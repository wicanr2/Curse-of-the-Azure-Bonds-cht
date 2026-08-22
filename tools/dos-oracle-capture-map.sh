#!/usr/bin/env bash
# 擷取一張第一人稱地圖：**先驗身分、再拍、拍完再驗一次**，任何一次不符就整批丟棄。
#
# ★ 為什麼要這樣。 擷取腳本只核對**座標**——座標對而圖錯時它一句話都不會說。
# 第 678 輪收了 54 張、第 682 輪收了 41 張，事後才發現拍到的是別張圖，兩批全廢。
# 第 682 輪那次**識別是成功的**，問題在於識別與擷取不在同一次開機、中間工具還
# 改過，於是「上次確認過」變成一句空話。
#
# ⇒ 身分驗證必須與擷取綁在**同一次呼叫**裡，而且前後各驗一次：前面那次確認起點
# 對，後面那次確認整批期間沒有飄走。
#
# 用法：
#   tools/dos-oracle-capture-map.sh <底檔> <索引檔> <area> <ecl段> <geo段> <前綴> <x> <y> [x y ...]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
base="$1"; index="$2"; AREA="$3"; ECL_BLOCK="$4"; GEO_BLOCK="$5"; PREFIX="$6"
shift 6
DST="$(cd "$(dirname "$index")" && pwd)"

# 預期指紋：兩格就分得開全部 16 張圖（cmd/geo-move-mask 的測試釘住這件事）。
expect="$(cd "$ROOT" && tools/go.sh run ./cmd/geo-move-mask -set "$AREA" -block "$GEO_BLOCK" \
  -cells 3,3:12,9 2>/dev/null | tail -1)"
want33="$(printf '%s' "$expect" | sed 's/.*(3,3)=\([0-9A-F]\).*/\1/')"
want129="$(printf '%s' "$expect" | sed 's/.*(12,9)=\([0-9A-F]\).*/\1/')"
echo "預期指紋：(3,3)=$want33 (12,9)=$want129"

# identify 回傳 0 表示這張圖是我們以為的那張。
identify() {
  local phase="$1" got33=0 got129=0 bit dir moved
  for spec in "3 3" "12 9"; do
    set -- $spec
    local mask=0 i=0
    for dir in N E S W; do
      moved="$(AREA=$AREA ECL_BLOCK=$ECL_BLOCK GEO_BLOCK=$GEO_BLOCK \
        "$ROOT/tools/dos-oracle-move-probe.sh" "$base" "$1" "$2" "$dir" 2>/dev/null | tail -1 | awk '{print $4}')"
      case "$moved" in 1) mask=$((mask | (1 << i)));; 0) ;; *) echo "  ⚠ $phase 探測失敗（$1,$2,$dir）"; return 1;; esac
      i=$((i+1))
    done
    if [ "$1" = "3" ]; then got33=$(printf '%X' "$mask"); else got129=$(printf '%X' "$mask"); fi
  done
  echo "  $phase 實測指紋：(3,3)=$got33 (12,9)=$got129"
  [ "$got33" = "$want33" ] && [ "$got129" = "$want129" ]
}

before="$(wc -l < "$index")"
if ! identify "拍之前"; then
  echo "✘ 起點就不是預期的那張圖：這一批不拍"; exit 1
fi

AREA=$AREA ECL_BLOCK=$ECL_BLOCK GEO_BLOCK=$GEO_BLOCK PREFIX=$PREFIX \
  "$ROOT/tools/dos-oracle-jump-capture.sh" "$base" "$index" "$@"

if ! identify "拍之後"; then
  echo "✘ 拍完之後身分對不上：**整批丟棄**"
  # 只刪這一次新增的那些列與檔案。
  tail -n +"$((before+1))" "$index" | awk '{print $1}' | while read -r name; do
    rm -f "$DST/$name"
  done
  head -n "$before" "$index" > "$index.tmp" && mv "$index.tmp" "$index"
  exit 1
fi
echo "✔ 前後兩次身分都對得上，這一批可以收"
