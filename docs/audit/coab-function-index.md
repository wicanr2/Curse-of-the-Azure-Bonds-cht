# CoAB 全函式覆蓋台帳

本檔由 `cmd/re-ledger` 產生，不要手改。狀態來源是
`docs/audit/re-function-ledger.json`（人工判定），預設 `待解讀`。
函式全集來自 `tools/re-sweep.sh` 的 IDA 匯出，可重跑重生。

`引用` 欄只是提示：有文件提到同一個 overlay 與同一個十六進位值，
**不等於**該函式已被解讀；沒命中也不代表沒人寫過。

## DOS

模組 37／函式 1386：已解讀 918、不阻塞 133、邊界碎片 246、待解讀 89；已定義程式碼 260651 bytes，未定義 16065 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| START.EXE | — | 325 | 168 | 133 | 20 | 4 | 23290 | 0 | [明細](function-index/dos-START.EXE.md) |
| overlay-00 | MEMORY | 2 | 2 | 0 | 0 | 0 | 78 | 24 | [明細](function-index/dos-overlay-00.md) |
| overlay-01 | INTRO | 4 | 4 | 0 | 0 | 0 | 1511 | 405 | [明細](function-index/dos-overlay-01.md) |
| overlay-02 | INTERPET | 90 | 55 | 0 | 29 | 6 | 14062 | 292 | [明細](function-index/dos-overlay-02.md) |
| overlay-03 | PROTECT | 3 | 1 | 0 | 1 | 1 | 834 | 384 | [明細](function-index/dos-overlay-03.md) |
| overlay-04 | TEMPLE | 14 | 12 | 0 | 1 | 1 | 3435 | 340 | [明細](function-index/dos-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 13 | 0 | 3 | 1 | 5057 | 555 | [明細](function-index/dos-overlay-05.md) |
| overlay-06 | SHOP | 6 | 4 | 0 | 0 | 2 | 2146 | 274 | [明細](function-index/dos-overlay-06.md) |
| overlay-07 | ECL2 | 41 | 28 | 0 | 8 | 5 | 8745 | 191 | [明細](function-index/dos-overlay-07.md) |
| overlay-08 | COMBAT | 21 | 12 | 0 | 7 | 2 | 4243 | 343 | [明細](function-index/dos-overlay-08.md) |
| overlay-09 | COMPTACT | 38 | 15 | 0 | 22 | 1 | 7219 | 92 | [明細](function-index/dos-overlay-09.md) |
| overlay-10 | COMPREP | 38 | 23 | 0 | 14 | 1 | 6727 | 43 | [明細](function-index/dos-overlay-10.md) |
| overlay-11 | INIT | 4 | 3 | 0 | 1 | 0 | 2236 | 743 | [明細](function-index/dos-overlay-11.md) |
| overlay-12 | EFFPROCS | 152 | 126 | 0 | 25 | 1 | 12862 | 455 | [明細](function-index/dos-overlay-12.md) |
| overlay-13 | COMSTUFF | 58 | 38 | 0 | 15 | 5 | 17022 | 722 | [明細](function-index/dos-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 12 | 0 | 0 | 2 | 3735 | 308 | [明細](function-index/dos-overlay-14.md) |
| overlay-15 | CAMP | 42 | 32 | 0 | 10 | 0 | 8139 | 898 | [明細](function-index/dos-overlay-15.md) |
| overlay-16 | LOADSAVE | 28 | 11 | 0 | 8 | 9 | 16063 | 787 | [明細](function-index/dos-overlay-16.md) |
| overlay-17 | GEN | 39 | 9 | 0 | 21 | 9 | 19521 | 1536 | [明細](function-index/dos-overlay-17.md) |
| overlay-18 | ENDSTUFF | 11 | 8 | 0 | 0 | 3 | 4446 | 1177 | [明細](function-index/dos-overlay-18.md) |
| overlay-19 | LIBRARY | 42 | 22 | 0 | 15 | 5 | 13042 | 1293 | [明細](function-index/dos-overlay-19.md) |
| overlay-20 | CLOCK | 16 | 16 | 0 | 0 | 0 | 3536 | 191 | [明細](function-index/dos-overlay-20.md) |
| overlay-21 | MONEY | 27 | 20 | 0 | 4 | 3 | 6469 | 385 | [明細](function-index/dos-overlay-21.md) |
| overlay-22 | SPELLS | 139 | 111 | 0 | 26 | 2 | 25236 | 2174 | [明細](function-index/dos-overlay-22.md) |
| overlay-23 | EFFECTS | 40 | 25 | 0 | 12 | 3 | 9183 | 376 | [明細](function-index/dos-overlay-23.md) |
| overlay-24 | GENERIC | 56 | 47 | 0 | 3 | 6 | 12483 | 290 | [明細](function-index/dos-overlay-24.md) |
| overlay-25 | TRAINING | 15 | 13 | 0 | 0 | 2 | 4931 | 733 | [明細](function-index/dos-overlay-25.md) |
| overlay-26 | MENUS | 21 | 15 | 0 | 0 | 6 | 4571 | 163 | [明細](function-index/dos-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 5 | 0 | 0 | 0 | 134 | 0 | [明細](function-index/dos-overlay-27.md) |
| overlay-28 | DRAWWIN | 6 | 5 | 0 | 0 | 1 | 446 | 0 | [明細](function-index/dos-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 9 | 0 | 1 | 2 | 2102 | 114 | [明細](function-index/dos-overlay-29.md) |
| overlay-30 | THREED | 13 | 11 | 0 | 0 | 2 | 5205 | 84 | [明細](function-index/dos-overlay-30.md) |
| overlay-31 | LOS | 9 | 8 | 0 | 0 | 1 | 2999 | 0 | [明細](function-index/dos-overlay-31.md) |
| overlay-32 | TACMAP | 24 | 23 | 0 | 0 | 1 | 5064 | 0 | [明細](function-index/dos-overlay-32.md) |
| overlay-33 | SQRPAK24 | 7 | 6 | 0 | 0 | 1 | 1515 | 57 | [明細](function-index/dos-overlay-33.md) |
| overlay-34 | BUG | 3 | 2 | 0 | 0 | 1 | 1749 | 562 | [明細](function-index/dos-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 4 | 0 | 0 | 0 | 615 | 74 | [明細](function-index/dos-overlay-35.md) |

## PC98

模組 37／函式 1488：已解讀 1003、不阻塞 29、邊界碎片 338、待解讀 118；已定義程式碼 270352 bytes，未定義 20321 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| PC98-GAME.EXE | — | 333 | 260 | 29 | 25 | 19 | 25067 | 0 | [明細](function-index/pc98-PC98-GAME.EXE.md) |
| overlay-00 | MEMORY | 2 | 1 | 0 | 1 | 0 | 77 | 2 | [明細](function-index/pc98-overlay-00.md) |
| overlay-01 | INTRO | 4 | 3 | 0 | 1 | 0 | 2002 | 726 | [明細](function-index/pc98-overlay-01.md) |
| overlay-02 | INTERPET | 86 | 58 | 0 | 28 | 0 | 14044 | 266 | [明細](function-index/pc98-overlay-02.md) |
| overlay-03 | PROTECT | 4 | 3 | 0 | 0 | 1 | 752 | 278 | [明細](function-index/pc98-overlay-03.md) |
| overlay-04 | TEMPLE | 23 | 12 | 0 | 10 | 1 | 3710 | 830 | [明細](function-index/pc98-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 12 | 0 | 4 | 1 | 5212 | 458 | [明細](function-index/pc98-overlay-05.md) |
| overlay-06 | SHOP | 8 | 4 | 0 | 2 | 2 | 2363 | 283 | [明細](function-index/pc98-overlay-06.md) |
| overlay-07 | ECL2 | 46 | 30 | 0 | 16 | 0 | 8889 | 363 | [明細](function-index/pc98-overlay-07.md) |
| overlay-08 | COMBAT | 24 | 10 | 0 | 12 | 2 | 3967 | 372 | [明細](function-index/pc98-overlay-08.md) |
| overlay-09 | COMPTACT | 36 | 15 | 0 | 20 | 1 | 6325 | 111 | [明細](function-index/pc98-overlay-09.md) |
| overlay-10 | COMPREP | 35 | 21 | 0 | 12 | 2 | 7373 | 95 | [明細](function-index/pc98-overlay-10.md) |
| overlay-11 | INIT | 6 | 5 | 0 | 0 | 1 | 1960 | 128 | [明細](function-index/pc98-overlay-11.md) |
| overlay-12 | EFFPROCS | 150 | 130 | 0 | 18 | 2 | 13011 | 578 | [明細](function-index/pc98-overlay-12.md) |
| overlay-13 | COMSTUFF | 72 | 40 | 0 | 27 | 5 | 17459 | 905 | [明細](function-index/pc98-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 11 | 0 | 0 | 3 | 3435 | 321 | [明細](function-index/pc98-overlay-14.md) |
| overlay-15 | CAMP | 45 | 30 | 0 | 15 | 0 | 8121 | 1126 | [明細](function-index/pc98-overlay-15.md) |
| overlay-16 | LOADSAVE | 47 | 15 | 0 | 14 | 18 | 18585 | 1552 | [明細](function-index/pc98-overlay-16.md) |
| overlay-17 | GEN | 54 | 11 | 0 | 34 | 9 | 20976 | 2752 | [明細](function-index/pc98-overlay-17.md) |
| overlay-18 | ENDSTUFF | 23 | 15 | 0 | 4 | 4 | 4957 | 1311 | [明細](function-index/pc98-overlay-18.md) |
| overlay-19 | LIBRARY | 57 | 23 | 0 | 29 | 5 | 12619 | 1629 | [明細](function-index/pc98-overlay-19.md) |
| overlay-20 | CLOCK | 16 | 16 | 0 | 0 | 0 | 3555 | 225 | [明細](function-index/pc98-overlay-20.md) |
| overlay-21 | MONEY | 30 | 20 | 0 | 8 | 2 | 7968 | 2243 | [明細](function-index/pc98-overlay-21.md) |
| overlay-22 | SPELLS | 138 | 95 | 0 | 27 | 16 | 25053 | 2127 | [明細](function-index/pc98-overlay-22.md) |
| overlay-23 | EFFECTS | 39 | 27 | 0 | 12 | 0 | 9201 | 458 | [明細](function-index/pc98-overlay-23.md) |
| overlay-24 | GENERIC | 65 | 40 | 0 | 17 | 8 | 13567 | 376 | [明細](function-index/pc98-overlay-24.md) |
| overlay-25 | TRAINING | 14 | 12 | 0 | 0 | 2 | 5200 | 167 | [明細](function-index/pc98-overlay-25.md) |
| overlay-26 | MENUS | 20 | 15 | 0 | 0 | 5 | 4937 | 183 | [明細](function-index/pc98-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 5 | 0 | 0 | 0 | 134 | 0 | [明細](function-index/pc98-overlay-27.md) |
| overlay-28 | DRAWWIN | 3 | 2 | 0 | 0 | 1 | 196 | 0 | [明細](function-index/pc98-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 10 | 0 | 1 | 1 | 1846 | 114 | [明細](function-index/pc98-overlay-29.md) |
| overlay-30 | THREED | 13 | 11 | 0 | 0 | 2 | 4851 | 117 | [明細](function-index/pc98-overlay-30.md) |
| overlay-31 | LOS | 9 | 8 | 0 | 0 | 1 | 2988 | 0 | [明細](function-index/pc98-overlay-31.md) |
| overlay-32 | TACMAP | 25 | 23 | 0 | 0 | 2 | 6784 | 0 | [明細](function-index/pc98-overlay-32.md) |
| overlay-33 | SQRPAK24 | 7 | 6 | 0 | 0 | 1 | 2106 | 146 | [明細](function-index/pc98-overlay-33.md) |
| overlay-34 | BUG | 2 | 2 | 0 | 0 | 0 | 14 | 0 | [明細](function-index/pc98-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 2 | 0 | 1 | 1 | 1048 | 79 | [明細](function-index/pc98-overlay-35.md) |
