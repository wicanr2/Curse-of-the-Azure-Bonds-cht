# 第一百九十七輪：真實 ECL2 encounter Battle

狀態：`READY`

## Evidence

上一輪 smoke report 找到 ECL2 block `3` 的 entry 3（payload `+0x2B0`）會讀取兩個 `LOAD MONSTER` descriptors 並抵達 `COMBAT`。以原始 `MON2CHA.DAX` 載入 records 後，ECL result 可交給 `game.State.StartEncounter`。

## Record adapter

真實 MON records 的 ArmorClass byte 不是直接 signed combat AC：ECL2 `FIRE KNIFE` 的 raw value 是 `59`，同一章其他 records 也落在 `52..60`。已觀察公式為：

```text
combat AC = 60 - raw ArmorClass
```

`monster.CombatArmorClass` 只轉換 50..60；小於 50 的 intermediate／synthetic record 保持原值，避免破壞既有 parser contract。

## Regression and CLI

`TestRealECL2EncounterBuildsBattleFromMON2CHA` 使用 repo 原始 image，解析 ECL2 block 3、執行 entry `0x2B0`、載入 MON2CHA records，最後驗證 `StartEncounter` 建立 active Battle。

可用下列 direct-entry 命令重現：

```text
go run ./cmd/azure-bonds-game \
  -encounter -encounter-block 3 -encounter-start 688 \
  -encounter-monster-member MON2CHA.DAX
```

這是可玩的 cross-chapter encounter slice，不代表普通玩家從 ECL1 開場已自動走到 ECL2，也不代表 ECL2 全部 variable operand／劇情已完成。

## Reuse boundary

後續 Gold Box 遊戲可沿用 `MonsterSpawn → MON*CHA → Fighter` 三層 adapter 與 packed AC normalization pattern，但必須以該作品的 MON record bytes 驗證 AC encoding；不可把 CoAB 的 `60-raw` 直接套到未驗證的作品。
