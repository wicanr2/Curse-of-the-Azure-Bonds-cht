# 第五十一輪：ECL RANDOM 與可重現 seed

狀態：READY（bounded VM arithmetic/random slice）

## 已完成

- 實作 reference command table 的 `RANDOM`（opcode `0x08`）：讀取上限與目的 memory address，寫入 `[0, maximum]` 的整數。
- 對小於 `0xFF` 的上限採原版 inclusive increment 行為；`0xFF` 仍代表 256 個可能值。
- 新增 seeded runner 與 `game.State.SetECLSeed`，讓事件／分支 regression 可重播，不把 UI 時間直接混入 VM。
- `RunResult.RandomValues` 保存本次 bounded execution 的隨機輸出，供測試與分析工具核對。
- regression 驗證同 seed 結果一致且落在 inclusive range；full suite 通過。

## 明確邊界

- `RANDOM` 只完成 VM memory side effect；遭遇表、外部 `PROGRAM` routine、完整 wilderness movement 與 ECL 全 opcode 仍待完成。
