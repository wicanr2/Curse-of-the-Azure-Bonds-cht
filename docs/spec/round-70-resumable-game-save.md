# 第七十輪：可恢復的 remake game save

狀態：`READY`（限 remake 已實作的 game state）

## 本輪成果

- F5/F9 現在使用版本 2 `GameFile`，除了 party roster，也保存 `area.State`、mode、location 與 map 座標。
- 舊版本 1 的 party-only JSON 仍可載入，並安全回到 wilderness。
- save package 保持與 `internal/game` 解耦，以數字欄位保存 UI-independent state。
- 這是 remake 的 JSON save，不是原版 DOS save slot；原始 Area1／Area2 raw records 仍由 `internal/area` codec 個別處理。

## 驗證

`internal/save` 測試涵蓋 game state round-trip 與舊 party JSON 相容性；`internal/game` 測試涵蓋 F5/F9 adapter。
