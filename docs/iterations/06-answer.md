# 第6回 解説

解答のコードは`mise run iteration:solution 6`で当てられる．
差分は`iterations/06/solution.patch`にある．

## 6-1 準備

`ci.yml`は，`actions/checkout@v5`のようにタグでactionを参照している．
タグを付け替えられると，次の実行で別のコードが動く．

`uses:`にハッシュを書いても，そのactionが中で使うactionはタグで参照されていることがある．
推移的な依存まで固定するには，ロックファイルを使う．

## 6-2 学ぶこと

この回の要点は，「何を使って作ったか」と「何を出荷したか」の両方を，後から検証できる形で固定することである．
actionはロックファイルで，ロックファイルが固定しないものはハッシュやダイジェストで固定する．

## 6-3 テストリスト

### ゲートのテスト(`tests/gates/pinning.test.ts`)

- [x] ロックファイルにないactionや，ロックファイルに載っていないワークフローがあれば失敗し，ファイルと参照を表示する．
- [x] 別のリポジトリのreusable workflow(`lint.yml@main`)と，コンテナのイメージ(`docker://alpine:3.22`)は，ハッシュかダイジェストで固定していなければ失敗する．
- [x] ロックファイルに載ったaction，ハッシュで固定した参照，同じリポジトリの中の参照だけなら通る．
  サブディレクトリのaction(`actions/cache/restore@v4.2.3`)は，リポジトリの単位(`actions/cache@v4.2.3`)で載っていれば通る．

`gh actions-lock --verify`はGitHubのAPIを使うので，ゲートのテストの対象にはしていない．
CIの`pinning`ジョブで動かす．
`release.yml`もGitHubでしか動かないので，6-5の手順のとおり，タグをpushして確かめる．

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
    evidence: ロックファイル(actions.lock)の変更の記録，CIのpinningジョブの結果
    timing: 定常(変更ごと)
    actor: 自動
    runbook: runbooks/actions-lock.md

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

`actions-lock.md`には，ロックファイルの差分のレビューの観点を書いた．
`commit`が変わったactionは変更点を確かめ，`owner_id`か`repo_id`が変わったらマージせずに知らせる．
`verify-image.md`には，タグではなくダイジェストで確かめることと，検証の失敗のときはデプロイせずに知らせることを書いた．

## 6-5 テスト駆動の実装

### gate:pinning

ロックファイルと突き合わせて，固定されていないものを返す関数は次のとおりである．

```ts
export function unpinned(
  file: string,
  workflow: Workflow,
  lock: Lockfile,
): string[] {
  const locked = lock.workflows?.[`.github/workflows/${file}`];
  const problems = locked ? [] : [`${file}: ロックファイルに載っていない`];
  for (const job of Object.values(workflow.jobs ?? {})) {
    // ジョブのuses:(reusable workflow)は，ロックファイルが固定しない．
    if (
      job.uses &&
      !job.uses.startsWith("./") &&
      !/@[0-9a-f]{40}$/.test(job.uses)
    ) {
      problems.push(`${file}: ${job.uses} がハッシュで固定されていない`);
    }
    for (const ref of (job.steps ?? []).flatMap((s) =>
      s.uses ? [s.uses] : [],
    )) {
      if (ref.startsWith("./")) {
        continue;
      }
      if (ref.startsWith("docker://")) {
        if (!/@sha256:[0-9a-f]{64}$/.test(ref)) {
          problems.push(`${file}: ${ref} がダイジェストで固定されていない`);
        }
        continue;
      }
      const key = lockKey(ref);
      if (
        locked &&
        !(locked.includes(key) && lock.dependencies?.[key]?.commit)
      ) {
        problems.push(`${file}: ${ref} がロックファイルにない`);
      }
    }
  }
  return problems;
}
```

ロックファイルは，actionをリポジトリの単位で記録し，`owner/repo`を小文字にする．
`lockKey`は，`uses:`の値をその形に直す．

このゲートは，ロックファイルが揃っているかを手元で確かめるだけで，記録したコミットが正しいかは確かめない．
それは，CIの`pinning`ジョブで`gh actions-lock --verify`が確かめる．

```yaml
      - run: mise run gate:pinning
      # ロックファイルのコミットが今も元のタグから辿れるかを，GitHubのAPIで確かめる．actでは動かさない．
      - if: ${{ !env.ACT }}
        env:
          GH_TOKEN: ${{ github.token }}
        run: mise run actions:verify
```

`gh actions-lock`の版は，`mise.toml`の`actions:setup`タスクで固定した．

```toml
[tasks."actions:setup"]
description = "ロックファイルを作るghの拡張(gh actions-lock)を入れる．版はここで固定する"
run = "gh extension install github/gh-actions-lock --pin v0.4.0 --force"
```

### release.yml

外部のactionは，`checkout`，`mise-action`，`attest-build-provenance`，`upload-artifact`の4つにした．
署名(cosign)とイメージの操作(docker)は，miseで入れた道具やランナーにあるコマンドで行う．
`attest-build-provenance`は中で`actions/attest`を使う．ロックファイルの`uses`に，この推移的な依存が載る．

```yaml
    'actions/attest-build-provenance@v4.1.0':
        ref: 'v4.1.0'
        commit: 'sha1-a2bbfa25375fe432b6a289bc6b6cd05ecd0c4c32'
        owner_id: 44036562
        repo_id: 760702757
        uses:
            - 'actions/attest@59d89421af93a897026c735860bf21b6eb4f7b26'
```

署名と検証の要点は次のとおりである．

```yaml
      - name: キーレスで署名する
        run: cosign sign --yes "${{ steps.push.outputs.image }}@${{ steps.push.outputs.digest }}"
      - name: provenanceを付ける
        uses: actions/attest-build-provenance@v4.1.0
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

1. ジョブの`uses:`(reusable workflow)も調べるテストを入れたかを確かめる．第13回でreusable workflowを使うようになる．ロックファイルはその呼び出しを固定しないので，ゲートでハッシュを求める．
2. GitHub Actionsは，タグではなくロックファイルの`commit`を動かすからである．
   レビューでは，`commit`が変わったactionの変更点と，`owner_id`と`repo_id`が変わっていないこと，`uses`に増えた推移的な依存を見る．
3. 第10回で，Renovateにactionの更新のプルリクエストを作らせる．Renovateは，ワークフローのタグとロックファイルを一緒に更新する．
4. 秘密鍵の漏れる心配がなく，鍵の交換や保管の運用も要らない．署名した者は，証明書に記録されたワークフローで分かる．
5. 「Sigstoreで誰かが署名した」ことしか確かめていない．攻撃者が自分のリポジトリで署名したイメージも通る．
6. 解答では，`pinning`ジョブが`ci.yml`に，`release`ジョブが`release.yml`にあり，`ops:verify`が通る．

## 6-7 発展課題

`unpinned`と同じように，ワークフローの最上位とジョブの`permissions`を調べる関数を足す．
ワークフローの最上位に`permissions`がなければ失敗にする．
ゲートのテストには，`permissions`のないワークフローのfixtureを足す．
