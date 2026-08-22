#!/usr/bin/env bash
# 用「自製存檔」把原版直接放到指定的格子上，再拍下四個朝向。
#
# ★ 為什麼不用走的。 走路只到得了「從存檔點連得過去」的地方，而且每一步都可能
# 被 ECL 事件插隊推歪。自製存檔是主選單按 L 就到，一步都不用走。
#
# 換圖用環境變數 AREA／ECL_BLOCK／GEO_BLOCK／PREFIX（第 678 輪起）。原版的第一人稱
# 地圖是**當前 ECL 段的進入碼**跑 `LOAD FILES` 決定的，不是存檔裡某個編號決定的，
# 所以換圖＝把那一段的位元組碼換進存檔的程式碼視窗（`dos-save-export -ecl-block`）。
#
# ⚠ 每一格都要重開一次遊戲（載入存檔只在主選單做得到），約 40 秒。
#
# 用法：tools/dos-oracle-jump-capture.sh <底檔> <索引檔> <x> <y> [x y ...]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/tools/dos-oracle-session.sh"
base="$1"; index="$2"; shift 2
# ⚠ 路徑要給**相對**的：tools/go.sh 把 repo 掛在容器的 /src，主機的絕對路徑
# 在容器裡不存在，工具會去建 /home/... 然後以 permission denied 收場。
SAVE="workplace/dos-oracle/game/SAVE"
# 預設是提爾佛頓（底檔本來就在那張圖）。換圖時三個都要給，PREFIX 是檔名前綴。
AREA="${AREA:-2}"
ECL_BLOCK="${ECL_BLOCK:-1}"
GEO_BLOCK="${GEO_BLOCK:-1}"
PREFIX="${PREFIX:-geo2-b01}"

screen() { "$S" text 2>/dev/null | head -30; }

# gate <鍵> <期待的畫面字串>：送鍵之後等畫面出現該字串，逾時就回非零。
gate() {
  "$S" key "$1" 1.2 >/dev/null
  for _ in $(seq 12); do
    if screen | grep -q "$2"; then return 0; fi
    sleep 1
  done
  return 1
}

while [ "$#" -ge 2 ]; do
  x="$1"; y="$2"; shift 2
  echo "=== 目標 ($x,$y) ==="
  (cd "$ROOT" && tools/go.sh run ./cmd/dos-save-export -base "$base" -out "$SAVE" -slot A \
    -area "$AREA" -ecl-block "$ECL_BLOCK" -map-block "$GEO_BLOCK" \
    -x "$x" -y "$y" -facing 0 >/dev/null 2>&1)
  "$S" start >/dev/null
  for _ in $(seq 20); do screen | grep -q "PLAY DEMO" && break; sleep 1; done
  gate Escape "CHOOSE A FUNCTION" || { echo "  主選單沒出現，跳過"; continue; }
  gate l "LOAD WHICH GAME"        || { echo "  載入選單沒出現，跳過"; continue; }
  gate a "BEGIN ADVENTURING\|CHOOSE A FUNCTION" || { echo "  載入沒完成，跳過"; continue; }
  gate b "AREA CAST VIEW"         || { echo "  沒進到冒險畫面，跳過"; continue; }

  # ⚠ 位置一定要從畫面讀回來核對。存檔寫進去不等於原版照著站——不核對的話
  # 索引會標成目標格，實際拍的是別格，而畫面本身看起來完全正常。
  # ⚠ 不是每一張圖都顯示座標：提爾佛頓是 `7,13 N 00:00`，GEO3 段 0x15 那類只有
  # `N 00:00`。這種圖只核對得到朝向，座標只能靠存檔——而存檔的座標是驗過會生效的
  # （提爾佛頓與下水道共 289 張逐格對得上）。分開處理，不要把「這張圖不顯示座標」
  # 與「畫面根本不對」混成同一種。
  read -r gx gy gd < <(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/_live.png")
  if [ "$gx" = "?" ] && [ "$gd" != "?" ]; then
    echo "  （這張圖不顯示座標，只核對朝向）"
  elif [ "$gx" != "$x" ] || [ "$gy" != "$y" ]; then
    echo "  ⚠ 畫面讀到 ($gx,$gy)，不是 ($x,$y)：不收這一格"
    continue
  fi
  for d in N E S W; do
    name="$PREFIX-x$(printf '%02d' "$x")-y$(printf '%02d' "$y")-$d.png"
    "$S" shot "$name" >/dev/null
    read -r rx ry rd < <(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/$name")
    if { { [ "$rx" = "$x" ] && [ "$ry" = "$y" ]; } || [ "$rx" = "?" ]; } && [ "$rd" = "$d" ]; then
      printf '%s\t%s\t%s\t%s\n' "$name" "$x" "$y" "$d" >> "$index"
      echo "  收下 $name"
    else
      echo "  ⚠ $name 讀到 ($rx,$ry,$rd)，與預期不符：不收"
    fi
    "$S" key Right 0.6 >/dev/null
  done
done
