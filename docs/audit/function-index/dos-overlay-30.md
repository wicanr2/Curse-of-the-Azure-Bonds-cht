# dos-overlay-30 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 20Bh:34h 後交給 19Ch:2Ah | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `0011` | sub_11 | — | 346 | 150 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>context/50-log-2026-08-09-13.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md<br>spec/README.md |
| `016B` | sub_16B | — | 33 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>四個 byte 的 setter(retf 8)：四個 word 參數只取低位 byte，依序寫 DS:720Ah/720Bh/720Ch/720Dh(第一個宣告的參數在 [bp+0Ch] → 720Ah) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-30.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/753-small-utility-routines.md |
| `018C` | sub_18C | — | 700 | 274 | 1 | 2 | ✓ | 待解讀 | — | — | spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/687-far-call-flattening-and-stack-leftover.md |
| `0448` | sub_448 | — | 270 | 108 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0556` | sub_556 | — | 49 | 18 | 4 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>範圍判斷(retf 4)：兩個有號參數都落在 0..15(含端點)才回 1，否則 0。範圍是 16×16 | audit/function-index/dos-overlay-30.md<br>audit/function-index/pc98-overlay-30.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md<br>spec/754-small-predicates-and-wrappers.md |
| `0587` | sub_587 | — | 310 | 123 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `06BD` | sub_6BD | — | 265 | 106 | 3 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `07C6` | sub_7C6 | — | 123 | 46 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/763-dungeon-map-second-plane-and-stone-to-flesh.md<br>讀迷宮地圖第二平面(retf 4)：先用本模組 0556h 檢查 0..15；不在範圍且 DS:8B5Eh 是 0 或 0Ah 就回 0，否則座標環繞(>0Fh→0、<0→0Fh)後讀 DS:7206h^[200h + (row shl 4) + col]。與 spec 761 的清牆(+300h)共用索引式，兩個平面各 256 bytes 首尾相接。DS:8B5Eh 決定地圖邊界要不要接起來 | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/517-reverse-engineering-gap-inventory.md<br>spec/518-dos-start-ecl-call-address-space-audit.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `0841` | sub_841 | — | 2059 | 822 | 0 | 6 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/517-reverse-engineering-gap-inventory.md<br>spec/518-dos-start-ecl-call-address-space-audit.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `104C` | sub_104C | — | 708 | 282 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `133A` | sub_133A | — | 318 | 125 | 0 | 1 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/521-dos-getmem-buffer-owner.md<br>spec/522-dos-buffer-four-plane-fill.md<br>spec/524-dos-overlay30-geo-loader-source.md |
| `1478` | sub_1478 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
