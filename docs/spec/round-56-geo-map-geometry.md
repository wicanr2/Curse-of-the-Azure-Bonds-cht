# 第五十六輪：GEO 16×16 地圖幾何資料層

狀態：`READY`（限 GEO block geometry parser）

## 已確認行為

參考 engine 的 `GeoBlock.LoadData` 與原始 DAX decoded size 一致：每個 GEO block 為 `0x402` bytes，跳過前 2 bytes 後有四個 `0x100` plane：

1. 每格兩個 4-bit wall direction 欄位；
2. 第二組兩個 4-bit wall direction 欄位；
3. 每格一個未修改的 terrain／background byte；
4. 每格四個 2-bit direction detail 欄位。

`internal/geo.Parse` 以 16×16 `Grid`／`Cell` 暴露上述資料，保留原始 direction order `0／2／4／6`，也接受已移除 prefix 的 `0x400` payload。

## 驗證

- synthetic packed nibble／2-bit planes 解碼 regression 通過；
- `GEO2.DAX`、`GEO3.DAX`、`GEO4.DAX`、`GEO5.DAX`、`GEO6.DAX` 的每個原始 block 都是 `0x402`，且可成功 parse；
- `go test -vet=off ./...` 通過。

## 邊界與未完成項目

本輪只完成幾何資料，不宣稱 tile art、`TILES.DAX`／`8X8D*.DAX` 對應、`WALLDEF*.DAX` wall rendering、移動成本、碰撞、地圖選擇或 Ebiten 畫面已完成。
