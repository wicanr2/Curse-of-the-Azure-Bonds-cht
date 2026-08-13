# 函式分流報告

由 `scripts/profile_triage.py` 產生，不要手改。**只排序，不改台帳狀態。**
分組是候選判斷：`thunk` 可能不只是 thunk，「只呼叫 RTL」也可能是關鍵邏輯。

## DOS（1344 個函式）

| 分組 | 數量 |
|---|---:|
| 其他 | 865 |
| 極小函式／thunk 候選 | 218 |
| 無呼叫且無寫入（純計算候選） | 158 |
| 軟體中斷（有 int） | 84 |
| 硬體存取（有 in／out） | 13 |
| 有字串參照 | 6 |

最常被呼叫的共用 routine（依 far call 次數）：

| routine | 次數 |
|---|---:|
| overlay-23 entry#9 ROLLDICE | 118 |
| overlay-24 entry#20 CHARMSG | 76 |
| overlay-32 entry#15 FINDX | 73 |
| overlay-32 entry#16 FINDY | 73 |
| overlay-24 entry#27 SPELLON | 71 |
| overlay-07 entry#1 ADDRESSVALUE | 55 |
| overlay-23 entry#11 ADDEFFECT | 49 |
| overlay-23 entry#3 SPELLOFF | 47 |
| overlay-07 entry#2 READVAR | 44 |
| overlay-24 entry#19 HORZMSG | 42 |
| overlay-07 entry#15 STOREVALUE | 33 |
| overlay-24 entry#34 QUIT | 32 |
| overlay-23 entry#8 MAKESAVE | 29 |
| overlay-26 entry#3 INIMNUBUF | 29 |
| overlay-24 entry#2 SHOWALL | 27 |

最大的 20 個函式（優先解讀候選）：

| 模組 | 位址 | 大小 | 被呼叫 | 主要 far call |
|---|---|---:|---:|---|
| overlay-16 | `228Eh` | 3994 | 1 | overlay-23 entry#11 ADDEFFECT×10、overlay-24 entry#46 ×4 |
| overlay-17 | `2868h` | 3454 | 1 | overlay-24 entry#7 DOCOM×3、overlay-19 entry#1 SHOWACTIVECHAR×1 |
| overlay-17 | `0782h` | 2622 | 1 | overlay-23 entry#11 ADDEFFECT×13、overlay-26 entry#8 YESNO×6 |
| overlay-23 | `03FEh` | 2313 | 9 | — |
| overlay-12 | `2E3Ch` | 1840 | 0 | — |
| overlay-17 | `3EE3h` | 1798 | 1 | overlay-16 entry#4 LOADACTIVEICON×7、overlay-32 entry#4 SETCOMBATCOLORS×1 |
| overlay-02 | `2086h` | 1791 | 1 | overlay-07 entry#15 STOREVALUE×15、overlay-07 entry#1 ADDRESSVALUE×6 |
| overlay-34 | `0240h` | 1735 | 0 | — |
| START.EXE | `12C1Dh` | 1611 | 1 | — |
| overlay-17 | `1ECDh` | 1608 | 3 | overlay-26 entry#6 CLEARMENU×2、overlay-19 entry#3 DISPLAYSTAT×1 |
| START.EXE | `1236Ch` | 1537 | 1 | — |
| overlay-16 | `3748h` | 1509 | 0 | overlay-26 entry#4 ITEMCOUNT×2、overlay-26 entry#3 INIMNUBUF×1 |
| overlay-16 | `0DEAh` | 1436 | 1 | overlay-26 entry#3 INIMNUBUF×1、overlay-24 entry#10 BYTETOSTR×1 |
| overlay-16 | `19EAh` | 1335 | 2 | — |
| overlay-13 | `2220h` | 1307 | 0 | overlay-32 entry#15 FINDX×6、overlay-32 entry#16 FINDY×6 |
| overlay-19 | `0083h` | 1300 | 7 | overlay-24 entry#1 PRINTITEMNAME×2、overlay-24 entry#22 PRINTCHARNAME×1 |
| overlay-13 | `33ACh` | 1278 | 2 | overlay-32 entry#15 FINDX×2、overlay-32 entry#16 FINDY×2 |
| overlay-18 | `10FFh` | 1265 | 0 | overlay-26 entry#4 ITEMCOUNT×3、overlay-29 entry#7 LOADCHARACTERPORTRAIT×1 |
| overlay-24 | `1AF7h` | 1217 | 0 | overlay-31 entry#1 SGN×4、overlay-32 entry#11 ONVISSCREEN×3 |
| overlay-24 | `0C28h` | 1187 | 0 | overlay-25 entry#10 OLDCLASSOK×2 |

## PC98（1481 個函式）

| 分組 | 數量 |
|---|---:|
| 其他 | 938 |
| 極小函式／thunk 候選 | 246 |
| 無呼叫且無寫入（純計算候選） | 176 |
| 軟體中斷（有 int） | 98 |
| 硬體存取（有 in／out） | 19 |
| 有字串參照 | 4 |

最常被呼叫的共用 routine（依 far call 次數）：

| routine | 次數 |
|---|---:|
| overlay-23 entry#9 ROLLDICE | 134 |
| overlay-24 entry#20 CHARMSG | 81 |
| overlay-24 entry#27 SPELLON | 72 |
| overlay-32 entry#15 FINDX | 72 |
| overlay-32 entry#16 FINDY | 72 |
| overlay-07 entry#1 ADDRESSVALUE | 52 |
| overlay-24 entry#19 HORZMSG | 48 |
| overlay-23 entry#3 SPELLOFF | 46 |
| overlay-23 entry#11 ADDEFFECT | 43 |
| overlay-07 entry#2 READVAR | 41 |
| overlay-24 entry#2 SHOWALL | 35 |
| overlay-24 entry#34 QUIT | 35 |
| overlay-07 entry#15 STOREVALUE | 32 |
| overlay-26 entry#8 YESNO | 31 |
| overlay-26 entry#3 INIMNUBUF | 30 |

最大的 20 個函式（優先解讀候選）：

| 模組 | 位址 | 大小 | 被呼叫 | 主要 far call |
|---|---|---:|---:|---|
| overlay-17 | `1627h` | 3756 | 2 | overlay-19 entry#1 SHOWACTIVECHAR×2、overlay-23 entry#9 ROLLDICE×2 |
| overlay-17 | `2E9Bh` | 3160 | 1 | overlay-24 entry#7 DOCOM×3、overlay-19 entry#1 SHOWACTIVECHAR×1 |
| overlay-16 | `2A6Dh` | 3141 | 1 | overlay-23 entry#11 ADDEFFECT×10、overlay-17 entry#6 TRAINCHARACTER×1 |
| overlay-17 | `546Fh` | 2527 | 1 | overlay-24 entry#10 BYTETOSTR×2、overlay-24 entry#22 PRINTCHARNAME×1 |
| overlay-23 | `03FEh` | 2287 | 9 | — |
| PC98-GAME.EXE | `145CAh` | 2253 | 2 | — |
| overlay-30 | `078Bh` | 1977 | 0 | — |
| overlay-12 | `2ED4h` | 1840 | 0 | — |
| overlay-17 | `02CBh` | 1834 | 0 | overlay-24 entry#40 PICKAPERSON×6、overlay-16 entry#11 SAVECURRENTGAME×3 |
| overlay-02 | `222Ch` | 1812 | 1 | overlay-07 entry#15 STOREVALUE×15、overlay-07 entry#1 ADDRESSVALUE×6 |
| overlay-17 | `45D0h` | 1771 | 2 | overlay-16 entry#4 LOADACTIVEICON×4、overlay-33 entry#4 DISPOSEFIGURE×3 |
| overlay-26 | `0133h` | 1771 | 2 | overlay-27 entry#5 HIDEWILDCURSOR×3、overlay-27 entry#4 SHOWWILDCURSOR×2 |
| overlay-01 | `0321h` | 1551 | 1 | — |
| overlay-23 | `19CAh` | 1501 | 2 | overlay-24 entry#27 SPELLON×7 |
| overlay-21 | `10FCh` | 1455 | 0 | overlay-23 entry#9 ROLLDICE×14 |
| overlay-22 | `66EBh` | 1434 | 0 | — |
| overlay-13 | `329Fh` | 1353 | 2 | overlay-32 entry#13 REFRESHCOMBATMAP×3、overlay-32 entry#15 FINDX×2 |
| overlay-17 | `3BF4h` | 1326 | 5 | overlay-24 entry#19 HORZMSG×3、overlay-16 entry#12 ATTACHCHARACTER×2 |
| overlay-32 | `0CCBh` | 1321 | 4 | overlay-33 entry#6 PUTFIGURE×1 |
| overlay-22 | `52E0h` | 1312 | 0 | overlay-24 entry#36 FIGCASTERLEVEL×1、overlay-23 entry#11 ADDEFFECT×1 |
