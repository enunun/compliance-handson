# 第11回 解説

解答のコードは`mise run iteration:solution 11`で当てられる．
差分は`iterations/11/solution.patch`にある．

## 11-1 準備

- APIのタスクのロールは，AWSのすべての操作を許されている．APIが乗っ取られると，アカウント全体を操作される．
- CIのIAMの利用者は，長期のアクセスキーを持つ．キーが漏れると，無効化するまで使われ続ける．
- GitHub Actions向けのロールは`sub`が`repo:*`なので，GitHubのどのリポジトリのジョブからでも引き受けられる．

## 11-2 学ぶこと

この回の要点は，権限を「作るときに絞る」「定期的に見直す」「認証情報を交換し続ける」の3つの仕組みにすることである．

## 11-3 テストリスト

### ゲートのテスト(`tests/gates/iam.test.ts`)

- [x] ポリシーをコードの中に書いていれば失敗する．
- [x] 長期のアクセスキーを作っていれば失敗する．
- [x] すべての操作やすべての対象を許すポリシーがあれば失敗する．
- [x] OIDCの信頼ポリシーが`aud`と`sub`を絞っていなければ失敗する．
- [x] 権限を絞ったポリシーだけなら通る．

### 運用テスト(`tests/ops/access.test.ts`)

- [x] 交換の後は，新しいパスワードで接続でき，古いパスワードでは接続できない．
- [x] 棚卸しの報告に，IAMの利用者，アクセスキーを持つ利用者(いない)，CloudTrailの記録が残る．

### 既存のテストへの影響

第8回と第9回の運用テストは，OpenTofuの出力の`database_url`ではなく，`tools/ops/db.ts`の`databaseUrl()`で接続先を得る．
`database_url`は，交換するとすぐに古くなるからである．

## 11-4 設計書

棚卸しの項目の問いは2つにした．

```yaml
    records:
      - question: 決めた間隔で権限を棚卸ししていたことを示す
        asker: 監査人
        within: 依頼から1週間以内
        keep_days: 400
        log: ops
        fields: [time, job, result, detail]
      - question: 棚卸しの期間に，誰が誰の権限を変えたかを示す
        asker: セキュリティ責任者
        within: 棚卸しの日
        keep_days: 90
        log: audit
        fields: [time, actor_id, action, target_id, result]
```

2つ目の問いは，第8回の監査ログで答えられる．
ログの設計を変えずに済むのは，第8回で問いから項目を決めていたからである．
`audit`の保管期間(400日)は，この問いの90日を満たしている(`ops:verify`が確かめる)．

棚卸しの実施者は「人」にした．
報告は自動で作るが，外すべき権限の判断は人が行う．

## 11-5 テスト駆動の実装

### gate:iam

信頼ポリシーの`sub`の検査は次のとおりである．
`${repository}`は固定の値に置き換えてから，1つのリポジトリのブランチか環境に絞った形かを確かめる．

```ts
const narrowSub = /^repo:[^*/]+\/[^*]+:(ref:refs\/heads\/[^*]+|environment:[^*]+)$/;

const fixed = sub.map((v) => v.replace(/\$\{[^}]+\}/g, "owner/repo"));
if (fixed.length === 0 || !fixed.every((v) => narrowSub.test(v))) {
  problems.push(`${name}: OIDCの sub を1つのリポジトリのブランチか環境に絞っていない(${sub.join(", ") || "なし"})`);
}
```

変数の値は，OpenTofuの`validation`で確かめる．

```hcl
  validation {
    condition     = can(regex("^[^*/]+/[^*/]+$", var.github_repository))
    error_message = "github_repository は owner/repo の形で書き，ワイルドカード(*)を使わない．"
  }
```

### 権限を直す

APIのタスクのロールのポリシーは次のとおりである．
ロググループのARNの一覧は，`templatefile`へ渡す前に`join`で文字列にした．

```json
{
  "Sid": "WriteLogs",
  "Effect": "Allow",
  "Action": ["logs:CreateLogStream", "logs:PutLogEvents"],
  "Resource": ["${log_group_arns}"]
}
```

### パスワードの交換

MiniStackでは，`aws rds modify-db-instance`でパスワードを変えても，実際のPostgreSQLのパスワードは変わらない．
解答では，`ALTER ROLE`でデータベースのパスワードを変え，シークレットを更新した．
実際のAWSでも，Secrets Managerのパスワードの交換はこの方式で行う．

OpenTofuは，シークレットの値とデータベースのパスワードの違いを差分から外した．
外さないと，次の`tofu apply`で交換したパスワードが元に戻る．

## 11-6 振り返り

1. 権限を絞ったポリシーが通ることも確かめたかを見る．
2. 解答では，ワイルドカードを含むポリシーを書いて`trivy config`を実行し，検出されないことを確かめた．
   道具の検出に頼るときは，だめな入力で本当に検出されるかを，ゲートのテストで確かめておく．
3. 長期のアクセスキーは，無効化するまで誰でも使える．OIDCの認証情報は1時間程度で切れ，引き受けられるのも決めたリポジトリのブランチだけである．
4. 交換の手順は「データベースとシークレットを変える」だけで済み，使う側の設定を書き換える手順が要らなくなった．
5. 解答では，3つの項目の仕組み(`iam`ジョブ，`access.yml`の2つのジョブ)があり，`ops:verify`が通る．

## 11-7 発展課題

危ない操作の一覧(例えば`s3:DeleteObject`，`iam:*`，`kms:ScheduleKeyDeletion`)を基準値に置き，それを許す文には`Sid`に理由を書かせる方法がある．
ゲートは，危ない操作を許す文の`Sid`が基準値の形(`Justified`で始まるなど)かを確かめる．
