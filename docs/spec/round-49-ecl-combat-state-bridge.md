# 第四十九輪：ECL COMBAT 到可操作 Battle

狀態：READY（已載入 party／MON*CHA 的 combat bridge）

## 已完成

- `game.State` 可接收解碼後的 `MON*CHA` records，並保存明確的 encounter combat seed。
- 當 ECL runner 回傳 `COMBAT`，且 State 已有 party 與 monster records 時，`Select` 直接呼叫 `StartEncounter`，由 `LOAD MONSTER`／`SETUP MONSTER` descriptors 與 records 建立 Battle。
- Ebiten 啟動時載入 `MON1CHA.DAX`；搭配角色建立後的 party 或 `-party-load`，ECL 觸發的戰鬥不再依賴 `-encounter` debug party。
- 沒有 party 或 monster table 時仍保留明確的 COMBAT event boundary，避免虛構敵人資料。
- regression 覆蓋 synthetic ECL `LOAD MONSTER → COMBAT` 到實際 `CombatActive` 的 state transition。

## 明確邊界

- `MON*ITM`／`MON*SPC` 尚未完整合併進玩家或敵人戰鬥；戰場格位、法術、逃跑／PARLAY、完整遭遇後 ECL continuation 仍待完成。
- `-encounter` 仍保留作為可重現的 direct-entry debug harness，不代表完整玩家流程已全部抵達該 encounter。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/game ./cmd/azure-bonds-game
```
