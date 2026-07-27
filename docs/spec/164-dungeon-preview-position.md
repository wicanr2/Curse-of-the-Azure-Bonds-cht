# 第一百六十四輪：dungeon preview position／camera slice

狀態：`READY`（限 GEO-safe position 與 preview 重建）

## 實作結果

- dungeon preview 保存目前 `(dungeonX,dungeonY)`，初始為 `(8,8)`。
- 方向鍵使用既有 `geo.Grid.CanMove` 的雙側 wall contract；被牆或邊界阻擋時不改變位置。
- 成功移動後重新生成 `mapdata.DungeonFloor`，並重新執行 Far／Mid／Near `WallLayoutCall` → 8×8D stamps。
- 畫面顯示目前 GEO map position，讓 wall／floor preview 可重現地隨位置變化。

## 邊界

這是 renderer preview 的 position slice，不是完整原版 party camera：尚未接 Area／save 的真實座標、方向轉向、ECL context wrap、遭遇觸發、movement cost 或 scroll animation。
