# dos-overlay-32 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 196h:43h 後交給 19Ch:2Ah | — |
| `0011` | sub_11 | — | 104 | 41 | 6 | 1 | ✓ | 待解讀 | — | — | — |
| `0079` | sub_79 | — | 123 | 51 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `00F4` | sub_F4 | — | 64 | 31 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>同上形狀,參數為 (0Fh,8)/(8,0) 等;兩個分支的參數順序不對稱,照抄 | spec/585-ecl-goto-and-display-mode-pair.md<br>spec/635-overlay32-grid-record-array.md |
| `0134` | sub_134 | — | 64 | 31 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>同上形狀,參數為 (0,0)/(8,8) | spec/585-ecl-goto-and-display-mode-pair.md |
| `0174` | sub_174 | — | 428 | 190 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0320` | sub_320 | — | 203 | 87 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `03EB` | sub_3EB | — | 286 | 122 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0509` | sub_509 | — | 111 | 43 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `0578` | sub_578 | — | 442 | 196 | 3 | 4 | ✓ | 待解讀 | — | — | — |
| `0732` | sub_732 | — | 49 | 18 | 5 | 1 | ✓ | 已解讀 | strong inference | docs/spec/635-overlay32-grid-record-array.md<br>與 pc98 overlay-32:0A1Eh 助憶碼序列完全相同（18 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：回傳 (0 <= arg_2 <= 6) and (0 <= arg_0 <= 6),用**有號**比較(jl/jg/jle)所以負值被擋掉而不是當成大正數。7×7 = 49 格的座標範圍檢查 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0763` | sub_763 | — | 186 | 79 | 5 | 4 | ✓ | 待解讀 | — | — | — |
| `081D` | sub_81D | — | 450 | 167 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `09DF` | sub_9DF | — | 352 | 142 | 3 | 6 | ✓ | 待解讀 | — | — | — |
| `0B3F` | sub_B3F | — | 247 | 101 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0C36` | sub_C36 | — | 40 | 17 | 3 | 1 | ✓ | 已解讀 | strong inference | docs/spec/635-overlay32-grid-record-array.md<br>與 PC-98 overlay-32:12EBh（entry#15）助憶碼序列完全相同，語意同該筆：i := <sub_1363>(arg_0,arg_2);回傳 byte[973Dh + 4*i + 0]。DS:973Dh 是每筆 4 bytes 的記錄陣列(位移 -68C3h 換成無號即 973Dh),與 1313h/133Bh 只差取哪個欄位 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0C5E` | sub_C5E | — | 40 | 17 | 3 | 1 | ✓ | 已解讀 | strong inference | docs/spec/635-overlay32-grid-record-array.md<br>與 PC-98 overlay-32:1313h（entry#16）助憶碼序列完全相同，語意同該筆：同 12EBh,但取 byte[973Dh + 4*i + 1](位移 -68C2h) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0C86` | sub_C86 | — | 40 | 17 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/635-overlay32-grid-record-array.md<br>與 PC-98 overlay-32:133Bh（entry#17）助憶碼序列完全相同，語意同該筆：同 12EBh,但取 byte[973Dh + 4*i + 3](位移 -68C0h)。+2 這三支都沒有讀 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0CAE` | sub_CAE | — | 88 | 32 | 9 | 1 | ✓ | 待解讀 | — | — | — |
| `0D06` | sub_D06 | — | 367 | 147 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0E75` | sub_E75 | — | 742 | 293 | 0 | 10 | ✓ | 待解讀 | — | — | — |
| `115B` | sub_115B | — | 506 | 197 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `1355` | sub_1355 | — | 108 | 41 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `13C1` | sub_13C1 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
