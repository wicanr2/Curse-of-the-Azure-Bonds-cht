#!/usr/bin/env bash
# 由 docs/audit/fp-screen-plan.md 的貪婪順序產生（第 751 輪第二批，466／585 起算）。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo3-b15（7 格）==="
AREA=3 ECL_BLOCK=21 GEO_BLOCK=21 PREFIX=geo3-b15 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 6 7 8 7 5 8 8 10 1 11 6 12 1 15
echo "=== geo4-b20（7 格）==="
AREA=4 ECL_BLOCK=32 GEO_BLOCK=32 PREFIX=geo4-b20 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 11 5 7 9 0 11 2 11 4 11 0 12 4 12
echo "=== geo4-b21（5 格）==="
AREA=4 ECL_BLOCK=33 GEO_BLOCK=33 PREFIX=geo4-b21 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 7 6 8 6 5 7 7 8 6 9
echo "=== geo4-b25（18 格）==="
AREA=4 ECL_BLOCK=37 GEO_BLOCK=37 PREFIX=geo4-b25 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 3 1 3 12 9 14 9 10 10 14 10 15 10 9 11 9 12 14 12 15 12 9 13 9 14 14 14 15 14 9 15 11 15 15 15
echo "=== geo5-b32（11 格）==="
AREA=5 ECL_BLOCK=50 GEO_BLOCK=50 PREFIX=geo5-b32 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 1 15 1 0 2 1 2 6 2 13 2 13 3 3 6 0 10 11 12 14 12
echo "=== geo5-b33（12 格）==="
AREA=5 ECL_BLOCK=51 GEO_BLOCK=51 PREFIX=geo5-b33 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 0 1 3 1 2 3 4 4 15 8 10 9 15 10 10 13 11 13 9 14 10 15 13 15
echo "=== geo5-b35（4 格）==="
AREA=5 ECL_BLOCK=53 GEO_BLOCK=53 PREFIX=geo5-b35 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 0 1 2 8 4 9 4
