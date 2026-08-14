# dos-overlay-06 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 37 | 11 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化(LOADSHOP):依序 far call 依賴單元的初始化 entry,最後呼叫 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/572-resident-service-functions.md |
| `0050` | sub_50 | — | 762 | 323 | 1 | 1 | ✓ | 待解讀 | — | — | audit/duplicate-strings.md<br>audit/embedded-strings.md<br>audit/function-strings.md |
| `0355` | sub_355 | — | 269 | 96 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/777-item-append-and-area-save-sweep.md<br>把物品複製進背包尾端(retf 8)：<00FDh:004Dh>(物品, DS:6506h) 為真就顯示 CS:034Ah 'OverLoaded' 並把輸出設 1；否則 GetMem 63 bytes、搬進去、next 清 NIL，接在 +14Dh 鏈的**尾端**(空鏈與非空鏈各寫一份程式碼)，最後叫 014Dh:0043h。⚠ 'OverLoaded' 與 overlay-19:2131h 的 'Overloaded' 大小寫不同，是兩條獨立常數 | audit/function-index/pc98-overlay-06.md<br>spec/777-item-append-and-area-save-sweep.md |
| `0474` | sub_474 | — | 440 | 166 | 1 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `06F6` | sub_6F6 | — | 631 | 270 | 0 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `096D` | sub_96D | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
