# 第7回 解説

解答のコードは`mise run iteration:solution 7`で当てられる．
差分は`iterations/07/solution.patch`にある．

## 7-1 準備

APIはメトリクスを出しておらず，エラーが増えても誰にも知らせない．

## 7-2 学ぶこと

この回の要点は，しきい値，ルール，通知先をすべてコードにし，テストで確かめられるようにすることである．
画面で作ったアラートは，誰がいつ変えたかが残らず，テストもできない．

## 7-3 テストリスト

### 単体テスト

- [x] API：`/metrics`は，ルートとステータスコードごとのリクエストの数を返す．

### 運用テスト(`tests/ops/alert-rules.test.yml`，`mise run test:alerts`)

- [x] エラーの割合が10%なら，5分続いた後で`ApiHighErrorRate`が発火する．3分の時点では発火しない．
- [x] エラーの割合が1%なら発火しない．
- [x] エラーが止まれば，アラートも止む．
- [x] `severity=page`はオンコールの担当者(`on-call`)に，`severity=ticket`はチームのチャット(`team-chat`)に届く．

### 既存のテストへの影響

- 運用項目のスキーマで，仕組みを「ワークフローとジョブ」か「アラート」のどちらか(`oneOf`)にした．
- `ops:verify`のfixtureの基準値に，監視の基準を足した．

## 7-4 設計書

基準値と運用項目は次のとおりである．

```yaml
monitoring:
  # APIのエラー(5xx)の割合がこれを超えた状態が error_rate_for の間続いたら，担当者に知らせる．
  error_rate_threshold: 0.05
  error_rate_for: 5m
```

```yaml
  - id: service-monitoring
    title: APIの監視
    category: 基盤運用
    requirements: [ISO27001:A.8.16, SOC2:CC7.2]
    mechanisms:
      - alert: ApiHighErrorRate
    evidence: アラートのルールとそのテストの結果，Alertmanagerの通知の記録
    timing: 定常(常時)
    actor: 自動
    runbook: runbooks/api-high-error-rate.md
```

手順書には，30分で収まらなければ開発チームの全員を呼ぶことを書いた．
人を呼ぶ基準がないと，担当者が1人で抱え込み，対応が遅れる．

## 7-5 テスト駆動の実装

### メトリクス

ハンドラを包んで数える関数は次のとおりである．
ルートの名前には，`http.ServeMux`に登録した形(`GET /api/notes`)を使う．
リクエストのURLをそのまま使うと，`/api/users/1/role`のようにIDごとに別の値になり，数値の種類が増え続ける．

```go
func (m *metrics) instrument(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		m.requests.WithLabelValues(route, strconv.Itoa(rec.code)).Inc()
		m.duration.WithLabelValues(route).Observe(time.Since(start).Seconds())
	})
}
```

メトリクスは，既定の登録先ではなく`prometheus.NewRegistry()`に登録した．
テストで`server.New`を何度呼んでも，登録が重ならない．

### アラートのルール

ルールは`tools/ops/alerts.ts`で基準値から作る．

```ts
export function alertRules(standards: Standards) {
  const { error_rate_threshold, error_rate_for } = standards.monitoring;
  return {
    groups: [
      {
        name: "api",
        rules: [
          {
            alert: "ApiHighErrorRate",
            expr: `sum(rate(http_requests_total{code=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > ${error_rate_threshold}`,
            for: error_rate_for,
            labels: { severity: "page" },
            annotations: {
              summary: `APIのエラーの割合が${error_rate_threshold * 100}%を超えている`,
              runbook: "ops/runbooks/api-high-error-rate.md",
            },
          },
        ],
      },
    ],
  };
}
```

`ops:verify`は，`rules.yml`が基準値から作ったものと同じかを確かめる．
運用項目の`alert`の仕組みは，`rules.yml`にそのアラートがあるかで確かめる．

テストの`eval_time`は，3分(窓が埋まる前)と15分(窓が埋まり，`for`の5分もたった後)にした．
止むことのテストでは，15分の後でエラーの数を増やすのをやめ，30分で発火していないことを確かめる．

### 通知先とDev Container

Prometheusの設定の`rule_files`は，設定ファイルからの相対パス(`rules.yml`)で書いた．
手元での`promtool check config`と，コンテナの中(`/etc/prometheus/`)の両方で同じファイルを指す．

Grafanaは，Web画面と重ならないように3001番で動かした．
Dev Containerの`forwardPorts`には，`"grafana:3001"`のようにサービスの名前を付けて書く．

## 7-6 振り返り

1. 発火しないことのテスト(低い割合，止んだ後)も入れたかを確かめる．発火することだけでは，常に発火するルールも通ってしまう．
2. 基準値と実際のしきい値がずれる．監査で「基準値は5%」と示しても，実際には違う値で動いていることになる．
3. 一瞬のエラーでも知らせることになり，夜中に何度も起こされる．通知が多すぎると，本物のアラートも無視されるようになる．
4. 対応が1人に偏ると，長引いたときに判断が遅れる．呼ぶ基準があれば，迷わずに人を増やせる．
5. 解答では，運用項目のアラートが`rules.yml`にあり，`ops:verify`が通る．

## 7-7 発展課題

基準値に`monitoring.latency_p95_seconds: 1`を足し，`alerts.ts`にルールを足す．

```text
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m]))) > 1
```

重大度は`ticket`にし，`test:alerts`の振り分けのテストで`team-chat`に届くことを確かめる．
