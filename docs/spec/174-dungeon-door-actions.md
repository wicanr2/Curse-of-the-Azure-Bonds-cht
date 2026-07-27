# 第一百七十四輪：dungeon door action adapter

狀態：`READY`

## 本輪成果

將第一百七十三輪的規則接到 remake 的 dungeon preview：`P` 對 `WallDoorFlags == 2` 執行 pick-lock，`K` 對 detail `2/3` 執行 Knock。State 以可重播的 dungeon seed 產生 d100；成功才由 GEO adapter 呼叫雙側 `UnlockDoorWrapped`，失敗保留「本次撬鎖機會已消耗」。Knock 成功前先依隊伍順序消耗第一個 `0x1F` memorized slot。

## 邊界與證據

- `internal/game.State.PickDungeonLock` 只負責 roster／seed／rules transaction，不擁有 GEO map。
- Ebiten adapter 只在 dungeon preview 提供 P/K 快捷鍵與繁中結果訊息，並在成功後重建 floor／wall preview。
- detail `1` 已可直接通行；detail `2` 是可撬鎖的 locked door；reference 的 detail `3` pick menu disabled，因此本輪只允許 Knock。
- 尚未宣稱完整 `locked_door` menu、bash strength/dice、door graphics 或從 ECL 劇情抵達每一扇門；這些仍需各自反組譯與 integration evidence。

## Regression

`internal/game/state_test.go` 覆蓋 seeded roster pick 與 Knock slot mutation；`internal/dungeon` tests 保留 d100 evaluation order、inclusive boundary 與 first-slot semantics。
