# 第六十六輪：原始 GEO map catalog

狀態：`READY`（限 GEO2–GEO6 DAX map ID catalog 與 preview selector）

## 已確認

- 原始映像含 `GEO2.DAX` 至 `GEO6.DAX`，共 16 個 decoded `0x402` blocks。
- block ID 保留 DAX index 的原值：GEO2 `{0x01,0x03,0x04}`、GEO3 `{0x10,0x11,0x15}`、GEO4 `{0x20,0x21,0x25}`、GEO5 `{0x32,0x33,0x35}`、GEO6 `{0x40,0x42,0x43,0x45}`。
- `internal/geo.Catalog` 以 `MapRef{Set, BlockID}` 查找，沒有把 block 順序誤當成 map ID。
- 遊戲加入 `-geo-set`／`-geo-block`，選中的 GEO block 會同時驅動 G 預覽與 D dungeon floor composition。

## 尚未完成

reference `Area1.current_3DMap_block_id` 與 ECL／area state 的完整 loader 尚未接通；目前 selector 是明確 CLI 選擇，尚非完整玩家流程的自動地圖切換。
