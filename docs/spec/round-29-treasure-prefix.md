# 第二十九輪：TREASURE bounded prefix

狀態：`READY`（限 operand framing 與安全前綴）

公開 ECL command table 將 `0x27` 定義為 `TREASURE`、8 個 operands。Shadowdale block `0x51` 的場所流程在此 command 停止，因此本輪：

- 保留 `0x27` 的 8-operand framing。
- bounded runner 消耗完整指令後繼續執行。
- 不虛構金錢、物品、角色或 party inventory 效果。

驗證：`TestRunSubsetConsumesTreasureOperandsAsBoundedNoOp`。這讓 trace 能繼續探索後續 event；它不是完整 TREASURE implementation。

- [x] 以 command metadata 解析 8 operands。
- [x] 加入安全 bounded no-op regression。
- [ ] 解碼 treasure table、party inventory 與原始獎勵規則。
