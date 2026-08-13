# pc98-overlay-07 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADECL2 | 52 | 14 | 0 | 8 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md<br>spec/522-dos-buffer-four-plane-fill.md<br>spec/523-dos-overlay07-vector26-entry.md |
| `008E` | sub_8E | READVAR | 520 | 206 | 2 | 4 | ✓ | 待解讀 | — | — | project-status.md<br>spec/README.md |
| `0296` | sub_296 | ADDRESSVALUE | 129 | 57 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0317` | sub_317 | INITECL | 362 | 134 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0499` | sub_499 | GETECL | 189 | 84 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0556` | sub_556 | GETMONSTERS | 103 | 36 | 0 | 1 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `05BD` | sub_5BD | MAXRANGE | 143 | 52 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `064C` | sub_64C | DRAWHEADBODY | 53 | 22 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `068B` | sub_68B | GODRAWWINDOW | 337 | 115 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `07DC` | sub_7DC | ADDFNC | 37 | 16 | 4 | 0 | ✓ | 待解讀 | — | — | — |
| `0801` | sub_801 | WHICHAREA | 91 | 28 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `085C` | sub_85C | FINDGUY | 97 | 35 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `08BD` | sub_8BD | CHECKSPECIALS | 875 | 292 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0C28` | sub_C28 | STORESPECIALS | 515 | 163 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0E2B` | sub_E2B | STOREVALUE | 465 | 162 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0FFC` | sub_FFC | GETVALUE | 30 | 12 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `101A` | sub_101A | — | 302 | 115 | 2 | 2 |  | 待解讀 | — | — | — |
| `1148` | sub_1148 | STORESTRING | 540 | 189 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1364` | sub_1364 | FINDSTR | 147 | 57 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `13F7` | sub_13F7 | GETSTR | 480 | 186 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `142D` | sub_142D | — | 83 | 30 | 2 | 2 |  | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md |
| `147D` | sub_147D | — | 31 | 13 | 2 | 1 |  | 待解讀 | — | — | — |
| `1669` | sub_1669 | — | 23 | 8 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `1697` | sub_1697 | — | 174 | 70 | 2 | 2 |  | 待解讀 | — | — | — |
| `1745` | sub_1745 | — | 5 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `174A` | sub_174A | — | 30 | 11 | 2 | 1 |  | 待解讀 | — | — | — |
| `1780` | sub_1780 | ECLMENUH | 20 | 10 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1794` | sub_1794 | — | 10 | 4 | 2 | 0 |  | 待解讀 | — | — | — |
| `179E` | sub_179E | — | 5 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `17A3` | sub_17A3 | — | 5 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `17A8` | sub_17A8 | — | 72 | 31 | 2 | 2 |  | 待解讀 | — | — | — |
| `17F4` | sub_17F4 | — | 11 | 6 | 3 | 0 |  | 待解讀 | — | — | — |
| `1808` | sub_1808 | — | 130 | 43 | 2 | 5 |  | 待解讀 | — | — | — |
| `188A` | sub_188A | ECLMENUV | 158 | 75 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1928` | sub_1928 | CHECKSTRING | 81 | 36 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1979` | sub_1979 | — | 133 | 50 | 2 | 1 |  | 待解讀 | — | — | — |
| `19FE` | sub_19FE | CHECKSTATUS | 104 | 37 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1A66` | sub_1A66 | SETUPGOSUBSTACK | 134 | 50 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1AEC` | sub_1AEC | MOVEFORWARD | 145 | 52 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1B89` | sub_1B89 | GODUEL | 584 | 201 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1DD1` | sub_1DD1 | ROBDOUGH | 303 | 82 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1F00` | sub_1F00 | ROBSTUFF | 176 | 62 | 0 | 3 | ✓ | 待解讀 | — | — | spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `1FB0` | sub_1FB0 | NONEXT | 277 | 138 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `20C5` | sub_20C5 | CHARSPEED | 205 | 76 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2235` | sub_2235 | KILLTHEDUDE | 516 | 208 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2439` | sub_2439 | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
