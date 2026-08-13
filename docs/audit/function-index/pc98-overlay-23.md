# pc98-overlay-23 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADEFFECTS | 18 | 5 | 0 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1627`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | knowledge/gold-box-ecl-interpreter.md |
| `0016` | sub_16 | KILLDUDE | 179 | 64 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `00C9` | sub_C9 | CALLEFFECT | 69 | 27 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `010E` | sub_10E | SPELLOFF | 315 | 115 | 6 | 3 | ✓ | 待解讀 | — | — | — |
| `0269` | sub_269 | — | 405 | 139 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `03FE` | sub_3FE | CHECKFX | 2287 | 1050 | 6 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md<br>spec/571-trytohit-attack-resolution.md |
| `0CED` | sub_CED | — | 251 | 97 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0E32` | sub_E32 | CHECKTERRAINFX | 914 | 399 | 0 | 11 | ✓ | 待解讀 | — | — | — |
| `11C4` | sub_11C4 | TRYTOHIT | 104 | 41 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/571-trytohit-attack-resolution.md<br>TRYTOHIT(modifier,target):d20 存全域 DS:A039h;自然 1 必失手、自然 20 改記為 100;CHECKFX(10h,target) 可改寫該全域;命中條件是 骰值+modifier > target^[19Bh](嚴格大於) | spec/571-trytohit-attack-resolution.md |
| `122C` | sub_122C | ATTEMPTTOHIT | 172 | 65 | 0 | 4 | ✓ | 待解讀 | — | — | spec/571-trytohit-attack-resolution.md |
| `12D8` | sub_12D8 | MAKESAVE | 144 | 56 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `1368` | sub_1368 | ROLLDICE | 75 | 29 | 5 | 1 | ✓ | 已解讀 | exact | docs/spec/568-rolldice-and-original-rng-entry.md<br>ROLLDICE(count,sides)=Σ(Random(sides)+1),回傳 byte;count<=0 回 0 | spec/568-rolldice-and-original-rng-entry.md |
| `13B3` | sub_13B3 | ROLLDAMAGEDICE | 36 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/568-rolldice-and-original-rng-entry.md<br>ROLLDAMAGEDICE(count,sides):把 count 寫入 DS:A032h 後委派 ROLLDICE;該全域用途未解 | audit/function-index/dos-overlay-23.md<br>spec/568-rolldice-and-original-rng-entry.md |
| `13D7` | sub_13D7 | ADDEFFECT | 175 | 58 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1486` | sub_1486 | LOSEDUDE | 138 | 51 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `1510` | sub_1510 | — | 21 | 7 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 21 bytes，已逐條讀完） | — |
| `1529` | sub_1529 | — | 25 | 7 | 4 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jz short near ptr sub_154C`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 25 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+0Ch]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `154C` | sub_154C | — | 6 | 3 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 6 bytes，已逐條讀完） | — |
| `1552` | sub_1552 | REMOVEINVIS | 9 | 4 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push [bp+arg_4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `155B` | sub_155B | — | 47 | 21 | 2 | 4 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 47 bytes，已逐條讀完） | — |
| `158A` | sub_158A | REMOVEFX | 98 | 39 | 3 | 3 | ✓ | 待解讀 | — | — | — |
| `15EC` | sub_15EC | ROUTINGREMOVEFX | 55 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1627` | sub_1627 | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | audit/function-triage.md |
| `1630` | sub_1630 | CUREEFFECT | 99 | 41 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `1693` | sub_1693 | CONVERTSTRTOSPEC | 41 | 17 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `16BC` | sub_16BC | CONVERTSPECTOSTR | 81 | 30 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `170D` | sub_170D | DONEWSTRENGTH | 73 | 28 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1756` | sub_1756 | — | 79 | 26 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `17A5` | sub_17A5 | — | 330 | 120 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `18F3` | sub_18F3 | RECALCULATESTATS | 18 | 6 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp+var_4], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | — |
| `1905` | sub_1905 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp+8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1919` | sub_1919 | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `shl ax, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `19CA` | sub_19CA | — | 1501 | 563 | 2 | 7 |  | 待解讀 | — | — | audit/function-triage.md |
| `1FFD` | sub_1FFD | PUTDAMAGE | 787 | 304 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `2325` | sub_2325 | PUTEFFECT | 212 | 87 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `2419` | sub_2419 | HEALDUDE | 209 | 71 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `251A` | sub_251A | STANDUP | 201 | 75 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `25E3` | sub_25E3 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
