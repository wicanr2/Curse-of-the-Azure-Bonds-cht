# dos-overlay-08 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 77 | 19 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 14 個呼叫，沒有其他動作：`call far ptr sub_104A`、`call far ptr 167h:4Dh`、`call far ptr 187h:43h`、`call far ptr 18Ch:93h`、`call far ptr 196h:43h`、`call far ptr 14Dh:101h`（body 共 77 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/748-cloud-effect-dispel-pair.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `004D` | sub_4D | — | 166 | 62 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/749-combat-teardown-and-battlefield-grid.md<br>戰鬥收尾(retf，不收參數)：逐節點清空 DS:755Bh 與 DS:755Fh 兩條雲霧鏈(每個 FreeMem 30 bytes)；FreeMem(DS:6E92h, 4E9h) 並歸零——4E9h=1257 剛好等於 24*50+49+7+1，因此戰場格陣是 50 欄×25 列、前 7 byte 為表頭；呼叫未辨識的 resident 0297:1D19h(@DS:75E0h) 與 overlay-32 entry#5；最後把 DS:72A0h 設回 overlay-22 entry#4 @10D2h(spec 719 的選目標程序)，可見 72A0h slot 0 是可換掉的掛鉤而非唯讀分派表 | audit/function-index/dos-overlay-08.md<br>spec/748-cloud-effect-dispel-pair.md<br>spec/749-combat-teardown-and-battlefield-grid.md<br>spec/769-combat-main-loop.md |
| `00F3` | sub_F3 | — | 184 | 67 | 0 | 7 | ✓ | 已解讀 | exact | docs/spec/769-combat-main-loop.md<br>戰鬥主迴圈(retf，無參數)：DS:4FBAh := 5、把 DS:72A0h 掛鉤指到 overlay-13 entry#18 @2220h(戰鬥版選目標，spec 749 記錄的收尾會換回 overlay-22 entry#4)；迴圈直到 DS:6F90h 或 DS:6F91h 為 0 或 本模組 0997h 設旗標，每輪走隊伍鏈、清 DS:4F9Dh^[596h]、用 01ABh 取行動者再 026Bh 讓它行動；離開後叫 本模組 004Dh(spec 749 的戰鬥收尾)並設 DS:4FC9h := 1。開場(overlay-10:1C3Eh)、主迴圈、收尾三段至此接成一條鏈 | spec/769-combat-main-loop.md |
| `01AB` | sub_1AB | — | 192 | 68 | 1 | 1 | ✓ | 待解讀 | — | — | spec/769-combat-main-loop.md |
| `026B` | sub_26B | — | 325 | 102 | 1 | 3 | ✓ | 待解讀 | — | — | spec/769-combat-main-loop.md |
| `03D5` | sub_3D5 | — | 810 | 302 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `076C` | sub_76C | — | 489 | 192 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0885` | sub_885 | — | 9 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+111h], 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | spec/769-combat-main-loop.md |
| `0997` | sub_997 | — | 317 | 112 | 1 | 3 | ✓ | 待解讀 | — | — | spec/769-combat-main-loop.md |
| `0B06` | sub_B06 | — | 415 | 152 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-08.md |
| `0C4E` | sub_C4E | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov es:[di], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0C6C` | sub_C6C | — | 11 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 18Ch:61h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `0C94` | sub_C94 | — | 2 | 1 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_CE0`，控制權轉交後不返回（body 共 2 bytes，已逐條讀完） | spec/769-combat-main-loop.md |
| `0CDA` | sub_CDA | — | 472 | 174 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0B06h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0EE9` | sub_EE9 | — | 184 | 67 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/769-combat-main-loop.md<br>攻擊前的兩道檢查(retf 0Ch，三個遠指標，第一個是輸出旗標)：<014Dh:00EDh>(攻擊者) 為真且 <014Dh:00F2h>(攻擊者) 為假時顯示 CS:0ED4h 'Not with that weapon'；否則 本模組 0C94h(目標, 攻擊者) 為真才把 攻擊者^[18Dh]^[0Ah] := 目標(⚠ 直接寫入不還原，與 spec 766 的借用後還原不同)，再依 <呼叫>(目標, 攻擊者) 分兩條路 | spec/769-combat-main-loop.md |
| `0FE8` | sub_FE8 | — | 83 | 33 | 1 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 83 bytes，已逐條讀完） | audit/function-index/dos-overlay-08.md |
| `1009` | sub_1009 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `104A` | sub_104A | — | 98 | 41 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FE8h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `11DF` | sub_11DF | — | 309 | 133 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1314` | sub_1314 | — | 81 | 24 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>清掉同陣營的目標(retf 4)：p^[198h]:=1；若 p^[18Dh]^[0Ah](目前目標)非 NIL 且其 +197h(陣營)與 p^[197h] 相同，就把目標設回 NIL | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-index/pc98-overlay-08.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/766-target-swap-idiom-slot-sort-and-percent-opcode.md |
| `1365` | sub_1365 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
