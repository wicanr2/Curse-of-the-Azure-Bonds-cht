# dos-overlay-18 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 50 | 19 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>送出命令 0Ch(retf 6)：填 DS:4FCAh 起的 16-byte 命令區塊(+1 命令碼=0Ch、+0=第三個參數、+3=0、+4=第一個參數、+6=第二個參數)後呼叫 097Fh:000Bh(@DS:4FCAh, 10h) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/612-ecl-main-loop.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/754-small-predicates-and-wrappers.md |
| `0032` | sub_32 | — | 56 | 21 | 3 | 0 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>送出命令 0Dh(retf 4)：同一個 16-byte 命令區塊，命令碼 0Dh，呼叫後把 DS:4FCAh 當回傳值讀出。全 DOS overlay 只有 overlay-18 這兩支寫 4FCBh、也只有這兩支呼叫 097Fh:000Bh | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>spec/754-small-predicates-and-wrappers.md |
| `006A` | sub_6A | — | 82 | 37 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/585-ecl-goto-and-display-mode-pair.md<br>依 DS:4FE6h(1 或 2)選一對顯示常式:模式 1 走 0297:2171、模式 2 走 0297:21B0,中間各夾一次 09AB:029E(1) | audit/embedded-strings.md<br>spec/585-ecl-goto-and-display-mode-pair.md |
| `00BC` | sub_BC | — | 1059 | 382 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `04DF` | sub_4DF | — | 673 | 234 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0780` | sub_780 | — | 195 | 75 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0843` | sub_843 | — | 408 | 151 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `09DB` | sub_9DB | — | 384 | 150 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B5F` | sub_B5F | — | 267 | 109 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `10FF` | sub_10FF | — | 1265 | 649 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `15F0` | sub_15F0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
