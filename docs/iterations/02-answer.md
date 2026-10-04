# 第2回 解説

解答のコードは`mise run iteration:solution 2`で当てられる．
差分は`iterations/02/solution.patch`にある．

## 2-1 準備

lodash 4.17.20にはHIGHの脆弱性が2件あり，golang-jwt 4.5.1にはHIGHの脆弱性(CVE-2025-30204)が1件ある．
`token.go`はトークンに署名するだけで，解析はしない．

## 2-2 学ぶこと

この回の要点は，ゲートで止めるだけでなく，止めたものを受け入れる道筋(例外)も仕組みにすることである．
例外の道筋がないと，開発者はゲートを外すか，検査を飛ばすことになる．

## 2-3 テストリスト

### ゲートのテスト(`tests/gates/sca.test.ts`)

- [x] 基準値以上の脆弱性があれば失敗し，証跡のJSONに残す．
- [x] 基準値以上の脆弱性がなければ通る．
- [x] 例外に載った脆弱性は外す．
- [x] 期限を過ぎた例外では外さない．

### ゲートのテスト(`tests/gates/exceptions.test.ts`)

- [x] 期限内の例外なら通り，件数を表示する．
- [x] 期限を過ぎた例外があれば失敗し，どの例外かを表示する．
- [x] 期限が基準値の最長の日数を超える例外があれば失敗する．

### 既存のテストへの影響

- `ops:verify`のfixtureの`standards.yaml`に，`vulnerability`と`exception`を足す．
  足さないと，正しい運用項目のfixtureがスキーマに合わなくなる．
- `gate:secrets`のテストは，証跡を`EVIDENCE_DIR`の一時的な場所から読むように変える．

## 2-4 設計書

基準値には2つの項目を足した．

```yaml
vulnerability:
  # この重大度以上の脆弱性があれば，ゲートを失敗させる．
  fail_severity: HIGH

exception:
  # 例外の期限は，承認した日から最長でこの日数までにする．
  max_days: 90
```

例外は次のとおりである．
理由には，到達しない根拠(呼んでいない関数)と，いつ直すかを書いた．

```yaml
exceptions:
  - id: EXC-001
    target: CVE-2025-30204
    package: github.com/golang-jwt/jwt/v4
    reason: >-
      脆弱なのはトークンを解析する処理(ParseUnverified)で，APIはトークンに署名するだけで解析しない．
      到達しないことを確かめたので，次の定期更新で4.5.2以上に上げるまで受け入れる．
    approver: "@security-lead"
    expires: 2026-12-31
```

運用項目は2つ足した．

- `dependency-vulnerabilities`：依存関係の脆弱性への対処は，パッチ運用(基盤運用)である．
- `exception-requests`：開発者からの依頼を受け付けて判断する運用なので，サポートデスク運用(業務運用)にした．
  実施者は「人」で，期限の検査だけを自動で行う．

例外のスキーマでは，`expires`を`format: date`にした．
Ajvで`format`を検証するには，`ajv-formats`を足す．

## 2-5 テスト駆動の実装

### gate:sca

例外をTrivyの除外の設定に変える関数は次のとおりである．
期限は`expired_at`に移すので，期限を過ぎた例外はTrivyが自動で無視する．

```ts
export function toTrivyIgnore(exceptions: Exception[]) {
  return {
    vulnerabilities: exceptions.map((e) => ({
      id: e.target,
      statement: `${e.id}: ${e.reason} (承認：${e.approver})`,
      expired_at: e.expires,
    })),
  };
}
```

作った除外の設定は，証跡の置き場所に`trivyignore.yaml`として残す．
その時点でどの例外を適用したかを，後から示せる．

証跡の置き場所は，環境変数`EVIDENCE_DIR`で替えられるようにした．

```ts
const evidenceDir = join(process.env.EVIDENCE_DIR ?? "out/evidence", "sca");
```

`tests/gates/helpers.ts`は，一時的な場所を作り，`run`で実行するコマンドに渡す．

```ts
export const evidenceDir = mkdtempSync(join(tmpdir(), "gate-evidence-"));

export function run(command: string, args: string[], cwd = appRoot) {
  const result = spawnSync(command, args, {
    cwd,
    encoding: "utf8",
    env: { ...process.env, EVIDENCE_DIR: evidenceDir },
  });
  // ...
}
```

`gate:secrets`のタスクも，シェルの`${EVIDENCE_DIR:-out/evidence}`で同じ置き場所を使う．

### gate:exceptions

期限の検査は次のとおりである．
`today`を引数にしたので，ゲートのテストでは`--today 2026-10-04`を渡して結果を固定できる．

```ts
export function checkExceptions(exceptions: Exception[], maxDays: number, today: string): string[] {
  const now = Date.parse(today);
  return exceptions.flatMap((e) => {
    const days = (Date.parse(e.expires) - now) / day;
    if (days < 0) {
      return [`${e.id}: 期限(${e.expires})を過ぎている`];
    }
    if (days > maxDays) {
      return [`${e.id}: 期限(${e.expires})が基準値の${maxDays}日を超えている`];
    }
    return [];
  });
}
```

### 仕込まれた脆弱性への対応

lodashは4.18.1に上げた．
golang-jwtは例外にした．

## 2-6 振り返り

1. 期限を過ぎた例外のテストを入れたかを確かめる．
2. `gate:sca`だけだと，期限を過ぎた例外は黙って無効になり，脆弱性が再び見つかって初めて気づく．
   `gate:exceptions`は，例外の側の問題として，理由とともに知らせる．
3. コードが変わり，解析の処理を呼ぶようになると，到達する脆弱性になる．期限は，見直しを強制する仕組みである．
4. 例外の期限が来た日から，テストが失敗し始める．ゲートのテストは，いつ実行しても同じ結果を返すようにする．
5. 解答では，2つの項目の仕組み(`sca`ジョブと`exceptions`ジョブ)が`ci.yml`にあり，`ops:verify`が通る．

## 2-7 発展課題

Trivyの結果のJSONは，`Results[].Vulnerabilities[].VulnerabilityID`に検出を持つ．
例外を適用しない検査の結果から識別子を集め，例外の`target`がその中になければ失敗にする．
ゲートのテストには，存在しない脆弱性を対象にした例外のfixtureを足す．
