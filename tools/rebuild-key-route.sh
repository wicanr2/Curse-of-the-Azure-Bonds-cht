#!/usr/bin/env bash
# 從目前完整 internal/game 測試原子重生正常鍵盤 session 的唯一現行 route oracle。
# 產物在 workplace/，不進 Git；正式文件與驗證命令一律只引用 route-current.json。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REL_TARGET="workplace/campaign-frames/route-current.json"
REL_CANDIDATE="${REL_TARGET}.candidate"
MIN_STEPS=10000
IMAGE="${GO_IMAGE:-coab-go-ebiten:1.24}"

mkdir -p "$ROOT/workplace/campaign-frames"

# 候選檔與正式檔分開；測試失敗或步數異常時不碰上一份可用 oracle。
docker run --rm --network none --memory 256m --cpus 1 --pids-limit 64 \
  -u "$(id -u):$(id -g)" -v "$ROOT:/src" -w /src "$IMAGE" \
  sh -lc "rm -f '$REL_CANDIDATE'"

docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
  -u "$(id -u):$(id -g)" -v "$ROOT:/workspace" -w /workspace \
  -e GOMODCACHE=/workspace/workplace/go-mod-cache \
  -e COAB_DECISION_LOG="/workspace/$REL_CANDIDATE" "$IMAGE" sh -lc '
    cp go.mod /tmp/coab-route.mod
    cp go.sum /tmp/coab-route.sum
    /usr/local/go/bin/go mod edit -modfile=/tmp/coab-route.mod \
      -replace=github.com/wicanr2/golden-box-remake-engine=/workspace/golden-box-remake-engine
    /usr/local/go/bin/go test -modfile=/tmp/coab-route.mod ./internal/game -count=1
  '

docker run --rm --network none --memory 256m --cpus 1 --pids-limit 64 \
  -u "$(id -u):$(id -g)" -v "$ROOT:/src" -w /src "$IMAGE" \
  sh -lc "steps=\$(grep -c '\"kind\"' '$REL_CANDIDATE'); \
    test \"\$steps\" -ge '$MIN_STEPS'; \
    mv '$REL_CANDIDATE' '$REL_TARGET'; \
    sha256sum '$REL_TARGET' > '${REL_TARGET}.sha256'; \
    printf 'route oracle: %s steps, ' \"\$steps\"; cat '${REL_TARGET}.sha256'"

# 最小一般強度玩家路徑 gate。可用 COAB_KEY_FRAMES 明確延長，但禁止開 boost。
docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
  -u "$(id -u):$(id -g)" -v "$ROOT:/workspace" -w /workspace \
  -e GOMODCACHE=/workspace/workplace/go-mod-cache \
  -e COAB_KEY_BOOST=0 -e COAB_KEY_FRAMES="${COAB_KEY_FRAMES:-1800}" \
  -e COAB_ROUTE_JSON="$REL_TARGET" "$IMAGE" sh -lc '
    cp go.mod /tmp/coab-route.mod
    cp go.sum /tmp/coab-route.sum
    /usr/local/go/bin/go mod edit -modfile=/tmp/coab-route.mod \
      -replace=github.com/wicanr2/golden-box-remake-engine=/workspace/golden-box-remake-engine
    trap '\''kill "$xvfb_pid" 2>/dev/null || true'\'' EXIT INT TERM
    Xvfb :99 -screen 0 1280x800x24 >/tmp/xvfb.log 2>&1 & xvfb_pid=$!
    DISPLAY=:99 /usr/local/go/bin/go test -modfile=/tmp/coab-route.mod \
      ./cmd/azure-bonds-game -run '\''^TestKeysDriveARealSessionFromTheTitle$'\'' -count=1 -v
  '
