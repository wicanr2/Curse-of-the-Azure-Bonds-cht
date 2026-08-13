# pc98-PC98-GAME.EXE 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `10000` | start | — | 756 | 219 | 0 | 69 |  | 待解讀 | — | — | — |
| `10320` | sub_10320 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10350` | sub_10350 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10355` | sub_10355 | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10390` | sub_10390 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10395` | sub_10395 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1039A` | sub_1039A | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `104E0` | sub_104E0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10520` | sub_10520 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10590` | sub_10590 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10600` | sub_10600 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10640` | sub_10640 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10710` | sub_10710 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10780` | sub_10780 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `107F0` | sub_107F0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10890` | sub_10890 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10895` | sub_10895 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1089A` | sub_1089A | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1089F` | sub_1089F | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `108D0` | sub_108D0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10BA0` | sub_10BA0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10CB0` | sub_10CB0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10D20` | sub_10D20 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10DE0` | sub_10DE0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10E2B` | sub_10E2B | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10EB0` | sub_10EB0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10EB5` | sub_10EB5 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10EC9` | sub_10EC9 | — | 2 | 1 | 2 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10F50` | sub_10F50 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `10FC0` | sub_10FC0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11070` | sub_11070 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `110F0` | sub_110F0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11190` | sub_11190 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11217` | sub_11217 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11400` | sub_11400 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `114C0` | sub_114C0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `115F0` | sub_115F0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11660` | sub_11660 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1167E` | sub_1167E | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `116F0` | sub_116F0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11740` | sub_11740 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11780` | sub_11780 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `117E0` | sub_117E0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `117EA` | sub_117EA | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11817` | sub_11817 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11830` | sub_11830 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `11860` | sub_11860 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `118B0` | sub_118B0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11960` | sub_11960 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `11990` | sub_11990 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `119C0` | sub_119C0 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `119E0` | sub_119E0 | — | 66 | 32 | 3 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1234F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 66 bytes，已逐條讀完） | — |
| `12095` | sub_12095 | — | 373 | 179 | 5 | 2 |  | 待解讀 | — | — | — |
| `1224A` | sub_1224A | — | 66 | 32 | 1 | 3 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1234F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 66 bytes，已逐條讀完） | — |
| `1230E` | sub_1230E | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `12340` | sub_12340 | — | 2 | 1 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1234F` | sub_1234F | — | 2 | 1 | 40 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>overlay stub：Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），由 overlay manager 轉派，不含遊戲邏輯（body 共 2 bytes，已逐條讀完） | — |
| `1259F` | sub_1259F | — | 1244 | 502 | 1 | 27 |  | 待解讀 | — | — | — |
| `12A80` | sub_12A80 | — | 24 | 10 | 1 | 0 |  | 待解讀 | — | — | — |
| `12E3D` | sub_12E3D | — | 580 | 211 | 2 | 2 |  | 待解讀 | — | — | — |
| `13231` | sub_13231 | — | 133 | 54 | 2 | 0 |  | 待解讀 | — | — | — |
| `139DD` | sub_139DD | — | 376 | 145 | 2 | 4 |  | 待解讀 | — | — | — |
| `13B55` | sub_13B55 | — | 159 | 56 | 1 | 2 |  | 待解讀 | — | — | — |
| `13BF4` | sub_13BF4 | — | 542 | 224 | 1 | 3 |  | 待解讀 | — | — | — |
| `13E1B` | sub_13E1B | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `13E22` | sub_13E22 | — | 138 | 60 | 1 | 1 |  | 待解讀 | — | — | — |
| `13EAC` | sub_13EAC | — | 197 | 85 | 1 | 1 |  | 待解讀 | — | — | — |
| `13F71` | sub_13F71 | — | 282 | 108 | 1 | 4 |  | 待解讀 | — | — | — |
| `1408B` | sub_1408B | — | 60 | 31 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 60 bytes，已逐條讀完） | — |
| `140C7` | sub_140C7 | — | 106 | 48 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 106 bytes，已逐條讀完） | — |
| `14131` | sub_14131 | — | 77 | 36 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 77 bytes，已逐條讀完） | — |
| `14180` | sub_14180 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `14187` | sub_14187 | — | 25 | 11 | 1 | 0 |  | 待解讀 | — | — | — |
| `141A0` | sub_141A0 | — | 164 | 77 | 4 | 4 |  | 待解讀 | — | — | — |
| `142B9` | sub_142B9 | — | 137 | 62 | 3 | 2 |  | 待解讀 | — | — | — |
| `14342` | sub_14342 | — | 48 | 20 | 1 | 1 |  | 待解讀 | — | — | — |
| `14372` | sub_14372 | — | 600 | 238 | 1 | 1 |  | 待解讀 | — | — | — |
| `145CA` | sub_145CA | — | 2253 | 814 | 2 | 4 |  | 待解讀 | — | — | — |
| `14E97` | sub_14E97 | — | 142 | 61 | 4 | 4 |  | 待解讀 | — | — | — |
| `152D0` | sub_152D0 | — | 265 | 123 | 1 | 5 |  | 待解讀 | — | — | — |
| `1546F` | sub_1546F | — | 78 | 40 | 1 | 4 |  | 待解讀 | — | — | — |
| `1562A` | sub_1562A | — | 29 | 13 | 1 | 1 |  | 待解讀 | — | — | — |
| `15647` | sub_15647 | — | 257 | 132 | 2 | 3 |  | 待解讀 | — | — | — |
| `1691B` | sub_1691B | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `1692D` | sub_1692D | — | 57 | 28 | 3 | 2 |  | 待解讀 | — | — | — |
| `16966` | sub_16966 | — | 80 | 42 | 1 | 3 |  | 待解讀 | — | — | — |
| `169B6` | sub_169B6 | — | 35 | 21 | 1 | 1 |  | 待解讀 | — | — | — |
| `169D9` | sub_169D9 | — | 20 | 12 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導位元組判定(第二族):81h..9Fh 或 E0h..F7h 時 stc(進位置起)表示前導,否則 clc。⚠ 上界是 F7h,與 18D5Dh 那族的 FCh 不同 | — |
| `169ED` | sub_169ED | — | 33 | 16 | 1 | 1 |  | 待解讀 | — | — | — |
| `16E70` | sub_16E70 | — | 94 | 48 | 2 | 7 |  | 待解讀 | — | — | — |
| `16ECE` | sub_16ECE | — | 39 | 22 | 1 | 1 |  | 待解讀 | — | — | — |
| `16F2B` | sub_16F2B | — | 86 | 33 | 1 | 3 |  | 待解讀 | — | — | — |
| `16F81` | sub_16F81 | — | 78 | 45 | 1 | 4 |  | 待解讀 | — | — | — |
| `16FCF` | sub_16FCF | — | 29 | 17 | 1 | 0 |  | 待解讀 | — | — | — |
| `16FEC` | sub_16FEC | — | 110 | 63 | 1 | 3 |  | 待解讀 | — | — | — |
| `1705A` | sub_1705A | — | 107 | 61 | 1 | 3 |  | 待解讀 | — | — | — |
| `170C5` | sub_170C5 | — | 90 | 57 | 1 | 2 |  | 待解讀 | — | — | — |
| `1711F` | sub_1711F | — | 41 | 24 | 3 | 1 |  | 待解讀 | — | — | — |
| `17148` | sub_17148 | — | 42 | 25 | 4 | 1 |  | 待解讀 | — | — | — |
| `17172` | sub_17172 | — | 20 | 12 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導判定(第二族),與 169D9h 逐指令相同 | — |
| `17186` | sub_17186 | — | 33 | 16 | 1 | 1 |  | 待解讀 | — | — | — |
| `171DF` | sub_171DF | — | 63 | 27 | 1 | 1 |  | 待解讀 | — | — | — |
| `17220` | sub_17220 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `17230` | sub_17230 | — | 9 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 9 bytes，已逐條讀完） | — |
| `17239` | sub_17239 | — | 23 | 8 | 1 | 0 |  | 待解讀 | — | — | — |
| `17250` | sub_17250 | — | 80 | 36 | 1 | 2 |  | 待解讀 | — | — | — |
| `172A0` | sub_172A0 | — | 53 | 26 | 1 | 1 |  | 待解讀 | — | — | — |
| `172D5` | sub_172D5 | — | 90 | 48 | 1 | 4 |  | 待解讀 | — | — | — |
| `173A1` | sub_173A1 | — | 75 | 32 | 1 | 6 |  | 待解讀 | — | — | — |
| `1741D` | sub_1741D | — | 569 | 231 | 1 | 10 |  | 待解讀 | — | — | — |
| `17656` | sub_17656 | — | 104 | 42 | 3 | 5 |  | 待解讀 | — | — | — |
| `177BD` | sub_177BD | — | 48 | 20 | 1 | 1 |  | 待解讀 | — | — | — |
| `1790D` | sub_1790D | — | 83 | 31 | 1 | 3 |  | 待解讀 | — | — | — |
| `17960` | sub_17960 | — | 215 | 99 | 1 | 7 |  | 待解讀 | — | — | — |
| `17A37` | sub_17A37 | — | 29 | 13 | 1 | 1 |  | 待解讀 | — | — | — |
| `17A54` | sub_17A54 | — | 609 | 227 | 2 | 12 |  | 待解讀 | — | — | — |
| `17CD7` | sub_17CD7 | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `17CE9` | sub_17CE9 | — | 57 | 28 | 2 | 2 |  | 待解讀 | — | — | — |
| `17D22` | sub_17D22 | — | 80 | 42 | 1 | 3 |  | 待解讀 | — | — | — |
| `17D72` | sub_17D72 | — | 35 | 21 | 1 | 1 |  | 待解讀 | — | — | — |
| `17D95` | sub_17D95 | — | 20 | 12 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導判定(第二族),與 169D9h 逐指令相同 | — |
| `17DA9` | sub_17DA9 | — | 33 | 16 | 1 | 1 |  | 待解讀 | — | — | — |
| `17DD5` | sub_17DD5 | — | 86 | 44 | 1 | 1 |  | 待解讀 | — | — | — |
| `17EA7` | sub_17EA7 | — | 363 | 152 | 1 | 11 |  | 待解讀 | — | — | — |
| `18036` | sub_18036 | — | 193 | 66 | 6 | 7 |  | 待解讀 | — | — | — |
| `180F7` | sub_180F7 | — | 28 | 13 | 3 | 3 |  | 待解讀 | — | — | — |
| `1812D` | sub_1812D | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `18194` | sub_18194 | — | 51 | 27 | 1 | 2 |  | 待解讀 | — | — | — |
| `181C7` | sub_181C7 | — | 40 | 26 | 1 | 1 |  | 待解讀 | — | — | — |
| `18229` | sub_18229 | — | 10 | 6 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call sub_18878`（body 共 10 bytes，已逐條讀完） | — |
| `18233` | sub_18233 | — | 351 | 124 | 1 | 4 |  | 待解讀 | — | — | — |
| `18392` | sub_18392 | — | 77 | 32 | 1 | 2 |  | 待解讀 | — | — | — |
| `18407` | sub_18407 | — | 311 | 129 | 2 | 12 |  | 待解讀 | — | — | — |
| `1856A` | sub_1856A | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `18571` | sub_18571 | — | 121 | 48 | 2 | 1 |  | 待解讀 | — | — | — |
| `185EA` | sub_185EA | — | 40 | 18 | 2 | 1 |  | 待解讀 | — | — | — |
| `18612` | sub_18612 | — | 113 | 52 | 2 | 5 |  | 待解讀 | — | — | — |
| `18683` | sub_18683 | — | 124 | 56 | 2 | 4 |  | 待解讀 | — | — | — |
| `186FF` | sub_186FF | — | 58 | 30 | 2 | 1 |  | 待解讀 | — | — | — |
| `18767` | sub_18767 | — | 67 | 37 | 1 | 1 |  | 待解讀 | — | — | — |
| `187AA` | sub_187AA | — | 66 | 33 | 1 | 1 |  | 待解讀 | — | — | — |
| `187EC` | sub_187EC | — | 36 | 21 | 1 | 1 |  | 待解讀 | — | — | — |
| `18810` | sub_18810 | — | 46 | 25 | 2 | 1 |  | 待解讀 | — | — | — |
| `1883E` | sub_1883E | — | 36 | 20 | 2 | 0 |  | 待解讀 | — | — | — |
| `18862` | sub_18862 | — | 11 | 8 | 4 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出 BEL(07h):INT 21h AH=02h(DOS display output)、DL=07h,前後保存 AX/DX | — |
| `1886D` | sub_1886D | — | 11 | 5 | 1 | 3 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>sub_18571 回傳 0 時才呼叫 18862h 發出 BEL | — |
| `18878` | sub_18878 | — | 11 | 5 | 1 | 3 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>sub_185EA 回傳非 0 時才呼叫 18862h 發出 BEL | — |
| `18883` | sub_18883 | — | 46 | 19 | 1 | 3 |  | 待解讀 | — | — | — |
| `188B1` | sub_188B1 | — | 23 | 10 | 1 | 3 |  | 待解讀 | — | — | — |
| `188C8` | sub_188C8 | — | 27 | 10 | 1 | 1 |  | 待解讀 | — | — | — |
| `188E3` | sub_188E3 | — | 35 | 14 | 1 | 1 |  | 待解讀 | — | — | — |
| `18930` | sub_18930 | — | 269 | 100 | 3 | 4 |  | 待解讀 | — | — | — |
| `18A3D` | sub_18A3D | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `18A44` | sub_18A44 | — | 74 | 29 | 1 | 4 |  | 待解讀 | — | — | — |
| `18A8E` | sub_18A8E | — | 25 | 12 | 3 | 1 |  | 待解讀 | — | — | — |
| `18AA7` | sub_18AA7 | — | 232 | 95 | 1 | 3 |  | 待解讀 | — | — | — |
| `18B8F` | sub_18B8F | — | 7 | 5 | 1 | 0 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `18BB1` | sub_18BB1 | — | 26 | 14 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1AC99`（body 共 26 bytes，已逐條讀完） | — |
| `18BDB` | sub_18BDB | — | 58 | 40 | 3 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@INTR$q4BYTEm9REGISTERS` 逐位元組相同（58 bytes） | — |
| `18C72` | sub_18C72 | — | 37 | 19 | 2 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@GetTime$qm4Wordt1t1t1` 逐位元組相同（37 bytes） | — |
| `18D3A` | sub_18D3A | — | 35 | 21 | 2 | 1 |  | 待解讀 | — | — | — |
| `18D5D` | sub_18D5D | — | 19 | 11 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導位元組判定:al 無 bit7 → 非雙位元組;81h..9Fh 與 E0h..FCh → 前導位元組;A0h..DFh(半形片假名)與 FDh 以上 → 非前導 | — |
| `18D70` | sub_18D70 | — | 212 | 113 | 1 | 2 |  | 待解讀 | — | — | — |
| `18E44` | sub_18E44 | — | 19 | 11 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導位元組判定,與 18D5Dh 逐指令相同 | — |
| `18E59` | sub_18E59 | — | 113 | 50 | 2 | 3 |  | 待解讀 | — | — | — |
| `18ECA` | sub_18ECA | — | 18 | 10 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>字串搬移輔助,與 DOS START.EXE:19A94h 逐指令相同 | — |
| `18EE0` | sub_18EE0 | — | 48 | 24 | 1 | 4 |  | 待解讀 | — | — | — |
| `18F10` | sub_18F10 | — | 146 | 53 | 1 | 5 |  | 待解讀 | — | — | — |
| `18FA3` | sub_18FA3 | — | 34 | 17 | 4 | 1 |  | 待解讀 | — | — | — |
| `18FC5` | sub_18FC5 | — | 31 | 16 | 1 | 1 |  | 待解讀 | — | — | — |
| `18FE4` | sub_18FE4 | — | 17 | 7 | 2 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AH=0Ah 呼叫 1977Bh(INT 18h)、輸出 byte_280FF、再以 AH=0Ch 呼叫 1977Bh ⇒ PC-98 螢幕 BIOS 的成對開關包夾一次輸出 | — |
| `18FF5` | sub_18FF5 | — | 114 | 41 | 2 | 2 |  | 待解讀 | — | — | — |
| `19085` | sub_19085 | — | 141 | 54 | 2 | 10 |  | 待解讀 | — | — | — |
| `19112` | sub_19112 | — | 21 | 13 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>從 CS:247h 起以 INT 29h(fast putchar)輸出 5 個字元 | — |
| `1917C` | sub_1917C | — | 31 | 9 | 2 | 2 |  | 待解讀 | — | — | — |
| `191C2` | sub_191C2 | — | 26 | 10 | 2 | 2 |  | 待解讀 | — | — | — |
| `19259` | sub_19259 | — | 24 | 9 | 3 | 2 |  | 待解讀 | — | — | — |
| `19271` | sub_19271 | — | 21 | 11 | 1 | 1 |  | 待解讀 | — | — | — |
| `19286` | sub_19286 | — | 12 | 4 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>把堆疊上的 word 參數存進 word_280E6(以 ss:[bx+4] 取參數,retf 2) | — |
| `19293` | sub_19293 | — | 13 | 7 | 3 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>PC-98 鍵盤 BIOS:INT 18h AH=01h(鍵盤狀態偵測),BH 非零時回傳 1 否則 0 | — |
| `192A0` | sub_192A0 | — | 50 | 22 | 4 | 2 |  | 待解讀 | — | — | — |
| `192D2` | sub_192D2 | — | 87 | 27 | 1 | 2 |  | 待解讀 | — | — | — |
| `19386` | sub_19386 | — | 52 | 25 | 1 | 1 |  | 待解讀 | — | — | — |
| `193BA` | sub_193BA | — | 14 | 8 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 CS:word_193CD 的 handle 呼叫 INT 21h AH=3Eh 關閉檔案,前後保存 AX/BX | — |
| `1948C` | sub_1948C | — | 7 | 3 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `19493` | sub_19493 | — | 291 | 115 | 5 | 12 |  | 待解讀 | — | — | — |
| `195B6` | sub_195B6 | — | 70 | 29 | 2 | 4 |  | 待解讀 | — | — | — |
| `195FC` | sub_195FC | — | 8 | 3 | 6 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>呼叫 sub_19A8B 後把 word_280F9 載入 DX 回傳 | — |
| `19604` | sub_19604 | — | 18 | 7 | 7 | 3 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>以 dx 呼叫 sub_19B00 與 sub_19B4F,結果存進 word_280F9 後呼叫 sub_19A29 | — |
| `19616` | sub_19616 | — | 179 | 69 | 1 | 8 |  | 待解讀 | — | — | — |
| `196C9` | sub_196C9 | — | 178 | 83 | 1 | 6 |  | 待解讀 | — | — | — |
| `1977B` | sub_1977B | — | 3 | 2 | 5 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>int 18h 後返回。⚠ IDA 的註解寫 TRANSFER TO ROM BASIC 是 IBM PC 的語意;PC-98 的 INT 18h 是螢幕／鍵盤 BIOS,不可沿用該註解 | — |
| `1977E` | sub_1977E | — | 19 | 11 | 6 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>Shift-JIS 前導位元組判定,與 18D5Dh 逐指令相同 | — |
| `19791` | sub_19791 | — | 44 | 20 | 2 | 1 |  | 待解讀 | — | — | — |
| `197BD` | sub_197BD | — | 50 | 29 | 2 | 4 |  | 待解讀 | — | — | — |
| `197EF` | sub_197EF | — | 54 | 31 | 1 | 4 |  | 待解讀 | — | — | — |
| `19825` | sub_19825 | — | 27 | 13 | 2 | 0 |  | 待解讀 | — | — | — |
| `19840` | sub_19840 | — | 110 | 56 | 4 | 4 |  | 待解讀 | — | — | — |
| `198AE` | sub_198AE | — | 153 | 83 | 1 | 4 |  | 待解讀 | — | — | — |
| `19947` | sub_19947 | — | 174 | 93 | 1 | 4 |  | 待解讀 | — | — | — |
| `199F5` | sub_199F5 | — | 18 | 12 | 3 | 0 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>文字 VRAM 位址計算:di := ((di × 50h) + dx) × 2 ⇒ 每列 80 欄、每格 2 bytes | — |
| `19A07` | sub_19A07 | — | 9 | 6 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出 BEL(07h):INT 21h AH=06h、DL=07h,前後保存 DX | — |
| `19A10` | sub_19A10 | — | 25 | 10 | 1 | 1 |  | 待解讀 | — | — | — |
| `19A29` | sub_19A29 | — | 48 | 19 | 2 | 1 |  | 待解讀 | — | — | — |
| `19A59` | sub_19A59 | — | 16 | 9 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>把 byte_280E5 寫回 0000:0712h 的 BIOS 資料區位元組(19A69h 的反向操作) | — |
| `19A69` | sub_19A69 | — | 13 | 7 | 3 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>讀 0000:0712h 的 BIOS 資料區位元組並加一後回傳 | — |
| `19A76` | sub_19A76 | — | 21 | 12 | 3 | 1 |  | 待解讀 | — | — | — |
| `19A8B` | sub_19A8B | — | 23 | 11 | 2 | 0 |  | 待解讀 | — | — | — |
| `19AA2` | sub_19AA2 | — | 32 | 13 | 2 | 2 |  | 待解讀 | — | — | — |
| `19AC2` | sub_19AC2 | — | 12 | 10 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_19A29`（body 共 12 bytes，已逐條讀完） | — |
| `19ACE` | sub_19ACE | — | 50 | 27 | 1 | 1 |  | 待解讀 | — | — | — |
| `19B00` | sub_19B00 | — | 79 | 36 | 2 | 1 |  | 待解讀 | — | — | — |
| `19B4F` | sub_19B4F | — | 46 | 19 | 1 | 2 |  | 待解讀 | — | — | — |
| `19B7D` | sub_19B7D | — | 31 | 13 | 6 | 1 |  | 待解讀 | — | — | — |
| `19B9C` | sub_19B9C | — | 32 | 15 | 6 | 1 |  | 待解讀 | — | — | — |
| `19BBC` | sub_19BBC | — | 52 | 25 | 1 | 3 |  | 待解讀 | — | — | — |
| `19BF0` | sub_19BF0 | — | 129 | 58 | 1 | 4 |  | 待解讀 | — | — | — |
| `19C71` | sub_19C71 | — | 32 | 12 | 1 | 3 |  | 待解讀 | — | — | — |
| `19C91` | sub_19C91 | — | 112 | 51 | 1 | 3 |  | 待解讀 | — | — | — |
| `19D02` | sub_19D02 | — | 28 | 13 | 4 | 2 |  | 待解讀 | — | — | — |
| `19D1E` | sub_19D1E | — | 64 | 32 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 64 bytes，已逐條讀完） | — |
| `19E98` | sub_19E98 | — | 8 | 5 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AL 查 CS:0FC0h 起的轉換表(xlat),前後保存 BX | — |
| `19EB0` | sub_19EB0 | — | 15 | 7 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call sub_1A721`（body 共 15 bytes，已逐條讀完） | — |
| `19EC4` | sub_19EC4 | — | 186 | 71 | 1 | 5 |  | 待解讀 | — | — | — |
| `19F7E` | sub_19F7E | — | 12 | 7 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 12 bytes，已逐條讀完） | — |
| `19F8A` | sub_19F8A | — | 59 | 31 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 59 bytes，已逐條讀完） | — |
| `19FC5` | sub_19FC5 | — | 88 | 48 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 88 bytes，已逐條讀完） | — |
| `1A01D` | sub_1A01D | — | 27 | 14 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 27 bytes，已逐條讀完） | — |
| `1A038` | sub_1A038 | — | 19 | 11 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 19 bytes，已逐條讀完） | — |
| `1A04B` | sub_1A04B | — | 88 | 30 | 1 | 2 |  | 待解讀 | — | — | — |
| `1A0A3` | sub_1A0A3 | — | 11 | 4 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 word_23AE0 − word_23ADC 的差值呼叫 sub_1A10C | — |
| `1A0FF` | sub_1A0FF | — | 13 | 6 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>nibble 合併:AX := (AX >> 4) or (ror(DX,4) and 0F000h) | — |
| `1A10C` | sub_1A10C | — | 13 | 6 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>nibble 拆解:rol(AX,4) 後 AX 取高 12 位、DX 取低 4 位 | — |
| `1A1E9` | sub_1A1E9 | — | 177 | 64 | 2 | 9 |  | 待解讀 | — | — | — |
| `1A29A` | sub_1A29A | — | 76 | 30 | 1 | 3 |  | 待解讀 | — | — | — |
| `1A2E6` | sub_1A2E6 | — | 54 | 20 | 1 | 2 |  | 待解讀 | — | — | — |
| `1A31C` | sub_1A31C | — | 51 | 19 | 2 | 2 |  | 待解讀 | — | — | — |
| `1A34F` | sub_1A34F | — | 72 | 32 | 2 | 2 |  | 待解讀 | — | — | — |
| `1A397` | sub_1A397 | — | 32 | 14 | 1 | 2 |  | 待解讀 | — | — | — |
| `1A3B7` | sub_1A3B7 | — | 12 | 5 | 2 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>呼叫 sub_1A3C3 取得 bx;bx 非 0 時把 cx 與 ss:[bx+2] 互換 | — |
| `1A3C3` | sub_1A3C3 | — | 30 | 15 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 30 bytes，已逐條讀完） | — |
| `1A3E1` | sub_1A3E1 | — | 27 | 10 | 1 | 1 |  | 待解讀 | — | — | — |
| `1A3FC` | sub_1A3FC | — | 12 | 5 | 3 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>回傳 (es:[8] + 0Fh) >> 4,即位元組數無條件進位換算成 paragraph;與 DOS START.EXE:1A42Ch 同義 | — |
| `1A410` | sub_1A410 | — | 97 | 34 | 1 | 5 |  | 待解讀 | — | — | — |
| `1A485` | sub_1A485 | — | 22 | 12 | 1 | 0 |  | 待解讀 | — | — | — |
| `1A49B` | sub_1A49B | — | 63 | 23 | 1 | 1 |  | 待解讀 | — | — | — |
| `1A4DA` | sub_1A4DA | — | 106 | 47 | 1 | 4 |  | 待解讀 | — | — | — |
| `1A544` | sub_1A544 | — | 68 | 32 | 1 | 1 |  | 待解讀 | — | — | — |
| `1A5F1` | sub_1A5F1 | — | 38 | 19 | 1 | 1 |  | 待解讀 | — | — | — |
| `1A617` | sub_1A617 | — | 29 | 13 | 2 | 1 |  | 待解讀 | — | — | — |
| `1A634` | sub_1A634 | — | 23 | 10 | 2 | 1 |  | 待解讀 | — | — | — |
| `1A650` | sub_1A650 | — | 157 | 72 | 1 | 4 |  | 待解讀 | — | — | — |
| `1A721` | sub_1A721 | — | 4 | 3 | 11 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1A72A`，控制權轉交後不返回；先設定 `pop cx`、`pop bx`（body 共 4 bytes，已逐條讀完） | — |
| `1A726` | sub_1A726 | — | 194 | 83 | 3 | 6 |  | 待解讀 | — | — | — |
| `1A7E8` | sub_1A7E8 | — | 14 | 7 | 2 | 2 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出 CS 內的 ASCIIZ 字串:逐字元讀 cs:[bx],非 0 就呼叫 1A82Ah 輸出;與 DOS START.EXE:1A6D4h 同義 | — |
| `1A7F6` | sub_1A7F6 | — | 12 | 5 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1A806`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 12 bytes，已逐條讀完） | — |
| `1A802` | sub_1A802 | — | 14 | 8 | 1 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>輸出一位十進位數字:al := ax ÷ cl,加 30h 後呼叫 1A82Ah,餘數移回 al;與 DOS START.EXE:1A6EEh 同義 | — |
| `1A810` | sub_1A810 | — | 7 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1A817` | sub_1A817 | — | 11 | 6 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and al, 0Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1A822` | sub_1A822 | — | 8 | 4 | 1 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add al, 7`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 8 bytes，已逐條讀完） | — |
| `1A82A` | sub_1A82A | — | 7 | 4 | 4 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>字元輸出:與 DOS START.EXE:1A716h 逐指令相同(INT 21h AH=06h) | — |
| `1A8B2` | sub_1A8B2 | — | 24 | 9 | 4 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$basg$qm3Anyt14Word` 逐位元組相同（24 bytes） | — |
| `1A8CE` | sub_1A8CE | — | 27 | 16 | 9 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$brmul$q7Longintt1` 逐位元組相同（27 bytes） | — |
| `1A8E9` | sub_1A8E9 | — | 110 | 55 | 2 | 2 |  | 待解讀 | — | — | — |
| `1A9E1` | sub_1A9E1 | — | 68 | 26 | 1 | 3 |  | 待解讀 | — | — | — |
| `1AA72` | sub_1AA72 | — | 187 | 79 | 2 | 4 |  | 待解讀 | — | — | — |
| `1AB2D` | sub_1AB2D | — | 179 | 66 | 1 | 4 |  | 待解讀 | — | — | — |
| `1ABE0` | sub_1ABE0 | — | 36 | 15 | 1 | 1 |  | 待解讀 | — | — | — |
| `1AC04` | sub_1AC04 | — | 21 | 9 | 2 | 0 |  | 待解讀 | — | — | — |
| `1AC19` | sub_1AC19 | — | 73 | 27 | 3 | 1 |  | 待解讀 | — | — | — |
| `1AC62` | sub_1AC62 | — | 14 | 7 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>把 AX 拆成兩個 nibble:DX := AX >> 4、AX := AX and 0Fh;與 DOS START.EXE:1AB57h 同義 | — |
| `1AC70` | sub_1AC70 | — | 15 | 7 | 2 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>nibble 合併;與 DOS START.EXE:1AB65h 同義 | — |
| `1AC7F` | sub_1AC7F | — | 26 | 12 | 10 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$basg$qm6Stringt1` 逐位元組相同（26 bytes） | — |
| `1AC99` | sub_1AC99 | — | 36 | 16 | 102 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$basg$qm6Stringt14Byte` 逐位元組相同（36 bytes） | — |
| `1ACCB` | sub_1ACCB | — | 65 | 29 | 8 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Copy$qm6Stringt17Integert3` 逐位元組相同（65 bytes） | — |
| `1AD0C` | sub_1AD0C | — | 44 | 19 | 8 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Concat$qm6Stringt1` 逐位元組相同（44 bytes） | — |
| `1AD84` | sub_1AD84 | — | 43 | 20 | 71 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$bsub$qm6Stringt1` 逐位元組相同（43 bytes） | — |
| `1ADAF` | sub_1ADAF | — | 18 | 8 | 1 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@$basg$qm6String4Char` 逐位元組相同（18 bytes） | — |
| `1AF34` | sub_1AF34 | — | 33 | 13 | 3 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Set@MemberOf$q4Byte` 逐位元組相同（33 bytes） | — |
| `1AFE0` | sub_1AFE0 | — | 4 | 1 | 5 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor di, 8000h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 4 bytes，已逐條讀完） | — |
| `1AFE4` | sub_1AFE4 | — | 195 | 97 | 11 | 2 |  | 待解讀 | — | — | — |
| `1B0A7` | sub_1B0A7 | — | 261 | 123 | 7 | 2 |  | 待解讀 | — | — | — |
| `1B1AC` | sub_1B1AC | — | 119 | 58 | 5 | 2 |  | 待解讀 | — | — | — |
| `1B223` | sub_1B223 | — | 23 | 14 | 7 | 2 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `__RealCmp` 逐位元組相同（23 bytes） | — |
| `1B23A` | sub_1B23A | — | 19 | 10 | 1 | 1 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>多欄位相等比較,與 DOS START.EXE:1B0DAh 逐指令相同 | — |
| `1B24D` | sub_1B24D | — | 63 | 30 | 2 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `__RealFloat` 逐位元組相同（63 bytes） | — |
| `1B28C` | sub_1B28C | — | 92 | 44 | 2 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `__RealTrunc` 逐位元組相同（92 bytes） | — |
| `1B334` | sub_1B334 | — | 10 | 8 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1AFE4`（body 共 10 bytes，已逐條讀完） | — |
| `1B33E` | sub_1B33E | — | 10 | 8 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1AFE0`（body 共 10 bytes，已逐條讀完） | — |
| `1B348` | sub_1B348 | — | 10 | 8 | 1 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1B0A7`（body 共 10 bytes，已逐條讀完） | — |
| `1B352` | sub_1B352 | — | 10 | 8 | 2 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call sub_1B1AC`（body 共 10 bytes，已逐條讀完） | — |
| `1B35C` | sub_1B35C | — | 81 | 39 | 1 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Int$q4Real` 逐位元組相同（81 bytes） | — |
| `1B3AD` | sub_1B3AD | — | 20 | 13 | 1 | 2 |  | 已解讀 | exact | docs/spec/574-pc98-shiftjis-and-text-vram.md<br>呼叫 sub_1B35C 取得三個回傳值(ax/bx/dx → cx/si/di),還原原參數後呼叫 sub_1AFE0 | — |
| `1B737` | sub_1B737 | — | 31 | 18 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp sub_1B0A7`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1B756` | sub_1B756 | — | 79 | 33 | 1 | 3 |  | 待解讀 | — | — | — |
| `1B7A5` | sub_1B7A5 | — | 22 | 10 | 2 | 2 |  | 待解讀 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal @Random$q4Word:body 與 DOS START.EXE:1B645h 逐位元組相同 ⇒ 兩平台亂數演算法一致。影響玩家可見結果,依 spec 566 保留待解讀 | — |
| `1B7F4` | sub_1B7F4 | — | 54 | 22 | 3 | 0 |  | 待解讀 | — | — | — |
| `1B839` | sub_1B839 | — | 82 | 37 | 2 | 1 |  | 待解讀 | — | — | — |
| `1B8B3` | sub_1B8B3 | — | 152 | 74 | 2 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `__Str2Int` 逐位元組相同（152 bytes） | — |
| `1B996` | sub_1B996 | — | 49 | 22 | 1 | 2 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Val__Longint$qm6Stringm7Integer` 逐位元組相同（49 bytes） | — |
| `1B9C7` | sub_1B9C7 | — | 69 | 37 | 4 | 1 |  | 待解讀 | — | — | — |
| `1BA37` | sub_1BA37 | — | 5 | 2 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1BA44`，控制權轉交後不返回；先設定 `mov dx, 0D7B1h`（body 共 5 bytes，已逐條讀完） | — |
| `1BA3C` | sub_1BA3C | — | 5 | 2 | 4 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp short loc_1BA44`，控制權轉交後不返回；先設定 `mov dx, 0D7B2h`（body 共 5 bytes，已逐條讀完） | — |
| `1BA41` | sub_1BA41 | — | 75 | 27 | 1 | 3 |  | 待解讀 | — | — | — |
| `1BA90` | sub_1BA90 | — | 59 | 19 | 5 | 2 |  | 待解讀 | — | — | — |
| `1BACB` | sub_1BACB | — | 17 | 11 | 2 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>與 DOS START.EXE:1B96Bh 同義,結果存進 word_23B08 | — |
| `1BB5E` | sub_1BB5E | — | 90 | 40 | 1 | 1 |  | 待解讀 | — | — | — |
| `1BC46` | sub_1BC46 | — | 36 | 10 | 4 | 1 |  | 待解讀 | — | — | — |
| `1BC6A` | sub_1BC6A | — | 28 | 14 | 5 | 2 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `__GetChar` 逐位元組相同（28 bytes） | — |
| `1BC86` | sub_1BC86 | — | 36 | 10 | 4 | 1 |  | 待解讀 | — | — | — |
| `1BCAA` | sub_1BCAA | — | 15 | 7 | 4 | 1 |  | 待解讀 | — | — | — |
| `1BCB9` | sub_1BCB9 | — | 49 | 24 | 2 | 1 |  | 待解讀 | — | — | — |
| `1BCEA` | sub_1BCEA | — | 41 | 20 | 1 | 4 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@ReadLn$qm4Text` 逐位元組相同（41 bytes） | — |
| `1BD13` | sub_1BD13 | — | 31 | 14 | 3 | 4 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@WriteLn$qm4Text` 逐位元組相同（31 bytes） | — |
| `1BD32` | sub_1BD32 | — | 38 | 14 | 4 | 1 |  | 待解讀 | — | — | — |
| `1BD58` | sub_1BD58 | — | 30 | 15 | 1 | 3 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Read$qm4Text4Char` 逐位元組相同（30 bytes） | — |
| `1BD76` | sub_1BD76 | — | 45 | 20 | 1 | 3 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Write$qm4Text4Char4Word` 逐位元組相同（45 bytes） | — |
| `1BDA3` | sub_1BDA3 | — | 56 | 29 | 1 | 3 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Read$qm4Textm6String4Word` 逐位元組相同（56 bytes） | — |
| `1BDDB` | sub_1BDDB | — | 62 | 32 | 5 | 3 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Write$qm4Textm6String4Word` 逐位元組相同（62 bytes） | — |
| `1BE71` | sub_1BE71 | — | 73 | 34 | 2 | 4 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Write$qm4Text7Longint4Word` 逐位元組相同（73 bytes） | — |
| `1BEBA` | sub_1BEBA | — | 46 | 23 | 3 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Assign$qm4Filem6String` 逐位元組相同（46 bytes） | — |
| `1BEE8` | sub_1BEE8 | — | 101 | 42 | 2 | 2 |  | 待解讀 | — | — | — |
| `1BF69` | sub_1BF69 | — | 37 | 13 | 3 | 2 |  | 待解讀 | — | — | — |
| `1BF8E` | sub_1BF8E | — | 15 | 4 | 5 | 1 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>檢查 es:[di+2] 是否為 0D7B3h 簽章;不符時把 word_23B08 設為 67h;與 DOS START.EXE:1BE2Eh 同義 | — |
| `1BFD3` | sub_1BFD3 | — | 104 | 45 | 2 | 2 |  | 待解讀 | — | — | — |
| `1C03B` | sub_1C03B | — | 48 | 19 | 2 | 2 |  | 待解讀 | — | — | — |
| `1C0B2` | sub_1C0B2 | — | 69 | 32 | 3 | 1 |  | 待解讀 | — | — | — |
| `1C0F7` | sub_1C0F7 | — | 23 | 10 | 1 | 1 |  | 待解讀 | — | — | — |
| `1C15D` | sub_1C15D | — | 35 | 16 | 27 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@Move$qm3Anyt14Word` 逐位元組相同（35 bytes） | — |
| `1C180` | sub_1C180 | — | 20 | 7 | 13 | 0 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@FillChar$qm3Any4Word4Byte` 逐位元組相同（20 bytes） | — |
| `1C250` | sub_1C250 | — | 27 | 17 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 27 bytes，已逐條讀完） | — |
| `1C26B` | sub_1C26B | — | 15 | 9 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：讀寫 `[bp-N]` 區域變數但沒有 `sub sp` 配置框架；這是別的函式被切開的後半段，不是完整函式（body 共 15 bytes，已逐條讀完） | — |
| `1C27A` | sub_1C27A | — | 19 | 8 | 3 | 1 |  | 不阻塞 | — | docs/spec/570-cross-platform-rtl-byte-match.md<br>Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `@UpCase$q4Char` 逐位元組相同（19 bytes） | — |
