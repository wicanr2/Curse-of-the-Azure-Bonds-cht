# 第五十九輪：Ebiten 原始 tile gallery

狀態：`READY`（限 TILES graphics pipeline preview）

## 已確認行為

`cmd/azure-bonds-game` 啟動時載入 `TILES.DAX` 的兩個 block，使用 `internal/gfx.ParsePicture` 與 reference EGA16 `RGBA` adapter 建立 48 個 `*ebiten.Image`。遊戲中按 `T` 可開啟繁中「原始圖塊預覽」，每個 24×24 tile 以 2× scale 顯示並標出原始 item index；按 `T` 或 `Esc` 返回原本 state。

這個 preview 是 graphics pipeline 的可見驗收入口：DAX → indexed nibbles → EGA RGBA → Ebiten image。

## 驗證

- `go test -vet=off ./...` 通過；
- `go build ./cmd/azure-bonds-game` 通過；
- app source 已將 TILES images 接入 Draw path。

## 邊界與未完成項目

preview 尚不是完整 GEO map renderer。GEO cell 到 background tile／wall definition 的原始 index mapping、碰撞、攝影機、地圖移動、音效與完整場所流程仍待完成。
