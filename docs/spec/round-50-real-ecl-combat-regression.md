# 第五十輪：真實 ECL1 COMBAT boundary regression

狀態：READY（ECL1 block 0x51 的已觀察 journey slice）

## 已完成

- regression 直接讀 repository 中的原始 ZIP、解析 `ECL1.DAX` 與 `MON1CHA.DAX`。
- 以真實 opening 選擇 `JOURNEY ON → STORE` 跑 `game.State`，驗證 ECL1 block `0x51` 到達 `COMBAT` boundary。
- 該路徑的 bounded result 沒有 `LOAD MONSTER` descriptors；State 因此保留事件 boundary，不虛構敵人或錯誤建立空 Battle。
- party 由測試明確注入，monster table 由真實 `MON1CHA` records 解析；`-run-subset -interactive -select 1,1` 可人工重現同一觀察。

## 明確邊界

- 這只證明 ECL1 的一條 journey slice；需再反組該 COMBAT 對應的 encounter／隨機遭遇資料，才能由完整玩家流程建立 Battle。城市／荒野 encounter table、ECL2–ECL6 流程與戰鬥後 continuation 仍未完成。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/game
```
