# 第二百一十二輪：DOS saveVerse 與 ECL DAMAGE adapter

狀態：`READY`（限已驗證 selected-target／whole-party damage branches）

## 證據

公開 CoAB reference 的 `Player.saveVerse` 位於 decompressed player record
`0xDF–0xE3`，`RollSavingThrow` 使用 1..20 natural 1／20 規則，並以 ECL DAMAGE
flags 的低五位作 bonus。`CMD_Damage` 的 selected-character branch 在
`saveFlags & 0x80` 時使用 `saveFlags & 7`（非零時減一）對應 saveVerse；whole-party
branch 使用原始 save type。

## Contract

- DOS parser／JSON 保存五個 raw saving-throw bytes，不重算或覆蓋原版數值。
- `party.Roster.ApplyECLDamage` 只處理 `flags & 0x80` 的 selected／whole-party
  branches，dice 與 saving throw 由注入函式提供，並將 HP clamp 到零。
- State `ResolvePendingECLDamage` 以 working roster transactional writeback，成功後
  清空 pending queue，並依 stable character ID 同步 renderer fighter HP。
- `flags & 0x80 == 0` 的 random-target／`CanHitTarget` branch，以及 affect save bonus、
  party death continuation，仍明確回傳 boundary error，不能假裝已完成。

## 驗收

`internal/party/ecl_damage_test.go` 覆蓋 selected target、natural 1、whole-party
natural 20；`internal/game/state_test.go` 覆蓋 roster／fighter HP writeback 與 queue
consume。下一步是以真實 ECL damage entry 解出 selected-character address 與 random
target context，再擴張剩餘 branches。
