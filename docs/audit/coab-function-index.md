# CoAB 全函式覆蓋台帳

本檔由 `cmd/re-ledger` 產生，不要手改。狀態來源是
`docs/audit/re-function-ledger.json`（人工判定），預設 `待解讀`。
函式全集來自 `tools/re-sweep.sh` 的 IDA 匯出，可重跑重生。

`引用` 欄只是提示：有文件提到同一個 overlay 與同一個十六進位值，
**不等於**該函式已被解讀；沒命中也不代表沒人寫過。

## DOS

模組 37／函式 1386：已解讀 447、不阻塞 133、邊界碎片 249、待解讀 557；已定義程式碼 260651 bytes，未定義 16065 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| START.EXE | — | 325 | 107 | 133 | 17 | 68 | 23290 | 0 | [明細](function-index/dos-START.EXE.md) |
| overlay-00 | MEMORY | 2 | 0 | 0 | 0 | 2 | 78 | 24 | [明細](function-index/dos-overlay-00.md) |
| overlay-01 | INTRO | 4 | 0 | 0 | 0 | 4 | 1511 | 405 | [明細](function-index/dos-overlay-01.md) |
| overlay-02 | INTERPET | 90 | 47 | 0 | 29 | 14 | 14062 | 292 | [明細](function-index/dos-overlay-02.md) |
| overlay-03 | PROTECT | 3 | 1 | 0 | 1 | 1 | 834 | 384 | [明細](function-index/dos-overlay-03.md) |
| overlay-04 | TEMPLE | 14 | 1 | 0 | 1 | 12 | 3435 | 340 | [明細](function-index/dos-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 3 | 0 | 3 | 11 | 5057 | 555 | [明細](function-index/dos-overlay-05.md) |
| overlay-06 | SHOP | 6 | 2 | 0 | 0 | 4 | 2146 | 274 | [明細](function-index/dos-overlay-06.md) |
| overlay-07 | ECL2 | 41 | 18 | 0 | 8 | 15 | 8745 | 191 | [明細](function-index/dos-overlay-07.md) |
| overlay-08 | COMBAT | 21 | 3 | 0 | 7 | 11 | 4243 | 343 | [明細](function-index/dos-overlay-08.md) |
| overlay-09 | COMPTACT | 38 | 2 | 0 | 22 | 14 | 7219 | 92 | [明細](function-index/dos-overlay-09.md) |
| overlay-10 | COMPREP | 38 | 4 | 0 | 14 | 20 | 6727 | 43 | [明細](function-index/dos-overlay-10.md) |
| overlay-11 | INIT | 4 | 1 | 0 | 1 | 2 | 2236 | 743 | [明細](function-index/dos-overlay-11.md) |
| overlay-12 | EFFPROCS | 152 | 91 | 0 | 27 | 34 | 12862 | 455 | [明細](function-index/dos-overlay-12.md) |
| overlay-13 | COMSTUFF | 58 | 1 | 0 | 16 | 41 | 17022 | 722 | [明細](function-index/dos-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 3 | 0 | 0 | 11 | 3735 | 308 | [明細](function-index/dos-overlay-14.md) |
| overlay-15 | CAMP | 42 | 5 | 0 | 11 | 26 | 8139 | 898 | [明細](function-index/dos-overlay-15.md) |
| overlay-16 | LOADSAVE | 28 | 1 | 0 | 8 | 19 | 16063 | 787 | [明細](function-index/dos-overlay-16.md) |
| overlay-17 | GEN | 39 | 0 | 0 | 21 | 18 | 19521 | 1536 | [明細](function-index/dos-overlay-17.md) |
| overlay-18 | ENDSTUFF | 11 | 2 | 0 | 0 | 9 | 4446 | 1177 | [明細](function-index/dos-overlay-18.md) |
| overlay-19 | LIBRARY | 42 | 2 | 0 | 15 | 25 | 13042 | 1293 | [明細](function-index/dos-overlay-19.md) |
| overlay-20 | CLOCK | 16 | 2 | 0 | 0 | 14 | 3536 | 191 | [明細](function-index/dos-overlay-20.md) |
| overlay-21 | MONEY | 27 | 6 | 0 | 4 | 17 | 6469 | 385 | [明細](function-index/dos-overlay-21.md) |
| overlay-22 | SPELLS | 139 | 57 | 0 | 26 | 56 | 25236 | 2174 | [明細](function-index/dos-overlay-22.md) |
| overlay-23 | EFFECTS | 40 | 21 | 0 | 12 | 7 | 9183 | 376 | [明細](function-index/dos-overlay-23.md) |
| overlay-24 | GENERIC | 56 | 38 | 0 | 3 | 15 | 12483 | 290 | [明細](function-index/dos-overlay-24.md) |
| overlay-25 | TRAINING | 15 | 2 | 0 | 2 | 11 | 4931 | 733 | [明細](function-index/dos-overlay-25.md) |
| overlay-26 | MENUS | 21 | 3 | 0 | 0 | 18 | 4571 | 163 | [明細](function-index/dos-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 3 | 0 | 0 | 2 | 134 | 0 | [明細](function-index/dos-overlay-27.md) |
| overlay-28 | DRAWWIN | 6 | 2 | 0 | 0 | 4 | 446 | 0 | [明細](function-index/dos-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 2 | 0 | 1 | 9 | 2102 | 114 | [明細](function-index/dos-overlay-29.md) |
| overlay-30 | THREED | 13 | 2 | 0 | 0 | 11 | 5205 | 84 | [明細](function-index/dos-overlay-30.md) |
| overlay-31 | LOS | 9 | 2 | 0 | 0 | 7 | 2999 | 0 | [明細](function-index/dos-overlay-31.md) |
| overlay-32 | TACMAP | 24 | 8 | 0 | 0 | 16 | 5064 | 0 | [明細](function-index/dos-overlay-32.md) |
| overlay-33 | SQRPAK24 | 7 | 1 | 0 | 0 | 6 | 1515 | 57 | [明細](function-index/dos-overlay-33.md) |
| overlay-34 | BUG | 3 | 2 | 0 | 0 | 1 | 1749 | 562 | [明細](function-index/dos-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 2 | 0 | 0 | 2 | 615 | 74 | [明細](function-index/dos-overlay-35.md) |

## PC98

模組 37／函式 1488：已解讀 593、不阻塞 29、邊界碎片 340、待解讀 526；已定義程式碼 270352 bytes，未定義 20321 bytes。

| 模組 | 原始單元 | 函式 | 已解讀 | 不阻塞 | 碎片 | 待解讀 | 程式碼 | 未定義 | 明細 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|
| PC98-GAME.EXE | — | 333 | 220 | 29 | 25 | 59 | 25067 | 0 | [明細](function-index/pc98-PC98-GAME.EXE.md) |
| overlay-00 | MEMORY | 2 | 0 | 0 | 1 | 1 | 77 | 2 | [明細](function-index/pc98-overlay-00.md) |
| overlay-01 | INTRO | 4 | 0 | 0 | 1 | 3 | 2002 | 726 | [明細](function-index/pc98-overlay-01.md) |
| overlay-02 | INTERPET | 86 | 58 | 0 | 28 | 0 | 14044 | 266 | [明細](function-index/pc98-overlay-02.md) |
| overlay-03 | PROTECT | 4 | 1 | 0 | 0 | 3 | 752 | 278 | [明細](function-index/pc98-overlay-03.md) |
| overlay-04 | TEMPLE | 23 | 2 | 0 | 10 | 11 | 3710 | 830 | [明細](function-index/pc98-overlay-04.md) |
| overlay-05 | POSTCOM | 17 | 2 | 0 | 4 | 11 | 5212 | 458 | [明細](function-index/pc98-overlay-05.md) |
| overlay-06 | SHOP | 8 | 2 | 0 | 2 | 4 | 2363 | 283 | [明細](function-index/pc98-overlay-06.md) |
| overlay-07 | ECL2 | 46 | 30 | 0 | 16 | 0 | 8889 | 363 | [明細](function-index/pc98-overlay-07.md) |
| overlay-08 | COMBAT | 24 | 1 | 0 | 12 | 11 | 3967 | 372 | [明細](function-index/pc98-overlay-08.md) |
| overlay-09 | COMPTACT | 36 | 3 | 0 | 20 | 13 | 6325 | 111 | [明細](function-index/pc98-overlay-09.md) |
| overlay-10 | COMPREP | 35 | 2 | 0 | 12 | 21 | 7373 | 95 | [明細](function-index/pc98-overlay-10.md) |
| overlay-11 | INIT | 6 | 2 | 0 | 0 | 4 | 1960 | 128 | [明細](function-index/pc98-overlay-11.md) |
| overlay-12 | EFFPROCS | 150 | 99 | 0 | 19 | 32 | 13011 | 578 | [明細](function-index/pc98-overlay-12.md) |
| overlay-13 | COMSTUFF | 72 | 3 | 0 | 28 | 41 | 17459 | 905 | [明細](function-index/pc98-overlay-13.md) |
| overlay-14 | MOVEMENT | 14 | 3 | 0 | 0 | 11 | 3435 | 321 | [明細](function-index/pc98-overlay-14.md) |
| overlay-15 | CAMP | 45 | 4 | 0 | 15 | 26 | 8121 | 1126 | [明細](function-index/pc98-overlay-15.md) |
| overlay-16 | LOADSAVE | 47 | 7 | 0 | 14 | 26 | 18585 | 1552 | [明細](function-index/pc98-overlay-16.md) |
| overlay-17 | GEN | 54 | 3 | 0 | 34 | 17 | 20976 | 2752 | [明細](function-index/pc98-overlay-17.md) |
| overlay-18 | ENDSTUFF | 23 | 7 | 0 | 4 | 12 | 4957 | 1311 | [明細](function-index/pc98-overlay-18.md) |
| overlay-19 | LIBRARY | 57 | 3 | 0 | 29 | 25 | 12619 | 1629 | [明細](function-index/pc98-overlay-19.md) |
| overlay-20 | CLOCK | 16 | 1 | 0 | 0 | 15 | 3555 | 225 | [明細](function-index/pc98-overlay-20.md) |
| overlay-21 | MONEY | 30 | 7 | 0 | 8 | 15 | 7968 | 2243 | [明細](function-index/pc98-overlay-21.md) |
| overlay-22 | SPELLS | 138 | 55 | 0 | 27 | 56 | 25053 | 2127 | [明細](function-index/pc98-overlay-22.md) |
| overlay-23 | EFFECTS | 39 | 24 | 0 | 12 | 3 | 9201 | 458 | [明細](function-index/pc98-overlay-23.md) |
| overlay-24 | GENERIC | 65 | 27 | 0 | 17 | 21 | 13567 | 376 | [明細](function-index/pc98-overlay-24.md) |
| overlay-25 | TRAINING | 14 | 2 | 0 | 0 | 12 | 5200 | 167 | [明細](function-index/pc98-overlay-25.md) |
| overlay-26 | MENUS | 20 | 2 | 0 | 0 | 18 | 4937 | 183 | [明細](function-index/pc98-overlay-26.md) |
| overlay-27 | OVERLAND | 5 | 3 | 0 | 0 | 2 | 134 | 0 | [明細](function-index/pc98-overlay-27.md) |
| overlay-28 | DRAWWIN | 3 | 2 | 0 | 0 | 1 | 196 | 0 | [明細](function-index/pc98-overlay-28.md) |
| overlay-29 | PORTRAIT | 12 | 2 | 0 | 1 | 9 | 1846 | 114 | [明細](function-index/pc98-overlay-29.md) |
| overlay-30 | THREED | 13 | 2 | 0 | 0 | 11 | 4851 | 117 | [明細](function-index/pc98-overlay-30.md) |
| overlay-31 | LOS | 9 | 2 | 0 | 0 | 7 | 2988 | 0 | [明細](function-index/pc98-overlay-31.md) |
| overlay-32 | TACMAP | 25 | 8 | 0 | 0 | 17 | 6784 | 0 | [明細](function-index/pc98-overlay-32.md) |
| overlay-33 | SQRPAK24 | 7 | 1 | 0 | 0 | 6 | 2106 | 146 | [明細](function-index/pc98-overlay-33.md) |
| overlay-34 | BUG | 2 | 2 | 0 | 0 | 0 | 14 | 0 | [明細](function-index/pc98-overlay-34.md) |
| overlay-35 | SQRPAK8 | 4 | 1 | 0 | 1 | 2 | 1048 | 79 | [明細](function-index/pc98-overlay-35.md) |
