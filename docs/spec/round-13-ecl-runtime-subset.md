# 第十三輪：ECL bounded runtime subset

狀態：`READY`（限明確列出的命令 subset）

`internal/ecl.RunSubset` 是第一個可執行的 ECL 小型 runtime，具備步數上限與明確錯誤停機：

- `EXIT`、`GOTO`、`GOSUB`、`RETURN` 與 call stack。
- `COMPARE`、六種 `IF`；不成立時依原版行為跳過下一個完整 command。
- `SAVE` 的 scalar／memory subset。
- `PRINT`／`PRINTCLEAR` 的 `0x80` compressed-string operand。
- `PICTURE` 等尚未還原副作用的命令仍作 bounded no-op；`LOAD FILES` 已改為消耗三個 operand 並輸出可觀測 map-load selector signal，但仍未宣稱完成所有 DOS file／picture／wallset 副作用。

其他尚未收斂的 opcode 仍會回傳精確 payload offset；`ON GOTO/GOSUB` 已在後續輪次加入
bounded branch semantics。`ADD NPC (0x36)` 的舊單-ID boundary 已由第 277 輪修正為
ID＋morale，並接入 NPC table／party side effect。

這不是完整遊戲 VM：尚未包含完整 party／地圖／戰鬥／選單輸入／音效狀態，亦不應把 bounded command signal 視為所有原版副作用的等價實作。

## 驗收

```sh
go test ./internal/ecl
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset
```
