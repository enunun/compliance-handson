import { describe, expect, it } from "vitest";
import { MAX_NOTE_LENGTH, noteInputSchema } from "./index.ts";

describe("noteInputSchema", () => {
  it("前後の空白を除いた本文を受け付ける", () => {
    expect(noteInputSchema.parse({ body: "  hello " })).toEqual({
      body: "hello",
    });
  });

  it("空白だけの本文を拒む", () => {
    expect(noteInputSchema.safeParse({ body: "   " }).success).toBe(false);
  });

  it("最大文字数を超える本文を拒む", () => {
    const body = "a".repeat(MAX_NOTE_LENGTH + 1);
    expect(noteInputSchema.safeParse({ body }).success).toBe(false);
  });
});
