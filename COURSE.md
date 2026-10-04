# 講座計画：SaaSのコンプライアンスを仕組みで満たす

教材を作る人とエージェントのための計画書．
講座の前提，題材，設計書，開発環境と，各回の範囲を定める．
学習者向けの各回の詳細は`docs/ROADMAP.md`に書く．

## 対象者とゴール

対象者は，Git，CI，Dockerを日常的に使うが，セキュリティやコンプライアンスは専門ではないSaaS開発者である．

ゴールは，SaaSの開発と運用で求められるコントロールを，次の流れで自分のリポジトリへ作り込めるようになることである．

1. 要求：SOC 2，ISO/IEC 27001，NIST SSDF，SLSA，経済産業省のSBOM導入の手引などが，なぜそのコントロールを求めるか．
2. 仕組み：CIのゲート，アラート，定期ジョブなど，人の注意に頼らず要求を満たす仕組み．
3. 証跡：仕組みが動いたことを，監査や顧客に示せる記録．
4. 例外の管理：誤検知や受け入れたリスクを，期限と承認つきで扱うこと．

修了した学習者は，顧客のセキュリティチェックシートや監査の質問に，自分のリポジトリの仕組みと証跡を見せて答えられる．

## 範囲

対象は，コードを書いてから本番で運用するまでに開発チームが仕組みとして作るものである．

- 開発時のコントロール：変更管理，シークレット，依存関係の脆弱性，SBOMとライセンス，SAST，コンテナとIaCの設定，成果物の署名とprovenance．
- サービスの運用：監視，監査ログとログ管理，バックアップとリストア，パッチと定期ジョブ，アカウントと利用者の管理．
- 運用の管理：対応期限などの基準，検出結果のトリアージとエスカレーション，定期報告．

運用の項目と分類は，近藤誠司『運用設計の教科書【改訂新版】』(技術評論社，2023年)の3分類(業務運用，基盤運用，運用管理)に従う．

## 講座の形式

- 連続講座で，全15回(第0回から第14回)．1回は90分で，自分のリポジトリへの適用までをその回の中で行う．
- 教材は日本語で書く．
- 教材に載せる出力は，実際に実行した結果を写す．`docs/ROADMAP.md`の使用例の出力は形を示す例で，各回を作るときに実際の出力に置き換える．
- 各回は，1つの運用項目を，仕組みと証跡まで作る．
- テスト駆動で進める．各回の最初に，「だめな入力では仕組みが失敗を報告し，正しい入力では通る」ことを確かめるテストを書き，その後に仕組みを作る．運用の項目では，アラートのルールのテスト，リストアしたデータの検証，監査ログの出力の検証などがこのテストにあたる．

## 題材

あらかじめ問題を仕込んだ，小さなSaaSのモノレポを学習者が育てる．
題材は教材とは別の，テンプレートリポジトリ`compliance-handson-app`に置く．

```text
.devcontainer/      学習者が作業するDev Container
apps/
  web/              TypeScriptのWebアプリ(pnpm workspace)
  api/              GoのAPI
packages/
  shared/           TypeScriptの共通ライブラリ
infra/              OpenTofu
ops/
  standards.yaml    基準値
  items.yaml        運用項目
  exceptions.yaml   例外
  schema/           上の3つのJSON Schema
  runbooks/         手順書
  policy.md         運用方針
tools/ops/          YAMLの検証，運用項目一覧の生成
tests/gates/        ゲートのテストと，だめな入力のfixture
tests/ops/          運用テスト
policy/             conftestのルール
.github/
  workflows/        ゲートと定期ジョブ
  CODEOWNERS
mise.toml           道具の版とタスク
```

仕込む問題の例は，git履歴に残った認証情報，脆弱なバージョンの依存関係，rootで動くDockerfile，公開設定のS3バケットである．
CIはGitHub Actionsで動かす．

## 演習と解答

題材リポジトリのタグで，各回の開始時点と解答を示す．

- `iteration-N-exercise`：第N回の開始時点．第N-1回の解答に，その回で検出する問題を仕込んだもの．
- `iteration-N-solution`：第N回の解答．

学習者は，テンプレートから自分のGitHubリポジトリを作り，ブランチ保護やActionsを自分で設定する．
行き詰まったら，`git diff iteration-N-solution`で解答との差分を見る．
各回の解説は，教材リポジトリ(このリポジトリ)の`docs/iteration-N.md`に書く．

## 各回の範囲

| 回 | 作る仕組み | 主な道具 | 『運用設計の教科書』との対応 |
| --- | --- | --- | --- |
| 0 | 題材とCIの土台，変更管理(ブランチ保護，必須チェック，CODEOWNERS)，運用項目のYAMLと一覧の生成 | GitHub | 運用項目一覧(3章) |
| 1 | シークレット検出(pre-commit，CI，git履歴)，漏洩時の手順書 | gitleaks | 運用情報統制(5.3節) |
| 2 | 依存関係の脆弱性検査(npm，Go)，期限付きの例外とその申請 | Trivy | パッチ運用(4.2節)，サポートデスク運用(3.3節) |
| 3 | SBOMの生成(コンポーネントごと)，ライセンスポリシー，証跡の保管 | Trivy，S3 | ログ管理(4.6節) |
| 4 | SASTと誤検知の扱い | Semgrep | |
| 5 | コンテナイメージとOpenTofuの設定検査 | Trivy | |
| 6 | Actionsのハッシュ固定，成果物の署名とprovenance | cosign，artifact attestations | |
| 7 | 監視とアラート | Prometheus，Alertmanager，Grafana | 監視運用(4.5節) |
| 8 | 監査ログと保管期間 | CloudWatch Logs | ログ管理(4.6節) |
| 9 | バックアップとリストア，リストアの訓練 | pg_dump，S3，RDS | バックアップ/リストア運用(4.4節) |
| 10 | 依存関係とベースイメージの更新，定期的な再検査，EOLの管理 | Renovate，Trivy | パッチ運用(4.2節)，ジョブ/スクリプト運用(4.3節)，保守契約管理(4.8節) |
| 11 | クラウドとGitHubの権限の検査と棚卸し，シークレットのローテーション | Trivy，IAMの認証情報レポート，CloudTrail | 運用アカウント管理(4.7節)，利用者の管理(3.2節) |
| 12 | 対応期限(SLA)，VEXでのトリアージ，エスカレーション | OpenVEX | 運用維持管理(5.2節)，運用情報統制(5.3節) |
| 13 | 共通ポリシーとreusable workflowによる展開，デプロイ前の署名検証 | conftest | 運用維持管理(5.2節) |
| 14 | 証跡の集約，月次報告，セキュリティチェックシートへの回答 | OpenSSF Scorecard | 定期報告(5.4節) |

## 設計書

学習者は，毎回次の3つを更新する．

| 設計書 | 中身 |
| --- | --- |
| 運用項目と基準のYAML(`ops/`) | `standards.yaml`に基準値(対応期限，証跡の保管期間，例外の最長期間など)，`items.yaml`に運用項目，`exceptions.yaml`に例外を書く．各項目は，3分類のどれか，要求，仕組み，証跡，タイミング(定常か非定常か)，実施者(自動か人か)を持つ． |
| 手順書(`ops/runbooks/`) | 人が行う非定常の項目の手順．例外の申請，シークレットが漏れたときの対応，リストアなど． |
| 運用方針(`ops/policy.md`) | 体制と役割，エスカレーションの経路．YAMLで表せないことだけを書く． |

基準値と例外は，CODEOWNERSでセキュリティ責任者の承認を要するようにし，運用項目は開発チームが承認する．
第13回では，`standards.yaml`を組織で共通のリポジトリへ移す．

基準値はYAMLにだけ書き，ゲートはYAMLを読んで判定する．
例えば，期限を過ぎた例外が残っているとCIが失敗する．
人が読む運用項目一覧(Markdown)はYAMLから生成し，手で編集しない．
運用項目に書いた仕組みと，ワークフローのジョブや定期ジョブが一致するかを，スクリプトで突き合わせて検証のコマンドに含める．

## 開発環境

学習者は，講座のDev Containerの中で作業する．

- 言語とパッケージ管理：Node.jsとpnpm，Go．
- IaC：OpenTofu．
- 本番の環境：AWSのAPIはMiniStackで模す．監視はPrometheus，Alertmanager，Grafanaで行う．どちらもDev Containerの中でコンテナとして動かす．
- 検査の道具：Trivy，gitleaks，Semgrep，cosign，conftest．
- CI：GitHub Actions．MiniStackはジョブのサービスコンテナとして起動し，`tofu apply`とアプリの結合テストを流す．
- ワークフローの手元での確認：nektos/act．`mise run ci`で実行する．OIDCトークンとGitHub本体の機能(ブランチ保護，必須チェック，CODEOWNERS，Renovate)はactでは動かないので，それらを扱う第0，6，10，11回はGitHubにpushして確かめる．
- スクリプトとゲートのテスト：TypeScriptとVitestで書く．YAMLの検証にはAjvを使う．アラートのルールはpromtoolでテストする．

道具の版は，LTSや安定版を優先して選ぶ．
CLIは題材リポジトリの`mise.toml`と`mise.lock`で，コンテナはイメージのダイジェストで固定する．

### 検証のコマンド

各ゲートは`gate:secrets`や`gate:sca`のようなmiseのタスクとして書き，CIのジョブはそのタスクを呼んで結果を証跡として保存する．
ゲートの定義はmiseのタスクにだけあるので，手元とCIで同じ検査が走る．

`mise run check`は次をまとめて実行する．

- `lint`：コードと文書のリント．
- `test`：アプリの単体テスト．
- `test:gates`：ゲートのテスト．だめな入力のfixtureでゲートが失敗し，正しい入力で通ることを確かめる．
- `ops:verify`：YAMLのスキーマ検証と，運用項目に書いた仕組みとワークフローのジョブの突き合わせ．
- `gate:*`：その回までに作ったすべてのゲート．

教材リポジトリでは，`mise run check`がtextlintとmarkdownlintで文書を検査する．

## 落とし穴

### MiniStack(1.5.21で確認)

- S3のObject Lock(COMPLIANCEモード)は機能する．ロック中の版を削除すると`AccessDenied`になる．ライフサイクルも設定できる．
- CloudWatchのアラームは評価され，SNSからSQSへ通知が届く．
- CloudWatch Logsの保管期間とJSONのフィルタ検索は機能する．メトリクスフィルタからはメトリクスが出なかった．
- CloudTrailは，`CLOUDTRAIL_RECORDING=1`で起動すると`lookup_events`で操作を返す．
- RDSは実際の`postgres:15-alpine`コンテナを起動する．スナップショットからの復元は，状態が`available`になってもコンテナを作らない．バックアップは`pg_dump`とS3で行う．
- IAMのポリシーは既定では評価されない．AssumeRoleWithWebIdentityは，検証できないトークンでも認証情報を返す．権限は静的な検査で確かめる．
- Secrets Managerのローテーションは設定できるが，実行されない(READMEによる)．
- RDSやECSのコンテナを起動するには，Python版では`ministack[full]`(`cryptography`を含む)と，Dockerのソケットが要る．
- ECSのRunTaskとOpenTofuとの組み合わせは，まだ確認していない．Dev Containerを作る段階で最初に確かめる．
- 動作が版ごとに変わりやすいので，版を固定する．

### nektos/act(0.2.89で確認)

- 次は動く．
  - サービスコンテナ．
  - `actions/upload-artifact@v4`．`--artifact-server-path`を指定する．
  - ローカルのreusable workflow．
  - `act schedule`による定期ジョブの起動．
- サービスコンテナの作業ディレクトリをリポジトリのパスに上書きするので，MiniStackが起動に失敗する．
  対策として，サービスに`options: --workdir /opt/ministack`を書く．
  イメージの既定と同じ値なので，GitHub上でも害はない．
- ジョブはホストのネットワークで動くので，サービスには`localhost:4566`で届く．
- `ACTIONS_ID_TOKEN_REQUEST_URL`が設定されないので，OIDCトークンを使う処理は動かない．
- ホストの`GITHUB_TOKEN`が無効だと，actionの取得で認証エラーになる．
- ランナーのイメージ(`ghcr.io/catthehacker/ubuntu:act-24.04`)が大きいので，Dev Containerのディスク容量に注意する．
