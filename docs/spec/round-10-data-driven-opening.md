# 第十輪：ECL data-driven opening

狀態：`DRAFT`

## 實作結果

Ebiten opening prototype 的初始化流程現在是：

```text
ZIP ECL1.DAX
  -> DAX.Parse
  -> first decoded block
  -> ECL.FindPackedTextCandidates
  -> recognize opening marker
  -> game.NewStateFromECL
  -> zh-TW display state
```

`game.State.OriginalOpening` 保存由原始 payload 辨識出的英文 marker；畫面顯示仍由 locale catalog 提供繁中內容。這讓資料來源與翻譯顯示可分別驗證。

## 驗證

```sh
CGO_ENABLED=1 go test -vet=off ./...
go run ./cmd/azure-bonds-game \
  -image curseoftheazurebonds.zip \
  -font /path/to/traditional-chinese-font.ttf
```

GUI 執行需要 host display；無 display 的 CI 只驗證 parser、state transition 與 Ebiten compile。

## 未完成

目前只辨識第一個 opening marker，尚未從 ECL command branch 建立完整事件序列；城市、戰鬥、存檔、圖像與音效仍待實作。
