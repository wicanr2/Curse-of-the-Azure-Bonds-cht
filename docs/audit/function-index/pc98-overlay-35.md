# pc98-overlay-35 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSQRPAK8 | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>spec/1074-global-init-pc98.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/882-load-8x8-tileset.md |
| `004C` | sub_4C | LOAD8X8SET | 699 | 285 | 0 | 1 | ✓ | 已解讀 | exact | 882<br>同 DOS overlay-35:00047h 的目的，但★PC-98 自己處理圖形記憶體：先 FillChar(圖^[17h], 0D20h=3360, 0) 與 FillChar(遠指標(圖^[13h])^, 460h=1120, 0)，用 723h:824h 讀檔後逐欄抄表頭(⚠前兩欄對調：圖^[0]←檔案^[2]、圖^[2]←檔案^[0])，+8 下限 1，+11h := 寬×高×3，再叫 2A8h:3BDh 並 out 0A6h,0。★3360 = 3 × 1120 → +17h 是三個色彩平面、+13h 是一個遮罩平面，這解釋了 spec 880 的平台差異 | audit/function-strings.md<br>spec/882-load-8x8-tileset.md |
| `032A` | sub_32A | PUT8X8SYMBOL | 336 | 127 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/984-drop-character-and-pc98-put8x8symbol.md<br>Put8x8Symbol 的 PC-98 版；DS:488Ah 五個分段起始值 1/46/116/186/256 直接讀到；> 7FFFh 同樣不指派「組」 | audit/function-strings.md<br>spec/984-drop-character-and-pc98-put8x8symbol.md |
| `047A` | sub_47A | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
