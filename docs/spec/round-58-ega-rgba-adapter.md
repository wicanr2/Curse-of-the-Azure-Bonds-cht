# 第五十八輪：EGA indexed pixels 到 RGBA adapter

狀態：`READY`（限 palette／透明色 adapter）

## 已確認行為

`internal/gfx.Picture.RGBA` 將原始 4-bit indexed picture item 轉為標準 `image.RGBA`：

- 使用 reference engine 的 16 色 EGA palette（包含 173／82／255 三段亮度）；
- palette index `16` 轉為 alpha 0，保留 masked DAX picture 的 transparent sentinel；
- 其他 index 必須落在 0–15，越界會回傳 error。

這個 adapter 不依賴 Ebiten，因此同一份像素可供 Ebiten、PNG exporter 與 unit test 使用。

## 驗證

- synthetic indexed picture regression 確認 EGA blue 像素值與 mask alpha；
- 原始 TILES／8X8D picture parser regression 仍通過；
- `go test -vet=off ./...` 通過。

## 邊界與未完成項目

本輪尚未把 RGBA image 接入遊戲 viewport；EGA palette mutation、tile／symbol index mapping、GEO map composition、音效與實際 Ebiten map screen 仍待完成。
