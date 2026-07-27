# 第六十三輪：wilderness combat-floor construction

狀態：`READY`（限野外遭遇 combat background generation）

## 已確認規則

參考 `ovr011.SetupGroundTiles → SetupWildernessFloor`／`SetupWildernessFloor01–03` 與
`Struct_1D1BC` 後，這是野外遭遇的 50×25 **combat background** entry map，不是
world map，也不是 GEO 的 terrain byte：

- 初值是 entry `23`。
- 第一段依 city flags、1–100／1–4 骰點建立縱向地形帶與 entry `0x3B–0x41`。
- 第二段依相鄰格的實際 `tile_index == 22` 與 city flags 產生 entry `0x1F–0x2A`。
- 第三段依 city flags 的 group score 產生 entry `0x2C–0x3D` 的地形群組。
- background entry 再經 `BackgroundTiles[entry].TileIndex` 才得到 `TILES.DAX` 的 pixel tile index。

`internal/mapdata.GenerateWilderness` 以明確 seed 重現相同的規則順序，方便 regression
與之後接回 combat setup seed。早期 prototype 曾把它接到 Shadowdale `ModeMap`；
第 280 輪已確認這不是正式世界移動模型，production new-game 改走 16×16 GEO dungeon。

## 尚未完成

本輪尚未完成把此 50×25 combat floor 接入所有野外 encounter、完整 combat placement
與 area/city flags seed；它沒有 world-map place transition 語意。
