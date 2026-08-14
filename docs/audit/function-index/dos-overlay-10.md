# dos-overlay-10 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 72 | 18 | 0 | 10 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 13 個呼叫，沒有其他動作：`call loc_104A`、`call far ptr loc_18B1+2`、`call far ptr sub_1847`、`call far ptr sub_1953`、`call loc_EB5+1`、`call far ptr sub_16BD`（body 共 72 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/pc98-overlay-23.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md |
| `0048` | sub_48 | — | 132 | 55 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `00CC` | sub_CC | — | 554 | 208 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02F6` | sub_2F6 | — | 130 | 49 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0378` | sub_378 | — | 116 | 54 | 5 | 1 | ✓ | 待解讀 | — | — | — |
| `03EC` | sub_3EC | — | 165 | 78 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0491` | sub_491 | — | 125 | 73 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `050E` | sub_50E | — | 478 | 197 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `06EC` | sub_6EC | — | 477 | 186 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `08C9` | sub_8C9 | — | 196 | 89 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `098D` | sub_98D | — | 27 | 12 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>表格查詢:以 ds:4A14h 的值為索引,回傳 DS:352h 起的表中該項 byte | audit/function-index/pc98-overlay-10.md<br>spec/572-resident-service-functions.md |
| `09A8` | sub_9A8 | — | 86 | 36 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `09FE` | sub_9FE | — | 268 | 115 | 1 | 3 | ✓ | 待解讀 | — | — | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0B0A` | sub_B0A | — | 407 | 171 | 1 | 2 | ✓ | 待解讀 | — | — | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0CA1` | sub_CA1 | — | 12 | 7 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | audit/function-index/dos-overlay-10.md |
| `0CAD` | sub_CAD | — | 395 | 169 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0CA1h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0E49` | sub_E49 | — | 382 | 183 | 1 | 3 | ✓ | 待解讀 | — | — | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0FC7` | sub_FC7 | — | 80 | 30 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>清空戰場地圖(retf，無參數)：FillChar((DS:6E92h+7)^, 4E2h, 17h) — 4E2h=1250=50×25，獨立確認 spec 749 由 FreeMem 4E9h 反推的格陣尺寸與 7 byte 表頭；再抄 DS:4F99h^[342h] 到 DS:4A14h(另兩個 local 指派後沒用)，最後叫本模組 09FEh/0B0Ah/0E49h | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `102F` | sub_102F | — | 174 | 72 | 1 | 4 | ✓ | 待解讀 | — | — | spec/750-combat-setup.md |
| `10DD` | sub_10DD | — | 286 | 94 | 1 | 2 | ✓ | 待解讀 | — | — | spec/750-combat-setup.md |
| `11FB` | sub_11FB | — | 49 | 18 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>判斷式(retf 4)：if (a>=0) and (a<11) then false else if (b>=0) and (b<6) then false else true。⚠ 是「兩個都不在範圍內才回 true」，不是一般的界內檢查；兩個範圍不同(0..10 與 0..5) | audit/function-index/pc98-overlay-10.md<br>spec/754-small-predicates-and-wrappers.md |
| `122C` | sub_122C | — | 324 | 138 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-23.md |
| `1370` | sub_1370 | — | 519 | 212 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-10.md |
| `15D1` | sub_15D1 | — | 211 | 86 | 2 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1370h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `16A4` | sub_16A4 | — | 25 | 8 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr [bp-7], 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 268 | 110 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1370h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `17C9` | sub_17C9 | — | 30 | 12 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_13A4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `17E7` | sub_17E7 | — | 25 | 9 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 25 bytes，已逐條讀完） | — |
| `1800` | sub_1800 | — | 37 | 13 | 1 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `shl di, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 37 bytes，已逐條讀完） | audit/function-index/dos-overlay-10.md<br>spec/750-combat-setup.md |
| `1847` | sub_1847 | — | 109 | 44 | 2 | 0 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1800h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `192B` | sub_192B | — | 2 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short sub_1930`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `1930` | sub_1930 | — | 15 | 6 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `shl ax, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `193F` | sub_193F | — | 20 | 8 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [di+30Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 20 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1953` | sub_1953 | — | 55 | 24 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-0Ch]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 55 bytes，已逐條讀完） | spec/687-far-call-flattening-and-stack-leftover.md |
| `198A` | sub_198A | — | 25 | 12 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | — |
| `19A3` | sub_19A3 | — | 73 | 30 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr [bp-2], 5`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 73 bytes，已逐條讀完） | — |
| `1C3E` | sub_1C3E | — | 371 | 130 | 0 | 10 | ✓ | 已解讀 | exact | docs/spec/750-combat-setup.md<br>戰鬥開場(retf，不收參數)：畫出 'A battle begins...'(CS:1C2Bh)；把兩條雲霧鏈頭、DS:650Eh、DS:6E91h 歸零並 FillChar(@DS:6E59h,38h,0)——起點/長度/終點三數咬合，可證戰鬥員陣列上限 8 筆(每筆 7 bytes，計數緊接其後)；寫地圖表頭 +2/+3 = overlay-32 entry#15/#16(隊伍鏈頭) − 3；對每個隊員叫兩次 overlay-23 entry#4(角色, 8) 與 (角色, 16h)；結尾 DS:4FBAh := 5。三支本模組 retf 程序以 push cs + call near 呼叫 | spec/748-cloud-effect-dispel-pair.md<br>spec/749-combat-teardown-and-battlefield-grid.md<br>spec/750-combat-setup.md |
| `1DB1` | sub_1DB1 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
