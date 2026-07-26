# 第二十四輪：session runtime execution

狀態：`READY`（限 bounded session execution）

`BlockSession.RunInteractive` 已接入 game runtime：

- 使用 current block 的 initial entry。
- 依 `SelectionsConsumed` 保留 global selection sequence 的 offset。
- 收到 `NewECLBlockID` 時驗證並切換 target，再以剩餘 selections 繼續。
- game command 載入 ECL1 全部 blocks，state 使用同一 session。

真實 ECL1 initial-entry graph 目前沒有可達 `NEWECL` edge；下一輪在 ECL4 其他 event entry 定位到 real transition。memory／call stack／party state 仍未完整保存。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -session
```
