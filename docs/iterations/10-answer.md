# 第10回 解説

解答のコードは`mise run iteration:solution 10`で当てられる．
差分は`iterations/10/solution.patch`にある．

## 10-1 準備

Go 1.25のサポートは，2026年8月19日に終わっている．
脆弱性が見つかっても，修正が出ない．

## 10-2 学ぶこと

この回の要点は，ゲートが変更のときだけ動くことの穴を，定期的な仕組みで埋めることである．

## 10-3 テストリスト

### ゲートのテスト(`tests/gates/eol.test.ts`)

- [x] サポートの終わったベースイメージがあれば失敗し，終わった日を表示する．
- [x] 基準値の日数のうちにサポートの終わるランタイムがあれば失敗し，終わる日を表示する．
- [x] サポートの終わりが遠いものだけなら通る．

サポートの情報が見つからない版のテストは入れていない．
入れるなら，データにない版(例えば`go = "1.10.0"`)を書いた`mise.toml`のfixtureを足す．

### 既存のテストへの影響

`ops:verify`のfixtureの基準値に，`eol.warn_days`を足した．

## 10-4 設計書

運用項目は3つ足した．
依存関係の更新の仕組みは，設定ファイルである．

```yaml
  - id: dependency-updates
    title: 依存関係の更新
    category: 基盤運用
    requirements: [ISO27001:A.8.8, ISO27001:A.8.19]
    mechanisms:
      - file: renovate.json
    evidence: Renovateが作った更新のプルリクエストと，そのレビューの記録
    timing: 定常(毎週)と非定常(脆弱性の修正が出たとき)
    actor: 自動
```

`ops:verify`は，設定ファイルの場所をリポジトリの直下(ワークフローの場所の2つ上)からのパスとして確かめる．

サポートの終了の管理は，要求にISO/IEC 27001 A.5.21(ICTのサプライチェーンの管理)も書いた．
言語やOSの提供元のサポートも，サプライチェーンの一部だからである．

## 10-5 テスト駆動の実装

### gate:eol

版を探す処理は次のとおりである．

```ts
const release = (releases[t.product] ?? []).find(
  (r) => t.version === r.name || t.version.startsWith(`${r.name}.`),
);
```

`startsWith(r.name)`だけだと，Node.jsの`2`が`24.21.0`に当たってしまう．
`.`を付けて比べる．

終了日の比べ方は，第2回の例外の期限と同じく，`YYYY-MM-DD`の文字列で比べた．

### 例外を閉じる

golang-jwtを4.5.2に上げ，`EXC-001`を消した．
例外の一覧に残るのは，第4回の誤検知の2件だけになる．

### 再検査とRenovate

`rescan.yml`は，`git describe --tags --abbrev=0 --match 'v*'`で最新のリリースのタグを得る．
タグが指すイメージは第6回の`release.yml`がGHCRに置いたものなので，検査の前に署名を確かめてもよい(第13回でデプロイの前に行う)．

`renovate.json`は，RenovateのJSON Schemaで確かめた．

## 10-6 振り返り

1. 情報がない版のテストを入れたかを確かめる．情報がないものを通すと，名前の書き間違いも通る．
2. endoflife.dateのデータは日々変わる．問い合わせると，テストの結果が実行した日とデータに左右される．
3. 利点は，公開直後に取り消される版(壊れた版や，乗っ取られた版)を避けられることである．
   欠点は，脆弱性の修正が遅れることである．そのため，脆弱性の修正だけは`vulnerabilityAlerts`でスケジュールから外した．
4. ワークフローが失敗したことに，GitHubの通知で誰かが気づくのを待つことになる．気づく人も，対応の期限も決まっていない．
5. 解答では，3つの項目の仕組み(`renovate.json`，`rescan.yml`の`rescan`ジョブ，`ci.yml`の`eol`ジョブ)があり，`ops:verify`が通る．

## 10-7 発展課題

`collectTargets`の`imageProducts`に，distrolessのイメージの名前を足す．
タグがない(`nonroot`だけの)イメージは，名前の`debian12`から版`12`を取り出す．
endoflife.dateの`debian`の版`12`のデータで確かめる．
