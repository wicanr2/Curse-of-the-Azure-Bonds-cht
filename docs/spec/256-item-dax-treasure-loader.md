# 第 256 輪：ITEM DAX treasure loader

狀態：`READY`（限 ITEM DAX 解壓、area/block namespace 與 pending loot adapter）

## 已完成

- 啟動器現在會讀取原始 `ITEM1.DAX`～`ITEM6.DAX`。
- `internal/game.ParseTreasureItemBlocks` 透過既有 DAX RLE parser 與 0x3F-byte `ItemRecord` parser 解碼各 block。
- key 使用 `(area << 8) | blockID`，保留每個 `ITEM{area}.DAX` 的 area-local block ID，不會因不同 area 使用同一 block ID 而覆蓋。
- `State.ResolveTreasureRequests` 將已載入的 deterministic item block 轉成 pending loot，金錢依 reference conversion（Copper=1、Silver=10、Electrum=100、Gold=200、Platinum=1000 copper）進入 party gold pool；Gems／Jewelry 保留在 treasure pool。
- `State.TakeTreasureItem` 要求明確角色與物品 index，才將 loot 寫入 party equipment。

## 保留邊界

`TREASURE` operand `>= 0x80` 的 random item generation、完整原版 item selection UI、item name-number localization 與所有 loot event 的實際劇情入口仍待接入；`0xFF` no-item branch 已支援。

## 回歸

game test 使用兩個 area 的相同 DAX block ID 驗證 namespace 隔離，並驗證 TREASURE 金錢、Gems、Jewelry 與 item pickup 的 state mutation。
