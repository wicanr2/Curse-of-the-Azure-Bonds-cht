# 第二百七十九輪：正式序幕 EXIT → 荒野主迴圈

狀態：READY

## Evidence

真實 ECL2 block `0x01` initial entry 在兩次 Continue selection 後執行 34 steps，
最後一條指令是 `+0x1CBB EXIT`，結果 PC 為 `+0x1CBC`。序幕前段已寫入：

- `SAVE 7, 0xC04B`：map X；
- `SAVE 13, 0xC04C`：map Y；
- `SAVE 1, 0xC04D`：方向。

Reference `ovr003.sub_29758` 在 initial `RunEclVm` 返回後，若不是 demo，保存目前 ECL
block 並進入 world loop。作品資料／Area 的 outdoor 狀態使 `game_state` 選為
`WildernessMap`。這不是另一個 ECL menu。

## Remake transaction

`RunResult.Exited` 明確表示 opcode `0x00` 已完成該次 VM lifecycle；menu pause、步數用盡
或 unsupported opcode 不得冒充 EXIT。`BlockSession.MemoryValue` 只讀 shared VM word，
讓作品 adapter 取得 script 寫入的 world registers，而不把 CoAB 位址硬編進共用 VM。

State 僅在 active new-game block `0x01` initial transaction 收到 EXIT 時：

1. 讀取 `0xC04B..0xC04D`；
2. 設定 wilderness location；
3. 進入 `ModeMap (7,13)`、方向 `1`；
4. 清除 Continue choices，不顯示人造「進入城市／繼續旅程／紮營」選單。

## Regression

真實 image integration 從無隊伍 title 開始，完成角色建立與兩次 Continue，驗證最後為
block `0x01`、`ModeMap`、座標 `(7,13)`、方向 `1`、wilderness location 且 choices 為空。

```text
go test ./internal/ecl ./internal/game
```
