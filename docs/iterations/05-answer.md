# 第5回 解説

解答のコードは`mise run iteration:solution 5`で当てられる．
差分は`iterations/05/solution.patch`にある．

## 5-1 準備

Dockerfileには`USER`がないので，APIはrootで動く．
エクスポートのバケットには，誰にでも`s3:GetObject`を許すポリシーが付いている．

## 5-2 学ぶこと

この回の要点は，動かす環境の設定も，コードと同じようにゲートで検査することである．

## 5-3 テストリスト

### ゲートのテスト(`tests/gates/config.test.ts`)

- [x] rootで動くDockerfileがあれば失敗し，`DS-0002`を表示する．
- [x] 誰でも読めるS3のバケットがあれば失敗し，`AWS-0087`を表示する．
- [x] 設定の誤りがなければ通る．

### ゲートのテスト(`tests/gates/image.test.ts`)

- [x] サポートが終わったOSのイメージなら失敗し，OSの名前を表示する．
- [x] サポート中で脆弱性のないイメージなら通る．

## 5-4 設計書

運用項目は2つとも基盤運用である．
仕組みは`ci.yml`の`image`ジョブと`config`ジョブである．

## 5-5 テスト駆動の実装

### Trivyの共通の処理

`gate:sca`，`gate:image`，`gate:config`は，どれも「基準値の重大度と例外でTrivyを動かし，結果をJSONで残す」．
この処理を`tools/gates/trivy.ts`の`runTrivy`にまとめた．

```ts
export function runTrivy(options: TrivyOptions): { status: number; report: string } {
  const severity = severitiesFrom(readStandards(options.ops).vulnerability.fail_severity);
  const ignoreFile = join(options.evidenceDir, "trivyignore.yaml");
  writeFileSync(ignoreFile, stringify(toTrivyIgnore(readExceptions(options.exceptions))));
  const report = join(options.evidenceDir, "trivy.json");
  const scan = spawnSync(
    "trivy",
    [...options.args, "--severity", severity.join(","), "--ignorefile", ignoreFile,
     "--exit-code", "1", "--quiet", "--format", "json", "--output", report],
    { stdio: "inherit" },
  );
  spawnSync("trivy", ["convert", "--quiet", "--format", "table", report], { stdio: "inherit" });
  return { status: scan.status ?? 1, report };
}
```

例外は，`vulnerabilities`と`misconfigurations`の両方に入れる．
どちらに当たるかは識別子で決まるので，両方に入れても誤って外すことはない．

### gate:image

`runTrivy`の後で，結果のJSONからOSのサポートを確かめる．

```ts
const os = (JSON.parse(readFileSync(report, "utf8")) as Report).Metadata?.OS;
if (os?.EOSL) {
  console.error(`${values.image}: ${os.Family} ${os.Name} はサポートが終わっている`);
  process.exit(1);
}
```

`image:build`タスクは，`DOCKER_BUILD_ARGS`をビルドの引数に足す．
社内のプロキシを通す環境などで，Dockerfileを変えずに引数を渡せる．
Dockerfileの`ARG GOPROXY`も，同じ目的で置いた．

### 誤りへの対応

Dockerfileの実行の段階は次のようになる．

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /usr/local/bin/api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
```

distrolessの`nonroot`のイメージは，もともとUID 65532で動く．
それでも`USER`を書くのは，Dockerfileを読むだけで利用者が分かり，検査でも確かめられるからである．

バケットには，パブリックアクセスのブロックと，KMSの鍵での暗号化を足した．

```hcl
resource "aws_s3_bucket_public_access_block" "exports" {
  bucket                  = aws_s3_bucket.exports.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

鍵は`infra/kms.tf`に1つ作り，2つのバケットで使う．
`enable_key_rotation = true`で，鍵を毎年自動で交換する．

## 5-6 振り返り

1. イメージの検査で，サポート中のイメージが通ることも確かめたかを見る．
2. 解答では，見つかった誤りをこの回ですべて直した．
   直す手間の大きい誤りは，期限つきの例外で1件ずつ受け入れ，ゲートの有効化を先に済ませる．
   ゲートを有効にしないと，その間に新しい誤りが増える．
3. サポートの終わったOSには，新しい脆弱性が見つかっても修正が出ない．今の時点で脆弱性がなくても，将来の修正を受けられない．
4. バケットを公開せず，APIが期限つきの署名つきURLを発行する．利用者ごとに誰が取得したかも記録できる．
5. 解答では，2つの項目の仕組み(`image`ジョブと`config`ジョブ)が`ci.yml`にあり，`ops:verify`が通る．

## 5-7 発展課題

`trivy image --format cyclonedx --output out/evidence/sbom/api-image.cdx.json app-api:local`で，イメージのSBOMを作れる．
`sbom`タスクに足すと，`evidence`ジョブがそのまま保管する．
ただし，OSのパッケージにはGPLのものも含まれる．`gate:license`の対象にするなら，OSのパッケージをどう扱うかを先に決める．
