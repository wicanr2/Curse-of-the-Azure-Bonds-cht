# pc98-overlay-33 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSQRPAK24 | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/overlay-init-graph.md |
| `004C` | sub_4C | LOAD24X24SET | 539 | 206 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0267` | sub_267 | PUT24X24SYMBOL | 97 | 39 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `02E8` | sub_2E8 | DISPOSEFIGURE | 116 | 48 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `038F` | sub_38F | LOADFIGURE | 1050 | 451 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `07A9` | sub_7A9 | PUTFIGURE | 285 | 114 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `08C6` | sub_8C6 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
