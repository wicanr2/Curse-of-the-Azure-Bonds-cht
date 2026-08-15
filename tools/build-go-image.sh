#!/usr/bin/env bash
# 建立本專案專用的 Go 映像（含 Ebiten／oto 的 X11 與 ALSA 開發標頭）。
# 只建立本專案自己的 tag，不動任何既有映像。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${COAB_GO_IMAGE:-coab-go-ebiten:1.24}"
docker build -t "$IMAGE" -f "$ROOT/tools/Dockerfile.go-ebiten" "$ROOT/tools"
echo "已建立 $IMAGE；tools/go.sh 會自動採用。"
