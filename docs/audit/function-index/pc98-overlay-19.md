# pc98-overlay-19 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADLIBRARY | 57 | 15 | 0 | 8 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call far ptr loc_1695+2`、`call far ptr sub_17B7`、`call far ptr loc_11C4+3`、`call sub_147D`、`call far ptr sub_1154`、`call far ptr loc_1083+1`（body 共 57 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-12.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md |
| `0098` | sub_98 | SHOWACTIVECHAR | 1305 | 537 | 4 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `05B1` | sub_5B1 | SHOWLOOT | 189 | 82 | 4 | 2 | ✓ | 待解讀 | — | — | — |
| `069E` | sub_69E | SHOWACTIVECOMBATSTUFF | 784 | 332 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `09BC` | sub_9BC | DISPLAYSTAT | 341 | 135 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0B82` | sub_B82 | VIEWACTIVECHAR | 673 | 262 | 0 | 13 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0E8E` | sub_E8E | — | 247 | 101 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1074` | sub_1074 | VIEWITEM | 144 | 70 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1154` | sub_1154 | — | 65 | 33 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 65 bytes，已逐條讀完） | — |
| `1195` | sub_1195 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `119A` | sub_119A | — | 15 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 418h:0D17h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `11A9` | sub_11A9 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `11AE` | sub_11AE | — | 6 | 4 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `142D` | sub_142D | — | 10 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr unk_14F2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1437` | sub_1437 | — | 48 | 21 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr unk_14F2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 48 bytes，已逐條讀完） | — |
| `146E` | sub_146E | — | 10 | 7 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1478` | sub_1478 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `147D` | sub_147D | — | 63 | 26 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 63 bytes，已逐條讀完） | — |
| `14E3` | sub_14E3 | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and [bp-7D99h], cl`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `14E8` | sub_14E8 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and ds:9320h, al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `150B` | sub_150B | — | 3 | 1 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp near ptr loc_1B2D+1`，控制權轉交後不返回（body 共 3 bytes，已逐條讀完） | — |
| `1515` | sub_1515 | — | 8 | 2 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and word ptr [di-7Dh], 0FF80h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `156A` | sub_156A | — | 14 | 5 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `test ax, 4881h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 14 bytes，已逐條讀完） | — |
| `1578` | sub_1578 | ITEMS | 1 | 1 | 2 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | audit/function-index/pc98-overlay-19.md |
| `1579` | sub_1579 | — | 10 | 3 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les ax, ds:9594h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1583` | sub_1583 | — | 5 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-3Fh], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1588` | sub_1588 | — | 25 | 8 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor ax, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | audit/function-index/dos-overlay-19.md |
| `15A1` | sub_15A1 | — | 125 | 42 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1578h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1604` | sub_1604 | — | 15 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr ds:0A335h, 52h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `1613` | sub_1613 | — | 9 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+197h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `1622` | sub_1622 | — | 6 | 1 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp word ptr es:[di+1CAh], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1683` | sub_1683 | — | 9 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+196h], 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `174A` | sub_174A | — | 109 | 41 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1578h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `17B7` | sub_17B7 | — | 617 | 261 | 2 | 13 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1578h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18C4` | sub_18C4 | — | 7 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:62Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `19CA` | sub_19CA | — | 23 | 11 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, 59h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | audit/function-triage.md |
| `1A8F` | sub_1A8F | — | 1055 | 342 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `1EAE` | sub_1EAE | READYITEM | 423 | 148 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2055` | sub_2055 | — | 64 | 28 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-7]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 64 bytes，已逐條讀完） | — |
| `2095` | sub_2095 | — | 79 | 29 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 79 bytes，已逐條讀完） | — |
| `2102` | sub_2102 | — | 166 | 61 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `21C5` | sub_21C5 | — | 176 | 66 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `2275` | sub_2275 | — | 465 | 154 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `246F` | sub_246F | USEITEM | 786 | 280 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `27DC` | sub_27DC | — | 482 | 183 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `2A32` | sub_2A32 | — | 482 | 199 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `2C69` | sub_2C69 | — | 828 | 321 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `2FE5` | sub_2FE5 | — | 722 | 283 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `32B7` | sub_32B7 | OVERLOADED | 132 | 45 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `333B` | sub_333B | ACTIVECHAR | 207 | 71 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `340A` | sub_340A | CASH | 117 | 44 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `34E4` | sub_34E4 | PICKSPELL | 372 | 157 | 2 | 6 | ✓ | 待解讀 | — | — | — |
| `3658` | sub_3658 | CANDOHEAL | 101 | 37 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `36BD` | sub_36BD | CANDOCURE | 89 | 30 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `37AB` | sub_37AB | DOHEAL | 373 | 156 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `39CA` | sub_39CA | DOCURE | 525 | 210 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `3BD7` | sub_3BD7 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
