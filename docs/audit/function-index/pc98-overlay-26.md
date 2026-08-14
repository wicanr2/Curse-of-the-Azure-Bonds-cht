# pc98-overlay-26 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADMENUS | 17 | 7 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化:far call 176h:57h 後交給 19Ah:2Ah | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-24.md |
| `0011` | sub_11 | LOCATEITEM | 98 | 38 | 5 | 1 | ✓ | 已解讀 | exact | docs/spec/812-coin-values-and-class-name-table.md<br>取鏈上第 N 個節點：retf 6，宣告順序 (索引, 串列頭)，回傳 dx:ax。n := 0；while (p <> NIL) and (n <> 索引) do p := p^[52h]、inc(n)；n = 索引 才回 p，否則回 NIL。串列比索引短時也回 NIL，呼叫端分不出兩種情形。spec 812（差異只在 mov/les 的等價碼型） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/dos-overlay-25.md |
| `0073` | sub_73 | ITEMCOUNT | 65 | 25 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>數鏈長(retf 4)：next 在 +52h，與 DOS 的 +2Ah 不同——PC-98 的節點配置重排過，引用位移前須各自確認 | audit/function-index/dos-overlay-22.md<br>audit/function-index/dos-overlay-25.md<br>audit/function-index/pc98-overlay-25.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `00B4` | sub_B4 | — | 93 | 41 | 1 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0133` | sub_133 | MENU | 1771 | 581 | 2 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `081E` | sub_81E | CLEARMENU | 51 | 27 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>清最下面一列文字(retf)：0A65h:1B30h(A000h:0F00h, 0A0h, 0) 與 (A200h:0F00h, 0A0h, 0)。參數形狀與效果可認定 0A65h:1B30h 是 FillChar(DOS 對應 0A54h:1AE0h)——注意兩平台 resident 沒有固定位移，不能用位址換算 | spec/755-per-character-slots-sound-driver-and-text-io.md |
| `0858` | sub_858 | INIMNUBUF | 64 | 26 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/755-per-character-slots-sound-driver-and-text-io.md<br>十筆欄位初始化(retf)：for i:=1 to 10，把 CS:0851h 的 6 bytes 複製到 DS:0A338h+i*7，再把 DS:0A334h+i 設成空白。兩個平行陣列(7 bytes 一筆的記錄 + byte 旗標) | spec/755-per-character-slots-sound-driver-and-text-io.md |
| `08AE` | sub_8AE | MKEMNULST | 304 | 126 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `09DE` | sub_9DE | — | 140 | 52 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0A6A` | sub_A6A | — | 319 | 122 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `0BA9` | sub_BA9 | — | 85 | 34 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>retf 4：參數是 SS 相對的記錄指標；用 +14h 起的遠指標與第二個參數叫本模組 0011h 取得遠指標，再叫 0418h:0D17h(該指標, 0, [+20h], [+1Ch] + (參數 − DS:0A32Ch), [+1Eh]) | spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `0BFE` | sub_BFE | — | 146 | 58 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0C90` | sub_C90 | — | 224 | 77 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0D70` | sub_D70 | — | 146 | 49 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0E02` | sub_E02 | — | 124 | 44 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-24.md |
| `0ED3` | sub_ED3 | VERTICALLIST | 897 | 360 | 0 | 10 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `129D` | sub_129D | YESNO | 130 | 61 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `131F` | sub_131F | BUILDMENULIST | 168 | 59 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `13C7` | sub_13C7 | DISPOSEMENULIST | 88 | 34 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>釋放整條鏈(retf 4)：next 在 +52h、每節點 FreeMem 56h(86 bytes)，離開前有把鏈頭寫回 NIL。⚠ DOS 的對應功能 overlay-26:1241h 是 2Eh(46 bytes)、next +2Ah，而且沒有清鏈頭 — 兩平台在節點大小與收尾行為都不同 | spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md |
| `141F` | sub_141F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/duplicate-strings.md<br>audit/embedded-strings.md |
