# dos-overlay-07 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 52 | 14 | 0 | 5 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md<br>spec/522-dos-buffer-four-plane-fill.md<br>spec/523-dos-overlay07-vector26-entry.md |
| `0034` | sub_34 | — | 319 | 118 | 2 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/536-tilverton-sewers-e2-search-route.md |
| `0173` | sub_173 | — | 137 | 58 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `01FC` | sub_1FC | — | 357 | 133 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0380` | sub_380 | — | 214 | 92 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0456` | sub_456 | — | 103 | 36 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `04BD` | sub_4BD | — | 149 | 53 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0552` | sub_552 | — | 53 | 22 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `0591` | sub_591 | — | 343 | 117 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `06E8` | sub_6E8 | — | 37 | 16 | 4 | 0 | ✓ | 待解讀 | — | — | — |
| `070D` | sub_70D | — | 40 | 16 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0735` | sub_735 | — | 91 | 28 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `0790` | sub_790 | — | 97 | 35 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `07F1` | sub_7F1 | — | 892 | 293 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0B6D` | sub_B6D | — | 515 | 163 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0D70` | sub_D70 | — | 302 | 107 | 0 | 3 | ✓ | 待解讀 | — | — | spec/547-normal-beholder-cave-presentation-state.md |
| `0E98` | sub_E98 | — | 160 | 53 | 2 | 2 |  | 待解讀 | — | — | — |
| `0F3A` | sub_F3A | — | 339 | 128 | 1 | 3 |  | 待解讀 | — | — | — |
| `108D` | sub_108D | — | 545 | 190 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `12AE` | sub_12AE | — | 430 | 167 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `145C` | sub_145C | — | 508 | 194 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1652` | sub_1652 | — | 77 | 31 | 2 | 0 |  | 待解讀 | — | — | — |
| `169F` | sub_169F | — | 9 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `16D3` | sub_16D3 | — | 173 | 68 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `1775` | sub_1775 | — | 5 | 2 | 2 | 0 |  | 待解讀 | — | — | — |
| `177A` | sub_177A | — | 63 | 25 | 2 | 0 |  | 待解讀 | — | — | — |
| `17C4` | sub_17C4 | — | 6 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `17EA` | sub_17EA | — | 58 | 27 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1824` | sub_1824 | — | 13 | 6 | 3 | 1 |  | 待解讀 | — | — | — |
| `18EE` | sub_18EE | — | 141 | 67 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `197B` | sub_197B | — | 214 | 86 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1A51` | sub_1A51 | — | 104 | 37 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1AB9` | sub_1AB9 | — | 134 | 50 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1B3F` | sub_1B3F | — | 151 | 53 | 0 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/518-dos-start-ecl-call-address-space-audit.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `1BE0` | sub_1BE0 | — | 566 | 194 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1E16` | sub_1E16 | — | 303 | 82 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1F45` | sub_1F45 | — | 176 | 62 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1FF5` | sub_1FF5 | — | 322 | 139 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2137` | sub_2137 | — | 205 | 76 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2252` | sub_2252 | — | 335 | 134 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `23A1` | sub_23A1 | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
