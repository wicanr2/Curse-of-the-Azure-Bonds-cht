# dos-overlay-06 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 37 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化(LOADSHOP):依序 far call 依賴單元的初始化 entry,最後呼叫 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md |
| `0050` | sub_50 | — | 762 | 323 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0355` | sub_355 | — | 269 | 96 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0474` | sub_474 | — | 440 | 166 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `06F6` | sub_6F6 | — | 631 | 270 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `096D` | sub_96D | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
