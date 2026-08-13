# dos-overlay-15 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `01CB` | sub_1CB | — | 115 | 42 | 3 | 2 |  | 待解讀 | — | — | — |
| `023E` | sub_23E | — | 68 | 27 | 1 | 2 |  | 待解讀 | — | — | — |
| `03A9` | nullsub_1 | — | 3 | 1 | 0 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 3 bytes，已逐條讀完） | — |
| `04CB` | sub_4CB | — | 182 | 77 | 1 | 6 |  | 待解讀 | — | — | — |
| `085C` | sub_85C | — | 155 | 57 | 1 | 1 |  | 待解讀 | — | — | — |
| `0942` | sub_942 | — | 439 | 180 | 1 | 6 |  | 待解讀 | — | — | — |
| `0E69` | sub_E69 | — | 63 | 25 | 1 | 0 |  | 待解讀 | — | — | — |
| `1016` | sub_1016 | — | 12 | 6 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `1022` | sub_1022 | — | 12 | 4 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les ax, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `11DE` | sub_11DE | — | 1096 | 472 | 3 | 7 |  | 待解讀 | — | — | — |
| `14D2` | sub_14D2 | — | 86 | 41 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1522` | sub_1522 | — | 92 | 38 | 4 | 6 |  | 待解讀 | — | — | — |
| `1554` | sub_1554 | — | 2 | 1 | 7 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1581`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `158A` | sub_158A | — | 31 | 12 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov dx, es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `15A9` | sub_15A9 | — | 219 | 68 | 7 | 2 |  | 待解讀 | — | — | — |
| `15D1` | sub_15D1 | — | 35 | 11 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-8], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 35 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:650Ah, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16AE` | sub_16AE | — | 6 | 2 | 6 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_16C7`，控制權轉交後不返回；先設定 `mov ds:650Ch, dx`（body 共 6 bytes，已逐條讀完） | — |
| `16B8` | sub_16B8 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 14 | 5 | 2 | 0 |  | 待解讀 | — | — | — |
| `16CB` | sub_16CB | — | 255 | 82 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1834` | sub_1834 | — | 250 | 103 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `199C` | sub_199C | — | 330 | 136 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `1B4D` | sub_1B4D | — | 348 | 146 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1D0A` | sub_1D0A | — | 459 | 198 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `1EDA` | sub_1EDA | — | 35 | 14 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1EFD` | sub_1EFD | — | 188 | 74 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1FB9` | sub_1FB9 | — | 178 | 70 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `206B` | sub_206B | — | 90 | 33 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `20C5` | sub_20C5 | — | 400 | 143 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2255` | sub_2255 | — | 180 | 63 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2309` | sub_2309 | — | 164 | 68 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `23FC` | sub_23FC | — | 485 | 206 | 0 | 11 | ✓ | 待解讀 | — | — | — |
| `25E1` | sub_25E1 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
