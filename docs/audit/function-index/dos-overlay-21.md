# dos-overlay-21 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 27 | 9 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 4 個呼叫，沒有其他動作：`call far ptr loc_16BC+1`、`call far ptr loc_15CF+2`、`call far ptr loc_14AC+1`、`call loc_19E9+1`（body 共 27 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `001B` | sub_1B | — | 32 | 12 | 3 | 1 | ✓ | 已解讀 | strong inference | docs/spec/637-overlay21-small-batch.md<br>與 pc98 overlay-21:001Bh 助憶碼序列完全相同（12 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：回傳 <far 0176:0A31>(arg_0,arg_2) + 5DCh(1500) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-21.md<br>spec/637-overlay21-small-batch.md |
| `003B` | sub_3B | — | 28 | 10 | 3 | 0 | ✓ | 已解讀 | strong inference | docs/spec/637-overlay21-small-batch.md<br>與 pc98 overlay-21:003Bh 助憶碼序列完全相同（10 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：arg_2^[188h] -= arg_0(word)。與 0057h 逐指令相同只差 sub/add,兩支都不做上下限檢查——減到負數會回捲 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-21.md<br>audit/function-index/pc98-overlay-21.md<br>spec/637-overlay21-small-batch.md |
| `0057` | sub_57 | — | 28 | 10 | 4 | 0 | ✓ | 已解讀 | strong inference | docs/spec/637-overlay21-small-batch.md<br>與 pc98 overlay-21:0057h 助憶碼序列完全相同（10 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：arg_2^[188h] += arg_0(word)。與 003Bh 成對 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-21.md<br>audit/function-index/pc98-overlay-21.md<br>spec/637-overlay21-small-batch.md |
| `0073` | sub_73 | — | 84 | 31 | 5 | 2 | ✓ | 待解讀 | — | — | — |
| `00C7` | sub_C7 | — | 130 | 51 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0149` | sub_149 | — | 85 | 31 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/622-character-money-block.md<br>與 pc98 overlay-21:0148h 助憶碼序列完全相同（31 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：把 DS:9594h 指向的紀錄設成 arg_0 金額:先用 i=0..4 迴圈把 [di+0FBh+2i] 五個硬幣欄位歸零,再 div 5——商存 +103h、餘存 +101h。這證明 +FBh 起是五元素硬幣陣列(不是五個獨立欄位),且依 AD&D 的 5 gp = 1 pp 定出 +103h 是白金、+101h 是金幣,陣列由低價值排到高價值 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `019E` | sub_19E | — | 86 | 34 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `021C` | sub_21C | — | 145 | 53 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `02AD` | sub_2AD | — | 435 | 192 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `046B` | sub_46B | — | 188 | 71 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0527` | sub_527 | — | 210 | 79 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `05F9` | sub_5F9 | — | 86 | 30 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `064F` | sub_64F | — | 948 | 382 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0A03` | sub_A03 | — | 124 | 50 | 0 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0A8A` | sub_A8A | — | 206 | 87 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B91` | sub_B91 | — | 295 | 125 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0CF0` | sub_CF0 | — | 619 | 243 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `0F5B` | sub_F5B | — | 84 | 30 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0FAF` | sub_FAF | — | 62 | 26 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0FED` | sub_FED | — | 1910 | 669 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-21.md |
| `145D` | sub_145D | — | 180 | 59 | 4 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FEDh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1513` | sub_1513 | — | 45 | 16 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp ax, 39h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 45 bytes，已逐條讀完） | audit/embedded-strings.md |
| `154F` | sub_154F | — | 15 | 6 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `imul cx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `155E` | sub_155E | — | 277 | 99 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FEDh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `198C` | sub_198C | — | 133 | 40 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2247` | sub_2247 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
