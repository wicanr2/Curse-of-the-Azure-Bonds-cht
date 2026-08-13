# pc98-overlay-14 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADMOVEMENT | 62 | 16 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call far ptr 13Eh:9Dh`、`call far ptr 14Ah:101h`、`call far ptr 0FAh:7Ah`、`call far ptr 172h:2Ah`、`call far ptr 0DCh:66h`、`call far ptr 194h:43h`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md |
| `003E` | sub_3E | — | 270 | 6 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `jmp loc_146`（body 共 270 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md |
| `014C` | sub_14C | — | 303 | 126 | 2 | 1 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/527-pc98-moveparty-action-writer-boundary.md<br>spec/533-pc98-moveparty-helper-gate.md<br>spec/534-chinese-manual-moveparty-character-transfer.md |
| `027B` | sub_27B | — | 122 | 43 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02F5` | sub_2F5 | — | 703 | 299 | 1 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md<br>spec/528-pc98-moveparty-action-transaction-boundary.md |
| `05B4` | sub_5B4 | — | 201 | 81 | 1 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md<br>spec/528-pc98-moveparty-action-transaction-boundary.md |
| `067D` | sub_67D | — | 67 | 25 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `06C0` | sub_6C0 | — | 84 | 33 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0714` | sub_714 | — | 69 | 29 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md |
| `0759` | sub_759 | — | 174 | 55 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0807` | sub_807 | — | 203 | 73 | 1 | 1 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/528-pc98-moveparty-action-transaction-boundary.md<br>spec/533-pc98-moveparty-helper-gate.md |
| `08E4` | sub_8E4 | PREMOVEPARTY | 699 | 267 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0BCC` | sub_BCC | MOVEPARTY | 471 | 190 | 0 | 8 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/526-pc98-moveparty-map-writer-searchrec-correction.md<br>spec/527-pc98-moveparty-action-writer-boundary.md<br>spec/533-pc98-moveparty-helper-gate.md<br>spec/534-chinese-manual-moveparty-character-transfer.md |
| `0DA3` | sub_DA3 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
