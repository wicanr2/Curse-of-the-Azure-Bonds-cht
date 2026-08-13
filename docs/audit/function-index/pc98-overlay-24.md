# pc98-overlay-24 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADGENERIC | 37 | 11 | 0 | 5 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>spec/507-pc98-general-target-object-order-projection.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/516-fire-knife-external-map-handoff-audit.md<br>spec/525-pc98-tempsearch-display-state.md |
| `0025` | sub_25 | — | 504 | 179 | 1 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/525-pc98-tempsearch-display-state.md |
| `021D` | sub_21D | — | 138 | 45 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `02A7` | sub_2A7 | — | 286 | 104 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `03C5` | sub_3C5 | — | 141 | 52 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `046C` | sub_46C | PRINTITEMNAME | 1087 | 417 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0917` | sub_917 | SHOWALL | 549 | 229 | 3 | 5 | ✓ | 待解讀 | — | — | — |
| `0B41` | sub_B41 | PRINTAC | 146 | 61 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0BD5` | sub_BD5 | PRINTHP | 99 | 38 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0C53` | sub_C53 | SHOWINFO | 431 | 196 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `0E02` | sub_E02 | HELPLESS | 69 | 28 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0E47` | sub_E47 | DOCOM | 1182 | 390 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `12E5` | sub_12E5 | BYTETOSTR | 60 | 28 | 4 | 0 | ✓ | 待解讀 | — | — | — |
| `1321` | sub_1321 | WORDTOSTR | 58 | 27 | 1 | 0 | ✓ | 待解讀 | — | — | — |
| `135B` | sub_135B | LONGTOSTR | 57 | 25 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1394` | sub_1394 | DEXDEFBONUS | 130 | 54 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1416` | sub_1416 | DEXRABONUS | 130 | 54 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1498` | sub_1498 | — | 152 | 62 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1530` | sub_1530 | STRHITBONUS | 163 | 67 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `15D3` | sub_15D3 | STRDAMBONUS | 79 | 31 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `1622` | sub_1622 | — | 55 | 23 | 3 | 1 |  | 待解讀 | — | — | — |
| `1659` | sub_1659 | STRWGTBONUS | 22 | 9 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `166F` | sub_166F | — | 40 | 16 | 2 | 2 |  | 待解讀 | — | — | — |
| `1697` | sub_1697 | — | 162 | 64 | 2 | 1 |  | 待解讀 | — | — | — |
| `1739` | sub_1739 | LOSEMEMSPELL | 65 | 24 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1785` | sub_1785 | — | 15 | 4 | 1 | 0 |  | 待解讀 | — | — | — |
| `1794` | sub_1794 | — | 12 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `17D1` | sub_17D1 | LOSEITEM | 154 | 58 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1865` | sub_1865 | — | 52 | 19 | 4 | 1 |  | 待解讀 | — | — | — |
| `186A` | sub_186A | — | 31 | 14 | 2 | 1 |  | 待解讀 | — | — | — |
| `18B8` | sub_18B8 | GAINITEM | 52 | 22 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `18EC` | sub_18EC | — | 96 | 32 | 2 | 1 |  | 待解讀 | — | — | — |
| `194C` | sub_194C | HORZMSG | 92 | 46 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `19A8` | sub_19A8 | CHARMSG | 34 | 14 | 4 | 1 | ✓ | 待解讀 | — | — | — |
| `19CA` | sub_19CA | — | 198 | 94 | 2 | 4 |  | 待解讀 | — | — | — |
| `1A90` | sub_1A90 | CLEARMSGWINDOW | 50 | 26 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1AC2` | sub_1AC2 | PRINTCHARNAME | 109 | 41 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `1B2F` | sub_1B2F | COPYSPRITE | 287 | 113 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1C48` | sub_1C48 | — | 40 | 15 | 2 | 0 |  | 待解讀 | — | — | — |
| `1C76` | sub_1C76 | CREATERADIALSPRITE | 77 | 45 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1CC3` | sub_1CC3 | MOVESPRITE | 978 | 329 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `2095` | sub_2095 | — | 437 | 142 | 7 | 5 |  | 待解讀 | — | — | — |
| `224A` | sub_224A | — | 448 | 146 | 2 | 8 |  | 待解讀 | — | — | — |
| `240A` | sub_240A | TWINKLE | 451 | 162 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `25CD` | sub_25CD | SPELLON | 107 | 37 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `2658` | sub_2658 | SAVEDAMAGE | 254 | 80 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `2776` | sub_2776 | HEALMSG | 109 | 47 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `27E3` | sub_27E3 | ENEMYOF | 36 | 13 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `2807` | sub_2807 | FIGGOODBAD | 84 | 29 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `285B` | sub_285B | FINDFOES | 292 | 110 | 0 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>spec/508-pc98-general-target-scan-producer.md |
| `297F` | sub_297F | FIGRANGE | 220 | 81 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2A5B` | sub_2A5B | QUIT | 70 | 20 | 1 | 0 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md |
| `2AAE` | sub_2AAE | GUARD | 60 | 24 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2AEA` | sub_2AEA | FIGCASTERLEVEL | 438 | 149 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2CA4` | sub_2CA4 | SETUPSCREEN | 456 | 192 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `2E8C` | sub_2E8C | SHOWLOCATION | 603 | 248 | 1 | 4 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md<br>project-status.md<br>spec/516-fire-knife-external-map-handoff-audit.md<br>spec/525-pc98-tempsearch-display-state.md |
| `30E7` | sub_30E7 | DOCOMBATSCREEN | 54 | 22 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `314B` | sub_314B | PICKAPERSON | 374 | 141 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `32B6` | sub_32B6 | — | 123 | 41 | 2 | 2 |  | 待解讀 | — | — | — |
| `3334` | sub_3334 | USINGMISSLEWEAPON | 67 | 23 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3377` | sub_3377 | USINGHURLEDWEAPON | 69 | 27 | 0 | 2 | ✓ | 待解讀 | — | — | context/50-log-2026-08-09-13.md |
| `33BC` | sub_33BC | VALIDMISSLE | 194 | 69 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `347E` | sub_347E | LEVELLIMIT | 327 | 110 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `35D8` | sub_35D8 | CHECKDYING | 168 | 57 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `3680` | sub_3680 | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
