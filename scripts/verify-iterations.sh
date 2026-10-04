#!/usr/bin/env bash
# 一時的なコピーに，第0回から各回の仕込みと解答のパッチを順に当て，
# 各回の解答で題材の検査(mise run check)が通ることを確かめる．
# 使い方：scripts/verify-iterations.sh [最後の回の番号]
set -euo pipefail

root=$(git rev-parse --show-toplevel)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 作業中のファイル(未コミットを含む)を，題材の開始時点としてコピーする．
cd "$root"
git ls-files -z --cached --others --exclude-standard | tar --null -T - -cf - | tar -xf - -C "$work"
cd "$work"
git init -q
git add -A
git -c user.name=verify -c user.email=verify@example.com commit -qm base
mise trust -q . && mise trust -q app

last=${1:-}
for dir in iterations/*/; do
  n=$(basename "$dir")
  if [ -n "$last" ] && [ "$((10#$n))" -gt "$((10#$last))" ]; then
    break
  fi
  echo "== 第$((10#$n))回 =="
  if [ -f "$dir/problems.patch" ]; then
    git apply --whitespace=nowarn "$dir/problems.patch"
  fi
  git apply --whitespace=nowarn "$dir/solution.patch"
  git add -A
  git -c user.name=verify -c user.email=verify@example.com commit -qm "iteration $n"
  mise -C app run check
done
echo "すべての回の解答で検査が通った．"
