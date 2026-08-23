#!/usr/bin/env bash
# 由 docs/audit/fp-screen-plan.md 的貪婪順序產生（第 751 輪，覆蓋 382／585 → 目標 466／585）。
# ⚠ 每一格約 60..90 秒：載入存檔只在主選單做得到，而且現在每張都要等畫面穩定
#   （settle）、四張拍完再用區域地圖核對位置。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo5-b33（8 格）==="
AREA=5 ECL_BLOCK=51 GEO_BLOCK=51 PREFIX=geo5-b33 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 15 12 14 13 14 14 15 14 11 15 15 15 15 11 9 13
echo "=== geo5-b32（1 格）==="
AREA=5 ECL_BLOCK=50 GEO_BLOCK=50 PREFIX=geo5-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 15
echo "=== geo6-b42（5 格）==="
AREA=6 ECL_BLOCK=66 GEO_BLOCK=66 PREFIX=geo6-b42 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 6 2 10 2 12 7 6 10 7 10
echo "=== geo2-b03（14 格）==="
AREA=2 ECL_BLOCK=3 GEO_BLOCK=3 PREFIX=geo2-b03 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 7 2 2 4 4 4 7 4 13 4 7 5 13 6 3 9 7 10 10 12 12 12 5 13 7 13 8 13
echo "=== geo2-b04（3 格）==="
AREA=2 ECL_BLOCK=4 GEO_BLOCK=4 PREFIX=geo2-b04 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 11 4 12 8 15 8
echo "=== geo3-b10（16 格）==="
AREA=3 ECL_BLOCK=16 GEO_BLOCK=16 PREFIX=geo3-b10 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 0 11 2 6 4 14 4 3 5 11 5 11 6 3 7 6 7 13 8 3 9 13 9 0 10 3 13 3 14 7 15
echo "=== geo3-b11（6 格）==="
AREA=3 ECL_BLOCK=17 GEO_BLOCK=17 PREFIX=geo3-b11 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 10 1 13 1 10 3 13 3 7 9 12 15
echo "=== geo3-b15（11 格）==="
AREA=3 ECL_BLOCK=21 GEO_BLOCK=21 PREFIX=geo3-b15 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 3 0 14 1 7 2 14 2 15 3 6 4 11 4 12 4 3 5 4 5 14 5
