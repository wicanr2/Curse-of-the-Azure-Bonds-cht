# dos-overlay-04 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 71 | 16 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 9 個相依 unit 的 0000h — overlay-21、overlay-19、overlay-26、overlay-24、overlay-07、overlay-34、overlay-22、overlay-23、overlay-25。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/dos-overlay-25.md |
| `0047` | sub_47 | — | 172 | 50 | 6 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `00F3` | sub_F3 | — | 397 | 143 | 7 | 1 | ✓ | 待解讀 | — | — | — |
| `0280` | sub_280 | — | 134 | 53 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0324` | sub_324 | — | 217 | 75 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `043D` | sub_43D | — | 513 | 177 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `063E` | sub_63E | — | 251 | 81 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0892` | sub_892 | — | 205 | 77 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `097B` | sub_97B | — | 202 | 72 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0A63` | sub_A63 | — | 175 | 48 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0B12` | sub_B12 | — | 506 | 220 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `0DD6` | sub_DD6 | — | 563 | 243 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `1009` | sub_1009 | — | 22 | 8 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 22 bytes，已逐條讀完） | — |
| `101F` | sub_101F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
