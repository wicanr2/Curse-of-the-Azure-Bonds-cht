# dos-overlay-20 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 32 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call far ptr 19Ch:2Ah`、`call far ptr 11Ah:57h`、`call far ptr 141h:9Dh`、`call far ptr 14Dh:101h`、`call far ptr 167h:4Dh`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/754-small-predicates-and-wrappers.md |
| `0020` | sub_20 | — | 696 | 239 | 1 | 1 | ✓ | 待解讀 | — | — | spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md |
| `02D8` | sub_2D8 | — | 173 | 69 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0385` | sub_385 | — | 52 | 20 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>上限 99 的累加(retf，無參數)：先叫本模組 sub_2D8h(@DS:7566h)；若 DS:7570h>0 則 DS:756Eh += DS:3E32h * DS:7570h、DS:7570h:=0，最後 DS:756Eh 夾到 99。全部無號 word | audit/function-index/pc98-overlay-20.md<br>spec/754-small-predicates-and-wrappers.md |
| `03B9` | sub_3B9 | — | 165 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `045E` | sub_45E | — | 233 | 92 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0549` | sub_549 | — | 126 | 54 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md<br>兩位數補零(retf 2)：先 14Dh:0052h 把數值轉成字串，n < 0Ah(無號)時在前面接上 CS:0547h 的 '0'，最後 0A54h:0680h(緩衝A, 緩衝B, 1, 2) 取固定寬度再寫回輸出參數。⚠ 補的是 ASCII '0'，換全形數字會讓寬度計算不一致 | spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md |
| `05D4` | sub_5D4 | — | 246 | 110 | 5 | 2 | ✓ | 待解讀 | — | — | — |
| `0712` | sub_712 | — | 310 | 131 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0862` | sub_862 | — | 173 | 65 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `091D` | sub_91D | — | 182 | 68 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `09DF` | sub_9DF | — | 294 | 105 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0B05` | sub_B05 | — | 177 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0BB6` | sub_BB6 | — | 178 | 65 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0C9C` | sub_C9C | — | 492 | 196 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `0E88` | sub_E88 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
