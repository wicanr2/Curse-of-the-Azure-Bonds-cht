# 第三十七輪：可操作戰鬥狀態與 Ebiten 畫面

狀態：READY（戰鬥垂直切片；不是完整 AD&D／DOS 戰鬥宣稱）

## 本輪證據

- `internal/combat.Battle.Attack` 使用 Battle 自有、可由 seed 重現的 d20 與傷害骰。
- `internal/game.State.StartCombat` 接收已解碼的 party／enemy fighters，拒絕空陣營或錯誤 side，建立第一回合。
- `State.CombatAct` 執行目前玩家角色攻擊，敵方回合自動攻擊第一個存活 party，直到下一個玩家回合或戰鬥結束。
- `cmd/azure-bonds-game` 的 `ModeCombat` 畫面顯示繁中戰鬥訊息、雙方 HP、目標游標；左右鍵切換敵人，Enter 攻擊。
- `internal/game/state_test.go` 以一名玩家與一名敵人驗證可操作回合及勝利轉移。

## 邊界與未完成項目

- party roster、裝備效果、法術、逃跑／PARLAY、戰場格位、死亡／狀態效果仍未完成。
- 本輪 API 由資料層傳入 fighters；`COMBAT` 的 ECL signal 尚未自動把已捕獲的 spawn descriptors 與 `MON*CHA` records 建成 Battle。
- 未宣稱數值已完整符合原版；`ResolveAttack` 的已驗證規則仍是精確測試基準。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/combat ./internal/game ./cmd/azure-bonds-game
```
