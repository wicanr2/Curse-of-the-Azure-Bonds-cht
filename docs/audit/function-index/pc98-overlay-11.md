# pc98-overlay-11 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADINIT | 47 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 8 個呼叫，沒有其他動作：`call far ptr 0C9h:2Fh`、`call far ptr 117h:57h`、`call far ptr 232h:34h`、`call far ptr 194h:43h`、`call far ptr 176h:57h`、`call far ptr 8Bh:2Ah`（body 共 47 bytes，已逐條讀完） | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md<br>spec/525-pc98-tempsearch-display-state.md<br>spec/572-resident-service-functions.md |
| `0048` | sub_48 | INITALL | 1209 | 503 | 0 | 3 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md |
| `0508` | sub_508 | INITVARS | 606 | 200 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0766` | sub_766 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `07CD` | sub_7CD | — | 51 | 27 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0800` | sub_800 | — | 40 | 26 | 1 | 1 |  | 待解讀 | — | — | project-status.md |
