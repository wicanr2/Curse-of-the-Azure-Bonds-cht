# dos-overlay-19 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 57 | 15 | 0 | 6 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call far ptr loc_16BB+2`、`call far ptr loc_17E5+2`、`call far ptr loc_11F6+1`、`call far ptr loc_14AC+1`、`call far ptr loc_1183+1`、`call far ptr loc_10B2+2`（body 共 57 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-17.md<br>audit/overlay-init-graph.md |
| `0083` | sub_83 | — | 1300 | 535 | 4 | 6 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0597` | sub_597 | — | 183 | 79 | 4 | 1 | ✓ | 已解讀 | exact | docs/spec/772-gods-intervene-money-display-and-sound-commands.md<br>七欄的金錢顯示(retf，無參數)：for i := 6 downto 0，DS:6506h^[0FBh + i*2] > 0(無號)才佔一列 — 名稱取自 DS:0F93h + i*0Bh(每筆 11 bytes 的 Pascal 短字串)、右對齊到第 20 欄(14h − 長度)，數值固定從第 21 欄起，列號從 7 往下遞增。⚠ 右對齊用 byte 數算，全形中文會算錯；名稱表在常駐資料段，scan_pascal_strings.py 掃不到。補記(spec 773)：DS:0F93h 的七個名稱已讀出 — Copper/Silver/Electrum/Gold/Platinum/Gems/Jewelry | spec/772-gods-intervene-money-display-and-sound-commands.md |
| `0698` | sub_698 | — | 786 | 332 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `09B3` | sub_9B3 | — | 341 | 135 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/782-strength-percentile-display.md<br>力量百分位的顯示(retf 4)：能力值在 DS:6506h^[11h + 索引*2]，畫在 索引+7 列；值 < 0Ah 時欄由 5 改 6 靠右對齊；索引 0 且值 = 12h(18) 且 +1Ch > 0 時再畫 '(' + 百分位 + ')'(CS:09AFh/09B1h)，百分位 < 0Ah 補 CS:09AAh 的 '0'、等於 64h(100) 直接用 CS:09ACh 的 '00' — 即 AD&D 的 18/00 記法。顏色依第一個參數在 0Dh/0Ah 二選一。⚠ 欄位靠硬編碼對齊，沒有量測字串寬度 | audit/function-index/pc98-overlay-19.md<br>spec/782-strength-percentile-display.md |
| `0B75` | sub_B75 | — | 768 | 303 | 0 | 14 | ✓ | 待解讀 | — | — | — |
| `0EC5` | sub_EC5 | — | 247 | 101 | 2 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `10AB` | sub_10AB | — | 282 | 138 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-19.md |
| `11C5` | sub_11C5 | — | 6 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `11E3` | sub_11E3 | — | 24 | 13 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:634h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 24 bytes，已逐條讀完） | — |
| `1494` | sub_1494 | — | 10 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr unk_1522`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `149E` | sub_149E | — | 10 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | audit/function-index/dos-overlay-15.md<br>spec/789-heal-budget-tick-and-halve-stack.md |
| `14A8` | sub_14A8 | — | 6 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1518` | sub_1518 | — | 9 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and [si+72h], dl`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `1554` | sub_1554 | — | 949 | 374 | 2 | 10 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 10ABh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1588` | sub_1588 | — | 33 | 11 | 2 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov dx, es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 33 bytes，已逐條讀完） | audit/function-index/dos-overlay-19.md |
| `15A9` | sub_15A9 | — | 11 | 4 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-37h], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `15B8` | sub_15B8 | — | 32 | 13 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 32 bytes，已逐條讀完） | — |
| `1643` | sub_1643 | — | 85 | 29 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr ds:4FBAh, 5`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 85 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 556 | 229 | 3 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1588h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18F9` | sub_18F9 | — | 21 | 7 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_1ADB`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 21 bytes，已逐條讀完） | — |
| `19EA` | sub_19EA | — | 22 | 12 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 16h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 22 bytes，已逐條讀完） | audit/function-triage.md |
| `1A00` | sub_1A00 | — | 250 | 104 | 4 | 8 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1588h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1B00` | sub_1B00 | — | 877 | 320 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `1EE9` | sub_1EE9 | — | 567 | 203 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `213C` | sub_213C | — | 166 | 61 | 1 | 4 | ✓ | 已解讀 | exact | docs/spec/767-ac-display-sign-and-trade-overload.md<br>交易給誰(retf 4)：以 DS:4AF8h 記住上次選的人當預設，用 CS:2120h 'Trade with Whom?' 選人；取消就離開；否則寫回 DS:4AF8h，用 overlay-19:3258h(spec 762)判超重 — 超重就顯示 CS:2131h 'Overloaded'，否則做三個呼叫(從物品端移除、加到 DS:6506h、更新對象) | audit/function-index/pc98-overlay-19.md<br>spec/767-ac-display-sign-and-trade-overload.md |
| `21F3` | sub_21F3 | — | 176 | 66 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/789-heal-budget-tick-and-halve-stack.md<br>物品堆對半分：半 := 有號(節點^[39h]) div 2；半 > 0（無號比較）才做。GetMem(新,3Fh) + Move(節點^,新^,3Fh) 複製整個 63-byte 節點，新^[39h]:=半、新^[34h]:=0、新^[2Ah]:=節點^[2Ah]、節點^[39h]:=數量−半、節點^[2Ah]:=新（插在原節點後）。否則顯示 'Can't halve that'(16)。數量 ≥ 80h 時有號 div 2 為負而無號比較會通過，無防護。retf 4。spec 789 | spec/789-heal-budget-tick-and-halve-stack.md |
| `22A3` | sub_22A3 | — | 465 | 154 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-12.md |
| `248D` | sub_248D | — | 742 | 263 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `27D6` | sub_27D6 | — | 482 | 183 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `2A42` | sub_2A42 | — | 465 | 193 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `2C57` | sub_2C57 | — | 793 | 307 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `2FA9` | sub_2FA9 | — | 687 | 269 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `3258` | sub_3258 | — | 132 | 45 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/762-ega-glyph-blit-and-movement-rate.md<br>拿不拿得下(retf 8)：角色^[14Ch] > 0Fh(物品數上限 16)就超過；w := 物品^[37h]，物品^[39h] > 0 時再乘上去(單位重量×數量)；上限 := <呼叫>(角色) + 5DCh(1500)；角色^[187h](累計負重) + w > 上限(無號)也算超過 | audit/function-index/dos-overlay-19.md<br>audit/function-index/pc98-overlay-19.md<br>spec/762-ega-glyph-blit-and-movement-rate.md<br>spec/767-ac-display-sign-and-trade-overload.md<br>spec/777-item-append-and-area-save-sweep.md |
| `32DC` | sub_32DC | — | 207 | 71 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `33AB` | sub_33AB | — | 118 | 44 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3474` | sub_3474 | — | 381 | 158 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `35F1` | sub_35F1 | — | 101 | 37 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>與 3656h 成對的判斷(retf 4)：前三個條件與 overlay-19:3656h(spec 756)逐字相同，最後一項改成呼叫 <overlay>(輸出參數, 8Ch, p) 並要求回 0 | audit/function-index/pc98-overlay-19.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `3656` | sub_3656 | — | 89 | 30 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>四條件合取(retf 4)：((p^[75h]=3) or ((p^[114h]>0) and <overlay 呼叫>(p))) and (DS:4FBAh<>5) and (p^[195h]=0) and (p^[191h]>0)。DS:4FBAh 就是戰鬥開場(spec 750)結尾設 5 的狀態變數；+114h 有號比較、+191h 無號 | audit/function-index/dos-overlay-14.md<br>audit/function-index/dos-overlay-19.md<br>audit/function-index/pc98-overlay-19.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/758-morale-field-0f7h-round-trip.md<br>spec/769-combat-main-loop.md |
| `36D8` | sub_36D8 | — | 224 | 91 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `37EC` | sub_37EC | — | 375 | 145 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `3963` | sub_3963 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
