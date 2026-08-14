# dos-overlay-33 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ch:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md |
| `002E` | sub_2E | — | 203 | 75 | 0 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>spec/749-combat-teardown-and-battlefield-grid.md |
| `00F9` | sub_F9 | — | 97 | 39 | 0 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `015A` | sub_15A | — | 98 | 40 | 0 | 1 | ✓ | 待解讀 | — | — | spec/749-combat-teardown-and-battlefield-grid.md |
| `01D4` | sub_1D4 | — | 820 | 353 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0508` | sub_508 | — | 278 | 111 | 0 | 1 | ✓ | 待解讀 | — | — | spec/749-combat-teardown-and-battlefield-grid.md |
| `061E` | sub_61E | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
