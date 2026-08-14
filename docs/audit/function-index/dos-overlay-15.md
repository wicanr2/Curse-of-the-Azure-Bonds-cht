# dos-overlay-15 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 52 | 14 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 9 個相依 unit 的 0000h — overlay-34、overlay-22、overlay-20、overlay-24、overlay-16、overlay-26、overlay-19、overlay-28、overlay-23。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/dos-overlay-28.md |
| `0034` | sub_34 | — | 340 | 73 | 1 | 1 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-index/dos-overlay-14.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/dos-overlay-23.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `0188` | sub_188 | — | 67 | 24 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>清掉超界值(retf 4)：for i:=0 to 53h，若 p^[1Eh+i] > 7Fh(無號)則歸零；最後 p^[72h]:=0。1Eh+53h=71h 與 72h 相接，可證角色記錄 +1Eh 起是 84 格陣列。⚠ IDA 匯出在 01BDh 吃掉 ES 前綴 26h，看似兩平台差異，回查原始 bytes 兩邊相同 | audit/function-index/dos-overlay-14.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `01CB` | sub_1CB | — | 115 | 42 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/760-item-effect-flag-clear-and-file-exists.md<br>清掉物品三個效果槽的旗標位元(retf 4)：走角色 +14Dh 的物品鏈，本模組 11DEh(p) 為真時把 p^[3Ch]/[3Dh]/[3Eh] 各 and 7Fh(保留效果 ID、清掉最高位旗標)，next 在 +2Ah。呼叫端是 overlay-15:023Eh(spec 755) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `023E` | sub_23E | — | 68 | 27 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>整隊清理(retf，無參數)：從 DS:650Ah 沿 +189h 走隊伍鏈，對每個成員叫本模組 0188h 與 01CBh | audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-15.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `0282` | sub_282 | — | 271 | 44 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0391` | sub_391 | — | 24 | 7 | 3 | 0 | ✓ | 待解讀 | — | — | — |
| `03A9` | nullsub_1 | — | 3 | 1 | 0 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 3 bytes，已逐條讀完） | — |
| `04CB` | sub_4CB | — | 182 | 77 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `05CA` | sub_5CA | — | 488 | 142 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `07B2` | sub_7B2 | — | 170 | 66 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `085C` | sub_85C | — | 155 | 57 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md<br>把 84 格陣列排序(retf，無參數)：對 DS:6506h(目前角色)的 +1Eh 起 84 格做選擇排序，比較時 and 7Fh、交換時整個 byte 一起搬(高位旗標跟著值走)，遞增。固定跑 84×85/2 次比較，沒有提早結束 | spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md |
| `0942` | sub_942 | — | 439 | 180 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `0B9B` | sub_B9B | — | 718 | 257 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `0E69` | sub_E69 | — | 63 | 25 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, [bp+arg_2]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 63 bytes，已逐條讀完） | — |
| `1016` | sub_1016 | — | 12 | 6 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md |
| `1022` | sub_1022 | — | 12 | 4 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les ax, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `11DE` | sub_11DE | — | 1096 | 472 | 3 | 7 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1016h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-index/dos-overlay-15.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `14D2` | sub_14D2 | — | 86 | 41 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1522` | sub_1522 | — | 92 | 38 | 4 | 8 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_14DC`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 92 bytes，已逐條讀完） | — |
| `1554` | sub_1554 | — | 2 | 1 | 7 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1581`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | — |
| `158A` | sub_158A | — | 31 | 12 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov dx, es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-23.md |
| `15A9` | sub_15A9 | — | 219 | 68 | 7 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 158Ah 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `15D1` | sub_15D1 | — | 35 | 11 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-8], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 35 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:650Ah, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16AE` | sub_16AE | — | 6 | 2 | 5 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_16C7`，控制權轉交後不返回；先設定 `mov ds:650Ch, dx`（body 共 6 bytes，已逐條讀完） | — |
| `16B8` | sub_16B8 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16BD` | sub_16BD | — | 14 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 14 bytes，已逐條讀完） | — |
| `16CB` | sub_16CB | — | 255 | 82 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1834` | sub_1834 | — | 250 | 103 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `199C` | sub_199C | — | 330 | 136 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `1B4D` | sub_1B4D | — | 348 | 146 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1D0A` | sub_1D0A | — | 459 | 198 | 1 | 7 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1EDA` | sub_1EDA | — | 35 | 14 | 2 | 1 | ✓ | 已解讀 | strong inference | docs/spec/637-overlay21-small-batch.md<br>與 pc98 overlay-15:1F46h 助憶碼序列完全相同（14 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：回傳 (arg_2^[196h] = 0):是則 1、否則 0。+196h 是狀態碼(spec 623 記到 6 與 5→4 兩條) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1EFD` | sub_1EFD | — | 188 | 74 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1FB9` | sub_1FB9 | — | 178 | 70 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/770-party-step-and-hug-attack.md<br>三級治療量累加(retf 2，SS 相對記錄)：70 條指令的助憶碼與 PC-98 overlay-15:2020h(spec 763)完全相同 — ss:[p−4] 次 1d8、ss:[p−6] 次 2d8+1、ss:[p−8] 次 3d8+3，累加到 ss:[p−2] | spec/770-party-step-and-hug-attack.md |
| `206B` | sub_206B | — | 90 | 33 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>全隊受傷總量(retf 2，參數沒被讀)：走隊伍鏈累加 p^[78h] − p^[1A4h](最大 HP 減目前 HP)，回 word。⚠ 兩個 byte 零延伸後相減，目前 HP 大於最大 HP 時會變成很大的正數 | audit/function-index/pc98-overlay-15.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `20C5` | sub_20C5 | — | 400 | 143 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2255` | sub_2255 | — | 180 | 63 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2309` | sub_2309 | — | 164 | 68 | 1 | 9 | ✓ | 待解讀 | — | — | — |
| `23FC` | sub_23FC | — | 485 | 206 | 0 | 12 | ✓ | 待解讀 | — | — | — |
| `25E1` | sub_25E1 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
