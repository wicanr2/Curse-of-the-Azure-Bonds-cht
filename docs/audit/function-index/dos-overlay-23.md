# dos-overlay-23 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 22 | 8 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 3 個呼叫，沒有其他動作：`call loc_15D0+1`、`call far ptr sub_19EA`、`call loc_1657`（body 共 22 bytes，已逐條讀完） | knowledge/gold-box-ecl-interpreter.md |
| `0016` | sub_16 | — | 179 | 64 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `00C9` | sub_C9 | — | 69 | 27 | 4 | 1 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:00C9h 助憶碼序列完全相同（27 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：CALLEFFECT:DS:A66Eh 非零時一律 far call DS:A28Ch(忽略 id);否則以 id*4 索引 DS:A040h 起的 far pointer 表分派。五個參數原樣轉傳 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-23.md<br>spec/576-adnd-strength-encoding-and-effect-removal.md |
| `010E` | sub_10E | — | 315 | 115 | 6 | 3 | ✓ | 待解讀 | — | — | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `0269` | sub_269 | — | 405 | 139 | 1 | 7 | ✓ | 已解讀 | strong inference | docs/spec/577-attempttohit-and-effect-chain-walk.md<br>與 pc98 overlay-23:0269h 助憶碼序列完全相同（139 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：effect 鏈遍歷入口(161 個呼叫者):先查 subject 自身;未命中且 id 屬於集合 {15h,2Dh,2Eh,31h}(函式前方 32 bytes 的 Turbo Pascal set 常數,兩平台逐位元組相同)才沿 DS:9598h 的鏈以 +18Ah 走訪,期間借用 DS:9F31h 起 0D8h bytes 並在前後 Move 備份/還原;命中則呼叫 CALLEFFECT 分派 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/577-attempttohit-and-effect-chain-walk.md |
| `03FE` | sub_3FE | — | 2313 | 1051 | 6 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md<br>spec/571-trytohit-attack-resolution.md |
| `0D07` | sub_D07 | — | 252 | 97 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0E45` | sub_E45 | — | 914 | 399 | 0 | 12 | ✓ | 待解讀 | — | — | — |
| `11D7` | sub_11D7 | — | 104 | 41 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/571-trytohit-attack-resolution.md<br>與 pc98 overlay-23:11C4h 助憶碼序列完全相同（41 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：TRYTOHIT(modifier,target):d20 存全域 DS:A039h;自然 1 必失手、自然 20 改記為 100;CHECKFX(10h,target) 可改寫該全域;命中條件是 骰值+modifier > target^[19Bh](嚴格大於) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `123F` | sub_123F | — | 172 | 65 | 0 | 4 | ✓ | 已解讀 | strong inference | docs/spec/577-attempttohit-and-effect-chain-walk.md<br>與 pc98 overlay-23:122Ch 助憶碼序列完全相同（65 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：ATTEMPTTOHIT(need, attacker, target):d20 存進全域 DS:A039h,<=1 必失手、=20 改成 100;呼叫 CHECKFX(0Ah,target) 與 CHECKFX(10h,attacker)(兩者可改寫該全域);依 target^[198h] 取 bank1^[6E2h] 或 [6E0h];A039h<0 再失手一次;命中條件 A039h + target^[19Ah] + v >= need ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/577-attempttohit-and-effect-chain-walk.md |
| `12EB` | sub_12EB | — | 148 | 57 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `137F` | sub_137F | — | 75 | 29 | 5 | 1 | ✓ | 已解讀 | exact | docs/spec/568-rolldice-and-original-rng-entry.md<br>ROLLDICE(count,sides):與 PC-98 同構;亂數 far call 0A54:1105 = @Random$q4Word(image offset B645h) | spec/568-rolldice-and-original-rng-entry.md |
| `13CA` | sub_13CA | — | 36 | 16 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/568-rolldice-and-original-rng-entry.md<br>與 PC-98 overlay-23:13B3h（entry#10）助憶碼序列完全相同，語意同該筆：ROLLDAMAGEDICE(count,sides):把 count 寫入 DS:A032h 後委派 ROLLDICE;該全域用途未解 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `13EE` | sub_13EE | — | 175 | 58 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `149D` | sub_149D | — | 95 | 38 | 0 | 5 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+arg_8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 95 bytes，已逐條讀完） | — |
| `1540` | sub_1540 | — | 21 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1599+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 21 bytes，已逐條讀完） | — |
| `1559` | sub_1559 | — | 16 | 7 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 16 bytes，已逐條讀完） | — |
| `1569` | sub_1569 | — | 14 | 6 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 19h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 14 bytes，已逐條讀完） | — |
| `1577` | sub_1577 | — | 5 | 3 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `157C` | sub_157C | — | 22 | 9 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 22 bytes，已逐條讀完） | — |
| `15A1` | sub_15A1 | — | 98 | 39 | 3 | 3 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:158Ah 助憶碼序列完全相同（39 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：REMOVEFX:以 1..19 索引 DS:159Dh 起的位元組表逐一呼叫 SPELLOFF(char, id, 0, 0);之後查找 effect 4Dh,命中且 char^[0F7h]=0B3h 時把 char^[198h] 清零。清單內容在 DS,未定位 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `1603` | sub_1603 | — | 55 | 24 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:15ECh 助憶碼序列完全相同（24 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：ROUTINGREMOVEFX:以 1..4 索引 DS:15B1h 起的位元組表逐一呼叫 SPELLOFF(char, id, 0, 0)。清單內容在 DS,未定位 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `1643` | sub_1643 | — | 99 | 41 | 0 | 4 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:1630h 助憶碼序列完全相同（41 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：CUREEFFECT(id, char):查不到該 effect 回 false;否則格式化輸出、停頓(1,10)、呼叫 SPELLOFF 解除,回 true ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `16A6` | sub_16A6 | — | 41 | 17 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:1693h 助憶碼序列完全相同（17 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：CONVERTSTRTOSPEC(percentile, str):str=18 回 percentile+1(1..101),否則回 str+100(103..118)。與 CONVERTSPECTOSTR 互逆 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `16CF` | sub_16CF | — | 81 | 30 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:16BCh 助憶碼序列完全相同（30 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：CONVERTSPECTOSTR(rec, out_pct, out_str):取 rec^[3] and 7Fh;<=101 則 str=18、pct=值-1;否則 str=值-100、pct=0。角色記錄 +3 是力量編碼欄位,最高位是旗標 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `1720` | sub_1720 | — | 73 | 28 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:170Dh 助憶碼序列完全相同（28 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：DONEWSTRENGTH:新力量 > char^[10h],或(新力量=18 且新百分位 > char^[1Dh])時,把 CONVERTSTRTOSPEC 的結果寫進 out^ 並回 true。釘死 +10h=力量、+1Dh=百分位、12h=18 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | spec/576-adnd-strength-encoding-and-effect-removal.md |
| `1769` | sub_1769 | — | 79 | 26 | 2 | 1 | ✓ | 已解讀 | strong inference | docs/spec/576-adnd-strength-encoding-and-effect-removal.md<br>與 pc98 overlay-23:1756h 助憶碼序列完全相同（26 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：同 DONEWSTRENGTH 的比較規則,對 SS 上的區域結構操作:p[-2]:p[-1] 為力量、p[-4]:p[-3] 為百分位,較大者寫回 p[-1]/p[-3](當前值同步到最大值欄位) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `17B8` | sub_17B8 | — | 246 | 88 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `18AE` | sub_18AE | — | 42 | 15 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 42 bytes，已逐條讀完） | — |
| `18D8` | sub_18D8 | — | 1089 | 409 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `1935` | sub_1935 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+34h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `193A` | sub_193A | — | 5 | 2 | 3 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_1B1B`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `193F` | sub_193F | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+3Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `194E` | sub_194E | — | 156 | 56 | 3 | 2 |  | 待解讀 | — | — | — |
| `19EA` | sub_19EA | — | 323 | 115 | 2 | 3 |  | 待解讀 | — | — | audit/function-triage.md |
| `1FD6` | sub_1FD6 | — | 803 | 309 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `2307` | sub_2307 | — | 212 | 87 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `23FB` | sub_23FB | — | 209 | 71 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `24ED` | sub_24ED | — | 201 | 75 | 0 | 5 | ✓ | 待解讀 | — | — | — |
