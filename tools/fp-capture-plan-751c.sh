#!/usr/bin/env bash
# 由 docs/audit/fp-screen-plan.md 的貪婪順序產生（第 751 輪第三批，530／585 起算）。
# 這一批把剩下的 55 種簽章全部蓋掉 → 585／585。
set -u
cd "$(dirname "$0")/.." || exit 1
IDX=docs/reference/original-dos/first-person/index.tsv
echo "=== geo5-b35（7 格）==="
AREA=5 ECL_BLOCK=53 GEO_BLOCK=53 PREFIX=geo5-b35 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 11 9 11 10 13 10 3 11 8 11 9 11 11 13
echo "=== geo6-b40（27 格）==="
AREA=6 ECL_BLOCK=64 GEO_BLOCK=64 PREFIX=geo6-b40 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 5 1 10 3 5 4 12 5 10 6 11 6 13 6 14 6 4 7 4 8 3 9 2 10 1 11 7 11 9 11 11 11 4 12 11 12 1 13 2 13 3 13 13 13 2 14 3 14 12 14 2 15 3 15
echo "=== geo6-b42（19 格）==="
AREA=6 ECL_BLOCK=66 GEO_BLOCK=66 PREFIX=geo6-b42 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 2 2 3 2 5 2 3 3 6 5 6 6 13 6 6 7 6 8 12 8 3 9 6 9 12 9 9 10 12 10 7 11 12 11 12 12 15 12
echo "=== geo6-b43（1 格）==="
AREA=6 ECL_BLOCK=67 GEO_BLOCK=67 PREFIX=geo6-b43 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 7 8
echo "=== geo6-b45（1 格）==="
AREA=6 ECL_BLOCK=69 GEO_BLOCK=69 PREFIX=geo6-b45 \
  tools/dos-oracle-jump-capture.sh workplace/orig-savgamb.dat "$IDX" 9 9
