# pc98-overlay-00 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `001B` | sub_1B | HEAPERRORHANDLER | 58 | 23 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>retf 2：與 DOS overlay-00:0017h 同構(0A65h:062Fh、0418h:12EFh、overlay-16 entry#11、DS:8CF7h、07E3h:0077h)；同樣回傳未被本函式寫過的 [bp-2] | audit/function-index/pc98-overlay-16.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `0055` | sub_55 | — | 19 | 8 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 19 bytes，已逐條讀完） | audit/function-index/pc98-overlay-16.md |
