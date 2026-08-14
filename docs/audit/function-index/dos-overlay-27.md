# dos-overlay-27 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md |
| `0007` | sub_7 | — | 43 | 15 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>取值(retf，無參數)：q := DS:4F99h^[342h]，把 byte[q+0A48h] 與 byte[q+0A68h](相差 20h)零延伸成 word 存進 DS:8C78h / 8C7Ah | audit/function-index/pc98-overlay-27.md<br>spec/753-small-utility-routines.md |
| `0032` | sub_32 | — | 49 | 15 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>兩次繪製(retf，無參數)：0297h:0EFFh(DS:8C78h, DS:8C7Ah, 遠指標@DS:65CAh) 與 0297h:1110h(DS:8C78h, DS:8C7Ah, 遠指標@DS:8C7Ch)。參數形狀由推入順序解出：末兩個 word 先高後低＝一個遠指標 | audit/function-index/dos-overlay-02.md<br>audit/function-index/pc98-overlay-27.md<br>spec/1060-option-row-menu-core.md<br>spec/753-small-utility-routines.md |
| `0063` | sub_63 | — | 28 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr 297h:1110h`（body 共 28 bytes，已逐條讀完） | — |
| `007F` | sub_7F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-02.md |
