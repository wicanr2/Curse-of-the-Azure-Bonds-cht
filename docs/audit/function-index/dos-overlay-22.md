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
| `0981` | sub_981 | — | 123 | 44 | 1 | 3 | ✓ | 已解讀 | strong inference | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>與 pc98 overlay-22:09EEh 助憶碼序列完全相同（44 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：清 DS:7CCEh 起每筆 4 bytes 的 far pointer 陣列(後測迴圈,i 走 0..30h 共 **49 筆**不是 48),再把 DS:7D92h 清 0;之後走 DS:9594h^[14Eh] 物品鏈(next 在 +52h),<sub_64E8>(p) 非 0 時呼叫 <sub_90E>(arg_0)——注意那支只吃 arg_0 不吃目前這個物品,作用對象另有來源 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `09FC` | sub_9FC | — | 875 | 326 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0D67` | sub_D67 | — | 170 | 67 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0E11` | sub_E11 | — | 245 | 100 | 10 | 3 | ✓ | 待解讀 | — | — | — |
| `0F06` | sub_F06 | — | 441 | 168 | 47 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/string-pairs.md |
| `10D2` | sub_10D2 | — | 282 | 102 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1263` | sub_1263 | — | 663 | 266 | 0 | 9 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-22.md |
| `145D` | sub_145D | — | 5 | 3 | 22 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1462` | sub_1462 | — | 7 | 3 | 14 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:755Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1499` | sub_1499 | — | 5 | 2 | 10 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp ax, 2Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `149E` | sub_149E | — | 195 | 71 | 6 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1263h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1513` | sub_1513 | — | 18 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+9]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1586` | sub_1586 | — | 10 | 6 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1590` | sub_1590 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:634h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1595` | sub_1595 | — | 5 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 11 | 5 | 6 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, 59h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `15CC` | sub_15CC | — | 5 | 1 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_153C+4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 7 | 2 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_13FE`，控制權轉交後不返回；先設定 `mov byte ptr [bp-1], 0`（body 共 7 bytes，已逐條讀完） | — |
| `15FB` | sub_15FB | — | 67 | 26 | 3 | 1 | ✓ | 已解讀 | exact | docs/spec/681-dos-target-array-and-reduce.md<br>四個 byte 參數各自 cbw(有號)擴成 word,依序寫進 ss:[arg_0-1Ch/-1Ah/-18h/-16h],再把該塊位址傳給 <far loc_1898h>。arg_0 是**呼叫端的 BP**(用 ss: 定址),所以這支直接寫進呼叫端的區域變數區而不是自己的 frame;參數順序在傳遞時反過來(arg_8 放最前) | spec/681-dos-target-array-and-reduce.md |
| `163E` | sub_163E | — | 77 | 27 | 2 | 1 | ✓ | 已解讀 | strong inference | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>與 pc98 overlay-22:1897h 助憶碼序列完全相同（27 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：把 arg_8^ 夾在 0..31h(49)、arg_4^ 夾在 0..18h(24),即 50 × 25。用 jle/jge(**有號**)所以負值真的會被夾到 0——與 spec 671 的 15731h 對照,那支用無號 jb 所以下限判斷是死碼,同一專案裡兩種寫法都有。參數是指標,夾限結果寫回呼叫端 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `168B` | sub_168B | — | 30 | 13 | 1 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | audit/function-index/dos-overlay-22.md |
| `16A9` | sub_16A9 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ss:[di-1Dh], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `16B8` | sub_16B8 | — | 5 | 2 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_1748`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 158 | 55 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 168Bh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `175B` | sub_175B | — | 324 | 128 | 5 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-22.md |
| `189F` | sub_189F | — | 5 | 3 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18A4` | sub_18A4 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-78h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `18A9` | sub_18A9 | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr ds:6E94h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `190D` | sub_190D | — | 5 | 3 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-73h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1912` | sub_1912 | — | 11 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr [bp-72h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `192B` | sub_192B | — | 5 | 2 | 16 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1930` | sub_1930 | — | 405 | 173 | 16 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 175Bh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `193A` | sub_193A | — | 16 | 6 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, [bp-72h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 16 bytes，已逐條讀完） | — |
| `1ABF` | sub_1ABF | — | 378 | 143 | 3 | 7 | ✓ | 待解讀 | — | — | — |
| `1C39` | sub_1C39 | — | 181 | 75 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `1CF9` | sub_1CF9 | — | 39 | 18 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:1F53h 助憶碼序列完全相同（18 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：<sub_1E8D>(DS:9594h^[198h], 'は祝福を受けた。')——薄包裝,除了取欄位與備妥字串外沒有其他動作 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1D2A` | sub_1D2A | — | 43 | 19 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is Cursed」（unk_1D20，長度 9）呼叫訊息 routine（body 共 43 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1D55` | sub_1D55 | — | 61 | 24 | 0 | 4 | ✓ | 已解讀 | strong inference | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>與 pc98 overlay-22:1FB6h 助憶碼序列完全相同（24 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：DS:0A520h(選定目標數)為 0 就直接返回,連骰都不擲;否則 <far 013Eh:008Eh>(t, ROLLDICE(1,8), 0) 非 0 時再呼叫 <far 014Ah:054Fh>(t)。1d8 正是 AD&D 的 Cure Light Wounds——加上 4416h 的 2d8+1 與 46CAh 的 3d8+3,cure 系列三個標準級距都齊了 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1D93` | sub_1D93 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_1D92，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `1DD5` | sub_1DD5 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is affected」（unk_1DC9，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E0F` | sub_1E0F | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is protected」（unk_1E02，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E4E` | sub_1E4E | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is cold-resistant」（unk_1E3C，長度 17）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1E7C` | sub_1E7C | — | 52 | 27 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_1E7B，長度 0）呼叫訊息 routine（body 共 52 bytes，已逐條讀完） | — |
| `1EC9` | sub_1EC9 | — | 6 | 3 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 16h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1FA0` | sub_1FA0 | — | 304 | 117 | 0 | 5 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md |
| `20E1` | sub_20E1 | — | 147 | 62 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/681-dos-target-array-and-reduce.md<br>p := DS:7435h(far pointer),為 nil 或 DS:7434h(byte 計數)為 0 就返回——這對「計數 + 第 1 筆」與 PC-98 的目標陣列(0A520h/0A521h)同一個形狀,推得 DOS 基底是 7431h(strong inference)。之後四道條件:⚠ 第三道要 <loc_1456h>(0,4,p) 等於 0、第四道要 <loc_1573h>(p,0Ch,@var) 不等於 0,兩道相鄰而方向相反。通過才移除效果 0Ch、呼叫 <far loc_14A4h> 並印英文原文 'has been reduced' | spec/681-dos-target-array-and-reduce.md |
| `2180` | sub_2180 | — | 70 | 34 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is friendly」（unk_2174，長度 11）呼叫訊息 routine（body 共 70 bytes，已逐條讀完） | — |
| `21C7` | sub_21C7 | — | 95 | 47 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/682-dos-spell-power-calc.md<br>n := <far 013Eh:05A4h>(DS:6F97h) + 1;v := <sub_1462>(DS:6F97h, 0, 0, n div 2, 4);再**重算一次** n div 2 當加項,<sub_F06>(8, v + n div 2, 空字串)。同一個除法算兩遍,與 129EBh 的重複除法同一種寫法 | audit/function-index/pc98-overlay-22.md<br>spec/682-dos-spell-power-calc.md |
| `2232` | sub_2232 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is shielded」（unk_2226，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2260` | sub_2260 | — | 71 | 37 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/682-dos-spell-power-calc.md<br>⚠ 推六個但 sub_1462 只吃五個:最前面那個 DS:6F97h 留在堆疊上,呼叫後用 pop dx 取回來當加項。這是刻意把「要傳的參數」與「事後要用的值」用同一個推入完成——照「呼叫前的推入都是參數」去數,會把 sub_1462 算成 6 個參數。中間還有巢狀呼叫 <far loc_15A3h>(DS:6F97h)。字串同樣是空的 | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-index/pc98-overlay-22.md<br>spec/682-dos-spell-power-calc.md |
| `22B4` | sub_22B4 | — | 310 | 122 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `23F2` | sub_23F2 | — | 87 | 36 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/702-more-overlay22-spell-handlers.md<br>套用 'is held'(CS:23EAh):v := DS:7434h;v=1 時再看 DS:6F97h 是否為 17h 決定 r 是 −2 還是 −3;v=2 → −1;v=3 或 4 → 0;**v 為其他值時 r 不被寫入,回傳堆疊殘值**。最後 <sub_1ABFh>(r, 訊息)。⚠ DS:7434h 在治療系(spec 700)只當非 0 旗標用,這裡卻是 1..4 的列舉 | spec/702-more-overlay22-spell-handlers.md |
| `245B` | sub_245B | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is fire resistant」（unk_2449，長度 17）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/spell-status-message-strings.md |
| `2494` | sub_2494 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is silenced」（unk_2488，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `24CD` | sub_24CD | — | 168 | 73 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2580` | sub_2580 | — | 174 | 67 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/630-spell-target-array.md<br>與 pc98 overlay-22:282Dh 助憶碼序列完全相同（67 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：挑目標填進陣列:budget := DS:9594h^[1A5h](存 DS:7D93h),掃 DS:9598h 角色鏈,+11Ah(RACETYPE)=0Eh 且 budget >= p^[1A5h] 時扣掉該值、inc DS:0A520h、把 far pointer 存進 [0A51Dh + 4×索引]。索引由 1 開始,所以 DS:0A521h 就是第 1 筆——這解釋了其他法術函式都在讀的那個位址。逐個扣預算不是固定選 N 個,扣不動的整個跳過(不會部分影響),而且預算歸零後仍走完整條鏈。訊息「は魅了された。」 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `262F` | sub_262F | — | 70 | 36 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/702-more-overlay22-spell-handlers.md<br>先發訊息再套效果:<sub_F06h>(DS:6F97h, 0, 1, 0, 0, 空字串 CS:262Eh);再 <overlay-23 entry#2>(17h, DS:7435h 的目標, 0(dword), 0)。⚠ sub_F06h 的第三/第四個參數與 spec 701 致傷三支不同(那裡是 0 與傷害量,這裡是 1 與 0),所以它有模式之分 | spec/702-more-overlay22-spell-handlers.md<br>spec/703-item-effect-slots-and-uninitialised-args.md |
| `2682` | sub_2682 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_2675，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `26BB` | sub_26BB | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「Knock-Knock」（unk_26AF，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `26F6` | sub_26F6 | — | 82 | 39 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/629-spell-pack-idiom-and-uninit.md<br>與 pc98 overlay-22:29A3h 助憶碼序列完全相同（39 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：v := (ROLLDICE(1,4) shl 4) or <far 014A:00D4>(DS:0A031h)——高 4 bit 放數量、低 4 bit 放另一個值,是打包不是乘法;再推入 DS:0A031h、v、三個 0,備妥「は分身した。」呼叫 <sub_F62>。1d4 個分身與 AD&D Mirror Image 一致(推論,不寫進結論)。與 3804h 用同一套打包寫法 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2754` | sub_2754 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is weakened」（unk_2748，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2799` | sub_2799 | — | 1120 | 433 | 0 | 6 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md |
| `2BF9` | sub_2BF9 | — | 433 | 148 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2DB6` | sub_2DB6 | — | 416 | 137 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `2F5E` | sub_2F5E | — | 61 | 25 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:3223h 助憶碼序列完全相同（25 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：if <far 013E:0070>(21h, DS:0A521h:0A523h) <> 0 then <far 014A:00A2>(目標,1,'の視力は戻った。')。與 41AFh 分支方向相反,交叉確定 013E:0070 是「目標身上有沒有這個效果」,21h 是失明 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2FA4` | sub_2FA4 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is blind」（unk_2F9B，長度 8）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2FD1` | sub_2FD1 | — | 165 | 61 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3076` | sub_3076 | — | 17 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call near ptr sub_2FD1`（body 共 17 bytes，已逐條讀完） | — |
| `3093` | sub_3093 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is diseased」（unk_3087，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `30C0` | sub_30C0 | — | 178 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `317E` | sub_317E | — | 932 | 349 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `352D` | sub_352D | — | 73 | 35 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:3804h 助憶碼序列完全相同（35 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：x := <far 014A:00D4>(DS:0A031h, DS:0A031h);n := (signext(DS:9594h^[198h]) shl 4) + x——把 [198h] 放高 4 bit、x 放低位打包成一個 word(不是相加),再推入 n 與三個 0、備妥 'は祈っている。' 呼叫 <sub_F62> ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `3599` | sub_3599 | — | 253 | 87 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `36A7` | sub_36A7 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「has been cursed!」（unk_3696，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `36E0` | sub_36E0 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is blinking」（unk_36D4，長度 11）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `370E` | sub_370E | — | 246 | 103 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `3804` | sub_3804 | — | 271 | 109 | 2 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/628-spell-effect-wrappers.md<br>spec/629-spell-pack-idiom-and-uninit.md<br>spec/631-strength-spell-exceptional.md |
| `391D` | sub_391D | — | 42 | 20 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:3BCBh 助憶碼序列完全相同（20 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：<sub_3AE5>(2Ah, DS:9594h^[198h], 'は加速された。')。效果 id 2Ah 與 41AFh 指同一個(加速) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `3947` | sub_3947 | — | 188 | 71 | 5 | 3 | ✓ | 待解讀 | — | — | — |
| `3A03` | sub_3A03 | — | 661 | 263 | 4 | 8 | ✓ | 待解讀 | — | — | — |
| `3C98` | sub_3C98 | — | 75 | 39 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/702-more-overlay22-spell-handlers.md<br>N d6,顆數算出來、面數寫死 6:n := <overlay-24 entry#36>(DS:6F97h);r := ROLLDICE(n, 6);<sub_3947h>(DS:7559h, DS:755Ah, r, 4, 0, @區域);<sub_3A03h>(7, r, 4, 1)。同一個 r 被兩支下游各用一次但參數位置不同(第三個 vs 第二個) | spec/702-more-overlay22-spell-handlers.md |
| `3CED` | sub_3CED | — | 46 | 21 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is Slowed」（unk_3CE3，長度 9）呼叫訊息 routine（body 共 46 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `3D27` | sub_3D27 | — | 433 | 158 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3EE2` | sub_3EE2 | — | 65 | 32 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:41AFh 助憶碼序列完全相同（32 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：if <far 013E:0070>(2Ah, DS:0A521h:0A523h) = 0 then 推入 DS:0A031h 與四個 0、備妥'はすばやくなった。' 呼叫 <sub_F62>。沒有該效果才施放,故 2Ah 是加速 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `3F23` | sub_3F23 | — | 64 | 26 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/700-pascal-parameter-order-rule.md<br>治療 opcode(2d8+1):DS:7434h 為 0 時整支什麼都不做(連訊息都沒有);否則 量 := ROLLDICE(2,8)+1,HEALDUDE(DS:7435h/7437h 組成的目標角色, 量, 0),成功才呼叫 overlay-24 entry#29 顯示 is fully/partially healed。第三個參數固定 0 = 滿血也照治,但 HEALDUDE 內部仍會因狀態不在 {0,1,4,5} 而回 false | audit/function-index/dos-overlay-22.md<br>spec/700-pascal-parameter-order-rule.md<br>spec/701-cure-cause-wounds-ladder.md |
| `3F6F` | sub_3F6F | — | 133 | 56 | 0 | 4 | ✓ | 已解讀 | strong inference | docs/spec/631-strength-spell-exceptional.md<br>與 pc98 overlay-22:423Fh 助憶碼序列完全相同（56 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：與 222Ah 同一族的力量法術入口:if <far 013E:007F>(t,0,15h,@var_1) <> 0 則備妥「は強くなった。」呼叫 <far 014A:0084>(1,0Ah);接著 n := ROLLDICE(1,4)*0Ah + 28h(擲骰乘 10 再加 40),<far 013E:0057>(t,92h,n,var_1,1),最後 <far 013E:0098>(t,0)。⚠ 92h 是外層 013E:0057 的參數不是 ROLLDICE 的 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `3FF4` | sub_3FF4 | — | 66 | 37 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/628-spell-effect-wrappers.md<br>與 pc98 overlay-22:42C4h 助憶碼序列完全相同（37 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：n := ROLLDICE(1,6) + 14h (1d6+20),接著 <sub_3BF5>(n,4,0,…) 與 <sub_3CB1>(0,4,14h,3)。整支不印任何訊息 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `4043` | sub_4043 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is paralyzed」（unk_4036，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `407A` | sub_407A | — | 78 | 33 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/701-cure-cause-wounds-ladder.md<br>治療 2d4+2:量 := ROLLDICE(2,4)+2,HEALDUDE(DS:7435h/7437h 的目標, 量, 0),成功則 <overlay-24 entry#26>(目標, 1, 'is Healed'(CS:4070h))。與 3F23h/43A7h 同族但訊息路徑不同 | audit/function-index/dos-overlay-22.md<br>spec/701-cure-cause-wounds-ladder.md<br>spec/702-more-overlay22-spell-handlers.md |
| `40D5` | sub_40D5 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_40C8，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4103` | sub_4103 | — | 59 | 31 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/701-cure-cause-wounds-ladder.md<br>致傷 2d4+2:sub_F06h(DS:6F97h, 0, 0, ROLLDAMAGEDICE(2,4)+2, 8, 空字串 CS:4102h)。與 407Ah 骰子相同、方向相反(Cause 對 Cure) | spec/700-pascal-parameter-order-rule.md<br>spec/701-cure-cause-wounds-ladder.md |
| `413F` | sub_413F | — | 57 | 31 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/701-cure-cause-wounds-ladder.md<br>致傷 2d8+1:sub_F06h(DS:6F97h, 0, 0, ROLLDAMAGEDICE(2,8)+1, 8, 空字串 CS:413Eh)。⚠ 明確推入只有 5 個 word,而 sub_F06h 是 retf 0Eh = 7 個——差的兩個是 0A54:0634h 留在堆疊上的字串結果。**這是 spec 690 那條殘留規則第一次被 retf N 獨立驗證** | spec/701-cure-cause-wounds-ladder.md |
| `4194` | sub_4194 | — | 244 | 95 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4289` | sub_4289 | — | 117 | 48 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/703-item-effect-slots-and-uninitialised-args.md<br>用哨兵值判斷下游有沒有動過:sub_F06h(DS:6F97h,0,0,0,8,空字串) → p := DS:6506h^[18Dh]^[0Ah] → **DS:6F93h := 40h(哨兵)** → <overlay-23 entry#4>(p, 9) → 若 DS:6F93h 仍是 40h 才 <overlay-23 entry#2>(0, [bp−4], [bp−2], DS:6506h, 40h)。⚠ **[bp−4] 與 [bp−2] 整支函式沒有任何一條指令寫過**(bp−9 是拼接緩衝、bp−8..bp−5 存 p),傳出去的是堆疊殘值;同一支 entry#2 在 262Fh 的同樣兩個位置推的是明確的 0。需實機驗證再決定 remake 怎麼寫 | spec/703-item-effect-slots-and-uninitialised-args.md |
| `4311` | sub_4311 | — | 150 | 64 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `43A7` | sub_43A7 | — | 66 | 26 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/700-pascal-parameter-order-rule.md<br>治療 opcode(3d8+3):與 3F23h 完全同構,只有骰子不同。配上 1d8 那支就是 AD&D 治療三階(輕傷/重傷/危傷) | audit/function-index/dos-overlay-22.md<br>spec/700-pascal-parameter-order-rule.md<br>spec/701-cure-cause-wounds-ladder.md |
| `43EA` | sub_43EA | — | 59 | 31 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/701-cure-cause-wounds-ladder.md<br>致傷 3d8+3:sub_F06h(DS:6F97h, 0, 0, ROLLDAMAGEDICE(3,8)+3, 8, 空字串 CS:43E9h)。與 43A7h 骰子相同、方向相反 | spec/701-cure-cause-wounds-ladder.md |
| `4432` | sub_4432 | — | 96 | 50 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/703-item-effect-slots-and-uninitialised-args.md<br>三段接力:r := <sub_E11h>(DS:6506h, 4, 49h) → <overlay-23 entry#21 PUTEFFECT>(r, 0,0,0,0, 空字串 CS:4425h) → <sub_F06h>(DS:6F97h, 0,0,0,0, 'is affected' CS:4426h)。⚠ CS:4425h 與 CS:4426h 緊鄰(前者長度 0、後者長度 11),反組譯裡兩個 offset 只差 1 很容易看成同一個字串的偏移 | spec/703-item-effect-slots-and-uninitialised-args.md |
| `4493` | sub_4493 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_4492，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `44D3` | sub_44D3 | — | 203 | 72 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `45B5` | sub_45B5 | — | 176 | 72 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `4672` | sub_4672 | — | 171 | 72 | 0 | 4 | ✓ | 已解讀 | strong inference | docs/spec/632-spell-target-array-loop.md<br>與 pc98 overlay-22:496Bh 助憶碼序列完全相同（72 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：bank0^[1CCh] 非 0 就直接結束;否則 for i := 1 to DS:0A520h,**逐筆檢查 nil** 後 v := <far 013E:0048>(0,4,t)、x := <sub_E75>(t,88h,88h),備妥「は絡みつかれた。」呼叫 <far 013E:0089>(v,1,0,0,x,訊息)。與 4A9Ah 相反,這支會檢查 nil——而 2776h 確實會把陣列第 1 筆清成 nil ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `472C` | sub_472C | — | 32 | 17 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is highlighted」（unk_471D，長度 14）呼叫訊息 routine（body 共 32 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4759` | sub_4759 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is invisible」（unk_474C，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4791` | sub_4791 | — | 133 | 57 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/632-spell-target-array-loop.md<br>與 pc98 overlay-22:4A9Ah 助憶碼序列完全相同（57 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：對整個目標陣列施放:訊息「は魅了された。」在迴圈外只印一次,迴圈 for i := 1 to DS:0A520h 內對每筆<far 014A:00A7>(0Bh,陣列[i],@var) 非 0 時接 <far 013E:002A>(0Bh,陣列[i],var,0)。數量為 0 時整段跳過。與 2147h 同一套效果接法,差別只在那支處理單一目標。不檢查陣列元素是否為 nil ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `4822` | sub_4822 | — | 184 | 79 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `48E4` | sub_48E4 | — | 398 | 161 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `4A8F` | sub_4A8F | — | 292 | 115 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `4BEC` | sub_4BEC | — | 416 | 194 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `4DA0` | sub_4DA0 | — | 275 | 132 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `4EB4` | sub_4EB4 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_4EB3，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `4EF7` | sub_4EF7 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「is protected」（unk_4EEA，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4F3E` | sub_4F3E | — | 1453 | 548 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `54EB` | sub_54EB | — | 164 | 72 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5590` | sub_5590 | — | 170 | 70 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `563B` | sub_563B | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_563A，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5669` | sub_5669 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5668，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5697` | sub_5697 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5696，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `56C4` | sub_56C4 | — | 189 | 75 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `578B` | sub_578B | — | 78 | 33 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/702-more-overlay22-spell-handlers.md<br>與 407Ah(spec 701)逐指令相同的第二份治療 2d4+2:ROLLDICE(2,4)+2 → HEALDUDE(目標,量,0) → 成功則 <overlay-24 entry#26>(目標, 1, 'is Healed')。⚠ 差別只有訊息字串位址(407Ah 用 CS:4070h、本支用 CS:5781h),**內容相同的一句話在同一個 overlay 存了兩份**——中文化要兩份都改,只改一份會出現時中時英且難重現 | spec/702-more-overlay22-spell-handlers.md |
| `57E3` | sub_57E3 | — | 368 | 158 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `5974` | sub_5974 | — | 294 | 114 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `5AA8` | sub_5AA8 | — | 464 | 173 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5C86` | sub_5C86 | — | 381 | 145 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `5E11` | sub_5E11 | — | 269 | 102 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `5F2F` | sub_5F2F | — | 36 | 12 | 0 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push [bp+arg_6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 36 bytes，已逐條讀完） | audit/function-index/dos-overlay-22.md |
| `5F53` | sub_5F53 | — | 185 | 87 | 4 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 5F2Fh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `6022` | sub_6022 | — | 268 | 105 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `612E` | sub_612E | — | 77 | 29 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/700-pascal-parameter-order-rule.md<br>物品類別是否落在 11..13:物品為 NIL 回 0;否則 v := byte [5CF6h + 物品^[2Eh]×16](16 bytes 一格,格內 +0),回傳 (10 < v < 14)。兩個比較都是無號(jbe/jnb),0Ah 與 0Eh 是開區間兩端,成立的只有 11/12/13。⚠ 表的基底是 DS:5CF6h 不是 5D02h(5D02h 是格內 +0Ch、5D04h 是 +0Eh),spec 691 已更正 | audit/function-index/pc98-overlay-22.md<br>spec/700-pascal-parameter-order-rule.md |
| `617B` | sub_617B | — | 121 | 42 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/703-item-effect-slots-and-uninitialised-args.md<br>從物品身上拔掉一個效果:掃 物品^[3Ch..3Eh] 三個槽(and 7Fh 後比對 arg_0,**不提早跳出,取最後一個符合的**),找到就清 0、物品^[30h] 減 1,遞減後**小於 0D2h(無號)**才呼叫 overlay-24 entry#17 把物品移除並釋放(spec 695)。那三個槽正好是 spec 696 建構子最後三個參數的落點,低 7 位是編號、最高位是旗標;+30h 是次數。⚠ 0D2h..0FFh 這段遞減後仍不觸發移除,當成「用完就沒了」會提早銷毀物品 | spec/703-item-effect-slots-and-uninitialised-args.md |
| `6209` | sub_6209 | — | 258 | 119 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `630B` | sub_630B | — | 1434 | 158 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:7332h, dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1434 bytes，已逐條讀完） | — |
| `68A5` | sub_68A5 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
