# pc98-overlay-10 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCOMPREP | 72 | 18 | 0 | 9 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 13 個呼叫，沒有其他動作：`call far ptr unk_101A`、`call far ptr loc_1882+1`、`call far ptr loc_1815+2`、`call loc_1922+1`、`call loc_E26`、`call far ptr loc_1694+3`（body 共 72 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/pc98-overlay-23.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md |
| `0048` | sub_48 | — | 132 | 55 | 4 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/771-extra-attacks-and-weapon-class-whitelist.md |
| `00CC` | sub_CC | — | 554 | 208 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/776-first-person-view-scan-and-ega-rect-fill.md |
| `02F6` | sub_2F6 | — | 130 | 49 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/784-cross-platform-pairs-second-batch.md<br>巢狀的牆面查詢(retf 8)：與 DOS overlay-10:02F6h 49 條同形，隊伍 y 在 DS:0A2AAh(差異 3 條，已逐條列出) | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md<br>spec/784-cross-platform-pairs-second-batch.md |
| `0378` | sub_378 | — | 116 | 54 | 5 | 1 | ✓ | 已解讀 | exact | docs/spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md<br>牆面從兩側各查一次(retf 6)：54 條指令與 DOS overlay-10:0378h 只差 dx/dy 表位址(489Eh/48A7h)。更正：本支的宣告順序是 (x, y, 方向)，先前寫成推入順序 | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md<br>spec/776-first-person-view-scan-and-ega-rect-fill.md<br>spec/784-cross-platform-pairs-second-batch.md |
| `03EC` | sub_3EC | — | 165 | 78 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/776-first-person-view-scan-and-ega-rect-fill.md |
| `0491` | sub_491 | — | 125 | 73 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/771-extra-attacks-and-weapon-class-whitelist.md<br>與 DOS overlay-10:0491h（entry#9）助憶碼序列完全相同，語意同該筆：八組寫死的參數(retf 2)：DS:4A0Eh = 1 時叫 本模組 0048h 四次 —(3,0,5)、(4,0,5)、(3,1,0Ah)、(4,1,0Ah)；否則四組第三欄全用 16h。原始碼是八行手寫呼叫不是迴圈，這 24 個數字要當資料照抄 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/771-extra-attacks-and-weapon-class-whitelist.md<br>spec/776-first-person-view-scan-and-ega-rect-fill.md |
| `050E` | sub_50E | — | 472 | 196 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-10.md<br>audit/function-index/pc98-overlay-10.md<br>spec/776-first-person-view-scan-and-ega-rect-fill.md |
| `06E6` | sub_6E6 | — | 471 | 185 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `08BD` | sub_8BD | — | 196 | 89 | 1 | 8 | ✓ | 已解讀 | strong inference | docs/spec/776-first-person-view-scan-and-ega-rect-fill.md<br>與 DOS overlay-10:08C9h（entry#7）助憶碼序列完全相同，語意同該筆：第一人稱視野掃描(retf，無參數)：兩層迴圈以全域 DS:4A0Ch(dx，−6..+6)與 DS:4A0Dh(dy，−2..+2)掃 13×5 = 65 格；每格 x := DS:720Fh + dx、y := DS:7210h + dy，對四個方向(6 西/0 北/2 東/4 南)各叫 本模組 0378h 取牆旗標，分別存進 DS:4A0Eh(北)/4A0Fh(西)/4A10h(東)/4A11h(南)；接著四支巢狀繪圖程序(03ECh/0491h/050Eh/06ECh，都以 push bp 傳靜態鏈)，再 DS:4A15h := <呼叫>(x,y) and 40h、本模組 00CCh()。⚠ 迴圈計數器是全域，不可重入。spec 771 的 0491h 判的 DS:4A0Eh 就是北面牆旗標 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md |
| `0981` | sub_981 | — | 27 | 12 | 3 | 0 | ✓ | 已解讀 | strong inference | docs/spec/572-resident-service-functions.md<br>與 dos overlay-10:098Dh 助憶碼序列完全相同（12 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：表格查詢:以 ds:4A14h 的值為索引,回傳 DS:352h 起的表中該項 byte ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `099C` | sub_99C | — | 86 | 36 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>在地圖上補兩格(retf 6)：與 DOS overlay-10:09A8h 同 | spec/758-morale-field-0f7h-round-trip.md |
| `09F2` | sub_9F2 | — | 268 | 115 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0AFE` | sub_AFE | — | 326 | 138 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0C4A` | sub_C4A | — | 75 | 31 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 75 bytes，已逐條讀完） | — |
| `0C95` | sub_C95 | — | 424 | 182 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0E3D` | sub_E3D | — | 382 | 183 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0FBB` | sub_FBB | — | 80 | 30 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>清空戰場地圖(retf)：FillChar((DS:9F2Ch+7)^, 4E2h, 17h)，DS:7F05h^[342h] → DS:7ACAh，與 DOS overlay-10:0FC7h 同 | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `1023` | sub_1023 | — | 174 | 72 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `10D1` | sub_10D1 | — | 286 | 94 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `11EF` | sub_11EF | — | 49 | 18 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>判斷式(retf 4)：與 DOS overlay-10:11FBh 相同的「兩個都不在範圍內才回 true」 | spec/754-small-predicates-and-wrappers.md |
| `1220` | sub_1220 | — | 324 | 138 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1364` | sub_1364 | — | 902 | 365 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-10.md |
| `142D` | sub_142D | — | 81 | 36 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-16h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `14E3` | sub_14E3 | — | 38 | 16 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1587`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 38 bytes，已逐條讀完） | audit/embedded-strings.md |
| `167E` | sub_167E | — | 316 | 129 | 2 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1364h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `17B7` | sub_17B7 | — | 55 | 21 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 55 bytes，已逐條讀完） | — |
| `17EE` | sub_17EE | — | 6 | 3 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 10h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | audit/function-index/pc98-overlay-10.md |
| `17F4` | sub_17F4 | — | 5 | 1 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_155A+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `17F9` | sub_17F9 | — | 8 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr [bp-0Eh], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `18FB` | sub_18FB | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `1900` | sub_1900 | — | 756 | 277 | 1 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 17EEh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `196A` | sub_196A | — | 26 | 11 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:7AC8h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | — |
| `1C28` | sub_1C28 | INITCOMBAT | 239 | 86 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `1DAA` | sub_1DAA | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
