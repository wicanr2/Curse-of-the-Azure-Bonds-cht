#!/usr/bin/env bash
# 用 IDAPython 把 prologue 匯出裡的 undefined 缺口強制解成指令後重新匯出。
#
# 為什麼需要這一步：`export_by_prologue.py` 只匯出 `is_code` 為真的位元組，
# IDA 沒認出來的那幾個 byte 會整段消失，而相鄰指令的 `ea` 中間少了幾格**在
# 逐條讀的時候看不出來**。實測 `pc98/overlay-22:048AAh` 匯出只有 14 條指令，
# 補完是 77 條——少了八成，沒有任何警告。
#
# 輸出寫到 `<平台>/overlays/filled/`，**不覆蓋原本的 prologue 匯出**，方便
# 逐支比對「補完之後指令數變了幾條」，回頭稽核已經判讀過的函式。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SWEEP="$ROOT/workplace/re-sweep"
JOBS="${JOBS:-3}"

one() {
  local plat="$1" m="$2" W="$SWEEP/$1/overlays"
  local out="$W/filled/$plat-$m.json"
  [ -s "$out" ] && { echo "skip $plat $m"; return 0; }
  mkdir -p "$W/filled"
  "$ROOT/tools/ida.sh" py "$W/$m.bin.i64" fill_gaps_and_export.py \
    "/work/filled/$plat-$m.json" >/dev/null 2>&1 || true
  if [ -s "$out" ]; then
    echo "ok   $plat $m $(python3 -c "import json;d=json.load(open('$out'));print(len(d['functions']),'函式,補',len(d.get('filled',[])),'段')")"
  else echo "FAIL $plat $m"; fi
}
export -f one; export SWEEP ROOT

for plat in dos pc98; do
  for f in "$SWEEP/$plat/overlays"/overlay-*.bin; do
    [ -e "$f" ] && printf '%s\t%s\n' "$plat" "$(basename "$f" .bin)"
  done
done | xargs -P "$JOBS" -n 1 -d '\n' bash -c 'IFS=$'"'"'\t'"'"' read -r p m <<<"$0"; one "$p" "$m"'
