# dos-overlay-02 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 82 | 20 | 0 | 10 | ✓ | 待解讀 | — | — | audit/ecl-opcode-dispatch.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/pc98-overlay-02.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `0052` | sub_52 | — | 150 | 49 | 3 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 00h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>context/50-log-2026-08-09-13.md<br>spec/559-full-module-re-sweep.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `00E8` | sub_E8 | — | 31 | 14 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 01h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>spec/559-full-module-re-sweep.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `0107` | sub_107 | — | 23 | 11 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 02h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>spec/559-full-module-re-sweep.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `011E` | sub_11E | — | 125 | 50 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 03h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>spec/559-full-module-re-sweep.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `019B` | sub_19B | — | 169 | 64 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 04h／05h／06h／07h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0244` | sub_244 | — | 94 | 35 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 08h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `02A2` | sub_2A2 | — | 78 | 30 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 09h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `02F0` | sub_2F0 | — | 218 | 73 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Ah 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `03CA` | sub_3CA | — | 151 | 52 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Ch 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0466` | sub_466 | — | 841 | 284 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Bh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `06D5` | sub_6D5 | — | 6 | 2 | 33 | 0 |  | 待解讀 | — | — | — |
| `06DF` | sub_6DF | — | 5 | 2 | 3 | 0 |  | 待解讀 | — | — | — |
| `06E4` | sub_6E4 | — | 5 | 1 | 3 | 0 |  | 待解讀 | — | — | — |
| `06E9` | sub_6E9 | — | 8 | 2 | 2 | 0 |  | 待解讀 | — | — | — |
| `0739` | sub_739 | — | 9 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `0748` | sub_748 | — | 8 | 2 | 3 | 0 |  | 待解讀 | — | — | — |
| `0752` | sub_752 | — | 9 | 3 | 2 | 0 |  | 待解讀 | — | — | — |
| `0801` | sub_801 | — | 60 | 22 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Dh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0841` | sub_841 | — | 235 | 77 | 1 | 8 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Eh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/518-dos-start-ecl-call-address-space-audit.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `092C` | sub_92C | — | 68 | 28 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 0Fh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0972` | sub_972 | — | 120 | 51 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 10h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `09EA` | sub_9EA | — | 156 | 69 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 11h／12h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0A86` | sub_A86 | — | 80 | 30 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 13h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0AD6` | sub_AD6 | — | 101 | 40 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 14h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0B3B` | sub_B3B | — | 128 | 43 | 1 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 16h／17h／18h／19h／1Ah／1Bh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0BBB` | sub_BBB | — | 90 | 33 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 20h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0C15` | sub_C15 | — | 309 | 104 | 1 | 5 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 21h／37h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0D4A` | sub_D4A | — | 5 | 2 | 2 | 0 |  | 待解讀 | — | — | — |
| `0D4F` | sub_D4F | — | 85 | 25 | 2 | 4 |  | 待解讀 | — | — | — |
| `0DA4` | sub_DA4 | — | 27 | 11 | 1 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 2Fh／30h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0DBF` | sub_DBF | — | 84 | 32 | 4 | 3 |  | 待解讀 | — | — | — |
| `0E13` | sub_E13 | — | 94 | 35 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 2Ah 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0E71` | sub_E71 | — | 76 | 29 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 35h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0EBD` | sub_EBD | — | 82 | 29 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 15h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `0F0F` | sub_F0F | — | 13 | 4 | 4 | 1 |  | 待解讀 | — | — | — |
| `1082` | sub_1082 | — | 396 | 161 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 2Bh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `120E` | sub_120E | — | 99 | 37 | 1 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 1Ch 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `1271` | sub_1271 | — | 325 | 120 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 1Dh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `13B6` | sub_13B6 | — | 96 | 33 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `1416` | sub_1416 | — | 382 | 141 | 1 | 6 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 1Eh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `1577` | sub_1577 | — | 48 | 16 | 2 | 1 |  | 待解讀 | — | — | — |
| `15AE` | sub_15AE | — | 10 | 5 | 4 | 0 |  | 待解讀 | — | — | — |
| `15B8` | sub_15B8 | — | 10 | 6 | 2 | 2 |  | 待解讀 | — | — | — |
| `15D1` | sub_15D1 | — | 76 | 23 | 2 | 2 |  | 待解讀 | — | — | — |
| `1636` | sub_1636 | — | 30 | 11 | 1 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 22h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md<br>context/50-log-2026-08-09-13.md |
| `16B8` | sub_16B8 | — | 8 | 3 | 2 | 1 |  | 待解讀 | — | — | — |
| `16D2` | sub_16D2 | — | 165 | 66 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 23h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `177A` | sub_177A | — | 32 | 14 | 2 | 2 |  | 待解讀 | — | — | — |
| `179A` | sub_179A | — | 61 | 18 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 24h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `17B5` | sub_17B5 | — | 3 | 1 | 4 | 1 |  | 待解讀 | — | — | — |
| `17C4` | sub_17C4 | — | 3 | 1 | 2 | 1 |  | 待解讀 | — | — | — |
| `17DD` | sub_17DD | — | 5 | 1 | 4 | 0 |  | 待解讀 | — | — | — |
| `17E2` | sub_17E2 | — | 5 | 1 | 3 | 1 |  | 待解讀 | — | — | — |
| `17E7` | sub_17E7 | — | 74 | 19 | 2 | 5 |  | 待解讀 | — | — | — |
| `182E` | sub_182E | — | 14 | 3 | 3 | 0 |  | 待解讀 | — | — | — |
| `184B` | sub_184B | — | 222 | 55 | 1 | 5 |  | 待解讀 | — | — | — |
| `18B3` | sub_18B3 | — | 42 | 11 | 2 | 2 |  | 待解讀 | — | — | — |
| `1953` | sub_1953 | — | 3 | 1 | 4 | 1 |  | 待解讀 | — | — | — |
| `1956` | sub_1956 | — | 77 | 21 | 2 | 8 |  | 待解讀 | — | — | — |
| `19A3` | sub_19A3 | — | 76 | 19 | 2 | 2 |  | 待解讀 | — | — | — |
| `19EF` | sub_19EF | — | 91 | 24 | 2 | 4 |  | 待解讀 | — | — | — |
| `1A4F` | sub_1A4F | — | 76 | 24 | 2 | 2 |  | 待解讀 | — | — | — |
| `1A9B` | sub_1A9B | — | 149 | 57 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 25h／26h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `1B53` | sub_1B53 | — | 1011 | 398 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 27h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `1F46` | sub_1F46 | — | 227 | 78 | 1 | 3 |  | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 28h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2086` | sub_2086 | — | 1791 | 649 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 29h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `27A8` | sub_27A8 | — | 159 | 66 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 2Ch 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2847` | sub_2847 | — | 172 | 61 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 32h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `28F3` | sub_28F3 | — | 16 | 7 | 1 | 0 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Ah 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2942` | sub_2942 | — | 845 | 305 | 1 | 4 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 2Eh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2C8F` | sub_2C8F | — | 38 | 12 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 31h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2CB5` | sub_2CB5 | — | 53 | 22 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 34h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2CEA` | sub_2CEA | — | 43 | 14 | 1 | 1 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 33h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2D15` | sub_2D15 | — | 73 | 23 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Dh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2D5E` | sub_2D5E | — | 75 | 36 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 39h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2DA9` | sub_2DA9 | — | 109 | 38 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 36h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2E16` | sub_2E16 | — | 236 | 83 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Bh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `2F02` | sub_2F02 | — | 369 | 124 | 1 | 10 | ✓ | 已解讀 | exact | docs/spec/561-ecl-external-call-registry.md<br>ECL CALL（opcode 2Dh）handler；operand-7FFFh=selector，認得 7 個 external routine：2E10h／6803h／8000h／8001h／B200h／C018h／C01Eh | audit/ecl-external-call-registry.md<br>audit/ecl-opcode-dispatch.md<br>spec/561-ecl-external-call-registry.md |
| `3073` | sub_3073 | — | 71 | 24 | 2 | 6 | ✓ | 待解讀 | — | — | audit/ecl-external-call-registry.md |
| `30DD` | sub_30DD | — | 322 | 104 | 1 | 9 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 38h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `321F` | sub_321F | — | 50 | 20 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Ch 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `3251` | sub_3251 | — | 51 | 19 | 1 | 2 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Eh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `3284` | sub_3284 | — | 84 | 34 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 3Fh 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `32D8` | sub_32D8 | — | 159 | 56 | 1 | 3 | ✓ | 待解讀 | — | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode 40h 的 handler（binding exact，語意未解讀） | audit/ecl-opcode-dispatch.md |
| `3377` | sub_3377 | — | 682 | 290 | 1 | 53 | ✓ | 已解讀 | exact | docs/spec/560-ecl-opcode-dispatch-table.md<br>ECL opcode dispatcher；由 ds:75FFh 讀 opcode，00h..40h 中除 1Fh 外共 64 個 opcode 對應 52 個 handler | audit/ecl-opcode-dispatch.md<br>context/50-log-2026-08-09-13.md<br>spec/559-full-module-re-sweep.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `3621` | sub_3621 | — | 112 | 39 | 3 | 3 | ✓ | 待解讀 | — | — | audit/ecl-opcode-dispatch.md<br>spec/560-ecl-opcode-dispatch-table.md |
| `3691` | sub_3691 | — | 225 | 70 | 1 | 6 | ✓ | 待解讀 | — | — | — |
| `3772` | sub_3772 | — | 679 | 205 | 0 | 12 | ✓ | 待解讀 | — | — | — |
| `3A19` | sub_3A19 | — | 7 | 5 | 0 | 0 | ✓ | 待解讀 | — | — | — |
