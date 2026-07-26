# 第四十一輪：可翻頁的繁中冒險手札

狀態：READY（八頁繁中摘要資料與 UI 導航）

## 本輪證據

- locale 新增八頁手札內容，涵蓋序章、達倫地理、五方勢力、解除枷印、三件神器、調查提示、戰鬥提示與 Gold Box 歷史。
- `State.JournalPages`／`JournalPage` 保存目前頁面；`NextJournalPage`／`PreviousJournalPage` 有邊界行為。
- Ebiten `J` 開啟手札，左右／上下鍵翻頁，`Esc` 或 `J` 關閉；戰鬥中禁止開啟。
- `internal/game/state_test.go` 驗證開啟、翻到第二頁、返回第一頁與關閉。

## 邊界與未完成項目

- 目前是把 Adventurer's Journal 整理成八頁繁中遊戲內摘要，不是逐條翻譯全部 59 個 Journal Entry 與 Tavern Tale。
- 完整條目觸發／已讀標記、手冊翻譯輪與原始圖片仍未接入。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
