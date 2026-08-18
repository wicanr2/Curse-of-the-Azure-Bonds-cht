#!/usr/bin/env bash
# 由 tools/dos-oracle.sh 掛進容器執行；說明與已知的坑寫在那一支。
# 在容器裡跑原版 DOS 版並依序送鍵，最後擷取畫面。
# 用法（容器內）：capture.sh <輸出png> <boot等待秒> <"鍵:延遲 鍵:延遲 ...">
set -u
out="$1"; boot="$2"; seq="$3"
Xvfb :82 -screen 0 1280x960x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
export DISPLAY=:82
for i in $(seq 1 50); do xset q >/dev/null 2>&1 && break; sleep 0.2; done
dosbox -noconsole -machine vgaonly \
  -c "mount c /game" -c "c:" -c "START.EXE STING Wooden" >/tmp/dosbox.log 2>&1 &
dosbox_pid=$!
# 顯示介面卡的提示只活約兩秒，所以不能用固定 sleep 等視窗——要一出現就搶焦點。
win=""
for i in $(seq 1 60); do
  win=$(xdotool search --onlyvisible --name DOSBox 2>/dev/null | head -n 1)
  [ -n "$win" ] && break
  sleep 0.1
done
echo "window=$win" >&2
if [ -n "$win" ]; then
  # 沒有視窗管理員時 windowactivate 會失敗，要直接設輸入焦點；
  # 送鍵一律走 XTEST（不加 --window），SDL 會忽略 XSendEvent 的合成事件。
  xdotool windowfocus --sync "$win" 2>/dev/null || xdotool windowfocus "$win" 2>/dev/null || true
  xdotool windowraise "$win" 2>/dev/null || true
fi
sleep "$boot"
for step in $seq; do
  key="${step%%:*}"; delay="${step##*:}"
  if [ "$key" = "-" ]; then sleep "$delay"; continue; fi
  # 每次送鍵前重設焦點：DOSBox 的 SDL 視窗剛映射時還吃不到 XTEST，
  # 只在啟動時設一次焦點會讓最早的幾個按鍵整個掉光。
  # ⚠ 視窗要**每次重新找**：DOSBox 切換顯示模式時會重建 SDL 視窗，
  # 啟動時抓到的 id 之後就是死的，對它設焦點不會有任何錯誤訊息，
  # 但按鍵會整批掉光。
  cur=$(xdotool search --onlyvisible --name DOSBox 2>/dev/null | tail -n 1)
  [ -n "$cur" ] && xdotool windowfocus "$cur" 2>/dev/null
  xdotool key --clearmodifiers "$key" 2>/dev/null || true
  sleep "$delay"
done
# 連拍：$4 張、每張間隔 $5 秒。示範模式是自動播放的，逐輪只拍一張太貴。
frames="${4:-1}"; interval="${5:-3}"
i=0
while [ "$i" -lt "$frames" ]; do
  import -window root /tmp/shot.xwd >/dev/null 2>&1
  target="$out"
  if [ "$frames" -gt 1 ]; then target="${out%.png}-$(printf %02d "$i").png"; fi
  convert /tmp/shot.xwd -crop 640x400+0+0 +repage "$target" >/dev/null 2>&1 || convert /tmp/shot.xwd "$target" >/dev/null 2>&1
  i=$((i + 1))
  [ "$i" -lt "$frames" ] && sleep "$interval"
done
kill "$dosbox_pid" 2>/dev/null || true
