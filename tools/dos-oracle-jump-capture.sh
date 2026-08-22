#!/usr/bin/env bash
# 用「自製存檔」把原版直接放到指定的格子上，再拍下四個朝向。
#
# ★ 為什麼不用走的。 走路只到得了「從存檔點連得過去」的地方，而且每一步都可能
# 被 ECL 事件插隊推歪。自製存檔是主選單按 L 就到，一步都不用走。
#
# ⚠ 只換得到**同一張圖裡的**格子。`-map-block` 換不到別張圖：原版的第一人稱
# 地圖由存檔裡的 ECL 狀態決定（第 675 輪實測，見 cmd/dos-save-export 的註解）。
# 所以底檔在哪張圖，這支就只拍得到那張圖。
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
  (cd "$ROOT" && tools/go.sh run ./cmd/dos-save-export -base "$base" -out "$SAVE" -slot A -area 2 -x "$x" -y "$y" -facing 0 >/dev/null)
  "$S" start >/dev/null
  for _ in $(seq 20); do screen | grep -q "PLAY DEMO" && break; sleep 1; done
  gate Escape "CHOOSE A FUNCTION" || { echo "  主選單沒出現，跳過"; continue; }
  gate l "LOAD WHICH GAME"        || { echo "  載入選單沒出現，跳過"; continue; }
  gate a "BEGIN ADVENTURING\|CHOOSE A FUNCTION" || { echo "  載入沒完成，跳過"; continue; }
  gate b "AREA CAST VIEW"         || { echo "  沒進到冒險畫面，跳過"; continue; }

  # ⚠ 位置一定要從畫面讀回來核對。存檔寫進去不等於原版照著站——不核對的話
  # 索引會標成目標格，實際拍的是別格，而畫面本身看起來完全正常。
  got="$(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/_live.png" | cut -f1,2)"
  if [ "$got" != "$(printf '%s\t%s' "$x" "$y")" ]; then
    echo "  ⚠ 畫面讀到 ($got)，不是 ($x,$y)：不收這一格"
    continue
  fi
  for d in N E S W; do
    name="geo2-b01-x$(printf '%02d' "$x")-y$(printf '%02d' "$y")-$d.png"
    "$S" shot "$name" >/dev/null
    read -r rx ry rd < <(python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/$name")
    if [ "$rx" = "$x" ] && [ "$ry" = "$y" ] && [ "$rd" = "$d" ]; then
      printf '%s\t%s\t%s\t%s\n' "$name" "$x" "$y" "$d" >> "$index"
      echo "  收下 $name"
    else
      echo "  ⚠ $name 讀到 ($rx,$ry,$rd)，與預期不符：不收"
    fi
    "$S" key Right 0.6 >/dev/null
  done
done
