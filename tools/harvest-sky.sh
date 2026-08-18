#!/usr/bin/env bash
# 收一張第一人稱地圖的天空色：把原版存檔改到指定的章節與地圖區塊，載入、
# 進遊戲、紮營存檔，再把 Area1 的 1FAh／1FCh 讀回來。
#
# ★ 為什麼可以這樣量：那兩個欄位**不是純存檔狀態**。把它們清成 0 之後載入、
# 走一段再存回來，值會變回原來的 11／9 ——原版每次進地圖都會重算。
#
# 用法：tools/harvest-sky.sh <base 存檔> <area> <block> [x] [y]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/tools/dos-oracle-session.sh"
SAVE="$ROOT/workplace/dos-oracle/game/SAVE"
RELSAVE="workplace/dos-oracle/game/SAVE"
base="$1"; area="$2"; block="$3"; x="${4:-7}"; y="${5:-7}"

rm -f "$SAVE"/savgam*.dat "$SAVE"/SAVGAM*.DAT
(cd "$ROOT" && ./tools/go.sh run ./cmd/dos-save-export -base "$base" -out "$RELSAVE" \
  -slot A -area "$area" -map-block "$block" -x "$x" -y "$y" -facing 2) >/dev/null

# ⚠ 一律用 wait／keyuntil，不要用固定延遲：開機時間會漂，漂掉的那一鍵
# 會讓後面每一步都錯位，而錯位的鍵可能剛好按到 EXIT TO DOS。
"$S" stop >/dev/null 2>&1 || true
"$S" start 8 >/dev/null
"$S" keyuntil Return "CHOOSE A FUNCTION" 12 >/dev/null
"$S" key l 0.6 >/dev/null
"$S" wait "LOAD WHICH GAME" >/dev/null
"$S" key a 0.6 >/dev/null
"$S" wait "BEGIN ADVENTURING" >/dev/null
"$S" key b 0.6 >/dev/null
"$S" wait "ENCAMP" >/dev/null
"$S" shot "sky-a${area}-b${block}.png" >/dev/null
"$S" key e 0.6 >/dev/null
"$S" wait "CAMP:SAVE" >/dev/null
"$S" key s 0.6 >/dev/null
"$S" wait "SAVE WHICH GAME" >/dev/null
"$S" key c 0.6 >/dev/null
"$S" wait "QUIT TO DOS" >/dev/null
"$S" key n 0.6 >/dev/null
"$S" stop >/dev/null

python3 - "$SAVE/SAVGAMC.DAT" "$area" "$block" <<'PY'
import struct, sys
path, area, block = sys.argv[1], sys.argv[2], sys.argv[3]
data = open(path, "rb").read()
a1 = data[1:1 + 0x800]
outdoor = struct.unpack("<H", a1[0x1FA:0x1FC])[0]
indoor = struct.unpack("<H", a1[0x1FC:0x1FE])[0]
print("area=%s block=%s GameArea=%d MapBlockID=%d outdoor=%d indoor=%d 位置=%s" % (
    area, block, data[0], a1[0x18A], outdoor, indoor, data[12801:12806].hex()))
PY
