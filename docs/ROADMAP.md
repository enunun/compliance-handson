# ロードマップ

この講座では，問題を仕込んだ小さなSaaSのモノレポに，コンプライアンスの仕組みを1回に1つずつ作り込む．
全15回を終えると，自分のリポジトリの仕組みと証跡を見せて，セキュリティチェックシートや監査の質問に答えられる．

## 完成したときの姿

第14回を終えたリポジトリでは，次のことができる．

すべての仕組みを手元で検査する．
同じ検査がCIでも走り，結果が証跡として保管される．

```console
$ mise run check
[lint] ok
[test] ok
[test:gates] 14 gates, 31 tests passed
[ops:verify] 18 items, all mechanisms found in workflows
[gate:secrets] no leaks found
[gate:sca] 0 findings over threshold (2 covered by exceptions)
[gate:exceptions] 2 active, 0 expired
...
```

月次報告を作る．
運用項目ごとに，仕組みが動いた記録と，未対応の脆弱性，期限が近い例外をまとめる．

```console
$ mise run ops:report -- --month 2026-11
wrote out/report-2026-11.md
  items: 18 (evidence found: 18)
  vulnerabilities over SLA: 0
  exceptions expiring within 30 days: 1 (EXC-003)
  OpenSSF Scorecard: 8.1
```

セキュリティチェックシートの各質問に，運用項目と証跡を対応づけた回答の下書きを作る．

```console
$ mise run ops:checksheet -- checksheets/sample.yaml
Q12 脆弱性管理を行っていますか → はい
    運用項目: dependency-vulnerabilities, vulnerability-triage
    証跡: s3://evidence/sca/2026-11/, out/report-2026-11.md
```

## 各回の進め方

各回は，同じ順で進める．

1. 要求を読む．その回のコントロールを，どの規格がなぜ求めるかを確かめる．
2. 運用項目を書く．`ops/items.yaml`に項目を足し，必要なら`ops/standards.yaml`に基準値を足す．第8回からは，その項目がログで答える問いも`records`に書き，`ops/logs.yaml`を更新する．
3. テストを書く．だめな入力で仕組みが失敗し，正しい入力で通ることを確かめるテストを先に書き，失敗することを見る．
4. 仕組みを作る．miseのタスクとCIのジョブを作り，テストを通す．
5. 証跡を残す．仕組みの結果を，決めた場所に保管する．
6. 手順書と運用方針を直す．人の手で行う作業があれば手順書を書く．
7. 自分のリポジトリに適用する．`mise run ci`でワークフローを確かめてから，GitHubにpushする．

各回の始めに`mise run iteration:start N`で，その回で検出する問題を自分のコードに当てる．
遅れたときや行き詰まったときは，`mise run iteration:solution N`で解答を当てる．

ファイルの場所は，`.github/`を除き`app/`からの相対パスで示す．

## テストの分け方

| 種類 | 確かめること | 置き場所 | 実行するタスク |
| --- | --- | --- | --- |
| 単体テスト | アプリと`tools/ops/`のスクリプトの振る舞い | 各パッケージ | `test` |
| ゲートのテスト | だめな入力のfixtureでゲートが失敗し，正しい入力で通ること | `tests/gates/` | `test:gates` |
| 運用テスト | アラートのルール，ログの出力，リストアの結果など，運用の仕組みが期待どおり動くこと | `tests/ops/` | `test:ops` |

ゲートのテストと運用テストは，ゲートや仕組みを作る前に書く．

## 各回の一覧

| 回 | 作る仕組み | 学ぶこと |
| --- | --- | --- |
| 0 | 変更管理と運用項目の土台 | ブランチ保護，CODEOWNERS，運用項目のYAML，ゲートのテストの書き方，MiniStack，act |
| 1 | シークレット検出 | gitleaks，pre-commit，漏洩時の対応 |
| 2 | 依存関係の脆弱性検査と例外 | Trivy，CVSSと重大度，期限付きの例外 |
| 3 | SBOMとライセンス，証跡の保管 | CycloneDX，ライセンスポリシー，S3のObject Lock，OpenTofu |
| 4 | SAST | Semgrep，誤検知の扱い |
| 5 | コンテナとIaCの設定検査 | イメージの検査，設定の誤りの検出 |
| 6 | ワークフローと成果物の完全性 | ハッシュ固定，cosign，SLSA provenance，OIDC |
| 7 | 監視とアラート | Prometheus，Alertmanager，Grafana，promtool |
| 8 | ログ設計 | 運用からのログの導出，構造化ログ，監査ログ，CloudWatch Logs |
| 9 | バックアップとリストア | pg_dump，RPO，リストアの訓練，定期ジョブ |
| 10 | 更新と定期的な再検査 | Renovate，SBOMの再検査，EOL |
| 11 | 権限の管理 | 最小権限，棚卸し，ローテーション |
| 12 | 脆弱性のトリアージ | 対応期限，OpenVEX，エスカレーション |
| 13 | 組織への展開 | conftest，reusable workflow，デプロイ前の検証 |
| 14 | 定期報告とチェックシート | 証跡の集約，OpenSSF Scorecard |

## 第0回 変更管理と運用項目の土台

- 要求：本番に入る変更は，レビューと検査を通ったものだけにする．誰が何を承認したかを示せるようにする(SOC 2 CC8.1，ISO/IEC 27001 A.8.32)．
- 仕込む問題：レビューなしで`main`へ直接pushできる．
- 使用例：`mise run up`でMiniStackを起動し，`tofu apply`で環境を作ってアプリを動かす．
  `mise run check`と`mise run ci`が通る．
- 追加するもの：
  - `ops/standards.yaml`：`change.required_approvals`．
  - `ops/items.yaml`：`change-management`．
  - `ops/schema/`：基準値と運用項目のJSON Schema．
  - `tools/ops/verify.ts`：YAMLのスキーマ検証と，運用項目の仕組みとワークフローのジョブの突き合わせ．
  - `tools/ops/render.ts`：運用項目一覧(Markdown)の生成．
  - `.github/workflows/ci.yml`：`mise run check`を呼ぶジョブ．
  - `.github/CODEOWNERS`：`ops/standards.yaml`をセキュリティ責任者の承認対象にする．
  - `mise run ops:evidence:change`：ブランチ保護の設定をGitHubのAPIで取得し，JSONで保存する．
- ゲートのテスト：存在しないジョブを仕組みに書いた運用項目で，`ops:verify`が失敗する．スキーマに合わない基準値で失敗する．
- 証跡：ブランチ保護の設定(JSON)と，プルリクエストのレビュー記録．
- 設計書の更新：YAMLの初版を書く．運用方針に体制と役割を書く．手順書に緊急変更の手順を書く．
- 既存のテストへの影響：なし(最初の回)．
- 学習者が行う道具の操作：このリポジトリをforkするか，テンプレートとして使って自分のリポジトリを作る．
  ブランチ保護と必須チェックを設定する．`mise run ci`でactを動かす．

## 第1回 シークレット検出

- 要求：認証情報をリポジトリに入れない．入ってしまったら無効化と交換をすぐに行う(ISO/IEC 27001 A.5.17，NIST SSDF PS.1)．
- 仕込む問題：メール配信サービスのAPIキーが，コードに直書きされてコミットされている．
- 使用例：キーを含むコミットを，pre-commitで止める．`mise run gate:secrets`で，履歴全体を検査する．
- 追加するもの：
  - `ops/items.yaml`：`secret-detection`．
  - `mise run gate:secrets`と`mise run gate:secrets:staged`：gitleaksでの検査．
  - `.gitleaks.toml`：fixtureの置き場所と，無効化済みのキーを許可する．
  - `lefthook.yml`：コミット時に`gate:secrets:staged`を実行する．
  - `.github/workflows/ci.yml`：`secrets`ジョブ．
- リファクタリング：ゲートが増えるので，`mise run check`を，コードの検査(`check:code`)とゲート(`gates`)に分ける．
  CIの`check`ジョブは`check:code`を実行し，ゲートはゲートごとのジョブで実行する．
- ゲートのテスト：キーを含むコミットやステージした変更で失敗し，キーを含まないもので通る．出力と証跡にキーそのものを書かない．
- 証跡：gitleaksの結果(JSON，キーは伏せ字)をCIのartifactとして保存する．
- 設計書の更新：手順書に漏洩時の対応(無効化，交換，利用の確認，記録)を書く．運用方針に連絡先を足す．
- 既存のテストへの影響：なし．
- 学習者が行う道具の操作：仕込まれたキーをコミットする．キーを無効化した前提で，履歴に残った検出を許可リストに記録する．

## 第2回 依存関係の脆弱性検査と例外

- 要求：既知の脆弱性を持つ依存関係を本番に入れない．受け入れるリスクは期限と承認者を決めて記録する(ISO/IEC 27001 A.8.8，SOC 2 CC7.1)．
- 仕込む問題：`apps/web`と`apps/api`に，重大度が高い脆弱性を持つ依存関係がある．1件は到達しないコードにあり，例外として扱う．
- 使用例：`mise run gate:sca`がnpmとGoの依存関係を検査する．例外に載った脆弱性は失敗にしない．期限を過ぎた例外が残ると失敗する．
- 追加するもの：
  - `ops/standards.yaml`：`vulnerability.fail_severity`と`exception.max_days`．
  - `ops/exceptions.yaml`：例外の一覧．対象，理由，承認者，期限を持つ．
  - `ops/items.yaml`：`dependency-vulnerabilities`と`exception-requests`．
  - `mise run gate:sca`：Trivyでの検査．例外の一覧からTrivyの除外設定を生成して使う．
  - `mise run gate:exceptions`：期限切れと，最長期間を超える例外を検出する．
- リファクタリング：ゲートのテストと本物のゲートが同じ証跡のファイルに書き合わないように，証跡の置き場所を環境変数`EVIDENCE_DIR`で替えられるようにする．
- ゲートのテスト：脆弱な依存関係のfixtureで失敗し，例外に載せると通る．期限切れの例外で`gate:exceptions`が失敗する．
- 証跡：Trivyの結果(JSON)と，例外の一覧の変更履歴(承認つきのプルリクエスト)．
- 設計書の更新：手順書に例外の申請を書く．CODEOWNERSで`ops/exceptions.yaml`をセキュリティ責任者の承認対象にする．
- 既存のテストへの影響：`ops:verify`のfixtureの基準値に，脆弱性と例外の基準を足す．`gate:secrets`のテストは，証跡を一時的な場所から読む．
- 学習者が行う道具の操作：依存関係を更新する(`pnpm update`，`go get`)．例外を申請するプルリクエストを作る．

## 第3回 SBOMとライセンス，証跡の保管

- 要求：出荷するソフトウェアの構成部品を一覧にして渡せるようにする．
  ライセンスの義務を守る．
  証跡は消せない形で決めた期間だけ保管する(経済産業省のSBOM導入の手引，NIST SSDF PS.3)．
- 仕込む問題：`apps/web`に，GPL-2.0-or-laterのリッチテキストエディタ(TinyMCE 7)が足されている．
- 使用例：`mise run sbom`がコンポーネントごとにCycloneDXのSBOMを作る．`mise run gate:license`が許可されたライセンスかを検査する．
  CIの`evidence`ジョブが，ゲートのジョブの証跡を集めて証跡の保管場所(S3)に送る．
- 追加するもの：
  - `ops/standards.yaml`：`license.allowed`と`evidence.retention_days`．
  - `ops/items.yaml`：`sbom`，`license-policy`，`evidence-retention`．
  - `tools/gates/license.ts`と`mise run gate:license`．
  - `infra/evidence.tf`：Object LockとライフサイクルつきのS3バケット．保管期間は`ops/standards.yaml`から読む．
  - `mise run evidence:upload`：証跡を日付とコミットごとにS3へ送る．
  - `mise run test:ops`：運用テストを実行する．
  - `.github/workflows/ci.yml`：`sbom`ジョブと，MiniStackをサービスコンテナで動かす`evidence`ジョブ．
- ゲートのテスト：許可していないライセンスや，ライセンスが分からない部品を含むSBOMで失敗する．
- 運用テスト：証跡の保管場所が基準値の日数だけ消せない設定であること，保管した証跡を版を指定しても消せないことを確かめる．
- 証跡：SBOM(CycloneDX)をS3に保管する．以降の回の証跡もここに保管する．
- 設計書の更新：運用方針に証跡の保管場所と期間を書く．
- 既存のテストへの影響：`ops:verify`のfixtureの基準値に，ライセンスと証跡の基準を足す．
  第1回と第2回のジョブは，`out/evidence/`全体をartifactとして残す形に変わる．
- 学習者が行う道具の操作：`tofu plan`と`tofu apply`で変更を確かめる．AWS CLIで，保管した証跡を一覧する．

## 第4回 SAST

- 要求：よく知られた種類の脆弱性(インジェクションなど)を，コードの段階で見つける(ISO/IEC 27001 A.8.28，NIST SSDF PW.7)．
- 仕込む問題：APIのメモの検索にSQLインジェクション，Web画面の改行の表示にXSSがある．
  キャッシュの判定に使うSHA-1と，TLSなしのHTTPサーバーも検出されるが，これらは誤検知である．
- 使用例：`mise run gate:sast`が，レジストリのルールセットと自前のルールでTypeScriptとGoを検査する．誤検知は例外の一覧に載せる．
- 追加するもの：
  - `ops/items.yaml`：`sast`．
  - `.semgrep/rules.yml`：題材が使うライブラリ(pgx，Hono)に合わせた自前のルール．
  - `tools/gates/sast.ts`と`mise run gate:sast`：Semgrepでの検査．第2回と同じ例外の一覧を読む．
  - `ops/schema/exceptions.schema.json`：例外の場所を絞る`path`を足す．
  - `.semgrepignore`：Semgrepの既定の除外(`tests/`)を置き換える．
- ゲートのテスト：SQLを文字列で組み立てるコードと，`raw()`に変数を渡すコードで失敗し，直したコードで通る．例外に載った検出は外す．
- 証跡：Semgrepの結果(JSONとSARIF)をS3に保管する．
- 設計書の更新：手順書の例外の申請に，誤検知と判断する基準を足す．
- 既存のテストへの影響：APIのテスト用のStoreに，検索の関数が加わる．
- 学習者が行う道具の操作：検出されたコードを直す．誤検知を例外として申請する．

## 第5回 コンテナとIaCの設定検査

- 要求：コンテナとクラウドの設定の誤りを，デプロイの前に見つける(ISO/IEC 27001 A.8.9)．
- 仕込む問題：APIのDockerfileが，サポートの終わったOS(alpine 3.20)の上でrootのまま動く．
  `infra/`に，誰でも読める公開のポリシーを付けたバケットがある．
- 使用例：`mise run gate:image`がビルドしたイメージを検査する．`mise run gate:config`がDockerfileとOpenTofuのコードを検査する．
  これまでの回で作ったバケットとデータベースの設定の誤りも見つかる．
- 追加するもの：
  - `ops/items.yaml`：`container-image`と`iac-config`．
  - `apps/api/Dockerfile`：distrolessのイメージの上で，rootではない利用者で動かす．
  - `tools/gates/image.ts`と`mise run gate:image`：脆弱性に加え，サポートの終わったOSでも失敗する．
  - `tools/gates/config.ts`と`mise run gate:config`．
  - `infra/kms.tf`：バケットを暗号化する鍵．バケットにはパブリックアクセスのブロックと暗号化を足し，データベースも暗号化する．
- リファクタリング：Trivyを呼ぶ処理を`tools/gates/trivy.ts`にまとめ，`gate:sca`もそれを使う．
  例外は，脆弱性と設定の誤りの両方の除外に使う．
- ゲートのテスト：rootで動くDockerfileと公開のバケットで`gate:config`が失敗し，直した設定で通る．
  サポートの終わったOSのイメージで`gate:image`が失敗し，サポート中のイメージで通る．
- 証跡：検査結果をS3に保管する．
- 設計書の更新：なし．
- 既存のテストへの影響：なし．
- 学習者が行う道具の操作：イメージをビルドする．Dockerfileに実行する利用者を足す．バケットの公開設定を外す．

## 第6回 ワークフローと成果物の完全性

- 要求：CIで使う部品のすり替えを防ぐ．出荷したイメージが，このリポジトリのCIで作られたことを検証できるようにする(SLSA，NIST SSDF PS.2)．
- 仕込む問題：第0回から，ワークフローの`uses:`をタグで参照している．この回では新たに仕込まない．
- 使用例：`mise run gate:pinning`が，ハッシュで固定されていない`uses:`を検出する．
  `v`で始まるタグをpushすると，`release.yml`がイメージをGHCRに置き，キーレスの署名とprovenanceを付けて検証する．
- 追加するもの：
  - `ops/items.yaml`：`actions-pinning`と`artifact-signing`．
  - `tools/gates/pinning.ts`と`mise run gate:pinning`．
  - `.github/workflows/release.yml`：イメージのpush，cosignでのキーレスの署名，`actions/attest-build-provenance`，検証．
  - `tools/gates/verify-image.sh`と`mise run release:verify`：署名とprovenanceの検証．第13回のデプロイの前の検証でも使う．
- ゲートのテスト：タグやブランチを参照するワークフローなら失敗し，ハッシュに固定した参照と同じリポジトリの中の参照だけなら通る．
- 証跡：署名とprovenance．検証の結果(JSON)をartifactに残す．
- 設計書の更新：手順書に，イメージの署名とprovenanceを確かめる手順を書く．
- 既存のテストへの影響：なし．
- 学習者が行う道具の操作：OIDCのトークンはactでは発行されないので，この回はGitHubにタグをpushして確かめる．

## 第7回 監視とアラート

- 要求：サービスの異常に気づき，決めた担当者へ知らせる(ISO/IEC 27001 A.8.16)．
- 仕込む問題：APIのエラー率が上がっても，誰にも通知されない．この回では新たに仕込まない．
- 使用例：Dev ContainerでPrometheus，Alertmanager，Grafanaも起動する．APIのエラー率がしきい値を超えると，Alertmanagerが重大度に応じた通知先へ知らせる．
- 追加するもの：
  - `ops/standards.yaml`：`monitoring.error_rate_threshold`と`monitoring.error_rate_for`．
  - `ops/items.yaml`：`service-monitoring`．仕組みはアラートのルールである．
  - `apps/api`：`/metrics`のエンドポイントと，ルートとステータスコードごとのリクエストの数．
  - `ops/monitoring/`：Prometheusの設定，アラートのルール(基準値から`ops:render`で生成する)，Alertmanagerの通知先，Grafanaのダッシュボード．
  - `.devcontainer/compose.yml`：Prometheus，Alertmanager，Grafanaのサービス．
  - `mise run test:alerts`：監視の設定の検査，アラートのルールのテスト，通知先の振り分けのテスト．
- 運用テスト：`promtool test rules`で，エラー率が上がるとアラートが発火し，下がると止むこと，低いエラー率では発火しないことを確かめる．
  `amtool config routes test`で，重大度ごとの通知先を確かめる．
- 証跡：アラートのルールとそのテストの結果．
- 設計書の更新：手順書にアラートへの対応を書く．運用方針に通知先を足す．
- 既存のテストへの影響：運用項目の仕組みに，ワークフローのジョブに加えてアラートを書けるようにスキーマを変える．`ops:verify`は，アラートのルールが基準値と合うかも確かめる．
- 学習者が行う道具の操作：Dev Containerを作り直す(Rebuild Container)．Grafanaでダッシュボードを確かめる．

## 第8回 ログ設計

- 要求：運用で答えるべき問い(誰がいつ何をしたか，障害の原因はどのリクエストか，不正な操作はないか)の答えを，必要な期間だけ残す．
  根拠はISO/IEC 27001 A.8.15，A.8.16とSOC 2 CC7.2である．
- 仕込む問題：APIのログは形式の決まらないテキストで，リクエストを追えない．ログインの失敗のログにパスワードを，成功のログにトークンを書いている．権限の変更は記録されない．
- 設計の進め方：ログの構造は，運用から次の順に導く．
  1. 運用項目ごとに，ログで答える問いを`records`に書く．誰が，何を，いつまでに知りたいかを書く．
     例えば`service-monitoring`は「アラートの原因になったリクエストを，発生から1時間以内に特定する」，`audit-logging`は「権限を変えた人と時刻を，1年後にも示す」である．
  2. 問いをまとめて，ログの種類(アプリログ，監査ログ)を分ける．種類ごとに，問いの答えに要る項目，保管期間，閲覧できる人，記録してはならない項目を決める．
  3. 保管期間は，その種類を使う運用のうち最も長いものに合わせ，基準値に書く．
  4. 実装と基盤は，決めた種類と項目に従う．
- 使用例：APIがJSONの構造化ログを出し，すべての行にリクエストIDが付く．ログインと権限の変更は，監査ログとして別のロググループに送られる．
  `mise run ops:records`で，運用項目とログの対応表を生成する．
- 追加するもの：
  - `ops/items.yaml`：既存の運用項目に`records`を書き足す．`log-design`と`audit-logging`を足す．
  - `ops/logs.yaml`：ログの種類ごとの項目，保管期間，閲覧できる人，記録してはならない項目．
  - `ops/standards.yaml`：`logs.retention_days`(種類ごと)．
  - `ops/schema/`：`logs.yaml`のスキーマ．ログの各行を検証するJSON Schemaは`logs.yaml`から生成する．
  - `apps/api`：`log/slog`の構造化ログ，リクエストID(応答の`X-Request-Id`)，監査イベント，ログの種類ごとのロググループへ送るハンドラ．
  - `infra/logging.tf`：種類ごとのロググループ．保管期間は`ops/standards.yaml`から読む．
  - `ops:verify`：`records`が参照するログの種類と項目が`ops/logs.yaml`にあること，保管期間が問いの遡る期間以上であることを確かめる．
- リファクタリング：APIのログ出力を，`ops/logs.yaml`のとおりの構造化ログに置き換える．
- ゲートのテスト：`ops/logs.yaml`にない項目を参照する`records`と，問いの遡る期間より短い保管期間で，`ops:verify`が失敗する．
- 運用テスト：APIでログインして権限を変え，出たログが生成したJSON Schemaに合うこと，パスワードとトークンを含まないこと，監査ログのロググループで検索できることを確かめる．
- 証跡：ロググループの設定と，運用項目とログの対応表．
- 設計書の更新：運用方針に，ログの種類ごとに閲覧できる人を書く．手順書のアラートへの対応に，リクエストIDでログを追う手順を足す．
- 既存のテストへの影響：`server.New`がロガーを受け取るので，APIの単体テストの作り方が変わる．`ops:verify`のfixtureの基準値に，ログの保管期間を足す．
- 学習者が行う道具の操作：CloudWatch Logsで，リクエストIDと監査イベントを検索する．

## 第9回 バックアップとリストア

- 要求：データを決めた間隔でバックアップし，戻せることを定期的に確かめる(ISO/IEC 27001 A.8.13)．
- 仕込む問題：データベースのバックアップがない．この回では新たに仕込まない．
- 使用例：定期ジョブが`pg_dump`の結果を，消せないバケットに保管する．`mise run ops:restore-test`が新しいデータベースに戻し，バックアップした時点の行数と比べる．
  作業の結果は，運用の作業の記録(ログの種類`ops`)に残る．
- 追加するもの：
  - `ops/standards.yaml`：`backup.interval_hours`(RPO)と`backup.retention_days`，ログの種類`ops`の保管期間．
  - `ops/items.yaml`：`backup`と`restore-test`．`records`に，バックアップと訓練が行われたことを示す問いを書く．
  - `ops/logs.yaml`：問いから導いたログの種類`ops`．
  - `infra/backup.tf`：Object Lockつきのバックアップのバケット．
  - `tools/ops/backup.ts`と`mise run ops:backup`．
  - `tools/ops/restore-test.ts`と`mise run ops:restore-test`．
  - `.github/workflows/backup.yml`：`schedule`で動くバックアップとリストアの訓練．
- 運用テスト：バックアップを取り，新しいデータベースに戻した行数が一致することを確かめる．
  保管場所が消せない設定であることと，作業の記録がログの設計に合うことも確かめる．
- 証跡：バックアップのファイルと，リストアの訓練の結果．
- 設計書の更新：手順書にリストアの手順を書く．
- 既存のテストへの影響：第8回のログの運用テストは，APIが書くログの種類(`api`と`audit`)だけを確かめる形に変わる．
- 学習者が行う道具の操作：Dev Containerを作り直す(PostgreSQLのクライアントが入る)．`act schedule`で定期ジョブを手元で動かす．

## 第10回 更新と定期的な再検査

- 要求：依存関係とベースイメージを最新に保つ．リリース後に公開された脆弱性を見つける．サポートが終わる部品を使い続けない(ISO/IEC 27001 A.8.8，A.8.19)．
- 仕込む問題：APIのDockerfileのビルド用のイメージが，サポートの終わったGo(1.25)である．
- 使用例：Renovateが更新のプルリクエストを作る．定期ジョブが，リリース済みのイメージを最新の脆弱性のデータで検査し直す．
  `mise run gate:eol`が，サポートが終わった，または終わりが近いランタイムとベースイメージを検出する．
- 追加するもの：
  - `ops/standards.yaml`：`eol.warn_days`．
  - `ops/items.yaml`：`dependency-updates`，`release-rescan`，`eol-tracking`．仕組みに，外部のサービスが読む設定ファイルも書けるようにする．
  - `renovate.json`：更新の方針．GitHub Actionsのハッシュも更新する．
  - `.github/workflows/rescan.yml`と`mise run rescan:image`：リリース済みのイメージの再検査．
  - `tools/gates/eol.ts`と`mise run gate:eol`：`mise.toml`とDockerfileの版を，endoflife.dateのデータと照らし合わせる．
- 依存関係の更新：第2回の例外(golang-jwt)を，直った版に上げて閉じる．
- ゲートのテスト：サポートが終わったベースイメージと，基準値の日数のうちにサポートが終わるランタイムで失敗する．
- 証跡：再検査の結果をartifactに残す．Renovateのプルリクエストの記録．
- 設計書の更新：運用方針に，更新を取り込む方針を書く．
- 既存のテストへの影響：`ops:verify`のfixtureの基準値に，サポートの終了の基準を足す．
- 学習者が行う道具の操作：Renovate(GitHubのアプリ)をリポジトリに入れる．GitHubの機能なので，この回はGitHubで確かめる．

## 第11回 権限の管理

- 要求：権限を必要最小限にし，定期的に棚卸しする．認証情報を定期的に交換する(ISO/IEC 27001 A.5.15，A.5.18，A.8.2)．
- 仕込む問題：`infra/`に，すべての操作を許すIAMポリシーがある．使われていないアクセスキーが残っている．
- 使用例：`mise run gate:config`がIAMポリシーの過剰な権限を検出する．
  `mise run ops:access-review`がIAMとGitHubの権限の棚卸しの報告を作る．
  `mise run ops:rotate-db-password`がデータベースのパスワードを交換する．
- 追加するもの：
  - `ops/standards.yaml`：`access.review_interval_days`と`credentials.max_age_days`．
  - `ops/items.yaml`：`least-privilege`，`access-review`，`credential-rotation`．`records`に，棚卸しの期間の権限の変更を，監査ログとCloudTrailから示せることを書く．
  - `infra/`：GitHub Actions向けのOIDCの信頼ポリシー(対象のリポジトリとブランチを絞る)．
- ゲートのテスト：すべての操作を許すポリシーと，対象を絞らない信頼ポリシーのfixtureで失敗する．
- 証跡：棚卸しの報告(IAMの認証情報レポート，CloudTrailの操作記録，GitHubのメンバー一覧)をS3に保管する．
- 設計書の更新：手順書に棚卸しと交換の手順を書く．運用方針に，権限を承認する人を書く．
- 既存のテストへの影響：なし．
- 学習者が行う道具の操作：GitHubのメンバーとチームの権限を確かめる．

## 第12回 脆弱性のトリアージ

- 要求：見つかった脆弱性を重大度ごとの期限内に直す．
  影響がない脆弱性はその根拠を示す．
  期限を過ぎたら上位者に知らせる(ISO/IEC 27001 A.8.8，SOC 2 CC7.4)．
- 仕込む問題：再検査で見つかった脆弱性が，期限を過ぎても残っている．
- 使用例：`mise run gate:sla`が，期限を過ぎた未対応の脆弱性を検出する．期限を過ぎるとAlertmanagerが上位者に通知する．影響のない脆弱性は，例外の一覧からOpenVEXの文書として出力される．
- 追加するもの：
  - `ops/standards.yaml`：`vulnerability.sla_days`(重大度ごとの日数)．
  - `ops/items.yaml`：`vulnerability-triage`と`escalation`．`records`に，検出から対応までの日数と，エスカレーションした相手と時刻を示せることを書く．
  - `ops/exceptions.yaml`：例外に，VEXの状態(`not_affected`など)と根拠を足す．
  - `mise run gate:sla`：初めて検出された日を記録し，基準値と比べる．
- リファクタリング：第2回の，例外の一覧からTrivyの除外設定を作る処理を，OpenVEXの文書を作る処理に置き換える．
- ゲートのテスト：期限を過ぎた検出結果のfixtureで失敗する．
- 証跡：OpenVEXの文書と，トリアージの結果をS3に保管する．
- 設計書の更新：運用方針にエスカレーションの経路を書く．手順書にトリアージの手順を書く．
- 既存のテストへの影響：第2回と第4回のゲートのテストで，例外の指定がOpenVEXの形に変わる．
- 学習者が行う道具の操作：検出された脆弱性をトリアージし，直すか例外にするかを決める．

## 第13回 組織への展開

- 要求：ほかのリポジトリでも同じ基準と検査を使えるようにする．検査を通っていないイメージはデプロイしない(NIST SSDF PO.1，SLSA)．
- 仕込む問題：デプロイのジョブが，署名を検証せずにイメージを使っている．
- 使用例：ゲートのジョブを`.github/workflows/gates.yml`のreusable workflowにまとめ，`ci.yml`はそれを呼ぶ．
  `mise run gate:policy`がconftestで組織の共通ルールを検査する．
  デプロイの前に署名を検証する．
- 追加するもの：
  - `ops/items.yaml`：`shared-policy`と`deploy-verification`．
  - `policy/`：conftestのルール．例えば，すべての運用項目に証跡があること，すべてのワークフローがゲートのreusable workflowを呼ぶこと．
  - `.github/workflows/gates.yml`．
  - `.github/workflows/deploy.yml`：イメージを使う前に署名を検証する．
- リファクタリング：`ci.yml`のゲートのジョブを`gates.yml`に移す．`ops/standards.yaml`を組織で共通に使う場所へ移す前提で，読み込む場所を1か所にまとめる．
- ゲートのテスト：証跡のない運用項目と，署名のないイメージのfixtureで失敗する．
- 証跡：デプロイ前の検証結果をS3に保管する．
- 設計書の更新：運用方針に，共通の基準を変えるときの承認の流れを書く．
- 既存のテストへの影響：ゲートのテストは変わらない．`ops:verify`の突き合わせの対象が`gates.yml`に変わる．
- 学習者が行う道具の操作：`mise run ci`でreusable workflowの呼び出しを確かめる．

## 第14回 定期報告とチェックシート

- 要求：運用の状況を定期的に報告する．顧客や監査の質問に，証跡を示して答える(SOC 2 CC2.2，ISO/IEC 27001 A.5.35)．
- 仕込む問題：なし．これまでの仕組みと証跡をまとめる．
- 使用例：「完成したときの姿」のとおり，`mise run ops:report`で月次報告を，`mise run ops:checksheet`でチェックシートの回答の下書きを作る．
- 追加するもの：
  - `ops/items.yaml`：`monthly-report`と`checksheet-response`．各運用項目の要求に，チェックシートの質問を対応づける．月次報告は，各運用項目の`records`が指すログと証跡から作る．
  - `tools/ops/report.ts`と`tools/ops/checksheet.ts`．
  - `.github/workflows/scorecard.yml`：OpenSSF Scorecard．
  - `checksheets/sample.yaml`：講座で用意するチェックシートの例．
- 運用テスト：証跡の欠けた運用項目があると，月次報告がそれを示すことを確かめる．
- 証跡：月次報告とScorecardの結果をS3に保管する．
- 設計書の更新：運用方針に，報告の宛先と頻度を書く．
- 既存のテストへの影響：なし．
- 学習者が行う道具の操作：自分のリポジトリでScorecardを動かす．チェックシートの回答を見直す．
