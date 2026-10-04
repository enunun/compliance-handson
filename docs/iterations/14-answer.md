# 第14回 解説

解答のコードは`mise run iteration:solution 14`で当てられる．
差分は`iterations/14/solution.patch`にある．

## 14-1 準備

依存関係の脆弱性検査は，証跡の保管場所の`<日付>/<コミット>/sca/trivy.json`を見せられる．
第3回から，CIの`evidence`ジョブが毎回置いている．

ワークフローの部品の固定は，見せるものがない．
`gate:pinning`は結果を表示するだけで，証跡を残していない．
CIの実行の記録はGitHubに残るが，90日ほどで消え，保管場所にもない．

## 14-2 学ぶこと

この回の要点は，「仕組みの有無」と「仕組みの動いた証跡の有無」を分けて確かめることである．
運用項目に仕組みと証跡を書いてきたので，報告もチェックシートの回答も，運用項目から作れる．

## 14-3 テストリスト

### 単体テスト(`tools/ops/checksheet.test.ts`)

- [x] 質問の要求に合う運用項目があれば「はい」と答え，その項目を示す．
- [x] 合う運用項目がなければ「要確認」と答える．

### 運用テスト(`tests/ops/report.test.ts`)

- [x] その月の証跡が見つかった運用項目を示す．
- [x] 定常の運用項目で証跡が欠けていれば，それを示す．
- [x] 非定常の運用項目は，記録がなくても欠けたとしない．

### ゲートのテストに足したもの

- [x] `ops:verify`：人の動かすタスクが`mise.toml`になければ失敗する．
- [x] `ops:verify`：ログを絞る項目(`match`)が，ログの設計になければ失敗する．
- [x] `gate:exceptions`，`gate:pinning`，`gate:eol`，`gate:iam`：結果を証跡(`result.json`)に残す．

## 14-4 設計書

月次報告の項目の仕組みは，人が動かすタスクにした．

```yaml
  - id: monthly-report
    title: 月次報告
    category: 運用管理
    requirements: [SOC2:CC2.2, SOC2:CC4.1, ISO27001:A.5.35]
    mechanisms:
      - task: ops:report
    evidence: 月次報告(out/report-YYYY-MM.md)を証跡の保管場所に保管
    timing: 定常(毎月)
    actor: 人
    runbook: runbooks/monthly-report.md
```

報告は自動で作れるが，「なし」の原因を調べて対応を決めるのは人である．
CIで定期的に作ることもできるが，本物の証跡の保管場所を読む権限が要る(発展課題)．

`records`の`match`は，問いに答えるログを絞る条件である．

```yaml
    records:
      - question: バックアップが決めた間隔で取れていたことを示す
        asker: 監査人
        within: 依頼から1週間以内
        keep_days: 400
        log: ops
        fields: [time, job, result, detail]
        match: { job: backup }
```

第8回で「ログは問いから決める」としたので，報告が探すものも問いから決まる．

## 14-5 テスト駆動の実装

### 月次報告

本物の証跡で動かすと，4つの運用項目で証跡が見つからない．

```text
wrote out/report-2026-10.md
  items: 30 (checked: 19, evidence found: 15)
  missing evidence: exception-requests, actions-pinning, eol-tracking, least-privilege
  vulnerabilities over SLA: 0
  exceptions expiring within 30 days: 0
  OpenSSF Scorecard: 未取得
```

どれも，結果を表示するだけで，証跡を残さないゲートである．

### 欠けた証跡を直す

ゲートの結果を`<ゲート>/result.json`に残す関数を作り，4つのゲートから呼んだ．
証跡の置き場所を決める`evidenceDirOf`も，`trivy.ts`から`tools/gates/evidence.ts`に移した．
Trivyを使わないゲート(`sla`，`policy`)も使うからである．

```ts
export function writeResult(
  gate: string,
  result: { problems: string[] } & Record<string, unknown>,
): void {
  writeFileSync(
    join(evidenceDirOf(gate), "result.json"),
    `${JSON.stringify(result, null, 2)}\n`,
  );
}
```

`gates.yml`の4つのジョブにも，`evidence-<ゲート>`のartifactを足した．
直した後の報告は次のとおりである．

```text
wrote out/report-2026-10.md
  items: 30 (checked: 19, evidence found: 19)
  missing evidence: なし
  vulnerabilities over SLA: 0
  exceptions expiring within 30 days: 0
  OpenSSF Scorecard: 未取得
```

### チェックシート

```text
Q3 利用しているOSSの脆弱性を管理していますか → はい
    運用項目: dependency-vulnerabilities, exception-requests, container-image, iac-config, dependency-updates, release-rescan, eol-tracking, vulnerability-triage, escalation
...
Q9 データセンターへの物理的な入退室を管理していますか → 要確認
```

Q9は，SaaSの事業者ではなく，クラウドの事業者の責任の範囲である．
AWSのSOC 2の報告書(AWS Artifactで取得できる)を示して答える．

## 14-6 振り返り

1. 証跡の有無を確かめるテストで，「あり」だけでなく「なし」と「該当なし」も確かめたかを見る．
2. 脆弱性検査は保管場所の`sca/`，部品の固定は`pinning/result.json`を見せる．あわせて，月次報告で毎月「あり」だったことを示す．
3. 変更管理は，ブランチ保護の設定(第0回の`ops:evidence:change`)とプルリクエストの記録で確かめる．
   対象外の項目にも証跡を自動で集める仕組みを足せば，報告で確かめられる．
4. よくない．要求が重なっても，質問が問う範囲と運用項目の範囲が同じとは限らない．Q3では9つの運用項目が出るが，回答には，質問に直接答えるもの(依存関係の脆弱性検査とトリアージ)を選んで書く．
5. 例えば，次のものはこの講座で扱っていない．
   インシデント対応の手順と訓練，退職者のアカウントの削除の記録，個人情報の保管場所と削除の手順，委託先の管理．

## 14-7 発展課題

報告のロールのポリシーには，証跡のバケットの`s3:ListBucket`と`s3:GetObject`，ロググループの`logs:FilterLogEvents`だけを許す．
信頼ポリシーは，第11回と同じく`sub`をmainのブランチに絞る．
`report.yml`は`schedule`で毎月1日に動かし，前の月を`--month`で渡す．
