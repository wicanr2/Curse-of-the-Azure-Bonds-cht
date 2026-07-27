# 第一百六十七輪：dungeon coordinate wrap

狀態：`READY`（限 dungeon GEO／wall preview 的 16×16 wrap）

## 反組譯證據

- reference `ovr031.getMap_XXX` 與 `get_wall_x2` 對 `mapX/mapY` 在有效 dungeon context 以 0..15 wrap；`ovr008.MovePositionForward` 也以 `DecrimentWrap`／`IncrementWrap` 移動。
- `ovr015` 的 dungeon input 先以 GEO wall type 做雙側碰撞，再更新位置並重繪 `Draw3dWorld`。
- 同一份原始程式保留 ECL block 0／10 的 invalid-coordinate 特例，因此 wrap 不是所有 Area／ECL context 的全域規則。

## 實作

- `geo.WrapCoordinate`、`CellWrapped`、`WallWrapped` 與 `CanMoveWrapped` 建立明確 wrapped API；原本嚴格邊界的 `Cell`／`Wall`／`CanMove` 不變。
- `gfx.TraverseWallViewWrapped` 讓 Far／Mid／Near 的 wall lookup 讀取 wrapped cell，但保留未包裝 traversal 給其他 context。
- dungeon preview 使用 wrapped movement／traversal，跨越 16×16 邊緣後仍重建 floor 與 wall stamps。

## 明確 boundary

本輪只完成 dungeon preview 的 GEO coordinate wrap；尚未完成從完整 ECL／Area loader 判斷何時啟用例外、原版 `mapWallType/mapWallRoof` 五-byte save segment、encounter distance、movement time 或完整 3D overlay。

## 驗證

新增 GEO wrapped movement、wrapped wall traversal regressions；既有 strict-boundary tests、完整 `go test ./...` 與 Ebiten build 覆蓋兩種 context。
