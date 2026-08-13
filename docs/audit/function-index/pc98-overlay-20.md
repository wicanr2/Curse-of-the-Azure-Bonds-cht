# pc98-overlay-20 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCLOCK | 32 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call far ptr 19Ah:2Ah`、`call far ptr 117h:57h`、`call far ptr 13Eh:9Dh`、`call far ptr 14Ah:101h`、`call far ptr 164h:57h`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0020` | sub_20 | — | 696 | 239 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02D8` | sub_2D8 | — | 173 | 69 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0385` | sub_385 | — | 52 | 20 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `03B9` | sub_3B9 | TICKCLOCK | 165 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `045E` | sub_45E | — | 233 | 92 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0549` | sub_549 | — | 126 | 54 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `05D2` | sub_5D2 | — | 246 | 110 | 5 | 2 | ✓ | 待解讀 | — | — | — |
| `073A` | sub_73A | — | 329 | 140 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0898` | sub_898 | — | 173 | 65 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `094E` | sub_94E | — | 182 | 68 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0A11` | sub_A11 | — | 294 | 105 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0B37` | sub_B37 | — | 177 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BE8` | sub_BE8 | — | 178 | 65 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0CD0` | sub_CD0 | REST | 492 | 196 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `0EBC` | sub_EBC | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
