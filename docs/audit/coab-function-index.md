# CoAB 全函式覆蓋台帳

本檔由 `cmd/re-ledger` 產生，不要手改。狀態來源是
`docs/audit/re-function-ledger.json`（人工判定），預設 `待解讀`。
函式全集來自 `tools/re-sweep.sh` 的 IDA 匯出，可重跑重生。

`引用` 欄只是提示：有文件提到同一個 overlay 與同一個十六進位值，
**不等於**該函式已被解讀；沒命中也不代表沒人寫過。

## DOS

模組 37／函式 1344：已解讀 186、不阻塞 133、邊界碎片 160、待解讀 865；已定義程式碼 238602 bytes，未定義 16044 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| START.EXE | — | 325 | 85 | 133 | 9 | 98 | 23290 | 0 | [明細](function-index/dos-START.EXE.md) |
| overlay-00 | MEMORY | 1 | 0 | 0 | 0 | 1 | 58 | 24 | [明細](function-index/dos-overlay-00.md) |
| overlay-01 | INTRO | 3 | 0 | 0 | 0 | 3 | 1504 | 405 | [明細](function-index/dos-overlay-01.md) |
| overlay-02 | INTERPET | 90 | 6 | 0 | 17 | 67 | 14062 | 292 | [明細](function-index/dos-overlay-02.md) |
| overlay-03 | PROTECT | 3 | 1 | 0 | 1 | 1 | 834 | 384 | [明細](function-index/dos-overlay-03.md) |
| overlay-04 | TEMPLE | 9 | 1 | 0 | 1 | 7 | 2107 | 334 | [明細](function-index/dos-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 2 | 0 | 1 | 14 | 5057 | 555 | [明細](function-index/dos-overlay-05.md) |
| overlay-06 | SHOP | 6 | 2 | 0 | 0 | 4 | 2146 | 274 | [明細](function-index/dos-overlay-06.md) |
| overlay-07 | ECL2 | 41 | 8 | 0 | 3 | 30 | 8745 | 191 | [明細](function-index/dos-overlay-07.md) |
| overlay-08 | COMBAT | 21 | 2 | 0 | 4 | 15 | 4243 | 343 | [明細](function-index/dos-overlay-08.md) |
| overlay-09 | COMPTACT | 37 | 1 | 0 | 18 | 18 | 7212 | 92 | [明細](function-index/dos-overlay-09.md) |
| overlay-10 | COMPREP | 38 | 3 | 0 | 7 | 28 | 6727 | 43 | [明細](function-index/dos-overlay-10.md) |
| overlay-11 | INIT | 2 | 1 | 0 | 0 | 1 | 563 | 743 | [明細](function-index/dos-overlay-11.md) |
| overlay-12 | EFFPROCS | 151 | 15 | 0 | 24 | 112 | 12855 | 455 | [明細](function-index/dos-overlay-12.md) |
| overlay-13 | COMSTUFF | 57 | 1 | 0 | 15 | 41 | 16498 | 722 | [明細](function-index/dos-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 1 | 0 | 0 | 13 | 3735 | 308 | [明細](function-index/dos-overlay-14.md) |
| overlay-15 | CAMP | 34 | 4 | 0 | 6 | 24 | 6009 | 887 | [明細](function-index/dos-overlay-15.md) |
| overlay-16 | LOADSAVE | 25 | 1 | 0 | 3 | 21 | 14690 | 785 | [明細](function-index/dos-overlay-16.md) |
| overlay-17 | GEN | 34 | 0 | 0 | 12 | 22 | 15142 | 1534 | [明細](function-index/dos-overlay-17.md) |
| overlay-18 | ENDSTUFF | 11 | 1 | 0 | 0 | 10 | 4446 | 1177 | [明細](function-index/dos-overlay-18.md) |
| overlay-19 | LIBRARY | 42 | 1 | 0 | 10 | 31 | 13042 | 1293 | [明細](function-index/dos-overlay-19.md) |
| overlay-20 | CLOCK | 16 | 1 | 0 | 0 | 15 | 3536 | 191 | [明細](function-index/dos-overlay-20.md) |
| overlay-21 | MONEY | 26 | 1 | 0 | 2 | 23 | 4559 | 385 | [明細](function-index/dos-overlay-21.md) |
| overlay-22 | SPELLS | 136 | 26 | 0 | 20 | 90 | 22254 | 2174 | [明細](function-index/dos-overlay-22.md) |
| overlay-23 | EFFECTS | 39 | 2 | 0 | 5 | 32 | 9176 | 376 | [明細](function-index/dos-overlay-23.md) |
| overlay-24 | GENERIC | 56 | 1 | 0 | 1 | 54 | 12483 | 290 | [明細](function-index/dos-overlay-24.md) |
| overlay-25 | TRAINING | 8 | 2 | 0 | 0 | 6 | 1295 | 733 | [明細](function-index/dos-overlay-25.md) |
| overlay-26 | MENUS | 21 | 2 | 0 | 0 | 19 | 4571 | 163 | [明細](function-index/dos-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 3 | 0 | 0 | 2 | 134 | 0 | [明細](function-index/dos-overlay-27.md) |
| overlay-28 | DRAWWIN | 6 | 1 | 0 | 0 | 5 | 446 | 0 | [明細](function-index/dos-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 2 | 0 | 1 | 9 | 2102 | 114 | [明細](function-index/dos-overlay-29.md) |
| overlay-30 | THREED | 12 | 1 | 0 | 0 | 11 | 3146 | 84 | [明細](function-index/dos-overlay-30.md) |
| overlay-31 | LOS | 9 | 2 | 0 | 0 | 7 | 2999 | 0 | [明細](function-index/dos-overlay-31.md) |
| overlay-32 | TACMAP | 24 | 1 | 0 | 0 | 23 | 5064 | 0 | [明細](function-index/dos-overlay-32.md) |
| overlay-33 | SQRPAK24 | 6 | 1 | 0 | 0 | 5 | 1508 | 57 | [明細](function-index/dos-overlay-33.md) |
| overlay-34 | BUG | 3 | 2 | 0 | 0 | 1 | 1749 | 562 | [明細](function-index/dos-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 2 | 0 | 0 | 2 | 615 | 74 | [明細](function-index/dos-overlay-35.md) |

## PC98

模組 37／函式 1481：已解讀 193、不阻塞 29、邊界碎片 226、待解讀 1033；已定義程式碼 269027 bytes，未定義 20319 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| PC98-GAME.EXE | — | 333 | 74 | 29 | 17 | 213 | 25067 | 0 | [明細](function-index/pc98-PC98-GAME.EXE.md) |
| overlay-00 | MEMORY | 2 | 0 | 0 | 1 | 1 | 77 | 2 | [明細](function-index/pc98-overlay-00.md) |
| overlay-01 | INTRO | 4 | 0 | 0 | 1 | 3 | 2002 | 726 | [明細](function-index/pc98-overlay-01.md) |
| overlay-02 | INTERPET | 85 | 5 | 0 | 19 | 61 | 14037 | 266 | [明細](function-index/pc98-overlay-02.md) |
| overlay-03 | PROTECT | 3 | 1 | 0 | 0 | 2 | 745 | 276 | [明細](function-index/pc98-overlay-03.md) |
| overlay-04 | TEMPLE | 23 | 1 | 0 | 9 | 13 | 3710 | 830 | [明細](function-index/pc98-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 1 | 0 | 3 | 13 | 5212 | 458 | [明細](function-index/pc98-overlay-05.md) |
| overlay-06 | SHOP | 8 | 1 | 0 | 2 | 5 | 2363 | 283 | [明細](function-index/pc98-overlay-06.md) |
| overlay-07 | ECL2 | 46 | 6 | 0 | 9 | 31 | 8889 | 363 | [明細](function-index/pc98-overlay-07.md) |
| overlay-08 | COMBAT | 24 | 0 | 0 | 9 | 15 | 3967 | 372 | [明細](function-index/pc98-overlay-08.md) |
| overlay-09 | COMPTACT | 36 | 2 | 0 | 16 | 18 | 6325 | 111 | [明細](function-index/pc98-overlay-09.md) |
| overlay-10 | COMPREP | 35 | 0 | 0 | 6 | 29 | 7373 | 95 | [明細](function-index/pc98-overlay-10.md) |
| overlay-11 | INIT | 6 | 1 | 0 | 0 | 5 | 1960 | 128 | [明細](function-index/pc98-overlay-11.md) |
| overlay-12 | EFFPROCS | 149 | 38 | 0 | 16 | 95 | 13004 | 578 | [明細](function-index/pc98-overlay-12.md) |
| overlay-13 | COMSTUFF | 71 | 4 | 0 | 19 | 48 | 17224 | 905 | [明細](function-index/pc98-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 1 | 0 | 0 | 13 | 3435 | 321 | [明細](function-index/pc98-overlay-14.md) |
| overlay-15 | CAMP | 45 | 3 | 0 | 8 | 34 | 8121 | 1126 | [明細](function-index/pc98-overlay-15.md) |
| overlay-16 | LOADSAVE | 47 | 2 | 0 | 5 | 40 | 18585 | 1552 | [明細](function-index/pc98-overlay-16.md) |
| overlay-17 | GEN | 54 | 2 | 0 | 20 | 32 | 20976 | 2752 | [明細](function-index/pc98-overlay-17.md) |
| overlay-18 | ENDSTUFF | 23 | 1 | 0 | 0 | 22 | 4957 | 1311 | [明細](function-index/pc98-overlay-18.md) |
| overlay-19 | LIBRARY | 56 | 2 | 0 | 22 | 32 | 11564 | 1629 | [明細](function-index/pc98-overlay-19.md) |
| overlay-20 | CLOCK | 15 | 0 | 0 | 0 | 15 | 3548 | 225 | [明細](function-index/pc98-overlay-20.md) |
| overlay-21 | MONEY | 30 | 2 | 0 | 3 | 25 | 7968 | 2243 | [明細](function-index/pc98-overlay-21.md) |
| overlay-22 | SPELLS | 138 | 24 | 0 | 22 | 92 | 25053 | 2127 | [明細](function-index/pc98-overlay-22.md) |
| overlay-23 | EFFECTS | 39 | 5 | 0 | 10 | 24 | 9201 | 458 | [明細](function-index/pc98-overlay-23.md) |
| overlay-24 | GENERIC | 65 | 1 | 0 | 7 | 57 | 13567 | 376 | [明細](function-index/pc98-overlay-24.md) |
| overlay-25 | TRAINING | 14 | 1 | 0 | 0 | 13 | 5200 | 167 | [明細](function-index/pc98-overlay-25.md) |
| overlay-26 | MENUS | 20 | 1 | 0 | 0 | 19 | 4937 | 183 | [明細](function-index/pc98-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 3 | 0 | 0 | 2 | 134 | 0 | [明細](function-index/pc98-overlay-27.md) |
| overlay-28 | DRAWWIN | 3 | 1 | 0 | 0 | 2 | 196 | 0 | [明細](function-index/pc98-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 2 | 0 | 1 | 9 | 1846 | 114 | [明細](function-index/pc98-overlay-29.md) |
| overlay-30 | THREED | 13 | 1 | 0 | 0 | 12 | 4851 | 117 | [明細](function-index/pc98-overlay-30.md) |
| overlay-31 | LOS | 9 | 2 | 0 | 0 | 7 | 2988 | 0 | [明細](function-index/pc98-overlay-31.md) |
| overlay-32 | TACMAP | 25 | 1 | 0 | 0 | 24 | 6784 | 0 | [明細](function-index/pc98-overlay-32.md) |
| overlay-33 | SQRPAK24 | 6 | 1 | 0 | 0 | 5 | 2099 | 146 | [明細](function-index/pc98-overlay-33.md) |
| overlay-34 | BUG | 2 | 2 | 0 | 0 | 0 | 14 | 0 | [明細](function-index/pc98-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 1 | 0 | 1 | 2 | 1048 | 79 | [明細](function-index/pc98-overlay-35.md) |
