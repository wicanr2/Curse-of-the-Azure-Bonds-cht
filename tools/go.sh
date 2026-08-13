#!/usr/bin/env bash
# Docker 內的 Go 工具鏈（本專案一律不用主機 Go）。
# 用法：tools/go.sh <go 子命令與參數...>
#   tools/go.sh build ./cmd/ovr-manifest
#   tools/go.sh test ./internal/pc98ovr
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${GO_IMAGE:-golang:1.24}"
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
  -w /src "$IMAGE" go "$@"
