# 第二百一十四輪：ECL DAMAGE AC／effect projection

狀態：`READY`（限 DOS field_186、target invisibility／blink）

## 證據

CoAB reference `CanHitTarget` 以 d20 與 target AC 比較，natural 1 miss、natural 20
hit；`CheckAffectsEffect(Type_16)` 會處理 `invisibility` `0x19`、`invisible` `0x47`
與其他 combat-round effects，`AffectInvisible` 使 attack roll 減 4，`AffectBlink`
則在 `actions.delay == 0` 將 attack roll 設為 `-1`。Reference 先把 natural 20
放大成 100 再呼叫 Type_16，因此 blink 仍可覆寫 natural 20。Reference player
`field_186` 位於 record `0x186`、為 signed save bonus，`RollSavingThrow` 將它加入
ECL damage save bonus。

## Contract

- DOS parser／Character／record writeback 保存 `SavingThrowBonus int8`。
- `CanHitECLDamageTargetWithContext` 實作 raw AC、natural 1／20、active invisibility
  `-4` 與 action-delay-aware blink；State default resolver 優先使用 decoded equipment
  AC，未就緒時使用 Character AC，並提供 context variant。
- displace 仍需要 persistent affect-data context，由 injected hit resolver 處理；death
  transition 與其他 `CheckAffectsEffect` 項目不在本輪假裝完成。

## 驗收

party parser／damage tests 覆蓋 `field_186=-2`、saving roll、invisibility modifier 與
blink 覆寫 natural 20；game test 覆蓋 State default resolver、projected AC 與 blink
context。相關 packages Docker 測試通過。
