# 第六十輪：GEO geometry Ebiten viewport

狀態：`READY`（限 GEO raw geometry preview）

## 已確認行為

`cmd/azure-bonds-game` 啟動時載入 `GEO2.DAX` 第一個 decoded block，使用 `internal/geo.Parse` 保存 16×16 cells。按 `G` 可開啟繁中 GEO geometry viewport：每格顯示原始幾何格線，四方向非零 wall fields 以青線繪出；按 `G` 或 `Esc` 返回。

viewport 明確把 raw geometry 與 tile／碰撞分開，避免把 `MapInfo` 的 terrain 或 wall byte 直接冒充 background tile index。

## 驗證

- `go test -vet=off ./...` 通過；
- `go build ./cmd/azure-bonds-game` 通過；
- app source 已將原始 GEO block 接入 Ebiten Draw path。

## 邊界與未完成項目

這是原始 geometry viewport，不是完整遊戲地圖：GEO area/block 選擇、background floor construction、TILES／WALLDEF index mapping、碰撞／移動成本、camera、場所事件與音效仍待完成。
