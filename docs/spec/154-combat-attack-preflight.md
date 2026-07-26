# 第一百五十四輪：combat attack preflight

## 問題與資料證據

第 150 輪已建立彈藥扣除的 atomic contract：攻擊失敗時不能先改變 inventory。第 153 輪新增 missile 近身限制後，原本的 CombatAct 順序會先扣除整回合彈藥、才由 Battle 拒絕相鄰攻擊，形成無效攻擊消耗箭／弩矢的 transaction bug。

## READY 實作邊界

- `Battle.ValidateAttack` 驗證存活 fighter、戰鬥狀態與已核對的 missile adjacency guard，不擲骰、不修改 HP。
- `Battle.Attack` 先呼叫 preflight，再消耗 deterministic RNG；直接使用 Battle API 的 invalid attack 不會消耗骰序。
- `State.combatAttackSequence` 在 ammunition transaction 前，先驗證目前 target；被拒絕的相鄰 missile attack 不修改 party inventory 或 target HP。
- 原本的 `ResolveAttack` 仍保留 deterministic injected-roll API，並共用相同 preflight。

## 明確 boundary

本輪不推測 raw `Range` 單位、line-of-sight、障礙物、Aim cursor 或其他 thrown weapon。多目標／多次攻擊若在第一個目標倒下後切換 target，仍由既有 Battle sequence 與後續 target adapter 負責；完整 ranged transaction 需要更多實機規則證據。
