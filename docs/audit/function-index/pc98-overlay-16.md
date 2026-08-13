# pc98-overlay-16 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADLOADSAVE | 32 | 10 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0020` | sub_20 | — | 87 | 36 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0079` | sub_79 | — | 214 | 84 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0164` | sub_164 | — | 1176 | 456 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0614` | sub_614 | LOADCHARLIST | 450 | 204 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `07DC` | sub_7DC | — | 93 | 40 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `08E4` | sub_8E4 | — | 453 | 198 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B4F` | sub_B4F | CHECKSAVEDRIVE | 920 | 370 | 6 | 4 | ✓ | 待解讀 | — | — | — |
| `0ECE` | sub_ECE | — | 24 | 9 | 3 | 1 |  | 待解讀 | — | — | — |
| `0ED3` | sub_ED3 | — | 46 | 18 | 3 | 1 |  | 待解讀 | — | — | — |
| `0F17` | sub_F17 | MAKEPLAYERICON | 241 | 87 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0FF7` | sub_FF7 | — | 94 | 33 | 2 | 1 |  | 待解讀 | — | — | — |
| `1061` | sub_1061 | LOADACTIVEICON | 498 | 183 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1261` | sub_1261 | DELETECHARACTER | 289 | 117 | 2 | 0 | ✓ | 待解讀 | — | — | — |
| `1437` | sub_1437 | — | 70 | 23 | 1 | 0 |  | 待解讀 | — | — | — |
| `14C9` | sub_14C9 | SAVECHARACTER | 26 | 12 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:649h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | — |
| `15A1` | sub_15A1 | — | 90 | 39 | 2 | 3 |  | 待解讀 | — | — | — |
| `15F5` | sub_15F5 | — | 50 | 22 | 3 | 0 |  | 待解讀 | — | — | — |
| `1627` | sub_1627 | — | 1221 | 480 | 2 | 5 |  | 待解讀 | — | — | audit/function-triage.md |
| `166F` | sub_166F | — | 10 | 5 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 5 | 3 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `167E` | sub_167E | — | 11 | 3 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 723h:48Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1B0D` | sub_1B0D | — | 264 | 109 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1C15` | sub_1C15 | — | 1152 | 449 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `2095` | sub_2095 | — | 374 | 112 | 3 | 1 |  | 待解讀 | — | — | — |
| `26CF` | sub_26CF | — | 798 | 241 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `2A6D` | sub_2A6D | LOADCHARACTER | 3141 | 1140 | 1 | 11 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `36B2` | sub_36B2 | — | 953 | 346 | 2 | 3 |  | 待解讀 | — | — | — |
| `3AA8` | sub_3AA8 | LOADMONSTER | 850 | 324 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `3DFF` | sub_3DFF | LOADNPC | 108 | 42 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3E8B` | sub_3E8B | ATTACHCHARACTER | 300 | 98 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `3FB7` | sub_3FB7 | — | 115 | 53 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `402A` | sub_402A | — | 303 | 120 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `4193` | sub_4193 | — | 183 | 77 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `424A` | sub_424A | — | 51 | 27 | 2 | 0 | ✓ | 待解讀 | — | — | — |
| `42FE` | sub_42FE | LOADSAVEDGAME | 772 | 317 | 0 | 12 | ✓ | 待解讀 | — | — | — |
| `46CE` | sub_46CE | SAVECURRENTGAME | 659 | 279 | 0 | 11 | ✓ | 待解讀 | — | — | — |
| `4961` | sub_4961 | LOADALL | 17 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_49C6`（body 共 17 bytes，已逐條讀完） | — |
| `49C6` | sub_49C6 | LOADGAME | 1183 | 458 | 2 | 9 | ✓ | 待解讀 | — | — | — |
| `4E65` | sub_4E65 | SAVEALL | 28 | 17 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4E81` | sub_4E81 | — | 23 | 9 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp+var_4], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | — |
| `5008` | sub_5008 | SAVEGAME | 718 | 296 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `5581` | sub_5581 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `5588` | sub_5588 | — | 236 | 90 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `5674` | sub_5674 | — | 196 | 93 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `5738` | sub_5738 | — | 23 | 12 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `574F` | sub_574F | — | 31 | 18 | 3 | 0 |  | 待解讀 | — | — | — |
