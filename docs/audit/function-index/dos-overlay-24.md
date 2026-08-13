# dos-overlay-24 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 37 | 11 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 6 個呼叫，沒有其他動作：`call loc_19A2+1`、`call loc_19E9+1`、`call loc_17E7`、`call sub_177A`、`call sub_16BD`、`call far ptr loc_1655+2`（body 共 37 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md<br>context/50-log-2026-08-09-13.md<br>spec/507-pc98-general-target-object-order-projection.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/516-fire-knife-external-map-handoff-audit.md |
| `0025` | sub_25 | — | 504 | 179 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/525-pc98-tempsearch-display-state.md<br>spec/564-ecl-operand-decoding-and-arity-validation.md<br>spec/687-far-call-flattening-and-stack-leftover.md |
| `021D` | sub_21D | — | 138 | 45 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/683-encumbrance-movement.md<br>負重決定移動力:byte[arg_0^[2Eh]×16 + 5CF6h] = 2 才生效;w := arg_0^[37h](有號),0..96h(150)用基準值 arg_4^[0E4h]、97h..18Fh(151..399)設 9、其餘設 6;之後 arg_0^[32h] 非 0 且值 <= 9 時再加 3(最多補到 12,正好是 AD&D 未負重的移動力)。**無條件覆寫** | audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md<br>spec/683-encumbrance-movement.md |
| `02A7` | sub_2A7 | — | 286 | 104 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `03C5` | sub_3C5 | — | 141 | 52 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/683-encumbrance-movement.md<br>w := arg_0^[187h] - <sub_148F>(arg_0)(32-bit 相減),負值**夾到 0**;0..200h 維持原值、201h..300h → 9、301h..400h → 6、其餘 → 3;⚠ **只在新值更小時才覆寫**(jnb 跳過寫回),所以可以在 021Dh 之後再跑一次而不會把值加回去——兩支的先後順序會影響結果 | audit/function-index/pc98-overlay-24.md<br>spec/683-encumbrance-movement.md |
| `0467` | sub_467 | — | 802 | 314 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0795` | sub_795 | — | 418 | 170 | 3 | 5 | ✓ | 待解讀 | — | — | — |
| `0939` | sub_939 | — | 136 | 57 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/684-biased-modifier-display.md<br>顯示一個偏移 60 的修正值:v := |arg_4^[19Ah] - 3Ch| 轉成字串,而負號**由獨立的一次比較**決定(arg_4^[19Ah] > 3Ch 才加 '-')。所以顯示值 = 60 - 儲存值,儲存得越大顯示越負——是「數字越小越好」的欄位用偏移量存的做法。⚠ 只看 sub + neg 那段會得到相反結論 | spec/684-biased-modifier-display.md |
| `09C1` | sub_9C1 | — | 89 | 34 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/684-biased-modifier-display.md<br>與 PC-98 0BD5h(spec 643)結構完全相同的 HP 顯示:arg_6^[1A4h] < arg_6^[78h](上限)時取一個參數、否則取另一個,arg_0 非 0 再覆蓋成第三個。⚠ 三個常數裡「受傷」不同:DOS 是 0Eh、PC-98 是 6(滿血 0Ah 與覆蓋 0Dh 兩平台相同)。兩邊的顏色編號本來就不是同一套,相同的兩個是巧合 | spec/684-biased-modifier-display.md |
| `0A34` | sub_A34 | — | 431 | 196 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0BE3` | sub_BE3 | — | 69 | 28 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/643-hp-fields-and-slot-array.md<br>與 pc98 overlay-24:0E02h 助憶碼序列完全相同（28 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：for i := 1 to 4:if <sub_25CD>(arg_0,arg_2,byte[49D3h + i],@var_6) <> 0 則 result := 1。⚠ 沒有提前結束,第一個成功後剩下三個照跑,副作用會發生四次。索引由 1 到 4 不是 0..3 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md |
| `0C28` | sub_C28 | — | 1187 | 391 | 0 | 9 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `10CB` | sub_10CB | — | 60 | 28 | 4 | 0 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:12E5h 助憶碼序列完全相同（28 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：Str 的 byte 版包裝:mov al + xor ah,ah + xor dx,dx 把輸入湊成無號 32-bit,呼叫 0A65h:12FBh(Str),再用 0A65h:649h 把結果指派給呼叫端的字串變數(上限 0FFh)。無號,0FFh 不會印成 -1 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1107` | sub_1107 | — | 58 | 27 | 1 | 0 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:1321h 助憶碼序列完全相同（27 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：Str 的 word 版包裝:mov ax + xor dx,dx(無號),其餘與 12E5h 相同 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1141` | sub_1141 | — | 57 | 25 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:135Bh 助憶碼序列完全相同（25 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：Str 的 longint 版包裝:直接推 arg_2:arg_0 整個 dword,retf 4;其餘與 12E5h/1321h 相同 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `117A` | sub_117A | — | 144 | 55 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `120A` | sub_120A | — | 144 | 55 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `129A` | sub_129A | — | 171 | 64 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1345` | sub_1345 | — | 184 | 69 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `13FD` | sub_13FD | — | 146 | 55 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `148F` | sub_148F | — | 244 | 90 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1583` | sub_1583 | — | 65 | 24 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/643-hp-fields-and-slot-array.md<br>與 pc98 overlay-24:1739h 助憶碼序列完全相同（24 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：在紀錄^[1Eh + 0..53h](84 格)裡找出第一個等於 arg_0 的格子清成 0,然後把迴圈變數設成 54h 提前結束(Pascal 沒有 break 的標準寫法)。只清第一個相符的。這 84 格與 spec 624 的 STORESPECIALS 寫入範圍(+1Fh..+6Fh,81 格)完全吻合,兩邊各自獨立量到同一個陣列 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `15FF` | sub_15FF | — | 157 | 62 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1652` | sub_1652 | — | 8 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+14Dh], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `169F` | sub_169F | — | 18 | 7 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_16E0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 41 | 15 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 41 bytes，已逐條讀完） | — |
| `16E6` | sub_16E6 | — | 148 | 54 | 0 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `177A` | sub_177A | — | 92 | 46 | 2 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `17D6` | sub_17D6 | — | 232 | 108 | 4 | 4 | ✓ | 待解讀 | — | — | — |
| `18BE` | sub_18BE | — | 50 | 26 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>依條件呼叫 loc_1ECC+1:一邊傳 (15h,...),另一邊傳 (16h,26h,12h,1) | audit/function-index/pc98-overlay-24.md<br>spec/585-ecl-goto-and-display-mode-pair.md |
| `18F3` | sub_18F3 | — | 130 | 53 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1975` | sub_1975 | — | 309 | 123 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1AAA` | sub_1AAA | — | 77 | 45 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1AF7` | sub_1AF7 | — | 1217 | 410 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `21DD` | sub_21DD | — | 441 | 158 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `2396` | sub_2396 | — | 107 | 37 | 2 | 1 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md |
| `2421` | sub_2421 | — | 254 | 80 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2543` | sub_2543 | — | 109 | 47 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `25B0` | sub_25B0 | — | 36 | 13 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:27E3h 助憶碼序列完全相同（13 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：回傳 (arg_0^[198h] = 0):是則 1、否則 0 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `25D4` | sub_25D4 | — | 84 | 29 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2628` | sub_2628 | — | 292 | 110 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `274C` | sub_274C | — | 220 | 81 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2828` | sub_2828 | — | 70 | 20 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `2877` | sub_2877 | — | 60 | 24 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:2AAEh 助憶碼序列完全相同（24 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：r := <sub_2A5B>(arg_0);arg_0^[18Eh]^[7] := 1;備妥「防御している」呼叫 <sub_194C>;回傳 r。回傳值是進來時就算好的,設旗標與顯示訊息都在那之後,所以不受旗標影響 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `28B3` | sub_28B3 | — | 438 | 149 | 0 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `2A6D` | sub_2A6D | — | 290 | 113 | 0 | 7 | ✓ | 待解讀 | — | — | audit/function-triage.md<br>spec/687-far-call-flattening-and-stack-leftover.md |
| `2BAA` | sub_2BAA | — | 603 | 248 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `2E05` | sub_2E05 | — | 54 | 22 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/634-str-conversion-family.md<br>與 pc98 overlay-24:30E7h 助憶碼序列完全相同（22 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：<far 018C:0301>() → <sub_1F6E>() → <sub_18F1>(DS:9F2Ch^[2]+3, DS:9F2Ch^[3]+3, 0FFh, 8)。兩個 byte 都用 cbw(有號)各加固定的 3 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2E6A` | sub_2E6A | — | 430 | 163 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `3018` | sub_3018 | — | 67 | 23 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `305B` | sub_305B | — | 69 | 27 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `30A0` | sub_30A0 | — | 194 | 69 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `3162` | sub_3162 | — | 327 | 110 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `32A9` | sub_32A9 | — | 173 | 58 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `3387` | sub_3387 | — | 168 | 57 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `342F` | sub_342F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
