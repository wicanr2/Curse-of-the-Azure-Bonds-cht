# 第六十九輪：Area1／Area2 binary codec

狀態：`READY`（限目前已定位欄位）

## 本輪成果

- `internal/area` 可從原版固定大小 `0x800` bytes 的 Area1 record 解碼目前已確認的 map block、地城旗標、座標、ECL block 與城市欄位。
- Area2 可解碼／寫回 `game_area`。
- encode 會複製原始 record，只覆寫已確認欄位，因此未知 bytes 不會在 remake 的局部 round-trip 中消失。
- 短於或長於 `0x800` bytes 的 record 會拒絕處理。

## 已確認 offsets

| record | offset | 欄位 | 型別 |
|---|---:|---|---|
| Area1 | `0x18A` | `current_3DMap_block_id` | byte |
| Area1 | `0x1CC` | `inDungeon` | little-endian signed word |
| Area1 | `0x1E0` | `lastXPos` | little-endian signed word |
| Area1 | `0x1E2` | `lastYPos` | little-endian signed word |
| Area1 | `0x1E4` | `LastEclBlockId` | little-endian word |
| Area1 | `0x342` | `current_city` | byte |
| Area2 | `0x624` | `game_area` | byte |

這是原版 save/import 的可重用 binary 邊界，不代表所有 Area1／Area2 欄位或完整 DOS save slot 已完成。

## 驗證

`go test ./internal/area ./...` 會測試欄位解碼、負座標、未知 bytes preservation 與 record size boundary。
