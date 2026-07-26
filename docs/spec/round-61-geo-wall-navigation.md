# 第六十一輪：GEO wall navigation contract

狀態：`READY`（限 raw GEO wall traversal）

## 已確認行為

`internal/geo.Grid.CanMove` 依 reference 的 direction order `0=north／2=east／4=south／6=west` 檢查：

- 目前 cell 對應方向 wall type 必須為 0；
- 相鄰 cell 的 opposite wall type 也必須為 0；
- 16×16 grid 外的移動必須停止。

Ebiten `GEO2` geometry viewport 的黃色游標使用同一個 contract；方向鍵只能沿已開放的 GEO wall 邊移動。

## 驗證

- current cell wall、neighbor opposite wall、open edge 與 grid boundary 均有 synthetic regression；
- viewport source 使用 `Grid.CanMove`，不是另寫一套 UI-only collision；
- `go test -vet=off ./...` 與 `go build ./cmd/azure-bonds-game` 通過。

## 邊界與未完成項目

這只是 GEO wall traversal，不是完整 AD&D movement：background tile movement cost、party placement、encounter trigger、wraparound、camera、floor construction 與場所事件仍待完成。
