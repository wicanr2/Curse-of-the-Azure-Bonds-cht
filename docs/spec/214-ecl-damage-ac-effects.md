# 第二百一十四輪：ECL DAMAGE AC／effect projection

狀態：`READY`（限 DOS field_186、target effects 與 health-state projection）

## 證據

CoAB reference `CanHitTarget` 以 d20 與 target AC 比較，natural 1 miss、natural 20
hit；`CheckAffectsEffect(Type_16)` 會處理 `invisibility` `0x19`、`invisible` `0x47`
與其他 combat-round effects，`AffectInvisible` 使 attack roll 減 4，`AffectBlink`
則在 `actions.delay == 0` 將 attack roll 設為 `-1`。Reference 先把 natural 20
放大成 100 再呼叫 Type_16，因此 blink 仍可覆寫 natural 20。Reference player
`field_186` 位於 record `0x186`、為 signed save bonus，`RollSavingThrow` 將它加入
ECL damage save bonus。Reference `damage_player` 將 exact zero 設為 unconscious，
1..9 overkill 設為 dying，10+ overkill 設為 dead；原本 animated 且 exact zero 也
設為 dead，非 OK／animated 狀態的 HP 寫回為 0。

## Contract

- DOS parser／Character／record writeback 保存 `SavingThrowBonus int8`。
- `CanHitECLDamageTargetWithContext` 實作 raw AC、natural 1／20、active invisibility
  `-4` 與 action-delay-aware blink；State default resolver 優先使用 decoded equipment
  AC，未就緒時使用 Character AC，並提供 context variant。
- displace 使用 FX effect-data 第一 byte 的 persistent `0x10` consumed bit，並依
  `combat_round == 0 && attack_roll == 0` 清除該 bit；State context adapter 會 deep-copy
  effect slice，成功時持久化、失敗時 rollback。death transition 與其他
  `CheckAffectsEffect` 項目仍不在本輪假裝完成。
- `Character.HealthStatus`／`DamageOutcome.Health` 保存 OK、animated、unconscious、
  dying、dead projection；`Bleeding` 保存 overkill。`Character.ApplyDeathEffects`／
  `State.ResolveDeathEffects` 已接入已證實的 affect_63 recovery 與 troll_fire_or_acid
  `0x64`→TrollRegen `0x66`（damage flags／combat heal 條件必須明確注入）。完整
  combatant removal 與其他 Death routine 仍由 combat adapter 處理；Battle position
  removal 已由 `HasCombatPosition=false` 對應 reference CombatMap size 0，並發出
  renderer-neutral `Fighter.DeathOverlay` signal。Ebiten 目前保留死亡時的 CombatX/Y
  anchor 並顯示繁中「倒下」overlay；原版 `CombatantKilled` 使用 `combat_icons[24]/[25]`
  skull animation，但目前只證實呼叫索引，尚未證實 CPIC/COMSPR extracted byte family，
  故不硬編素材索引。dragon-slayer `0x4B`
  已由 `ResolveDragonSlayer` 接受 explicit target monster kind、strength damage bonus
  與 d12 roller，避免 Character／ECL DAMAGE 猜測 target。若 State 正在 active combat，
  `Character.RemoveCombatAffects`
  會依 reference 19-kind table 清理 combat-only effects（blink／invisibility 不在清單
  中而保留），`Battle.SetHitPoints` 會同步
  ECL HP 並重新計算 party／enemy status；status 非 active 時走既有 `finishCombat`。

## 驗收

party parser／damage tests 覆蓋 `field_186=-2`、saving roll、invisibility modifier 與
blink 覆寫 natural 20、displace 首次 miss／bit consume／後續命中與 round-start reset；
game test 覆蓋 State default resolver、projected AC、blink context 與 transactional
rollback；party test 覆蓋 exact zero、overkill、animated death、affect_63 recovery 與
troll fire/acid gate，game test 覆蓋 State death-effect transaction。相關 packages Docker
測試通過。
