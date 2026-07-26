# 第四輪：Go 核心層與繁中資源

狀態：`READY`（只涵蓋 DAX container API 與 locale catalog，不涵蓋 ECL 語意）

## 允許實作的契約

- Go `internal/dax.Parse` 讀取已確認的 DAX container。
- parser 必須驗證 header、block boundary、RLE output size，錯誤不可靜默忽略。
- Go `internal/locale` 載入 JSON 字串 catalog，缺少翻譯時回退英文原文。
- 上層遊戲狀態不可假定 ECL opcode 已知；目前只能把 block 當作 payload。

## 驗證

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX
```

## 未涵蓋

ECL 事件控制流、圖片像素格式、音效、戰鬥規則與完整中文翻譯仍需後續規格；這些項目尚不可標示 `READY`。
