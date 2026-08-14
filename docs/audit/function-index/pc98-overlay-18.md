# pc98-overlay-18 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 38 | 18 | 2 | 0 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>retf，無參數：02A8h:1392h(9, 0Fh)；08EEh:0379h(1)；02A8h:1392h(9, 9) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>audit/overlay-init-graph.md<br>spec/612-ecl-main-loop.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/754-small-predicates-and-wrappers.md |
| `0026` | sub_26 | — | 950 | 343 | 1 | 2 | ✓ | 待解讀 | — | — | spec/585-ecl-goto-and-display-mode-pair.md |
| `03DC` | sub_3DC | — | 678 | 235 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0682` | sub_682 | — | 198 | 76 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0748` | sub_748 | — | 419 | 154 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `08EB` | sub_8EB | — | 384 | 150 | 1 | 4 | ✓ | 已解讀 | exact | 1<br>同 DOS overlay-18:009DBh（module_align 對齊，助憶碼序列完全相同）。只差兩條：DS:7BA4h↔4ACEh 與 GetMem 呼叫目標 | — |
| `0A6F` | sub_A6F | — | 281 | 114 | 2 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0B9B` | sub_B9B | — | 376 | 163 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `1213` | sub_1213 | FINAL | 870 | 437 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-18.md<br>audit/function-strings.md |
| `167E` | sub_167E | — | 216 | 105 | 3 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1213h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1756` | sub_1756 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md<br>spec/612-ecl-main-loop.md |
| `175D` | sub_175D | — | 67 | 30 | 3 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add bh, 8`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 67 bytes，已逐條讀完） | — |
| `17AA` | sub_17AA | — | 73 | 35 | 3 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>讀一個像素(retn 4，參數 x, y)：位移 := y*80 + x div 8、遮罩 := 80h shr (x and 7)，依序測 A800h/B000h/B800h 三個平面組成 0..7 的值。同樣不含第四平面 | spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `17F3` | sub_17F3 | — | 170 | 80 | 1 | 4 |  | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>載入 3 平面圖檔並置中貼上(近呼叫 retn)：out 0A6h,1 後開檔(AH=3Dh)，讀 4 bytes 檔頭(寬像素、高列數)到 0C29h:7BCCh；寬位元組 := (寬+7) div 8，水平置中 (80−寬位元組) div 2、垂直置中 ((400−高) div 2)*80；逐列貼 A800h/0B000h/0B800h 三個平面(每列 80 bytes)。成功回 0、失敗回 0FFFFh。⚠ 只有三平面(8 色)，沒有 E000h；0C29h 是寫死的段值 | spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `18AF` | sub_18AF | — | 17 | 11 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為資料段、CX 取自 ds:7BCAh 的長度,呼叫 INT 21h AH=3Fh 讀檔 | — |
| `18C0` | sub_18C0 | — | 39 | 23 | 1 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 189Dh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18E7` | sub_18E7 | — | 16 | 10 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為段位址,從 offset 0 起 rep stosw 寫入 3E80h 個 0 ⇒ 清空 32000 bytes 的顯示平面 | audit/function-index/pc98-overlay-18.md |
| `18F7` | sub_18F7 | — | 17 | 10 | 1 | 0 |  | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>以 AX 為段位址,從 offset 0 起 rep stosw 寫入 3E80h 個 0FFFFh ⇒ 把 32000 bytes 的顯示平面填滿(18E7h 的填 0 版本) | — |
| `1908` | sub_1908 | — | 22 | 9 | 1 | 2 |  | 已解讀 | exact | docs/spec/575-random-core-and-pc98-vram.md<br>以 si 從 0 起、每次加 2B67h 並 and 7FFFh,重複 8000h 次呼叫 sub_191E ⇒ 對 32KB 平面做打散順序的逐格處理 | spec/575-random-core-and-pc98-vram.md |
| `191E` | sub_191E | — | 51 | 23 | 1 | 0 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 189Dh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/embedded-strings.md |
| `1951` | sub_1951 | — | 40 | 22 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>近呼叫 retn 4：把 Pascal 短字串就地改成 NUL 結尾(rep movsb 把內容往前搬一格蓋掉長度 byte，再補 0)，之後 bx:=段、cx:=位移、dl:=另一參數，呼叫 sub_17F3h。⚠ 原字串在呼叫後已被覆寫。同一段匯出後面還接兩支獨立小程序 1979h/1985h(PC-98 INT 18h AH=42h 畫面模式控制，IDA 的 ROM BASIC 註解不適用) | spec/754-small-predicates-and-wrappers.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `1979` | sub_1979 | — | 12 | 8 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>PC-98 螢幕 BIOS:INT 18h AH=42h、CX=0C000h(CRT 顯示模式設定)。⚠ IDA 的 TRANSFER TO ROM BASIC 註解是 IBM PC 語意,不適用 PC-98 | audit/function-index/pc98-overlay-18.md<br>spec/754-small-predicates-and-wrappers.md |
| `1985` | sub_1985 | — | 16 | 10 | 1 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>PC-98 螢幕 BIOS:INT 18h AH=42h、CX=8000h(另一種 CRT 模式),再 out 68h, 8 | audit/function-index/pc98-overlay-18.md<br>spec/754-small-predicates-and-wrappers.md |
