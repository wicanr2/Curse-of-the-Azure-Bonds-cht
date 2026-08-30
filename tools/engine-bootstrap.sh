#!/usr/bin/env bash
# 讓一份**乾淨 clone** 建得起來：把 go.mod 鎖住的那個 engine commit 準備好，
# 並重建檔案型 Go module proxy。
#
# 為什麼需要這一支：engine 是固定放在 CoAB 同層的獨立 repo，而 `workplace/`
# 是忽略的工作目錄，所以剛 clone 完的 CoAB **可能沒有 engine clone，也沒有
# proxy**。engine 是私有 repo，
# `proxy.golang.org` 取不到，容器裡又不該放 GitHub 憑證。
#
# 用法（在 CoAB 根目錄）：
#   tools/engine-bootstrap.sh
#
# 做完之後 `tools/go.sh test ./...` 就能跑。之後要升級 engine 相依，走的是
# AGENTS.md 那條流程（engine push → engine-proxy.sh → go get），不是這一支。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENGINE="${GOLDEN_BOX_REMAKE_ENGINE_DIR:-$ROOT/../golden-box-remake-engine}"
MODULE="github.com/wicanr2/golden-box-remake-engine"
REMOTE="https://github.com/wicanr2/golden-box-remake-engine.git"

version="$(awk -v m="$MODULE" '$1 == m { print $2 }' "$ROOT/go.mod" | head -1)"
[ -n "$version" ] || { echo "go.mod 裡找不到 $MODULE 的版本" >&2; exit 2; }
# pseudo-version 的尾巴就是 12 碼 short commit。
commit="${version##*-}"
echo "go.mod 鎖的是 $version（commit $commit）"

if [ ! -d "$ENGINE/.git" ]; then
  echo "找不到固定共用 engine repo，改用 clone：$REMOTE"
  git clone "$REMOTE" "$ENGINE"
fi

# 只要 repo 裡有那個 commit 就夠——**不動工作區、不 checkout**，
# 免得踩掉開發者自己在 engine 上的改動。
if ! git -C "$ENGINE" cat-file -e "$commit^{commit}" 2>/dev/null; then
  echo "本機 engine repo 沒有 $commit，先 fetch"
  git -C "$ENGINE" fetch origin
fi
if ! git -C "$ENGINE" cat-file -e "$commit^{commit}" 2>/dev/null; then
  cat >&2 <<MSG
engine repo 裡找不到 commit $commit。
可能的原因：那個 commit 還沒 push 上 GitHub，或這台機器取不到私有 repo。
先在有那份 commit 的機器上 \`git -C golden-box-remake-engine push origin main\`。
MSG
  exit 3
fi

built="$("$ROOT/tools/engine-proxy.sh" "$commit")"
if [ "$built" != "$version" ]; then
  echo "打包出來的版本是 $built，與 go.mod 鎖的 $version 不同" >&2
  exit 4
fi
echo "proxy 已就緒：$ROOT/workplace/engine-proxy"
echo "接下來可以跑：tools/go.sh test ./..."
