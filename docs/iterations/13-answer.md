# 第13回 解説

解答のコードは`mise run iteration:solution 13`で当てられる．
差分は`iterations/13/solution.patch`にある．

## 13-1 準備

言えない．
`deploy.yml`はタグでイメージを指定し，署名とprovenanceを確かめずにデプロイする．
タグは付け替えられるので，GHCRに書き込める人や乗っ取られたトークンがあれば，別のイメージを本番で動かせる．
第6回で署名を付けても，使う側で確かめなければ意味がない．

## 13-2 学ぶこと

この回の要点は，基準と検査を「各リポジトリに書き写すもの」から「1か所から参照するもの」に変えることである．
決まりはconftestのルールとして書き，テストできる形にする．

## 13-3 テストリスト

### ゲートのテスト(`tests/gates/policy.test.ts`)

- [x] 署名を検証する前にデプロイするジョブがあれば失敗する．
- [x] 最上位に`permissions`のないワークフローがあれば失敗する．
- [x] 人が行う運用項目に手順書がなければ失敗する．
- [x] ルールに合うものだけなら通る．

### 単体テスト(`tools/ops/deploy.test.ts`)

- [x] ダイジェストで指定したイメージは受け付ける．
- [x] タグで指定したイメージは拒む．

### 運用テスト(`tests/ops/deploy.test.ts`)

- [x] ダイジェストで指定していないイメージはデプロイしない．

### 既存のテストへの影響

ゲートのテストは変わらない．
ゲートのジョブを移すと，`ops:verify`が，運用項目の仕組みのジョブが`ci.yml`にないことを12件指摘する．

```text
ops/items.yaml: secret-detection: ci.yml にジョブ secrets がない
ops/items.yaml: dependency-vulnerabilities: ci.yml にジョブ sca がない
...
```

`items.yaml`の仕組みを`gates.yml`に替えて直した．

## 13-4 設計書

デプロイ前の検証の項目は次のとおりである．

```yaml
  - id: deploy-verification
    title: デプロイ前の検証
    category: 基盤運用
    requirements: [SSDF:PS.2, SLSA:Build-L2, ISO27001:A.8.32]
    mechanisms:
      - workflow: deploy.yml
        job: deploy
    evidence: デプロイの前の署名とprovenanceの検証結果，デプロイの結果
    timing: 非定常(デプロイのとき)
    actor: 人
    runbook: runbooks/verify-image.md
    records:
      - question: 本番で動いているイメージを，いつ誰の操作でデプロイしたかを示す
        asker: 監査人
        within: 依頼から1週間以内
        keep_days: 400
        log: ops
        fields: [time, job, result, detail]
```

「誰の操作で」に答えるため，`deploy.ts`は`detail`に`GITHUB_ACTOR`(ワークフローを動かした人)を書く．
ログの設計は変えずに済んだ．

`actor`を「人」にしたのは，デプロイを人が決めて動かすからである．
すると，この回で作った`items.rego`のルールにより，手順書が必須になる．

## 13-5 テスト駆動の実装

### gate:policy

デプロイのルールは次のとおりである．

```rego
deny contains msg if {
	some name, job in input.jobs
	some i, step in job.steps
	contains(step.run, "mise run deploy")
	not verified_before(job, i)
	msg := sprintf("ジョブ %s が，mise run release:verify より前に mise run deploy を動かしている", [name])
}

verified_before(job, i) if {
	some j, step in job.steps
	j < i
	contains(step.run, "mise run release:verify")
}
```

`tools/gates/policy.ts`は，ワークフローと運用項目を別の`--namespace`で検査し，結果を1つの証跡(`policy/conftest.json`)にまとめる．
conftestの終了コードは，ルールに反すると1，ルールの読み込みなどに失敗すると2以上になる．
1以外の失敗は例外にして，ルールの書き間違いを「通った」と扱わないようにした．

### デプロイを直す

`deploy.yml`の手順は次の順にした．

1. GHCRにログインし，タグからダイジェストを得る．
2. `mise run release:verify`で署名とprovenanceを検証する．
3. OIDCでAWSのロールを引き受ける．
4. ダイジェストで指定してデプロイする．

```yaml
      - name: 署名とprovenanceを検証し，証跡に残す
        env:
          GH_TOKEN: ${{ github.token }}
          REF: ${{ steps.image.outputs.ref }}
        run: mise run release:verify "${REF}"
```

`${{ steps.image.outputs.ref }}`を`run`に直接書くと，入力のタグに`"; curl ...`のような値を入れられたとき，シェルのコマンドとして動く．
`env`で渡すと，値は変数の中身として扱われる．

### MiniStackのECS

AWSプロバイダは，ネットワークの設定のないECSのサービスを作ると，MiniStackの応答を読むところで落ちる．
題材では，Fargateのサービスにして，VPC，サブネット，セキュリティグループを足した．
本番のAWSでも，Fargateのサービスにはネットワークの設定が要る．

## 13-6 振り返り

1. ルールのテストで，良い例が通ることも確かめたかを見る．だめな例だけだと，何でも失敗するルールも通ってしまう．
2. 検証に失敗したら，AWSの認証情報を一度も得ずに終わるからである．権限を使う前に止めるほど，すり替えられたイメージで何かをされる余地が減る．
3. タグは付け替えられるので，呼ぶ側の知らないうちにゲートの中身が変わる．ゲートを弱める変更や，悪意のある変更も入りうる．第6回のactionと同じく，ハッシュで固定し，更新はRenovateのプルリクエストで確かめる．
4. 新しいルールを，まず警告だけにする方法がある(conftestの`warn`)．各リポジトリが直してから`deny`に変える．ほかのリポジトリはハッシュで固定しているので，取り込む時期を各自で選べる．
5. 解答では，2つの項目の仕組み(`gates.yml`の`policy`ジョブ，`deploy.yml`の`deploy`ジョブ)があり，`ops:verify`が通る．

## 13-7 発展課題

`pull_request_target`は，フォークからのプルリクエストでも，書き込みの権限とシークレットを持って動く．
プルリクエストのコードをチェックアウトすると，そのコードが権限を持って動く．
ルールでは，`input.on`に`pull_request_target`があり，`actions/checkout`のステップの`with.ref`に`github.event.pull_request.head`を含むものを`deny`にする．
