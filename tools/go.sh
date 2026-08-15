#!/usr/bin/env bash
# Docker 內的 Go 工具鏈（本專案一律不用主機 Go）。
# 用法：tools/go.sh <go 子命令與參數...>
#   tools/go.sh build ./cmd/ovr-manifest
#   tools/go.sh test ./internal/pc98ovr
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# 預設用本專案的映像（含 Ebiten／oto 需要的 X11 與 ALSA 開發標頭），
# 沒建過就退回 golang:1.24 並提示——那顆缺標頭，internal/sound 與
# cmd/azure-bonds-game 會建置失敗，`go test ./...` 就不能當全套 gate。
# 建立方式：tools/build-go-image.sh
DEFAULT_IMAGE="coab-go-ebiten:1.24"
if [ -n "${GO_IMAGE:-}" ]; then
  IMAGE="$GO_IMAGE"
elif docker image inspect "$DEFAULT_IMAGE" >/dev/null 2>&1; then
  IMAGE="$DEFAULT_IMAGE"
else
  IMAGE="golang:1.24"
  echo "提示：$DEFAULT_IMAGE 尚未建立，改用 golang:1.24（缺 X11／ALSA 標頭）。" >&2
  echo "      要讓 go test ./... 能當全套 gate，請先跑 tools/build-go-image.sh。" >&2
fi
# Ebiten 在 package init 階段就開 GLFW，沒有 DISPLAY 會直接 panic，容器又不會
# 繼承主機的 DISPLAY。本專案映像裡有 with-xvfb（起 Xvfb :99 再執行命令），
# 用它包起來對不需要顯示的命令無害。
XVFB=""
if [ "$IMAGE" = "$DEFAULT_IMAGE" ]; then
  XVFB="with-xvfb"
fi
mkdir -p "$ROOT/workplace/go-build-cache" "$ROOT/workplace/go-mod-cache"
exec docker run --rm \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ROOT:/src" \
  -v "$ROOT/workplace/go-build-cache:/gocache" \
  -v "$ROOT/workplace/go-mod-cache:/gomod" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomod \
  -e GOFLAGS="-mod=mod -buildvcs=false" \
  `# buildvcs=false：本 repo 的 .git 在 workplace/azure-bonds-git，根目錄的 .git 是空殼` \
  -w /src "$IMAGE" $XVFB go "$@"
