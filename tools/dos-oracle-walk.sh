#!/usr/bin/env bash
# 在互動 session 裡走一段路，每一步都擷取並把「座標＋朝向」記進索引檔。
#
# 用法：tools/dos-oracle-walk.sh <前綴> "鍵 鍵 鍵 ..."
#   tools/dos-oracle-walk.sh w "Up Up Right Up"
# 產物：workplace/dos-oracle/out/<前綴>-NN.png ＋ workplace/dos-oracle/<前綴>.tsv
#
# 索引檔每列是 `檔名 x y 朝向`，直接餵給 remake 的
# `-dungeon-x/-dungeon-y/-dungeon-facing` 做逐格比對。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
prefix="$1"; keys="$2"
index="$ROOT/workplace/dos-oracle/$prefix.tsv"
: > "$index"
i=0
record() {
  local name="$prefix-$(printf %02d "$i").png"
  "$ROOT/tools/dos-oracle-session.sh" shot "$name" >/dev/null
  printf '%s\t%s\n' "$name" \
    "$(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/$name")" >> "$index"
  i=$((i+1))
}
record
for k in $keys; do
  "$ROOT/tools/dos-oracle-session.sh" key "$k" 0.7 >/dev/null
  record
done
cat "$index"
