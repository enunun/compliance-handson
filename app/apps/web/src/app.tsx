import { type Note, noteInputSchema } from "@app/shared";
import { Hono } from "hono";
import { deleteCookie, getCookie, setCookie } from "hono/cookie";

type Options = {
  /** APIのベースURL(例：http://localhost:8080)． */
  apiUrl: string;
  /** APIを呼ぶ関数．テストでは差し替える． */
  fetch?: typeof fetch;
};

const TOKEN_COOKIE = "token";

/** 題材のSaaSの画面を返す． */
export function createApp({ apiUrl, fetch: fetchApi = fetch }: Options) {
  const app = new Hono();

  const callApi = (path: string, init: RequestInit = {}, token?: string) =>
    fetchApi(`${apiUrl}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });

  app.get("/", async (c) => {
    const token = getCookie(c, TOKEN_COOKIE);
    if (!token) {
      return c.html(<LoginPage />);
    }
    const res = await callApi("/api/notes", {}, token);
    if (res.status === 401) {
      deleteCookie(c, TOKEN_COOKIE);
      return c.redirect("/");
    }
    const notes = (await res.json()) as Note[];
    return c.html(<NotesPage notes={notes} />);
  });

  app.post("/login", async (c) => {
    const form = await c.req.parseBody();
    const res = await callApi("/api/login", {
      method: "POST",
      body: JSON.stringify({ email: form.email, password: form.password }),
    });
    if (!res.ok) {
      return c.html(
        <LoginPage error="メールアドレスかパスワードが違います" />,
        401,
      );
    }
    const { token } = (await res.json()) as { token: string };
    setCookie(c, TOKEN_COOKIE, token, { httpOnly: true, sameSite: "Lax" });
    return c.redirect("/");
  });

  app.post("/notes", async (c) => {
    const token = getCookie(c, TOKEN_COOKIE);
    if (!token) {
      return c.redirect("/");
    }
    const input = noteInputSchema.safeParse(await c.req.parseBody());
    if (!input.success) {
      return c.text("メモの本文は1〜1000文字で書いてください", 400);
    }
    await callApi(
      "/api/notes",
      { method: "POST", body: JSON.stringify(input.data) },
      token,
    );
    return c.redirect("/");
  });

  return app;
}

function Layout({ children }: { children: unknown }) {
  return (
    <html lang="ja">
      <head>
        <meta charset="utf-8" />
        <title>Notes</title>
      </head>
      <body>{children}</body>
    </html>
  );
}

function LoginPage({ error }: { error?: string }) {
  return (
    <Layout>
      <h1>ログイン</h1>
      {error && <p role="alert">{error}</p>}
      <form method="post" action="/login">
        <input
          name="email"
          type="email"
          placeholder="メールアドレス"
          required
        />
        <input
          name="password"
          type="password"
          placeholder="パスワード"
          required
        />
        <button type="submit">ログイン</button>
      </form>
    </Layout>
  );
}

function NotesPage({ notes }: { notes: Note[] }) {
  return (
    <Layout>
      <h1>メモ</h1>
      <form method="post" action="/notes">
        <textarea name="body" required />
        <button type="submit">追加</button>
      </form>
      <ul>
        {notes.map((note) => (
          <li>{note.body}</li>
        ))}
      </ul>
    </Layout>
  );
}
