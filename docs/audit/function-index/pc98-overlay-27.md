# pc98-overlay-27 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADOVERLAND | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md |
| `0007` | sub_7 | INITWILDCURSOR | 43 | 15 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>取值(retf，無參數)：q := DS:7F05h^[342h]，byte[q+160Ch] 與 byte[q+162Ch](相差 20h)→ DS:0BE04h / 0BE06h。與 DOS overlay-27:0007h 同形 | audit/function-index/pc98-overlay-27.md<br>spec/753-small-utility-routines.md |
| `0032` | sub_32 | SHOWWILDCURSOR | 49 | 15 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>兩次繪製(retf，無參數)：02A8h:098Fh(DS:0BE04h, DS:0BE06h, 遠指標@DS:9660h) 與 02A8h:0A86h(…, 遠指標@DS:0BE08h)。與 DOS overlay-27:0032h 同形 | audit/function-index/dos-overlay-02.md<br>audit/function-index/pc98-overlay-27.md<br>spec/1060-option-row-menu-core.md<br>spec/753-small-utility-routines.md |
| `0063` | sub_63 | HIDEWILDCURSOR | 28 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr 2A8h:0A86h`（body 共 28 bytes，已逐條讀完） | — |
| `007F` | sub_7F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-02.md |
