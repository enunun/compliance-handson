# 第3回 解説

解答のコードは`mise run iteration:solution 3`で当てられる．
差分は`iterations/03/solution.patch`にある．

## 3-1 準備

TinyMCE 7のライセンスはGPL-2.0-or-laterである．
初期化の設定の`license_key: "gpl"`も，GPLで使うことを示している．

## 3-2 学ぶこと

この回の要点は，証跡を作るだけでなく，信用できる形で保管するところまでを仕組みにすることである．
SBOMも，ゲートの結果も，保管して初めて後から示せる．

## 3-3 テストリスト

### ゲートのテスト(`tests/gates/license.test.ts`)

- [x] 許可したライセンスだけなら通る．
- [x] 許可していないライセンスがあれば失敗し，部品とライセンスを表示する．
- [x] ライセンスの分からない部品があれば失敗する．

### 運用テスト(`tests/ops/evidence.test.ts`)

- [x] 証跡の保管場所は，基準値の日数だけ消せない(COMPLIANCEモードの既定の保管期間)．
- [x] 保管した証跡は，版を指定しても消せない(`AccessDenied`)．

### 既存のテストへの影響

`ops:verify`のfixtureの`standards.yaml`に，`license`と`evidence`を足す．

## 3-4 設計書

基準値には，許可するライセンスと保管期間を足した．

```yaml
license:
  # 出荷する依存関係に許可するライセンス．ここにないものは，法務の確認を経て足す．
  allowed:
    - MIT
    - ISC
    - BSD-2-Clause
    - BSD-3-Clause
    - Apache-2.0
    - 0BSD

evidence:
  # 証跡を消せない状態で保管する日数．監査の対象期間(1年)に余裕を持たせる．
  retention_days: 400
```

運用項目は3つ足した．

- `sbom`：部品の記録は基盤の管理なので基盤運用にした．
- `license-policy`：ライセンスの方針を守らせる運用なので運用管理にした．仕組みは`sbom`ジョブを共有する．
- `evidence-retention`：証跡の保管場所の運用なので基盤運用にした．

## 3-5 テスト駆動の実装

### gate:license

検査の中心は次の関数である．

```ts
export function checkLicenses(components: Component[], allowed: string[]): string[] {
  return components
    .filter((c) => c.type !== "application" && c.version)
    .flatMap((c) => {
      const licenses = licensesOf(c);
      const name = `${c.name}@${c.version}`;
      if (licenses.length === 0) {
        return [`${name}: ライセンスが分からない`];
      }
      const denied = licenses.filter((l) => !allowed.includes(l));
      return denied.length > 0 ? [`${name}: ${denied.join(", ")} は許可されていない`] : [];
    });
}
```

`sbom`タスクは，Webの画面(`app/`のpnpmのロックファイル)とAPI(`apps/api`の`go.mod`)を別々に調べる．
`gate:license`は`depends`で`sbom`を先に動かす．

```toml
[tasks."gate:license"]
description = "SBOMの依存関係のライセンスが，許可したものかを検査する"
depends = ["sbom"]
run = 'pnpm exec tsx tools/gates/license.ts "${EVIDENCE_DIR:-out/evidence}"/sbom/*.cdx.json'
```

### 証跡の保管場所

`infra/evidence.tf`は，基準値を読んで保管期間に使う．

```hcl
locals {
  standards = yamldecode(file("${path.module}/../ops/standards.yaml"))
}
```

ライフサイクルでは，保管期間の翌日に古い版と今の版を消す．
保管期間の間はObject Lockが守るので，ライフサイクルでも消えない．

運用テストでは，`aws s3api`の結果のJSONを読んで確かめた．
削除できないことは，終了コードと`AccessDenied`の両方で確かめる．

### CI

`evidence`ジョブの要点は次のとおりである．

```yaml
  evidence:
    needs: [secrets, sca, sbom]
    if: always()
    runs-on: ubuntu-latest
    services:
      ministack:
        image: ministackorg/ministack:1.5.21@sha256:7a30ff0670578e10528b5a9a21e795ef494f96bbc35691c9b0742ea35de71808
        ports:
          - 4566:4566
        options: --workdir /opt/ministack
    # (env，defaultsは省略)
    steps:
      - uses: actions/checkout@v5
      - run: echo "127.0.0.1 000000000000.localhost" | sudo tee -a /etc/hosts
      - uses: jdx/mise-action@v3
      - uses: actions/download-artifact@v4
        with:
          pattern: evidence-*
          path: app/out/evidence/
          merge-multiple: true
      - run: mise run up
      - run: mise run test:ops
      - run: mise run evidence:upload
```

`options: --workdir /opt/ministack`は，actでサービスコンテナを動かすために要る．
イメージの既定と同じ値なので，GitHubでも害はない．

`upload-artifact`は，指定した場所からの相対で中身を残す．
各ジョブが`app/out/evidence/`全体を残すと，`merge-multiple`で集めたときに`secrets/`，`sca/`，`sbom/`の形がそのまま保たれる．

## 3-6 振り返り

1. 運用テストを，ゲートのテストと分けたかを確かめる．運用テストは本番の環境が要る．
2. 分からないものを通すと，許可していないライセンスが「分からない」の中に紛れても気づけない．
   分からない部品は，調べてから許可の一覧に足すか，部品を替える．
3. ライセンスの義務は，期限が来れば解決するものではない．使い続ける限り義務が続くので，直すか，法務の確認を経て許可の一覧を変える．
4. 証跡を集めて送る仕組みと，保管場所の設定(Object Lock，保管期間)が正しく動くことを確かめている．
   本物の環境では，保管場所はジョブの外にあり，CIはOIDCで得た権限を使って送る．第11回でこの権限を扱う．
5. 解答では，3つの項目の仕組み(`sbom`ジョブと`evidence`ジョブ)が`ci.yml`にあり，`ops:verify`が通る．

## 3-7 発展課題

2つのSBOMの`components`から`name@version`の集合を作り，差を表示する．
前回のSBOMは，証跡の保管場所から`aws s3 cp`で取得する．
ゲートのテストには，部品が1つ増えたSBOMと減ったSBOMの組を用意する．
