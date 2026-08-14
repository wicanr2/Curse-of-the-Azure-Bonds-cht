# pc98-overlay-29 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPORTRAIT | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-17.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/754-small-predicates-and-wrappers.md |
| `000C` | sub_C | SHOWPORTRAIT | 133 | 51 | 3 | 1 | ✓ | 已解讀 | exact | docs/spec/785-cross-platform-pairs-third-batch.md<br>retf 0Ah：與 DOS overlay-29:000Ch 51 條同形，DS:7F05h、緩衝 DS:165Ch/166Ch、resident 呼叫改為 2A8h:0C32h/77Dh/0B7Dh/0A86h(差異 10 條，已逐條列出) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-29.md<br>spec/753-small-utility-routines.md<br>spec/754-small-predicates-and-wrappers.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `0091` | sub_91 | SHOWHEAD | 34 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_C`（body 共 34 bytes，已逐條讀完） | audit/function-index/dos-overlay-29.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `00B3` | sub_B3 | SHOWBODY | 37 | 18 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>包裝(retf 8)：與 DOS overlay-29:00B3h 位元組相同 | audit/function-index/dos-overlay-29.md<br>audit/function-index/pc98-overlay-29.md<br>spec/754-small-predicates-and-wrappers.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `010B` | sub_10B | LOADSEQUENCE | 847 | 340 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `045A` | sub_45A | DISPOSESEQUENCE | 109 | 42 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/760-item-effect-flag-clear-and-file-exists.md<br>逐項釋放後清空(retf 4)：p^[0] 為 0 就離開；否則由高索引往低，對 p + 8i − 1 叫 02A8h:10D5h，最後 FillChar(p^, 43h, 0)、DS:0A313h:=0、DS:0A325h:=0FFh。⚠ 計數器用全域 DS:0A9D9h 而非 local，不可重入 | audit/function-index/pc98-overlay-29.md<br>spec/758-morale-field-0f7h-round-trip.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `04E0` | sub_4E0 | LOADCHARACTERPORTRAIT | 281 | 120 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `05F9` | sub_5F9 | SHOWCHARACTERPORTRAIT | 84 | 36 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>兩種畫法的分岔(retf 6)：指標在 DS:0A30Ah 與 DS:0A30Fh，其餘同 DOS overlay-29:06E1h | spec/758-morale-field-0f7h-round-trip.md |
| `066C` | sub_66C | SHOW3DSPRITE | 184 | 74 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `072B` | sub_72B | LOADBIGPIC | 76 | 36 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>載入 bigpic(retf 2)：先叫本模組 045Ah(@DS:0A2C6h)；DS:0A327h 非 NIL 就離開；否則組出 CS:0724h 的 'bigpic' 字串，叫 02A8h:025Bh(參數, 0, 0, @DS:0A327h, 0)，最後 DS:0A32Bh := 參數(該指標的哨兵 byte) | spec/758-morale-field-0f7h-round-trip.md |
| `0777` | sub_777 | SHOWBIGPIC | 43 | 17 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>retf，無參數：019Eh:0384h() 後 02A8h:0A86h(1, 1, 遠指標@DS:0A327h)，再 02A8h:10D5h(@DS:0A327h)。最後這一步 DOS 版沒有 | spec/753-small-utility-routines.md<br>spec/760-item-effect-flag-clear-and-file-exists.md |
| `07A2` | sub_7A2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
