# dos-overlay-13 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 292 | 94 | 0 | 5 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-24.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/754-small-predicates-and-wrappers.md |
| `0124` | sub_124 | — | 110 | 36 | 3 | 2 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>加成後夾制再乘 2(retf 4)：v := p^[1A5h]；p^[197h]=0(陣營)時 v += DS:4F9Dh^[6E4h]；DS:6F9Ah:=1；v 不在 1..60h 就設成 1(不是夾到邊界)；DS:6F96h := (v*2) and 0FFh；叫 <呼叫>(p, 12h)；DS:6F9Ah:=0；回傳 DS:6F96h | audit/embedded-strings.md<br>audit/function-index/dos-overlay-13.md<br>audit/function-index/pc98-overlay-13.md<br>spec/758-morale-field-0f7h-round-trip.md<br>spec/762-ega-glyph-blit-and-movement-rate.md<br>spec/763-dungeon-map-second-plane-and-stone-to-flesh.md |
| `0192` | sub_192 | — | 196 | 74 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md |
| `031D` | sub_31D | — | 841 | 334 | 3 | 4 | ✓ | 待解讀 | — | — | — |
| `0666` | sub_666 | — | 233 | 87 | 1 | 6 | ✓ | 待解讀 | — | — | spec/750-combat-setup.md |
| `074F` | sub_74F | — | 523 | 193 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `095A` | sub_95A | — | 935 | 334 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0D1C` | sub_D1C | — | 189 | 74 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0DD9` | sub_DD9 | — | 313 | 105 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0F12` | sub_F12 | — | 45 | 19 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>依相位進位的折半(retf 2)：f(x: byte) = (x + (DS:758Dh and 1)) div 2。shr 只用來取 bit 0 進 CF，值本身丟掉。DS:758Dh 在戰鬥開場(spec 750)被設 0 | audit/function-index/pc98-overlay-13.md<br>spec/754-small-predicates-and-wrappers.md |
| `0F46` | sub_F46 | — | 510 | 184 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1144` | sub_1144 | — | 169 | 57 | 4 | 2 | ✓ | 已解讀 | exact | docs/spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md<br>暫時借用目前目標欄位(retf 8)：甲為 NIL 回 false、乙=甲 回 true；否則清 DS:6F9Bh、叫 <呼叫>(1, 甲)，未中止就把甲塞進 乙^[18Dh]^[0Ah]、叫 <呼叫>(0, 乙)、再還原，最後回 DS:6F9Bh = 0。⚠ 目前目標欄位同時是參數傳遞通道，中途離開會留下錯誤的目標 | spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md<br>spec/769-combat-main-loop.md |
| `1227` | sub_1227 | — | 524 | 190 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `1433` | sub_1433 | — | 18 | 7 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | — |
| `1476` | sub_1476 | — | 31 | 12 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `14E8` | sub_14E8 | — | 34 | 10 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr ds:75D7h, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 34 bytes，已逐條讀完） | — |
| `1513` | sub_1513 | — | 12 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+14h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md |
| `156D` | sub_156D | — | 8 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1590` | sub_1590 | — | 8 | 2 | 9 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+1A4h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 10 | 6 | 10 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `15A4` | sub_15A4 | — | 15 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1470+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `15B3` | sub_15B3 | — | 12 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp-15h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `1652` | sub_1652 | — | 8 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+1A2h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `18A4` | sub_18A4 | — | 6 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr es:[di+19Bh], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `18AE` | sub_18AE | — | 19 | 7 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add di, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 19 bytes，已逐條讀完） | — |
| `1921` | sub_1921 | — | 5 | 2 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1926` | sub_1926 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1944` | sub_1944 | — | 6 | 3 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 6 bytes，已逐條讀完） | — |
| `194A` | sub_194A | — | 142 | 53 | 3 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `19D8` | sub_19D8 | — | 778 | 268 | 6 | 7 | ✓ | 待解讀 | — | — | — |
| `1CE2` | sub_1CE2 | — | 187 | 76 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/771-extra-attacks-and-weapon-class-whitelist.md<br>由門檻算出額外攻擊(retf 0Ch，第一個遠指標是輸出累加器)：n := <呼叫>(a, b, 角色)；<呼叫>(角色) 為真時 m := (byte[5D02h + 角色^[151h]^[2Eh]*10h] − 1) div 3，否則 m := n；n−m > 0 時 n -= m 並讓 輸出^ += 2，再判一次 += 3。兩段是先後不是迴圈，最多加兩次。5D02h 是物品屬性表(DS:5CF6h，16 bytes 一筆)的 +0Ch | spec/771-extra-attacks-and-weapon-class-whitelist.md |
| `1D9D` | sub_1D9D | — | 57 | 15 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call loc_104A`、`call far ptr loc_18B0+3`、`call loc_1953`、`call loc_19A3`、`call far ptr loc_15D0+1`、`call far ptr unk_11F7`（body 共 57 bytes，已逐條讀完） | audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md |
| `1DF6` | sub_1DF6 | — | 537 | 195 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `200F` | sub_200F | — | 507 | 176 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `2220` | sub_2220 | — | 1307 | 483 | 0 | 5 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-08.md<br>audit/function-triage.md<br>spec/769-combat-main-loop.md |
| `275A` | sub_275A | — | 361 | 132 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `28C3` | sub_28C3 | — | 223 | 73 | 3 | 3 | ✓ | 已解讀 | exact | docs/spec/771-extra-attacks-and-weapon-class-whitelist.md<br>物品類別白名單(retf 8)：(角色^[10Fh] > 0) 或 ((角色^[117h] > 0) 且 <呼叫>(角色)) 成立時，裝備 角色^[151h] 為 NIL 或其類別 +2Eh 落在 {7, 8, 23h, 24h, 25h, 61h} 就通過(無號比較，區間是 22h < 類別 < 26h)；再要求 行動者^[18Dh]^[0Fh] > 1 且 (行動者^[0DEh] and 7Fh) <= 1，最後回傳 本模組 29A2h(行動者, 角色) = 行動者^[18Dh]^[09h] | spec/771-extra-attacks-and-weapon-class-whitelist.md |
| `29A2` | sub_29A2 | — | 601 | 239 | 6 | 2 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-13.md<br>spec/771-extra-attacks-and-weapon-class-whitelist.md |
| `2BFB` | sub_2BFB | — | 551 | 216 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `2E22` | sub_2E22 | — | 138 | 49 | 0 | 1 | ✓ | 待解讀 | — | — | spec/750-combat-setup.md |
| `2EAC` | sub_2EAC | — | 130 | 46 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/762-ega-glyph-blit-and-movement-rate.md<br>取指定陣營中最快的移動速率(retf 4)：走 DS:650Ah 隊伍鏈，<呼叫>(參數) 等於 p^[197h](陣營)且 p^[196h] 非 0 的成員，取 本模組 0124h(p) div 2 的最大值(無號比較)。0124h 回的是速率乘 2，這裡除回來 — 與既有『距離單位是半格』的 2 是同一個 | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-13.md<br>spec/762-ega-glyph-blit-and-movement-rate.md<br>spec/763-dungeon-map-second-plane-and-stone-to-flesh.md |
| `2F3C` | sub_2F3C | — | 205 | 67 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3040` | sub_3040 | — | 432 | 177 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `31F0` | sub_31F0 | — | 361 | 132 | 2 | 10 | ✓ | 待解讀 | — | — | — |
| `33AC` | sub_33AC | — | 1278 | 493 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `38AA` | sub_38AA | — | 161 | 60 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `394B` | sub_394B | — | 428 | 142 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3B37` | sub_3B37 | — | 799 | 315 | 1 | 10 | ✓ | 待解讀 | — | — | — |
| `3E56` | sub_3E56 | — | 480 | 159 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `403F` | sub_403F | — | 291 | 117 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `4162` | sub_4162 | — | 106 | 39 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/759-combatant-xy-lookup-and-bank-select-bit.md<br>取兩個戰鬥員的 x/y 後畫(retf 4)：先 overlay-24 entry#24 @1AAAh(參數+0Dh)；再用 overlay-32 entry#15/#16 對 SS 相對記錄的 +0Ch 與 −0Ah 兩個遠指標各取 x、y；最後 overlay-24 entry#25 @1AF7h(x1, y1, x2, y2, 1, 1Eh)。entry#15/#16 = 由 overlay-32:0CAEh 換索引後取 4-byte 記錄的 +0/+1，即戰鬥員座標 — 補上 spec 750「地圖表頭 = 隊伍位置 −3」的機制依據 | audit/function-index/pc98-overlay-13.md<br>spec/759-combatant-xy-lookup-and-bank-select-bit.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `41CC` | sub_41CC | — | 186 | 63 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `42FD` | sub_42FD | — | 722 | 312 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `45CF` | sub_45CF | — | 308 | 111 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `4703` | sub_4703 | — | 264 | 93 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `4811` | sub_4811 | — | 183 | 70 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/770-party-step-and-hug-attack.md<br>hugs 攻擊(retf 0Ah)：DS:6F9Fh < 12h(有號)就離開；把 攻擊者^[18Dh]^[0Ah] 抄到全域 DS:6FA3h，組出「攻擊者 + CS:480Bh 'hugs ' + 目標名」後顯示；再對目標下 3Ah、對攻擊者自己下 90h 兩個效果碼。⚠ 'hugs ' 的尾隨空白是常數的一部分(長度 byte 為 5)，用來直接接目標名 | spec/770-party-step-and-hug-attack.md |
| `48DC` | sub_48DC | — | 214 | 76 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/772-gods-intervene-money-display-and-sound-commands.md<br>神明介入(retf，無參數)：<0A54h:0724h>(@DS:8C80h) 以 ZF 回報，成立才做 — 顯示 CS:48C8h 'The Gods intervene!'，走隊伍鏈把 +197h(陣營)為 1 的成員 +196h 設 0、+195h(狀態)設 6，並把 byte[66A6h + 4*索引] 歸零；最後 <呼叫>(地圖^[2]+3, 地圖^[3]+3, 0FFh, 8) — 加回 spec 750 減掉的 3，印證戰場原點是隊伍位置退三格。⚠ DS:66A6h 既是 DS:6D35h 的項數又是 1-based 陣列的第 0 格 | spec/772-gods-intervene-money-display-and-sound-commands.md |
| `49B2` | sub_49B2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
