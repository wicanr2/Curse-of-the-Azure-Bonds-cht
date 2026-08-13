# dos-overlay-22 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 32 | 10 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call far ptr sub_15D1`、`call loc_1656+1`、`call far ptr loc_14AB+2`、`call sub_16BD`、`call loc_19E7+3`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `0020` | sub_20 | — | 367 | 125 | 1 | 2 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `0219` | sub_219 | — | 470 | 194 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `03F5` | sub_3F5 | — | 436 | 168 | 1 | 1 | ✓ | 待解讀 | — | — | spec/517-reverse-engineering-gap-inventory.md |
| `05B4` | sub_5B4 | — | 749 | 291 | 1 | 2 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md |
| `08A1` | sub_8A1 | — | 224 | 76 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `0981` | sub_981 | — | 123 | 44 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `09FC` | sub_9FC | — | 875 | 326 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0D67` | sub_D67 | — | 170 | 67 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0E11` | sub_E11 | — | 245 | 100 | 10 | 3 | ✓ | 待解讀 | — | — | — |
| `0F06` | sub_F06 | — | 441 | 168 | 47 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/string-pairs.md |
| `10D2` | sub_10D2 | — | 282 | 102 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1263` | sub_1263 | — | 663 | 266 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `145D` | sub_145D | — | 5 | 3 | 22 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1462` | sub_1462 | — | 7 | 3 | 14 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:755Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1499` | sub_1499 | — | 5 | 2 | 10 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp ax, 2Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `149E` | sub_149E | — | 195 | 71 | 6 | 4 |  | 待解讀 | — | — | — |
| `1513` | sub_1513 | — | 18 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+9]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1586` | sub_1586 | — | 10 | 6 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1590` | sub_1590 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:634h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1595` | sub_1595 | — | 5 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 11 | 5 | 6 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, 59h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `15CC` | sub_15CC | — | 5 | 1 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_153C+4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 7 | 2 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_13FE`，控制權轉交後不返回；先設定 `mov byte ptr [bp-1], 0`（body 共 7 bytes，已逐條讀完） | — |
| `15FB` | sub_15FB | — | 67 | 26 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `163E` | sub_163E | — | 77 | 27 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `168B` | sub_168B | — | 30 | 13 | 1 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ss:[di-1Dh], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `16B8` | sub_16B8 | — | 5 | 2 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_1748`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 158 | 55 | 2 | 3 |  | 待解讀 | — | — | — |
| `175B` | sub_175B | — | 324 | 128 | 5 | 4 | ✓ | 待解讀 | — | — | — |
| `189F` | sub_189F | — | 5 | 3 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18A4` | sub_18A4 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-78h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18A9` | sub_18A9 | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr ds:6E94h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `190D` | sub_190D | — | 5 | 3 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-73h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1912` | sub_1912 | — | 11 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr [bp-72h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `192B` | sub_192B | — | 5 | 2 | 16 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1930` | sub_1930 | — | 405 | 173 | 16 | 6 |  | 待解讀 | — | — | — |
| `193A` | sub_193A | — | 16 | 6 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, [bp-72h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 16 bytes，已逐條讀完） | — |
| `1ABF` | sub_1ABF | — | 378 | 143 | 3 | 7 | ✓ | 待解讀 | — | — | — |
| `1C39` | sub_1C39 | — | 181 | 75 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `1CF9` | sub_1CF9 | — | 39 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1D2A` | sub_1D2A | — | 43 | 19 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is Cursed」（unk_1D20，長度 9）呼叫訊息 routine（body 共 43 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1D55` | sub_1D55 | — | 61 | 24 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `1D93` | sub_1D93 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_1D92，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `1DD5` | sub_1DD5 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is affected」（unk_1DC9，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E0F` | sub_1E0F | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is protected」（unk_1E02，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E4E` | sub_1E4E | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is cold-resistant」（unk_1E3C，長度 17）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E7C` | sub_1E7C | — | 52 | 27 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_1E7B，長度 0）呼叫訊息 routine（body 共 52 bytes，已逐條讀完） | — |
| `1EC9` | sub_1EC9 | — | 6 | 3 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 16h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1FA0` | sub_1FA0 | — | 304 | 117 | 0 | 5 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md |
| `20E1` | sub_20E1 | — | 147 | 62 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2180` | sub_2180 | — | 70 | 34 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is friendly」（unk_2174，長度 11）呼叫訊息 routine（body 共 70 bytes，已逐條讀完） | — |
| `2232` | sub_2232 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is shielded」（unk_2226，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2260` | sub_2260 | — | 71 | 37 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `22B4` | sub_22B4 | — | 310 | 122 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `23F2` | sub_23F2 | — | 87 | 36 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `245B` | sub_245B | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is fire resistant」（unk_2449，長度 17）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/spell-status-message-strings.md |
| `2494` | sub_2494 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is silenced」（unk_2488，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `24CD` | sub_24CD | — | 168 | 73 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2580` | sub_2580 | — | 174 | 67 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `262F` | sub_262F | — | 70 | 36 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2682` | sub_2682 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_2675，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `26BB` | sub_26BB | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「Knock-Knock」（unk_26AF，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `26F6` | sub_26F6 | — | 82 | 39 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2754` | sub_2754 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is weakened」（unk_2748，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2799` | sub_2799 | — | 1120 | 433 | 0 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `2BF9` | sub_2BF9 | — | 433 | 148 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2DB6` | sub_2DB6 | — | 416 | 137 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `2F5E` | sub_2F5E | — | 61 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2FA4` | sub_2FA4 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is blind」（unk_2F9B，長度 8）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2FD1` | sub_2FD1 | — | 165 | 61 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3076` | sub_3076 | — | 17 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call near ptr sub_2FD1`（body 共 17 bytes，已逐條讀完） | — |
| `3093` | sub_3093 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is diseased」（unk_3087，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `30C0` | sub_30C0 | — | 178 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `317E` | sub_317E | — | 932 | 349 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `352D` | sub_352D | — | 73 | 35 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `3599` | sub_3599 | — | 253 | 87 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `36A7` | sub_36A7 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「has been cursed!」（unk_3696，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `36E0` | sub_36E0 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is blinking」（unk_36D4，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `370E` | sub_370E | — | 246 | 103 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `3804` | sub_3804 | — | 271 | 109 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `391D` | sub_391D | — | 42 | 20 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `3947` | sub_3947 | — | 188 | 71 | 5 | 3 | ✓ | 待解讀 | — | — | — |
| `3A03` | sub_3A03 | — | 661 | 263 | 4 | 8 | ✓ | 待解讀 | — | — | — |
| `3C98` | sub_3C98 | — | 75 | 39 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3CED` | sub_3CED | — | 46 | 21 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is Slowed」（unk_3CE3，長度 9）呼叫訊息 routine（body 共 46 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `3D27` | sub_3D27 | — | 433 | 158 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3EE2` | sub_3EE2 | — | 65 | 32 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `3F23` | sub_3F23 | — | 64 | 26 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3F6F` | sub_3F6F | — | 133 | 56 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3FF4` | sub_3FF4 | — | 66 | 37 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4043` | sub_4043 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is paralyzed」（unk_4036，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `407A` | sub_407A | — | 78 | 33 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `40D5` | sub_40D5 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_40C8，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4103` | sub_4103 | — | 59 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `413F` | sub_413F | — | 57 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4194` | sub_4194 | — | 244 | 95 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4289` | sub_4289 | — | 117 | 48 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4311` | sub_4311 | — | 150 | 64 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `43A7` | sub_43A7 | — | 66 | 26 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `43EA` | sub_43EA | — | 59 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4432` | sub_4432 | — | 96 | 50 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4493` | sub_4493 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_4492，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `44D3` | sub_44D3 | — | 203 | 72 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `45B5` | sub_45B5 | — | 176 | 72 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `4672` | sub_4672 | — | 171 | 72 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `472C` | sub_472C | — | 32 | 17 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is highlighted」（unk_471D，長度 14）呼叫訊息 routine（body 共 32 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4759` | sub_4759 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_474C，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4791` | sub_4791 | — | 133 | 57 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4822` | sub_4822 | — | 184 | 79 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `48E4` | sub_48E4 | — | 398 | 161 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `4A8F` | sub_4A8F | — | 292 | 115 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `4BEC` | sub_4BEC | — | 416 | 194 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `4DA0` | sub_4DA0 | — | 275 | 132 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `4EB4` | sub_4EB4 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_4EB3，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `4EF7` | sub_4EF7 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is protected」（unk_4EEA，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `54EB` | sub_54EB | — | 164 | 72 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5590` | sub_5590 | — | 170 | 70 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `563B` | sub_563B | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_563A，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5669` | sub_5669 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5668，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5697` | sub_5697 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5696，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `56C4` | sub_56C4 | — | 189 | 75 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `578B` | sub_578B | — | 78 | 33 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `57E3` | sub_57E3 | — | 368 | 158 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `5974` | sub_5974 | — | 294 | 114 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `5AA8` | sub_5AA8 | — | 464 | 173 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5C86` | sub_5C86 | — | 381 | 145 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `5E11` | sub_5E11 | — | 269 | 102 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5F2F` | sub_5F2F | — | 36 | 12 | 0 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push [bp+arg_6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 36 bytes，已逐條讀完） | — |
| `5F53` | sub_5F53 | — | 185 | 87 | 4 | 6 |  | 待解讀 | — | — | — |
| `6022` | sub_6022 | — | 268 | 105 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `612E` | sub_612E | — | 77 | 29 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `617B` | sub_617B | — | 121 | 42 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `6209` | sub_6209 | — | 258 | 119 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `68A5` | sub_68A5 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
