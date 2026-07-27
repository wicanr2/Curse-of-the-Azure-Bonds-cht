# 第 255 輪：ECL TREASURE signal

狀態：`READY`（限 raw treasure request 與 State exactly-once queue）

## 已確認

- 公開 CoAB reference 的 command table 將 `0x27` 定義為 `TREASURE`，固定 8 個 operand。
- `ovr003.CMD_Treasure` 將前 7 個 operand 依序寫入 pooled money：Copper、Silver、Electrum、Gold、Platinum、Gems、Jewelry。
- 第 8 個 operand 是物品來源 block：`< 0x80` 讀取目前區域的 `ITEM{game_area}.DAX` block；`>= 0x80` 進入 random item branch；`0xFF` 表示沒有物品。
- 本輪 `internal/ecl` 會保留七種數量與 item block，不在 VM 內假造 `ITEM*.DAX` 載入、隨機抽物或 inventory mutation。

## 實作邊界

`RunResult.TreasureRequests` 是資料 signal；`BlockSession` 會跨 `NEWECL` 聚合；`game.State` 以 `ConsumeTreasureRequests()` 提供 exactly-once adapter 入口。後續可由 active area loader 將 request 接到 `ITEM1.DAX`～`ITEM6.DAX` 與 party inventory。

## 回歸

synthetic `TREASURE` block 驗證 8 個 operand 會被完整消耗，且七種 coin count 與 item block 均原樣保留，之後仍可繼續到 `EXIT`。
