# dos-START.EXE 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `1002F` | PROGRAM | — | 50 | 10 | 0 | 10 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_120D0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 50 bytes，已逐條讀完） | — |
| `103B0` | sub_103B0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `103E0` | sub_103E0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `103E5` | sub_103E5 | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10420` | sub_10420 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10425` | sub_10425 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1042A` | sub_1042A | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10570` | sub_10570 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10575` | sub_10575 | — | 2 | 1 | 3 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `105B0` | sub_105B0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10620` | sub_10620 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10690` | sub_10690 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `106D0` | sub_106D0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `107A0` | sub_107A0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10810` | sub_10810 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10880` | sub_10880 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10920` | sub_10920 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10925` | sub_10925 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1092A` | sub_1092A | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1092F` | sub_1092F | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10960` | sub_10960 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10C30` | sub_10C30 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10D40` | sub_10D40 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10DB0` | sub_10DB0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10E70` | sub_10E70 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10F00` | sub_10F00 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10F05` | sub_10F05 | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10F19` | sub_10F19 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10F90` | sub_10F90 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10FF0` | sub_10FF0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `110A0` | sub_110A0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11120` | sub_11120 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `111C0` | sub_111C0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11430` | sub_11430 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `114F0` | sub_114F0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11620` | sub_11620 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11638` | sub_11638 | — | 23 | 12 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `int 3Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | — |
| `11690` | sub_11690 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1169F` | sub_1169F | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11720` | sub_11720 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11770` | sub_11770 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `117B0` | sub_117B0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11810` | sub_11810 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1181A` | sub_1181A | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11847` | sub_11847 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11860` | sub_11860 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `11890` | sub_11890 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `118E0` | sub_118E0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11980` | sub_11980 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `119B0` | sub_119B0 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `119E0` | sub_119E0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11A00` | sub_11A00 | — | 66 | 32 | 4 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_120DF`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 66 bytes，已逐條讀完） | — |
| `11ECD` | sub_11ECD | — | 472 | 205 | 3 | 3 |  | 待解讀 | — | — | — |
| `120A5` | sub_120A5 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `120D0` | sub_120D0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `120DF` | sub_120DF | — | 2 | 1 | 24 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `12110` | sub_12110 | — | 111 | 44 | 1 | 7 |  | 已解讀 | exact | docs/spec/672-numeric-menu-input-and-nibble-array.md<br>數字選單輸入:迴圈 ReadKey,兩道檢查——字元要在 CS:byte_120F0h 寫死的集合裡、(c-30h) 要落在 [arg_0..arg_2](用 Pascal 的集合 + 運算建子範圍);任一不過就回到 ReadKey。⚠ 回顯在兩道檢查之後,所以按錯鍵完全沒有回饋、不合法的按鍵永遠不會顯示出來。通過才 Write 並回傳 c-30h | — |
| `1236C` | sub_1236C | — | 1537 | 609 | 1 | 37 |  | 待解讀 | — | — | — |
| `12970` | sub_12970 | — | 123 | 47 | 1 | 1 |  | 已解讀 | exact | docs/spec/674-ega-pixel-mask-and-bit-reverse.md<br>算「哪幾個像素等於某顏色」的 8-bit 遮罩:4 bytes = 8 個 4-bit 像素,每個 byte 先比高 nibble 後比低 nibble,命中就把目前的 bit 加進遮罩(bit 由 80h 起每次右移一位)。bit 7 對應第 0 個像素。正好餵給 EGA 的 Bit Mask 暫存器(3CEh 索引 8,見 spec 673)。高 nibble 在前的順序與 129EBh 一致 | — |
| `129EB` | sub_129EB | — | 92 | 41 | 4 | 2 |  | 已解讀 | exact | docs/spec/672-numeric-menu-input-and-nibble-array.md<br>從 arg_0 指的 8 bytes(= 16 個 nibble)取第 arg_4 個:奇數索引取低 nibble、偶數取高 nibble——**高 nibble 在前**。奇偶判斷用 shr al,1 看進位,但商下面又用 idiv 2 重算一次,同一個除法做了兩遍 | — |
| `12A47` | sub_12A47 | — | 48 | 19 | 2 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>設定 BIOS 顯示模式:把 Registers record 的 AH=0、AL=arg_0 後呼叫 INTR(10h) ⇒ INT 10h AH=00h;呼叫前把 byte_211A5 存進 byte_211A4(前一個模式),呼叫後把 arg_0 記進 byte_211A5 | — |
| `12A77` | sub_12A77 | — | 15 | 9 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>還原先前的 BIOS 顯示模式:以 byte_211A4(前一個模式)呼叫 sub_12A47 | — |
| `12C1D` | sub_12C1D | — | 1611 | 617 | 1 | 5 |  | 待解讀 | — | — | — |
| `133E2` | sub_133E2 | — | 133 | 54 | 2 | 0 |  | 已解讀 | exact | docs/spec/674-ega-pixel-mask-and-bit-reverse.md<br>byte 位元反轉:八段固定的 and + shl/shr(bit0→7、1→6、…、7→0)。已對 0..255 全部逐值驗算,與位元字串反轉不一致 0 筆。第一段沒有先 and 1,直接 shl 7 靠「只存回 byte」把其餘位元擠掉,其他七段都有遮罩 | — |
| `13467` | sub_13467 | — | 141 | 49 | 2 | 0 |  | 已解讀 | exact | docs/spec/675-horizontal-mirror-pair.md<br>16-bit 裡 8 組兩位元的順序反轉:遮罩取自 DS:1E73Dh 起的八個 word,實際讀出是 C000h/3000h/0C00h/0300h/00C0h/0030h/000Ch/0003h(八組互不重疊的兩位元),移位量 ±14/±10/±6/±2。已對 0..65535 全部逐值驗算,不一致 0 筆。與 133E2h(1-bit 版)成一套——同一個水平鏡射在兩種像素深度下的兩個版本。遮罩放資料段可由資料改變,133E2h 的則是寫死的立即值 | — |
| `134F4` | sub_134F4 | — | 49 | 20 | 2 | 0 |  | 已解讀 | exact | docs/spec/671-dos-video-page-and-dead-bounds.md<br>交換一個 byte 的高低 nibble:(arg_0 shl 4) + ((arg_0 and 0F0h) shr 4)。用 add 而非 or(兩半不重疊結果相同);shl 後存進 byte 變數,高位自然被截掉不需額外遮罩 | — |
| `145C9` | sub_145C9 | — | 192 | 72 | 2 | 3 |  | 待解讀 | — | — | — |
| `14689` | sub_14689 | — | 116 | 42 | 2 | 2 |  | 已解讀 | exact | docs/spec/673-ega-graphics-controller-reset.md<br>釋放主記錄與附帶緩衝:size := byte(p^[8]) × word(p^[11h])(記錄自己記著「幾筆 × 每筆多大」);p^[13h] 的 far pointer 非 nil 時先 FreeMem(@p^[13h], size),再 FreeMem(主記錄, size + 17h)(前 17h bytes 是標頭,+13h..+16h 就是那個指標),最後把呼叫端指標清成 nil。兩塊用同一個 size。用的是 Turbo Pascal 的 FreeMem,不是 spec 655 那套自製配置器 | — |
| `14B4C` | sub_14B4C | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `14B53` | sub_14B53 | — | 135 | 49 | 1 | 1 |  | 已解讀 | exact | docs/spec/675-horizontal-mirror-pair.md<br>blit 前的參數計算:arg_A shl 3(×8,一個 byte 裝 8 個像素);src^[2] 與 dst^[2] 是各自的每列位元組數,var_12 = src stride - dst stride - arg_C 即每列要跳過多少;var_10/14/12/C 最後都乘 2,所以 sub_15329 以 word 為單位前進。dst^[11h] 與 14689h 算釋放大小用的是同一個欄位。參數靠 bx/cx 與一整片區域變數傳,不是堆疊參數 | — |
| `14BDA` | sub_14BDA | — | 148 | 52 | 1 | 1 |  | 待解讀 | — | — | — |
| `14C6E` | sub_14C6E | — | 357 | 136 | 1 | 3 |  | 待解讀 | — | — | — |
| `14DD3` | sub_14DD3 | — | 143 | 52 | 1 | 1 |  | 待解讀 | — | — | — |
| `14E62` | sub_14E62 | — | 157 | 56 | 1 | 1 |  | 待解讀 | — | — | — |
| `14EFF` | sub_14EFF | — | 379 | 144 | 1 | 3 |  | 待解讀 | — | — | — |
| `1507A` | sub_1507A | — | 149 | 54 | 1 | 1 |  | 待解讀 | — | — | — |
| `1510F` | sub_1510F | — | 162 | 57 | 1 | 1 |  | 待解讀 | — | — | — |
| `151B1` | sub_151B1 | — | 376 | 143 | 1 | 3 |  | 待解讀 | — | — | — |
| `15329` | sub_15329 | — | 95 | 36 | 7 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 95 bytes，已逐條讀完） | — |
| `15388` | sub_15388 | — | 147 | 54 | 7 | 1 |  | 待解讀 | — | — | — |
| `15420` | sub_15420 | — | 115 | 59 | 1 | 0 |  | 已解讀 | exact | docs/spec/673-ega-graphics-controller-reset.md<br>把 EGA/VGA 繪圖暫存器設回預設:Graphics Controller(3CEh/3CFh)索引 0..5 全寫 0、7 寫 0Fh、8 寫 0FFh,Sequencer(3C4h/3C5h)索引 2 寫 0Fh(四平面全開)。索引 6(Miscellaneous)被跳過——模式本身由別處設定,這裡只清理繪圖行為。⚠ DOS 版走 EGA/VGA 平面圖形,與 PC-98 版的文字 VRAM + 三平面是兩套顯示模型,繪圖程式碼不可能共用 | — |
| `15493` | sub_15493 | — | 49 | 20 | 1 | 1 |  | 已解讀 | exact | docs/spec/671-dos-video-page-and-dead-bounds.md<br>切換 BIOS 顯示頁:把 Registers.AH := 5、AL := arg_0 後呼叫 Turbo Pascal 的 Intr(10h, Regs);同時自己算好 word_2119Eh := (arg_0 shl 9) + 0A000h——shl 9 是 ×512 段落 = 8KB,所以那是該頁的段位址 | — |
| `154C4` | sub_154C4 | — | 25 | 11 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>word_211A0 := (arg_0 << 9) + 0A000h ⇒ 由列號算出顯示記憶體的段位址(A000h 起、每單位 512 bytes) | — |
| `154DD` | sub_154DD | — | 188 | 89 | 4 | 4 |  | 待解讀 | — | — | — |
| `15606` | sub_15606 | — | 299 | 126 | 3 | 6 |  | 待解讀 | — | — | — |
| `15731` | sub_15731 | — | 65 | 30 | 4 | 2 |  | 已解讀 | exact | docs/spec/671-dos-video-page-and-dead-bounds.md<br>範圍檢查後呼叫 <sub_15606>(1,20h,arg_0,arg_2,arg_2,arg_4,arg_6)(arg_2 被推入兩次)。上限 27h(39)與 18h(24)即 40 欄 × 25 列。⚠ 兩個下限檢查(cmp x,0 + jb)是死碼——無號比較,任何值都不小於 0;是原作的冗餘不是判讀不確定,remake 直接寫上限即可 | — |
| `15772` | sub_15772 | — | 133 | 53 | 3 | 3 |  | 待解讀 | — | — | — |
| `157F7` | sub_157F7 | — | 119 | 45 | 3 | 3 |  | 待解讀 | — | — | — |
| `1586E` | sub_1586E | — | 60 | 20 | 1 | 1 |  | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>跳過開頭空白(retf 2)：參數是 SS 相對位址，-103h 是游標、-101h 是長度、-100h 起是內容；游標小於長度且該位置是空白就前進。無號比較 | — |
| `15B42` | sub_15B42 | — | 401 | 171 | 1 | 10 |  | 待解讀 | — | — | — |
| `15D66` | sub_15D66 | — | 78 | 40 | 1 | 4 |  | 已解讀 | exact | docs/spec/759-combatant-xy-lookup-and-bank-select-bit.md<br>顯示一行並等一個鍵(retf 8)：StoreString(參數字串, 緩衝, 28h)；本模組 15731h(0, 18h, 0, 28h)；本模組 15772h(0, 18h, c, b, @緩衝)；ReadKey(結果存進不再讀的 local)。28h/18h 成對，形狀是 40 欄×24 列的文字視窗 | — |
| `15F53` | sub_15F53 | — | 23 | 11 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>呼叫 RTL @DELAY(byte_21169 × 100) | — |
| `15F6A` | sub_15F6A | — | 470 | 207 | 2 | 3 |  | 待解讀 | — | — | — |
| `16140` | sub_16140 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `16147` | sub_16147 | — | 159 | 78 | 1 | 1 |  | 待解讀 | — | — | — |
| `161E6` | sub_161E6 | — | 96 | 46 | 1 | 1 |  | 待解讀 | — | — | — |
| `16246` | sub_16246 | — | 271 | 131 | 1 | 1 |  | 待解讀 | — | — | — |
| `16360` | sub_16360 | — | 29 | 13 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>呼叫 RTL @FreeMem(ptr, (size+7) and 0FFF8h) ⇒ 釋放對齊到 8 bytes 的區塊 | — |
| `1637D` | sub_1637D | — | 29 | 13 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>呼叫 RTL @GetMem(ptr, (size+7) and 0FFF8h) ⇒ 配置對齊到 8 bytes 的區塊 | — |
| `1639A` | sub_1639A | — | 110 | 40 | 2 | 2 |  | 待解讀 | — | — | — |
| `16408` | sub_16408 | — | 112 | 44 | 1 | 3 |  | 待解讀 | — | — | — |
| `16478` | sub_16478 | — | 32 | 17 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_1639A`（body 共 32 bytes，已逐條讀完） | — |
| `16498` | sub_16498 | — | 90 | 48 | 2 | 4 |  | 待解讀 | — | — | — |
| `16566` | sub_16566 | — | 75 | 32 | 1 | 6 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>等待 overlay 磁片(retf)：反覆用 1685Fh 檢查 'game.ovr' 是否存在，不在就印 'Please insert overlay disk.' 加換行、等一個按鍵再試。⚠ 沒有離開的出口 | — |
| `16645` | sub_16645 | — | 538 | 219 | 1 | 11 |  | 待解讀 | — | — | — |
| `1685F` | sub_1685F | — | 74 | 33 | 4 | 3 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>檔案是否存在(retf 4)：把參數字串複製進 50h bytes 緩衝，FindFirst(路徑, 0, 記錄)，回傳 (word_24E58 = 0) and (字串長度 <> 0)。word_24E58 是 DosError | — |
| `168A9` | sub_168A9 | — | 207 | 82 | 1 | 5 |  | 待解讀 | — | — | — |
| `16A62` | sub_16A62 | — | 194 | 83 | 1 | 3 |  | 待解讀 | — | — | — |
| `16B24` | sub_16B24 | — | 243 | 109 | 1 | 7 |  | 待解讀 | — | — | — |
| `16C17` | sub_16C17 | — | 29 | 13 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>依序 @Close 兩個 File 參數(先 arg_4 再 arg_0) | — |
| `16C3E` | sub_16C3E | — | 559 | 235 | 2 | 17 |  | 待解讀 | — | — | — |
| `16E8F` | sub_16E8F | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `16EA0` | sub_16EA0 | — | 81 | 28 | 2 | 5 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call @Halt$q4Word`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `16F03` | sub_16F03 | — | 122 | 51 | 1 | 4 |  | 待解讀 | — | — | — |
| `16FAD` | sub_16FAD | — | 319 | 116 | 6 | 12 |  | 待解讀 | — | — | — |
| `170EC` | sub_170EC | — | 28 | 13 | 3 | 3 |  | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>清空鍵盤緩衝區(retf，無參數)：while KEYPRESSED do 讀一個鍵丟掉(以 push cs + call near 呼叫 sub_16FADh) | — |
| `17122` | sub_17122 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `17150` | sub_17150 | — | 182 | 79 | 10 | 6 |  | 待解讀 | — | — | — |
| `17206` | sub_17206 | — | 58 | 25 | 7 | 4 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>依音效裝置決定是否安裝(retf)：dword_21154 := seg045:0124h；byte_21DAB := byte_21DAA；byte_21DAC := 1；byte_21DAA <> 2 時才叫 19320h() 與 196E0h(0)、1974Ah(0) | — |
| `17240` | sub_17240 | — | 16 | 8 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>把 dword_21154 所指的 word 清成 0 | — |
| `19320` | sub_19320 | — | 53 | 24 | 1 | 0 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>聲音驅動安裝(retf)：存 INT 08h 舊向量到 CS:dword_1931Ch，換成 seg045:225Eh；out 43h,0B6h 後把 13B1h(≈236.7 Hz)分兩次寫 port 40h。⚠ 控制字 0B6h 選的是 counter 2，資料卻寫進 counter 0 | — |
| `19355` | sub_19355 | — | 57 | 27 | 1 | 0 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>聲音驅動卸載(retf)：還原 INT 08h；counter 0 寫回 0FFFFh(18.2 Hz)；in 61h/and 0FCh/out 61h 關喇叭閘；對 port 0C0h 依序寫 9Fh/0BFh/0DFh/0FFh — SN76489(Tandy 三音)四聲道全部最大衰減，即靜音 | — |
| `1941B` | sub_1941B | — | 117 | 40 | 1 | 2 |  | 待解讀 | — | — | — |
| `19490` | sub_19490 | — | 142 | 57 | 1 | 2 |  | 待解讀 | — | — | — |
| `1951E` | sub_1951E | — | 145 | 55 | 2 | 2 |  | 待解讀 | — | — | — |
| `195AF` | sub_195AF | — | 305 | 121 | 1 | 1 |  | 待解讀 | — | — | — |
| `196E0` | sub_196E0 | — | 106 | 43 | 2 | 1 |  | 待解讀 | — | — | — |
| `1974A` | sub_1974A | — | 106 | 43 | 2 | 1 |  | 待解讀 | — | — | — |
| `197B4` | sub_197B4 | — | 20 | 12 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>跨段寫入:切到 seg045 後把 arg_0 寫進 byte_17286,再還原 DS(前後以 push/pop 保存 AX 與 DS) | — |
| `197D0` | sub_197D0 | — | 26 | 14 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：字串指派：把字面值「Wooden」寫入目的字串變數（body 共 26 bytes，已逐條讀完） | — |
| `197F0` | @MSDOS$qm9REGISTERS | — | 11 | 10 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@MSDOS$qm9REGISTERS`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `197FB` | @INTR$q4BYTEm9REGISTERS | — | 58 | 40 | 8 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@INTR$q4BYTEm9REGISTERS`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1985C` | @GetDate$qm4Wordt1t1t1 | — | 34 | 18 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@GetDate$qm4Wordt1t1t1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1987E` | @SetDate$q4Wordt1t1 | — | 20 | 9 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@SetDate$q4Wordt1t1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19892` | @GetTime$qm4Wordt1t1t1 | — | 37 | 19 | 1 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@GetTime$qm4Wordt1t1t1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `198B7` | @SetTime$q4Wordt1t1t1 | — | 23 | 10 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@SetTime$q4Wordt1t1t1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `198CE` | @DiskFree$q4Byte | — | 25 | 12 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@DiskFree$q4Byte`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `198E7` | @DiskSize$q4Byte | — | 27 | 13 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@DiskSize$q4Byte`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19902` | @FINDFIRST$q7PATHSTR4WORDm9SEARCHREC | — | 62 | 33 | 2 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FINDFIRST$q7PATHSTR4WORDm9SEARCHREC`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19940` | @FINDNEXT$qm9SEARCHREC | — | 26 | 13 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FINDNEXT$qm9SEARCHREC`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1995A` | sub_1995A | — | 35 | 21 | 2 | 1 |  | 待解讀 | — | — | — |
| `1997D` | @FEXPAND$q7PATHSTR | — | 197 | 104 | 1 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FEXPAND$q7PATHSTR`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19A42` | @FSplit$q7PathStrm6DirStrm7NameStrm6ExtStr | — | 82 | 37 | 2 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FSplit$q7PathStrm6DirStrm7NameStrm6ExtStr`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19A94` | sub_19A94 | — | 18 | 10 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>字串搬移輔助:長度夾到 bx 上限後 stosb 寫長度,再 rep movsb 複製內容並更新來源指標 | — |
| `19AB0` | @__CRTInit$qv | — | 48 | 24 | 1 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@__CRTInit$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19AE0` | unknown_libname_1 | — | 105 | 45 | 1 | 5 |  | 待解讀 | — | — | — |
| `19B49` | sub_19B49 | — | 78 | 31 | 2 | 2 |  | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>CRT 畫面模式設定(近呼叫 retn)：清 0040h:0087h bit 0；模式非 7 且 >=4 就強制成 3 後設模式；再 AX=1112h/BL=0 載入 8×8 字型、AX=1130h 取列數，列數為 2Ah(42)時才設回 0040h:0087h bit 0、設游標形狀(AX=0100h/CX=0600h)與 AH=12h/BL=20h | — |
| `19B97` | sub_19B97 | — | 72 | 30 | 2 | 2 |  | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>CRT 畫面參數偵測(近呼叫 retn)：INT 10h AH=0Fh 取模式，AX=1130h/BH=0/DL=0 取字型資訊(DL=列數−1)；DL=0 時改用 18h 且模式 ≤3 另設 byte_24E5Dh:=1；結果存 word_24E5Eh(模式與欄數)、word_24E68h(列數)、byte_24E5Ch:=1、word_24E62h:=0、word_24E64h:=dx；列數 > 18h 時 AH:=1 | — |
| `19BF5` | sub_19BF5 | — | 61 | 23 | 2 | 5 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>Ctrl-Break 處理(retf 2)：byte_24E6C 非 0 才動作 — 清旗標、用 INT 16h 把鍵盤緩衝區抽乾、印 ^C 與換行，然後 INT 23h | — |
| `19C32` | @WINDOW$q4BYTEt1t1t1 | — | 64 | 23 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@WINDOW$q4BYTEt1t1t1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19C72` | @CLRSCR$qv | — | 26 | 8 | 1 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@CLRSCR$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19C8C` | @CLREOL$qv | — | 20 | 7 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@CLREOL$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19CA0` | @INSLINE$qv | — | 37 | 15 | 0 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@INSLINE$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19CC5` | @GOTOXY$q4BYTEt1 | — | 44 | 15 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@GOTOXY$q4BYTEt1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19CF1` | @WHEREX$qv | — | 12 | 5 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@WHEREX$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19CFD` | @WHEREY$qv | — | 12 | 5 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@WHEREY$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D09` | @TEXTCOLOR$q4BYTE | — | 26 | 9 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@TEXTCOLOR$q4BYTE`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D23` | @TEXTBACKGROUND$q4BYTE | — | 24 | 8 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@TEXTBACKGROUND$q4BYTE`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D3B` | @LOWVIDEO$qv | — | 6 | 2 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@LOWVIDEO$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D41` | @HIGHVIDEO$qv | — | 6 | 2 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@HIGHVIDEO$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D47` | @NORMVIDEO$qv | — | 7 | 3 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@NORMVIDEO$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19D4E` | @DELAY$q4WORD | — | 32 | 13 | 3 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>@DELAY 實作:取毫秒參數後反覆讀 0000:0000 的 BIOS timer 位元組,每次變動遞減計數(輔助迴圈在 sub_19D6E) | — |
| `19D6E` | sub_19D6E | — | 8 | 4 | 3 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>@DELAY 的忙碌等待迴圈:比較 al 與 es:[di] 的 BIOS timer 位元組,相同就 loop | — |
| `19D76` | @SOUND$q4WORD | — | 45 | 20 | 0 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>@SOUND 實作:divisor := 1234DDh ÷ 頻率(頻率 <= 12h 直接返回);in 61h 開喇叭閘、out 43h 送 0B6h(PIT ch2 方波)、分兩次 out 42h 送 divisor 高低位元組 | — |
| `19DA3` | @NOSOUND$qv | — | 7 | 4 | 0 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>@NOSOUND 實作:in 61h 後 and 0FCh 再 out 61h,關閉 PC 喇叭的閘門與計時器輸出 | — |
| `19DAA` | @KEYPRESSED$qv | — | 18 | 8 | 3 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@KEYPRESSED$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19DBC` | @READKEY$qv | — | 34 | 14 | 3 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@READKEY$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19DDE` | @ASSIGNCRT$qm4TEXT | — | 43 | 13 | 1 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ASSIGNCRT$qm4TEXT`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `19F24` | sub_19F24 | — | 7 | 3 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `19F2B` | sub_19F2B | — | 89 | 43 | 8 | 5 |  | 待解讀 | — | — | — |
| `19F84` | sub_19F84 | — | 33 | 14 | 2 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>捲動一行:若 dh+1 超過 word_24E64 高位元組(視窗下界)則以 INT 10h AX=0601h 上捲一行,屬性取 byte_24E60、範圍取 word_24E62/word_24E64 | — |
| `19FA5` | sub_19FA5 | — | 7 | 3 | 5 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp sub_1A0B8`，控制權轉交後不返回；先設定 `mov ah, 3`、`xor bh, bh`（body 共 7 bytes，已逐條讀完） | — |
| `19FAC` | sub_19FAC | — | 7 | 3 | 4 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp sub_1A0B8`，控制權轉交後不返回；先設定 `mov ah, 2`、`xor bh, bh`（body 共 7 bytes，已逐條讀完） | — |
| `19FB3` | sub_19FB3 | — | 157 | 73 | 1 | 4 |  | 待解讀 | — | — | — |
| `1A050` | sub_1A050 | — | 104 | 56 | 1 | 1 |  | 待解讀 | — | — | — |
| `1A0B8` | sub_1A0B8 | — | 11 | 10 | 11 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>BIOS 視訊呼叫包裝:保存 si/di/es → int 10h → 還原。被呼叫 18 次,是 CRT 單元對 INT 10h 的統一入口 | — |
| `1A0D0` | sub_1A0D0 | — | 11 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call sub_1A611`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1A0E8` | @OVRINIT$q6String | — | 109 | 46 | 1 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@OVRINIT$q6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A155` | sub_1A155 | — | 11 | 7 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 11 bytes，已逐條讀完） | — |
| `1A160` | sub_1A160 | — | 58 | 31 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 58 bytes，已逐條讀完） | — |
| `1A19A` | sub_1A19A | — | 87 | 48 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 87 bytes，已逐條讀完） | — |
| `1A1F1` | sub_1A1F1 | — | 24 | 13 | 3 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>開檔:把 Pascal 字串複製成 ASCIIZ 到堆疊緩衝(lodsb 取長度後 rep movsb 再補 0),再以 INT 21h AX=3D00h 唯讀開啟 | — |
| `1A209` | @OVRSETBUF$q7LONGINT | — | 99 | 36 | 1 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@OVRSETBUF$q7LONGINT`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A26C` | @OVRGETBUF$qv | — | 20 | 8 | 1 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@OVRGETBUF$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A280` | @OVRCLEARBUF$qv | — | 47 | 19 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@OVRCLEARBUF$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A374` | sub_1A374 | — | 184 | 74 | 1 | 5 |  | 待解讀 | — | — | — |
| `1A42C` | sub_1A42C | — | 12 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>回傳 (es:[8] + 0Fh) >> 4,即把 es:[8] 的位元組數無條件進位換算成 paragraph 數 | — |
| `1A438` | sub_1A438 | — | 93 | 35 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 93 bytes，已逐條讀完） | — |
| `1A495` | sub_1A495 | — | 78 | 38 | 1 | 1 |  | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>搬移 overlay 區並改寫 stub(近呼叫 retn)：新段 := word_2096C、舊段 := es:10h；沿鏈把 [bp+4] 等於舊段的節點改成新段；用 std 反向搬 es:8 個 bytes；最後由 di=23h 起、每格前進 5、共 es:0Ch 次寫入新段。與 spec 757 的 1A4E3h 合起來確定 stub 陣列在 +20h、每格 5 bytes、格數在控制區塊 +0Ch；⚠ 兩支寫入位置在 +3 重疊，本輪不宣稱 stub 內部欄位切法 | — |
| `1A4E3` | sub_1A4E3 | — | 88 | 35 | 2 | 1 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>重建 VROOMM stub(近呼叫 retn)：從控制區塊 +20h 起、依 es:0Ch 的項數，每項寫 5 bytes = CD 3F(INT 3Fh) + 原本 es:[di+1] 的 2 bytes + 1 個 0。這是 stub 版面的產生端證據(既有結論是由 stub 內容反推) | — |
| `1A540` | @__SystemInit$qv | — | 157 | 72 | 1 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@__SystemInit$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A611` | sub_1A611 | — | 4 | 3 | 13 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1A61C`，控制權轉交後不返回；先設定 `pop cx`、`pop bx`（body 共 4 bytes，已逐條讀完） | — |
| `1A618` | @Halt$q4Word | — | 188 | 79 | 4 | 6 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Halt$q4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A6D4` | sub_1A6D4 | — | 14 | 7 | 2 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出 CS 內的 ASCIIZ 字串:逐字元讀 cs:[bx],非 0 就呼叫 sub_1A716 輸出並前進 | — |
| `1A6E2` | sub_1A6E2 | — | 12 | 5 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出三位十進位數:以除數 100 與 10 呼叫 sub_1A6EE 印百位與十位,再續印個位 | — |
| `1A6EE` | sub_1A6EE | — | 14 | 8 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出一位十進位數字:al := ax ÷ cl,加 30h 轉 ASCII 後呼叫 sub_1A716 輸出,並把餘數(ah)移回 al 供下一位使用 | — |
| `1A6FC` | sub_1A6FC | — | 7 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1A703` | sub_1A703 | — | 11 | 6 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and al, 0Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1A70E` | sub_1A70E | — | 8 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add al, 7`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1A716` | sub_1A716 | — | 7 | 4 | 4 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>字元輸出:mov dl, al 後以 INT 21h AH=06h(直接主控台輸出)送出。這就是十進位輸出與字串輸出共用的底層 routine | — |
| `1A747` | @IOResult$qv | — | 7 | 3 | 1 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@IOResult$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A74E` | @__IOCheck$qv | — | 14 | 5 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@__IOCheck$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A75C` | @__RangeCheck$q7Longintpa2$7Longint | — | 40 | 15 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@__RangeCheck$q7Longintpa2$7Longint`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A784` | @__StackCheck$q4Word | — | 25 | 10 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@__StackCheck$q4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A79D` | @$basg$qm3Anyt14Word | — | 24 | 9 | 6 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$basg$qm3Anyt14Word`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1A7B5` | @Sqr$q7Longint | — | 4 | 2 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Sqr$q7Longint`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A7B9` | @$brmul$q7Longintt1 | — | 27 | 16 | 7 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brmul$q7Longintt1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1A7D4` | @$brdiv$q7Longintt1 | — | 110 | 55 | 2 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brdiv$q7Longintt1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1A842` | @$brrsh$q7Longint7Integer | — | 12 | 6 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brrsh$q7Longint7Integer`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1A84E` | @$brlsh$q7Longint7Integer | — | 12 | 6 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brlsh$q7Longint7Integer`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1A85A` | @Abs$q7Longint | — | 15 | 7 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Abs$q7Longint`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A869` | @GetMem$qm7Pointer4Word | — | 59 | 24 | 3 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@GetMem$qm7Pointer4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A8A4` | @FreeMem$qm7Pointer4Word | — | 32 | 13 | 4 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FreeMem$qm7Pointer4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A8C4` | @Mark$qm7Pointer | — | 22 | 7 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Mark$qm7Pointer`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A8DA` | @Release$qm7Pointer | — | 27 | 9 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Release$qm7Pointer`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A8F5` | @MemAvail$qv | — | 68 | 26 | 1 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@MemAvail$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A939` | @MaxAvail$qv | — | 77 | 31 | 0 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@MaxAvail$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1A986` | sub_1A986 | — | 156 | 62 | 1 | 4 |  | 待解讀 | — | — | — |
| `1AA22` | sub_1AA22 | — | 179 | 66 | 1 | 4 |  | 待解讀 | — | — | — |
| `1AAD5` | sub_1AAD5 | — | 36 | 15 | 1 | 1 |  | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>向下配置 8 bytes 並檢查下界(近呼叫，失敗時 CF=1)：di := word[2097Eh]−8，為 0 則失敗；si := (di shr 4) + word[20980h]，si <= word[2097Ch] 則失敗；成功才把 di 寫回。減法未處理借位 | — |
| `1AAF9` | sub_1AAF9 | — | 21 | 9 | 2 | 0 |  | 已解讀 | exact | docs/spec/575-random-core-and-pc98-vram.md<br>以 dword_2097E 為來源連續四次 movsw 複製 8 bytes,之後 di 回退 8 並更新來源指標 | — |
| `1AB0E` | sub_1AB0E | — | 73 | 27 | 3 | 1 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>指標正規化與下界夾制(近呼叫 retn)：dword_2097Eh 減 word_20982h，低 4 bits 留 SI、其餘右移 4 加到段值，與 word_2097Ch:word_2097Ah 這個下界比，低於就夾到下界；offset 為 0 有段值 +1000h 的特別路徑。與 spec 753 的 1AAD5h 同族 | — |
| `1AB57` | sub_1AB57 | — | 14 | 7 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>把 AX 拆成兩個 nibble:DX := AX >> 4、AX := AX and 0Fh | — |
| `1AB65` | sub_1AB65 | — | 15 | 7 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>nibble 合併:rol(DX,4) 後把高 12 位併進 AX、DX 留低 4 位 | — |
| `1AB74` | @$basg$qm6Stringt1 | — | 26 | 12 | 15 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$basg$qm6Stringt1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AB8E` | @$basg$qm6Stringt14Byte | — | 36 | 16 | 26 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$basg$qm6Stringt14Byte`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1ABB2` | @Length$qm6String | — | 14 | 5 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Length$qm6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1ABC0` | @Copy$qm6Stringt17Integert3 | — | 65 | 29 | 5 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Copy$qm6Stringt17Integert3`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AC01` | @Concat$qm6Stringt1 | — | 44 | 19 | 8 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Concat$qm6Stringt1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AC2D` | @Pos$qm6Stringt1 | — | 55 | 31 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Pos$qm6Stringt1`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AC64` | @$bsub$qm6Stringt1 | — | 43 | 20 | 6 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$bsub$qm6Stringt1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AC8F` | @$basg$qm6String4Char | — | 18 | 8 | 3 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$basg$qm6String4Char`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1ACA1` | @$basg$qm6Stringn4Char4Byte | — | 27 | 11 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$basg$qm6Stringn4Char4Byte`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1ACBC` | @Insert$qm6Stringt14Word7Integer | — | 84 | 43 | 0 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Insert$qm6Stringt14Word7Integer`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AD10` | @Delete$qm6String7Integert2 | — | 86 | 43 | 1 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Delete$qm6String7Integert2`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AD66` | @Set@$bctr$qn4Byte4Word | — | 42 | 18 | 3 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@$bctr$qn4Byte4Word`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AD90` | @Set@Clear$qv | — | 15 | 7 | 1 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@Clear$qv`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AD9F` | @Set@$brplu$q4Byte | — | 33 | 13 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@$brplu$q4Byte`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1ADC0` | @Set@$brplu$q4Bytet1 | — | 52 | 21 | 1 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@$brplu$q4Bytet1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1ADF4` | @Set@$bctr$qm3Set4Word | — | 32 | 13 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@$bctr$qm3Set4Word`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AE14` | @Set@MemberOf$q4Byte | — | 33 | 13 | 7 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@Set@MemberOf$q4Byte`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AE35` | @$brplu$qm3Sett1 | — | 28 | 12 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brplu$qm3Sett1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AE51` | @$brmin$qm3Sett1 | — | 30 | 13 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brmin$qm3Sett1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AE6F` | @$brmul$qm3Sett1 | — | 28 | 12 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brmul$qm3Sett1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AE8B` | @$beql$qm3Sett1 | — | 23 | 9 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$beql$qm3Sett1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AEA2` | @$bgeq$qm3Sett1 | — | 30 | 13 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$bgeq$qm3Sett1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1AEC0` | __RealSub | — | 4 | 1 | 6 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealSub`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AEC4` | __RealAdd | — | 214 | 105 | 9 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealAdd`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1AF9A` | __RealMul | — | 125 | 63 | 8 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealMul`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B017` | __RealDiv | — | 137 | 74 | 5 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealDiv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B0A0` | sub_1B0A0 | — | 35 | 18 | 2 | 1 |  | 待解讀 | — | — | — |
| `1B0C3` | __RealCmp | — | 23 | 14 | 3 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealCmp`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B0DA` | sub_1B0DA | — | 19 | 10 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>多欄位相等比較:依序比較 al/cl、dx/di、bx/si、ah/ch,任一不等就提前返回 | — |
| `1B0ED` | __RealFloat | — | 63 | 30 | 3 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealFloat`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B12C` | __RealTrunc | — | 92 | 44 | 2 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__RealTrunc`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B188` | @$brplu$q4Realt1 | — | 76 | 35 | 0 | 9 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `@$brplu$q4Realt1`，`$b` 前綴是 Borland 的運算子編碼 | — |
| `1B1D4` | sub_1B1D4 | — | 10 | 8 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call __RealAdd`（body 共 10 bytes，已逐條讀完） | — |
| `1B1DE` | sub_1B1DE | — | 10 | 8 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call __RealSub`（body 共 10 bytes，已逐條讀完） | — |
| `1B1E8` | sub_1B1E8 | — | 10 | 8 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call __RealMul`（body 共 10 bytes，已逐條讀完） | — |
| `1B1F2` | sub_1B1F2 | — | 10 | 8 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call __RealDiv`（body 共 10 bytes，已逐條讀完） | — |
| `1B1FC` | @Int$q4Real | — | 81 | 39 | 1 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Int$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B24D` | @Frac$q4Real | — | 20 | 13 | 1 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Frac$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B261` | @Sqrt$q4Real | — | 93 | 43 | 0 | 5 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Sqrt$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B2BE` | @Sin$q4Real | — | 116 | 52 | 0 | 9 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Sin$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B405` | @Exp$q4Real | — | 115 | 56 | 0 | 8 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Exp$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B4A8` | @ArcTan$q4Real | — | 225 | 98 | 0 | 8 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ArcTan$q4Real`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B5D7` | sub_1B5D7 | — | 31 | 18 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp __RealMul`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1B5F6` | sub_1B5F6 | — | 79 | 33 | 2 | 3 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>6-byte real 的 Horner 多項式求值(近呼叫 retn)：入口 AX:BX:DX 是 x、CS:DI 指向係數表、CX 是項數；迴圈用 __RealAdd/__RealMul，最後加上 CX=81h,SI=0,DI=0(Turbo Pascal real 的 1.0)。屬 RTL 數學核心 | — |
| `1B645` | @Random$q4Word | — | 22 | 10 | 5 | 2 |  | 已解讀 | exact | docs/spec/575-random-core-and-pc98-vram.md<br>@Random(n) 本體:呼叫 LCG 後 xor ax,ax 丟棄低位字,n 為 0 回 0,否則回傳 (RandSeed shr 16) mod n | — |
| `1B65B` | @Random__Real$qv | — | 29 | 13 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Random__Real$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B678` | @Random__Extended$qv | — | 26 | 8 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Random__Extended$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B694` | sub_1B694 | — | 54 | 22 | 3 | 0 |  | 已解讀 | exact | docs/spec/575-random-core-and-pc98-vram.md<br>Turbo Pascal RTL 的 LCG:RandSeed := RandSeed*134775813 + 1 (mod 2^32);高位乘數 0808h 由 shl/add ch,cl/add dh,bl 湊出,只有低位 8405h 是常數(cs:word_1B6CA);已用 20 萬組種子數值驗證等價 | — |
| `1B6CC` | @Randomize$qv | — | 13 | 5 | 0 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>@Randomize 實作:INT 21h AH=2Ch 取系統時間,把 CX:DX 存進 dword_20998 ⇒ dword_20998 就是 Turbo Pascal 的 RandSeed(原版隨機序列的種子來源) | — |
| `1B6D9` | __Int2Str | — | 82 | 37 | 2 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__Int2Str`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B753` | __Str2Int | — | 152 | 74 | 2 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__Str2Int`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B7EB` | @Str$q7Longint4Wordm6String4Byte | — | 75 | 39 | 1 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Str$q7Longint4Wordm6String4Byte`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B836` | @Val__Longint$qm6Stringm7Integer | — | 49 | 22 | 1 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Val__Longint$qm6Stringm7Integer`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B867` | @Assign$qm4Textm6String | — | 69 | 37 | 3 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Assign$qm4Textm6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B8AC` | @SetTextBuf$qm4Textm3Any4Word | — | 43 | 12 | 0 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@SetTextBuf$qm4Textm3Any4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B8D7` | @Reset$qm4Text | — | 5 | 2 | 3 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Reset$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B8DC` | @Rewrite$qm4Text | — | 80 | 29 | 4 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Rewrite$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B92C` | @Flush$qm4Text | — | 4 | 2 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Flush$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B930` | @Close$qm4Text | — | 59 | 19 | 5 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Close$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1B96B` | sub_1B96B | — | 17 | 11 | 2 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 es:[bx+di] 的 far pointer 呼叫例外/結束處理常式(自身的 es:di 傳兩次當參數);回傳非 0 就存進 word_20996 | — |
| `1B9FE` | sub_1B9FE | — | 90 | 40 | 1 | 1 |  | 已解讀 | exact | docs/spec/759-combatant-xy-lookup-and-bank-select-bit.md<br>依 ^Z 截斷文字檔(近呼叫 retn，DI 指向檔案記錄)：seek 檔尾取大小，回退 80h(不足則 0)讀 80h bytes，在其中找第一個 1Ah；找到就以 AX=4202h 相對檔尾往回 seek 到該處，再用 AH=40h 配 CX=0 截斷。找不到就直接返回 | — |
| `1BAE6` | __GetEntry | — | 36 | 10 | 4 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__GetEntry`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BB0A` | __GetChar | — | 28 | 14 | 4 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__GetChar`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BB26` | __PutEntry | — | 36 | 10 | 4 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__PutEntry`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BB4A` | __PutChar | — | 15 | 7 | 4 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `__PutChar`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BB59` | sub_1BB59 | — | 49 | 24 | 2 | 1 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>Text 檔案的 I/O 派工(近呼叫 retn)：以 Turbo Pascal TextRec 配置對上 — +08h BufPos 先由 BX 寫入，call dword ptr es:[di+14h](InOutFunc)，結果非 0 存進 word_20996(IOResult)，再把 +0Ah BufEnd→AX、+08h BufPos→BX、+04h BufSize→DX、+0Ch BufPtr→ES:DI | — |
| `1BB8A` | @ReadLn$qm4Text | — | 41 | 20 | 1 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ReadLn$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BBB3` | @WriteLn$qm4Text | — | 31 | 14 | 3 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@WriteLn$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BBD2` | @Write$qm4Text | — | 38 | 14 | 2 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Write$qm4Text`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BBF8` | @Read$qm4Text4Char | — | 30 | 15 | 1 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Read$qm4Text4Char`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BC16` | @Write$qm4Text4Char4Word | — | 45 | 20 | 2 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Write$qm4Text4Char4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BC43` | @Read$qm4Textm6String4Word | — | 56 | 29 | 1 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Read$qm4Textm6String4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BC7B` | @Write$qm4Textm6String4Word | — | 62 | 32 | 3 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Write$qm4Textm6String4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BCB9` | @Read$qm4Text7Longint | — | 88 | 42 | 0 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Read$qm4Text7Longint`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BD11` | @Write$qm4Text7Longint4Word | — | 73 | 34 | 2 | 4 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Write$qm4Text7Longint4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BD5A` | @Assign$qm4Filem6String | — | 46 | 23 | 2 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Assign$qm4Filem6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BD88` | @Reset$qm4File4Word | — | 101 | 42 | 1 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Reset$qm4File4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BDED` | @Truncate$qm4File | — | 28 | 11 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Truncate$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BE09` | @Close$qm4File | — | 37 | 13 | 3 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Close$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BE2E` | sub_1BE2E | — | 15 | 4 | 5 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>檢查 es:[di+2] 是否為 0D7B3h 簽章;不符時把 word_20996 設為 67h | — |
| `1BE3D` | @Read$qm4Filem3Any | — | 54 | 25 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Read$qm4Filem3Any`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BE73` | @BlockRead$qm4Filem3Any4Wordm4Word | — | 104 | 45 | 2 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@BlockRead$qm4Filem3Any4Wordm4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BEDB` | @Seek$qm4File7Longint | — | 48 | 19 | 2 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Seek$qm4File7Longint`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BF0B` | @FilePos$qm4File | — | 23 | 9 | 1 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FilePos$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BF22` | @FileSize$qm4File | — | 27 | 11 | 0 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FileSize$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BF3D` | @Eof$qm4File | — | 21 | 9 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Eof$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BF52` | sub_1BF52 | — | 69 | 32 | 3 | 1 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>取檔案大小(近呼叫 retn)：檔案記錄 +02h 不等於 0D7B3h(fmInOut)就把 word_20996(IOResult) 設 67h(103 檔案未開啟)並回 CF；否則用 INT 21h AH=42h 做 AL=01/02/00 三次 LSEEK(取現位置、取檔尾、移回)，把檔案大小放進 BX:CX | — |
| `1BF97` | @Erase$qm4File | — | 23 | 10 | 1 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Erase$qm4File`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BFAE` | @Rename$qm4Filem6String | — | 79 | 42 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Rename$qm4Filem6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1BFFD` | @Move$qm3Anyt14Word | — | 35 | 16 | 26 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@Move$qm3Anyt14Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C020` | @FillChar$qm3Any4Word4Byte | — | 20 | 7 | 19 | 0 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@FillChar$qm3Any4Word4Byte`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C034` | @ParamStr$qm6String4Word | — | 79 | 38 | 4 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ParamStr$qm6String4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C083` | @ParamCount$qv | — | 7 | 4 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ParamCount$qv`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C08A` | sub_1C08A | — | 50 | 24 | 2 | 1 |  | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>掃 PSP 命令列(近呼叫 retn)：從 ES:0080h 起跳過空白(<= 20h)、量一段非空白長度，DX 每數完一段減一，BX 累計段數，SI/DI 留下該段起訖。ParamCount/ParamStr 的掃描核心 | — |
| `1C0BC` | @GetDir$q4Bytem6String4Word | — | 81 | 40 | 0 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@GetDir$q4Bytem6String4Word`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C10D` | @ChDir$qm6String | — | 65 | 27 | 0 | 3 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@ChDir$qm6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C14E` | @MkDir$qm6String | — | 21 | 9 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@MkDir$qm6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C163` | @RmDir$qm6String | — | 21 | 9 | 0 | 2 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@RmDir$qm6String`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
| `1C178` | sub_1C178 | — | 27 | 17 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 27 bytes，已逐條讀完） | — |
| `1C193` | sub_1C193 | — | 15 | 9 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 15 bytes，已逐條讀完） | — |
| `1C1A2` | @UpCase$q4Char | — | 19 | 8 | 4 | 1 |  | 不阻塞 | — | docs/spec/566-turbo-pascal-rtl-not-blocking.md<br>Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `@UpCase$q4Char`；屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則 | — |
