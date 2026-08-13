#!/usr/bin/env bash
# 把兩平台每個模組的「全部函式完整逐指令 body」匯出。
#
# 與 small 匯出的差別只有門檻：small 用 48 bytes（挑得出「一眼看完」的），
# 這裡用 4096 bytes ≒ 全部。目的是讓逐條閱讀與跨平台助憶碼配對不再受工具
# 門檻限制——48 bytes 以上的函式佔待解讀的 93%。
#
# 輸出：<i64 所在目錄>/full/<模組>.json（ida.sh 把 i64 所在目錄掛成 /work）。
# 併發刻意壓在 5：這台是共用機器，14 核但常態 load 已有 4~5。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SWEEP="$ROOT/workplace/re-sweep"
LIMIT="${LIMIT:-4096}"
JOBS="${JOBS:-5}"

one() {
  local i64="$1" name="$2"
  local dir out
  dir="$(dirname "$i64")"; out="$dir/full/$name.json"
  [ -s "$out" ] && { echo "skip $name"; return 0; }
  mkdir -p "$dir/full"
  "$ROOT/tools/ida.sh" py "$i64" export_small_functions.py "/work/full/$name.json" "$LIMIT" >/dev/null 2>&1 || true
  if [ -s "$out" ]; then echo "ok   $name $(stat -c%s "$out")"; else echo "FAIL $name"; fi
}
export -f one; export ROOT SWEEP LIMIT

{
  for plat in dos pc98; do
    for i64 in "$SWEEP/$plat/overlays"/overlay-*.i64; do
      [ -e "$i64" ] && printf '%s\t%s-%s\n' "$i64" "$plat" "$(basename "$i64" .bin.i64)"
    done
  done
  printf '%s\t%s\n' "$SWEEP/dos/START.EXE.i64" "dos-START.EXE"
  printf '%s\t%s\n' "$SWEEP/PC98-GAME.EXE.i64" "pc98-PC98-GAME.EXE"
} | xargs -P "$JOBS" -n 1 -d '\n' bash -c 'IFS=$'"'"'\t'"'"' read -r f n <<<"$0"; one "$f" "$n"'
