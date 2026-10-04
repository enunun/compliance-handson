# 第6回 解説

解答のコードは`mise run iteration:solution 6`で当てられる．
差分は`iterations/06/solution.patch`にある．

## 6-1 準備

`ci.yml`は，`actions/checkout@v5`のようにタグでactionを参照している．
タグを付け替えられると，次の実行で別のコードが動く．

## 6-2 学ぶこと

この回の要点は，「何を使って作ったか」と「何を出荷したか」の両方を，後から検証できる形で固定することである．

## 6-3 テストリスト

### ゲートのテスト(`tests/gates/pinning.test.ts`)

- [x] タグやブランチで参照するactionやワークフローがあれば失敗し，ファイルと参照を表示する．
  ステップの`uses:`(`actions/checkout@v5`)と，ジョブの`uses:`(`lint.yml@main`)の両方を見つける．
- [x] ハッシュで固定した参照と，同じリポジトリの中の参照だけなら通る．

`release.yml`はGitHubでしか動かないので，ゲートのテストの対象にはしていない．
6-5の手順のとおり，タグをpushして確かめる．

## 6-4 設計書

運用項目は2つ足した．

```yaml
  - id: actions-pinning
    title: ワークフローの部品の固定
    category: 基盤運用
    requirements: [SLSA:Build, SSDF:PO.3]
    mechanisms:
      - workflow: ci.yml
        job: pinning
    evidence: CIのpinningジョブの結果
    timing: 定常(変更ごと)
    actor: 自動

  - id: artifact-signing
    title: 成果物の署名とprovenance
    category: 基盤運用
    requirements: [SLSA:Build-L2, SSDF:PS.2]
    mechanisms:
      - workflow: release.yml
        job: release
    evidence: 署名とprovenance(GHCRとGitHubのattestation)，検証の結果(JSON)
    timing: 定常(リリースごと)
    actor: 自動
    runbook: runbooks/verify-image.md
```

手順書には2つのことを書いた．タグではなくダイジェストで確かめることと，検証の失敗のときはデプロイせずに知らせることである．

## 6-5 テスト駆動の実装

### gate:pinning

参照を集めて，固定されていないものを返す関数は次のとおりである．

```ts
export function unpinned(workflow: Workflow): string[] {
  const refs = Object.values(workflow.jobs ?? {}).flatMap((job) => [
    ...(job.uses ? [job.uses] : []),
    ...(job.steps ?? []).flatMap((step) => (step.uses ? [step.uses] : [])),
  ]);
  return refs.filter(
    (ref) =>
      !ref.startsWith("./") &&
      !/@[0-9a-f]{40}$/.test(ref) &&
      !/^docker:\/\/.+@sha256:[0-9a-f]{64}$/.test(ref),
  );
}
```

`docker://`でコンテナを直接使う参照も，ダイジェストで固定していれば通す．

### release.yml

外部のactionは，`checkout`，`mise-action`，`attest-build-provenance`，`upload-artifact`の4つにした．
署名(cosign)とイメージの操作(docker)は，miseで入れた道具やランナーにあるコマンドで行う．

署名と検証の要点は次のとおりである．

```yaml
      - name: キーレスで署名する
        run: cosign sign --yes "${{ steps.push.outputs.image }}@${{ steps.push.outputs.digest }}"
      - name: provenanceを付ける
        uses: actions/attest-build-provenance@96278af6caaf10aea03fd8d33a09a777ca52d62f # v3.2.0
        with:
          subject-name: ${{ steps.push.outputs.image }}
          subject-digest: ${{ steps.push.outputs.digest }}
          push-to-registry: true
      - name: 署名とprovenanceを検証し，証跡に残す
        env:
          GH_TOKEN: ${{ github.token }}
        run: mise run release:verify "${{ steps.push.outputs.image }}@${{ steps.push.outputs.digest }}"
```

`verify-image.sh`は，署名した者がこのリポジトリの`release.yml`であることまで確かめる．

```bash
cosign verify \
  --certificate-identity-regexp "^https://github.com/${repo}/\.github/workflows/release\.yml@" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  "$ref" > "$dir/cosign-verify.json"

gh attestation verify "oci://${ref}" \
  --repo "$repo" \
  --signer-workflow "${repo}/.github/workflows/release.yml" \
  --format json > "$dir/provenance-verify.json"
```

## 6-6 振り返り

1. ジョブの`uses:`(reusable workflow)も調べるテストを入れたかを確かめる．第13回でreusable workflowを使うようになる．
2. 第10回で，Renovateにactionの更新のプルリクエストを作らせる．Renovateはハッシュとコメントの版を一緒に更新する．
3. 秘密鍵の漏れる心配がなく，鍵の交換や保管の運用も要らない．署名した者は，証明書に記録されたワークフローで分かる．
4. 「Sigstoreで誰かが署名した」ことしか確かめていない．攻撃者が自分のリポジトリで署名したイメージも通る．
5. 解答では，`pinning`ジョブが`ci.yml`に，`release`ジョブが`release.yml`にあり，`ops:verify`が通る．

## 6-7 発展課題

`unpinned`と同じように，ワークフローの最上位とジョブの`permissions`を調べる関数を足す．
ワークフローの最上位に`permissions`がなければ失敗にする．
ゲートのテストには，`permissions`のないワークフローのfixtureを足す．
