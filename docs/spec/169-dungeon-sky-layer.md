# 第一百六十九輪：dungeon sky layer

狀態：`READY`（限 Area1 sky colour codec 與 dungeon preview sky background）

## 反組譯證據

- reference `Area1.cs` 的 `outdoor_sky_colour` 位於 `0x1FA`、`indoor_sky_colour` 位於 `0x1FC`，均為 word。
- reference `ovr029.RedrawView` 先讀 `mapWallRoof = get_wall_x2(...)`；`mapWallRoof > 0x7F` 選 indoor sky，否則選 outdoor sky。
- reference sky palette table 為 `{0x00,0x0F,0x04,0x0B,0x0D,0x02,0x09,0x0E}` 重複兩次；目前共用 EGA16 palette 可直接承接這些 index。

## 實作

- `area.State`／Area1 codec 讀寫兩個 sky colour word，並保留未知 bytes。
- dungeon preview 以目前 `DungeonWallRoof` 的 high bit 選 indoor/outdoor sky，將 Area sky index 映射成 EGA16 background；wall stamps 仍在 sky layer 上方繪製。
- map wall cache 仍由 GEO refresh 重算，sky layer 不把 cache 當成碰撞或地圖真相。

## 明確 boundary

本輪尚未完成原版 Area sky colour 的完整 UI／save container 寫回、sky animation、roof geometry、door overlay 或完整 `Draw3dWorldBackground` layer；目前是可驗證的 colour-selection vertical slice。

## 驗證

Area1 codec regression 覆蓋 `0x1FA/0x1FC` word round-trip；Docker gate 覆蓋完整 tests 與 Ebiten build。
