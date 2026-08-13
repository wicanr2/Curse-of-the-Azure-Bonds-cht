# dos-overlay-14 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 62 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call far ptr 141h:9Dh`、`call far ptr 14Dh:101h`、`call far ptr 0FDh:7Ah`、`call far ptr 175h:2Ah`、`call loc_EB2+4`、`call far ptr 196h:43h`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md |
| `003E` | sub_3E | — | 277 | 6 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md |
| `0153` | sub_153 | — | 309 | 127 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0288` | sub_288 | — | 122 | 43 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0302` | sub_302 | — | 742 | 303 | 1 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `05E8` | sub_5E8 | — | 201 | 81 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `06B1` | sub_6B1 | — | 67 | 25 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `06F4` | sub_6F4 | — | 85 | 33 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0749` | sub_749 | — | 69 | 29 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `078E` | sub_78E | — | 174 | 55 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `083C` | sub_83C | — | 196 | 70 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `090A` | sub_90A | — | 644 | 248 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BAF` | sub_BAF | — | 780 | 334 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `0EBB` | sub_EBB | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
