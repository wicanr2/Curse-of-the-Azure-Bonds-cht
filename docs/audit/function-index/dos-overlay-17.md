# dos-overlay-17 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 62 | 17 | 0 | 6 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 11 個相依 unit 的 0000h — overlay-24、overlay-19、overlay-22、overlay-23、overlay-29、overlay-33、overlay-16、overlay-26、overlay-32、overlay-34、overlay-25。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-25.md |
| `003E` | sub_3E | — | 415 | 87 | 1 | 1 | ✓ | 已解讀 | exact | 849<br>角色記錄解構子(retf 4，指向角色遠指標的遠指標)：釋放 +18Dh 戰鬥狀態(16h)、+14Dh 物品鏈(每個 3Fh)、+0F2h 一條 9 bytes 的鏈(next 在 +5)，最後 FreeMem(角色, 1A6h) 並寫回 NIL。★DOS 角色記錄 = 422 bytes 的直接證據 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-19.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-25.md<br>audit/function-index/dos-overlay-32.md |
| `01DD` | sub_1DD | — | 1445 | 363 | 0 | 11 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp loc_1F8`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1445 bytes，已逐條讀完） | audit/function-strings.md |
| `0782` | sub_782 | — | 2622 | 999 | 1 | 7 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-17.md<br>audit/function-strings.md<br>audit/function-triage.md |
| `0E93` | sub_E93 | — | 62 | 15 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1048`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 62 bytes，已逐條讀完） | — |
| `0FF5` | sub_FF5 | — | 9 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov byte ptr es:[di+10Eh], 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `11F7` | sub_11F7 | — | 652 | 246 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0782h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1467` | sub_1467 | — | 280 | 107 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0782h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `14FF` | sub_14FF | — | 47 | 15 | 2 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short near ptr loc_15D2+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 47 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1522` | sub_1522 | — | 48 | 19 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [di+4060h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 48 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 81 | 28 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short sub_1625`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `1625` | sub_1625 | — | 5 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 6`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `162A` | sub_162A | — | 6 | 2 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_145B+2`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1639` | sub_1639 | — | 88 | 39 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 88 bytes，已逐條讀完） | — |
| `164D` | sub_164D | — | 9 | 4 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `169F` | sub_169F | — | 10 | 4 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `16A9` | sub_16A9 | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, [bp-4]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16AE` | sub_16AE | — | 6 | 2 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+11h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `17E7` | sub_17E7 | — | 413 | 159 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0782h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18F4` | sub_18F4 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di+11h], dl`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1994` | sub_1994 | — | 10 | 4 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mul dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `199E` | sub_199E | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [di+4124h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1ECD` | sub_1ECD | — | 1608 | 600 | 3 | 13 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0782h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-index/dos-overlay-19.md<br>audit/function-triage.md |
| `2515` | sub_2515 | — | 146 | 53 | 2 | 1 | ✓ | 已解讀 | exact | 866<br>體質的 HP 加值(retf 2，職業碼)：讀 遠指標(DS:6506h)^[19h](體質，spec 857)，3→−2、4..6→−1、7..14→0、15→+1、16→+2、17..19→戰士類(職業碼 2/3/4)給 體質−0Eh(即 +3/+4/+5)否則 +2。★與 AD&D 一版逐格相符，同時是 +19h=體質 與那組職業編號的第二份佐證。⚠體質 ≥20 時所有分支都不成立，直接回傳未初始化的堆疊殘值 | audit/function-index/pc98-overlay-17.md<br>audit/resident-data-tables.md<br>spec/866-constitution-hp-bonus.md<br>spec/869-constitution-bonus-per-class-slot.md |
| `260C` | sub_260C | — | 278 | 111 | 1 | 6 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `2724` | sub_2724 | — | 285 | 121 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2868` | sub_2868 | — | 3454 | 1290 | 1 | 9 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
| `3680` | sub_3680 | — | 982 | 378 | 1 | 8 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `3A56` | sub_3A56 | — | 323 | 104 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `3B99` | sub_3B99 | — | 223 | 101 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `3C78` | sub_3C78 | — | 15 | 7 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cl, 3`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `3EE3` | sub_3EE3 | — | 1798 | 708 | 1 | 8 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
| `45E9` | sub_45E9 | — | 259 | 93 | 3 | 1 | ✓ | 已解讀 | exact | 869<br>體質加值總和(retf 4，角色)：對八個職業槽 c=0..7，該槽等級 >0 且未達 DS:3EB9h 的 HD 上限時，加一份 byte[427Fh + 體質]；職業碼 2/3/4(戰士類)再依體質 17/18/19-20/21-23/24-25 額外加 1/2/3/4/5；c=4(遊俠)且該槽等級=1 時把累積總和乘 2。★DS:427Fh 是體質加值表(3→−2、4..6→−1、7..14→0、15→+1、16..25→+2)，配上戰士類分支後與 AD&D 一版完整體質表(含 20-25)逐格相符。與 spec 866 在 3..19 完全一致，866 只是做到 19 | audit/function-index/dos-overlay-17.md<br>spec/850-hit-dice-and-effective-level.md<br>spec/869-constitution-bonus-per-class-slot.md |
| `46EC` | sub_46EC | — | 218 | 88 | 1 | 2 | ✓ | 已解讀 | exact | 850<br>有效等級(retf 4)：各職業等級總和（Ranger c=4 與 Monk c=7 各多算 1）加上 45E9h 的調整後除以職業數取平均。負調整時有下限保護。⚠ 職業數 = 0 會除以零 | audit/function-index/pc98-overlay-17.md<br>spec/850-hit-dice-and-effective-level.md |
| `47C6` | sub_47C6 | — | 588 | 242 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `4A12` | sub_4A12 | — | 256 | 98 | 2 | 2 | ✓ | 已解讀 | exact | 850<br>升級擲 HD(retf 6，(角色, 職業遮罩))：DS:0818h 骰數 / DS:0820h 面數 / DS:3EA8h 遮罩 / DS:3EB9h HD 上限，四張 8 bytes 平行表 = AD&D 生命骰表。等級 > 1 就只擲 1 顆；擲兩次取大（非 AD&D 規則）。過上限給固定 3/2/1 但用 mov 覆寫而非累加 | audit/function-index/pc98-overlay-17.md<br>spec/850-hit-dice-and-effective-level.md |
| `4B12` | sub_4B12 | — | 344 | 126 | 1 | 1 | ✓ | 已解讀 | exact | 1<br>付錢與找零(retf 8，(角色, 金額 longint))：金額先 × 0C8h(200) 換成銅幣（AD&D 1 gp = 200 cp）。第一段 i 由 0 往上，每種幣值取 min(金額 div 幣值[i] + 1, 持有量) 扣掉——+1 造成刻意多付；第二段若金額變負就取絕對值，i 由 4 往下把差額以大面額找回。幣值表是 DS:0CA2h（1/10/100/200/1000 銅幣）。⚠ 角色身上錢不夠時 n 恆為 0、金額不減，i 會一路遞增讀出幣值表外，可能除以零 | audit/function-index/pc98-overlay-17.md |
| `4D2A` | sub_4D2A | — | 2450 | 756 | 2 | 10 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `56BC` | sub_56BC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
