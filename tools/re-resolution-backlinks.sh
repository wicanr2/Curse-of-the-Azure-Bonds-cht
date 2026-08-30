#!/bin/sh
# 驗證「較晚 exact RE 已解出較早 spec 的 unknown」是否確實回填。
# 鍵必須包含平台＋overlay／模組＋位址；單獨的 sub_83 之類不可跨模組配對。
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
LEDGER=${1:-"$ROOT/docs/audit/re-resolution-backlinks.tsv"}
failed=0

tail -n +2 "$LEDGER" | while IFS="$(printf '\t')" read -r platform address semantic evidence older required; do
    test -n "$platform" || continue
    if ! test -f "$ROOT/$evidence"; then
        echo "缺少 evidence spec：$evidence" >&2
        failed=1
    elif ! grep -F "$address" "$ROOT/$evidence" >/dev/null; then
        echo "evidence 沒有原始位址：$platform $address ($evidence)" >&2
        failed=1
    fi
    if ! test -f "$ROOT/$older"; then
        echo "缺少 older spec：$older" >&2
        failed=1
    elif ! grep -F "$required" "$ROOT/$older" >/dev/null; then
        echo "較早 spec 尚未回填：$platform $address $semantic ($older)" >&2
        failed=1
    fi
done

# POSIX pipeline 的 while 可能在 subshell；再做一次不依賴狀態傳遞的總檢。
if tail -n +2 "$LEDGER" | while IFS="$(printf '\t')" read -r platform address semantic evidence older required; do
    test -n "$platform" || continue
    test -f "$ROOT/$evidence" && grep -F "$address" "$ROOT/$evidence" >/dev/null &&
        test -f "$ROOT/$older" && grep -F "$required" "$ROOT/$older" >/dev/null || exit 1
done; then
    echo "RE resolution backlinks: OK"
else
    exit 1
fi
