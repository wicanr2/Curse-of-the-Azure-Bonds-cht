#!/usr/bin/env bash
# 把「開場到結局」戰役的每一個劇情檢查點，用**真的前端**開起來畫一張。
#
# ★ 存在的理由：`TestRealNewGameRunsToTheEnding` 驅動的是 `*State`，
# 畫面那一層一次都沒被跑到。remake-status 的「開場到結局的同一 session」那一列
# 因此寫著「那是**測試路徑**；真人從開場玩到結局沒有自動化的證明」。
# 這支腳本補的就是那一句：同一條路線的每個檢查點，都由玩家真正會執行的
# `cmd/azure-bonds-game` 載入並畫出一張 640×480。
#
# 用法：
#   tools/campaign-frames.sh            # 先跑戰役導出快照，再逐張畫
#   tools/campaign-frames.sh -skip-run  # 沿用既有快照，只重畫
#
# ⚠ 字型在 repo 外（`/home/anr2/cht/etan_font`），所以要另外唯讀掛進容器。
# 少了它每一個字都會變成豆腐框，而**畫面其他部分完全正常**——只看「有沒有
# 產生 PNG」是驗不出來的。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FONT_DIR="${COAB_ETEN_FONT_DIR:-/home/anr2/cht/etan_font}"
SAVES="$ROOT/workplace/campaign-frames/saves"
PNGS="$ROOT/workplace/campaign-frames/png"
IMAGE="${GO_IMAGE:-coab-go-ebiten:1.24}"

if [ ! -f "$FONT_DIR/stdfont.15" ]; then
  echo "找不到倚天字型 $FONT_DIR/stdfont.15；設 COAB_ETEN_FONT_DIR 指過去。" >&2
  exit 1
fi

if [ "${1:-}" != "-skip-run" ]; then
  echo "== 跑戰役、導出劇情檢查點 =="
  rm -rf "$SAVES"
  mkdir -p "$SAVES"
  COAB_CAMPAIGN_SNAPSHOT_DIR=/src/workplace/campaign-frames/saves \
    "$ROOT/tools/go.sh" test ./internal/game/ -run TestRealNewGameRunsToTheEnding -count=1
fi

mkdir -p "$PNGS"
rm -f "$PNGS"/*.png
count=0
for save in "$SAVES"/*.json; do
  [ -e "$save" ] || continue
  name="$(basename "$save" .json)"
  echo "== 畫 $name =="
  # ⚠ 一張畫不出來不要整批中斷：**要看的是「哪幾張畫不出來」**，
  # 中斷會讓後面全部變成「沒資料」而不是「失敗」。
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$ROOT:/src" \
    -v "$FONT_DIR:/fonts:ro" \
    -v "$ROOT/workplace/go-build-cache:/gocache" \
    -v "$ROOT/workplace/go-mod-cache:/gomod" \
    -e GOCACHE=/gocache -e GOMODCACHE=/gomod \
    -e GOFLAGS="-mod=mod -buildvcs=false" \
    -e GOPROXY="file:///src/workplace/engine-proxy,https://proxy.golang.org,direct" \
    -e GOSUMDB=off \
    -w /src "$IMAGE" with-xvfb go run ./cmd/azure-bonds-game \
      -eten-font /fonts/stdfont.15 \
      -party-load "/src/workplace/campaign-frames/saves/$name.json" \
      -screenshot "/src/workplace/campaign-frames/png/$name.png" \
      >/dev/null 2>&1 || echo "   ⚠ $name 畫不出來" >&2
  count=$((count + 1))
done
echo "== 完成：$count 個檢查點，$(ls -1 "$PNGS"/*.png 2>/dev/null | wc -l) 張畫出來 =="
