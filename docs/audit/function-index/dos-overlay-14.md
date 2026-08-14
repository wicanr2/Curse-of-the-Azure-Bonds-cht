# dos-overlay-14 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 62 | 16 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call far ptr 141h:9Dh`、`call far ptr 14Dh:101h`、`call far ptr 0FDh:7Ah`、`call far ptr 175h:2Ah`、`call loc_EB2+4`、`call far ptr 196h:43h`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `003E` | sub_3E | — | 277 | 6 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/761-filled-export-and-dungeon-wall-bits.md<br>清掉一個方向的牆面位元(retf 6)：格 := DS:7206h^[300h + (row shl 4) + col]；dir=6 → and 3Fh、4 → and 0CFh、2 → and 0F3h、0 → and 0FCh。每列 16 格、每格 1 byte、四個方向各 2 bits。case 沒有 else，奇數方向碼整支不做事。⚠ 先前 spec 569 的註記(「只準備參數並執行 jmp」)是根據被截斷成 6 條指令的匯出寫的，已由補洞重匯出更正(見 spec 761) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/pc98-overlay-14.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md |
| `0153` | sub_153 | — | 309 | 127 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0288` | sub_288 | — | 122 | 43 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0302` | sub_302 | — | 742 | 303 | 1 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `05E8` | sub_5E8 | — | 201 | 81 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `06B1` | sub_6B1 | — | 67 | 25 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>找格子(retf 6)：f(v: byte, p: 遠指標) 在 p^[1Eh..1Eh+53h] 線性搜尋等於 v 的索引，找不到回 0FFh。與 overlay-15:0188h 同一個 84 格陣列 | audit/function-index/dos-overlay-14.md<br>audit/function-index/pc98-overlay-14.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `06F4` | sub_6F4 | — | 85 | 33 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>找持有某值的隊員(retf 2，回遠指標)：走隊伍鏈，對每個成員叫 06B1h 找格子(v, 成員)，回傳值不是 0FFh 就回傳該成員；走完沒找到回 NIL。spec 756 的 0749h 是其呼叫端 | audit/function-index/dos-overlay-14.md<br>audit/function-index/pc98-overlay-14.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `0749` | sub_749 | — | 69 | 29 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>移除值為 1Fh 的那一格(retf，無參數)：p := 本模組 06F4h(1Fh)；p 非 NIL 時 i := 找格子(1Fh, p)(overlay-14:06B1h)，把 p^[1Eh+i] 歸零並回 true，否則回 false | audit/function-index/dos-overlay-14.md<br>audit/function-index/pc98-overlay-14.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `078E` | sub_78E | — | 174 | 55 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `083C` | sub_83C | — | 196 | 70 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/770-party-step-and-hug-attack.md<br>隊伍在迷宮裡前進一格(retf，無參數)：DS:720Fh(x) += byte[2694h + DS:7211h(面向)]、DS:7210h(y) += byte[269Dh + 面向]，兩軸各自無條件環繞(<0 → 0Fh、>0Fh → 0，有號比較)；DS:7212h := 017Fh:0034h(x, y, 面向)、FillChar(@DS:7588h, 3, 1)、DS:7213h := 017Fh:003Eh(y, x)；最後依 DS:4F9Dh^[594h] bit 0 二選一叫 0108h:002Ah。⚠ 移動一律環繞，沒有 overlay-30:07C6h 那道 DS:8B5Eh 判斷 | spec/770-party-step-and-hug-attack.md |
| `090A` | sub_90A | — | 644 | 248 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BAF` | sub_BAF | — | 780 | 334 | 0 | 8 | ✓ | 待解讀 | — | — | spec/749-combat-teardown-and-battlefield-grid.md |
| `0EBB` | sub_EBB | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
