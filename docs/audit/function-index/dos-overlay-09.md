# dos-overlay-09 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 62 | 16 | 0 | 8 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call loc_1046+4`、`call loc_18B2+1`、`call far ptr loc_1951+2`、`call far ptr sub_1847`、`call far ptr loc_16BC+1`、`call far ptr sub_15D1`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/583-ledger-denominator-repair.md |
| `004D` | sub_4D | — | 512 | 186 | 0 | 14 | ✓ | 待解讀 | — | — | spec/687-far-call-flattening-and-stack-leftover.md |
| `024D` | sub_24D | — | 100 | 33 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02B1` | sub_2B1 | — | 256 | 96 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `03B1` | sub_3B1 | — | 249 | 97 | 2 | 5 | ✓ | 待解讀 | — | — | spec/687-far-call-flattening-and-stack-leftover.md |
| `04AA` | sub_4AA | — | 347 | 122 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0605` | sub_605 | — | 304 | 111 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0735` | sub_735 | — | 630 | 246 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `09C7` | sub_9C7 | — | 671 | 242 | 2 | 8 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-09.md |
| `0C53` | sub_C53 | — | 5 | 2 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `idiv cx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0C58` | sub_C58 | — | 10 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `0C62` | sub_C62 | — | 9 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr ds:47F1h, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `0C7B` | sub_C7B | — | 41 | 16 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 41 bytes，已逐條讀完） | — |
| `0CB7` | sub_CB7 | — | 228 | 75 | 3 | 7 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 09C7h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0DB1` | sub_DB1 | — | 1043 | 349 | 1 | 15 | ✓ | 待解讀 | — | — | — |
| `11CF` | sub_11CF | — | 38 | 15 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 38 bytes，已逐條讀完） | — |
| `1201` | sub_1201 | — | 95 | 31 | 5 | 4 | ✓ | 待解讀 | — | — | — |
| `1273` | sub_1273 | — | 248 | 84 | 2 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1388` | sub_1388 | — | 197 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1444` | sub_1444 | — | 7 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:6FA2h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1458` | sub_1458 | — | 3 | 1 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_1539`，控制權轉交後不返回（body 共 3 bytes，已逐條讀完） | — |
| `150E` | sub_150E | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 13 | 5 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+2Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-09.md |
| `154F` | sub_154F | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cl, 4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1577` | sub_1577 | — | 15 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jle short sub_159F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `1586` | sub_1586 | — | 11 | 4 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cx, 3`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1595` | sub_1595 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor ah, ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 5 | 2 | 7 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-2], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159F` | sub_159F | — | 32 | 12 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+2Eh], 55h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 32 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 10 | 3 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `15DB` | sub_15DB | — | 126 | 44 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1542h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1681` | sub_1681 | — | 1129 | 371 | 2 | 6 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-09.md |
| `16A4` | sub_16A4 | — | 26 | 10 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub ax, dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | — |
| `1847` | sub_1847 | — | 98 | 35 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1681h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18A9` | sub_18A9 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-0Eh], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1921` | sub_1921 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and al, 10h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1926` | sub_1926 | — | 664 | 206 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1681h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1BC6` | sub_1BC6 | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
