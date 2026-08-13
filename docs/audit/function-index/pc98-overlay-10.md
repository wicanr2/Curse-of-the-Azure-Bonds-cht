# pc98-overlay-10 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCOMPREP | 72 | 18 | 0 | 9 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 13 個呼叫，沒有其他動作：`call far ptr unk_101A`、`call far ptr loc_1882+1`、`call far ptr loc_1815+2`、`call loc_1922+1`、`call loc_E26`、`call far ptr loc_1694+3`（body 共 72 bytes，已逐條讀完） | spec/572-resident-service-functions.md |
| `0048` | sub_48 | — | 132 | 55 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `00CC` | sub_CC | — | 554 | 208 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `02F6` | sub_2F6 | — | 130 | 49 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0378` | sub_378 | — | 116 | 54 | 5 | 1 | ✓ | 待解讀 | — | — | — |
| `03EC` | sub_3EC | — | 165 | 78 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0491` | sub_491 | — | 125 | 73 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `050E` | sub_50E | — | 472 | 196 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `06E6` | sub_6E6 | — | 471 | 185 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `08BD` | sub_8BD | — | 196 | 89 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `0981` | sub_981 | — | 27 | 12 | 3 | 0 | ✓ | 待解讀 | — | — | — |
| `099C` | sub_99C | — | 86 | 36 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `09F2` | sub_9F2 | — | 268 | 115 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0AFE` | sub_AFE | — | 326 | 138 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0C4A` | sub_C4A | — | 75 | 31 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 75 bytes，已逐條讀完） | — |
| `0C95` | sub_C95 | — | 424 | 182 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0E3D` | sub_E3D | — | 382 | 183 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0FBB` | sub_FBB | — | 80 | 30 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1023` | sub_1023 | — | 174 | 72 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `10D1` | sub_10D1 | — | 286 | 94 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `11EF` | sub_11EF | — | 49 | 18 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1220` | sub_1220 | — | 324 | 138 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1364` | sub_1364 | — | 902 | 365 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `142D` | sub_142D | — | 81 | 36 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-16h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `14E3` | sub_14E3 | — | 38 | 16 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1587`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 38 bytes，已逐條讀完） | — |
| `167E` | sub_167E | — | 316 | 129 | 2 | 5 |  | 待解讀 | — | — | — |
| `17B7` | sub_17B7 | — | 55 | 21 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 55 bytes，已逐條讀完） | — |
| `17EE` | sub_17EE | — | 6 | 3 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 10h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `17F4` | sub_17F4 | — | 5 | 1 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_155A+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `17F9` | sub_17F9 | — | 8 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr [bp-0Eh], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `18FB` | sub_18FB | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `1900` | sub_1900 | — | 756 | 277 | 1 | 3 |  | 待解讀 | — | — | — |
| `196A` | sub_196A | — | 26 | 11 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:7AC8h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | — |
| `1C28` | sub_1C28 | INITCOMBAT | 239 | 86 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `1DAA` | sub_1DAA | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
