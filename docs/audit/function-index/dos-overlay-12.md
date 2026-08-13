# dos-overlay-12 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 27 | 9 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0000h（entry#2）助憶碼序列完全相同，語意同該筆：unit 初始化:依序呼叫四個本 overlay 內的 routine ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-12.md<br>audit/string-pairs.md<br>spec/569-small-function-batch-reading.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `001B` | sub_1B | — | 33 | 12 | 17 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:001Bh（entry#3）助憶碼序列完全相同，語意同該筆：傷害取消 routine(spec 412 稱 Protected):參數非 0 且等於 DS:A02Dh 時,把 DS:A02Eh 與 DS:A02Dh 都歸零 ⇒ A02Dh 是傷害來源 ID、A02Eh 是傷害值 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>audit/function-index/pc98-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `003C` | sub_3C | — | 57 | 22 | 6 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0075` | sub_75 | — | 24 | 10 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0075h（entry#5）助憶碼序列完全相同，語意同該筆：以 arg_6／arg_8 呼叫外部 routine 並檢查回傳值(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `008D` | sub_8D | — | 17 | 7 | 1 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:008Dh（entry#6）助憶碼序列完全相同，語意同該筆：DS:A02Ch 加一、DS:A039h(命中骰)加一(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `009E` | sub_9E | — | 18 | 7 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:009Eh（entry#7）助憶碼序列完全相同，語意同該筆：DS:A03Ch 加 5、DS:A039h(命中骰)加一(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `00B0` | sub_B0 | — | 32 | 11 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:00B0h（entry#8）助憶碼序列完全相同，語意同該筆：DS:A03Ch 小於 5 時歸零,否則減 5;並將 DS:A039h(命中骰)減一 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `00E8` | sub_E8 | — | 131 | 49 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `016B` | sub_16B | — | 29 | 11 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0166h（entry#10）助憶碼序列完全相同，語意同該筆：若 DS:9594h 所指 record 的 +14Ch bit0 為 1,則 DS:A039h(命中骰)減 7 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md |
| `0188` | sub_188 | — | 105 | 35 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `01F1` | sub_1F1 | — | 71 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0238` | sub_238 | — | 55 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `026F` | sub_26F | — | 55 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `02A6` | sub_2A6 | — | 37 | 16 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:029Bh（entry#15）助憶碼序列完全相同，語意同該筆：傷害旗標 bit 1 非零時:DS:A02Eh 有號減半,並把 DS:A02Ch 加 3 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `02CB` | sub_2CB | — | 193 | 62 | 0 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `038C` | sub_38C | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `03A0` | sub_3A0 | — | 60 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `03DC` | sub_3DC | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `03E5` | sub_3E5 | — | 94 | 36 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `0443` | sub_443 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | audit/function-index/dos-overlay-12.md |
| `044C` | sub_44C | — | 45 | 14 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0443h（entry#22）助憶碼序列完全相同，語意同該筆：目標 record 的 +19Bh 小於 39h 時夾到 39h,並把 DS:A02Ch 加一;若 DS:A031h 等於 0Fh 則 DS:A02Eh 歸零 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0479` | sub_479 | — | 52 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `04AD` | sub_4AD | — | 43 | 18 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:04A4h（entry#24）助憶碼序列完全相同，語意同該筆：arg_0 為 0 且傷害旗標 bit 0 非零時:DS:A02Eh 有號減半,DS:A02Ch 加 3 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md<br>audit/string-pairs.md |
| `04E4` | sub_4E4 | — | 85 | 30 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `054A` | sub_54A | — | 94 | 39 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `05B6` | sub_5B6 | — | 316 | 108 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `06F2` | sub_6F2 | — | 53 | 20 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0727` | sub_727 | — | 52 | 18 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0769` | sub_769 | — | 131 | 55 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `07EC` | sub_7EC | — | 32 | 15 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:07F6h（entry#32）助憶碼序列完全相同，語意同該筆：DS:A02Eh := DS:A02Eh − (DS:A02Eh ÷ 4) ⇒ 取四分之三 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0818` | sub_818 | — | 171 | 54 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `08CD` | sub_8CD | — | 181 | 59 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0982` | sub_982 | — | 37 | 11 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0994h（entry#35）助憶碼序列完全相同，語意同該筆：DS:A039h(命中骰)減 4、arg_6 所指 record 的 +19Bh 與 +19Ch 各減 4、DS:A02Ch 減 4 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `09A7` | sub_9A7 | — | 57 | 23 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0A0E` | sub_A0E | — | 417 | 168 | 0 | 8 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BAF` | sub_BAF | — | 19 | 7 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0BC8h（entry#38）助憶碼序列完全相同，語意同該筆：DS:A039h(命中骰)減 4、DS:A02Ch 減 4(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0BC2` | sub_BC2 | — | 34 | 11 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:0BDBh（entry#39）助憶碼序列完全相同，語意同該筆：若 arg_6 → +18Eh 所指 record 的 +3 為 0,則 DS:A035h := 1 且 DS:A039h(命中骰) := 0FFh ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0BE9` | sub_BE9 | — | 93 | 37 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0C61` | sub_C61 | — | 718 | 268 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0F39` | sub_F39 | — | 113 | 41 | 3 | 4 | ✓ | 待解讀 | — | — | — |
| `0FAA` | sub_FAA | — | 101 | 36 | 8 | 2 | ✓ | 待解讀 | — | — | — |
| `100F` | sub_100F | — | 7 | 3 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr ds:6508h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `104C` | sub_104C | — | 23 | 11 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:106Dh（entry#45）助憶碼序列完全相同，語意同該筆：DS:A030h := DS:A030h ÷ 2(有號)(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `106F` | sub_106F | — | 138 | 60 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `10F9` | sub_10F9 | — | 145 | 59 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `118A` | sub_118A | — | 77 | 24 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `11D7` | sub_11D7 | — | 41 | 14 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:11F9h（entry#49）助憶碼序列完全相同，語意同該筆：DS:9594h 所指 record 的 +11Ah 等於 1、且 +0DEh 的低 7 位元等於 2 時,DS:A039h(命中骰)減 4 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1200` | sub_1200 | — | 75 | 29 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `124B` | sub_124B | — | 51 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `127E` | sub_127E | — | 51 | 21 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `12B1` | sub_12B1 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `12BA` | sub_12BA | — | 33 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr loc_1466+1`（body 共 33 bytes，已逐條讀完） | — |
| `12DB` | sub_12DB | — | 34 | 11 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:12FDh（entry#55）助憶碼序列完全相同，語意同該筆：把 arg_6 → +18Eh 所指 record 的 +6 清 0;若 DS:A034h 非零則 DS:A030h := 0 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `12FD` | sub_12FD | — | 32 | 16 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:131Fh（entry#56）助憶碼序列完全相同，語意同該筆：以 (0, 0FFh, 0, 62h, arg_6, arg_8) 呼叫 sub_1437 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/dos-overlay-12.md<br>spec/573-effprocs-effect-handlers-first-batch.md |
| `131D` | sub_131D | — | 87 | 32 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1374` | sub_1374 | — | 91 | 36 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `13CF` | sub_13CF | — | 70 | 30 | 0 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1415` | sub_1415 | — | 32 | 14 | 0 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push cs`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 32 bytes，已逐條讀完） | — |
| `1435` | sub_1435 | — | 5 | 2 | 8 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov sp, bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `143A` | sub_143A | — | 4 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 4 bytes，已逐條讀完） | — |
| `1454` | sub_1454 | — | 6 | 3 | 4 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub sp, 0Ch`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `145D` | sub_145D | — | 5 | 1 | 15 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1462` | sub_1462 | — | 6 | 2 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov dx, es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1494` | sub_1494 | — | 5 | 3 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1499` | sub_1499 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:634h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `149E` | sub_149E | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `14A3` | sub_14A3 | — | 11 | 3 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 542h:0B33h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `14F9` | sub_14F9 | — | 1 | 1 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `14FA` | sub_14FA | — | 17 | 5 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les ax, es:[di+0Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 17 bytes，已逐條讀完） | — |
| `1536` | sub_1536 | — | 10 | 6 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1540` | sub_1540 | — | 5 | 1 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:634h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1545` | sub_1545 | — | 5 | 1 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1572`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `154A` | sub_154A | — | 8 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr ds:6FA3h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1554` | sub_1554 | — | 6 | 3 | 19 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ch`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1569` | sub_1569 | — | 9 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push [bp+arg_8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1572` | sub_1572 | — | 7 | 4 | 4 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call sub_1454`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `157F` | sub_157F | — | 3 | 2 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov bp, sp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 3 bytes，已逐條讀完） | — |
| `158B` | sub_158B | — | 10 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 10 bytes，已逐條讀完） | — |
| `1595` | sub_1595 | — | 6 | 3 | 3 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push [bp+arg_A]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `15AB` | sub_15AB | — | 30 | 15 | 0 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call sub_14F9`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `15E6` | sub_15E6 | — | 115 | 42 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1687` | sub_1687 | — | 59 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `16C2` | sub_16C2 | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1454`（body 共 22 bytes，已逐條讀完） | — |
| `16D8` | sub_16D8 | — | 19 | 7 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:1713h（entry#70）助憶碼序列完全相同，語意同該筆：DS:A035h := 1、DS:A039h(命中骰)減 4(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `16EB` | sub_16EB | — | 48 | 23 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:1726h（entry#71）助憶碼序列完全相同，語意同該筆：擲 d100;結果 <= 5Fh(95) 時以 (0, 0Ch, 1, 19h, arg_6, arg_8) 呼叫 sub_1437 ⇒ 95% 機率套用該效果 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1729` | sub_1729 | — | 60 | 28 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `1765` | sub_1765 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 12 bytes，已逐條讀完） | — |
| `1771` | sub_1771 | — | 87 | 32 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `17C8` | sub_17C8 | — | 52 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1809` | sub_1809 | — | 165 | 52 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `18AE` | sub_18AE | — | 127 | 40 | 2 | 2 |  | 待解讀 | — | — | — |
| `192D` | sub_192D | — | 3 | 2 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov bp, sp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 3 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1930` | sub_1930 | — | 6 | 2 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+0Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `196D` | sub_196D | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `19A2` | sub_19A2 | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `19D7` | sub_19D7 | — | 23 | 11 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:1A1Dh（entry#80）助憶碼序列完全相同，語意同該筆：DS:A02Eh := DS:A02Eh ÷ 2(有號)(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `19EE` | sub_19EE | — | 87 | 37 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1A65` | sub_1A65 | — | 439 | 162 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `1C1C` | sub_1C1C | — | 51 | 22 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `1C4F` | sub_1C4F | — | 65 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1C90` | sub_1C90 | — | 75 | 26 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1CF6` | sub_1CF6 | — | 718 | 268 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1FC4` | sub_1FC4 | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:202Ch（entry#87）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 0 非零時,DS:A02Eh 有號除以 2(減半) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1FE4` | sub_1FE4 | — | 77 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `203C` | sub_203C | — | 60 | 23 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2078` | sub_2078 | — | 130 | 49 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `20FA` | sub_20FA | — | 50 | 16 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `212C` | sub_212C | — | 169 | 61 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `21D5` | sub_21D5 | — | 61 | 29 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2212` | sub_2212 | — | 82 | 37 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2264` | sub_2264 | — | 63 | 26 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `22A3` | sub_22A3 | — | 123 | 48 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `231E` | sub_231E | — | 22 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_F39`（body 共 22 bytes，已逐條讀完） | — |
| `2334` | sub_2334 | — | 94 | 40 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `2392` | sub_2392 | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_2334`（body 共 16 bytes，已逐條讀完） | — |
| `23A2` | sub_23A2 | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_2334`（body 共 16 bytes，已逐條讀完） | — |
| `23B2` | sub_23B2 | — | 38 | 20 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:2414h（entry#101）助憶碼序列完全相同，語意同該筆：擲 d100(count=1,sides=64h);結果 <= 5Ah(90) 時依序呼叫 1Bh(35h) 與 1Bh(0Bh) ⇒ 90% 機率取消這兩類傷害 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `23D8` | sub_23D8 | — | 23 | 13 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:243Ah（entry#102）助憶碼序列完全相同，語意同該筆：依序以常數 0Bh 與 35h 呼叫 overlay-12 local 1Bh(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `23EF` | sub_23EF | — | 16 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_1B`（body 共 16 bytes，已逐條讀完） | — |
| `23FF` | sub_23FF | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:2461h（entry#104）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 1 非零時呼叫 overlay-12 local 1Bh(0)(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2418` | sub_2418 | — | 35 | 16 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:247Ah（entry#105）助憶碼序列完全相同，語意同該筆：依序呼叫 1Bh(37h) 與 1Bh(34h);若 DS:A041h 為 0 則 DS:A02Ch := 64h ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `243B` | sub_243B | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:249Dh（entry#106）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 0(依 spec 412 為 Fire)非零時呼叫 1Bh(0)(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2454` | sub_2454 | — | 69 | 27 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2499` | sub_2499 | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:24FBh（entry#108）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 2 非零時,DS:A02Eh 有號除以 2(減半) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `24B9` | sub_24B9 | — | 99 | 39 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `251C` | sub_251C | — | 62 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `255A` | sub_255A | — | 64 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `259A` | sub_259A | — | 32 | 15 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:25FCh（entry#112）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 1 非零時,DS:A02Eh 有號除以 2(減半) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `25BA` | sub_25BA | — | 76 | 25 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2606` | sub_2606 | — | 70 | 26 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2657` | sub_2657 | — | 286 | 109 | 0 | 10 | ✓ | 待解讀 | — | — | — |
| `2782` | sub_2782 | — | 120 | 47 | 0 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `27FA` | sub_27FA | — | 53 | 21 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `282F` | sub_282F | — | 38 | 20 | 0 | 3 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:289Dh（entry#118）助憶碼序列完全相同，語意同該筆：擲 d100;結果 <= 1Eh(30) 時依序呼叫 1Bh(0Bh) 與 1Bh(35h) ⇒ 30% 機率取消 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md |
| `2855` | sub_2855 | — | 49 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `289C` | sub_289C | — | 222 | 88 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `297A` | sub_297A | — | 34 | 16 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:29F2h（entry#121）助憶碼序列完全相同，語意同該筆：DS:A02Dh 為 0 且傷害旗標 bit 3 為 0 時直接返回,否則呼叫 1Bh(0) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `299C` | sub_299C | — | 144 | 46 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2A2C` | sub_2A2C | — | 124 | 50 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2AA8` | sub_2AA8 | — | 46 | 25 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:2B20h（entry#124）助憶碼序列完全相同，語意同該筆：依序呼叫 1Bh(8Eh)、1Bh(1Dh)、1Bh(44h);傷害旗標 bit 6 非零時再呼叫 1Bh(0) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2AD6` | sub_2AD6 | — | 57 | 19 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2B0F` | sub_2B0F | — | 25 | 13 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:2B87h（entry#126）助憶碼序列完全相同，語意同該筆：傷害旗標 DS:A02Fh bit 2(依 spec 412 為 Electricity)非零時呼叫 1Bh(0)(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2B28` | sub_2B28 | — | 22 | 8 | 0 | 0 | ✓ | 已解讀 | strong inference | docs/spec/573-effprocs-effect-handlers-first-batch.md<br>與 PC-98 overlay-12:2BA0h（entry#127）助憶碼序列完全相同，語意同該筆：經兩層 far pointer(arg_6 → +18Eh)把目標 record 的 +6 清 0(retf 0Ah,5 個 word 參數) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2B3E` | sub_2B3E | — | 240 | 73 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2C2E` | sub_2C2E | — | 33 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr loc_1466+1`（body 共 33 bytes，已逐條讀完） | — |
| `2C4F` | sub_2C4F | — | 69 | 28 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2C94` | sub_2C94 | — | 57 | 17 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2CD9` | sub_2CD9 | — | 131 | 49 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2D7D` | sub_2D7D | — | 182 | 71 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2E33` | sub_2E33 | — | 9 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `2E3C` | sub_2E3C | — | 1840 | 569 | 0 | 0 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `356C` | sub_356C | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
