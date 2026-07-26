# 第一百六十二輪：3D wall viewport layout slice

狀態：`READY`（限 wall stamp layout 與 preview 垂直切片）

## 已確認規則

reference `draw_3D_8x8_titles` 使用十組固定 viewport metadata：

```text
idxOffset  = [0, 2, 6, 10, 22, 38, 54, 110, 132, 154]
colCount   = [1, 1, 1, 3, 2, 2, 7, 2, 2, 1]
rowCount   = [2, 4, 4, 4, 8, 8, 8, 11, 11, 2]
```

`wallType` 1–5、6–10、11–15 分別選取 WALLDEF slot 1、2、3；slice 是 `(wallType-1) % 5`，column 由 `idxOffset` 依 viewport row-major 消費。

## 實作結果

- `gfx.BuildWallLayout` 封裝上述 metadata，輸出 renderer-neutral `WallStamp`（邏輯 row／column、global symbol ID、local item 與 picture）。
- dungeon preview 會從目前 GEO 找到第一個 wall type，產生一個 front-wall stamp sample，並以原始 8×8D pixel item 顯示。
- regression 覆蓋第一組 layout 的 index 0、位置 `(4,5)` 與 local item。

## 邊界

本輪未宣稱完成 `Draw3dWorldFar/Mid/Near` 的方向遍歷、遮擋、sky／roof layer、door layer 或 camera movement；目前是可驗證的 wall layout 與素材 preview 垂直切片。
