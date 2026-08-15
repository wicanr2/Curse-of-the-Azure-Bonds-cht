# 第六十二輪 README 與可重現截圖

狀態：`READY`（限 repository progress evidence）

## 目標

讓 GitHub 首頁能直接展示目前 prototype 的可見成果，同時不把 tile gallery 或 GEO geometry viewport 誤稱為完整地圖與完整遊戲。

## 已確認內容

- `scripts/render_previews/` 從原始 ZIP 讀取 `TILES.DAX`、`GEO2.DAX`。
- `TILES.DAX` 透過 `internal/dax`、`internal/gfx` 與 reference EGA16 palette 輸出 tile gallery。
- `GEO2.DAX` 透過 `internal/dax`、`internal/geo` 輸出 16×16 geometry，青線代表非零 wall field，黃點代表預覽游標。
- PNG 位於 `docs/screenshots/`，可由 README 的相對連結直接顯示。
- `internal/mapdata` 保存 reference `BackGroundTiles` 的 74 筆 metadata；這是 floor construction 的資料基礎，不是已完成的 tile mapping。

## 重現

```sh
./tools/go.sh run ./scripts/render_previews
```
