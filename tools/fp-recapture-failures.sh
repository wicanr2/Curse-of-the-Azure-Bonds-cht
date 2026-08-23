#!/usr/bin/env bash
# 重拍所有目前逐格比對不過的格子（第 750 輪：擷取加了畫面穩定判定與區域地圖位置核對）。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo2-b01（1 格）==="
AREA=2 ECL_BLOCK=1 GEO_BLOCK=1 PREFIX=geo2-b01 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 1 7
echo "=== geo5-b32（2 格）==="
AREA=5 ECL_BLOCK=50 GEO_BLOCK=50 PREFIX=geo5-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 13 0 14
echo "=== geo5-b33（16 格）==="
AREA=5 ECL_BLOCK=51 GEO_BLOCK=51 PREFIX=geo5-b33 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 4 1 4 2 4 8 15 9 15 12 15 13 14 14 7 14 15 15 0 15 4 15 5 15 6 15 7 15 9 15 13
echo "=== geo5-e31-b32（1 格）==="
AREA=5 ECL_BLOCK=49 GEO_BLOCK=50 PREFIX=geo5-e31-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 15
