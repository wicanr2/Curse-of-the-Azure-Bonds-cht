#!/usr/bin/env bash
# 在隔離 Wine／Xvfb 容器內啟動 Windows full-local 發行版並截圖。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
IMAGE="coab-wine-smoke:ubuntu-noble-20260826"
RELEASE="$ROOT/dist/$VERSION/full-local/windows"
OUT="$ROOT/dist/$VERSION/smoke"

if [[ -z "$VERSION" || ! -x "$RELEASE/azure-bonds-game.exe" ]]; then
  echo "用法：tools/windows-release-smoke.sh <已建置版本>" >&2
  exit 2
fi

docker image inspect "$IMAGE" >/dev/null
mkdir -p "$OUT"
rm -f "$OUT/coab-windows-release-smoke.png"
docker run --rm --network none --memory 1g --cpus 2 --pids-limit 256 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp/wine-home -e WINEPREFIX=/tmp/wine-prefix -e WINEDEBUG=-all \
  -v "$RELEASE:/release:ro" -v "$OUT:/smoke" -w /release "$IMAGE" sh -c '
    set -eu
    mkdir -p "$HOME" "$WINEPREFIX"
    Xvfb :99 -screen 0 800x600x24 >/tmp/xvfb.log 2>&1 &
    xvfb=$!
    trap "kill $xvfb 2>/dev/null || true" EXIT
    DISPLAY=:99 /usr/lib/wine/wine64 azure-bonds-game.exe \
      -opening -screenshot Z:\\smoke\\coab-windows-release-smoke.png
  '

test -s "$OUT/coab-windows-release-smoke.png"
echo "完成：$OUT/coab-windows-release-smoke.png"
