# 第六十八輪：Area1／Area2 map-load boundary

狀態：`READY`（限 area state 與 `CMD_LoadFiles` branch contract）

## 已確認

reference `Area1` 的已知欄位包括 `current_3DMap_block_id`（`0x18A`）、`inDungeon`（`0x1CC`）、`lastXPos`／`lastYPos`、`LastEclBlockId` 與 `current_city`；全域 `game_area` 決定 `GEO<area>.DAX` 檔案 set。

`internal/area.State` 現在保存這些已驗證的 map-selection 邊界欄位，並實作 `ApplyLoadFiles`：

- dungeon 且第三 operand 不是 `0xFF`／`0x7F`：更新 `Current3DMapBlockID`，產生 GEO map effect。
- 非 dungeon 且第一 operand 有效、last DAX block 不是 `0x50`：產生 big-picture effect。
- 其他情況不虛構 map／picture side effect。

`game.State` 已使用此 contract；ECL map request 只有在 `SetInDungeon(true)` 後才會轉給 GEO catalog，避免把 wilderness `LOAD FILES` 誤當 dungeon map。

## 尚未完成

完整 Area1／Area2 二進位 save/import、`game_area` 從 ECL／save 的自動載入、WALLDEF set reload、big-picture renderer 與完整 dungeon entry flow 仍待完成。
