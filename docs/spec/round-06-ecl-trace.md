# 第六輪：ECL 安全 trace walker

狀態：`DRAFT`

`internal/ecl.Trace` 現在能以已知 command table 追蹤 decoded block 的 cursor：

- command arity 只來自公開 ECL dump table 的已記錄資料。
- operand 由第五輪 framing parser 消費。
- 未知 opcode、截斷 operand 或 block 邊界會停止並回傳錯誤，不執行猜測語意。
- 目前不做 GOTO/GOSUB 分支、不修改 party、戰鬥或世界狀態。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -trace
```

這是 ECL 直譯器的安全前置層，不是可玩遊戲；需要後續用原版畫面／DOSBox trace 對齊控制流與每個 command 的副作用。
