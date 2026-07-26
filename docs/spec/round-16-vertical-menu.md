# 第十六輪：VERTICAL MENU

狀態：`READY`（限 menu framing、prompt、selection observability）

`RunSubset` 已依原版 `CMD_VertMenu` 支援 `0x15`：

```text
[destination word][delay string][option count][option strings...]
```

結果保留 `Menu.Vertical`、`Menu.Prompt`、options 與 selected index，並寫回 destination memory。實際 ECL1 已讀到城市內的垂直／後續 menu 選項，包括 `INN`、`STORE`、`BAR`、`LEAVE`。

尚未完成的是所有 menu command 的 UI 行為、CAMP／城市場所事件與後續戰鬥流程；連續 menu input state 已接入 game session，runner 仍可作為 deterministic harness。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset -select 0,0
```
