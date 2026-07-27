# 第一百七十輪：WallDoorFlags detail

狀態：`READY`（限 GEO door/detail flag adapter 與 preview evidence）

## 反組譯證據

reference `ovr031.WallDoorFlagsGet(mapDir,mapY,mapX)`：

- 接受 cardinal `mapDir` `0/2/4/6`，以 16×16 dungeon wrap 讀取 MapInfo。
- 初值為 `1`；若目前方向沒有 wall type，直接回傳 `1`。
- 若有 wall，回傳該方向的 `x3_dir_*` detail 值（GEO packed 2-bit field，`0..3`）。
- `ovr015` 的 `bash_door`／`locked_door` 會把回傳 `3` 作為其中一個 door action condition；完整 skill／strength／random transaction 另在同一 routine。

## 實作

- `geo.Grid.WallDoorFlagsWrapped` 精確保留上述 default、wall/detail 分支與 invalid-direction rejection。
- dungeon preview 顯示目前 facing 的 `WallDoorFlags`，並標示來源為 GEO `x3 detail`；只有目前 `mapWallType != 0` 才顯示該值，避免把 no-wall default `1` 說成 door。

## 明確 boundary

本輪沒有實作開門、解鎖、撬門、撞門、角色 skill／strength、dice、door state mutation 或 door symbol `0x104+bitmask` overlay；這些需要完整 Area／party／rules side effects 證據。

## 驗證

GEO regression 覆蓋 no-wall default `1`、walled detail `3` 與 diagonal rejection；完整 Docker gate 驗證 preview build。
