# dos-overlay-01 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 81 | 28 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>可中斷的計時等待(retf 2)：06EAh:024Ch() 開頭結尾成對；起點 := 0542h:0994h()(32-bit 計數)，終點 := 起點 + 參數*100；迴圈中 09ABh:02FAh() 非 0 就提早離開，否則等到計數 >= 終點(無號 32-bit 比較) | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `01DF` | sub_1DF | — | 1103 | 551 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `0634` | sub_634 | — | 320 | 142 | 0 | 3 | ✓ | 待解讀 | — | — | spec/749-combat-teardown-and-battlefield-grid.md |
| `0774` | sub_774 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
