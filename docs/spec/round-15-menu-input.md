# 第十五輪：ECL menu selection input

狀態：`READY`（限 selection injection 與 opening state）

`RunSubsetWithSelections` 接受 successive `HORIZONTAL MENU` 的 zero-based selection index；每個 menu 將選取值寫回原始 destination memory，並保留 `Menu.Selected`。未提供或越界時使用 0，確保研究工具 deterministic。

`game.State.Select(index)` 將選項索引轉成繁中訊息，並把同一個 index 傳回 ECL subset。Ebiten prototype 使用 Up/Down/Left/Right 移動 cursor、Enter/Space 選擇。

實際 `ECL1.DAX` 驗證：

- selection `0`（ENTER CITY）進入 block 80 的 `0x15 VERTICAL MENU` 路徑；該 parser 已在下一輪完成。
- selection `1`（JOURNEY ON）進入不同的 block 80 路徑，停在另一個未支援 command。

這已是可驗證的 input-to-ECL 分支，不代表戰鬥或完整事件 VM 已完成。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset -select 1
```
