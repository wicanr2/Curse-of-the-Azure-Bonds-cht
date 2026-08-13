# pc98-overlay-03 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPROTECT | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：唯一實質動作是 `call far ptr 19Ah:2Ah`，參數原樣傳遞（body 共 12 bytes，已逐條讀完） | spec/569-small-function-batch-reading.md |
| `000C` | sub_C | GAMECLEAN | 18 | 8 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `0131` | sub_131 | DOPROTECT | 715 | 310 | 0 | 1 | ✓ | 待解讀 | — | — | — |
