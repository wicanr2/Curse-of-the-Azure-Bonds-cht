#!/usr/bin/env bash
# 在隔離 Xvfb 容器內啟動 Linux full-local AppImage 並截圖。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
RELEASE="$ROOT/dist-all/$VERSION/full-local"
APPIMAGE="$RELEASE/azure-bonds-remake-$VERSION-x86_64.AppImage"
OUT="$ROOT/dist-all/$VERSION/smoke"

if [[ -z "$VERSION" || ! -x "$APPIMAGE" ]]; then
  echo "用法：tools/linux-release-smoke.sh <已建置版本>" >&2
  exit 2
fi

mkdir -p "$OUT"
rm -f "$OUT/coab-linux-release-smoke.png"
docker run --rm --network none --memory 2g --cpus 2 --pids-limit 384 \
  -u "$(id -u):$(id -g)" \
  --tmpfs /tmp/.X11-unix:rw,mode=1777 \
  -v "$RELEASE:/release:ro" -v "$OUT:/smoke" coab-go-ebiten:1.24 sh -c '
    set -eu
    Xvfb :99 -screen 0 800x600x24 >/tmp/xvfb.log 2>&1 &
    xvfb=$!
    trap "kill $xvfb 2>/dev/null || true" EXIT
    n=0
    until test -S /tmp/.X11-unix/X99; do
      n=$((n+1)); test "$n" -lt 50 || { cat /tmp/xvfb.log; exit 1; }
      sleep 0.1
    done
    DISPLAY=:99 APPIMAGE_EXTRACT_AND_RUN=1 \
      /release/'"$(basename "$APPIMAGE")"' \
      -opening -screenshot /smoke/coab-linux-release-smoke.png
  '

test -s "$OUT/coab-linux-release-smoke.png"
echo "完成：$OUT/coab-linux-release-smoke.png"
