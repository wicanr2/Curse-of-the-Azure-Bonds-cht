# pc98-overlay-24 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADGENERIC | 37 | 11 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 6 個呼叫，沒有其他動作：`call loc_1982+1`、`call sub_19CA`、`call far ptr loc_17B7`、`call loc_1749+1`、`call sub_1697`、`call loc_1626+1`（body 共 37 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md<br>context/50-log-2026-08-09-13.md<br>spec/507-pc98-general-target-object-order-projection.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/516-fire-knife-external-map-handoff-audit.md |
| `0025` | sub_25 | — | 504 | 179 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/525-pc98-tempsearch-display-state.md<br>spec/564-ecl-operand-decoding-and-arity-validation.md |
| `021D` | sub_21D | — | 138 | 45 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02A7` | sub_2A7 | — | 286 | 104 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `03C5` | sub_3C5 | — | 141 | 52 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `046C` | sub_46C | PRINTITEMNAME | 1087 | 417 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0917` | sub_917 | SHOWALL | 549 | 229 | 3 | 5 | ✓ | 待解讀 | — | — | — |
| `0B41` | sub_B41 | PRINTAC | 146 | 61 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0BD5` | sub_BD5 | PRINTHP | 99 | 38 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/643-hp-fields-and-slot-array.md<br>顯示 HP:if arg_6^[1A5h](目前 HP)< arg_6^[78h] then v := 6 else v := 0Ah;arg_0 非零則 v := 0Dh;再用 <sub_12E5>(Str 的 byte 版)把 HP 轉成字串連同 v 傳給顯示 routine,接上一個空白後 <far 0418:0D17h>。這確認 +78h 是 HP 上限(DOS 20FAh 的夾限是另一條證據,兩平台同偏移) | spec/643-hp-fields-and-slot-array.md |
| `0C53` | sub_C53 | SHOWINFO | 431 | 196 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0E02` | sub_E02 | HELPLESS | 69 | 28 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/643-hp-fields-and-slot-array.md<br>for i := 1 to 4:if <sub_25CD>(arg_0,arg_2,byte[49D3h + i],@var_6) <> 0 則 result := 1。⚠ 沒有提前結束,第一個成功後剩下三個照跑,副作用會發生四次。索引由 1 到 4 不是 0..3 | audit/function-index/dos-overlay-24.md<br>spec/643-hp-fields-and-slot-array.md |
| `0E47` | sub_E47 | DOCOM | 1182 | 390 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `12E5` | sub_12E5 | BYTETOSTR | 60 | 28 | 4 | 0 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br>Str 的 byte 版包裝:mov al + xor ah,ah + xor dx,dx 把輸入湊成無號 32-bit,呼叫 0A65h:12FBh(Str),再用 0A65h:649h 把結果指派給呼叫端的字串變數(上限 0FFh)。無號,0FFh 不會印成 -1 | audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md<br>spec/634-str-conversion-family.md |
| `1321` | sub_1321 | WORDTOSTR | 58 | 27 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br>Str 的 word 版包裝:mov ax + xor dx,dx(無號),其餘與 12E5h 相同 | audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md<br>spec/634-str-conversion-family.md |
| `135B` | sub_135B | LONGTOSTR | 57 | 25 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br>Str 的 longint 版包裝:直接推 arg_2:arg_0 整個 dword,retf 4;其餘與 12E5h/1321h 相同 | audit/function-index/dos-overlay-24.md<br>spec/634-str-conversion-family.md |
| `1394` | sub_1394 | DEXDEFBONUS | 130 | 54 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1416` | sub_1416 | DEXRABONUS | 130 | 54 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1498` | sub_1498 | — | 152 | 62 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1530` | sub_1530 | STRHITBONUS | 163 | 67 | 2 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `15D3` | sub_15D3 | STRDAMBONUS | 79 | 31 | 2 | 3 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jb short loc_1633`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 79 bytes，已逐條讀完） | — |
| `1622` | sub_1622 | — | 55 | 23 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 55 bytes，已逐條讀完） | — |
| `1659` | sub_1659 | STRWGTBONUS | 22 | 9 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp+var_3]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 22 bytes，已逐條讀完） | audit/function-index/pc98-overlay-24.md |
| `166F` | sub_166F | — | 40 | 16 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `ja short loc_169F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 40 bytes，已逐條讀完） | — |
| `1697` | sub_1697 | — | 162 | 64 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1659h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1739` | sub_1739 | LOSEMEMSPELL | 65 | 24 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/643-hp-fields-and-slot-array.md<br>在紀錄^[1Eh + 0..53h](84 格)裡找出第一個等於 arg_0 的格子清成 0,然後把迴圈變數設成 54h 提前結束(Pascal 沒有 break 的標準寫法)。只清第一個相符的。這 84 格與 spec 624 的 STORESPECIALS 寫入範圍(+1Fh..+6Fh,81 格)完全吻合,兩邊各自獨立量到同一個陣列 | audit/function-index/dos-overlay-24.md<br>spec/643-hp-fields-and-slot-array.md |
| `1785` | sub_1785 | — | 15 | 4 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub byte ptr [bp+si-5F7Eh], 82h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `1794` | sub_1794 | — | 12 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add word ptr [bx+di-7Dh], 43h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `17D1` | sub_17D1 | LOSEITEM | 154 | 58 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1865` | sub_1865 | — | 52 | 19 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 81Fh:1A2h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 52 bytes，已逐條讀完） | — |
| `186A` | sub_186A | — | 31 | 14 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 31 bytes，已逐條讀完） | — |
| `18B8` | sub_18B8 | GAINITEM | 52 | 22 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+54h], ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 52 bytes，已逐條讀完） | — |
| `18EC` | sub_18EC | — | 96 | 32 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 96 bytes，已逐條讀完） | — |
| `194C` | sub_194C | HORZMSG | 92 | 46 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `19A8` | sub_19A8 | CHARMSG | 34 | 14 | 4 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jnz short loc_1A16`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 34 bytes，已逐條讀完） | audit/function-index/pc98-overlay-24.md |
| `19CA` | sub_19CA | — | 198 | 94 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 19A8h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-triage.md |
| `1A90` | sub_1A90 | CLEARMSGWINDOW | 50 | 26 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>與 dos overlay-24:18BEh 助憶碼序列完全相同（26 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：依條件呼叫 loc_1ECC+1:一邊傳 (15h,...),另一邊傳 (16h,26h,12h,1) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1AC2` | sub_1AC2 | PRINTCHARNAME | 109 | 41 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1B2F` | sub_1B2F | COPYSPRITE | 287 | 113 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1C48` | sub_1C48 | — | 40 | 15 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:1B0Dh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 40 bytes，已逐條讀完） | — |
| `1C76` | sub_1C76 | CREATERADIALSPRITE | 77 | 45 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1CC3` | sub_1CC3 | MOVESPRITE | 978 | 329 | 0 | 9 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-24.md |
| `2095` | sub_2095 | — | 437 | 142 | 7 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1CC3h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `224A` | sub_224A | — | 448 | 146 | 2 | 8 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1CC3h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `240A` | sub_240A | TWINKLE | 451 | 162 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `25CD` | sub_25CD | SPELLON | 107 | 37 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `2658` | sub_2658 | SAVEDAMAGE | 254 | 80 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2776` | sub_2776 | HEALMSG | 109 | 47 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `27E3` | sub_27E3 | ENEMYOF | 36 | 13 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br>回傳 (arg_0^[198h] = 0):是則 1、否則 0 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-24.md<br>audit/string-pairs.md<br>spec/634-str-conversion-family.md |
| `2807` | sub_2807 | FIGGOODBAD | 84 | 29 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `285B` | sub_285B | FINDFOES | 292 | 110 | 0 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>spec/508-pc98-general-target-scan-producer.md |
| `297F` | sub_297F | FIGRANGE | 220 | 81 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2A5B` | sub_2A5B | QUIT | 70 | 20 | 1 | 0 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md |
| `2AAE` | sub_2AAE | GUARD | 60 | 24 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br>r := <sub_2A5B>(arg_0);arg_0^[18Eh]^[7] := 1;備妥「防御している」呼叫 <sub_194C>;回傳 r。回傳值是進來時就算好的,設旗標與顯示訊息都在那之後,所以不受旗標影響 | audit/function-index/dos-overlay-24.md<br>spec/634-str-conversion-family.md |
| `2AEA` | sub_2AEA | FIGCASTERLEVEL | 438 | 149 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2CA4` | sub_2CA4 | SETUPSCREEN | 456 | 192 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `2E8C` | sub_2E8C | SHOWLOCATION | 603 | 248 | 1 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/516-fire-knife-external-map-handoff-audit.md<br>spec/525-pc98-tempsearch-display-state.md |
| `30E7` | sub_30E7 | DOCOMBATSCREEN | 54 | 22 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/634-str-conversion-family.md<br><far 018C:0301>() → <sub_1F6E>() → <sub_18F1>(DS:9F2Ch^[2]+3, DS:9F2Ch^[3]+3, 0FFh, 8)。兩個 byte 都用 cbw(有號)各加固定的 3 | audit/function-index/dos-overlay-24.md<br>spec/634-str-conversion-family.md |
| `314B` | sub_314B | PICKAPERSON | 374 | 141 | 0 | 5 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-24.md |
| `32B6` | sub_32B6 | — | 123 | 41 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 314Bh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `3334` | sub_3334 | USINGMISSLEWEAPON | 67 | 23 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3377` | sub_3377 | USINGHURLEDWEAPON | 69 | 27 | 0 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md |
| `33BC` | sub_33BC | VALIDMISSLE | 194 | 69 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `347E` | sub_347E | LEVELLIMIT | 327 | 110 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `35D8` | sub_35D8 | CHECKDYING | 168 | 57 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `3680` | sub_3680 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
