#!/usr/bin/env bash
# 以新建置的 Linux full-local AppImage 擷取實機畫面並合成內部展示版推廣片。
# 用法：tools/build-promo.sh <版本>
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
SOURCE_COMMIT="$(git --git-dir="$ROOT/workplace/azure-bonds-git" --work-tree="$ROOT" rev-parse --short=12 HEAD)"
IMAGE="game-video:latest"
RELEASE="$ROOT/dist-all/$VERSION/full-local"
APPIMAGE="$RELEASE/azure-bonds-remake-$VERSION-x86_64.AppImage"
OUT="$ROOT/dist-all/$VERSION/promo"

if [[ -z "$VERSION" || ! -x "$APPIMAGE" ]]; then
  echo "用法：tools/build-promo.sh <已建置版本>" >&2
  exit 2
fi
docker image inspect "$IMAGE" >/dev/null
docker run --rm --network none --memory 64m --cpus 1 --pids-limit 32 \
  -u "$(id -u):$(id -g)" -v "$ROOT/dist-all/$VERSION:/dist-version" \
  busybox:1.37 sh -c 'rm -rf /dist-version/promo && mkdir -p /dist-version/promo'

docker run --rm --network none --memory 2g --cpus 2 --pids-limit 384 \
  -u "$(id -u):$(id -g)" \
  --tmpfs /tmp/.X11-unix:rw,mode=1777 \
  -v "$RELEASE:/release:ro" -v "$OUT:/promo" \
  -e VERSION="$VERSION" -e SOURCE_COMMIT="$SOURCE_COMMIT" "$IMAGE" sh -c '
    set -eu
    mkdir -p /promo/clips /promo/frames
    app="/release/azure-bonds-remake-$VERSION-x86_64.AppImage"
    font="/release/linux/AppDir/assets/fonts/NotoSansTC-Regular.ttf"
    audio="/release/linux/AppDir/assets/audio/pc98-bgm-selector-01.ogg"
    export HOME=/tmp/coab-promo-home APPIMAGE_EXTRACT_AND_RUN=1 DISPLAY=:99
    mkdir -p "$HOME"
    Xvfb :99 -screen 0 640x480x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    xvfb=$!
    game=""
    trap "test -z \"$game\" || kill -TERM -$game 2>/dev/null || true; kill $xvfb 2>/dev/null || true" EXIT
    n=0
    until test -S /tmp/.X11-unix/X99; do
      n=$((n+1)); test "$n" -lt 50 || { cat /tmp/xvfb.log; exit 1; }
      sleep 0.1
    done

    record() {
      name=$1; caption=$2; duration=$3; shift 3
      setsid "$app" -sound-dir /tmp/coab-promo-no-audio "$@" >/tmp/coab-$name.log 2>&1 & game=$!
      n=0
      until xdotool search --onlyvisible --name "." >/dev/null 2>&1; do
        n=$((n+1)); test "$n" -lt 300 || { cat /tmp/coab-$name.log; exit 1; }
        sleep 0.1
      done
      # X11 關閉上一個 Ebiten 視窗後可能保留最後一幀；等新視窗穩定再收片，
      # 避免把上一段的正常畫面誤配到下一段字幕。
      sleep 1
      ffmpeg -hide_banner -loglevel error -y -f x11grab -framerate 30 \
        -video_size 640x480 -i :99.0+0,0 -t "$duration" -an \
        -vf "scale=960:720:flags=neighbor,pad=1280:720:160:0:black,drawbox=x=0:y=646:w=1280:h=74:color=black@0.74:t=fill,drawtext=fontfile=$font:text=$caption:fontcolor=white:fontsize=34:x=(w-text_w)/2:y=665,tpad=stop_mode=clone:stop_duration=$duration,fps=30,setpts=N/(30*TB)" \
        -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p "/promo/clips/$name.mp4"
      kill -TERM "-$game" 2>/dev/null || true
      wait "$game" 2>/dev/null || true
      game=""
      n=0
      while xdotool search --onlyvisible --name "." >/dev/null 2>&1; do
        n=$((n+1)); test "$n" -lt 50 || { echo "上一個視窗未關閉：$name" >&2; exit 1; }
        sleep 0.1
      done
    }

    # 在同一個正常遊戲 session 內錄下移動、F2 theme 雙向切換與 F6 語言循環。
    # 不能以多張靜態 preview 拼接，否則無法證明 runtime hotkey 真的接到玩家路徑。
    record_switches() {
      name=03-live-switches
      duration=12
      setsid "$app" -sound-dir /tmp/coab-promo-no-audio -tilverton-dungeon >/tmp/coab-$name.log 2>&1 & game=$!
      n=0
      win=""
      until win=$(xdotool search --onlyvisible --name "." 2>/dev/null | tail -n 1) && test -n "$win"; do
        n=$((n+1)); test "$n" -lt 300 || { cat /tmp/coab-$name.log; exit 1; }
        sleep 0.1
      done
      sleep 1
      xdotool windowfocus --sync "$win"
      ffmpeg -hide_banner -loglevel error -y -f x11grab -framerate 30 \
        -video_size 640x480 -i :99.0+0,0 -t "$duration" -an \
        -vf "scale=960:720:flags=neighbor,pad=1280:720:160:0:black,drawbox=x=0:y=646:w=1280:h=74:color=black@0.74:t=fill,drawtext=fontfile=$font:text=實際遊玩・F2 雙 theme・F6 四語言:fontcolor=white:fontsize=34:x=(w-text_w)/2:y=665,tpad=stop_mode=clone:stop_duration=$duration,fps=30,setpts=N/(30*TB)" \
        -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p "/promo/clips/$name.mp4" & capture=$!
      sleep 1
      xdotool key --clearmodifiers Up; sleep 0.8
      xdotool key --clearmodifiers Right; sleep 0.8
      xdotool key --clearmodifiers Up; sleep 0.8
      xdotool key --clearmodifiers F2; sleep 1.3
      xdotool key --clearmodifiers F2; sleep 1.3
      xdotool key --clearmodifiers F6; sleep 1
      xdotool key --clearmodifiers F6; sleep 1
      xdotool key --clearmodifiers F6
      wait "$capture"
      kill -TERM "-$game" 2>/dev/null || true
      wait "$game" 2>/dev/null || true
      game=""
    }

    record 01-opening "繁體中文・完整主線" 6 -opening
    record 02-party "原版流程建角・新增操作指引" 6 -guided-creation -guided-creation-step party
    record_switches
    record 03-guide "F3 原版 AREA 地圖・事件攻略" 6 -tilverton-dungeon -ui-preview guide-full
    record 04-combat "回合制戰鬥・原作 sprite 與地形" 7 -encounter -combat-visual-demo magic
    record 05-ending "從提爾佛頓到最終決戰" 6 -ending-page 4

    ffmpeg -hide_banner -loglevel error -y -f lavfi -i color=c=0x07101c:s=1280x720:r=30:d=3 \
      -vf "drawtext=fontfile=$font:text=青色枷的詛咒:fontcolor=0xffdc5c:fontsize=58:x=(w-text_w)/2:y=250,drawtext=fontfile=$font:text=繁體中文 Remake:fontcolor=white:fontsize=38:x=(w-text_w)/2:y=340" \
      -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p /promo/clips/00-title.mp4
    ffmpeg -hide_banner -loglevel error -y -f lavfi -i color=c=0x07101c:s=1280x720:r=30:d=4 \
      -vf "drawtext=fontfile=$font:text=非商業版・開發中:fontcolor=0xffdc5c:fontsize=48:x=(w-text_w)/2:y=270,drawtext=fontfile=$font:text=PolyForm Noncommercial 1.0.0:fontcolor=white:fontsize=28:x=(w-text_w)/2:y=350" \
      -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p /promo/clips/06-end.mp4

    printf "file %s\n" \
      /promo/clips/00-title.mp4 /promo/clips/01-opening.mp4 /promo/clips/02-party.mp4 \
      /promo/clips/03-live-switches.mp4 /promo/clips/03-guide.mp4 \
      /promo/clips/04-combat.mp4 /promo/clips/05-ending.mp4 \
      /promo/clips/06-end.mp4 >/tmp/concat.txt
    ffmpeg -hide_banner -loglevel error -y -f concat -safe 0 -i /tmp/concat.txt \
      -vf "fps=30,setpts=N/(30*TB)" -an -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p \
      /tmp/video.mp4
    ffmpeg -hide_banner -loglevel error -y -i /tmp/video.mp4 -stream_loop -1 -i "$audio" \
      -filter_complex "[1:a]volume=1.5,afade=t=in:st=0:d=1.5,afade=t=out:st=48:d=2[a]" \
      -map 0:v -map "[a]" -t 50 -c:v copy -c:a aac -b:a 160k -movflags +faststart \
      -metadata title="Curse of the Azure Bonds Traditional Chinese Remake - internal promo" \
      -metadata comment="Rights pending: internal review only; not cleared for public distribution" \
      /promo/coab-remake-promo-internal.mp4

    ffprobe -v error -show_format -show_streams -of json /promo/coab-remake-promo-internal.mp4 >/promo/ffprobe.json
    ffmpeg -hide_banner -y -i /promo/coab-remake-promo-internal.mp4 \
      -vf "fps=1/5,scale=256:144:flags=lanczos,tile=5x2" -frames:v 1 /promo/contact-sheet.png \
      >/tmp/contact.log 2>&1
    for second in 2 8 14 20 24 28 33 40 46 49; do
      ffmpeg -hide_banner -loglevel error -y -ss "$second" -i /promo/coab-remake-promo-internal.mp4 \
        -frames:v 1 "/promo/frames/frame-$second.png"
    done
    ffmpeg -hide_banner -i /promo/coab-remake-promo-internal.mp4 -af volumedetect -f null - \
      >/promo/volume-check.txt 2>&1
    ffmpeg -hide_banner -i /promo/coab-remake-promo-internal.mp4 -af silencedetect=n=-50dB:d=3 -f null - \
      >/promo/silence-check.txt 2>&1
    printf "version=%s\nsource_commit=%s\nvideo_source=packaged Linux full-local AppImage, Docker/Xvfb x11grab\ninteractive_evidence=normal Tilverton session with movement then injected F2,F2,F6,F6,F6\naudio_source=assets/audio/pc98-bgm-selector-01.ogg\nrights_status=internal review only; public distribution not cleared\n" "$VERSION" "$SOURCE_COMMIT" >/promo/metadata.txt
    cd /promo
    sha256sum coab-remake-promo-internal.mp4 contact-sheet.png ffprobe.json metadata.txt >SHA256SUMS.txt
  '

echo "完成：$OUT/coab-remake-promo-internal.mp4"
