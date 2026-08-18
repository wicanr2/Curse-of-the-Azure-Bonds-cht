#!/usr/bin/env bash
# 互動式的原版 DOS oracle：容器常駐，主機端逐步送鍵、擷取、讀畫面文字。
#
# ★ 存在的理由：`tools/dos-oracle.sh` 是「一次跑完一整串定時按鍵」，任何一步的
# 載入時間漂移都會讓後面每一個鍵錯位，而一個錯位的鍵可能剛好按到 EXIT TO DOS。
# 這一支把迴圈拆開——送一個鍵、讀一次畫面、再決定下一個鍵，
# 每一步約一秒，取代「猜一整串、等兩分半、看結果」。
#
# 用法：
#   tools/dos-oracle-session.sh start            # 起容器（COMMAND=… 可換命令列）
#   tools/dos-oracle-session.sh key Return       # 送一個鍵（按住 0.25 秒）
#   tools/dos-oracle-session.sh type BOB         # 逐字送
#   tools/dos-oracle-session.sh shot foo.png     # 存到 workplace/dos-oracle/out/
#   tools/dos-oracle-session.sh text             # 擷取後把畫面文字印出來
#   tools/dos-oracle-session.sh stop             # 只停自己這一顆容器
#
# ⚠ 只碰名為 coab-dos-oracle 的容器；不做任何 prune／rmi。
# ⚠ CURSE.CFG 預設保留（開機三題會被跳過、直接進 EGA）；
#    要重問三題就先自己 rm workplace/dos-oracle/game/CURSE.CFG。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$ROOT/workplace/dos-oracle"
IMAGE="${DOS_ORACLE_IMAGE:-dosbox-run:latest}"
NAME="coab-dos-oracle"
HOLD="${HOLD:-0.25}"

ex() { docker exec -e DISPLAY=:82 "$NAME" "$@"; }
focus() { ex bash -c 'w=$(xdotool search --onlyvisible --name DOSBox | tail -n 1); [ -n "$w" ] && xdotool windowfocus "$w"' >/dev/null 2>&1 || true; }

case "${1:-}" in
  start)
    mkdir -p "$WORK/game" "$WORK/out"
    if [ ! -f "$WORK/game/START.EXE" ]; then
      python3 - "$ROOT/curseoftheazurebonds.zip" "$WORK/game" <<'PY'
import sys, zipfile
zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])
PY
      mkdir -p "$WORK/game/SAVE"
    fi
    docker rm -f "$NAME" >/dev/null 2>&1 || true
    docker run -d --rm --name "$NAME" \
      --log-opt max-size=10m --log-opt max-file=3 \
      -u "$(id -u):$(id -g)" \
      -v "$WORK/game:/game" -v "$WORK/out:/out" \
      -v "$ROOT/tools/dos-oracle-session-inner.sh:/session.sh:ro" \
      -e HOME=/tmp -e COMMAND="${COMMAND:-}" \
      "$IMAGE" bash /session.sh >/dev/null
    sleep "${2:-8}"
    echo "started $NAME"
    ;;
  key)
    focus
    ex xdotool keydown --clearmodifiers "$2" >/dev/null 2>&1 || true
    sleep "$HOLD"
    ex xdotool keyup --clearmodifiers "$2" >/dev/null 2>&1 || true
    sleep "${3:-0.4}"
    ;;
  type)
    focus
    text="$2"
    for (( i=0; i<${#text}; i++ )); do
      ch="${text:i:1}"; [ "$ch" = " " ] && ch="space"
      ex xdotool keydown --clearmodifiers "$ch" >/dev/null 2>&1 || true
      sleep 0.12
      ex xdotool keyup --clearmodifiers "$ch" >/dev/null 2>&1 || true
      sleep 0.12
    done
    ;;
  shot)
    ex bash -c 'import -window root /tmp/s.xwd && convert /tmp/s.xwd -crop 640x400+0+0 +repage "/out/'"$2"'"' >/dev/null 2>&1
    echo "$WORK/out/$2"
    ;;
  text)
    name="${2:-_live.png}"
    ex bash -c 'import -window root /tmp/s.xwd && convert /tmp/s.xwd -crop 640x400+0+0 +repage "/out/'"$name"'"' >/dev/null 2>&1
    python3 "$ROOT/tools/dos_screen.py" "$WORK/out/$name"
    ;;
  stop)
    docker rm -f "$NAME" >/dev/null 2>&1 || true
    echo "stopped $NAME"
    ;;
  *)
    sed -n '2,30p' "$0"
    exit 2
    ;;
esac
