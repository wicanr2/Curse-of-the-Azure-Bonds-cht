# pc98-overlay-18 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 38 | 18 | 2 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>spec/612-ecl-main-loop.md |
| `0026` | sub_26 | — | 950 | 343 | 1 | 2 | ✓ | 待解讀 | — | — | spec/585-ecl-goto-and-display-mode-pair.md |
| `03DC` | sub_3DC | — | 678 | 235 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0682` | sub_682 | — | 198 | 76 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0748` | sub_748 | — | 419 | 154 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `08EB` | sub_8EB | — | 384 | 150 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0A6F` | sub_A6F | — | 281 | 114 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0B9B` | sub_B9B | — | 376 | 163 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1213` | sub_1213 | FINAL | 870 | 437 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `167E` | sub_167E | — | 216 | 105 | 3 | 4 |  | 待解讀 | — | — | — |
| `1756` | sub_1756 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>spec/612-ecl-main-loop.md |
| `175D` | sub_175D | — | 67 | 30 | 3 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add bh, 8`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 67 bytes，已逐條讀完） | — |
| `17AA` | sub_17AA | — | 73 | 35 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `17F3` | sub_17F3 | — | 170 | 80 | 1 | 4 |  | 待解讀 | — | — | — |
| `18AF` | sub_18AF | — | 17 | 11 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為資料段、CX 取自 ds:7BCAh 的長度,呼叫 INT 21h AH=3Fh 讀檔 | — |
| `18C0` | sub_18C0 | — | 39 | 23 | 1 | 3 |  | 待解讀 | — | — | — |
| `18E7` | sub_18E7 | — | 16 | 10 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為段位址,從 offset 0 起 rep stosw 寫入 3E80h 個 0 ⇒ 清空 32000 bytes 的顯示平面 | audit/function-index/pc98-overlay-18.md |
| `18F7` | sub_18F7 | — | 17 | 10 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為段位址,從 offset 0 起 rep stosw 寫入 3E80h 個 0FFFFh ⇒ 把 32000 bytes 的顯示平面填滿(18E7h 的填 0 版本) | — |
| `1908` | sub_1908 | — | 22 | 9 | 1 | 2 |  | 已解讀 | exact | docs/spec/575-random-core-and-pc98-vram.md<br>以 si 從 0 起、每次加 2B67h 並 and 7FFFh,重複 8000h 次呼叫 sub_191E ⇒ 對 32KB 平面做打散順序的逐格處理 | spec/575-random-core-and-pc98-vram.md |
| `191E` | sub_191E | — | 51 | 23 | 1 | 0 |  | 待解讀 | — | — | audit/embedded-strings.md |
| `1951` | sub_1951 | — | 40 | 22 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1979` | sub_1979 | — | 12 | 8 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>PC-98 螢幕 BIOS:INT 18h AH=42h、CX=0C000h(CRT 顯示模式設定)。⚠ IDA 的 TRANSFER TO ROM BASIC 註解是 IBM PC 語意,不適用 PC-98 | — |
| `1985` | sub_1985 | — | 16 | 10 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>PC-98 螢幕 BIOS:INT 18h AH=42h、CX=8000h(另一種 CRT 模式),再 out 68h, 8 | — |
