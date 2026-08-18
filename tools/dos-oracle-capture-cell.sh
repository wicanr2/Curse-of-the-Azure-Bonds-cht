#!/usr/bin/env bash
# 在互動 session 裡把目前這一格的四個朝向都拍下來，並把座標記進索引檔。
# 遇到 `PRESS <ENTER>` 之類的提示會自動按 Enter 消掉——原版的 ECL 事件
# 會在踏進某些格子時插進來，固定序列的走法會被它整個推歪。
#
# 用法：tools/dos-oracle-capture-cell.sh <前綴> <索引檔>
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
prefix="$1"; index="$2"
S="$ROOT/tools/dos-oracle-session.sh"

dismiss() {
  for _ in 1 2 3 4 5 6; do
    "$S" shot "_probe.png" >/dev/null
    if python3 "$ROOT/tools/dos_screen.py" "$ROOT/workplace/dos-oracle/out/_probe.png" \
        | grep -q "PRESS <ENTER>"; then
      "$S" key Return 1.0 >/dev/null
    else
      return 0
    fi
  done
}

dismiss
for d in 0 1 2 3; do
  name="$prefix-$d.png"
  "$S" shot "$name" >/dev/null
  printf '%s\t%s\n' "$name" \
    "$(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/$name")" >> "$index"
  "$S" key Right 0.6 >/dev/null
  dismiss
done
