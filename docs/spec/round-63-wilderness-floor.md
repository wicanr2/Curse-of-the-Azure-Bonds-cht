# 第六十三輪：wilderness floor construction 與荒野 tile composition

狀態：`READY`（限 reference wilderness floor generation 與目前 Shadowdale map slice）

## 已確認規則

參考 `ovr011.SetupWildernessFloor`／`SetupWildernessFloor01–03` 與 `Struct_1D1BC` 後，荒野地面是 50×25 的 background-table entry map，不是 GEO 的 terrain byte：

- 初值是 entry `23`。
- 第一段依 city flags、1–100／1–4 骰點建立縱向地形帶與 entry `0x3B–0x41`。
- 第二段依相鄰格的實際 `tile_index == 22` 與 city flags 產生 entry `0x1F–0x2A`。
- 第三段依 city flags 的 group score 產生 entry `0x2C–0x3D` 的地形群組。
- background entry 再經 `BackgroundTiles[entry].TileIndex` 才得到 `TILES.DAX` 的 pixel tile index。

`internal/mapdata.GenerateWilderness` 以明確 seed 重現相同的規則順序，方便 regression 與之後接回原版 area seed。`game.State` 在進入 Shadowdale map slice 時建立 floor；`Move` 會檢查 50×25 邊界與 `MoveCost != 0xFF`。

## 尚未完成

本輪尚未完成原版 area pointer／各城市 index 的完整載入、GEO／dungeon floor construction、camera／encounter trigger、完整 50×25 map 的 place transitions 與音效。
