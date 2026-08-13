# dos-overlay-25 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0C51` | nullsub_1 | — | 1 | 1 | 0 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 1 bytes，已逐條讀完） | — |
| `0EE1` | sub_EE1 | — | 945 | 360 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1292` | sub_1292 | — | 92 | 33 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `12EE` | sub_12EE | — | 92 | 33 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `134A` | sub_134A | — | 80 | 29 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `139A` | sub_139A | — | 34 | 14 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `13BC` | sub_13BC | — | 44 | 18 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `13E8` | sub_13E8 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
