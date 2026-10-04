# 第4回 解説

解答のコードは`mise run iteration:solution 4`で当てられる．
差分は`iterations/04/solution.patch`にある．

## 4-1 準備

検索はSQLの文を`fmt.Sprintf`で組み立て，改行の表示は本文を`raw()`で出力している．
どちらも利用者の入力がそのまま解釈される．

## 4-2 学ぶこと

この回の要点は，道具の既定のルールに頼り切らず，自分たちのコードに合わせてルールとその判断の基準を持つことである．

## 4-3 テストリスト

### ゲートのテスト(`tests/gates/sast.test.ts`)

- [x] SQLを文字列の組み立てで作るコードがあれば失敗し，ルールの識別子を表示する．
- [x] `raw()`に変数を渡すコードがあれば失敗する．
- [x] 直したコードでは通り，`sast: no findings`を表示する．
- [x] 例外に載った検出は外す．例外がなければ失敗し，あれば通る．

### 単体テスト

- [x] Web画面：メモの改行は`<br/>`で表示し，HTMLはエスケープする．
- [x] API：検索は，自分のメモのうち語を含むものだけを返す．

### 既存のテストへの影響

APIのテスト用の`memStore`に，`SearchNotes`を足す．
`Store`のインターフェースに関数が増えたので，足さないとテストのコードがコンパイルできない．

## 4-4 設計書

例外は2件足した．
どちらも，手順書の「誤検知と判断する基準」に当てはまる．

```yaml
  - id: EXC-002
    target: use-of-sha1
    path: apps/api/internal/server/server.go
    reason: >-
      メモの一覧のETag(キャッシュの判定)にSHA-1を使っている．改ざんの検知や署名には使っていないので，
      衝突攻撃の影響を受けない．誤検知として扱い，期限までにSHA-256へ替えるかを見直す．
    approver: "@security-lead"
    expires: 2026-12-31

  - id: EXC-003
    target: use-tls
    path: apps/api/cmd/api/main.go
    reason: >-
      本番ではロードバランサーがTLSを終端し，APIへは内部のネットワークだけで届く構成である．
      誤検知として扱い，期限までに構成を見直す．
    approver: "@security-lead"
    expires: 2026-12-31
```

`target`にはルールの識別子の末尾を書いた．
`path`で場所を絞ると，同じルールの別の場所の検出は外れない．

## 4-5 テスト駆動の実装

### 自前のルール

SQLのルールは，`fmt.Sprintf`の結果を直接渡す形，`+`でつないだ文字列を渡す形，変数に入れてから渡す形の3つを見つける．
関数の名前は`metavariable-regex`で`Query`，`QueryRow`，`Exec`に絞った．

```yaml
    patterns:
      - pattern-either:
          - pattern: $DB.$METHOD($CTX, fmt.Sprintf(...), ...)
          - pattern: $DB.$METHOD($CTX, $A + $B, ...)
          - pattern: |
              $SQL := fmt.Sprintf(...)
              ...
              $DB.$METHOD($CTX, $SQL, ...)
      - metavariable-regex:
          metavariable: $METHOD
          regex: ^(Query|QueryRow|Exec)$
```

`metavariable-regex`を`patterns`の外に書くと効かない．
その場合，`w.Header().Set("ETag", fmt.Sprintf(...))`にも当たる．
ルールの誤りは，誤検知の山として開発者に返ってくるので，ルールにもゲートのテストを書く．

`raw()`のルールは，`pattern-not: raw("...")`で文字列の定数を除いた．

### ゲート

検出が例外に当たるかは，次の関数で判断する．

```ts
export function isExcepted(finding: Finding, exceptions: Exception[], today: string): boolean {
  return exceptions.some(
    (e) =>
      e.expires >= today &&
      (finding.check_id === e.target || finding.check_id.endsWith(`.${e.target}`)) &&
      (!e.path || finding.path.endsWith(e.path)),
  );
}
```

`YYYY-MM-DD`の文字列は，文字列として比べても日付の順になる．

ゲートのテストの入力は`tests/gates/fixtures/`にあるので，題材全体を調べるときだけ`--exclude fixtures`で外す．
`.semgrepignore`はgitのリポジトリの直下から読まれるので，リポジトリ直下に置いた．

### 検出への対応

SQLは次のように直した．

```go
	rows, err := p.pool.Query(ctx,
		`SELECT id, owner_id, body, created_at FROM notes
		 WHERE owner_id = $1 AND body LIKE '%' || $2 || '%' ORDER BY id`, ownerID, query)
```

改行の表示は次のように直した．

```tsx
          <li>
            {truncate(note.body, { length: 100 })
              .split("\n")
              .map((line, i) => (
                <>
                  {i > 0 && <br />}
                  {line}
                </>
              ))}
          </li>
```

## 4-6 振り返り

1. 例外のテストで，例外がない場合に失敗することも確かめたかを見る．例外の効き目を確かめるには，両方が要る．
2. レジストリのルールセットは，標準ライブラリの`database/sql`のような広く使われる書き方を対象にしている．
   pgxの関数やHonoの`raw()`は対象外だった．
3. SHA-256に替えれば，例外の管理が要らず，次の監査でも説明が要らない．
   変える手間の小さいときは直す方がよい．例外は，直す手間の大きいときに選ぶ．
4. 開発者が検出を信用しなくなり，本物の検出まで例外にされる．ルールも，だめな入力と正しい入力で確かめる．
5. 解答では，`sast`の項目の仕組み(`sast`ジョブ)が`ci.yml`にあり，`ops:verify`が通る．

## 4-7 発展課題

Honoの`c.redirect()`に，`c.req.query()`の値をそのまま渡す形を見つけるルールの例は次のとおりである．

```yaml
  - id: hono-open-redirect
    languages: [typescript]
    severity: ERROR
    message: 利用者の入力をそのままリダイレクト先にしない．
    patterns:
      - pattern-either:
          - pattern: $C.redirect($C.req.query(...))
          - pattern: |
              const $URL = $C.req.query(...)
              ...
              $C.redirect($URL)
```

ゲートのテストには，この形のコードと，決まった場所へだけリダイレクトするコードを用意する．
