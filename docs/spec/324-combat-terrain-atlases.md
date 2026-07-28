# 第三百二十四輪：原版 combat terrain atlas

狀態：`READY`

## 原始資料

CoAB 發行檔包含三個獨立於 `TILES.DAX` 的戰鬥地形 atlas：

| member | decoded bytes | items | 用途證據 |
|---|---:|---:|---|
| `DUNGCOM.DAX` | 7217 | 25 | 石牆、45°斜牆、轉角 |
| `WILDCOM.DAX` | 9809 | 34 | 樹、草、石、水岸 |
| `RANDCOM.DAX` | 1745 | 6 | 桌椅與特殊障礙 |

三者都是單一 DAX block。decoded payload 是 17-byte SSI picture header，byte 8
為 item count；剩餘長度恰為 `count × 24 × 24 ÷ 2`，每 byte 高低 nibble
各是一個 EGA palette index。palette 0 在 terrain overlay 中透明。

## Renderer contract

- `gfx.ParseCombatTiles` 嚴格驗證 header、count 與完整 payload 長度。
- 每張 tile 固定 24×24；640×480 renderer 只以 nearest-neighbour 2× 畫成 48×48。
- 地城戰場使用 `mapdata.GenerateDungeon` 已產生的 background entries，
  再由 `BackgroundTile.TileIndex` 查 `DUNGCOM`，不可依序鋪 atlas。
- 目前 7×7 viewport 取既有 50×25 buffer 的 `(18..24,7..13)`；
  這是 SetupDungeonFloor 寫入區的可見 slice。
- `Area.InDungeon` 或非世界章節的 combat fallback 使用 DUNGCOM；完整
  encounter terrain-mode selector、WILDCOM generation 與 RANDCOM decoration
  仍需追 engine routine，不把 heuristic 宣稱為最終原版規則。

## 驗收

- 原始三檔 regression 固定為 25／34／6 張 24×24 tile。
- generator 產生 `dungcom-tiles.png`、`wildcom-tiles.png`、
  `randcom-tiles.png`。
- `gold-box-layout-combat.png` 顯示由 GEO dungeon buffer lookup 得到的
  DUNGCOM 石牆，而非棋盤格或手工 mock。

