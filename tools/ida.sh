#!/usr/bin/env bash
# IDA Pro 9.4 headless 包裝（本專案唯一入口）。
#
# 為什麼固定用 ida-pro-9.4-idapython:py312-v1：
#   基底 image（ver2／ver3）跑 IDAPython 會「零輸出、零訊息」地失敗，
#   而且 exit code 不可信（同一種失敗在不同 image 分別回 0 與 1）。
#   只有這顆 image 同時修好 libpython3.12 與 idapyswitch 的 $HOME 寫入位置。
#   詳見 ~/.claude/knowledge-base/retro/ida-pro-9.4.md。
#
# 用法：
#   tools/ida.sh analyze <binary> [額外 idat 參數...]
#       在 binary 所在目錄產生 .i64 與 .asm（-A -B）。
#   tools/ida.sh binary16 <raw.bin> [entry_json]
#       以 16-bit 8086、base 0 載入 raw overlay，套用 entry point 後自動分析。
#   tools/ida.sh py <i64_or_binary> <script.py> [script_args...]
#       跑 IDAPython；script 由 tools/ida/ 掛載為唯讀。
#   tools/ida.sh idc <i64_or_binary> <script.idc> [script_args...]
#       跑 IDC（退路；新腳本一律優先寫 IDAPython）。
#   tools/ida.sh raw <idat 參數...>
#       直接傳給 idat，工作目錄是第一個參數所在目錄的父層。
#
# 硬規則：headless 的 print／Message() 不進 stdout，exit code 也不可信。
# 腳本一律把結果寫檔，收工前驗檔案存在且非空。
set -euo pipefail

IMAGE="${IDA_IMAGE:-ida-pro-9.4-idapython:py312-v1}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

die() { echo "ida.sh: $*" >&2; exit 2; }

run_in() {
  # $1 = 掛載為 /work 的主機目錄，其餘為 idat 參數
  local work="$1"; shift
  [ -d "$work" ] || die "工作目錄不存在：$work"
  docker run --rm \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$work:/work" \
    -v "$ROOT/tools/ida:/work-tools:ro" \
    -w /work \
    "$IMAGE" idat "$@"
}

cmd="${1:-}"; shift || true
case "$cmd" in
  analyze)
    target="${1:?用法: tools/ida.sh analyze <binary>}"; shift
    [ -f "$target" ] || die "找不到檔案：$target"
    dir="$(cd "$(dirname "$target")" && pwd)"
    run_in "$dir" -A -B "$@" "$(basename "$target")"
    ;;
  binary16)
    target="${1:?用法: tools/ida.sh binary16 <raw.bin> [entry_json]}"; shift
    [ -f "$target" ] || die "找不到檔案：$target"
    dir="$(cd "$(dirname "$target")" && pwd)"
    base="$(basename "$target")"
    entry="${1:-}"
    # -p 指定 8086、-b0 指定 base paragraph 0、-B 批次；
    # raw binary 沒有 entry point，交給 seed_binary16.py 依 entry_json 標記後再自動分析。
    run_in "$dir" -A -B -p8086 -b0 "$base"
    if [ -n "$entry" ]; then
      run_in "$dir" -A "-S/work-tools/seed_binary16.py $entry" "$base.i64"
    fi
    ;;
  py|idc)
    target="${1:?用法: tools/ida.sh $cmd <i64|binary> <script>}"; shift
    script="${1:?缺少 script}"; shift
    [ -f "$target" ] || die "找不到檔案：$target"
    [ -f "$ROOT/tools/ida/$(basename "$script")" ] || die "script 必須放在 tools/ida/：$script"
    dir="$(cd "$(dirname "$target")" && pwd)"
    run_in "$dir" -A "-S/work-tools/$(basename "$script") $*" "$(basename "$target")"
    ;;
  raw)
    target="${1:?用法: tools/ida.sh raw <idat 參數...>}"
    dir="$(pwd)"
    run_in "$dir" "$@"
    ;;
  *)
    sed -n '3,30p' "${BASH_SOURCE[0]}"
    exit 2
    ;;
esac
