# dos-overlay-06 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 37 | 11 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0050` | sub_50 | — | 762 | 323 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0355` | sub_355 | — | 269 | 96 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0474` | sub_474 | — | 440 | 166 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `06F6` | sub_6F6 | — | 631 | 270 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `096D` | sub_96D | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
