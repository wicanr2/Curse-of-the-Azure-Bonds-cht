# pc98-overlay-25 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTRAINING | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 19Ah:2Ah 與 164h:57h | audit/embedded-strings.md |
| `0011` | sub_11 | — | 928 | 342 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `03B1` | sub_3B1 | FIGLEVELSTUFF | 704 | 247 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0671` | sub_671 | ADJUSTCLERICALSPELLS | 370 | 121 | 2 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `07E3` | sub_7E3 | SETSAVETHROWS | 737 | 282 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0AC4` | sub_AC4 | SETTHIEFSKILLS | 595 | 224 | 2 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0D17` | sub_D17 | — | 370 | 145 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0F30` | sub_F30 | CHANGEHUMANCLASS | 1130 | 437 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `139A` | sub_139A | OLDHUMANCLASS | 92 | 33 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `13F6` | sub_13F6 | CURHUMANCLASS | 92 | 33 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1452` | sub_1452 | CURHUMANLEVEL | 80 | 29 | 2 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `14A2` | sub_14A2 | ISHUMAN | 34 | 14 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `14C4` | sub_14C4 | OLDCLASSOK | 44 | 18 | 4 | 2 | ✓ | 待解讀 | — | — | — |
| `14F0` | sub_14F0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
