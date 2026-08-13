# pc98-overlay-04 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTEMPLE | 52 | 14 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 9 個呼叫，沒有其他動作：`call far ptr loc_1153+1`、`call far ptr sub_101A`、`call far ptr 164h:57h`、`call far ptr 14Ah:101h`、`call loc_6DC+4`、`call far ptr 19Ah:2Ah`（body 共 52 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0051` | sub_51 | — | 99 | 42 | 6 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0103` | sub_103 | — | 368 | 139 | 7 | 3 | ✓ | 待解讀 | — | — | — |
| `02A3` | sub_2A3 | — | 134 | 53 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `035B` | sub_35B | — | 217 | 75 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `048E` | sub_48E | — | 483 | 192 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `06A4` | sub_6A4 | — | 4 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and word ptr [bp+si-7Dh], 68h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 4 bytes，已逐條讀完） | — |
| `06A8` | sub_6A8 | — | 610 | 192 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `093E` | sub_93E | — | 205 | 77 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0A33` | sub_A33 | — | 202 | 72 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B31` | sub_B31 | — | 141 | 46 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BFF` | sub_BFF | — | 561 | 240 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `0F42` | sub_F42 | GOTEMPLE | 217 | 86 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-04.md |
| `0FD9` | sub_FD9 | — | 25 | 11 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:262h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | — |
| `101A` | sub_101A | — | 306 | 128 | 2 | 10 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0F42h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1109` | sub_1109 | — | 10 | 7 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1113` | sub_1113 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1118` | sub_1118 | — | 30 | 14 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `1136` | sub_1136 | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:62Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1140` | sub_1140 | — | 6 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `114F` | sub_114F | — | 9 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_6A4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `11BD` | sub_11BD | — | 13 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 14Ah:2Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | — |
| `11D7` | sub_11D7 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
