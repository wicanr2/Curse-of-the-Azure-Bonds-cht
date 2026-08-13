# pc98-overlay-30 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTHREED | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 2 個呼叫，沒有其他動作：`call far ptr 232h:34h`、`call far ptr 19Ah:2Ah`（body 共 17 bytes，已逐條讀完） | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md<br>spec/522-dos-buffer-four-plane-fill.md |
| `0011` | sub_11 | — | 346 | 150 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md<br>spec/README.md |
| `016B` | sub_16B | SET3DCOLORS | 33 | 13 | 0 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `018C` | sub_18C | CLEAR3DVIEW | 531 | 206 | 1 | 2 | ✓ | 待解讀 | — | — | spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/687-far-call-flattening-and-stack-leftover.md |
| `039F` | sub_39F | — | 270 | 108 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `04AD` | sub_4AD | — | 49 | 18 | 4 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `04DE` | sub_4DE | BLOCKCODE | 303 | 122 | 0 | 3 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/525-pc98-tempsearch-display-state.md<br>spec/528-pc98-moveparty-action-transaction-boundary.md |
| `060D` | sub_60D | WALLCODE | 259 | 105 | 3 | 2 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/525-pc98-tempsearch-display-state.md |
| `0710` | sub_710 | SPECIALCODE | 123 | 46 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `078B` | sub_78B | BUILDVIEW | 1977 | 825 | 0 | 6 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0F8F` | sub_F8F | LOADWALLSET | 666 | 268 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1253` | sub_1253 | LOAD3DMAP | 270 | 105 | 0 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/525-pc98-tempsearch-display-state.md |
| `1361` | sub_1361 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/524-dos-overlay30-geo-loader-source.md |
