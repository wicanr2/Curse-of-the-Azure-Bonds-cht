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
# ⚠ 踩過的坑，每一條都花過好幾輪：
#   1. **按鍵要按住再放開**（`keydown` → `sleep 0.25` → `keyup`）。瞬時的
#      `xdotool key` 在開機那幾題文字模式提示上會整批被吃掉，看起來就像
#      「文字模式送不進按鍵」。改成按住之後三道提示全部答得進去。
#   2. 送鍵一律走 XTEST（`xdotool key` 不加 `--window`）：SDL 會忽略 XSendEvent
#      的合成事件。沒有視窗管理員時 `windowactivate` 會失敗，要用 `windowfocus`。
#   3. **開機三道提示**（沒有 `CURSE.CFG` 時才會出現）：顯示介面卡
#      `[1]CGA [2]EGA [3]Tandy` → 音效 `[1]PC [2]Tandy [3]Silent` → 存檔路徑。
#      答 `2` `3` `Return` 就進 EGA。⚠ `CURSE.CFG` **不存這些答案**（實測永遠
#      0 byte），它只是一個「已設定過」的旗標——留著它就會用預設值 CGA 起動，
#      所以本工具預設每次都刪掉它（`KEEP_CFG=1` 可保留）。
#   4. 判斷畫面**不要只看顏色數**。顯示介面卡那題與音效那題的像素統計幾乎一樣，
#      只看數字會誤判成「按鍵沒進去」；要把圖叫出來看。
#   5. 示範模式（`d`）走的是**示範專用地圖**，地名是 `NOWHERE IN THE REALMS`，
#      而且那一幕是 **PIC（夜營畫面）不是第一人稱視野**——
#      `docs/reference/original-dos/tilverton-first-person-demo.png` 的檔名會誤導。
#
# 已走通的選單路徑（`frames=0` 是逐鍵軌跡模式，一輪就看得到每一步）：
#   開機三題 → 標題列（Enter ＝ PLAY）→ `CHOOSE A FUNCTION`
#   → `c` 建角 → Enter×2 吃預設 → `REROLL STATS?` Enter
#   → `CHARACTER NAME:` 打字 ＋ Enter → 圖示編輯 → `e` → `IS THIS ICON OK?` `y`
#   → `SAVE 名字?` `y` → 確認 → `NEW FILE NAME:` 打字 ＋ Enter → 回主選單
#   → `a` → `ADD FROM WHERE?`（CURSE）→ 角色清單 → 選角色 → 離開
#
# ⚠ **定時送鍵的序列很脆**：任何一步的載入時間漂移都會讓後面每一個鍵錯位，
# 而一個錯位的鍵可能剛好按到 `EXIT TO DOS`。要穩定重現得改成**看畫面決定下一鍵**
# （底部選單列的像素指紋就夠當狀態判別），不能一直加長固定延遲。
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
  -e HOME=/tmp -e COMMAND="${COMMAND:-}" \
  "$IMAGE" bash /capture.sh "/out/$OUT" "$BOOT" "$SEQ" "$FRAMES" "$INTERVAL"
