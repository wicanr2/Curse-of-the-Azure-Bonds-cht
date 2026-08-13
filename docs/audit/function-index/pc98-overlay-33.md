# pc98-overlay-33 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSQRPAK24 | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md |
| `004C` | sub_4C | LOAD24X24SET | 539 | 206 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0267` | sub_267 | PUT24X24SYMBOL | 97 | 39 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `02E8` | sub_2E8 | DISPOSEFIGURE | 116 | 48 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `038F` | sub_38F | LOADFIGURE | 1050 | 451 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `07A9` | sub_7A9 | PUTFIGURE | 285 | 114 | 0 | 1 | ✓ | 待解讀 | — | — | — |
