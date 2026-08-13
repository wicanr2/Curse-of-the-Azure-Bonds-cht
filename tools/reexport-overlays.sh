#!/usr/bin/env bash
# 重新產生 out/<module>.json，讓 seed_missing 補進去的函式進到台帳分母。
#
# 兩個踩過的坑，都會**安靜地**給出錯的結果：
#
# 1. 輸出路徑必須落在掛載點內。ida.sh 把「i64 所在目錄」掛成 /work，所以
#    `/work/../out/x.json` 在容器裡是 `/out`——不存在，寫入被丟棄，out/ 的
#    時間戳完全不動，看起來像「跑完了但沒變」。
# 2. **不要再讓 IDA 重新分析。** 把 `overlays/x.bin.i64` 交給 `idat` 會從
#    `x.bin` 重新載入並覆蓋資料庫，`seed_missing_functions.py` 補上的函式
#    就沒了（overlay-25 會從 15 個變回 8 個）。這裡只跑 export_module.py，
#    純讀取。
#
# `overlay` 欄位（模組名、entry 種子、code_size…）由 analyze_overlay.py 產生，
# export_module.py 沒有；少了它 re-ledger 會把模組名認成 `overlay-25.bin`，
# 台帳整批 join 不上。改在主機端從 ovr-manifest.json 補回去。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SWEEP="$ROOT/workplace/re-sweep"
JOBS="${JOBS:-5}"

one() {
  local plat="$1" m="$2" W="$SWEEP/$1"
  "$ROOT/tools/ida.sh" py "$W/overlays/$m.bin.i64" export_module.py \
    "/work/$m.export.json" >"$W/log/$m.reexport.log" 2>&1 || true
  if [ -s "$W/overlays/$m.export.json" ]; then
    python3 "$ROOT/scripts/attach_overlay_field.py" "$plat" "$m" \
      "$W/overlays/$m.export.json" "$W/out/$m.json"
    rm -f "$W/overlays/$m.export.json"
    echo "ok   $plat $m $(python3 -c "import json;print(len(json.load(open('$W/out/$m.json'))['functions']))")"
  else
    echo "FAIL $plat $m"
  fi
}
export -f one; export SWEEP ROOT

for plat in dos pc98; do
  for f in "$SWEEP/$plat/overlays"/overlay-*.bin; do
    [ -e "$f" ] && printf '%s\t%s\n' "$plat" "$(basename "$f" .bin)"
  done
done | xargs -P "$JOBS" -n 1 -d '\n' bash -c 'IFS=$'"'"'\t'"'"' read -r p m <<<"$0"; one "$p" "$m"'
