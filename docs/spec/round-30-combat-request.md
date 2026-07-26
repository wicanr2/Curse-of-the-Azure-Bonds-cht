# 第三十輪：COMBAT request signal

狀態：`READY`（限控制轉移 signal）

ECL1 block `0x51` 的真實 trace 在 payload `+0x0643` 遇到：

```text
opcode: 0x24 COMBAT
operands: 0
```

公開重製引擎在這個 command 將遊戲狀態轉入 combat loop。本輪加入：

- `RunResult.CombatRequested` signal。
- `BlockSession` 聚合並傳遞該 signal。
- CLI 輸出 `subset requested COMBAT`。
- Game state 顯示繁中戰鬥入口訊息，保留 `OriginalEvent == "COMBAT"`。

這不是完整戰鬥實作；角色、敵人、地圖格、回合順序、命中／傷害、法術、逃跑與戰利品仍未完成，因此 runner 在 signal 處停止，不把 combat 當成普通 fallthrough。

驗證：`TestRunSubsetReportsCombatRequest`、`TestStateExposesCombatEntryFromECL`，以及 ECL1 block `0x51` trace。

- [x] 還原 `COMBAT` 無 operand framing。
- [x] 建立 ECL／session／game state signal。
- [x] 接入繁中戰鬥入口訊息。
- [ ] 建立 party／enemy combat model。
- [ ] 實作 AD&D 回合、骰點與戰鬥 UI。
