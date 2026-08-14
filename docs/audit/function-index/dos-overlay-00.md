# dos-overlay-00 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0017` | sub_17 | — | 58 | 23 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>retf 2：0A54h:0634h(0Eh, 0, @本地緩衝區, CS:0000h)、0542h:0946h()、overlay-16 entry#11()，DS:5CF1h 非 0 時再叫 06EAh:0000h；回傳 [bp-2]。⚠ 本函式沒有任何指令寫過 bp-2，是否為堆疊殘值取決於 0A54h:0634h 實際寫入長度，實作前需實機確認 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-00.md<br>audit/function-index/pc98-overlay-00.md<br>spec/753-small-utility-routines.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `0051` | sub_51 | — | 20 | 9 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>安裝回呼(retf，無參數)：把遠指標 0039h:0025h 寫進 DS:47C4h；0039h:0025h 是 VROOMM stub → overlay-00 entry#1 @0017h | audit/embedded-strings.md<br>spec/753-small-utility-routines.md |
