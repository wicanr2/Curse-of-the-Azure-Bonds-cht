#!/usr/bin/env bash
# 從指定格子、指定朝向往前走一步，回報走不走得動。
#
# ★ 存在的理由：要判「原版現在站在哪一張圖」，需要一把**與 remake 的算圖無關**
# 的尺。畫面比對是循環論證；AREA 自動地圖只有 63.3% 對得上 GEO（spec 1185）。
# 這一支只用原版的**行為**：走得動還是走不動，把座標讀回來就知道。
#
# ⚠ 每個方向都**重新載一次存檔**，朝向直接寫進存檔。不在遊戲裡轉向再走：
# 轉向與走路的按鍵都可能被吞掉，而誤差會累積——實測「走一步再走回來」會漂到
# 隔壁格，然後整格的探測就廢了。重載慢，但每一次都是乾淨的起點。
#
# 輸出：`x y 朝向 0|1`（1 ＝ 走得動）。
#
# 用法：AREA=2 ECL_BLOCK=1 GEO_BLOCK=1 tools/dos-oracle-move-probe.sh <底檔> <x> <y> <朝向 N|E|S|W>
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/tools/dos-oracle-session.sh"
base="$1"; x="$2"; y="$3"; face="$4"
SAVE="workplace/dos-oracle/game/SAVE"
AREA="${AREA:-2}"; ECL_BLOCK="${ECL_BLOCK:-1}"; GEO_BLOCK="${GEO_BLOCK:-1}"
case "$face" in N) fnum=0;; E) fnum=2;; S) fnum=4;; W) fnum=6;; *) echo "朝向要是 N/E/S/W"; exit 2;; esac

pos() { "$S" shot "_probe.png" >/dev/null; python3 "$ROOT/tools/dos_screen_pos.py" "$ROOT/workplace/dos-oracle/out/_probe.png"; }
screen() { "$S" text 2>/dev/null | head -30; }
gate() { "$S" key "$1" 1.2 >/dev/null; for _ in $(seq 12); do screen | grep -q "$2" && return 0; sleep 1; done; return 1; }

(cd "$ROOT" && tools/go.sh run ./cmd/dos-save-export -base "$base" -out "$SAVE" -slot A \
  -area "$AREA" -ecl-block "$ECL_BLOCK" -map-block "$GEO_BLOCK" -x "$x" -y "$y" -facing "$fnum" >/dev/null 2>&1)
"$S" start >/dev/null
for _ in $(seq 20); do screen | grep -q "PLAY DEMO" && break; sleep 1; done
gate Escape "CHOOSE A FUNCTION" >/dev/null || { echo "$x $y $face FAIL"; exit 0; }
gate l "LOAD WHICH GAME" >/dev/null        || { echo "$x $y $face FAIL"; exit 0; }
gate a "BEGIN\|CHOOSE A FUNCTION" >/dev/null || { echo "$x $y $face FAIL"; exit 0; }
gate b "AREA CAST VIEW" >/dev/null          || { echo "$x $y $face FAIL"; exit 0; }

read -r cx cy cd < <(pos)
if [ "$cx" != "$x" ] || [ "$cy" != "$y" ] || [ "$cd" != "$face" ]; then
  echo "$x $y $face FAIL(讀到 $cx,$cy,$cd)"; exit 0
fi
"$S" key Up 1.2 >/dev/null
read -r nx ny nd < <(pos)
if [ "$nx" = "$x" ] && [ "$ny" = "$y" ]; then echo "$x $y $face 0"; else echo "$x $y $face 1 → ($nx,$ny)"; fi
