# pc98-overlay-32 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADTACMAP | 17 | 7 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 2 個呼叫，沒有其他動作：`call loc_1982+1`、`call loc_19C8+2`（body 共 17 bytes，已逐條讀完） | — |
| `0011` | sub_11 | CALCBIGOFFSET | 104 | 41 | 7 | 1 | ✓ | 待解讀 | — | — | — |
| `0079` | sub_79 | CALCSCREENCOORDS | 123 | 51 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `00F4` | sub_F4 | SETCOMBATCOLORS | 29 | 15 | 0 | 0 | ✓ | 待解讀 | — | — | spec/585-ecl-goto-and-display-mode-pair.md |
| `0111` | sub_111 | RESETCOMBATCOLORS | 29 | 15 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `012E` | sub_12E | SHOWCURSOR | 428 | 190 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `02DA` | sub_2DA | — | 409 | 184 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `0473` | sub_473 | HIDECURSOR | 612 | 255 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `06D7` | sub_6D7 | CALCWHOISWHERE | 286 | 122 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `07F5` | sub_7F5 | FINDOBJECT | 111 | 43 | 5 | 1 | ✓ | 待解讀 | — | — | — |
| `0864` | sub_864 | PUTMAPSYMBOL | 442 | 196 | 3 | 5 | ✓ | 待解讀 | — | — | — |
| `0A1E` | sub_A1E | ONVISSCREEN | 49 | 18 | 6 | 1 | ✓ | 待解讀 | — | — | — |
| `0A4F` | sub_A4F | ONSCREEN | 186 | 79 | 6 | 4 | ✓ | 待解讀 | — | — | — |
| `0B09` | sub_B09 | — | 450 | 167 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0CCB` | sub_CCB | REFRESHCOMBATMAP | 1321 | 579 | 3 | 8 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `11F4` | sub_11F4 | DOCOMBATSHAPE | 247 | 101 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `12EB` | sub_12EB | FINDX | 40 | 17 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1313` | sub_1313 | FINDY | 40 | 17 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `133B` | sub_133B | FINDSIZE | 40 | 17 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1363` | sub_1363 | FINDID | 88 | 32 | 9 | 1 | ✓ | 待解讀 | — | — | — |
| `13BB` | sub_13BB | FINDOBJECTS | 363 | 146 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1526` | sub_1526 | SUBTRACTDUDE | 749 | 296 | 0 | 10 | ✓ | 待解讀 | — | — | — |
| `1813` | sub_1813 | ADDDUDE | 506 | 197 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `1A0D` | sub_1A0D | SHOWFIG | 108 | 41 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `1A79` | sub_1A79 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
