# 第十七輪：interactive menu session

狀態：`READY`（限 ECL menu pause／selection sequence）

`RunSubsetInteractive` 與 `game.State` 現在保存 successive menu selections：

- selection sequence 用完時，在下一個 `HORIZONTAL MENU`／`VERTICAL MENU` 回傳 `WaitingForMenu`。
- UI 每次選擇後重新執行相同 ECL prefix，帶入完整 sequence，直到下一個 menu 或 unsupported command。
- game state 將英文 menu option 映射成繁中：`客棧`、`商店`、`酒館`、`離開`，以及 `請按任意鍵或 Enter 繼續`。
- 未提供 menu selection 的 research runner 仍維持 deterministic index 0 行為。

這解決了 menu framing 與連續輸入保存；城市場所的地點選擇已在下一輪對齊，CAMP、場所功能、戰鬥與音效仍未完成。

## 驗收

```sh
go test ./...
go test ./cmd/azure-bonds-game
```
