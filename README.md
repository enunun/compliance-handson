# SaaSのコンプライアンスを仕組みで満たすハンズオン

SaaSの開発と運用で求められるコンプライアンスのコントロールを，開発者が仕組みとして作る連続講座である．
問題を仕込んだ小さなSaaS(`app/`)に，変更管理，シークレット検出，SBOM，監視，ログ設計，バックアップなどの仕組みを1回に1つずつ作り込む．
全15回を終えると，自分のリポジトリの仕組みと証跡を見せて，セキュリティチェックシートや監査の質問に答えられる．

## 対象者

Git，CI，Dockerを日常的に使うSaaSの開発者を対象にする．
セキュリティやコンプライアンスの知識は前提にしない．

## 始め方

1. このリポジトリをforkするか，GitHubの「Use this template」で自分のリポジトリを作る．
2. 自分のリポジトリをcloneし，VSCodeで開いて「Reopen in Container」を実行する．
   Dev Containerには，講座で使う道具とVSCodeの拡張機能がすべて入っている．
3. 題材の検査が通ることを確かめる．

   ```console
   cd app
   mise run check
   ```

4. [第0回](docs/iterations/00.md)から始める．

## 各回の進め方

各回の始めに，その回で検出する問題を自分のコードに当てる．

```console
mise run iteration:start 1
```

遅れたときや行き詰まったときは，解答を当てて追いつく．

```console
mise run iteration:solution 1
```

## 資料

- [ロードマップ](docs/ROADMAP.md)：各回で作る仕組みと学ぶこと．
- 各回の演習の手順と解説：`docs/iterations/`．

| 回 | 演習 | 解説 |
| --- | --- | --- |
| 0 | [変更管理と運用項目の土台](docs/iterations/00.md) | [解説](docs/iterations/00-answer.md) |
