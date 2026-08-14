# pc98-overlay-15 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCAMP | 52 | 14 | 0 | 7 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 9 個呼叫，沒有其他動作：`call loc_19C8+2`、`call sub_11C7`、`call far ptr sub_1084`、`call far ptr loc_159F+2`、`call far ptr loc_E22+4`、`call sub_1697`（body 共 52 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/dos-overlay-28.md |
| `0034` | sub_34 | — | 340 | 121 | 1 | 2 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-index/dos-overlay-14.md<br>audit/function-index/dos-overlay-19.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/dos-overlay-23.md |
| `0188` | sub_188 | — | 67 | 24 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>清掉超界值(retf 4)：與 DOS overlay-15:0188h 位元組相同 | audit/function-index/dos-overlay-14.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/788-memorised-spell-slots-identified.md<br>spec/792-memorization-timer-closed.md |
| `01CB` | sub_1CB | — | 115 | 42 | 3 | 2 | ✓ | 已解讀 | exact | docs/spec/760-item-effect-flag-clear-and-file-exists.md<br>清掉物品三個效果槽的旗標位元(retf 4)：鏈頭 +14Eh、槽在 +64h/+65h/+66h、next 在 +52h — PC-98 物品節點配置與 DOS 不同，其餘同 DOS overlay-15:01CBh | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `023E` | sub_23E | — | 68 | 27 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>整隊清理(retf)：鏈頭 DS:9598h、next 在 +18Ah，其餘同 DOS overlay-15:023Eh | audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `0282` | sub_282 | — | 165 | 63 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/788-memorised-spell-slots-identified.md<br>回傳某職業某等級「還能再記幾個法術」：掃角色 +1Eh..+71h 共 84 格，值 > 0 者取低 7 位當法術編號（最高位＝還在記憶中，見 spec 792），查法術屬性表 DS:61B4h + s*10h 的 +1(等級) 與 +0(職業)，相符就計數（不分最高位）；結果 := 角色^[12Ch + 職業*5 + 等級] − 已記數。retf 4。spec 788／792（PC-98：角色表 DS:9594h） | spec/788-memorised-spell-slots-identified.md |
| `0386` | sub_386 | — | 285 | 114 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `04C2` | sub_4C2 | DOCASTSPELL | 182 | 77 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `05CC` | sub_5CC | — | 488 | 207 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `07B4` | sub_7B4 | — | 170 | 66 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `085E` | sub_85E | — | 155 | 57 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md<br>與 DOS overlay-15:085Ch（entry#12）助憶碼序列完全相同，語意同該筆：把 84 格陣列排序(retf，無參數)：對 DS:6506h(目前角色)的 +1Eh 起 84 格做選擇排序，比較時 and 7Fh、交換時整個 byte 一起搬(高位旗標跟著值走)，遞增。固定跑 84×85/2 次比較，沒有提早結束 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `092D` | sub_92D | — | 444 | 181 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `0B86` | sub_B86 | — | 673 | 255 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `0E54` | sub_E54 | — | 131 | 49 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `0FC7` | sub_FC7 | — | 21 | 9 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 81Fh:43h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 21 bytes，已逐條讀完） | audit/function-index/pc98-overlay-15.md |
| `0FF2` | sub_FF2 | — | 10 | 6 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, 50h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | audit/embedded-strings.md |
| `0FFC` | sub_FFC | — | 31 | 10 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-0Eh], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1084` | sub_1084 | — | 306 | 134 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FC7h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/embedded-strings.md<br>audit/string-pairs.md |
| `11A9` | sub_11A9 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `11AE` | sub_11AE | — | 12 | 4 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_133F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `11C7` | sub_11C7 | — | 765 | 347 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FC7h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `147B` | sub_147B | — | 85 | 37 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-15.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `14CA` | sub_14CA | — | 119 | 53 | 8 | 9 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 147Bh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1524` | sub_1524 | — | 2 | 1 | 9 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_153D`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `1558` | sub_1558 | — | 35 | 13 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp dx, ds:9596h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 35 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 7 | 2 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:959Ah, dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1688` | sub_1688 | — | 13 | 3 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+18Ch], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | — |
| `1697` | sub_1697 | — | 2 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 2 bytes，已逐條讀完） | — |
| `1699` | sub_1699 | — | 251 | 79 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1794` | sub_1794 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+18Ah], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1799` | sub_1799 | — | 9 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 9 bytes，已逐條讀完） | — |
| `17F6` | sub_17F6 | — | 296 | 124 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `19A4` | sub_19A4 | — | 404 | 160 | 1 | 6 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1B94` | sub_1B94 | — | 184 | 71 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1D4C` | sub_1D4C | — | 506 | 204 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `1F46` | sub_1F46 | — | 35 | 14 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/637-overlay21-small-batch.md<br>回傳 (arg_2^[196h] = 0):是則 1、否則 0。+196h 是狀態碼(spec 623 記到 6 與 5→4 兩條) | audit/function-index/dos-overlay-15.md |
| `1F69` | sub_1F69 | — | 183 | 73 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-15.md |
| `2020` | sub_2020 | — | 117 | 46 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/763-dungeon-map-second-plane-and-stone-to-flesh.md<br>三級治療量累加(retf 2，SS 相對記錄)：ss:[p−4] 次 1d8、ss:[p−6] 次 2d8+1、ss:[p−8] 次 3d8+3，總量加到 ss:[p−2]。沒有任何上限檢查(理論上限 13260，word 裝得下)。⚠ IDA 把同一個位移 8b7e06 印成 arg_2 與 [bp+6] 兩種名字，判參數要看位移位元組 | audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/763-dungeon-map-second-plane-and-stone-to-flesh.md<br>spec/770-party-step-and-hug-attack.md |
| `2095` | sub_2095 | — | 61 | 24 | 3 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 61 bytes，已逐條讀完） | — |
| `20D2` | sub_20D2 | — | 90 | 33 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>全隊受傷總量(retf 2)：目前 HP 在 +1A5h，其餘同 DOS overlay-15:206Bh | audit/function-index/pc98-overlay-15.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `212C` | sub_212C | — | 400 | 143 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-15.md |
| `22BC` | sub_22BC | — | 180 | 63 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/789-heal-budget-tick-and-halve-stack.md<br>把一筆治療額度沿 DS:9598h 角色鏈由前往後分配：缺 := +78h(最大HP) − +1A5h(目前HP)，上限為呼叫端框架的 ss:[rec-2] 剩餘額度；呼叫 <far unk_146E>(角色, byte(缺), 0) 成功才扣額度。順序決定誰拿得到，額度用完的成員拿不到。retf 2。spec 789（PC-98 欄位 +1） | audit/function-index/pc98-overlay-15.md |
| `2370` | sub_2370 | — | 164 | 68 | 1 | 8 | ✓ | 已解讀 | exact | docs/spec/792-memorization-timer-closed.md<br>巢狀程序組成的流程（每個呼叫多推 bp 當靜態鏈）：輸出^:=0；20D2h 回 0 就離開；呼叫 1F69h；再問 20D2h，回 0 走 <14F8h+2>(DS:9594h)+<15ADh+1>；否則 Move(DS:0A652h→暫存,0Eh)、212Ch、輸出^ := <10AEh+1>(0)；輸出^=0 時 2020h、22BCh、<14F8h+2>+<15ADh+1>，並 Move(暫存→DS:0A652h,0Eh) 還原（回非 0 就不還原）。retf 4。spec 792 | — |
| `2457` | sub_2457 | DOCAMP | 481 | 198 | 0 | 15 | ✓ | 待解讀 | — | — | — |
| `2638` | sub_2638 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
