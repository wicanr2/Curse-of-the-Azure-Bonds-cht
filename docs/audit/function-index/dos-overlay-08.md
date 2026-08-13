# dos-overlay-08 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 77 | 19 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 14 個呼叫，沒有其他動作：`call far ptr sub_104A`、`call far ptr 167h:4Dh`、`call far ptr 187h:43h`、`call far ptr 18Ch:93h`、`call far ptr 196h:43h`、`call far ptr 14Dh:101h`（body 共 77 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `004D` | sub_4D | — | 166 | 62 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-08.md |
| `00F3` | sub_F3 | — | 184 | 67 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `01AB` | sub_1AB | — | 192 | 68 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `026B` | sub_26B | — | 325 | 102 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `03D5` | sub_3D5 | — | 810 | 302 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `076C` | sub_76C | — | 489 | 192 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0885` | sub_885 | — | 9 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+111h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `0997` | sub_997 | — | 317 | 112 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0B06` | sub_B06 | — | 415 | 152 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0C4E` | sub_C4E | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0C6C` | sub_C6C | — | 11 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 18Ch:61h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `0C94` | sub_C94 | — | 2 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_CE0`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `0CDA` | sub_CDA | — | 472 | 174 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0B06h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0EE9` | sub_EE9 | — | 184 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0FE8` | sub_FE8 | — | 83 | 33 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 83 bytes，已逐條讀完） | — |
| `1009` | sub_1009 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `104A` | sub_104A | — | 98 | 41 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FE8h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `11DF` | sub_11DF | — | 309 | 133 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1314` | sub_1314 | — | 81 | 24 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1365` | sub_1365 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
