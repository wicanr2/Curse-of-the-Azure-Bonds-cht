# pc98-overlay-06 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSHOP | 37 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 6 個呼叫，沒有其他動作：`call far ptr 10Dh:84h`、`call far ptr 0FAh:7Ah`、`call far ptr 164h:57h`、`call far ptr 14Ah:101h`、`call sub_6E0`、`call far ptr 19Ah:2Ah`（body 共 37 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md |
| `0037` | sub_37 | — | 860 | 364 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-02.md |
| `03A0` | sub_3A0 | GIVEITEM | 269 | 96 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `04BE` | sub_4BE | — | 440 | 166 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `06A4` | nullsub_1 | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `iret`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `06E0` | sub_6E0 | — | 22 | 10 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `iret`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 22 bytes，已逐條讀完） | — |
| `0778` | sub_778 | SHOPMAINMENU | 727 | 307 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0A4F` | sub_A4F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
