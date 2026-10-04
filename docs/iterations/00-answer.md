# 第0回 解説

解答のコードは`mise run iteration:solution 0`で当てられる．
差分は`iterations/00/solution.patch`にある．

## 0-1 準備

`mise run check`は，lint，型検査，単体テストを実行する．
この時点の題材には，運用項目とゲートがまだない．

## 0-2 学ぶこと

この回の要点は，仕組みと証跡を対で作ることである．
ブランチ保護は仕組みで，その設定を取得したJSONとプルリクエストのレビュー記録が証跡になる．

## 0-3 テストリスト

解答のテストリストは次のとおりである．

### ゲートのテスト(`tests/gates/ops-verify.test.ts`)

- [x] 正しい運用項目では，`ops:verify`が成功し`ops: ok`を表示する．
- [x] 仕組みのジョブがワークフローになければ，失敗し，どのジョブがないかを表示する．
- [x] 基準値がスキーマに合わなければ，失敗し，合わない場所(`/change/required_approvals`)を表示する．
- [x] 運用項目一覧が`items.yaml`と合わなければ，失敗し，`mise run ops:render`を案内する．

### 単体テスト(`tools/ops/render.test.ts`)

- [x] 運用項目を1行ずつ表にする．要求と仕組みが複数あれば`<br>`でつなぐ．
- [x] 手順書のない項目は，手順書の列を空にする．

手順書のファイルがないときの検査は，ゲートのテストに入れていない．
入れる場合は，`runbook`に存在しないファイルを書いたfixtureを足す．

## 0-4 設計書

### 基準値と運用項目

`ops/standards.yaml`は，承認の数だけを持つ．

```yaml
change:
  required_approvals: 1
```

`ops/items.yaml`の最初の項目は次のとおりである．
分類は，変更の承認という運用そのものの管理なので「運用管理」にした．

```yaml
items:
  - id: change-management
    title: 変更管理
    category: 運用管理
    requirements: [SOC2:CC8.1, ISO27001:A.8.32]
    mechanisms:
      - workflow: ci.yml
        job: check
    evidence: ブランチ保護の設定(JSON)と，プルリクエストのレビュー記録
    timing: 定常(変更ごと)
    actor: 自動
    runbook: runbooks/emergency-change.md
```

スキーマでは，`additionalProperties: false`で書き間違えた項目名を検出する．
`category`と`actor`は`enum`で値を限り，`id`は`pattern`で英小文字とハイフンに限った．
仕組み(`mechanisms`)は，後の回でアラートのルールなども入るように，ワークフローとジョブの組の配列にした．

### 手順書と運用方針

緊急変更の手順書には，実施できる人，手順，証跡を書いた．
ブランチ保護の設定は変えず，管理者の権限でマージする．
設定を外すと，外していた間の変更を追えなくなるからである．

運用方針には，開発チーム，セキュリティ責任者，リポジトリ管理者の3つの役割を書いた．
基準値の承認者はCODEOWNERSの`@security-lead`で，自分のアカウントに置き換えて使う．

## 0-5 テスト駆動の実装

### ops:verify

最初のテスト(正しい運用項目では通る)は，コマンドとして実行したときに`ops: ok`を表示するだけで通る．

```ts
if (import.meta.main) {
  console.log("ops: ok");
}
```

次のテスト(ジョブがない)で，運用項目を読んでワークフローと突き合わせる処理を書く．
`verify`は問題の配列を返し，表示と終了コードはコマンドの側で決める．
こうすると，`verify`の結果を後の回の別の道具からも使える．

```ts
for (const item of items.items) {
  for (const { workflow, job } of item.mechanisms) {
    const path = join(options.workflowsDir, workflow);
    if (!existsSync(path)) {
      problems.push({ file: itemsFile, message: `${item.id}: ワークフロー ${workflow} がない` });
      continue;
    }
    const jobs = (readYaml(path) as { jobs?: Record<string, unknown> }).jobs;
    if (!jobs || !(job in jobs)) {
      problems.push({ file: itemsFile, message: `${item.id}: ${workflow} にジョブ ${job} がない` });
    }
  }
}
```

3つ目のテスト(スキーマ)で，Ajvでの検証を足す．
`allErrors: true`にすると，最初の誤りで止まらずにすべての誤りを返す．
スキーマに合わないときは，その先の突き合わせを行わずに返す．

4つ目のテスト(古い一覧)では，`render`の結果と`ops/items.md`の中身を比べる．

実行すると，問題のあるファイルと理由を表示して失敗する．

```text
tests/gates/fixtures/ops-verify/missing-job/ops/items.yaml: change-management: ci.yml にジョブ check がない
```

```text
tests/gates/fixtures/ops-verify/invalid-standards/ops/standards.yaml: /change/required_approvals must be >= 1
```

### ops:render

`render`は，項目を表の1行に変える．
複数の要求と仕組みは`<br>`でつなぎ，1つのセルに収めた．

### 検証のコマンド

`mise.toml`の`check`は次のようになる．

```toml
[tasks.check]
description = "題材の検査をまとめて実行する(CIと同じ)"
depends = ["lint", "test", "test:gates", "ops:verify"]
```

### CIとブランチ保護

`ci.yml`は，`defaults.run.working-directory`で`app/`を作業ディレクトリにする．
GitHubが実行するのはリポジトリ直下の`.github/workflows/`だけだからである．

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: app
    steps:
      - uses: actions/checkout@v5
      - uses: jdx/mise-action@v3
      - run: mise run check
```

`uses:`はタグで参照している．
タグは後から別のコミットに付け替えられるので，第6回でハッシュに固定する．

### 証跡

`ops:evidence:change`は，`gh api`の結果を日付つきのファイルに保存する．
`gh api`のパスの`{owner}`と`{repo}`は，ghがgitのリモートから埋める．

## 0-6 振り返り

1. 手順書がないときの検査を入れたかを確かめる．入れていなければ，0-3の注記のとおりfixtureを足す．
2. 理由まで確かめないと，別の理由(例えばYAMLの読み込みの失敗)で失敗していても，テストが通ってしまう．
3. 基準値の変更は方針の決定で，セキュリティ責任者が承認する．運用項目の変更は開発チームが承認する．
   CODEOWNERSはファイル単位で承認者を決めるので，承認者が違うものはファイルを分ける．
4. 画面の設定は，いつ誰が変えたかが残りにくい．監査の時点で「その期間ずっと保護されていた」ことを示せない．
   設定を定期的にJSONで保存すると，その時点の設定を示せる．
5. 解答では，仕組みの`check`ジョブと`ci.yml`のジョブ名が一致している．`ops:verify`が通れば一致している．

## 0-7 発展課題

`verify`に，CODEOWNERSを読む処理を足す例は次のとおりである．
CODEOWNERSの場所は，`VerifyOptions`に`codeownersPath`を足して渡す．

```ts
const owners = readFileSync(options.codeownersPath, "utf8")
  .split("\n")
  .filter((line) => line.trim() && !line.startsWith("#"));
if (!owners.some((line) => line.split(/\s+/)[0] === "/app/ops/standards.yaml")) {
  problems.push({ file: options.codeownersPath, message: "standards.yaml の承認者がない" });
}
```

ゲートのテストには，`standards.yaml`の行がないCODEOWNERSのfixtureを足す．
