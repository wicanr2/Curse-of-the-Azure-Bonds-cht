# 第三十六輪：ECL-to-enemy encounter adapter

狀態：`READY`（限 enemy fighter 建立）

`BuildEnemies` 將 ECL `MonsterSpawn` 與 `MON*CHA Record` 合併成 `combat.Fighter`。以 ECL1 block `0x51` payload `+0x1293` 執行至第一個 `COMBAT` 的真實驗證：

```text
spawns=3 enemies=24
FIGHTER x4  (monster 0x59)
BUGBEAR x10 (monster 0x56)
WORG x10    (monster 0x57)
```

CLI：

```sh
go run ./cmd/azure-bonds \
  -image curseoftheazurebonds.zip -member ECL1.DAX \
  -block-id 81 -encounter-start 4755
```

adapter 在第一個 `COMBAT` 停止，避免把同一段 ECL 後續分支的 spawn 當成同一場戰鬥。此輪尚未加入 party、MON*ITM／SPC effects merge、battle map 或回合 UI。

- [x] ECL spawn → MON*CHA → combat.Fighter。
- [x] 真實 Shadowdale encounter regression。
- [x] 第一個 COMBAT 邊界。
- [ ] 建立 party fighter 與 Battle。
- [ ] 接入裝備／effects、戰場與完整 UI。
