# pc98-overlay-01 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 81 | 28 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>可中斷的計時等待(retf 2)：07E3h:02C7h／0418h:133Dh／08EEh:03B3h，其餘與 DOS overlay-01:0000h 同 | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-01.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0321` | sub_321 | — | 1551 | 775 | 1 | 0 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0936` | sub_936 | DOINTRO | 364 | 162 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0AA2` | sub_AA2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
