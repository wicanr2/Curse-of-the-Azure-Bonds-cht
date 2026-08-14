# pc98-overlay-25 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTRAINING | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 19Ah:2Ah 與 164h:57h | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-25.md |
| `0011` | sub_11 | — | 928 | 342 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-25.md<br>audit/function-index/pc98-overlay-12.md |
| `03B1` | sub_3B1 | FIGLEVELSTUFF | 704 | 247 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0671` | sub_671 | ADJUSTCLERICALSPELLS | 370 | 121 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/809-cleric-spell-progression-table.md<br>牧師系可記憶法術數重算：retf 6，宣告順序 (角色, 要不要重算)。等級 := 角色^[109h] + 角色^[111h] * 本模組 14C4h(角色)，有號 <= 0 就離開。要不要重算 <> 0 時：for j := 2 to 5 清 角色^[12Ch+j]、角色^[12Dh] := 1、for L := 2 to 等級 do for j := 1 to 5 do 角色^[12Ch+j] += byte[734Fh + L*5 + j]（進階表每等級 5 bytes 的增量，兩平台逐位元組相同，累計結果＝AD&D 一版牧師表；L >= 13 整列為 FFh 且迴圈無上界檢查）。接著六條智慧加成（角色^[15h] > 0Ch/0Dh → +12Dh、> 0Eh/0Fh → +12Eh、> 10h → +12Fh、> 11h → +130h，各自要求該等級目前 > 0）——要不要重算 = 0 時仍照跑，連呼兩次會累積。+12Ch 本身從未被碰，職業 0（牧師系）的區塊是 +12Dh..+131h。spec 809 | audit/embedded-strings.md<br>spec/809-cleric-spell-progression-table.md |
| `07E3` | sub_7E3 | SETSAVETHROWS | 737 | 282 | 2 | 1 | ✓ | 待解讀 | — | — | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0AC4` | sub_AC4 | SETTHIEFSKILLS | 595 | 224 | 2 | 2 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md |
| `0D17` | sub_D17 | — | 370 | 145 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-26.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `0F30` | sub_F30 | CHANGEHUMANCLASS | 1130 | 437 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `139A` | sub_139A | OLDHUMANCLASS | 92 | 33 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>第一個非零職業欄位的索引(retf 4)：掃 p^[111h..117h]，與 DOS overlay-25:1292h 同 | audit/function-index/pc98-overlay-25.md<br>spec/753-small-utility-routines.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `13F6` | sub_13F6 | CURHUMANCLASS | 92 | 33 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>第一個非零職業欄位的索引(retf 4)：掃 p^[109h..10Fh]，與 DOS overlay-25:12EEh 同 | spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `1452` | sub_1452 | CURHUMANLEVEL | 80 | 29 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>第一個非零的職業欄位(retf 4)：與 DOS overlay-25:134Ah 同，含同一個越界一格的行為 | audit/embedded-strings.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `14A2` | sub_14A2 | ISHUMAN | 34 | 14 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>判斷式(retf 4，一個遠指標參數)：回傳 p^[74h] = 7，與 DOS overlay-25:139Ah 同 | spec/753-small-utility-routines.md |
| `14C4` | sub_14C4 | OLDCLASSOK | 44 | 18 | 4 | 2 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>判斷式(retf 4)：回傳 本模組 sub_1452h(p) > p^[0E6h]，與 DOS overlay-25:13BCh 同 | spec/754-small-predicates-and-wrappers.md |
| `14F0` | sub_14F0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
