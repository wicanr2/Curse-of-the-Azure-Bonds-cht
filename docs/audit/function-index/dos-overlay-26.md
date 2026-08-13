# dos-overlay-26 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 179h:57h 後交給 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>spec/686-dos-ecl-operand-tables-and-menu.md |
| `0011` | sub_11 | — | 99 | 38 | 5 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>spec/686-dos-ecl-operand-tables-and-menu.md |
| `0074` | sub_74 | — | 65 | 25 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `00D5` | sub_D5 | — | 240 | 92 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `01E5` | sub_1E5 | — | 431 | 180 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `03D4` | sub_3D4 | — | 1041 | 365 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `07E5` | sub_7E5 | — | 24 | 14 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr 542h:311h`（body 共 24 bytes，已逐條讀完） | — |
| `07FD` | sub_7FD | — | 141 | 52 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `088A` | sub_88A | — | 314 | 121 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `09C4` | sub_9C4 | — | 76 | 31 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0A10` | sub_A10 | — | 70 | 29 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0A56` | sub_A56 | — | 162 | 71 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0AF8` | sub_AF8 | — | 257 | 111 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0BF9` | sub_BF9 | — | 224 | 77 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0CD9` | sub_CD9 | — | 146 | 49 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0D6B` | sub_D6B | — | 124 | 44 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0DF9` | sub_DF9 | — | 778 | 312 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `112A` | sub_112A | — | 111 | 52 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1199` | sub_1199 | — | 168 | 59 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1241` | sub_1241 | — | 76 | 30 | 0 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `128D` | sub_128D | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
