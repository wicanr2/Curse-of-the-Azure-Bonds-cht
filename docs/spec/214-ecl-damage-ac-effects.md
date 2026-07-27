# 第二百一十四輪：ECL DAMAGE AC／effect projection

狀態：`READY`（限 DOS field_186 與 target invisibility）

## 證據

CoAB reference `CanHitTarget` 以 d20 與 target AC 比較，natural 1 miss、natural 20
hit；`CheckAffectsEffect(Type_16)` 會處理 `invisibility` `0x19`、`invisible` `0x47`
與其他 combat-round effects，`AffectInvisible` 使 attack roll 減 4。Reference
player `field_186` 位於 record `0x186`、為 signed save bonus，`RollSavingThrow` 將它
加入 ECL damage save bonus。

## Contract

- DOS parser／Character／record writeback 保存 `SavingThrowBonus int8`。
- `CanHitECLDamageTarget` 實作 raw AC、natural 1／20 與 active invisibility `-4`；
  State default resolver 優先使用 decoded equipment AC，未就緒時使用 Character AC。
- blink／displace 需要 combat round／affect-data context，仍由 injected hit resolver
  處理；death transition 與其他 `CheckAffectsEffect` 項目不在本輪假裝完成。

## 驗收

party parser／damage tests 覆蓋 `field_186=-2`、saving roll、invisibility modifier；
game test 覆蓋 State default resolver 與 projected AC。相關 packages Docker 測試通過。
