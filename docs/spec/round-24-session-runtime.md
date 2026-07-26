# 第二十四輪：session runtime execution

狀態：`READY`（限 bounded session execution）

`BlockSession.RunInteractive` 已接入 game runtime：

- 使用 current block 的 initial entry。
- 依 `SelectionsConsumed` 保留 global selection sequence 的 offset。
- 收到 `NewECLBlockID` 時驗證並切換 target，再以剩餘 selections 繼續。
- 當時的 game command 先載入 ECL1 全部 blocks；第 55 輪已擴展為 ECL1–ECL6 global block namespace，state 使用同一 session。

真實 ECL1 initial-entry graph 目前沒有可達 `NEWECL` edge；後續已在 ECL4／ECL5 event entry 定位 transition。第 54 輪已完成 bounded memory／call stack 保存；party state 與完整劇情 continuation 仍未完成。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -session
```
