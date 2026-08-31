#!/usr/bin/env bash
# 以正式 Linux full-local AppImage 擷取繁中實機畫面，合成敘事型推廣片。
# 用法：tools/build-promo.sh <版本>
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
SOURCE_COMMIT="$(git --git-dir="$ROOT/workplace/azure-bonds-git" --work-tree="$ROOT" rev-parse --short=12 HEAD)"
GAME_SOURCE_COMMIT="${GAME_SOURCE_COMMIT:-$SOURCE_COMMIT}"
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

# 現代 A6 三種手繪框在同一 session 內依次解碼時，2 GiB 會在第二種樣式處觸發 OOM；
# 錄影必須保留全部熱切換素材，因此將媒體工作容器上限調為 4 GiB。
docker run --rm --network none --memory 4g --cpus 2 --pids-limit 384 \
  -u "$(id -u):$(id -g)" \
  --tmpfs /tmp/.X11-unix:rw,mode=1777 \
  -v "$RELEASE:/release:ro" -v "$OUT:/promo" \
  -v "$ROOT/workplace/dos-oracle/out:/dos-evidence:ro" \
  -v "$ROOT/tools/build-promo.sh:/recording/build-promo.sh:ro" \
  -e VERSION="$VERSION" -e GAME_SOURCE_COMMIT="$GAME_SOURCE_COMMIT" "$IMAGE" sh -c '
    set -eu
    mkdir -p /promo/clips /promo/frames
    app="/release/azure-bonds-remake-$VERSION-x86_64.AppImage"
    font="/release/linux/AppDir/assets/fonts/NotoSansTC-Regular.ttf"
    audio_title="/release/linux/AppDir/assets/audio/pc98-bgm-selector-01.ogg"
    audio_world="/release/linux/AppDir/assets/audio/pc98-bgm-selector-05.ogg"
    audio_combat="/release/linux/AppDir/assets/audio/pc98-bgm-selector-07.ogg"
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
      # 開場、戰鬥與結局 preview 的大型素材在視窗出現後仍會繼續解碼；
      # 以 10 秒有界穩定期避免把初始化黑幕當成遊戲畫面。
      sleep 10
      ffmpeg -hide_banner -loglevel error -y -f x11grab -framerate 30 \
        -video_size 640x480 -i :99.0+0,0 -t "$duration" -an \
        -vf "scale=880:660:flags=lanczos,pad=1280:720:200:0:0x07101c,drawtext=fontfile=$font:text=$caption:fontcolor=white:fontsize=31:x=(w-text_w)/2:y=674,tpad=stop_mode=clone:stop_duration=$duration,fps=30,setpts=N/(30*TB)" \
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

    # 從目前正式 AppImage 擷取建角固定狀態。DOS 側使用已登記的
    # runtime oracle PNG；對拍字幕明示是流程／版面對照，不宣稱同存檔逐像素。
    capture_still() {
      name=$1; shift
      setsid "$app" -sound-dir /tmp/coab-promo-no-audio "$@" >/tmp/coab-$name.log 2>&1 & game=$!
      n=0
      until xdotool search --onlyvisible --name "." >/dev/null 2>&1; do
        n=$((n+1)); test "$n" -lt 300 || { cat /tmp/coab-$name.log; exit 1; }
        sleep 0.1
      done
      sleep 10
      ffmpeg -hide_banner -loglevel error -y -f x11grab -video_size 640x480 \
        -i :99.0+0,0 -frames:v 1 "/promo/frames/source-$name.png"
      kill -TERM "-$game" 2>/dev/null || true
      wait "$game" 2>/dev/null || true
      game=""
      n=0
      while xdotool search --onlyvisible --name "." >/dev/null 2>&1; do
        n=$((n+1)); test "$n" -lt 50 || { echo "上一個視窗未關閉：$name" >&2; exit 1; }
        sleep 0.1
      done
    }

    compare_stills() {
      name=$1; dos=$2; remake=$3; caption=$4
      ffmpeg -hide_banner -loglevel error -y -loop 1 -i "$dos" -loop 1 -i "$remake" -t 4 \
        -filter_complex "[0:v]scale=560:420:force_original_aspect_ratio=decrease,pad=560:420:(ow-iw)/2:(oh-ih)/2:black[left];[1:v]scale=560:420:force_original_aspect_ratio=decrease,pad=560:420:(ow-iw)/2:(oh-ih)/2:black[right];[left][right]hstack=inputs=2[comparison];[comparison]pad=1280:720:80:90:black,drawtext=fontfile=$font:text=DOS 原版:fontcolor=0xffdc5c:fontsize=32:x=280-text_w/2:y=35,drawtext=fontfile=$font:text=Remake:fontcolor=0xffdc5c:fontsize=32:x=920-text_w/2:y=35,drawbox=x=0:y=570:w=1280:h=94:color=black@0.78:t=fill,drawtext=fontfile=$font:text=$caption:fontcolor=white:fontsize=31:x=(w-text_w)/2:y=595,fps=30,setpts=N/(30*TB)" \
        -an -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p "/promo/clips/$name.mp4"
    }

    # 在同一個正常遊戲 session 內錄下移動、F2 theme 切換、
    # F7 手繪樣式切換與 F3 原版 AREA 攻略地圖。
    # 不能以多張靜態 preview 拼接，否則無法證明 runtime hotkey 真的接到玩家路徑。
    record_switches() {
      name=03-live-switches
      setsid "$app" -sound-dir /tmp/coab-promo-no-audio -tilverton-dungeon >/tmp/coab-$name.log 2>&1 & game=$!
      n=0
      win=""
      until win=$(xdotool search --onlyvisible --name "." 2>/dev/null | tail -n 1) && test -n "$win"; do
        n=$((n+1)); test "$n" -lt 300 || { cat /tmp/coab-$name.log; exit 1; }
        sleep 0.1
      done
      # 互動分幕同樣要等現代素材完成首次解碼，否則會錄到約 2 秒黑幕。
      sleep 10
      xdotool windowfocus --sync "$win"
      focus_game() {
        current=""
        tries=0
        while test -z "$current" && test "$tries" -lt 300; do
          current=$(xdotool search --onlyvisible --pid "$game" 2>/dev/null | tail -n 1 || true)
          if test -z "$current"; then
            current=$(xdotool search --onlyvisible --name "." 2>/dev/null | tail -n 1 || true)
          fi
          test -n "$current" || sleep 0.1
          tries=$((tries+1))
        done
        test -n "$current" || { echo "找不到遊戲視窗" >&2; exit 1; }
        xdotool windowfocus --sync "$current"
      }
      tap() {
        focus_game
        xdotool key --clearmodifiers "$1"
      }
      : >/tmp/live-concat.txt
      capture_piece() {
        piece=$1; seconds=$2
        ffmpeg -hide_banner -loglevel error -y -f x11grab -framerate 30 \
          -video_size 640x480 -i :99.0+0,0 -t "$seconds" -an \
          -vf "scale=880:660:flags=lanczos,pad=1280:720:200:0:0x07101c,drawtext=fontfile=$font:text=同一場遊戲即時操作・F2 主題・F7 手繪樣式・F3 攻略地圖:fontcolor=white:fontsize=29:x=(w-text_w)/2:y=676,fps=30,setpts=N/(30*TB)" \
          -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p "/promo/clips/$name-$piece.mp4"
        printf "file %s\n" "/promo/clips/$name-$piece.mp4" >>/tmp/live-concat.txt
      }

      # 每次套用外觀或語言都可能讓 Ebiten 短暫重建視窗。載入空窗不收片，
      # 但所有操作仍發生在同一個正常遊戲程序與同一個 session。
      capture_piece 00-movement 3 & capture=$!
      sleep 0.4
      tap Up; sleep 0.8
      tap Right; sleep 0.8
      wait "$capture"

      tap F2
      focus_game
      capture_piece 10-original-theme 3
      tap F2
      focus_game
      capture_piece 11-modern-theme 3
      tap F7; sleep 0.4
      tap Right; sleep 0.4
      capture_piece 20-style-selection 1
      tap Return
      focus_game
      capture_piece 21-style-applied 2
      tap F3
      focus_game
      capture_piece 30-guide-map 3

      ffmpeg -hide_banner -loglevel error -y -f concat -safe 0 -i /tmp/live-concat.txt \
        -c copy "/promo/clips/$name.mp4"
      cp /tmp/coab-$name.log /promo/live-switches-runtime.log
      kill -TERM "-$game" 2>/dev/null || true
      wait "$game" 2>/dev/null || true
      game=""
      n=0
      while xdotool search --onlyvisible --name "." >/dev/null 2>&1; do
        n=$((n+1)); test "$n" -lt 50 || { echo "上一個視窗未關閉：$name" >&2; exit 1; }
        sleep 0.1
      done
    }

    record 01-opening "青色魔法紋章，喚醒被奪走的記憶" 5 -opening
    record 02-inn "場景與肖像對話・旅店裡的第一條線索" 5 -inn
    record 03-filani "完整繁中劇情・賢者費拉妮的警告" 5 -filani
    capture_still remake-party -guided-creation -guided-creation-step party -guided-creation-name 測試者
    capture_still remake-race -guided-creation
    capture_still remake-abilities -guided-creation -guided-creation-step abilities
    compare_stills 04-compare-party /dos-evidence/audit-function-20260830.png \
      /promo/frames/source-remake-party.png "DOS 原版資料對照・隊伍流程與 Gold Box 版面傳承"
    record 05-world-map "跨越月海諸城・原版世界地圖與現代重繪主題" 5 -world-map
    record_switches
    record 07-guide "F3 原版 AREA 地圖・事件點與繁中攻略" 5 -tilverton-dungeon -ui-preview guide-full
    record 08-combat "集結隊伍・以 AD&D 戰術迎戰青色枷鎖" 7 -encounter -combat-visual-demo magic
    record 09-ending "從提爾佛頓，征戰至最終決戰" 5 -ending-page 4

    ffmpeg -hide_banner -loglevel error -y -f lavfi -i color=c=0x07101c:s=1280x720:r=30:d=3 \
      -vf "drawtext=fontfile=$font:text=青色枷的詛咒:fontcolor=0xffdc5c:fontsize=62:x=(w-text_w)/2:y=240,drawtext=fontfile=$font:text=繁體中文完整版 Remake:fontcolor=white:fontsize=38:x=(w-text_w)/2:y=340" \
      -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p /promo/clips/00-title.mp4
    ffmpeg -hide_banner -loglevel error -y -f lavfi -i color=c=0x07101c:s=1280x720:r=30:d=4 \
      -vf "drawtext=fontfile=$font:text=枷鎖已烙印・冒險由你改寫:fontcolor=0xffdc5c:fontsize=48:x=(w-text_w)/2:y=250,drawtext=fontfile=$font:text=繁體中文・三平台・非商業授權:fontcolor=white:fontsize=30:x=(w-text_w)/2:y=345" \
      -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p /promo/clips/06-end.mp4

    printf "file %s\n" \
      /promo/clips/00-title.mp4 /promo/clips/01-opening.mp4 \
      /promo/clips/02-inn.mp4 /promo/clips/03-filani.mp4 \
      /promo/clips/04-compare-party.mp4 /promo/clips/05-world-map.mp4 \
      /promo/clips/03-live-switches.mp4 /promo/clips/07-guide.mp4 \
      /promo/clips/08-combat.mp4 /promo/clips/09-ending.mp4 \
      /promo/clips/06-end.mp4 >/tmp/concat.txt
    ffmpeg -hide_banner -loglevel error -y -f concat -safe 0 -i /tmp/concat.txt \
      -vf "fps=30,setpts=N/(30*TB)" -an -c:v libx264 -preset medium -crf 18 -pix_fmt yuv420p \
      /tmp/video.mp4
    duration=$(ffprobe -v error -show_entries format=duration -of default=nw=1:nk=1 /tmp/video.mp4)
    ffmpeg -hide_banner -loglevel error -y -i /tmp/video.mp4 \
      -stream_loop -1 -i "$audio_title" -stream_loop -1 -i "$audio_world" -stream_loop -1 -i "$audio_combat" \
      -filter_complex "[1:a]atrim=0:18,afade=t=in:st=0:d=1.2,afade=t=out:st=16:d=2,volume=1.35[a1];[2:a]atrim=0:22,afade=t=in:st=0:d=1,afade=t=out:st=20:d=2,adelay=18000|18000,volume=1.45[a2];[3:a]atrim=0:24,afade=t=in:st=0:d=0.8,afade=t=out:st=21:d=3,adelay=40000|40000,volume=1.55[a3];[a1][a2][a3]amix=inputs=3:duration=longest:normalize=0,volume=9dB,alimiter=limit=0.841[a]" \
      -map 0:v -map "[a]" -t "$duration" -c:v copy -c:a aac -b:a 192k -movflags +faststart \
      -metadata title="Curse of the Azure Bonds Traditional Chinese Remake - 繁中推廣片" \
      -metadata comment="Rights pending: internal review only; not cleared for public distribution" \
      /promo/coab-remake-promo-zh-TW.mp4

    ffprobe -v error -show_format -show_streams -of json /promo/coab-remake-promo-zh-TW.mp4 >/promo/ffprobe.json
    ffmpeg -hide_banner -y -i /promo/coab-remake-promo-zh-TW.mp4 \
      -vf "fps=1/2,scale=256:144:flags=lanczos,tile=5x7" -frames:v 1 /promo/contact-sheet.png \
      >/tmp/contact.log 2>&1
    for second in 2 6 11 16 21 26 31 36 41 46 51 56 61; do
      ffmpeg -hide_banner -loglevel error -y -ss "$second" -i /promo/coab-remake-promo-zh-TW.mp4 \
        -frames:v 1 "/promo/frames/frame-$second.png"
    done
    ffmpeg -hide_banner -i /promo/coab-remake-promo-zh-TW.mp4 -af volumedetect -f null - \
      >/promo/volume-check.txt 2>&1
    ffmpeg -hide_banner -i /promo/coab-remake-promo-zh-TW.mp4 -filter_complex \
      ebur128=peak=true -f null - >/promo/loudness-check.txt 2>&1
    ffmpeg -hide_banner -i /promo/coab-remake-promo-zh-TW.mp4 -af silencedetect=n=-50dB:d=3 -f null - \
      >/promo/silence-check.txt 2>&1
    ffmpeg -hide_banner -i /promo/coab-remake-promo-zh-TW.mp4 -vf blackdetect=d=1:pix_th=0.02 -an -f null - \
      >/promo/black-check.txt 2>&1
    script_sha=$(sha256sum /recording/build-promo.sh | cut -d" " -f1)
    printf "version=%s\ngame_source_commit=%s\nrecording_script_sha256=%s\nvideo_source=packaged Linux full-local AppImage, Docker/Xvfb x11grab\nlanguage=zh-TW primary\nstory_evidence=opening, Windlords Inn portrait dialogue, Filani portrait dialogue, world map, F3 guide map, combat, ending\ndos_comparison_source=registered DOS runtime oracle PNG; flow/layout comparison, not same-save pixel parity\ninteractive_evidence=one normal Tilverton session: movement, F2 original theme, F2 modern theme, F7 hand-painted style, F3 guide map\naudio_source=PC-98 derived OGG selectors 01 title, 05 wilderness, 07 combat\naudio_label=original-game-derived remake playback; not new original soundtrack\nrights_status=internal review only; public distribution not cleared\n" "$VERSION" "$GAME_SOURCE_COMMIT" "$script_sha" >/promo/metadata.txt
    cd /promo
    sha256sum coab-remake-promo-zh-TW.mp4 contact-sheet.png ffprobe.json metadata.txt \
      live-switches-runtime.log volume-check.txt loudness-check.txt silence-check.txt black-check.txt >SHA256SUMS.txt
  '

echo "完成：$OUT/coab-remake-promo-zh-TW.mp4"
