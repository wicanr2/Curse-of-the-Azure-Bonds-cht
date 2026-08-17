# pc98-overlay-28 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADDRAWWIN | 22 | 8 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 3 個呼叫，沒有其他動作：`call far ptr 194h:43h`、`call far ptr 19Ah:2Ah`、`call far ptr 176h:57h`（body 共 22 bytes，已逐條讀完） | audit/far-call-map-dos.md<br>audit/far-call-map-pc98.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-28.md<br>audit/function-index/pc98-overlay-02.md |
| `0016` | sub_16 | DRAWWINDOW | 167 | 52 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/973-pc98-text-print-signature.md<br>單邊（DOS 無）：在迷宮就更新視野三參數並重畫，否則顯示 bigpic 121 世界地圖 | audit/far-call-map-dos.md<br>audit/far-call-map-pc98.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-28.md<br>audit/function-index/pc98-overlay-02.md |
| `00BD` | sub_BD | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
