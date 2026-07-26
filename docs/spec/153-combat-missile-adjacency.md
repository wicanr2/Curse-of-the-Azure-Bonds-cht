# 第一百五十三輪：missile adjacency guard

## RuleBook／資料證據

RuleBook Combat Fighting 明確規定：角色不能用 missile weapon 攻擊相鄰目標，但 thrown weapon 可以。ITEMS 的 `Range` 會同時出現在遠程與投擲武器，因此不能只用非零 Range 判斷；本輪以已辨識的 bow／crossbow／sling item group（41–47）標記 missile，並保留 dart type 9 的 thrown exception。

## 實作結果

- `EquipmentEffect`／`combat.Fighter` 保存 `WeaponRange`、`MissileWeapon` 與 `ThrownWeapon` profile。
- `Battle.ResolveAttack` 在 attacker／target 都有 CombatMap position 時，拒絕 missile 對相鄰目標的攻擊；沒有 position 時維持 direct API compatibility。
- dart profile 明確允許相鄰攻擊；其餘 thrown weapon 不由本輪臆測成 exception。
- `AttackSequence` 與 CombatAct 共用同一 guard，避免多次攻擊繞過近身限制。

## 明確 boundary

本輪未接入 raw Range 的完整距離單位、line-of-sight、障礙物、手動 Aim cursor、其他 thrown weapon type、彈藥種類／消耗與 ranged attack UI。

