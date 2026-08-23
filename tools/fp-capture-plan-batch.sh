#!/usr/bin/env bash
# 由 docs/audit/fp-screen-plan.md 的貪婪順序產生（第 750 輪）。
# ⚠ 每一格約 45 秒：載入存檔只在主選單做得到，所以每格都要重開一次遊戲。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo5-b33（28 格）==="
AREA=5 ECL_BLOCK=51 GEO_BLOCK=51 PREFIX=geo5-b33 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 4 3 4 10 5 15 6 14 7 13 13 15 13 15 4 15 5 15 7 15 9 15 11 15 12 9 13 14 13 14 14 15 14 8 15 11 15 15 15 0 4 3 13 8 13 8 14 13 14 9 15 12 15 14 15
echo "=== geo3-b15（7 格）==="
AREA=3 ECL_BLOCK=21 GEO_BLOCK=21 PREFIX=geo3-b15 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 9 6 12 1 3 3 6 3 14 4 7 5 0 15
echo "=== geo6-b42（2 格）==="
AREA=6 ECL_BLOCK=66 GEO_BLOCK=66 PREFIX=geo6-b42 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 3 6 10 10
echo "=== geo3-b10（2 格）==="
AREA=3 ECL_BLOCK=16 GEO_BLOCK=16 PREFIX=geo3-b10 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 3 11 3 12
echo "=== geo3-b11（2 格）==="
AREA=3 ECL_BLOCK=17 GEO_BLOCK=17 PREFIX=geo3-b11 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 4 6 12 8
echo "=== geo4-b21（1 格）==="
AREA=4 ECL_BLOCK=33 GEO_BLOCK=33 PREFIX=geo4-b21 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 8 12
echo "=== geo4-b25（7 格）==="
AREA=4 ECL_BLOCK=37 GEO_BLOCK=37 PREFIX=geo4-b25 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 9 10 11 10 7 11 11 12 12 13 11 14 10 15
echo "=== geo5-b32（3 格）==="
AREA=5 ECL_BLOCK=50 GEO_BLOCK=50 PREFIX=geo5-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 3 2 4 3 9
echo "=== geo5-e31-b32（1 格）==="
AREA=5 ECL_BLOCK=49 GEO_BLOCK=50 PREFIX=geo5-e31-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 15
echo "=== geo5-b35（2 格）==="
AREA=5 ECL_BLOCK=53 GEO_BLOCK=53 PREFIX=geo5-b35 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 12 4 13 11
echo "=== geo6-b40（9 格）==="
AREA=6 ECL_BLOCK=64 GEO_BLOCK=64 PREFIX=geo6-b40 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 12 4 9 6 11 8 11 9 4 10 2 11 3 11 10 11 10 13
