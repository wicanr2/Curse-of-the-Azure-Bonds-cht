# 第二百一十三輪：ECL DAMAGE random-target／CanHitTarget

狀態：`READY`（限注入式 random target 與命中 resolver）

## 證據

CoAB reference `CMD_Damage` 在 `flags & 0x80 == 0` 時，以低八位作 target count，
每次用 party size random roll 選目標，呼叫 `CanHitTarget(var_6, target)`；命中後扣
damage，並為下一次嘗試重新擲 damage。`CanHitTarget` 的 reference contract 是
natural 1 miss、natural 20 hit，否則 d20 加 raw operand bonus 並與 target AC 比較。

## Contract

- `party.ApplyECLDamageWithHitResolver` 保存 target order／hit／applied damage，使用
  注入 `rollDie` 保持可重播。
- `DamageHitResolver` 接收 target、raw `saveFlags` byte 與 d20 source；AC／invisibility
  affect 等作品資料由 resolver 自己投影，不在 party damage core 猜測。
- State 提供 `ResolvePendingECLDamageWithHitResolver`，成功後 transactional 寫回 roster
  與 stable-ID fighter HP。
- 原版 death transition、`CheckAffectsEffect` 與完整 AC projection 仍是後續 adapter；
  resolver 回傳 error 時不清空 pending request。

## 驗收

`internal/party/ecl_damage_test.go` 與 `internal/game/state_test.go` 驗證 random target
順序、raw hit bonus、miss／hit 與 HP sync；selected／whole-party regressions 仍維持。
