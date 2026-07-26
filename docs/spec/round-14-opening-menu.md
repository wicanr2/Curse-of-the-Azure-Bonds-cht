# 第十四輪：開場文字與 menu 接線

狀態：`READY`（限 opening extraction 與 menu observability）

`RunSubset` 現在實作並測試：

- `0x14 COMPARE AND` 的 compare flags。
- `0x2A GETTABLE` indexed memory read。
- `0x2B HORIZONTAL MENU`：destination word、option count、可變 string operands。
- `0x80 length=0` 空 packed string，以及 `0x81` string-memory word operand。

以真實 `ECL1.DAX` initial entry 執行，已讀出：

```text
YOU ARE AT THE EDGE OF
TILVERTON / SHADOWDALE
. WILL YOU ENTER OR CONTINUE YOUR JOURNEY?
menu: ENTER CITY, JOURNEY ON, CAMP
```

`internal/game.NewStateFromECL` 會將第一個 ECL menu 的英文選項保留於 `OriginalChoices`，並以 locale 映射成繁中選項；Ebiten prototype 改為繪製 state 的 choices。未提供 selection 時 runner 仍以 index 0 作為 deterministic extraction，但現在可注入 successive menu selections；CAMP 行為仍未實作。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset
```
