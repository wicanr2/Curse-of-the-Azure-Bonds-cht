# 第三十八輪：ECL encounter 到 Battle 的資料橋

狀態：READY（真實 ECL1 encounter 的直接啟動路徑）

## 本輪證據

- `ecl.RunResult` 現在保留 `SETUP MONSTER` 與每個 `LOAD MONSTER` 的 descriptor；`CLEARMONSTERS` 會清除已累積的 `LOAD MONSTER` 清單，但**保留** `SETUP MONSTER` 寫下的 sprite／圖片／距離（原作 `1Ch` 只釋放怪物鏈與計數，不碰 `ds:7601h`／`7602h`；見 spec 1104 §九之二）。
- `BlockSession.RunFrom` 會聚合跨 bounded result 的 setup／spawn 資料。
- `game.State.StartEncounter` 以 `MonsterSpawn`、`MON*CHA` records 與外部 party roster 呼叫 `monster.BuildEnemies`，再建立 playable Battle。
- `cmd/azure-bonds-game -encounter` 直接執行 ECL1 block `0x51` payload `+0x1293`，載入 `MON1CHA.DAX` 並進入戰鬥畫面。
- `-encounter` 使用明確標示的 debug party，因原始 party save／creation 格式尚未完成；正常啟動仍從原始 opening state 開始。

## 邊界與未完成項目

- 完整玩家流程尚未抵達此 encounter；目前是可重現的 direct-entry vertical slice。
- party roster、`MON*ITM`／`MON*SPC` effects、戰場格位、其餘 monster spells、逃跑／PARLAY 與原版戰鬥數值仍未完整還原；目前僅有 MON*CHA raw spell slots 與 Magic Missile `0x0F` bounded slice。
- 不把 debug party 的數值當成原版資料。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
go run ./cmd/azure-bonds -block-id 81 -encounter-start 4755 -monster-member MON1CHA.DAX
```
