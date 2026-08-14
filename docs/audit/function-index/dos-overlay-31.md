# dos-overlay-31 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0007` | sub_7 | — | 46 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>Sign(x: integer): integer(retf 2)：x<0→−1、x>0→1、x=0→0，有號比較。與 PC-98 同位址那支 45 個 byte 完全相同 | audit/function-index/dos-overlay-31.md<br>audit/function-index/pc98-overlay-24.md<br>spec/753-small-utility-routines.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0035` | sub_35 | — | 360 | 148 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-12.md |
| `019D` | sub_19D | — | 169 | 58 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/787-bresenham-record-and-edge-clamp.md<br>Bresenham 走線初始化(retf 4)：r^[0Eh]/[10h] := 起點 x/y、r^[0Ah]/[0Ch] := |dx|/|dy|、r^[12h]/[14h] := Sign(dx)/Sign(dy)(overlay-31:0007h)、r^[8] := 0、r^[16h] := 0。記錄共 17h bytes：+0/+2 起點、+4/+6 終點、+8 誤差、+0Ah/+0Ch 絕對差、+0Eh/+10h 目前、+12h/+14h 步進、+16h 結束旗標。⚠ 兩平台逐位元組相同(差異 0 條) | audit/function-index/dos-overlay-12.md<br>audit/function-index/pc98-overlay-12.md<br>audit/function-index/pc98-overlay-31.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0246` | sub_246 | — | 420 | 132 | 1 | 1 | ✓ | 已解讀 | exact | 854<br>視線的 Bresenham 逐格推進(retf 4)：狀態記錄 +04h/+06h 終點、+08h 誤差、+0Ah/+0Ch 主軸判斷、+0Eh/+10h 目前座標、+12h/+14h 步進、+16h 累計步數（直走+2、斜走+3，與 spec 837 的移動花費同一套半格度量）、+17h 方向碼（查 DS:2558h 3×3 表） | spec/854-bresenham-stepper-and-direction-encoding.md |
| `03EA` | sub_3EA | — | 352 | 144 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `054A` | sub_54A | — | 921 | 411 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `08E3` | sub_8E3 | — | 717 | 294 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `0BB0` | sub_BB0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
