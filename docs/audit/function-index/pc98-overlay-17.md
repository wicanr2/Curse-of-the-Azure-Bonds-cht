# pc98-overlay-17 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADGEN | 62 | 16 | 0 | 7 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call far ptr sub_15A1`、`call far ptr loc_101A`、`call far ptr sub_11C7`、`call far ptr loc_147C+1`、`call loc_17B4+3`、`call far ptr loc_1981+2`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md |
| `003E` | sub_3E | — | 215 | 79 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0115` | sub_115 | — | 107 | 37 | 1 | 5 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `02CB` | sub_2CB | DOGEN | 1834 | 676 | 0 | 19 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0AFB` | sub_AFB | — | 825 | 325 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0DE5` | sub_DE5 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0DEA` | sub_DEA | — | 14 | 6 | 3 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 14 bytes，已逐條讀完） | audit/function-triage.md |
| `0E03` | sub_E03 | — | 15 | 9 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | audit/embedded-strings.md |
| `0E17` | sub_E17 | — | 6 | 2 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `0E26` | sub_E26 | — | 61 | 29 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 61 bytes，已逐條讀完） | — |
| `0FC5` | sub_FC5 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0FCA` | sub_FCA | — | 5 | 1 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1691+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0FCF` | sub_FCF | — | 5 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `0FD4` | sub_FD4 | — | 9 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 81Fh:1A2h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | audit/embedded-strings.md |
| `0FF7` | sub_FF7 | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | audit/embedded-strings.md |
| `11C7` | sub_11C7 | — | 722 | 63 | 1 | 3 |  | 待解讀 | — | — | — |
| `1437` | sub_1437 | — | 72 | 28 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-19h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 72 bytes，已逐條讀完） | — |
| `14CA` | sub_14CA | — | 6 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_168B+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `152E` | sub_152E | — | 131 | 50 | 2 | 2 |  | 待解讀 | — | — | — |
| `15A1` | sub_15A1 | — | 130 | 60 | 2 | 3 |  | 待解讀 | — | — | — |
| `15F5` | sub_15F5 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1691+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | audit/embedded-strings.md |
| `15FF` | sub_15FF | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1604` | sub_1604 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 81Fh:1A2h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1609` | sub_1609 | — | 3 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_2788`，控制權轉交後不返回（body 共 3 bytes，已逐條讀完） | — |
| `161D` | sub_161D | — | 10 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mul dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1627` | sub_1627 | — | 3756 | 1446 | 2 | 19 |  | 待解讀 | — | — | audit/function-triage.md |
| `166A` | sub_166A | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-55h], di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1683` | sub_1683 | — | 6 | 2 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_142C+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `18C4` | sub_18C4 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-1Bh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1923` | sub_1923 | — | 82 | 32 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-55h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 82 bytes，已逐條讀完） | audit/embedded-strings.md |
| `197E` | sub_197E | — | 7 | 2 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+74h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `278C` | sub_278C | — | 135 | 52 | 2 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `28E1` | sub_28E1 | — | 509 | 213 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `2B27` | sub_2B27 | — | 131 | 60 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `2C1C` | sub_2C1C | — | 411 | 184 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `2DB7` | sub_2DB7 | — | 155 | 71 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `2E9B` | sub_2E9B | — | 3160 | 1202 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `3BBD` | sub_3BBD | ADDCHARACTERTOPARTY | 55 | 24 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:262h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 55 bytes，已逐條讀完） | — |
| `3BF4` | sub_3BF4 | — | 1326 | 517 | 5 | 12 |  | 待解讀 | — | — | audit/function-triage.md |
| `4122` | sub_4122 | REMOVECHARACTERFROMPARTY | 347 | 110 | 3 | 3 | ✓ | 待解讀 | — | — | — |
| `427D` | sub_427D | — | 61 | 27 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp+arg_4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 61 bytes，已逐條讀完） | audit/embedded-strings.md |
| `4320` | sub_4320 | — | 525 | 208 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `45CA` | sub_45CA | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `dec ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `45D0` | sub_45D0 | SETACTIVEICON | 1771 | 687 | 2 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `4CBB` | sub_4CBB | CHARCONBONUS | 249 | 92 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `4DB4` | sub_4DB4 | — | 218 | 88 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `4E8E` | sub_4E8E | — | 9 | 4 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+arg_2]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `4E97` | sub_4E97 | — | 579 | 238 | 9 | 2 |  | 待解讀 | — | — | — |
| `50DA` | sub_50DA | — | 256 | 98 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `51DA` | sub_51DA | — | 246 | 88 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `52D0` | sub_52D0 | — | 98 | 38 | 3 | 2 |  | 待解讀 | — | — | — |
| `5420` | sub_5420 | TRAINCHARACTER | 72 | 25 | 2 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_FF7`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 72 bytes，已逐條讀完） | — |
| `546F` | sub_546F | — | 2527 | 935 | 1 | 9 |  | 待解讀 | — | — | audit/function-triage.md |
| `5E57` | sub_5E57 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
