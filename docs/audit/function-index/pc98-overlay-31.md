# pc98-overlay-31 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADLOS | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-24.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0007` | sub_7 | SGN | 46 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>Sign(x: integer): integer(retf 2)：x<0→−1、x>0→1、x=0→0，有號比較。與 DOS 同位址那支 45 個 byte 完全相同 | audit/function-index/dos-overlay-31.md<br>audit/function-index/pc98-overlay-24.md<br>spec/753-small-utility-routines.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0035` | sub_35 | — | 360 | 148 | 1 | 1 | ✓ | 已解讀 | exact | 852<br>同 DOS overlay-31:00035h（同位址、助憶碼序列完全相同）。14 條差異全是 DS:9F2Eh/9F30h↔6E94h/6E96h 與 Move 呼叫目標 | audit/function-index/dos-overlay-12.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-12.md<br>audit/function-index/pc98-overlay-31.md<br>spec/852-target-list-source.md<br>spec/860-ai-target-reachability-scan.md |
| `019D` | sub_19D | STARTVECTOR | 169 | 58 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/787-bresenham-record-and-edge-clamp.md<br>Bresenham 走線初始化(retf 4)：與 DOS overlay-31:019Dh 58 條**逐位元組相同**，無任何運算元差異 | audit/function-index/dos-overlay-12.md<br>audit/function-index/pc98-overlay-12.md<br>audit/function-index/pc98-overlay-31.md<br>spec/787-bresenham-record-and-edge-clamp.md |
| `0246` | sub_246 | STEPVECTOR | 420 | 132 | 1 | 1 | ✓ | 已解讀 | exact | 854<br>同 DOS overlay-31:00246h（同位址）。整支只差一條表基底 DS:47E6h↔2558h，表內容逐 byte 相同 | audit/function-index/pc98-overlay-31.md<br>spec/854-bresenham-stepper-and-direction-encoding.md |
| `03EA` | sub_3EA | LOSEXISTS | 352 | 144 | 1 | 3 | ✓ | 已解讀 | exact | 854<br>同 DOS overlay-31:003EAh（同位址、助憶碼序列完全相同）。只差三條表基底 DS:48ADh/48AEh↔26A3h/26A4h | audit/function-index/dos-overlay-31.md<br>audit/function-index/pc98-overlay-31.md<br>spec/860-ai-target-reachability-scan.md |
| `054A` | sub_54A | INARC | 910 | 410 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-31.md<br>spec/860-ai-target-reachability-scan.md |
| `08D8` | sub_8D8 | SCAN | 717 | 294 | 0 | 4 | ✓ | 已解讀 | exact | 860<br>同 DOS overlay-31:008E3h（助憶碼序列完全相同）。15 條差異全是位址：18Ch↔189h、座標快取 66A3h↔9740h−3、結果表 6E94h↔9F30h−2 | spec/860-ai-target-reachability-scan.md |
| `0BA5` | sub_BA5 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
