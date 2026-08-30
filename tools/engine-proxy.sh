#!/usr/bin/env bash
# 把主機固定共用 engine repo 打包成一個「檔案型 Go module proxy」，讓 CoAB 的
# go.mod 可以鎖到新的 engine commit。
#
# 為什麼需要這一支：`golden-box-remake-engine` 是**私有** repo，proxy.golang.org
# 取不到；而容器裡沒有 GitHub 憑證，也不該放（見 rules：不要把 git 憑證複製進
# 客戶主機／容器）。所以 `go get <module>@<commit>` 一定會失敗在
# "could not read Username for https://github.com"。
#
# 這一支不需要任何憑證：模組的 zip 內容就是本機那份已經 push 上去的 commit，
# 版本號用 Go 的 pseudo-version 規則從 commit 時間與雜湊算出來，因此與 GitHub
# 上那個 commit 對得起來。
#
# 用法：
#   tools/engine-proxy.sh            # 用 engine 的 HEAD
#   tools/engine-proxy.sh <commit>   # 指定 commit
#
# 產生後：
#   tools/go.sh get github.com/wicanr2/golden-box-remake-engine@<印出來的版本>
# （`tools/go.sh` 會自動帶上 GOPROXY 與 GOPRIVATE，見該檔）
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENGINE="${GOLDEN_BOX_REMAKE_ENGINE_DIR:-$ROOT/../golden-box-remake-engine}"
MODULE="github.com/wicanr2/golden-box-remake-engine"
OUT="$ROOT/workplace/engine-proxy"

[ -d "$ENGINE/.git" ] || { echo "找不到固定共用 engine repo：$ENGINE" >&2; exit 2; }

commit="$(git -C "$ENGINE" rev-parse "${1:-HEAD}")"
short="${commit:0:12}"
stamp="$(TZ=UTC git -C "$ENGINE" show -s --date='format-local:%Y%m%d%H%M%S' --format=%cd "$commit")"
rfc="$(TZ=UTC git -C "$ENGINE" show -s --date='format-local:%Y-%m-%dT%H:%M:%SZ' --format=%cd "$commit")"
version="v0.0.0-${stamp}-${short}"

# ⚠ 未提交的改動不會進 zip：模組內容一律取自 commit，不取自工作區。
# 否則 go.sum 的雜湊會對應到一份 GitHub 上不存在的東西。
#
# 這道檢查只在**沒有指定 commit**（也就是要打包 HEAD）時才有意義——那時
# 「我以為包進去了」的風險是真的。明確指定 commit 時打包的內容毫無歧義，
# 而且 `tools/engine-bootstrap.sh` 要能在開發者的 engine 工作區有改動時照樣跑。
if [ $# -eq 0 ] && [ -n "$(git -C "$ENGINE" status --porcelain)" ]; then
  echo "engine 工作區有未提交的改動；先 commit 再打包，或明確指定要打包哪個 commit" >&2
  exit 2
fi

dir="$OUT/$MODULE/@v"
mkdir -p "$dir"
printf '{"Version":"%s","Time":"%s"}\n' "$version" "$rfc" > "$dir/$version.info"
git -C "$ENGINE" show "$commit:go.mod" > "$dir/$version.mod"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
prefix="$MODULE@$version"
mkdir -p "$tmp/$prefix"
git -C "$ENGINE" archive --format=tar "$commit" | tar -x -C "$tmp/$prefix"
if command -v zip >/dev/null 2>&1; then
  ( cd "$tmp" && find "$prefix" -type f | sort | zip -X -q "$dir/$version.zip" -@ )
elif command -v python3 >/dev/null 2>&1; then
  # CoAB 的鎖版 test image 有 Python 標準庫但沒有 zip CLI；使用固定排序與
  # ZIP_STORED 產出符合 Go module proxy contract 的可重現 archive。
  python3 - "$tmp" "$prefix" "$dir/$version.zip" <<'PY'
import pathlib
import sys
import zipfile

root = pathlib.Path(sys.argv[1])
prefix = pathlib.Path(sys.argv[2])
output = pathlib.Path(sys.argv[3])
with zipfile.ZipFile(output, "w", compression=zipfile.ZIP_STORED) as archive:
    for path in sorted((root / prefix).rglob("*")):
        if path.is_file():
            archive.write(path, path.relative_to(root).as_posix())
PY
else
  echo "需要 zip 或 python3 才能建立 Go module archive" >&2
  exit 3
fi

# list 讓 `go list -m -versions` 之類的查詢也能運作。
grep -qxF "$version" "$dir/list" 2>/dev/null || echo "$version" >> "$dir/list"

echo "$version"
