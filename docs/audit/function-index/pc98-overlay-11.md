# pc98-overlay-11 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADINIT | 47 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 8 個呼叫，沒有其他動作：`call far ptr 0C9h:2Fh`、`call far ptr 117h:57h`、`call far ptr 232h:34h`、`call far ptr 194h:43h`、`call far ptr 176h:57h`、`call far ptr 8Bh:2Ah`（body 共 47 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md |
| `0048` | sub_48 | INITALL | 1209 | 503 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>context/50-log-2026-08-09-13.md<br>project-status.md |
| `0508` | sub_508 | INITVARS | 606 | 200 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0766` | sub_766 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
| `07CD` | sub_7CD | — | 51 | 27 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>16 色調色盤設定(retn 2)：out 6Ah,1 切 16 色；依參數選 CS:076Dh(非 0)或 CS:079Dh(0)這兩份 16×3=48 bytes 的調色盤，經 0A8h/0AAh/0ACh/0AEh 寫入(索引/綠/紅/藍，索引遞減)；再 int 18h AH=42h/40h/12h 與 out 68h,8。IDA 的 ROM BASIC 註解是 IBM PC 語意，PC-98 不適用 | audit/function-index/pc98-overlay-11.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/775-take-menu-and-palette-writer.md |
| `0800` | sub_800 | — | 40 | 26 | 1 | 1 |  | 已解讀 | exact | docs/spec/775-take-menu-and-palette-writer.md<br>寫入 16 組調色盤(近呼叫 retn，SI 指向 48 bytes 資料)：for i := 0Fh downto 0 — out 0A8h,i 後依序 out 0AAh/0ACh/0AEh(綠紅藍)，SI 往前遞增。⚠ 索引遞減而資料遞增，資料第一組對應索引 15。spec 755 的 overlay-11:07CDh 呼叫的就是這一段 | project-status.md<br>spec/775-take-menu-and-palette-writer.md |
