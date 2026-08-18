#!/usr/bin/env bash
# 由 tools/dos-oracle-session.sh 掛進容器當 PID 1：起 Xvfb ＋ DOSBox 之後就掛著，
# 讓主機端用 docker exec 逐步送鍵與擷取（互動式 oracle）。
set -u
export DISPLAY=:82
Xvfb :82 -screen 0 1280x960x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
for i in $(seq 1 50); do xset q >/dev/null 2>&1 && break; sleep 0.2; done
dosbox -noconsole -machine vgaonly \
  -c "mount c /game" -c "c:" -c "${COMMAND:-START.EXE STING Wooden}" >/tmp/dosbox.log 2>&1 &
echo $! > /tmp/dosbox.pid
for i in $(seq 1 60); do
  win=$(xdotool search --onlyvisible --name DOSBox 2>/dev/null | head -n 1)
  [ -n "$win" ] && break
  sleep 0.1
done
[ -n "${win:-}" ] && xdotool windowfocus "$win" 2>/dev/null
exec tail -f /dev/null
