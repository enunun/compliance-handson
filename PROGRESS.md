# 教材づくりの進み具合

教材を作る人とエージェントのための作業の記録である．
区切りごとに更新する．計画は`COURSE.md`，各回の要件は`docs/ROADMAP.md`にある．

## 今の状況

| 回 | 状況 |
| --- | --- |
| 0〜9 | 完成．パッチ(`iterations/NN/`)，演習の手順(`docs/iterations/NN.md`)，解説(`NN-answer.md`)がそろい，`mise run iterations:verify`が通る． |
| 10〜14 | 未着手． |

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
- `000000000000.localhost`を`/etc/hosts`に足した．MiniStackは`docker run --network chnet --network-alias 000000000000.ministack -p 4566:4566 -v /var/run/docker.sock:/var/run/docker.sock`で起動する．
- PostgreSQLのクライアントは，この検証環境のホストにある(16系)．Dev ContainerにはDockerfileで入れる．
- AWS CLI(v1)は`AWS_REGION`を読まないので，`.devcontainer/aws-config`にリージョンを書いた．
- AWSの環境変数：`AWS_ENDPOINT_URL=http://localhost:4566`，`AWS_ACCESS_KEY_ID=test`，`AWS_SECRET_ACCESS_KEY=test`，`AWS_REGION=ap-northeast-1`，`AWS_CONFIG_FILE=.devcontainer/aws-config`．

## 次にやること

第10回(更新と定期的な再検査)から始める．作業用のリポジトリの最新のタグは`sol9`である．

## 第9回から第14回でやること

各回の要件は`docs/ROADMAP.md`にある．作るときに決める点を書いておく．

- 第9回 バックアップとリストア：`pg_dump`の結果を証跡のバケットとは別のバケット(Object Lock)に保管する．RDSのスナップショットからの復元はMiniStackでは動かないので使わない．
  リストアの訓練は，新しいRDSを作って戻し，行数と内容を比べる運用テストにする．定期ジョブは`schedule`のワークフローで，`act schedule`で確かめる．`records`に，バックアップと訓練の成否を足す．
- 第10回 更新と定期的な再検査：`renovate.json`(actionのハッシュも更新する)，保管済みのSBOMの再検査(`trivy sbom`)の定期ジョブ，`mise.toml`の道具の版とendoflife.dateのデータからEOLを検出する`gate:eol`．
  第2回の例外(golang-jwt)をここで直してもよい．
- 第11回 権限の管理：`gate:config`でIAMポリシーの`*`を検出する(Trivyで足りなければconftestを先取りせずTypeScriptで書く)．GitHub Actions向けのOIDCの信頼ポリシー(`sub`をリポジトリとブランチに絞る)．
  棚卸しの報告(IAMの認証情報レポート，CloudTrail，`gh api`のメンバー一覧)を証跡にする．データベースのパスワードの交換のスクリプトと手順書．
- 第12回 脆弱性のトリアージ：初めて検出された日を証跡の保管場所に記録し，基準値の重大度ごとの日数を過ぎたら`gate:sla`が失敗する．期限切れはAlertmanagerのAPIへアラートを送り，第7回の振り分けを使う．
  例外の一覧からTrivyの除外の設定を作る処理を，OpenVEXの文書を作る処理に置き換える(リファクタリング)．
- 第13回 組織への展開：conftestのルール(`policy/`)，ゲートのジョブを`.github/workflows/gates.yml`(reusable workflow)へ移す．デプロイのジョブ(MiniStackのECS)の前に`release:verify`で署名を検証する．
  仕込みは，署名を検証しない`deploy.yml`．
- 第14回 定期報告とチェックシート：`ops:report`(月次報告)と`ops:checksheet`(チェックシートの回答の下書き)，OpenSSF Scorecardのワークフロー．`docs/ROADMAP.md`の冒頭の使用例を，実際の出力に置き換える．

## 最後にやること

- `COURSE.md`の「落とし穴」を見直す．
- すべての回で`mise run iterations:verify`を通す．
- 全体を`finalize-artifacts`の手順で読み直す．
