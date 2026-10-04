import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    projects: [
      {
        test: {
          name: "unit",
          include: [
            "apps/*/src/**/*.test.{ts,tsx}",
            "packages/*/src/**/*.test.ts",
            "tools/**/*.test.ts",
          ],
        },
      },
      {
        test: {
          name: "gates",
          include: ["tests/gates/**/*.test.ts"],
          // ゲートは外部のコマンドを実行するので，時間がかかる．
          testTimeout: 60_000,
        },
      },
    ],
  },
});
