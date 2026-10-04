#!/usr/bin/env bash
# 第N回の仕込み(start)か解答(solution)のパッチを，作業中のコードに当てる．
# 使い方：scripts/iteration.sh start|solution N
set -euo pipefail

usage() {
  echo "使い方：$0 start|solution <回の番号>" >&2
  exit 2
}

[ $# -eq 2 ] || usage
case "$1" in
  start) kind=problems ;;
  solution) kind=solution ;;
  *) usage ;;
esac
n=$(printf '%02d' "$((10#$2))")
root=$(git rev-parse --show-toplevel)
patch="iterations/$n/$kind.patch"

if [ ! -f "$root/$patch" ]; then
  if [ "$kind" = problems ]; then
    echo "第$((10#$n))回に仕込む問題はない．そのまま始める．"
    exit 0
  fi
  echo "$patch がない．" >&2
  exit 1
fi

# 自分の変更と重なる箇所は，3-wayマージで当てる．衝突したら，印の付いた箇所を直す．
git -C "$root" apply --3way --whitespace=nowarn "$patch"
echo "$patch を当てた．"
