# 第五十七輪：indexed picture 與 WALLDEF 資料層

狀態：`READY`（限 DAX indexed picture／wall definition parser）

## 已確認行為

原始 DAX picture block 的前 17 bytes 是 picture header：little-endian height／width units、位置、item count 與 8 bytes metadata；後續每 byte 儲存兩個 4-bit palette index，high nibble 先出現。`internal/gfx.ParsePicture` 會展開成每個 item 的 indexed pixels；masked picture 的 mask color 轉成 reference engine 使用的 palette index 16。

原始尺寸回歸：

- `TILES.DAX` blocks 是 24×24 indexed tile sets（height 24、width units 3）；
- `8X8D2–6.DAX` blocks 是 8×8 symbol sets；
- `WALLDEF2–6.DAX` 每個 record 是 5×156 bytes，部分 DAX block 會串接兩個 record；`ParseWallDefs` 會依 record 邊界拆開。

## 驗證

- packed nibble expansion 與 transparent mask synthetic regression 通過；
- 原始 TILES、8X8D、WALLDEF 全部 block 可成功 parse；
- `go test -vet=off ./...` 通過。

## 邊界與未完成項目

本輪只完成 indexed data layer，不宣稱 EGA palette、tile／symbol index mapping、wall definition offset／碰撞、TILES 與 GEO 的畫面組合或 Ebiten render 已完成。
