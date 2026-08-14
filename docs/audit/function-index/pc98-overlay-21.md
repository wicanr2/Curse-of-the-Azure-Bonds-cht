# pc98-overlay-21 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADMONEY | 23 | 6 | 0 | 4 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_19CA`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `001B` | sub_1B | MAXIUMWEIGHT | 32 | 12 | 3 | 1 | ✓ | 已解讀 | exact | docs/spec/637-overlay21-small-batch.md<br>回傳 <far 0176:0A31>(arg_0,arg_2) + 5DCh(1500) | audit/function-index/dos-overlay-21.md<br>spec/637-overlay21-small-batch.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `003B` | sub_3B | LOSEWEIGHT | 28 | 10 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/637-overlay21-small-batch.md<br>arg_2^[188h] -= arg_0(word)。與 0057h 逐指令相同只差 sub/add,兩支都不做上下限檢查——減到負數會回捲 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-21.md<br>audit/function-index/pc98-overlay-21.md<br>spec/637-overlay21-small-batch.md<br>spec/764-fsplit-dbcs-and-eight-slot-longint-table.md |
| `0057` | sub_57 | GAINWEIGHT | 28 | 10 | 4 | 0 | ✓ | 已解讀 | exact | docs/spec/637-overlay21-small-batch.md<br>arg_2^[188h] += arg_0(word)。與 003Bh 成對 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-21.md<br>audit/function-index/pc98-overlay-04.md<br>audit/function-index/pc98-overlay-21.md<br>spec/637-overlay21-small-batch.md |
| `0073` | sub_73 | TOOHEAVY | 84 | 31 | 3 | 2 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>會不會超過上限(retf 0Ah)：欄位是 +188h，其餘與 DOS overlay-21:0073h 相同 | audit/function-index/pc98-overlay-21.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/762-ega-glyph-blit-and-movement-rate.md |
| `00C7` | sub_C7 | CASHPOOL | 129 | 51 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0148` | sub_148 | CHANGECHARMONEY | 85 | 31 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/622-character-money-block.md<br>把 DS:9594h 指向的紀錄設成 arg_0 金額:先用 i=0..4 迴圈把 [di+0FBh+2i] 五個硬幣欄位歸零,再 div 5——商存 +103h、餘存 +101h。這證明 +FBh 起是五元素硬幣陣列(不是五個獨立欄位),且依 AD&D 的 5 gp = 1 pp 定出 +103h 是白金、+101h 是金幣,陣列由低價值排到高價值 | audit/function-index/dos-overlay-21.md<br>spec/622-character-money-block.md |
| `019D` | sub_19D | CHANGEPOOLMONEY | 86 | 34 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>清表並拆商餘(retf 2)：表在 DS:0A00Ah，商寫 DS:0A01Ah、餘寫 DS:0A016h，與 DOS overlay-21:019Eh 同 | audit/embedded-strings.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `0218` | sub_218 | ADDPLATINUM | 145 | 53 | 0 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `02AB` | sub_2AB | GETMONEYINPUT | 461 | 203 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0485` | sub_485 | GIVEMONEY | 188 | 71 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0541` | sub_541 | POOLMONEY | 210 | 79 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0613` | sub_613 | NUMOFPCS | 86 | 30 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>數隊伍(retf)：鏈頭 DS:9598h、next +18Ah，其餘同 DOS overlay-21:05F9h | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0669` | sub_669 | SHAREPOOL | 948 | 382 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0A1D` | sub_A1D | DROPCASH | 124 | 50 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/764-fsplit-dbcs-and-eight-slot-longint-table.md<br>扣一筆並累計(retf 8)：與 DOS overlay-21:0A03h 助憶碼 50 條完全相同，表在 DS:0A00Ah、狀態變數 DS:7F27h | audit/embedded-strings.md<br>spec/764-fsplit-dbcs-and-eight-slot-longint-table.md |
| `0AA6` | sub_AA6 | — | 206 | 87 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B9B` | sub_B9B | GETMONEYTYPE | 520 | 219 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0DE8` | sub_DE8 | TAKEMONEY | 12 | 4 | 0 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_19E0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `1070` | sub_1070 | CHECKTREASURE | 84 | 30 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>兩個旗標的輸出(retf 8)：表在 DS:0A00Ah，第 8 格 DS:0A026h，與 DOS overlay-21:0F5Bh 同 | audit/embedded-strings.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `10C4` | sub_10C4 | — | 56 | 25 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/637-overlay21-small-batch.md<br>r := ROLLDICE(1,14h)(1d20);1..0Eh 回傳 1、0Fh..14h 回傳 2,即 70%/30%。兩個區間剛好把 1..20 分完(無縫隙無重疊),所以回傳值雖無預設值也不會讀到未初始化。比較用無號 jb/ja | spec/637-overlay21-small-batch.md |
| `10FC` | sub_10FC | CREATERNDTREASURE | 1455 | 538 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-21.md<br>audit/function-triage.md |
| `147D` | sub_147D | — | 143 | 51 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 10FCh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `151F` | sub_151F | — | 15 | 5 | 5 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp al, 37h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `152E` | sub_152E | — | 115 | 38 | 1 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 10FCh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `15A1` | sub_15A1 | — | 382 | 141 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 10FCh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1683` | sub_1683 | — | 10 | 3 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1695`，控制權轉交後不返回；先設定 `les di, [bp-9]`、`mov byte ptr es:[di+59h], 0D1h`（body 共 10 bytes，已逐條讀完） | — |
| `19CA` | sub_19CA | — | 12 | 3 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp near ptr 39F6h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | audit/function-triage.md |
| `19E0` | sub_19E0 | — | 31 | 14 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp near ptr 5A80h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1A3E` | sub_1A3E | APPRAISE | 2263 | 151 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2315` | sub_2315 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
