# dos-overlay-27 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `0007` | sub_7 | — | 43 | 15 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `0032` | sub_32 | — | 49 | 15 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `0063` | sub_63 | — | 28 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：唯一實質動作是 `call far ptr 297h:1110h`，參數原樣傳遞（body 共 28 bytes，已逐條讀完） | — |
| `007F` | sub_7F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
