# dos-overlay-35 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/882-load-8x8-tileset.md |
| `0047` | sub_47 | — | 265 | 112 | 0 | 1 | ✓ | 已解讀 | exact | 882<br>載入 8×8 圖磚組(retf 4，(組, 編號))：組要先過 0A54h:8D4h 的合法表；檔名 '8x8d'＋Str(DS:5BEEh) 一位磁碟編號，叫 297h:11Bh 載入到圖庫表 DS:65B6h + 組*4；回 NIL 就組錯誤訊息 'Unable to load '＋編號＋' from 8x8D' 並收場 | audit/embedded-strings.md<br>audit/function-strings.md<br>spec/882-load-8x8-tileset.md |
| `0173` | sub_173 | — | 336 | 127 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/781-put8x8symbol-and-wall-bit-set.md<br>Put8x8Symbol(retf 0Ah；函式名由自身錯誤訊息 CS:0150h 'Bad symbol number in Put8x8Symbol.' 定住)：編號分五段(1..2Dh/2Eh..73h/74h..0B9h/0BAh..0FFh/100h..127h)選 DS:65B6h 起的遠指標，減掉 word[2680h + 組*2] 的起始編號；第二個參數非 0 走 0297h:08F8h 直接畫，否則先 Move(p^[17h + n*p^[11h]] → DS:65CAh^[17h], p^[11h]) 再叫 0297h:1110h。可證符號資源表頭 17h bytes、+11h 是每符號位元組數。⚠ 編號為負數時「組」這個 local 未指派，會用堆疊殘值當索引 | spec/781-put8x8symbol-and-wall-bit-set.md |
| `02C3` | sub_2C3 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
