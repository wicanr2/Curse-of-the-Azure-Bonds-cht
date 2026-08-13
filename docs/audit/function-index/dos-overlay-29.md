# dos-overlay-29 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ch:2Ah`（body 共 12 bytes，已逐條讀完） | — |
| `000C` | sub_C | — | 133 | 51 | 3 | 1 |  | 待解讀 | — | — | — |
| `0091` | sub_91 | — | 34 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_C`（body 共 34 bytes，已逐條讀完） | — |
| `00B3` | sub_B3 | — | 37 | 18 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `010B` | sub_10B | — | 1104 | 453 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `055B` | sub_55B | — | 105 | 41 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `05DD` | sub_5DD | — | 260 | 110 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `06E1` | sub_6E1 | — | 84 | 36 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `0754` | sub_754 | — | 184 | 74 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0813` | sub_813 | — | 110 | 51 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0881` | sub_881 | — | 33 | 13 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `08A2` | sub_8A2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
