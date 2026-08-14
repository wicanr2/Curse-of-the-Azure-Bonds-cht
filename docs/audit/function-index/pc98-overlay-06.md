# pc98-overlay-06 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSHOP | 37 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 6 個呼叫，沒有其他動作：`call far ptr 10Dh:84h`、`call far ptr 0FAh:7Ah`、`call far ptr 164h:57h`、`call far ptr 14Ah:101h`、`call sub_6E0`、`call far ptr 19Ah:2Ah`（body 共 37 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md<br>spec/823-shop-price-multiplier.md |
| `0037` | sub_37 | — | 860 | 364 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-02.md<br>audit/function-strings.md |
| `03A0` | sub_3A0 | GIVEITEM | 269 | 96 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/777-item-append-and-area-save-sweep.md<br>與 DOS overlay-06:0355h（entry#2）助憶碼序列完全相同，語意同該筆：把物品複製進背包尾端(retf 8)：<00FDh:004Dh>(物品, DS:6506h) 為真就顯示 CS:034Ah 'OverLoaded' 並把輸出設 1；否則 GetMem 63 bytes、搬進去、next 清 NIL，接在 +14Dh 鏈的**尾端**(空鏈與非空鏈各寫一份程式碼)，最後叫 014Dh:0043h。⚠ 'OverLoaded' 與 overlay-19:2131h 的 'Overloaded' 大小寫不同，是兩條獨立常數 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `04BE` | sub_4BE | — | 440 | 166 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/823-shop-price-multiplier.md<br>商店買東西的主迴圈：retf。repeat：選中 := NIL、清單 := DS:0A026h、旗標 := 0；<far 019Eh:0000h>；DS:78A4h := 1；本模組 0050h(@鍵, @選中, @清單, @旗標)；鍵 <> 42h('B') 且 <> 0Dh(Enter) 就結束。否則依 word(DS:7F09h^[62h]) 算價：1→選中^[62h] div 10h、2→div 8、4→div 4、8→div 2、20h→shl 1、40h→shl 2、80h→shl 3——**10h 那一段與 else 都沒有程式碼，價是未初始化的區域變數（缺陷）**。有 := <00FAh:0057h>(DS:9594h)；價 <= 有 → 本模組 0355h(@失敗, 選中)，失敗 = 0 才 <010Dh:006Bh>(有 − 價)；否則 池 := <010Dh:0075h>(@DS:0A00Ah)，價 <= 池 → 同樣先 0355h 再 <010Dh:0070h>(低字(池 − 價))；都不夠則顯示 'お金が足りません' 並 <014Ah:007Fh>。付款模型與 spec 822 相同，但多一個回滾點：0355h 失敗就不扣錢。spec 823 | audit/function-strings.md |
| `06A4` | nullsub_1 | — | 1 | 1 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `iret`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `06E0` | sub_6E0 | — | 22 | 10 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `iret`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 22 bytes，已逐條讀完） | — |
| `0778` | sub_778 | SHOPMAINMENU | 727 | 307 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0A4F` | sub_A4F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
