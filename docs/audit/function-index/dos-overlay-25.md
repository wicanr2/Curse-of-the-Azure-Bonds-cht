# dos-overlay-25 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 2 個相依 unit 的 0000h — overlay-34、overlay-26。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-25.md |
| `0011` | sub_11 | — | 928 | 101 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short near ptr dword_3C+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 928 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/pc98-overlay-12.md |
| `03B1` | sub_3B1 | — | 704 | 169 | 1 | 4 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `iret`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 704 bytes，已逐條讀完） | — |
| `0671` | sub_671 | — | 370 | 103 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `07E3` | sub_7E3 | — | 757 | 187 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0AD8` | sub_AD8 | — | 426 | 151 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0C51` | nullsub_1 | — | 1 | 1 | 0 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 1 bytes，已逐條讀完） | — |
| `0D2F` | sub_D2F | — | 434 | 148 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0EE1` | sub_EE1 | — | 945 | 360 | 0 | 8 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1292` | sub_1292 | — | 92 | 33 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `12EE` | sub_12EE | — | 92 | 33 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `134A` | sub_134A | — | 80 | 29 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `139A` | sub_139A | — | 34 | 14 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>判斷式(retf 4，一個遠指標參數)：回傳 p^[74h] = 7 | audit/function-index/pc98-overlay-25.md<br>spec/753-small-utility-routines.md |
| `13BC` | sub_13BC | — | 44 | 18 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>判斷式(retf 4，一個遠指標參數)：回傳 本模組 sub_134Ah(p) > p^[0E6h](有號比較) | audit/function-index/pc98-overlay-25.md<br>spec/754-small-predicates-and-wrappers.md |
| `13E8` | sub_13E8 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
