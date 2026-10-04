#!/usr/bin/env bash
# 一時的なコピーに，第0回から各回の仕込みと解答のパッチを順に当て，
# 各回の解答で題材の検査(mise run check)が通ることを確かめる．
# 使い方：scripts/verify-iterations.sh [最後の回の番号] [検査を始める回の番号]
# 検査を始める回より前の回は，パッチを当てるだけで検査しない．
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
first=${2:-0}
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
  if [ "$((10#$n))" -ge "$((10#$first))" ]; then
    mise -C app run check
  fi
done
echo "検査したすべての回の解答で，検査が通った．"
