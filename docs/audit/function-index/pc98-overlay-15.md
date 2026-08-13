# pc98-overlay-15 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCAMP | 52 | 14 | 0 | 7 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 9 個呼叫，沒有其他動作：`call loc_19C8+2`、`call sub_11C7`、`call far ptr sub_1084`、`call far ptr loc_159F+2`、`call far ptr loc_E22+4`、`call sub_1697`（body 共 52 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md<br>spec/583-ledger-denominator-repair.md |
| `0034` | sub_34 | — | 340 | 121 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0188` | sub_188 | — | 67 | 24 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `01CB` | sub_1CB | — | 115 | 42 | 3 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `023E` | sub_23E | — | 68 | 27 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0282` | sub_282 | — | 165 | 63 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0386` | sub_386 | — | 285 | 114 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `04C2` | sub_4C2 | DOCASTSPELL | 182 | 77 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `05CC` | sub_5CC | — | 488 | 207 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `07B4` | sub_7B4 | — | 170 | 66 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `085E` | sub_85E | — | 155 | 57 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `092D` | sub_92D | — | 444 | 181 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `0B86` | sub_B86 | — | 673 | 255 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `0E54` | sub_E54 | — | 131 | 49 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `0FC7` | sub_FC7 | — | 21 | 9 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 81Fh:43h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 21 bytes，已逐條讀完） | — |
| `0FF2` | sub_FF2 | — | 10 | 6 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, 50h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | audit/embedded-strings.md |
| `0FFC` | sub_FFC | — | 31 | 10 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-0Eh], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1084` | sub_1084 | — | 306 | 134 | 2 | 3 |  | 待解讀 | — | — | audit/embedded-strings.md<br>audit/string-pairs.md |
| `11A9` | sub_11A9 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `11AE` | sub_11AE | — | 12 | 4 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_133F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `11C7` | sub_11C7 | — | 765 | 347 | 2 | 6 |  | 待解讀 | — | — | — |
| `147B` | sub_147B | — | 85 | 37 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `14CA` | sub_14CA | — | 119 | 53 | 8 | 9 |  | 待解讀 | — | — | — |
| `1524` | sub_1524 | — | 2 | 1 | 9 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_153D`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `1558` | sub_1558 | — | 35 | 13 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp dx, ds:9596h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 35 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 7 | 2 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:959Ah, dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1688` | sub_1688 | — | 13 | 3 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+18Ch], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | — |
| `1697` | sub_1697 | — | 2 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 2 bytes，已逐條讀完） | — |
| `1699` | sub_1699 | — | 251 | 79 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1794` | sub_1794 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+18Ah], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1799` | sub_1799 | — | 9 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 9 bytes，已逐條讀完） | — |
| `17F6` | sub_17F6 | — | 296 | 124 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `19A4` | sub_19A4 | — | 404 | 160 | 1 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1B94` | sub_1B94 | — | 184 | 71 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1D4C` | sub_1D4C | — | 506 | 204 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `1F46` | sub_1F46 | — | 35 | 14 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1F69` | sub_1F69 | — | 183 | 73 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2020` | sub_2020 | — | 117 | 46 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2095` | sub_2095 | — | 61 | 24 | 3 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 61 bytes，已逐條讀完） | — |
| `20D2` | sub_20D2 | — | 90 | 33 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `212C` | sub_212C | — | 400 | 143 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `22BC` | sub_22BC | — | 180 | 63 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `2370` | sub_2370 | — | 164 | 68 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `2457` | sub_2457 | DOCAMP | 481 | 198 | 0 | 15 | ✓ | 待解讀 | — | — | — |
| `2638` | sub_2638 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
