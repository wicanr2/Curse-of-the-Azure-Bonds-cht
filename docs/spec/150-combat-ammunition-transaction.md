# 第一百五十輪：combat ammunition transaction

## RuleBook／資料證據

RuleBook 武器表標明弓必須有箭、弩必須有弩矢；ITEMS descriptor 的 `AmmunitionType` 保存武器所需的 raw code。實際資料顯示 raw ammunition code 與 inventory item type 不在同一 namespace：例如弓類使用 raw `11`、弩類 raw `138`，而箭／弩矢 inventory type 分別是其他 type。

## 實作結果

- `EquipmentEffect`／`combat.Fighter` 保存 readied weapon 的 raw `AmmunitionType`，不把它誤當 inventory index。
- `Character.ConsumeAmmunition` 接受上層注入的 `raw ammunition type → []inventory item type` mapping，支援堆疊扣除與單件彈藥，並在數量不足或 mapping 缺失時保持 atomic、不修改 inventory。
- `State.SetAmmunitionItemTypes` 保存 mapping copy；CombatAct 在有 mapping 且 fighter 有 raw requirement 時，先驗證並扣除整個本回合 shots，再執行攻擊，避免打到一半才發現彈藥不足。
- 未注入 mapping 時不猜測箭／弩矢對應，維持舊 compatibility path；這讓其他 Gold Box 遊戲可依自身 ITEMS／inventory table 注入 mapping。

## 明確 boundary

本輪未硬編碼 CoAB 的 ammo mapping，未處理彈藥裝填 UI、魔法箭／特殊彈藥、怪物彈藥、彈藥拾取／商店價格與 ranged line-of-sight。這些需要更多 MON*ITM／ECL／實機證據。

