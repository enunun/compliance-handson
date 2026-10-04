# 第9回 解説

解答のコードは`mise run iteration:solution 9`で当てられる．
差分は`iterations/09/solution.patch`にある．

## 9-1 準備

今の題材では，データベースが壊れたり，データを誤って消したりしても戻せない．

## 9-2 学ぶこと

この回の要点は，バックアップを取ることではなく，戻せることを定期的に確かめ，その記録を残すことである．

## 9-3 テストリスト

### 運用テスト(`tests/ops/backup.test.ts`)

- [x] バックアップを取り，バックアップの保管場所に置く．証跡に，バックアップした時点の行数が残る．
- [x] 最新のバックアップを新しいデータベースに戻し，行数が一致する．
- [x] バックアップの保管場所は，基準値の日数だけ消せない設定である．
- [x] 作業の記録が`ops`のロググループに残り，ログの設計に合う．

### 既存のテストへの影響

第8回のログの運用テストは，ログの設計のすべての種類を確かめていた．
`ops`の行はバックアップの運用テストが作るので，ログの運用テストはAPIが書く種類(`api`と`audit`)だけを確かめる形に変えた．

## 9-4 設計書

問いとログの種類は次のとおりである．

```yaml
    records:
      - question: バックアップが決めた間隔で取れていたことを示す
        asker: 監査人
        within: 依頼から1週間以内
        keep_days: 400
        log: ops
        fields: [time, job, result, detail]
```

```yaml
  ops:
    description: 定期ジョブなど，運用の作業の記録．作業が決めたとおりに行われたことを示すのに使う．
    destination: CloudWatch Logs の /app/ops
    readers: [開発チーム, セキュリティ責任者]
    # fields と required は省略
```

基準値には，次のものを足した．

```yaml
logs:
  retention_days:
    api: 30
    audit: 400
    ops: 400

backup:
  interval_hours: 24
  retention_days: 35
```

`infra/logging.tf`は基準値の`logs.retention_days`からロググループを作るので，`/app/ops`のロググループはOpenTofuのコードを変えずにできる．

## 9-5 テスト駆動の実装

### バックアップ

バックアップは，書き出す前に各テーブルの行数を数え，ダンプと一緒に`.rows.json`として置く．
証跡には，置き場所，ファイルのSHA-256，行数を残す．

```ts
const rows = countRows(databaseUrl);
sh("pg_dump", ["--format=custom", "--file", file, databaseUrl]);
const sha256 = createHash("sha256").update(readFileSync(file)).digest("hex");
```

### リストアの訓練

訓練は，最新のバックアップを選び，新しいデータベースを作って戻し，行数を比べる．
作ったデータベースは，成功と失敗のどちらでも`finally`で消す．

```ts
const tables = Object.fromEntries(
  Object.keys(expected).map((t) => [t, { backup: expected[t], restored: restored[t] ?? 0 }]),
);
const ok = Object.values(tables).every((t) => t.backup === t.restored);
```

運用の作業の記録は，`tools/ops/run.ts`の`opsLog`で書く．
APIのログと同じく，1行のJSONを標準出力に書き，`LOG_GROUP_PREFIX`があればCloudWatch Logsにも送る．

## 9-6 振り返り

1. 行数が一致しない場合(失敗)のテストはない．入れるなら，`.rows.json`を書き換えたバックアップを置き，訓練が失敗することを確かめる．
2. 今のデータベースは，バックアップの後も利用者に使われて変わる．比べるべきは，バックアップした時点の内容である．
3. 行数が同じでも，中身が違うことはありうる．テーブルごとのチェックサム(例えば全行のハッシュ)をバックアップの時点で記録し，戻した先と比べる方法がある．
4. 本物の環境では，データベースはジョブの外にあり，ジョブはOIDCで得た権限を使ってバックアップを取る．
   訓練用のデータベースも，本番とは別のネットワークに作る．
5. 解答では，2つの項目の仕組み(`backup`ジョブ)が`backup.yml`にあり，`ops:verify`が通る．

## 9-7 発展課題

`backup.yml`の`on.schedule[0].cron`を読み，時の欄から1日の回数を求める．
例えば`17 18 * * *`なら1日に1回なので，間隔は24時間である．
この間隔が`backup.interval_hours`以下でなければ失敗にする．
