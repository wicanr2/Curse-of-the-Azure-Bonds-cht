#!/usr/bin/env bash
# 第 710 輪：補 fp-screen-plan 貪婪清單剩下的 6 格（579/585 → 目標 585/585）。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo5-b35（3 格）==="
AREA=5 ECL_BLOCK=53 GEO_BLOCK=53 PREFIX=geo5-b35 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 11 9 3 11 8 11
echo "=== geo6-b40（3 格）==="
AREA=6 ECL_BLOCK=64 GEO_BLOCK=64 PREFIX=geo6-b40 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 5 4 11 6 4 7
