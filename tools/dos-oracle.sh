#!/usr/bin/env bash
# 在 Docker/Xvfb 裡跑原版 DOS 版並擷取畫面（原版行為 oracle）。
#
# 用法：tools/dos-oracle.sh <輸出檔名> [開機等待秒] ["鍵:延遲 ..."] [張數] [間隔秒]
#   tools/dos-oracle.sh title.png 8
#   tools/dos-oracle.sh demo.png 8 "d:6" 14 4        # 示範模式連拍 14 張
#
# 輸出落在 workplace/dos-oracle/out/；遊戲檔案由映像 zip 解到
# workplace/dos-oracle/game/（可寫，遊戲會在裡面建 CURSE.CFG 與存檔）。
#
# ★ 進入方式：`START.EXE STING Wooden` 是遊戲自己的測試模式，繞過翻譯輪
# （spec 530）。取證等級因此是 `cheat-assisted`，不能當「正常玩家路徑」。
#
# ⚠ 三個踩過的坑：
#   1. **沒有 CURSE.CFG 時**開機會先問顯示介面卡與存檔路徑，兩題都在**文字模式**；
#      實測按鍵在文字模式送不進去（3 秒與 15 秒都試過，每次送鍵前重新搜尋視窗
#      並重設焦點也一樣）。有 CURSE.CFG 時直接進圖形模式，按鍵就正常。
#      ⇒ 目前只能拿到 CGA（4 色）畫面；EGA 要嘛解出 CURSE.CFG 的格式
#      （字串在 `GAME.OVR` 偏移 114766 的 `\x05CURSE`），要嘛解決文字模式輸入。
#   2. 送鍵一律走 XTEST（`xdotool key` 不加 `--window`）：SDL 會忽略 XSendEvent
#      的合成事件。沒有視窗管理員時 `windowactivate` 會失敗，要用 `windowfocus`。
#   3. 示範模式（`d`）走的是**示範專用地圖**，地名是 `NOWHERE IN THE REALMS`，
#      不是提爾佛頓街道——`tilverton-first-person-demo.png` 的檔名會誤導。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$ROOT/workplace/dos-oracle"
IMAGE="${DOS_ORACLE_IMAGE:-dosbox-run:latest}"
OUT="$1"; BOOT="${2:-8}"; SEQ="${3:-}"; FRAMES="${4:-1}"; INTERVAL="${5:-3}"

mkdir -p "$WORK/game" "$WORK/out"
if [ ! -f "$WORK/game/START.EXE" ]; then
  python3 - "$ROOT/curseoftheazurebonds.zip" "$WORK/game" <<'PY'
import sys, zipfile
zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])
PY
  mkdir -p "$WORK/game/SAVE"
fi
# KEEP_CFG=1 保留上一次的設定；預設砍掉是為了每次都從同一個狀態開始。
[ "${KEEP_CFG:-}" = "1" ] || rm -f "$WORK/game/CURSE.CFG"

exec docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$WORK/game:/game" -v "$WORK/out:/out" \
  -v "$ROOT/tools/dos-oracle-capture.sh:/capture.sh:ro" \
  -e HOME=/tmp \
  "$IMAGE" bash /capture.sh "/out/$OUT" "$BOOT" "$SEQ" "$FRAMES" "$INTERVAL"
