#!/usr/bin/env bash
# 全模組 RE 全掃：把一個平台的 resident executable 與全部 TPOV overlay
# 建成 IDA database 並匯出結構 JSON。
#
# 這是「先徹底盤點、再逐系統語意閉合」的第一步。它只產生 IDA 已知的事實，
# 不做任何語意判斷。
#
# 用法：
#   tools/re-sweep.sh dos    workplace/ida-dos-probe/START.EXE workplace/ida-dos-probe/GAME.OVR
#   tools/re-sweep.sh pc98   workplace/ida406/PC98-GAME.EXE    workplace/ida406/PC98-GAME.OVR
#
# 產出（全部在 workplace/re-sweep/<platform>/，不進版控）：
#   ovr-manifest.json         overlay 結構清冊（entry stub → handler offset）
#   overlays/overlay-NN.bin   每段 overlay 的 raw code
#   out/<module>.json         每個模組的函式／xref／字串／未定義區
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${IDA_IMAGE:-ida-pro-9.4-idapython:py312-v1}"
JOBS="${JOBS:-4}"

platform="${1:?用法: tools/re-sweep.sh <dos|pc98> <exe> <ovr>}"
exe="${2:?缺少 resident executable}"
ovr="${3:?缺少 TPOV overlay 容器}"

work="$ROOT/workplace/re-sweep/$platform"
mkdir -p "$work/overlays" "$work/out" "$work/log"

ida() {
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$work:/work" -v "$ROOT/tools/ida:/work-tools:ro" -w /work \
    "$IMAGE" idat "$@"
}
export -f ida 2>/dev/null || true

echo "[1/4] overlay manifest"
# -o 是容器內路徑（repo 掛在 /src），所以只能給 repo 相對路徑
(cd "$ROOT" && tools/go.sh build -o workplace/re-sweep/ovr-manifest ./cmd/ovr-manifest)
test -x "$ROOT/workplace/re-sweep/ovr-manifest" || { echo "ovr-manifest 建置失敗"; exit 1; }
"$ROOT/workplace/re-sweep/ovr-manifest" -exe "$exe" -ovr "$ovr" -platform "$platform" \
  -code-dir "$work/overlays" -out "$work/ovr-manifest.json"

echo "[2/4] resident executable"
resident="$(basename "$exe")"
cp -f "$exe" "$work/$resident"
rm -f "$work/$resident".i64 "$work/$resident".id? "$work/$resident".nam "$work/$resident".til
ida -A -B "$resident" >"$work/log/$resident.build.log" 2>&1 || true
ida -A "-S/work-tools/export_module.py /work/out/$resident.json" "$resident.i64" \
  >"$work/log/$resident.export.log" 2>&1 || true
test -s "$work/out/$resident.json" || { echo "resident 匯出失敗，見 $work/out/$resident.json.error.log"; exit 1; }

# Borland 除錯符號（目前只有 PC-98 有）可以再補一批 code offset 種子。
SEEDS_ARG=""
if [ -f "$work/seeds.json" ]; then
  SEEDS_ARG="/work/seeds.json"
  echo "     使用額外種子 $work/seeds.json"
fi
export SEEDS_ARG

echo "[3/4] overlays（JOBS=$JOBS）"
modules="$(python3 -c "
import json
data = json.load(open('$work/ovr-manifest.json'))
print('\n'.join(o['module'] for o in data['overlays']))
")"

run_overlay() {
  local module="$1"
  local bin="overlays/$module.bin"
  rm -f "$work/$bin".i64 "$work/$bin".id? "$work/$bin".nam "$work/$bin".til
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$work:/work" -v "$ROOT/tools/ida:/work-tools:ro" -w /work \
    "$IMAGE" idat -A -p8086 -b0 \
    "-S/work-tools/analyze_overlay.py /work/ovr-manifest.json $module /work/out/$module.json $SEEDS_ARG" \
    "$bin" >"$work/log/$module.log" 2>&1 || true
  if [ -s "$work/out/$module.json" ]; then echo "ok   $module"; else echo "FAIL $module"; fi
}
export -f run_overlay
export work ROOT IMAGE

echo "$modules" | xargs -P "$JOBS" -I{} bash -c 'run_overlay "$@"' _ {}

echo "[4/4] 彙總"
python3 - "$work" <<'PY'
import glob, json, os, sys
work = sys.argv[1]
rows, missing = [], []
for path in sorted(glob.glob(os.path.join(work, "out", "*.json"))):
    if path.endswith(".error.log"):
        continue
    with open(path, encoding="utf-8") as handle:
        data = json.load(handle)
    ov = data.get("overlay")
    rows.append({
        "module": ov["module"] if ov else data["input"]["name"],
        "functions": data["totals"]["functions"],
        "instructions": data["totals"]["instructions"],
        "code_bytes": data["totals"]["code_bytes"],
        "seeded": len(ov["seeded_entries"]) if ov else None,
        "failed_entries": len(ov["failed_entries"]) if ov else None,
        "segment_bytes": sum(s["end"] - s["start"] for s in data["segments"]),
        "undefined_bytes": sum(r["end"] - r["start"] for r in data["undefined_ranges"]),
    })
for e in sorted(glob.glob(os.path.join(work, "out", "*.error.log"))):
    missing.append(os.path.basename(e))
summary = {"modules": rows, "errors": missing,
           "totals": {
               "modules": len(rows),
               "functions": sum(r["functions"] for r in rows),
               "instructions": sum(r["instructions"] for r in rows),
               "code_bytes": sum(r["code_bytes"] for r in rows),
               "segment_bytes": sum(r["segment_bytes"] for r in rows),
               "undefined_bytes": sum(r["undefined_bytes"] for r in rows),
           }}
with open(os.path.join(work, "sweep-summary.json"), "w", encoding="utf-8") as handle:
    json.dump(summary, handle, ensure_ascii=False, indent=1)
print(json.dumps(summary["totals"], ensure_ascii=False))
if missing:
    print("errors:", missing)
PY
