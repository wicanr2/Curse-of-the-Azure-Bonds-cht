#!/usr/bin/env bash
# 以既有 Docker 工具鏈產生 Linux AppImage、Windows ZIP 與 macOS 雙架構 ZIP。
# 用法：tools/package-release.sh <版本>
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
if [[ -z "$VERSION" || ! "$VERSION" =~ ^[0-9A-Za-z._-]+$ ]]; then
  echo "用法：tools/package-release.sh <版本（僅英數、點、底線、連字號）>" >&2
  exit 2
fi

GO_IMAGE="coab-go-ebiten:1.24"
MAC_IMAGE="u2cht-osxcross:20260826-r1"
APPIMAGE_IMAGE="u5cht/appimage:latest"
OUT="$ROOT/dist/$VERSION"
UID_NOW="$(id -u)"
GID_NOW="$(id -g)"

for image in "$GO_IMAGE" "$MAC_IMAGE" "$APPIMAGE_IMAGE"; do
  docker image inspect "$image" >/dev/null
done
test -f "$ROOT/workplace/engine-proxy/github.com/wicanr2/golden-box-remake-engine/@v/list"

docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
  -u "$UID_NOW:$GID_NOW" -v "$ROOT:/src" -w /src "$GO_IMAGE" sh -c '
    set -eu
    rm -rf "dist/'"$VERSION"'"
    mkdir -p "dist/'"$VERSION"'/build" "dist/'"$VERSION"'/patch" "dist/'"$VERSION"'/full-local"
  '

common_go_mounts=(
  -v "$ROOT:/src"
  -v "$ROOT/workplace/go-build-cache:/gocache"
  -v "$ROOT/workplace/go-mod-cache:/gomod"
  -e GOCACHE=/gocache -e GOMODCACHE=/gomod
  -e GOFLAGS=-mod=mod\ -buildvcs=false
  -e GOPROXY=file:///src/workplace/engine-proxy
  -e GOSUMDB=off
  -w /src
)

docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
  -u "$UID_NOW:$GID_NOW" "${common_go_mounts[@]}" \
  -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=1 "$GO_IMAGE" \
  go build -trimpath -ldflags=-s\ -w -o "dist/$VERSION/build/azure-bonds-game-linux-amd64" ./cmd/azure-bonds-game

docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
  -u "$UID_NOW:$GID_NOW" "${common_go_mounts[@]}" \
  -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 "$GO_IMAGE" \
  go build -trimpath -ldflags=-s\ -w -o "dist/$VERSION/build/azure-bonds-game.exe" ./cmd/azure-bonds-game

for tuple in "amd64:o64-clang:o64-clang++" "arm64:oa64-clang:oa64-clang++"; do
	arch="${tuple%%:*}"
	rest="${tuple#*:}"
	compiler="${rest%%:*}"
	cxx="${rest##*:}"
  docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
    -u "$UID_NOW:$GID_NOW" "${common_go_mounts[@]}" \
    -e GOOS=darwin -e GOARCH="$arch" -e CGO_ENABLED=1 -e CC="$compiler" -e CXX="$cxx" "$MAC_IMAGE" \
    go build -trimpath -ldflags=-s\ -w -o "dist/$VERSION/build/azure-bonds-game-darwin-$arch" ./cmd/azure-bonds-game
done

# 所有檔案操作也留在一次性容器內。patch 明確排除原版 ZIP 與 PC-98 BGM；
# full-local 只有本機存在原版 ZIP時才產生，且 dist/ 整體由 .gitignore 排除。
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
  -u "$UID_NOW:$GID_NOW" -v "$ROOT:/src" -w /src "$GO_IMAGE" sh -c '
    set -eu
    V='"$VERSION"'
    B="dist/$V/build"
    for flavor in patch full-local; do
      BASE="dist/$V/$flavor"
      mkdir -p "$BASE/linux/AppDir/usr/bin" "$BASE/windows" \
        "$BASE/macos-amd64/Azure Bonds Remake.app/Contents/MacOS" \
        "$BASE/macos-arm64/Azure Bonds Remake.app/Contents/MacOS"
      cp -R assets "$BASE/linux/AppDir/"
      cp -R assets "$BASE/windows/"
      cp -R assets "$BASE/macos-amd64/Azure Bonds Remake.app/Contents/MacOS/"
      cp -R assets "$BASE/macos-arm64/Azure Bonds Remake.app/Contents/MacOS/"
      find "$BASE" -type f -name "pc98-bgm-selector-*.ogg" -delete
      cp packaging/README-發行包.md "$BASE/linux/AppDir/README.md"
      cp packaging/README-發行包.md "$BASE/windows/README.md"
      cp packaging/README-發行包.md "$BASE/macos-amd64/Azure Bonds Remake.app/Contents/MacOS/README.md"
      cp packaging/README-發行包.md "$BASE/macos-arm64/Azure Bonds Remake.app/Contents/MacOS/README.md"
      for target in \
        "$BASE/linux/AppDir" \
        "$BASE/windows" \
        "$BASE/macos-amd64/Azure Bonds Remake.app/Contents/MacOS" \
        "$BASE/macos-arm64/Azure Bonds Remake.app/Contents/MacOS"; do
        cp LICENSE NOTICE.md "$target/"
      done
    done
    cp "$B/azure-bonds-game-linux-amd64" "dist/$V/patch/linux/AppDir/usr/bin/azure-bonds-game"
    cp "$B/azure-bonds-game.exe" "dist/$V/patch/windows/azure-bonds-game.exe"
    for arch in amd64 arm64; do
      APP="dist/$V/patch/macos-$arch/Azure Bonds Remake.app/Contents"
      cp packaging/macos/launcher "$APP/MacOS/azure-bonds-game"
      cp "$B/azure-bonds-game-darwin-$arch" "$APP/MacOS/azure-bonds-game-bin"
      sed "s/@VERSION@/$V/g" packaging/macos/Info.plist > "$APP/Info.plist"
      chmod 0755 "$APP/MacOS/azure-bonds-game" "$APP/MacOS/azure-bonds-game-bin"
    done
    cp packaging/windows/啟動遊戲.bat "dist/$V/patch/windows/"
    cp packaging/linux/AppRun "dist/$V/patch/linux/AppDir/AppRun"
    cp packaging/linux/azure-bonds-remake.desktop "dist/$V/patch/linux/AppDir/"
    cp assets/sprites/bigpic1-block-79-item-00.png "dist/$V/patch/linux/AppDir/azure-bonds-remake.png"
    chmod 0755 "dist/$V/patch/linux/AppDir/AppRun" "dist/$V/patch/linux/AppDir/usr/bin/azure-bonds-game"

    # AppImage 自帶非 glibc 的執行期依賴，避免只在建置 image 能啟動。
    mkdir -p "dist/$V/patch/linux/AppDir/usr/lib"
    ldd "$B/azure-bonds-game-linux-amd64" | awk "/=> \/.*\// {print \$3}" | while read -r lib; do
      case "$(basename "$lib")" in
        libc.so.*|libm.so.*|libdl.so.*|libpthread.so.*|librt.so.*|ld-linux-*) continue ;;
      esac
      cp -L "$lib" "dist/$V/patch/linux/AppDir/usr/lib/"
    done

    if [ -f curseoftheazurebonds.zip ]; then
      cp -R "dist/$V/patch/." "dist/$V/full-local/"
      for target in \
        "dist/$V/full-local/linux/AppDir" \
        "dist/$V/full-local/windows" \
        "dist/$V/full-local/macos-amd64/Azure Bonds Remake.app/Contents/MacOS" \
        "dist/$V/full-local/macos-arm64/Azure Bonds Remake.app/Contents/MacOS"; do
        cp curseoftheazurebonds.zip "$target/"
        cp assets/audio/pc98-bgm-selector-*.ogg "$target/assets/audio/"
      done
    fi
  '

for flavor in patch full-local; do
  [[ -f "$OUT/$flavor/linux/AppDir/AppRun" ]] || continue
  docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
    -u "$UID_NOW:$GID_NOW" -v "$ROOT:/src" -w /src "$APPIMAGE_IMAGE" bash -c \
    "ARCH=x86_64 appimagetool dist/$VERSION/$flavor/linux/AppDir dist/$VERSION/$flavor/azure-bonds-remake-$VERSION-x86_64.AppImage"
  docker run --rm --network none --memory 512m --cpus 1 --pids-limit 128 \
    -u "$UID_NOW:$GID_NOW" -v "$ROOT:/src" -w /src python:3.12-slim python -c '
import pathlib, sys, zipfile
root = pathlib.Path(sys.argv[1])
for source, name in [
    (root / "windows", f"azure-bonds-remake-{sys.argv[2]}-windows-x86_64.zip"),
    (root / "macos-amd64", f"azure-bonds-remake-{sys.argv[2]}-macos-x86_64.zip"),
    (root / "macos-arm64", f"azure-bonds-remake-{sys.argv[2]}-macos-arm64.zip"),
]:
    with zipfile.ZipFile(root / name, "w", zipfile.ZIP_DEFLATED) as archive:
        for path in sorted(source.rglob("*")):
            if path.is_file():
                archive.write(path, path.relative_to(source))
' "dist/$VERSION/$flavor" "$VERSION"
done

docker run --rm --network none --memory 512m --cpus 1 --pids-limit 128 \
  -u "$UID_NOW:$GID_NOW" -v "$ROOT:/src" -w /src python:3.12-slim python -c '
import hashlib, json, pathlib, sys
root = pathlib.Path(sys.argv[1])
rows = []
artifacts = []
for flavor in ("patch", "full-local"):
    artifacts.extend((root / flavor).glob("*.AppImage"))
    artifacts.extend((root / flavor).glob("*.zip"))
for path in sorted(artifacts):
    data = path.read_bytes()
    rows.append({"file": str(path.relative_to(root)), "bytes": len(data), "sha256": hashlib.sha256(data).hexdigest()})
(root / "SHA256SUMS.json").write_text(json.dumps(rows, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
' "dist/$VERSION"

echo "完成：$OUT"
