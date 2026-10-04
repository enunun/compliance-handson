# 第12回 解説

解答のコードは`mise run iteration:solution 12`で当てられる．
差分は`iterations/12/solution.patch`にある．

## 12-1 準備

`EXC-004`は，2026年8月20日の再検査で見つかったHIGHの脆弱性の例外である．
「修正版を待つ」として，期限が2026年12月31日になっている．
見つかってから4か月以上，影響を調べないまま残すことになる．
今の`gate:exceptions`は，期限が基準値の90日のうちにあるかしか見ないので，これを止められない．

## 12-2 学ぶこと

この回の要点は，例外を「期限付きで受け入れる」だけのものから，「影響を判断した結果」を記録するものに変えることである．
影響がないものはVEXで根拠を示し，影響があるものは見つかった日から数えた期限で管理する．

## 12-3 テストリスト

### ゲートのテスト(`tests/gates/sla.test.ts`)

- [x] 影響のある脆弱性が，直すまでの期限を過ぎていれば失敗する．
- [x] 例外の期限が，直すまでの期限を超えていれば失敗する．
- [x] 期限内のものと，影響がないものだけなら通る．

### 既存のゲートのテストに足したもの

- [x] `gate:sca`：影響がない(`not_affected`)例外は，OpenVEXの文書にして外す(`sca.test.ts`)．
- [x] `ops:verify`：脆弱性の例外にトリアージの結果がなければ失敗する(`ops-verify.test.ts`)．

### 運用テスト(`tests/ops/escalation.test.ts`)

- [x] 期限を過ぎた脆弱性を，オンコールの担当者へのアラートとして送る．
- [x] 知らせたことが`ops`のロググループに残り，ログの設計に合う．

### 既存のテストへの影響

第2回の`gate:sca`のテストの例外(`lodash.yaml`)には状態がないので，これまでどおりTrivyの除外の設定で外れる．
OpenVEXで外す場合のfixture(`lodash-vex.yaml`)を足した．

`ops:verify`のfixtureの基準値には，`vulnerability.sla_days`を足した．

運用テストのファイルが増え，第11回のパスワードの交換と第9回のバックアップが同時に動いて，バックアップが古いパスワードで失敗した．
どの運用テストも同じ環境を使うので，`vitest.config.ts`の`ops`で`fileParallelism: false`にし，1つずつ動かす．

## 12-4 設計書

直すまでの日数は次のとおりにした．

```yaml
vulnerability:
  fail_severity: HIGH
  # 見つかってから直すまでの最長の日数．影響がない(not_affected)と判断したものには適用しない．
  sla_days:
    CRITICAL: 7
    HIGH: 30
    MEDIUM: 90
    LOW: 180
```

例外のスキーマでは，`if`と`then`で，脆弱性の例外にだけトリアージの結果を求めた．

```json
"allOf": [
  {
    "if": { "properties": { "target": { "type": "string", "pattern": "^(CVE|GHSA)-" } } },
    "then": {
      "required": ["package", "severity", "found", "status"],
      "properties": { "package": { "type": "string", "pattern": "^pkg:" } }
    }
  },
  {
    "if": { "properties": { "status": { "const": "not_affected" } }, "required": ["status"] },
    "then": { "required": ["justification"] }
  }
]
```

トリアージの項目の`records`は書いていない．
トリアージの結果は`exceptions.yaml`の変更履歴に残り，ログでは答えないからである．
エスカレーションは，第9回の`ops`のログで答えられる．

## 12-5 テスト駆動の実装

### トリアージする

`CVE-2024-29415`は，`isPublic`(と，その中で使う`isPrivate`)がアドレスの種類を誤る脆弱性である．
題材は`isV4Format`と`isV6Format`しか使わない．

```console
$ grep -rnwo "ip\.[A-Za-z0-9]*" apps/web/src --include=*.ts --exclude=*.test.ts
apps/web/src/client-ip.ts:12:ip.isV4Format
apps/web/src/client-ip.ts:12:ip.isV6Format
```

そこで，`EXC-004`は次のようにした．

```yaml
  - id: EXC-004
    target: CVE-2024-29415
    package: pkg:npm/ip
    severity: HIGH
    found: 2026-08-20
    status: not_affected
    justification: vulnerable_code_not_in_execute_path
    reason: >-
      脆弱性は isPublic と isPrivate がアドレスの種類を誤ることである．apps/web は isV4Format と
      isV6Format だけを使い，この2つを呼ばない(grep で確かめた)．修正版が出たら更新する．
    approver: "@security-lead"
    expires: 2026-12-31
```

`gate:sla`の結果は次のとおりになる．

```text
sla: 0 affected, 1 not_affected, 0 overdue
```

### OpenVEXへのリファクタリング

`toOpenVex`は，期限内の`not_affected`の例外だけを文書にする．
例外の理由と承認した人は，`impact_statement`に入れた．

```ts
statements: exceptions
  .filter((e) => e.status === "not_affected" && e.expires >= today)
  .map((e) => ({
    vulnerability: { name: e.target },
    products: [{ "@id": e.package }],
    status: "not_affected",
    justification: e.justification,
    impact_statement: `${e.id}: ${e.reason} (承認：${e.approver})`,
  })),
```

`products`には版のないpURL(`pkg:npm/ip`)を書いた．
Trivyは，版のないpURLをすべての版に当てる．
版を上げても`not_affected`の判断が変わらないかは，更新のときに確かめる．

### エスカレーション

`gate:sla`は，`ALERTMANAGER_URL`があれば，期限を過ぎたものを次のアラートとして送る．

```ts
labels: {
  alertname: "VulnerabilityOverdue",
  severity: "page",
  exception: e.id,
  vulnerability: e.target,
},
```

`rescan.yml`では，リポジトリの変数`ALERTMANAGER_URL`を渡し，再検査が失敗しても`gate:sla`を動かす(`if: always()`)．

## 12-6 振り返り

1. ゲートのテストで，`--today`で日付を決めたかを見る．決めないと，テストの結果が実行した日に左右される．
2. 直すのがよい．`ip`の使い方はNode.jsの`net.isIP`で置き換えられるので，置き換えればパッケージごと要らなくなる．
   `not_affected`は「すぐに直さなくてよい」根拠であり，直さなくてよい理由ではない．次の更新のときなど，急がない時期に直す．
3. 担当者だけに知らせ続けると，ほかの作業を優先して放置されやすい．上位者は，人を割り当てたり，リスクとして受け入れたりする判断ができる．
4. 人が書き忘れたり，遅い日付を書いたりできる．見つかった日を再検査の証跡から自動で記録する仕組みにすれば防げる．
   再検査の結果は第10回から証跡として残っているので，それと例外の`found`を突き合わせる方法がある．
5. 解答では，2つの項目の仕組み(`ci.yml`の`sla`ジョブ，`rescan.yml`の`rescan`ジョブ)があり，`ops:verify`が通る．

## 12-7 発展課題

`gate:sla`に`--report`(Trivyの結果のJSON)を足し，基準値の重大度以上で例外にない脆弱性を数える．
`ALERTMANAGER_URL`があれば，`VulnerabilityUntriaged`(`severity: ticket`)として送る．
第7回の振り分けで，`team-chat`に届く．
