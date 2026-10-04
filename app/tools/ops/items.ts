import { readFileSync } from "node:fs";
import { parse } from "yaml";

/** ワークフローのジョブとしての仕組み． */
export type Mechanism = { workflow: string; job: string };

/** 運用項目．ops/schema/items.schema.json の形に合う． */
export type Item = {
  id: string;
  title: string;
  category: "業務運用" | "基盤運用" | "運用管理";
  requirements: string[];
  mechanisms: Mechanism[];
  evidence: string;
  timing: string;
  actor: "自動" | "人";
  runbook?: string;
};

export type ItemsFile = { items: Item[] };

/** YAMLのファイルを読む． */
export function readYaml(path: string): unknown {
  return parse(readFileSync(path, "utf8"));
}
