#!/usr/bin/env bash
# 以 `55 89 e5` prologue 為界匯出每個模組的函式（取代相信 IDA 的函式清單）。
#
# 為什麼換掉舊的 full 匯出：IDA 會把一支 Turbo Pascal 函式切成好幾塊，而且
# 有些位元組是 code 卻不屬於任何函式的 chunk，於是完全不出現在匯出裡。
# `overlay-02:149Ch` 的 IDA size 是 50 bytes，真實長度 544——差 11 倍，
# 逐條讀時看到的是一支空殼，沒有任何警告。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SWEEP="$ROOT/workplace/re-sweep"
JOBS="${JOBS:-3}"

one() {
  local plat="$1" m="$2" W="$SWEEP/$1/overlays"
  local out="$W/prologue/$plat-$m.json"
  [ -s "$out" ] && { echo "skip $plat $m"; return 0; }
  mkdir -p "$W/prologue"
  "$ROOT/tools/ida.sh" py "$W/$m.bin.i64" export_by_prologue.py \
    "/work/prologue/$plat-$m.json" >/dev/null 2>&1 || true
  if [ -s "$out" ]; then
    echo "ok   $plat $m $(python3 -c "import json;print(len(json.load(open('$out'))['functions']))")"
  else echo "FAIL $plat $m"; fi
}
export -f one; export SWEEP ROOT

for plat in dos pc98; do
  for f in "$SWEEP/$plat/overlays"/overlay-*.bin; do
    [ -e "$f" ] && printf '%s\t%s\n' "$plat" "$(basename "$f" .bin)"
  done
done | xargs -P "$JOBS" -n 1 -d '\n' bash -c 'IFS=$'"'"'\t'"'"' read -r p m <<<"$0"; one "$p" "$m"'
