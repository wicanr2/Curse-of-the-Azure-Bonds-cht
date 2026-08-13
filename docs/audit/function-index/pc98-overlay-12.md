# pc98-overlay-12 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADEFFPROCS | 27 | 9 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>unit 初始化:依序呼叫四個本 overlay 內的 routine | spec/569-small-function-batch-reading.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `001B` | sub_1B | — | 33 | 12 | 17 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害取消 routine(spec 412 稱 Protected):參數非 0 且等於 DS:A02Dh 時,把 DS:A02Eh 與 DS:A02Dh 都歸零 ⇒ A02Dh 是傷害來源 ID、A02Eh 是傷害值 | audit/function-index/pc98-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `003C` | sub_3C | — | 57 | 22 | 6 | 2 | ✓ | 待解讀 | — | — | — |
| `0075` | sub_75 | — | 24 | 10 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>以 arg_6／arg_8 呼叫外部 routine 並檢查回傳值(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `008D` | sub_8D | — | 17 | 7 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A02Ch 加一、DS:A039h(命中骰)加一(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `009E` | sub_9E | — | 18 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A03Ch 加 5、DS:A039h(命中骰)加一(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `00B0` | sub_B0 | — | 32 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A03Ch 小於 5 時歸零,否則減 5;並將 DS:A039h(命中骰)減一 | spec/573-effprocs-effect-handlers-first-batch.md |
| `00E3` | sub_E3 | — | 131 | 49 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `0166` | sub_166 | — | 29 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>若 DS:9594h 所指 record 的 +14Ch bit0 為 1,則 DS:A039h(命中骰)減 7 | spec/573-effprocs-effect-handlers-first-batch.md |
| `0183` | sub_183 | — | 99 | 34 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `01E6` | sub_1E6 | — | 71 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `022D` | sub_22D | — | 55 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0264` | sub_264 | — | 55 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `029B` | sub_29B | — | 37 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 bit 1 非零時:DS:A02Eh 有號減半,並把 DS:A02Ch 加 3 | spec/573-effprocs-effect-handlers-first-batch.md |
| `02C0` | sub_2C0 | — | 193 | 62 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0381` | sub_381 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `0397` | sub_397 | — | 60 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `03D3` | sub_3D3 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `03DC` | sub_3DC | — | 94 | 36 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `043A` | sub_43A | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `0443` | sub_443 | — | 45 | 14 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0470` | sub_470 | — | 52 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `04A4` | sub_4A4 | — | 43 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `04E2` | sub_4E2 | — | 85 | 30 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `054C` | sub_54C | — | 94 | 39 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `05BD` | sub_5BD | — | 316 | 108 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `06F9` | sub_6F9 | — | 53 | 20 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `072E` | sub_72E | — | 52 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0773` | sub_773 | — | 131 | 55 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `07F6` | sub_7F6 | — | 32 | 15 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A02Eh := DS:A02Eh − (DS:A02Eh ÷ 4) ⇒ 取四分之三 | spec/573-effprocs-effect-handlers-first-batch.md |
| `0829` | sub_829 | — | 171 | 54 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `08DF` | sub_8DF | — | 181 | 59 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0994` | sub_994 | — | 37 | 11 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A039h(命中骰)減 4、arg_6 所指 record 的 +19Bh 與 +19Ch 各減 4、DS:A02Ch 減 4 | spec/573-effprocs-effect-handlers-first-batch.md |
| `09B9` | sub_9B9 | — | 57 | 23 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0A32` | sub_A32 | — | 406 | 167 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0BC8` | sub_BC8 | — | 19 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A039h(命中骰)減 4、DS:A02Ch 減 4(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `0BDB` | sub_BDB | — | 34 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>若 arg_6 → +18Eh 所指 record 的 +3 為 0,則 DS:A035h := 1 且 DS:A039h(命中骰) := 0FFh | spec/573-effprocs-effect-handlers-first-batch.md |
| `0C0C` | sub_C0C | — | 93 | 37 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0C82` | sub_C82 | — | 718 | 268 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0F5B` | sub_F5B | — | 113 | 41 | 3 | 3 | ✓ | 待解讀 | — | — | — |
| `0FCC` | sub_FCC | — | 27 | 9 | 7 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `or ax, es:[di+154h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 27 bytes，已逐條讀完） | — |
| `1030` | sub_1030 | — | 61 | 25 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `106D` | sub_106D | — | 23 | 11 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A030h := DS:A030h ÷ 2(有號)(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `1091` | sub_1091 | — | 138 | 60 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `111B` | sub_111B | — | 145 | 59 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `11AC` | sub_11AC | — | 77 | 24 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `11F9` | sub_11F9 | — | 41 | 14 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1222` | sub_1222 | — | 75 | 29 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `126D` | sub_126D | — | 51 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `12A0` | sub_12A0 | — | 51 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `12D3` | sub_12D3 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `12DC` | sub_12DC | — | 33 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1437`（body 共 33 bytes，已逐條讀完） | — |
| `12FD` | sub_12FD | — | 34 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>把 arg_6 → +18Eh 所指 record 的 +6 清 0;若 DS:A034h 非零則 DS:A030h := 0 | spec/573-effprocs-effect-handlers-first-batch.md |
| `131F` | sub_131F | — | 32 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>以 (0, 0FFh, 0, 62h, arg_6, arg_8) 呼叫 sub_1437 | spec/573-effprocs-effect-handlers-first-batch.md |
| `133F` | sub_133F | — | 87 | 32 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1396` | sub_1396 | — | 91 | 36 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `13F1` | sub_13F1 | — | 20 | 9 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 20 bytes，已逐條讀完） | — |
| `1405` | sub_1405 | — | 5 | 3 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `140A` | sub_140A | — | 5 | 2 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `or al, al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `140F` | sub_140F | — | 11 | 5 | 9 | 1 |  | 待解讀 | — | — | — |
| `1414` | sub_1414 | — | 27 | 12 | 2 | 2 |  | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>以 (arg, 1) 呼叫 far sub_146E,回傳非零才續行後段處理 | spec/573-effprocs-effect-handlers-first-batch.md |
| `1437` | sub_1437 | — | 41 | 18 | 15 | 2 | ✓ | 待解讀 | — | — | — |
| `1464` | sub_1464 | — | 10 | 2 | 7 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `rep sub byte ptr [bx-427Eh], 81h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `146E` | sub_146E | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `or al, [bp+si-7133h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `147A` | sub_147A | — | 86 | 34 | 4 | 2 | ✓ | 待解讀 | — | — | — |
| `14CA` | sub_14CA | — | 18 | 4 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr ds:0A03Dh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 18 bytes，已逐條讀完） | — |
| `1515` | sub_1515 | — | 9 | 3 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor byte ptr [di-427Eh], 81h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `151F` | sub_151F | — | 6 | 3 | 4 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 0Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 5 | 3 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1547` | sub_1547 | — | 10 | 4 | 9 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jnz short loc_1589`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1551` | sub_1551 | — | 10 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `155B` | sub_155B | — | 10 | 6 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1565` | sub_1565 | — | 6 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:62Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1574` | sub_1574 | — | 27 | 12 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1437`（body 共 27 bytes，已逐條讀完） | — |
| `158F` | sub_158F | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_147A`（body 共 22 bytes，已逐條讀完） | — |
| `15A5` | sub_15A5 | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_147A`（body 共 22 bytes，已逐條讀完） | — |
| `15BB` | sub_15BB | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_147A`（body 共 22 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 36 | 18 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>以 (8, 2, arg_6, arg_8) 呼叫 far routine,回傳值零延伸後與 0 一起傳給 sub_151F | spec/573-effprocs-effect-handlers-first-batch.md |
| `1621` | sub_1621 | — | 1 | 1 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `1622` | sub_1622 | — | 5 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 1Ch`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1627` | sub_1627 | — | 155 | 54 | 2 | 3 |  | 待解讀 | — | — | audit/function-triage.md |
| `16C2` | sub_16C2 | — | 59 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `16FD` | sub_16FD | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_147A`（body 共 22 bytes，已逐條讀完） | — |
| `1713` | sub_1713 | — | 19 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A035h := 1、DS:A039h(命中骰)減 4(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `1726` | sub_1726 | — | 48 | 23 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `176D` | sub_176D | — | 60 | 28 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `17A9` | sub_17A9 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 12 bytes，已逐條讀完） | — |
| `17B5` | sub_17B5 | — | 87 | 32 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `180C` | sub_180C | — | 52 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `184F` | sub_184F | — | 212 | 66 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1973` | sub_1973 | — | 64 | 26 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `19B3` | sub_19B3 | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `19E8` | sub_19E8 | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1A1D` | sub_1A1D | — | 23 | 11 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A02Eh := DS:A02Eh ÷ 2(有號)(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `1A34` | sub_1A34 | — | 87 | 37 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1ABC` | sub_1ABC | — | 458 | 169 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `1C86` | sub_1C86 | — | 51 | 22 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1CB9` | sub_1CB9 | — | 65 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1CFA` | sub_1CFA | — | 75 | 26 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1D5E` | sub_1D5E | — | 718 | 268 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `202C` | sub_202C | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 0 非零時,DS:A02Eh 有號除以 2(減半) | spec/573-effprocs-effect-handlers-first-batch.md |
| `204C` | sub_204C | — | 77 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `20AA` | sub_20AA | — | 60 | 23 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `20E6` | sub_20E6 | — | 118 | 48 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `215C` | sub_215C | — | 50 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `218E` | sub_218E | — | 169 | 61 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2237` | sub_2237 | — | 61 | 29 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2274` | sub_2274 | — | 82 | 37 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `22C6` | sub_22C6 | — | 63 | 26 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2305` | sub_2305 | — | 123 | 48 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2380` | sub_2380 | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_F5B`（body 共 22 bytes，已逐條讀完） | — |
| `2396` | sub_2396 | — | 94 | 40 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `23F4` | sub_23F4 | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_2396`（body 共 16 bytes，已逐條讀完） | — |
| `2404` | sub_2404 | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_2396`（body 共 16 bytes，已逐條讀完） | — |
| `2414` | sub_2414 | — | 38 | 20 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `243A` | sub_243A | — | 23 | 13 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>依序以常數 0Bh 與 35h 呼叫 overlay-12 local 1Bh(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `2451` | sub_2451 | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_1B`（body 共 16 bytes，已逐條讀完） | — |
| `2461` | sub_2461 | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 1 非零時呼叫 overlay-12 local 1Bh(0)(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `247A` | sub_247A | — | 35 | 16 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>依序呼叫 1Bh(37h) 與 1Bh(34h);若 DS:A041h 為 0 則 DS:A02Ch := 64h | spec/573-effprocs-effect-handlers-first-batch.md |
| `249D` | sub_249D | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 0(依 spec 412 為 Fire)非零時呼叫 1Bh(0)(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `24B6` | sub_24B6 | — | 69 | 27 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `24FB` | sub_24FB | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 2 非零時,DS:A02Eh 有號除以 2(減半) | spec/573-effprocs-effect-handlers-first-batch.md |
| `251B` | sub_251B | — | 99 | 39 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `257E` | sub_257E | — | 62 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `25BC` | sub_25BC | — | 64 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `25FC` | sub_25FC | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 1 非零時,DS:A02Eh 有號除以 2(減半) | spec/573-effprocs-effect-handlers-first-batch.md |
| `261C` | sub_261C | — | 76 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2668` | sub_2668 | — | 70 | 26 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `26BD` | sub_26BD | — | 294 | 112 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `27F0` | sub_27F0 | — | 120 | 47 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2868` | sub_2868 | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `289D` | sub_289D | — | 38 | 20 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `28C3` | sub_28C3 | — | 49 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `290C` | sub_290C | — | 230 | 91 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `29F2` | sub_29F2 | — | 34 | 16 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>DS:A02Dh 為 0 且傷害旗標 bit 3 為 0 時直接返回,否則呼叫 1Bh(0) | spec/573-effprocs-effect-handlers-first-batch.md |
| `2A14` | sub_2A14 | — | 144 | 46 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2AA4` | sub_2AA4 | — | 124 | 50 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2B20` | sub_2B20 | — | 46 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2B4E` | sub_2B4E | — | 57 | 19 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2B87` | sub_2B87 | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>傷害旗標 DS:A02Fh bit 2(依 spec 412 為 Electricity)非零時呼叫 1Bh(0)(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `2BA0` | sub_2BA0 | — | 22 | 8 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>經兩層 far pointer(arg_6 → +18Eh)把目標 record 的 +6 清 0(retf 0Ah,5 個 word 參數) | spec/573-effprocs-effect-handlers-first-batch.md |
| `2BB6` | sub_2BB6 | — | 240 | 73 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2CA6` | sub_2CA6 | — | 33 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1437`（body 共 33 bytes，已逐條讀完） | — |
| `2CC7` | sub_2CC7 | — | 69 | 28 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2D0C` | sub_2D0C | — | 57 | 17 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2D56` | sub_2D56 | — | 131 | 49 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2E15` | sub_2E15 | — | 182 | 71 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `2ECB` | sub_2ECB | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `2ED4` | sub_2ED4 | INITEFFPROX | 1840 | 569 | 0 | 0 | ✓ | 待解讀 | — | — | audit/function-triage.md |
