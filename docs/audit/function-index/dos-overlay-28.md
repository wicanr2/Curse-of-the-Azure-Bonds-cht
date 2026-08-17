# dos-overlay-28 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 22 | 8 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 3 個呼叫，沒有其他動作：`call far ptr 196h:43h`、`call far ptr 19Ch:2Ah`、`call far ptr 179h:57h`（body 共 22 bytes，已逐條讀完） | audit/far-call-map-dos.md<br>audit/far-call-map-pc98.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-28.md<br>audit/function-index/pc98-overlay-02.md |
| `0016` | sub_16 | — | 57 | 22 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>三項查表(retf 4)：只用第一個宣告的參數([bp+8])，08h→00h、41h→00h、0E8h→0Fh、其他→0Fh。第二個參數整支沒讀；0E8h 分支與 default 同值(多餘的比較) | audit/far-call-map-dos.md<br>audit/far-call-map-pc98.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-28.md<br>audit/function-index/pc98-overlay-02.md |
| `004F` | sub_4F | — | 57 | 22 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>三項查表(retf 4)：08h→08h、41h→07h、0E8h→07h、其他→07h。第二個參數整支沒讀；0E8h 分支與 default 同值 | audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-28.md<br>audit/function-index/pc98-overlay-15.md<br>context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/754-small-predicates-and-wrappers.md |
| `0088` | sub_88 | — | 68 | 26 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>三項查表(retf 4)：只用第一個宣告的參數，0Bh→8、9→6、0DBh→6、6→1、其他→6。與 spec 754 的 overlay-28:0016h/004Fh 同族 | project-status.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/976-view-refresh-field-redraw-and-los-collect.md |
| `00CC` | sub_CC | — | 235 | 82 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/976-view-refresh-field-redraw-and-los-collect.md<br>視野更新完整版（pc98 0016h 是簡化版）：補齊九組跨平台全域對應；PC-98 砍掉 DS:8B67h 那條分支 | audit/far-call-map-dos.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/976-view-refresh-field-redraw-and-los-collect.md<br>spec/979-mass-cure-effect-and-two-piece-icon.md |
| `01B7` | sub_1B7 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
