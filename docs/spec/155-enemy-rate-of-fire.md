# 第一百五十五輪：enemy RateOfFire turn

## RuleBook／既有資料證據

第 149 輪已確認 RuleBook 的弓／飛鏢每回合攻擊次數，以及 ITEMS `RateOfFire` → `AttacksPerTurn` projection。既有 party turn 已使用 `Battle.AttackSequence`，但 enemy turn 仍固定呼叫單次 `Battle.Attack`，因此同一個已投影到 `combat.Fighter` 的武器 profile 在敵方回合會被忽略。

## 實作結果

- enemy `Fighter.AttacksPerTurn > 1` 時，`State.advanceCombatToParty` 改走 `Battle.AttackSequence`。
- sequence 仍使用 deterministic RNG、目標死亡即停止，並以繁中多次攻擊摘要顯示。
- `AttacksPerTurn <= 1` 維持單次攻擊相容路徑；本輪不猜測怪物額外職業攻擊、AI 換目標或特殊彈藥。

## 明確 boundary

敵方彈藥 inventory、敵方 Aim／line-of-sight、地形／移動 AI、back stab 與完整 enemy target policy 仍需 MON*／ECL／實機證據；本輪只修正已存在 `Fighter.AttacksPerTurn` profile 的回合套用。
