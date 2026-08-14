# pc98-overlay-13 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | INITDUDE | 292 | 94 | 0 | 7 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-24.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/754-small-predicates-and-wrappers.md |
| `0124` | sub_124 | FIGMOVE | 110 | 36 | 3 | 2 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>加成後夾制再乘 2(retf 4)：+1A6h/+198h/DS:7F09h/DS:0A034h/DS:0A030h，其餘同 DOS overlay-13:0124h | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-13.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `0192` | sub_192 | CALCWEAPDAMAGE | 196 | 74 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-24.md |
| `0358` | sub_358 | DESCRIBEWEAPONATTACK | 812 | 326 | 3 | 4 | ✓ | 待解讀 | — | — | — |
| `0684` | sub_684 | — | 266 | 98 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `078E` | sub_78E | REALMOVE | 523 | 193 | 0 | 5 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-24.md |
| `0999` | sub_999 | CHECKPARTINGBLOWS | 968 | 345 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `0D81` | sub_D81 | RUNAWAY | 189 | 74 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `0E3E` | sub_E3E | FIGBLOWS | 313 | 105 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0F77` | sub_F77 | ADJUSTBLOWS | 45 | 19 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>依相位進位的折半(retf 2)：f(x) = (x + (DS:0A81Fh and 1)) div 2，與 DOS overlay-13:0F12h 同 | audit/embedded-strings.md<br>spec/754-small-predicates-and-wrappers.md |
| `0FB1` | sub_FB1 | TRYTOSWEEP | 108 | 35 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-13.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `101A` | sub_101A | — | 396 | 146 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FB1h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `119F` | sub_119F | — | 3 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_1116`，控制權轉交後不返回（body 共 3 bytes，已逐條讀完） | — |
| `11A9` | sub_11A9 | — | 6 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 6 bytes，已逐條讀完） | audit/embedded-strings.md |
| `11AF` | sub_11AF | CHECKTARGET | 33 | 13 | 4 | 1 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-13.md |
| `11C7` | sub_11C7 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp dx, [bp+8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `11CC` | sub_11CC | — | 140 | 46 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 11AFh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `12A8` | sub_12A8 | TURNUNDEAD | 487 | 180 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `1428` | sub_1428 | — | 8 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1522+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `14B4` | sub_14B4 | ANYUNDEAD | 61 | 22 | 1 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_14F4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 61 bytes，已逐條讀完） | — |
| `14F2` | sub_14F2 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-3]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `14F7` | sub_14F7 | — | 25 | 10 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-8], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | — |
| `1510` | sub_1510 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1538` | sub_1538 | — | 8 | 3 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-3]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | audit/embedded-strings.md |
| `153D` | sub_153D | — | 5 | 1 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+0E9h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les ax, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | audit/embedded-strings.md |
| `155B` | sub_155B | — | 5 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jnz short near ptr unk_14F1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1560` | sub_1560 | — | 5 | 2 | 9 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov sp, bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1565` | sub_1565 | — | 4 | 2 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 4 bytes，已逐條讀完） | — |
| `1569` | sub_1569 | — | 1 | 1 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-13.md |
| `156A` | sub_156A | — | 11 | 4 | 10 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp+var_13], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1622` | sub_1622 | — | 7 | 4 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call near ptr sub_358`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 19 | 6 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor ah, ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 19 bytes，已逐條讀完） | — |
| `167E` | sub_167E | — | 11 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jbe short loc_16E8`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1697` | sub_1697 | — | 485 | 165 | 2 | 7 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1569h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1883` | sub_1883 | — | 81 | 29 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+197h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `18DD` | sub_18DD | — | 5 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+14h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18E2` | sub_18E2 | — | 8 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr es:[di+19Ch], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `18F1` | sub_18F1 | — | 5 | 2 | 7 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_17DF`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18F6` | sub_18F6 | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `or ax, [bp+8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `191E` | sub_191E | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+14h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1923` | sub_1923 | — | 174 | 58 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1569h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/embedded-strings.md |
| `19CB` | sub_19CB | NEWATTACKER | 142 | 53 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `1A59` | sub_1A59 | ATTACKE | 707 | 242 | 6 | 9 | ✓ | 待解讀 | — | — | — |
| `1D1C` | sub_1D1C | CALCRANGEMODS | 187 | 76 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1DD7` | sub_1DD7 | LOADCOMSTUFF | 57 | 15 | 0 | 7 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call far ptr sub_101A`、`call far ptr sub_1883`、`call sub_1923`、`call loc_1982+1`、`call far ptr loc_15A0+1`、`call far ptr sub_11C7`（body 共 57 bytes，已逐條讀完） | audit/overlay-init-graph.md |
| `1E30` | sub_1E30 | COMPTARGCURE | 537 | 195 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2049` | sub_2049 | — | 482 | 169 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `225F` | sub_225F | WHOZAP | 1300 | 483 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `27A1` | sub_27A1 | CASTCOMBATSPELL | 372 | 135 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `2915` | sub_2915 | THIEFATTACK | 223 | 73 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `29F4` | sub_29F4 | FIGDIR | 126 | 43 | 6 | 3 | ✓ | 待解讀 | — | — | — |
| `2A72` | sub_2A72 | SHOWARROW | 546 | 223 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `2C94` | sub_2C94 | SETMORALE | 138 | 49 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2D1E` | sub_2D1E | FASTESTENEMY | 130 | 46 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2DB9` | sub_2DB9 | CHECKBETRAYAL | 205 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2EBB` | sub_2EBB | — | 541 | 217 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `30D8` | sub_30D8 | — | 361 | 132 | 2 | 9 | ✓ | 待解讀 | — | — | — |
| `329F` | sub_329F | — | 1353 | 521 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `37E8` | sub_37E8 | — | 161 | 60 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3889` | sub_3889 | — | 427 | 142 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `3A74` | sub_3A74 | VIEWDUDES | 779 | 303 | 1 | 11 | ✓ | 待解讀 | — | — | — |
| `3D7F` | sub_3D7F | PICKTARGET | 480 | 159 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `3F71` | sub_3F71 | EFF57 | 327 | 128 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `40B8` | sub_40B8 | — | 114 | 42 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/760-item-effect-flag-clear-and-file-exists.md<br>取兩個戰鬥員的 x/y 後畫(retf 4)：結構與 DOS overlay-13:4162h(spec 759)逐條相同，但最後一個參數 DOS 是常數 1Eh、PC-98 是 DS:7F16h * 7(有號乘法)。真實平台差異，非匯出誤差 | spec/760-item-effect-flag-clear-and-file-exists.md |
| `412A` | sub_412A | — | 186 | 63 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `426B` | sub_426B | EFF87 | 722 | 312 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `453D` | sub_453D | EFF139 | 308 | 111 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `4671` | sub_4671 | EFF144 | 264 | 93 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `478B` | sub_478B | EFF96 | 235 | 75 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4876` | sub_4876 | KILLTHEBASTARDS | 194 | 67 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4938` | sub_4938 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
