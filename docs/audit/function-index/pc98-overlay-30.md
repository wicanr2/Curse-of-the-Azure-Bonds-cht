# pc98-overlay-30 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTHREED | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 2 個呼叫，沒有其他動作：`call far ptr 232h:34h`、`call far ptr 19Ah:2Ah`（body 共 17 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-30.md<br>audit/function-index/pc98-overlay-02.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md |
| `0011` | sub_11 | — | 346 | 150 | 1 | 2 | ✓ | 已解讀 | exact | 781<br>同 DOS overlay-30:00011h（同位址、助憶碼序列完全相同）。只差七條呼叫目標（sub_60Dh↔sub_6BDh、232h:2Fh↔20Bh:2Fh、2A8h:0B7Dh↔297h:1316h） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/pc98-overlay-02.md<br>audit/function-index/pc98-overlay-30.md<br>context/50-log-2026-08-09-13.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `016B` | sub_16B | SET3DCOLORS | 33 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>四個 byte 的 setter(retf 8)：依序寫 DS:0A2A4h..0A2A7h，與 DOS overlay-30:016Bh 同 | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-30.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/753-small-utility-routines.md |
| `018C` | sub_18C | CLEAR3DVIEW | 531 | 206 | 1 | 2 | ✓ | 待解讀 | — | — | spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/687-far-call-flattening-and-stack-leftover.md |
| `039F` | sub_39F | — | 270 | 108 | 1 | 1 | ✓ | 已解讀 | exact | 780<br>同 DOS overlay-30:00448h（助憶碼序列完全相同、0 個差異塊）。6 條差異全是位址：三張 10 bytes 平行表 DS:16ACh/16B6h/16C0h↔0AD8h/0AE2h/0AECh、DS:0A29Ch↔7202h、貼圖呼叫 232h:2Fh↔20Bh:2Fh | — |
| `04AD` | sub_4AD | — | 49 | 18 | 4 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>範圍判斷(retf 4)：兩個有號參數都落在 0..15 才回 1。與 DOS overlay-30:0556h 同 | audit/embedded-strings.md<br>spec/754-small-predicates-and-wrappers.md |
| `04DE` | sub_4DE | BLOCKCODE | 303 | 122 | 0 | 3 | ✓ | 已解讀 | exact | 781<br>同 DOS overlay-30:00587h（module_align 對齊，相似度 0.996）。唯一序列差異是 DOS 多一條 xor ah, ah，其餘只差一條呼叫目標 | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/525-pc98-tempsearch-display-state.md<br>spec/528-pc98-moveparty-action-transaction-boundary.md |
| `060D` | sub_60D | WALLCODE | 259 | 105 | 3 | 2 | ✓ | 已解讀 | exact | 778<br>同 DOS overlay-30:006BDh（讀迷宮某方向的 4-bit 碼）。唯一序列差異是 xor ah, ah 的有無，其餘是 DS:0BDF0h↔8B5Eh、DS:0A2A0h↔7206h | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/525-pc98-tempsearch-display-state.md |
| `0710` | sub_710 | SPECIALCODE | 123 | 46 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/763-dungeon-map-second-plane-and-stone-to-flesh.md<br>與 DOS overlay-30:07C6h（entry#6）助憶碼序列完全相同，語意同該筆：讀迷宮地圖第二平面(retf 4)：先用本模組 0556h 檢查 0..15；不在範圍且 DS:8B5Eh 是 0 或 0Ah 就回 0，否則座標環繞(>0Fh→0、<0→0Fh)後讀 DS:7206h^[200h + (row shl 4) + col]。與 spec 761 的清牆(+300h)共用索引式，兩個平面各 256 bytes 首尾相接。DS:8B5Eh 決定地圖邊界要不要接起來 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `078B` | sub_78B | BUILDVIEW | 1977 | 825 | 0 | 6 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0F8F` | sub_F8F | LOADWALLSET | 666 | 268 | 0 | 1 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `1253` | sub_1253 | LOAD3DMAP | 270 | 105 | 0 | 1 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md |
| `1361` | sub_1361 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/524-dos-overlay30-geo-loader-source.md |
