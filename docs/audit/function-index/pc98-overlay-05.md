# pc98-overlay-05 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPOSTCOM | 57 | 15 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0039` | sub_39 | — | 792 | 278 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0351` | sub_351 | — | 478 | 165 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `052F` | sub_52F | — | 1121 | 341 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0A5C` | sub_A5C | — | 632 | 263 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0CE9` | sub_CE9 | — | 227 | 100 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0DCC` | sub_DCC | — | 275 | 102 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0F0F` | sub_F0F | — | 265 | 107 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `1072` | sub_1072 | — | 168 | 65 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `11A9` | sub_11A9 | — | 32 | 14 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, 0A35Bh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 32 bytes，已逐條讀完） | — |
| `13C7` | sub_13C7 | — | 158 | 51 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `14C5` | sub_14C5 | — | 8 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub byte ptr [bx+4F91h], 82h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1506` | sub_1506 | — | 417 | 162 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1679` | sub_1679 | — | 10 | 4 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:6BCh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1683` | sub_1683 | — | 7 | 2 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_16C5`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1775` | sub_1775 | DOPOSTCOMBAT | 558 | 192 | 0 | 6 | ✓ | 待解讀 | — | — | spec/558-pc98-ecl-treasure-combat-boundary.md |
| `19A3` | sub_19A3 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
