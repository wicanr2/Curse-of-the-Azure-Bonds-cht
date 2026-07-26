# 第九輪：Ebiten opening prototype

狀態：`DRAFT`

`cmd/azure-bonds-game` 已將 `internal/game.State` 接到 Ebiten：

- 啟動時顯示繁中標題與提示。
- Enter／Space 從標題進入開場邊界。
- 方向鍵右側選擇繼續旅程，Enter 選城市。
- locale 與字型都是外部可替換檔案；`-font` 應指定包含繁中字形的 TTF/OTF，未指定時只保留 ASCII fallback。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds-game -font /path/to/cjk-font.ttf
```

目前仍是開場 prototype，不宣稱已完成原版地圖、戰鬥、音效或完整劇情。
