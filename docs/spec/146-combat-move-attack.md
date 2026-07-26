# 第一百四十六輪：MOVE attack

## RuleBook 證據

RuleBook 的 Combat MOVE 說明：Move 可用來移動角色與攻擊；讓角色移入敵方所在格即可攻擊。若角色移離敵人，敵人另有免費攻擊規則，已於第 145 輪收斂。

## 實作結果

- `Battle.MoveWithFreeAttacks` 先檢查目的格；存活敵人佔據的目的格會呼叫既有 `Battle.Attack`，回傳 `MoveResult.Attack`。
- 這個 bounded core transaction 不把 party fighter 寫入敵人格；攻擊仍從原格發生，直到真正的 CombatMap／occupancy／reach 規則解碼後再擴充。
- 存活隊友佔據的目的格仍拒絕移動；敵方 fighter 移入任何已佔格也仍拒絕。
- State 將移入敵格的攻擊結果沿用既有繁中命中／未命中訊息，並照既有戰鬥勝負 transition 消耗 party turn。

## 明確 boundary

本輪只實作 RuleBook 明確的「移入敵格即攻擊」語意；未猜測地形阻擋、戰場邊界、負重 movement allowance、facing／背面 AC、weapon reach、多格移動與完整動畫。

