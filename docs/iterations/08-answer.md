# 第8回 解説

解答のコードは`mise run iteration:solution 8`で当てられる．
差分は`iterations/08/solution.patch`にある．

## 8-1 準備

足されたログには，次の問題がある．

- 形式が決まっていないので，項目で検索できない．どのリクエストのログかも分からない．
- ログインの失敗のログにパスワードを，成功のログにトークンを書いている．ログを読める人なら誰でも，他人になりすませる．
- 権限の変更は記録されない．監査で「誰が権限を変えたか」を示せない．

## 8-2 学ぶこと

この回の要点は，ログの項目と保管期間に，運用上の理由を持たせることである．
理由のない項目は，漏れたときの危険だけが残る．
理由のない保管期間は，短すぎて答えられないか，長すぎて費用と危険が増える．

## 8-3 テストリスト

### 単体テスト(API)

- [x] リクエストのログに，リクエストID，ルート，ステータスコードがあり，リクエストIDは応答の`X-Request-Id`と同じである．
- [x] ログインの失敗を監査ログに書き，パスワードは書かない．
- [x] ログインの成功を監査ログに書き，トークンは書かない．
- [x] 権限の変更を，操作した人と対象とともに監査ログに書く．
- [x] CloudWatch Logsへ送るハンドラは，`log_type`ごとのロググループに送り，`log_type`のないログは送らない．

### ゲートのテスト(`ops:verify`)

- [x] 問いが，ログの設計にない項目を使えば失敗する．
- [x] ログの保管期間が，問いの遡る期間より短ければ失敗する．

### 運用テスト(`tests/ops/logs.test.ts`)

- [x] `api`と`audit`のロググループの各行が，ログの設計に合う．
- [x] 監査ログで，権限の変更とログインの失敗を検索できる．
- [x] どのロググループにも，パスワードとトークンを書かない．
- [x] ロググループの保管期間は，基準値と同じである．

### 既存のテストへの影響

`server.New`がロガーを受け取るので，APIの単体テストは`newServer(t, w)`でAPIを作るように変えた．
`ops:verify`のfixtureの基準値には，`logs.retention_days`を足した．

## 8-4 設計書

問い(`records`)は次のとおりである．

| 運用項目 | 問い | 誰が | いつまでに | 遡る期間 | ログ |
| --- | --- | --- | --- | --- | --- |
| service-monitoring | アラートの原因になったリクエストを特定する | オンコールの担当者 | 発生から1時間以内 | 30日 | api |
| audit-logging | 権限を変えた人と時刻を示す | 監査人 | 依頼から1週間以内 | 400日 | audit |
| audit-logging | ログインの失敗が続くアカウントを見つける | セキュリティ責任者 | 発生から1日以内 | 90日 | audit |

ここから，ログの種類を次のように分けた．

- `api`：障害の調査に使う．読むのは開発チームとオンコールの担当者で，30日遡れればよい．
- `audit`：監査と不正の調査に使う．読むのはセキュリティ責任者で，監査の対象期間(1年)に余裕を持たせて400日遡る．

保管期間は，基準値に次のように書いた．
`audit`は，問いの中で最も長い400日に合わせた．

```yaml
logs:
  retention_days:
    api: 30
    audit: 400
```

`audit`の項目は，2つの問いの項目を合わせたものである．
権限の変更には`actor_id`，`action`，`target_id`，`result`が，ログインの失敗には`actor_email`，`source_ip`が要る．
記録してはならない項目として，`password`，`token`，`authorization`，`cookie`を決めた．

## 8-5 テスト駆動の実装

### 設計の検証

`checkRecords`は，問いごとに種類と項目を確かめ，種類ごとに最も長い遡る期間を集めて保管期間と比べる．

```ts
for (const [name, type] of Object.entries(logs.types)) {
  const retention = standards.logs.retention_days[name];
  if (retention === undefined) {
    problems.push(`ログの種類 ${name} の保管期間が standards.yaml にない`);
  } else if (retention < (longest[name] ?? 0)) {
    problems.push(
      `ログの種類 ${name} の保管期間(${retention}日)が，問いの keep_days(${longest[name]}日)より短い`,
    );
  }
  // 記録してはならない項目の検査は省略
}
```

ログの1行を検証するJSON Schemaは，`additionalProperties: false`にした．
設計にない項目がログに現れたら，運用テストが失敗する．

### APIのログ

リクエストのログは，ハンドラを包む関数で書く．
リクエストIDは`context`に入れ，監査のログやエラーのログでも同じ値を使う．

```go
func withRequestLog(logger *slog.Logger, route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
		// ここで api のログを1行書く(省略)
	})
}
```

権限の変更では，成功と失敗のどちらでも監査ログを書く．
失敗した操作も，不正の試みの手がかりになるからである．

CloudWatch Logsへ送るハンドラは，1行ごとに`PutLogEvents`を呼ぶ．
講座の規模では足りるが，量の多い本番では，ログの転送の仕組み(ECSならFireLensなど)に替える．

### 運用テスト

運用テストは，APIをビルドして，`mise run up`で作ったデータベースにつないで起動する．
ログインの失敗，ログイン，権限の変更の後で，各ロググループの行を検証する．

## 8-6 振り返り

1. 運用テストで，設計にない項目を検出できることを確かめたかを見る．`additionalProperties: false`がなければ，検出できない．
2. 同じ種類にすると，保管期間は長い方(400日)へ，閲覧できる人は広い方へ合わせることになる．
   障害の調査のログを必要以上に長く保管し，監査ログを必要以上に多くの人が読めるようになる．
3. ログインの失敗が続くアカウントを見つけるには，どのアカウントが狙われたかが要る．
   そのため，閲覧できる人をセキュリティ責任者に絞り，運用方針に書いた．
4. ログの項目が増えるほど，漏れたときの影響が広がる．運用テストは，設計にない項目が現れたら失敗するので，項目を足すには先に設計(と問い)を変える必要がある．
5. 解答では，`audit-logging`と`log-design`の仕組みが`ci.yml`にあり，`records`の検証を含めて`ops:verify`が通る．

## 8-7 発展課題

問いは「同じメールアドレスのログインの失敗を，10分ごとに数える」である．
`audit`の`actor_email`，`action`，`result`，`time`で答えられるので，ログの設計は変えなくてよい．
仕組みとしては，CloudWatch Logsのメトリクスフィルタで数を数え，アラームを作る方法がある．
MiniStackではメトリクスフィルタが数値を出さないので，この講座の環境では，定期ジョブで`filter-log-events`の結果を数える．
