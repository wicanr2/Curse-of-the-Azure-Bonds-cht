# dos-overlay-05 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 57 | 15 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call far ptr unk_104A`、`call loc_770`、`call far ptr sub_1184`、`call far ptr loc_16BC+1`、`call loc_69D+2`、`call far ptr loc_15CF+2`（body 共 57 bytes，已逐條讀完） | — |
| `0039` | sub_39 | — | 792 | 278 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0351` | sub_351 | — | 491 | 166 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `053C` | sub_53C | — | 1128 | 343 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0A7E` | sub_A7E | — | 603 | 250 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0CE6` | sub_CE6 | — | 203 | 90 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0DB1` | sub_DB1 | — | 270 | 101 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `0ED7` | sub_ED7 | — | 79 | 31 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `106C` | sub_106C | — | 250 | 86 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1166` | sub_1166 | — | 30 | 12 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, 0FFh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `1184` | sub_1184 | — | 87 | 39 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-106h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 87 bytes，已逐條讀完） | — |
| `1370` | sub_1370 | — | 246 | 80 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `14AC` | sub_14AC | — | 242 | 87 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `15A9` | sub_15A9 | — | 112 | 45 | 5 | 2 |  | 待解讀 | — | — | — |
| `169F` | sub_169F | — | 9 | 3 | 4 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_15B3`，控制權轉交後不返回；先設定 `mov [bp-4], ax`、`mov [bp-2], dx`（body 共 9 bytes，已逐條讀完） | — |
| `1736` | sub_1736 | — | 451 | 161 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `18F9` | sub_18F9 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
