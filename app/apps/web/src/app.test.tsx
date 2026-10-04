import { describe, expect, it, vi } from "vitest";
import { createApp } from "./app.tsx";

const apiUrl = "http://api.test";

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function form(fields: Record<string, string>) {
  return { method: "POST", body: new URLSearchParams(fields) };
}

describe("画面", () => {
  it("ログインしていなければログイン画面を表示する", async () => {
    const res = await createApp({ apiUrl }).request("/");
    expect(await res.text()).toContain("ログイン");
  });

  it("ログインに成功するとトークンをCookieに保存する", async () => {
    const fetch = vi.fn(async () => json({ token: "t1" }));
    const res = await createApp({ apiUrl, fetch }).request(
      "/login",
      form({ email: "a@example.com", password: "secret" }),
    );
    expect(res.status).toBe(302);
    expect(res.headers.get("set-cookie")).toContain("token=t1");
  });

  it("ログインに失敗するとエラーを表示する", async () => {
    const fetch = vi.fn(async () => json({ error: "x" }, 401));
    const res = await createApp({ apiUrl, fetch }).request(
      "/login",
      form({ email: "a@example.com", password: "wrong" }),
    );
    expect(res.status).toBe(401);
    expect(await res.text()).toContain("違います");
  });

  it("ログインしていれば自分のメモを表示する", async () => {
    const fetch = vi.fn(async () =>
      json([{ id: 1, ownerId: 1, body: "hello", createdAt: "" }]),
    );
    const res = await createApp({ apiUrl, fetch }).request("/", {
      headers: { Cookie: "token=t1" },
    });
    expect(await res.text()).toContain("<li>hello</li>");
    expect(fetch).toHaveBeenCalledWith(
      `${apiUrl}/api/notes`,
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: "Bearer t1" }),
      }),
    );
  });

  it("空のメモはAPIに送らずに拒む", async () => {
    const fetch = vi.fn();
    const res = await createApp({ apiUrl, fetch }).request("/notes", {
      ...form({ body: "  " }),
      headers: { Cookie: "token=t1" },
    });
    expect(res.status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });
});
