# 第一百四十五輪：MOVE free attack

## RuleBook 證據

RuleBook 的 Combat MOVE 說明：若角色移動離開敵人鄰接範圍，敵人會對該角色進行免費攻擊；背面攻擊的完整 AC／facing 規則另有說明。

## 實作結果

- `Battle.MoveWithFreeAttacks` 先保存移動前座標，再在成功移動後檢查每個存活 enemy。
- 只有「移動前相鄰、移動後不再相鄰」的敵人會對移動者各進行一次 `Battle.Attack`。
- State 將免費反擊結果加入繁中移動訊息；若反擊使戰鬥結束，沿用既有 finish path。
- 舊 `Battle.Move` API 保持相容，僅回傳移動後 fighter；需要觀察反擊的 caller 使用新 API。

## 明確 boundary

目前尚未猜測背面 AC 修正、facing／weapon reach、地形阻擋、負重 movement allowance 或多格移動；本輪只收斂 RuleBook 明確的 adjacency transition 與 free attack trigger。
