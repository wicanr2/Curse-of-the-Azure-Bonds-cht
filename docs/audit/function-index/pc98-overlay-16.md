# pc98-overlay-16 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADLOADSAVE | 32 | 10 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call loc_1982+1`、`call loc_19C9+1`、`call sub_1627`、`call far ptr loc_147D`、`call sub_15A1`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-00.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-23.md |
| `0020` | sub_20 | — | 87 | 36 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>C 字串轉 Pascal 字串(retf 4)：0A65h:1B0Dh(Move)與 0A65h:649h(StoreString)，其餘同 DOS overlay-16:0020h，含同一個超出參數區的目的指標 | audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-02.md<br>audit/function-index/pc98-overlay-15.md<br>audit/function-index/pc98-overlay-16.md |
| `0079` | sub_79 | — | 214 | 84 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0164` | sub_164 | — | 1176 | 456 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-25.md<br>audit/function-strings.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `0614` | sub_614 | LOADCHARLIST | 450 | 204 | 0 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `07DC` | sub_7DC | — | 93 | 40 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `08E4` | sub_8E4 | — | 453 | 198 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-33.md<br>audit/function-strings.md |
| `0B4F` | sub_B4F | CHECKSAVEDRIVE | 920 | 370 | 6 | 4 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0ECE` | sub_ECE | — | 24 | 9 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 24 bytes，已逐條讀完） | — |
| `0ED3` | sub_ED3 | — | 46 | 18 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_F0E`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 46 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-strings.md |
| `0F17` | sub_F17 | MAKEPLAYERICON | 241 | 87 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0FF7` | sub_FF7 | — | 94 | 33 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 94 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1061` | sub_1061 | LOADACTIVEICON | 498 | 183 | 1 | 4 | ✓ | 已解讀 | exact | 1<br>同 DOS overlay-16:00A7Eh（module_align 對齊，助憶碼序列完全相同）。差異全是位址：DS:9594h↔6506h、DS:0D91h↔0515h、DS:96C4h/96C6h↔662Eh/6630h、[di−6992h]↔[di+65D8h] 與呼叫目標；'CHEAD'／'CBODY' 兩平台都沒翻譯 | audit/embedded-strings.md<br>audit/function-strings.md |
| `1261` | sub_1261 | DELETECHARACTER | 289 | 117 | 2 | 0 | ✓ | 已解讀 | exact | 868<br>同 DOS overlay-16:00C7Eh 的功能，但★整段重做三次而且每個副檔名配一個不同的前綴常數(DS:8BF6h ＋ loc_A2Ah+3 / loc_A37h / loc_A28h)——PC-98 把三個檔放在三個不同位置，DOS 三個同目錄。真的邏輯差異(spec 783 第 d 類)，117 對 63 條 | audit/function-strings.md<br>spec/868-character-file-triplet.md |
| `1437` | sub_1437 | — | 70 | 23 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add ah, 82h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 70 bytes，已逐條讀完） | — |
| `14C9` | sub_14C9 | SAVECHARACTER | 26 | 12 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:649h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | audit/function-index/pc98-overlay-16.md<br>audit/function-strings.md |
| `15A1` | sub_15A1 | — | 90 | 39 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 90 bytes，已逐條讀完） | — |
| `15F5` | sub_15F5 | — | 50 | 22 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 50 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1627` | sub_1627 | — | 1221 | 480 | 2 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 14C9h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-triage.md |
| `166F` | sub_166F | — | 10 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 5 | 3 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `167E` | sub_167E | — | 11 | 3 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 723h:48Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1B0D` | sub_1B0D | — | 264 | 109 | 1 | 1 | ✓ | 已解讀 | exact | 1<br>同 DOS overlay-16:01388h（module_align 對齊，助憶碼序列完全相同）。只差三條：DS:8BF6h↔5BF0h 與兩個 CS 常數／字串函式位址 | audit/function-index/pc98-overlay-16.md |
| `1C15` | sub_1C15 | — | 1152 | 449 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-16.md |
| `2095` | sub_2095 | — | 374 | 112 | 3 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1C15h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `26CF` | sub_26CF | — | 798 | 241 | 1 | 5 | ✓ | 已解讀 | exact | 925<br>同 DOS overlay-16:01F21h(相似度 0.922)。五個差異區塊全是同一件事：取角色遠指標時多一次 les di, es:[di]——ss:[父bp+0Ah] 存的是指向角色遠指標的遠指標；模板的父層位移 −1C4h → −1ECh。沒有任何邏輯差異 | spec/925-npc-template-floor.md |
| `2A6D` | sub_2A6D | LOADCHARACTER | 3141 | 1140 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-16.md<br>audit/function-strings.md<br>audit/function-triage.md |
| `36B2` | sub_36B2 | — | 953 | 346 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 2A6Dh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `3AA8` | sub_3AA8 | LOADMONSTER | 850 | 324 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `3DFF` | sub_3DFF | LOADNPC | 108 | 42 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/760-item-effect-flag-clear-and-file-exists.md<br>載入 CPIC(retf 2)：與 DOS overlay-16:351Ch(spec 756)同流程(DS:7F09h^[67Ch] > 7 就不做、GetMem 後填 +126h、組 CS:3DFAh 的 'CPIC'、最後叫 (參數, p^[143h]))。⚠ 配置大小 DOS 1A6h、PC-98 1A7h，多一個 byte | spec/760-item-effect-flag-clear-and-file-exists.md |
| `3E8B` | sub_3E8B | ATTACHCHARACTER | 300 | 98 | 2 | 2 | ✓ | 已解讀 | exact | 873<br>同 DOS overlay-16:035A8h，44 條差異全是位址(DS:650Ah↔9598h、+189h↔+18Ah、DS:4F9Dh↔7F09h)。★PC-98 側在 DOS 出現 hlt 的位置是正確的 lea di,[bp+var_C]，用來校正 DOS 的反組譯錯位 | spec/873-add-to-party-and-portrait-slot.md |
| `3FB7` | sub_3FB7 | — | 115 | 53 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `402A` | sub_402A | — | 303 | 120 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/780-submap-blit-entry-init-and-weighted-score.md<br>清單項目初始化(retf 4)：記錄陣列 DS:0A648h、每筆 29h(41) bytes、索引取自 參數^[0] — 長度 byte 設 14h 並用 ' ' 填滿 20 格、叫本模組 45C9h 後 StoreString(28h)、+09h 填 87h、+0Ah 填 i+3Fh；byte[0A80Ah + i] 不等於 5 時，從 DS:6812h 起每筆 15h bytes 的表搬一段到 +0Dh(長度取自該筆第一個 byte)。⚠ 20 個空白是寫死的欄寬 | spec/780-submap-blit-entry-init-and-weighted-score.md |
| `4193` | sub_4193 | — | 183 | 77 | 2 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `424A` | sub_424A | — | 51 | 27 | 2 | 0 | ✓ | 已解讀 | exact | docs/spec/636-pc98-text-vram-and-disk-bios.md<br>FillChar(0A000h:0A00h,500h,0) 與 FillChar(0A200h:0A00h,500h,1)——PC-98 文字 VRAM 的字元碼平面與屬性平面。位移 0A00h ÷2 = 第 1280 格 = 第 16 列(80 欄),長度 500h ÷2 = 640 格 = 8 列,所以清的是第 16..23 列的訊息區。⚠ 中文化不能沿用這條路徑:漢字 ROM 沒有繁體字集 | spec/636-pc98-text-vram-and-disk-bios.md |
| `42FE` | sub_42FE | LOADSAVEDGAME | 772 | 317 | 0 | 12 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `46CE` | sub_46CE | SAVECURRENTGAME | 659 | 279 | 0 | 11 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `4961` | sub_4961 | LOADALL | 17 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_49C6`（body 共 17 bytes，已逐條讀完） | — |
| `49C6` | sub_49C6 | LOADGAME | 1183 | 458 | 2 | 9 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `4E65` | sub_4E65 | SAVEALL | 28 | 17 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/636-pc98-text-vram-and-disk-bios.md<br>repeat until <sub_B4F>(0) <> 0 的純忙碌等待(jz 跳回迴圈開頭 4E68h,每輪重新推入參數 0,中間沒有延遲也沒有讓出),之後 <sub_5008>(0,4Bh) | spec/636-pc98-text-vram-and-disk-bios.md |
| `4E81` | sub_4E81 | — | 23 | 9 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp+var_4], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | — |
| `5008` | sub_5008 | SAVEGAME | 718 | 296 | 2 | 4 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `5581` | sub_5581 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/pc98-overlay-02.md<br>spec/612-ecl-main-loop.md |
| `5588` | sub_5588 | — | 236 | 90 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `5674` | sub_5674 | — | 196 | 93 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `5738` | sub_5738 | — | 23 | 12 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/636-pc98-text-vram-and-disk-bios.md<br>呼叫 574Fh 算好磁碟機碼後,al := CS:576Eh、ah := 4、int 1Bh(PC-98 磁碟 BIOS);沒進位回 1、有進位回 0 | audit/function-index/pc98-overlay-16.md<br>spec/636-pc98-text-vram-and-disk-bios.md |
| `574F` | sub_574F | — | 31 | 18 | 3 | 0 |  | 已解讀 | exact | docs/spec/636-pc98-text-vram-and-disk-bios.md<br>al := byte at 0C29h:8BF7h,減 1 再 and 0Fh 再 or 90h,存進 CS:576Eh。90h 是磁碟機類別碼、低 4 bit 是機號,所以 0C29h:8BF7h 存的是 1 起算的機號。與 5738h 靠程式碼段內的變數溝通而不是傳參數 | audit/function-index/pc98-overlay-16.md<br>spec/636-pc98-text-vram-and-disk-bios.md |
