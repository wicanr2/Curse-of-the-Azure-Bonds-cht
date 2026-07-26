# 第二十二輪：ECL block session loader

狀態：`READY`（限 decoded block ownership／switch contract）

`internal/ecl.BlockSession` 封裝 decoded ECL block ID→payload：

- `CurrentData`／`CurrentBlockID` 取得目前 block。
- `InitialEntry` 由該 block 的第五個初始化 entry 計算 payload offset。
- `Switch` 驗證 target block 存在後才切換。
- `ApplyResult` 可套用 runner 的 `NewECLBlockID` signal。

CLI 驗收 `ECL1.DAX -session`：block `0x50/0x51/0x52` 均成功建立，initial entry 均為 `+0x0014`。

這仍不是完整跨 block VM：selection／memory／call stack 尚未跨 block 保存，game state 目前尚未持有 BlockSession。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -session
```
