# 第六十四輪：GEO dungeon floor composition

狀態：`READY`（限 `SetupDungeonFloor` 的四段 tile composition 與 GEO2 preview）

## 已確認

- 參考 `ovr011.SetupDungeonFloor` 會以 map center 周圍 `row=-2..2`、`column=-6..6` 建立 13×5 window。
- `set_background_tile` 將 local tile id 以 `+1` 寫入 50×25 `Struct_1D1BC` background-entry buffer。
- `build_background_tiles_1–4` 依四方向 `get_dir_flags` 選擇牆角、門、地面與邊界 tile。
- `get_dir_flags` 綜合目前 GEO wall type、door/detail field，以及相鄰 cell 的 opposite direction；`internal/mapdata.GenerateDungeon` 已實作這個資料鏈。
- 遊戲 `D` 預覽與 `docs/screenshots/dungeon-floor.png` 使用同一個 `GenerateDungeon` 與 TILES RGBA adapter。

## 尚未完成

本輪已補上 `sub_370D3` 的 seeded table／chair decoration：GEO `terrain & 0x40`、四方向 flags、`tile_index == 0x16` 與 1d10 機率均有 regression。dungeon area／map selection、camera、完整 combat placement／encounter 與音效仍待接入。
