import { spawnSync } from "node:child_process";
import { resolve } from "node:path";

/** 題材のディレクトリ(app/)． */
export const appRoot = resolve(import.meta.dirname, "../..");

/** ゲートのテストに使う入力の場所． */
export function fixture(...parts: string[]): string {
  return resolve(import.meta.dirname, "fixtures", ...parts);
}

/** コマンドを実行し，終了コードと出力を返す． */
export function run(command: string, args: string[], cwd = appRoot) {
  const result = spawnSync(command, args, { cwd, encoding: "utf8" });
  return {
    status: result.status,
    output: `${result.stdout}${result.stderr}`,
  };
}
