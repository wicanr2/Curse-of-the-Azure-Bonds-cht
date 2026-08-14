# pc98-overlay-08 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCOMBAT | 77 | 19 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 14 個呼叫，沒有其他動作：`call far ptr unk_101A`、`call far ptr 164h:57h`、`call far ptr 184h:43h`、`call far ptr 189h:93h`、`call far ptr 194h:43h`、`call far ptr 14Ah:101h`（body 共 77 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/748-cloud-effect-dispel-pair.md<br>spec/751-overlay-init-chain-dependency-graph.md |
| `004D` | sub_4D | — | 246 | 86 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-08.md<br>spec/748-cloud-effect-dispel-pair.md<br>spec/749-combat-teardown-and-battlefield-grid.md |
| `0143` | sub_143 | GOCOMBAT | 184 | 67 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `01FB` | sub_1FB | — | 192 | 68 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `02BB` | sub_2BB | — | 325 | 102 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0414` | sub_414 | — | 776 | 288 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `0795` | sub_795 | — | 97 | 38 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `09A5` | sub_9A5 | — | 320 | 112 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0B20` | sub_B20 | — | 167 | 52 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-08.md |
| `0BA5` | sub_BA5 | — | 20 | 10 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 18h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 20 bytes，已逐條讀完） | — |
| `0BB9` | sub_BB9 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 19Eh:6B5h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `0BC3` | sub_BC3 | — | 6 | 2 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-20Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `0BFF` | sub_BFF | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:6BCh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0C04` | sub_C04 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0C09` | sub_C09 | — | 20 | 11 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 20 bytes，已逐條讀完） | — |
| `0C1D` | sub_C1D | — | 590 | 221 | 2 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0B20h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0C45` | sub_C45 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `0F17` | sub_F17 | — | 184 | 67 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0FD9` | sub_FD9 | — | 5 | 2 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1028` | sub_1028 | — | 82 | 32 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 82 bytes，已逐條讀完） | audit/function-index/pc98-overlay-08.md |
| `107A` | sub_107A | — | 14 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr ds:0A339h, 44h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 14 bytes，已逐條讀完） | — |
| `11C7` | sub_11C7 | — | 263 | 105 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1028h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `124D` | sub_124D | — | 296 | 123 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1375` | sub_1375 | — | 81 | 24 | 1 | 1 | ✓ | 待解讀 | — | — | — |
