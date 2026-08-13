# pc98-overlay-29 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPORTRAIT | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
| `000C` | sub_C | SHOWPORTRAIT | 133 | 51 | 3 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0091` | sub_91 | SHOWHEAD | 34 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_C`（body 共 34 bytes，已逐條讀完） | — |
| `00B3` | sub_B3 | SHOWBODY | 37 | 18 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `010B` | sub_10B | LOADSEQUENCE | 847 | 340 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `045A` | sub_45A | DISPOSESEQUENCE | 109 | 42 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `04E0` | sub_4E0 | LOADCHARACTERPORTRAIT | 281 | 120 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `05F9` | sub_5F9 | SHOWCHARACTERPORTRAIT | 84 | 36 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `066C` | sub_66C | SHOW3DSPRITE | 184 | 74 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `072B` | sub_72B | LOADBIGPIC | 76 | 36 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0777` | sub_777 | SHOWBIGPIC | 43 | 17 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `07A2` | sub_7A2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
