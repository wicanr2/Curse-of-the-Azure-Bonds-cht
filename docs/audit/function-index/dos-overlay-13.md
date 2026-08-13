# dos-overlay-13 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 292 | 94 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0124` | sub_124 | — | 110 | 36 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `0192` | sub_192 | — | 196 | 74 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `031D` | sub_31D | — | 841 | 334 | 3 | 4 | ✓ | 待解讀 | — | — | — |
| `0666` | sub_666 | — | 233 | 87 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `074F` | sub_74F | — | 523 | 193 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `095A` | sub_95A | — | 935 | 334 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0D1C` | sub_D1C | — | 189 | 74 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0DD9` | sub_DD9 | — | 313 | 105 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0F12` | sub_F12 | — | 45 | 19 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0F46` | sub_F46 | — | 510 | 184 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1144` | sub_1144 | — | 169 | 57 | 4 | 2 | ✓ | 待解讀 | — | — | — |
| `1433` | sub_1433 | — | 18 | 7 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | — |
| `1476` | sub_1476 | — | 31 | 12 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `14E8` | sub_14E8 | — | 34 | 10 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr ds:75D7h, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 34 bytes，已逐條讀完） | — |
| `1513` | sub_1513 | — | 12 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+14h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `156D` | sub_156D | — | 8 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1590` | sub_1590 | — | 8 | 2 | 9 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+1A4h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 10 | 6 | 10 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `15A4` | sub_15A4 | — | 15 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1470+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `15B3` | sub_15B3 | — | 12 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-15h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `1652` | sub_1652 | — | 8 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+1A2h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `18A4` | sub_18A4 | — | 6 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr es:[di+19Bh], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `18AE` | sub_18AE | — | 19 | 7 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add di, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 19 bytes，已逐條讀完） | — |
| `1921` | sub_1921 | — | 5 | 2 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1926` | sub_1926 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1944` | sub_1944 | — | 6 | 3 | 4 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 6 bytes，已逐條讀完） | — |
| `194A` | sub_194A | — | 142 | 53 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `19D8` | sub_19D8 | — | 778 | 268 | 6 | 7 | ✓ | 待解讀 | — | — | — |
| `1CE2` | sub_1CE2 | — | 187 | 76 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1D9D` | sub_1D9D | — | 57 | 15 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `1DF6` | sub_1DF6 | — | 537 | 195 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `200F` | sub_200F | — | 507 | 176 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `2220` | sub_2220 | — | 1307 | 483 | 0 | 5 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `275A` | sub_275A | — | 361 | 132 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `28C3` | sub_28C3 | — | 223 | 73 | 3 | 3 | ✓ | 待解讀 | — | — | — |
| `29A2` | sub_29A2 | — | 601 | 239 | 6 | 2 | ✓ | 待解讀 | — | — | — |
| `2BFB` | sub_2BFB | — | 551 | 216 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `2E22` | sub_2E22 | — | 138 | 49 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2EAC` | sub_2EAC | — | 130 | 46 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2F3C` | sub_2F3C | — | 205 | 67 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3040` | sub_3040 | — | 432 | 177 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `31F0` | sub_31F0 | — | 361 | 132 | 2 | 10 | ✓ | 待解讀 | — | — | — |
| `33AC` | sub_33AC | — | 1278 | 493 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `38AA` | sub_38AA | — | 161 | 60 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `394B` | sub_394B | — | 428 | 142 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3B37` | sub_3B37 | — | 799 | 315 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `3E56` | sub_3E56 | — | 480 | 159 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `403F` | sub_403F | — | 291 | 117 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4162` | sub_4162 | — | 106 | 39 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `41CC` | sub_41CC | — | 186 | 63 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `42FD` | sub_42FD | — | 722 | 312 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `45CF` | sub_45CF | — | 308 | 111 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `4703` | sub_4703 | — | 264 | 93 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `4811` | sub_4811 | — | 183 | 70 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `48DC` | sub_48DC | — | 214 | 76 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `49B2` | sub_49B2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
