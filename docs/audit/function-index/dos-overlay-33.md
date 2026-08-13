# dos-overlay-33 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：唯一實質動作是 `call far ptr 19Ch:2Ah`，參數原樣傳遞（body 共 12 bytes，已逐條讀完） | — |
| `002E` | sub_2E | — | 203 | 75 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `00F9` | sub_F9 | — | 97 | 39 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `015A` | sub_15A | — | 98 | 40 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `01D4` | sub_1D4 | — | 820 | 353 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0508` | sub_508 | — | 278 | 111 | 0 | 1 | ✓ | 待解讀 | — | — | — |
