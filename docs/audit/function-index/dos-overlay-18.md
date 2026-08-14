# dos-overlay-18 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 50 | 19 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>送出命令 0Ch(retf 6)：填 DS:4FCAh 起的 16-byte 命令區塊(+1 命令碼=0Ch、+0=第三個參數、+3=0、+4=第一個參數、+6=第二個參數)後呼叫 097Fh:000Bh(@DS:4FCAh, 10h) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/612-ecl-main-loop.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/754-small-predicates-and-wrappers.md |
| `0032` | sub_32 | — | 56 | 21 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>送出命令 0Dh(retf 4)：同一個 16-byte 命令區塊，命令碼 0Dh，呼叫後把 DS:4FCAh 當回傳值讀出。全 DOS overlay 只有 overlay-18 這兩支寫 4FCBh、也只有這兩支呼叫 097Fh:000Bh | audit/embedded-strings.md<br>audit/function-index/dos-overlay-18.md<br>audit/function-index/pc98-overlay-02.md<br>spec/754-small-predicates-and-wrappers.md |
| `006A` | sub_6A | — | 82 | 37 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>依 DS:4FE6h(1 或 2)選一對顯示常式:模式 1 走 0297:2171、模式 2 走 0297:21B0,中間各夾一次 09AB:029E(1) | audit/embedded-strings.md<br>spec/585-ecl-goto-and-display-mode-pair.md |
| `00BC` | sub_BC | — | 1059 | 382 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `04DF` | sub_4DF | — | 673 | 234 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0780` | sub_780 | — | 195 | 75 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0843` | sub_843 | — | 408 | 151 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `09DB` | sub_9DB | — | 384 | 150 | 1 | 4 | ✓ | 已解讀 | exact | 1<br>結局動畫的主迴圈(retf，無參數)：GetMem(DS:4ACEh, 0AC8h=2760) → 迴圈直到 KeyPressed(9ABh:2FAh) 為真才 ReadKey 存進 DS:4AE2h 離開 → FreeMem。每一輪先 Random(2710h) 需 < 1 才動作（1/10000）；FillChar(DS:4AE3h, 3, 1)；筆數 = Random(2)+1，每筆 DS:4AE3h+i := Random(5)+2；起點固定 (41h, 41h)，終點 DS:4AEBh := Random(14h)+23h、DS:4AEDh := −(Random(5)+32h)（另存一份到 4AEFh/4AF1h）；接著以模式 0 呼叫 sub_843h + sub_BCh，還原座標後再以模式 1 呼叫 sub_843h + sub_780h——擦掉再畫上的兩段式更新 | audit/function-index/pc98-overlay-18.md |
| `0B5F` | sub_B5F | — | 267 | 109 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `10FF` | sub_10FF | — | 1265 | 649 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
| `15F0` | sub_15F0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
