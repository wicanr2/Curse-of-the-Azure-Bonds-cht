# pc98-overlay-03 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPROTECT | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/569-small-function-batch-reading.md |
| `000C` | sub_C | GAMECLEAN | 18 | 8 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>永遠回 1(retf，無參數)：配置 19Ah bytes 的 local 卻一個都沒用到，直接 mov byte [bp-1],1 後回傳 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>spec/753-small-utility-routines.md |
| `0131` | sub_131 | DOPROTECT | 715 | 310 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `03FC` | sub_3FC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
