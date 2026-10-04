# 第1回 解説

解答のコードは`mise run iteration:solution 1`で当てられる．
差分は`iterations/01/solution.patch`にある．

## 1-1 準備

`mail.go`は，APIキーをコードに直書きしている．
コミットした時点で，キーは履歴に入った．

## 1-2 学ぶこと

この回の要点は，検出の仕組みと，漏れたときの対応の手順を対で作ることである．
検出だけでは，見つかった後に何をするかが人によって変わる．

## 1-3 テストリスト

### ゲートのテスト(`tests/gates/secrets.test.ts`)

- [x] `gate:secrets`：キーを含むコミットが履歴にあれば失敗し，`leaks found: 1`を表示する．
- [x] `gate:secrets`：キーを含まないリポジトリでは通り，`no leaks found`を表示する．
- [x] `gate:secrets`：出力と結果のファイルに，キーそのものを書かない．
- [x] `gate:secrets:staged`：コミットしようとしている変更にキーがあれば失敗する．
- [x] `gate:secrets:staged`：コミットしようとしている変更にキーがなければ通る．

### 既存のテストへの影響

ない．
`check`を分けたので，CIの`check`ジョブが実行するタスクの名前が変わる．

## 1-4 設計書

`ops/items.yaml`に足した項目は次のとおりである．
漏洩を見つけて連絡と対応につなぐ運用なので，分類は運用管理(運用情報統制)にした．

```yaml
  - id: secret-detection
    title: シークレット検出
    category: 運用管理
    requirements: [ISO27001:A.5.17, ISO27001:A.8.4, SSDF:PS.1]
    mechanisms:
      - workflow: ci.yml
        job: secrets
    evidence: gitleaksの結果(JSON，キーは伏せ字)をCIのartifactとして保存
    timing: 定常(変更ごと，コミットごと)
    actor: 自動
    runbook: runbooks/secret-leak.md
```

コミットの前の検査(lefthook)は，手元でしか動かず，`--no-verify`で飛ばせる．
そのため，運用項目の仕組みにはCIの`secrets`ジョブを書いた．
コミットの前の検査は，早く気づくための補助である．

手順書では，無効化を最初の手順にした．
記録欄には，日付，シークレット，見つけた場所，無効化，利用の確認，対応した人を書く．
シークレットの値は，先頭の数文字だけを書く．

## 1-5 テスト駆動の実装

### gate:secrets

最初のテストは，キーを含むfixtureのリポジトリで失敗することである．
タスクは次のようになる．

```toml
[tasks."gate:secrets"]
description = "gitの履歴全体からシークレットを探す(引数でリポジトリの場所を変えられる)"
run = [
  "mkdir -p out/evidence/secrets",
  "gitleaks git --config ../.gitleaks.toml --redact --report-format json --report-path out/evidence/secrets/gitleaks.json",
]
```

引数がなければ，gitleaksは`app/`を調べる．
gitは`app/`を含むリポジトリ全体の履歴を返すので，リポジトリ全体を調べることになる．

テストでは，一時的なリポジトリを次のように作る．
最初に空のコミットを作るのは，`gate:secrets:staged`で差分の基準にするためである．

```ts
function repoFrom(name: string, { commit = true } = {}): string {
  const dir = mkdtempSync(join(tmpdir(), "gate-secrets-"));
  git(dir, "init", "-q");
  git(dir, "commit", "-q", "--allow-empty", "-m", "init");
  cpSync(fixture("secrets", name), dir, { recursive: true });
  git(dir, "add", "-A");
  if (commit) {
    git(dir, "commit", "-qm", name);
  }
  return dir;
}
```

### .gitleaks.toml

許可リストは2つにした．

```toml
[extend]
useDefault = true

[[allowlists]]
description = "ゲートのテストとその入力，教材の仕込みのパッチ"
paths = ['''^app/tests/gates/''', '''^iterations/''']

[[allowlists]]
description = "無効化済みのキー．対応の記録は app/ops/runbooks/secret-leak.md の記録欄にある"
regexTarget = "secret"
regexes = ['''^mk_9f2c7a41e8b35d06c1f4a7e92b8d3c5f$''']
```

ゲートのテストは，確かめるためにキーを含む．`iterations/`は，この教材の仕込みのパッチを置く場所である．

無効化したキーは，コミットのハッシュではなく値で指定した．
gitleaksは検出ごとに，コミットのハッシュを含む指紋(`Fingerprint`)を出し，`.gitleaksignore`に指紋を書いても外せる．
ただし，ハッシュはリポジトリごとに違うので，この教材では値で指定した．
許可リストに値を書いても，キーは無効化済みなので害はない．

### 検証のコマンドとCI

`mise.toml`の検査は次のようになる．

```toml
[tasks."check:code"]
description = "コードと運用項目を検査する(CIのcheckジョブ)"
depends = ["lint", "test", "test:gates", "ops:verify"]

[tasks.gates]
description = "すべてのゲートを実行する(CIではゲートごとのジョブ)"
depends = ["gate:secrets"]

[tasks.check]
description = "題材の検査とゲートをまとめて実行する"
depends = ["check:code", "gates"]
```

`ci.yml`の`secrets`ジョブは，履歴全体を取得してゲートを実行し，結果を残す．

```yaml
  secrets:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: app
    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0
      - uses: jdx/mise-action@v3
      - run: mise run gate:secrets
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: evidence-secrets
          path: app/out/evidence/secrets/
```

`lefthook.yml`には，`pre-commit`のコマンドを足した．

```yaml
    secrets:
      run: mise -C app run gate:secrets:staged
```

## 1-6 振り返り

1. ステージした変更の検査を入れたかを確かめる．
2. 仕込みと同じキーを使うと，そのキーを許可リストに足した後はテストが通らなくなる．
   fixtureのキーは，許可リストのどの条件にも当たらないものにする．
3. 履歴を書き換えても，すでにcloneやforkをした人の手元には残る．
   キーを使えなくするには，発行元での無効化しかない．
4. コミットの前の検査は手元で飛ばせる．CIの検査は履歴全体を見るが，pushした後にしか動かない．
   両方を置くと，早く気づけて，飛ばしても止まる．
5. 解答では，運用項目の`secrets`ジョブが`ci.yml`にあり，`ops:verify`が通る．

## 1-7 発展課題

`.gitleaks.toml`にルールを足す例は次のとおりである．

```toml
[[rules]]
id = "mail-service-api-key"
description = "メール配信サービスのAPIキー"
regex = '''\bmk_[0-9a-f]{32}\b'''
keywords = ["mk_"]
```

ゲートのテストには，このルールでだけ見つかる書き方(例えば変数名に`key`を含まない代入)のfixtureを足す．
