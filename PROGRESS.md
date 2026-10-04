# 教材づくりの進み具合

教材を作る人とエージェントのための作業の記録である．
区切りごとに更新する．計画は`COURSE.md`，各回の要件は`docs/ROADMAP.md`にある．

## 今の状況

| 回 | 状況 |
| --- | --- |
| 0〜14 | 完成．パッチ(`iterations/NN/`)，演習の手順(`docs/iterations/NN.md`)，解説(`NN-answer.md`)がそろい，`mise run iterations:verify`が通る． |

## 作り方

各回の題材の変更は，作業用のリポジトリ(scratchpadの`build/`)にコミットとして積み，差分をパッチとして書き出す．
作業用のリポジトリは消えることがあるが，パッチが正なので作り直せる．

作り直す手順は次のとおりである．

1. このリポジトリの`app/`，`.github/`(あれば)，`.devcontainer/`，`mise.toml`，`mise.lock`，`.mise/`，`lefthook.yml`，`.gitignore`をコピーし，gitのリポジトリにする．
2. `iterations/00/`から順に，`problems.patch`(あれば)と`solution.patch`を当て，回ごとにコミットとタグ(`exN`，`solN`)を付ける．
3. 次の回は，`solN`の上に仕込み(`exN+1`)と解答(`solN+1`)を積む．
4. パッチは`git diff solN exN+1 > iterations/N+1/problems.patch`と`git diff exN+1 solN+1 > iterations/N+1/solution.patch`で書き出す．

回ごとの手順は次のとおりである．

1. 仕込みを作り，`mise run check:code`が通ることを確かめてコミットする．
2. 解答を作る．テストを先に書く．`mise run check`(と必要なら`mise run test:ops`)が通るまで直す．
3. 教材に載せる出力は，実際の実行から写す．仕込みの状態での失敗の出力は，`git worktree`で取る．
4. パッチを書き出し，`docs/ROADMAP.md`のその回を実際の内容に合わせ，演習の手順と解説を書く．`README.md`の表に1行足す．
5. `pnpm textlint`と`pnpm markdownlint-cli2`を通し，`mise run iterations:verify <回> <回>`で確かめてコミットする．

## この検証環境での回避策

- Dockerのデーモンは`dockerd`を手で起動する．止まっていたら起動し直す．
- Docker Hubの取得回数の制限を避けるため，`/etc/docker/daemon.json`にミラー(`mirror.gcr.io`)を設定した．
- イメージのビルドでは，ホストのGoのモジュールのキャッシュを`python3 -m http.server 8099`で配り，`DOCKER_BUILD_ARGS="--network host --build-arg GOPROXY=http://127.0.0.1:8099"`を渡す．
- コンテナが作り直されると，`/etc/hosts`の追記，`dockerd`，モジュールプロキシ(`http.server 8099`)が消える．作業を再開するときに起動し直す．
- `gate:eol`はendoflife.dateに問い合わせるので，ネットワークが要る．
- actは，ホストの`GITHUB_TOKEN`と`GH_TOKEN`を外し，プロキシとCAを渡して動かす．
  `env -u GITHUB_TOKEN -u GH_TOKEN act ... -P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-24.04 --env HTTPS_PROXY=$HTTPS_PROXY --env HTTP_PROXY=$HTTP_PROXY --env NO_PROXY=localhost,127.0.0.1 --env SSL_CERT_FILE=/ccr/ca-bundle.crt --env NODE_EXTRA_CA_CERTS=/ccr/ca-bundle.crt --env GIT_SSL_CAINFO=/ccr/ca-bundle.crt --env REQUESTS_CA_BUNDLE=/ccr/ca-bundle.crt --container-options "-v /root/.ccr:/ccr:ro"`
- `000000000000.localhost`を`/etc/hosts`に足した．MiniStackは`docker run --network chnet --network-alias 000000000000.ministack -p 4566:4566 -v /var/run/docker.sock:/var/run/docker.sock`で起動する．
- PostgreSQLのクライアントは，この検証環境のホストにある(16系)．Dev ContainerにはDockerfileで入れる．
- AWS CLI(v1)は`AWS_REGION`を読まないので，`.devcontainer/aws-config`にリージョンを書いた．
- AWSの環境変数：`AWS_ENDPOINT_URL=http://localhost:4566`，`AWS_ACCESS_KEY_ID=test`，`AWS_SECRET_ACCESS_KEY=test`，`AWS_REGION=ap-northeast-1`，`AWS_CONFIG_FILE=.devcontainer/aws-config`．

## 次にやること

全15回のパッチと教材がそろった．作業用のリポジトリの最新のタグは`sol14`である．残りは「最後にやること」である．

第14回で決めたこと：

- 運用項目のスキーマ：仕組みに`task`(人が動かすmiseのタスク．`ops:verify`が`app/mise.toml`にあるかを確かめる)を足す．`records`に`match`(ログを絞る項目と値．例：`{ job: backup }`)を足し，`ops:verify`が項目の有無を確かめる．
- `mise run ops:report --month YYYY-MM`：運用項目ごとに，`match`のある問いはCloudWatch Logsの記録を，`gates.yml`のジョブは証跡の保管場所の`<日付>/<コミット>/<ジョブ>/`を，その月に探す．定常の項目で見つからなければ「証跡なし」．あわせて，期限を過ぎた脆弱性，30日以内に期限の来る例外，Scorecardの点(`out/evidence/scorecard/results.json`があれば)をまとめ，`out/report-YYYY-MM.md`に書く．
- 運用テストは，2099-01の日付で証跡を置き，`dependency-vulnerabilities`が見つかり，`actions-pinning`が「証跡なし」になることを確かめる．
- 解答では，証跡のないゲート(exceptions，pinning，eol，iam)が結果のJSONを証跡に残し，`gates.yml`でartifactにする．
- `mise run ops:checksheet checksheets/sample.yaml`：質問の`requirements`と運用項目の`requirements`を突き合わせ，回答の下書き(`out/checksheet-sample.md`)を作る．
- `scorecard.yml`(OpenSSF Scorecard)，運用項目`monthly-report`，`checksheet-response`，`supply-chain-scorecard`．`escalation`の`timing`は非定常にする．

第13回で決めたこと：

- 仕込み：`infra/ecs.tf`(クラスタ`prod`，タスク定義，サービス`api`．台数0)，`tools/ops/deploy.ts`(`mise run deploy --image <参照>`がタスク定義の新しい版を作り`update-service`する)，運用テスト`tests/ops/deploy.test.ts`，`.github/workflows/deploy.yml`(手動で動かし，タグのイメージを署名を確かめずにデプロイする．OIDCで`github-deploy`のロールを引き受ける)．
- 解答：`deploy.yml`はタグからダイジェストを得て，`release:verify`の後で`deploy`する．`deploy.ts`はダイジェストでない参照を拒む．
  `policy/`のconftestのルール(Rego v1)：ワークフローは最上位に`permissions`を書く，`mise run deploy`の前に同じジョブで`mise run release:verify`を動かす，`actor: 人`の運用項目には手順書がある．`gate:policy`(`tools/gates/policy.ts`)が`../.github/workflows/*.yml`と`ops/items.yaml`を検査する．
  リファクタリング：`ci.yml`のゲートのジョブを`gates.yml`(`workflow_call`，入力`working-directory`)へ移し，`ci.yml`は`uses: ./.github/workflows/gates.yml`で呼ぶ．`items.yaml`の仕組みの参照も`gates.yml`に替える．
- `aws-actions/configure-aws-credentials`は v6.3.0(`e1253824e5c10ff9df46874f81ed3ec929e19cfd`)に固定する．
- MiniStackのECSは，`register-task-definition`，`create-service`，`update-service`，`describe-services`が動くことを確かめた．
- AWSプロバイダは，ネットワークの設定のないECSのサービスを読むと落ちる(`flattenNetworkConfiguration`)．FargateとVPC，サブネット，セキュリティグループを足して避けた．
- `deploy.ts`はタスク定義の`requiresCompatibilities`などを引き継ぐため，`register-task-definition --cli-input-json`を使う．
- 第11回の修正：シークレットの最初の値を`terraform_data`の`local-exec`で書くようにし，`ex12`以降を作り直した．

第12回で決めたこと(後の回で使う)：

- 脆弱性の例外は，`package`(pURL)，`severity`，`found`，`status`，`justification`を持つ．`not_affected`はOpenVEXの文書(`openvex.json`)で，それ以外はTrivyの除外の設定で外す．
- `gate:sla`は`ALERTMANAGER_URL`があれば`VulnerabilityOverdue`(severity=page)を送り，`ops`のログに`escalation`を書く．
- 運用テストは`fileParallelism: false`で1つずつ動く．この検証環境では，MiniStackのほかにAlertmanagerを`docker run -p 9093:9093 -v app/ops/monitoring:/etc/alertmanager:ro`で起動する．

## 第9回から第14回でやること

各回の要件は`docs/ROADMAP.md`にある．作るときに決める点を書いておく．

- 第9回 バックアップとリストア：`pg_dump`の結果を証跡のバケットとは別のバケット(Object Lock)に保管する．RDSのスナップショットからの復元はMiniStackでは動かないので使わない．
  リストアの訓練は，新しいRDSを作って戻し，行数と内容を比べる運用テストにする．定期ジョブは`schedule`のワークフローで，`act schedule`で確かめる．`records`に，バックアップと訓練の成否を足す．
- 第10回 更新と定期的な再検査：`renovate.json`(actionのハッシュも更新する)，保管済みのSBOMの再検査(`trivy sbom`)の定期ジョブ，`mise.toml`の道具の版とendoflife.dateのデータからEOLを検出する`gate:eol`．
  第2回の例外(golang-jwt)をここで直してもよい．
- 第11回 権限の管理(完成)：`gate:config`でIAMポリシーの`*`を検出する(Trivyで足りなければconftestを先取りせずTypeScriptで書く)．GitHub Actions向けのOIDCの信頼ポリシー(`sub`をリポジトリとブランチに絞る)．
  棚卸しの報告(IAMの認証情報レポート，CloudTrail，`gh api`のメンバー一覧)を証跡にする．データベースのパスワードの交換のスクリプトと手順書．
- 第12回 脆弱性のトリアージ：初めて検出された日を証跡の保管場所に記録し，基準値の重大度ごとの日数を過ぎたら`gate:sla`が失敗する．期限切れはAlertmanagerのAPIへアラートを送り，第7回の振り分けを使う．
  例外の一覧からTrivyの除外の設定を作る処理を，OpenVEXの文書を作る処理に置き換える(リファクタリング)．
- 第13回 組織への展開：conftestのルール(`policy/`)，ゲートのジョブを`.github/workflows/gates.yml`(reusable workflow)へ移す．デプロイのジョブ(MiniStackのECS)の前に`release:verify`で署名を検証する．
  仕込みは，署名を検証しない`deploy.yml`．
- 第14回 定期報告とチェックシート：`ops:report`(月次報告)と`ops:checksheet`(チェックシートの回答の下書き)，OpenSSF Scorecardのワークフロー．`docs/ROADMAP.md`の冒頭の使用例を，実際の出力に置き換える．

## 最後にやること

- [x] `COURSE.md`の「落とし穴」を見直す(第12〜14回で見つけたことを足した)．
- [x] すべての回で`mise run iterations:verify`を通す(第0〜14回が通った)．
- [x] 全体を`finalize-artifacts`の手順で読み直す．
