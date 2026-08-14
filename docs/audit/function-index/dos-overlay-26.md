# dos-overlay-26 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 179h:57h 後交給 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-24.md |
| `0011` | sub_11 | — | 99 | 38 | 5 | 1 | ✓ | 已解讀 | exact | docs/spec/812-coin-values-and-class-name-table.md<br>取鏈上第 N 個節點：retf 6，宣告順序 (索引, 串列頭)，回傳 dx:ax。n := 0；while (p <> NIL) and (n <> 索引) do p := p^[2Ah]、inc(n)；n = 索引 才回 p，否則回 NIL。串列比索引短時也回 NIL，呼叫端分不出兩種情形。spec 812 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-24.md |
| `0074` | sub_74 | — | 65 | 25 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>數鏈長(retf 4)：走遠指標鏈計數，next 在 +2Ah(＝物品節點的 far next)，回傳 word | audit/embedded-strings.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/dos-overlay-25.md<br>audit/function-index/pc98-overlay-04.md<br>audit/function-index/pc98-overlay-25.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `00D5` | sub_D5 | — | 240 | 92 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `01E5` | sub_1E5 | — | 431 | 180 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `03D4` | sub_3D4 | — | 1041 | 365 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `07E5` | sub_7E5 | — | 24 | 14 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call far ptr 542h:311h`（body 共 24 bytes，已逐條讀完） | spec/750-combat-setup.md |
| `07FD` | sub_7FD | — | 141 | 52 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `088A` | sub_88A | — | 314 | 121 | 2 | 3 | ✓ | 待解讀 | — | — | spec/848-scroll-list-widget.md |
| `09C4` | sub_9C4 | — | 76 | 31 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>第一個非空白的位置(retf 6)：同樣先複製進本地緩衝，再從第 1 個字元往後跳過 20h，回傳位置。與 0A10h 合起來是 trim 的左右兩半 | spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0A10` | sub_A10 | — | 70 | 29 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>裁掉尾端空白後的長度(retf 6)：先用 0A54h:064Eh 把第一個宣告的遠字串參數複製進 40 字元本地緩衝(上限 28h)，再從尾端往回跳過 20h，回傳長度(下限 1)。第二個參數整支沒讀；判準是 ASCII 空白，全形空白不算 | audit/function-index/dos-overlay-26.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `0A56` | sub_A56 | — | 162 | 71 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0AF8` | sub_AF8 | — | 257 | 111 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0BF9` | sub_BF9 | — | 224 | 77 | 2 | 2 | ✓ | 已解讀 | exact | 848<br>捲動清單：沿清單跳過整頁，節點 +29h（PC-98 +51h，即 spec 843 顯示行節點的屬性）為 0 就停。每次都從鏈頭重走到游標（0011h 取第 n 個節點），O(視窗高 × 游標位置) | audit/function-index/dos-overlay-26.md<br>audit/function-index/pc98-overlay-26.md<br>spec/848-scroll-list-widget.md |
| `0CD9` | sub_CD9 | — | 146 | 49 | 1 | 3 | ✓ | 已解讀 | exact | 848<br>捲動清單：翻頁。先記住游標在視窗裡的相對位置，頂端 DS:7292h 加減一整頁後夾限（上界 總筆數−視窗高、下界 0），再把游標放回同一個相對位置。⚠ 總筆數 < 視窗高 時上界為負，頂端會被夾成負值 | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-26.md<br>spec/848-scroll-list-widget.md |
| `0D6B` | sub_D6B | — | 124 | 44 | 1 | 2 | ✓ | 已解讀 | exact | 848<br>捲動清單：游標移一格。走出視窗就繞回另一端（不是捲動視窗），再叫 00BF9h 跳過屬性為 0 的行 | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-index/pc98-overlay-26.md<br>spec/848-scroll-list-widget.md |
| `0DF9` | sub_DF9 | — | 778 | 312 | 0 | 9 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `112A` | sub_112A | — | 111 | 52 | 0 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `1199` | sub_1199 | — | 168 | 59 | 0 | 1 | ✓ | 已解讀 | exact | 843<br>顯示行清單的建構子(retf 6，(鏈頭遠指標, 筆數))：GetMem 第一個節點 2Eh bytes、清 next/屬性/文字，之後 for i := 2 to 筆數 逐一 GetMem 掛在 +2Ah 並同樣清空。筆數 < 2 時只建一個。與 spec 843 的『加一行』是同一種節點 | audit/function-index/pc98-overlay-26.md |
| `1241` | sub_1241 | — | 76 | 30 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>釋放整條選單節點鏈(retf 4，參數是指向鏈頭指標的遠指標)：next 在 +2Ah、每個節點 FreeMem 2Eh(46 bytes)。⚠ 離開時沒有把鏈頭寫回 NIL，呼叫端不自己清就是懸空指標 | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-26.md<br>spec/733-cast-driver.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `128D` | sub_128D | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
