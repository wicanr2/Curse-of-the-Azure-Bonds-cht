# pc98-overlay-09 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCOMPTACT | 62 | 16 | 0 | 6 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call loc_1018+2`、`call far ptr loc_187F+4`、`call sub_1923`、`call far ptr loc_1817`、`call far ptr loc_1695+2`、`call far ptr sub_15A1`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/583-ledger-denominator-repair.md |
| `005D` | sub_5D | COMPUTERCONTROL | 530 | 191 | 0 | 14 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `026F` | sub_26F | — | 100 | 33 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `02D3` | sub_2D3 | — | 256 | 96 | 1 | 3 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>spec/493-pc98-quick-sleep-area-target.md |
| `03D3` | sub_3D3 | — | 249 | 97 | 2 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/493-pc98-quick-sleep-area-target.md<br>spec/503-pc98-quick-target-object-chain-boundary.md |
| `04CC` | sub_4CC | — | 347 | 122 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/493-pc98-quick-sleep-area-target.md<br>spec/503-pc98-quick-target-object-chain-boundary.md |
| `0627` | sub_627 | — | 304 | 111 | 1 | 3 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>spec/493-pc98-quick-sleep-area-target.md |
| `0757` | sub_757 | — | 630 | 246 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `09E9` | sub_9E9 | — | 494 | 178 | 2 | 6 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-09.md |
| `0BD7` | sub_BD7 | — | 32 | 11 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mul dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 32 bytes，已逐條讀完） | — |
| `0C22` | sub_C22 | — | 8 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [di+6AAh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `0C4A` | sub_C4A | — | 393 | 135 | 2 | 7 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 09E9h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0DD3` | sub_DD3 | — | 560 | 180 | 1 | 9 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-09.md |
| `0FE8` | sub_FE8 | — | 297 | 102 | 2 | 9 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0DD3h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `119F` | sub_119F | — | 14 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1592`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 14 bytes，已逐條讀完） | — |
| `1223` | sub_1223 | — | 95 | 31 | 4 | 5 | ✓ | 待解讀 | — | — | — |
| `1296` | sub_1296 | — | 239 | 80 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `13A5` | sub_13A5 | — | 111 | 36 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1414` | sub_1414 | — | 8 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+0F7h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `14D9` | sub_14D9 | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `idiv cx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `14F2` | sub_14F2 | — | 45 | 19 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_140F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 45 bytes，已逐條讀完） | — |
| `151F` | sub_151F | — | 2 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short near ptr sub_1556`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `1524` | sub_1524 | — | 5 | 1 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+13h], 5`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1529` | sub_1529 | — | 31 | 13 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_143A+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1556` | sub_1556 | — | 9 | 4 | 9 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 9 bytes，已逐條讀完） | — |
| `155F` | sub_155F | — | 1 | 1 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | audit/function-index/pc98-overlay-09.md |
| `1560` | sub_1560 | — | 5 | 2 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 12h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1565` | sub_1565 | — | 7 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+56h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `158D` | sub_158D | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1592` | sub_1592 | — | 7 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-2], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `15A1` | sub_15A1 | — | 10 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `15AB` | sub_15AB | — | 126 | 45 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 155Fh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `169E` | sub_169E | — | 606 | 217 | 2 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-09.md |
| `1900` | sub_1900 | — | 35 | 12 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 35 bytes，已逐條讀完） | — |
| `1923` | sub_1923 | — | 689 | 215 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 169Eh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/embedded-strings.md |
| `1BE3` | sub_1BE3 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
