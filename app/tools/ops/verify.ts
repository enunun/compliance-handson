/** 検証で見つかった問題． */
export type Problem = { file: string; message: string };

export type VerifyOptions = {
  /** standards.yaml と items.yaml がある場所． */
  opsDir: string;
  /** ワークフローがある場所． */
  workflowsDir: string;
  /** JSON Schemaがある場所． */
  schemaDir: string;
};

/** 基準値と運用項目を検証し，見つかった問題を返す． */
export function verify(_options: VerifyOptions): Problem[] {
  throw new Error("第0回で実装する");
}
