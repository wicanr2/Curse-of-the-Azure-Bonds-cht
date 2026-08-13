# dos-overlay-17 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0782` | sub_782 | — | 2622 | 999 | 1 | 7 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0E93` | sub_E93 | — | 62 | 15 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1048`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 62 bytes，已逐條讀完） | — |
| `0FF5` | sub_FF5 | — | 9 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr es:[di+10Eh], 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `11F7` | sub_11F7 | — | 652 | 246 | 2 | 6 |  | 待解讀 | — | — | — |
| `1467` | sub_1467 | — | 280 | 107 | 2 | 4 |  | 待解讀 | — | — | — |
| `14FF` | sub_14FF | — | 47 | 15 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short near ptr loc_15D2+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 47 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1522` | sub_1522 | — | 48 | 19 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [di+4060h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 48 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 81 | 28 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short sub_1625`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `1625` | sub_1625 | — | 5 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 6`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `162A` | sub_162A | — | 6 | 2 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_145B+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1639` | sub_1639 | — | 88 | 39 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 88 bytes，已逐條讀完） | — |
| `164D` | sub_164D | — | 9 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `169F` | sub_169F | — | 10 | 4 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16AE` | sub_16AE | — | 6 | 2 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+11h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `17E7` | sub_17E7 | — | 413 | 159 | 2 | 2 |  | 待解讀 | — | — | — |
| `18F4` | sub_18F4 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+11h], dl`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1994` | sub_1994 | — | 10 | 4 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mul dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `199E` | sub_199E | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [di+4124h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1ECD` | sub_1ECD | — | 1608 | 600 | 3 | 12 |  | 待解讀 | — | — | audit/function-triage.md |
| `2515` | sub_2515 | — | 146 | 53 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `260C` | sub_260C | — | 278 | 111 | 2 | 6 | ✓ | 待解讀 | — | — | — |
| `2724` | sub_2724 | — | 285 | 121 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2868` | sub_2868 | — | 3454 | 1290 | 1 | 9 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `3680` | sub_3680 | — | 982 | 378 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `3A56` | sub_3A56 | — | 323 | 104 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `3B99` | sub_3B99 | — | 223 | 101 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `3C78` | sub_3C78 | — | 15 | 7 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cl, 3`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `3EE3` | sub_3EE3 | — | 1798 | 708 | 1 | 8 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `45E9` | sub_45E9 | — | 259 | 93 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `46EC` | sub_46EC | — | 218 | 88 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `47C6` | sub_47C6 | — | 588 | 242 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `4A12` | sub_4A12 | — | 256 | 98 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `4B12` | sub_4B12 | — | 344 | 126 | 1 | 1 | ✓ | 待解讀 | — | — | — |
