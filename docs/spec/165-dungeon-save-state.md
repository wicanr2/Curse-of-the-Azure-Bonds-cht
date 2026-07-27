# 第一百六十五輪：dungeon 3D view save state

狀態：`READY`（限 remake JSON save 的地城位置／方向保存）

## 證據與實作

- 第 164 輪已讓 dungeon preview 以 `(dungeonX, dungeonY)` 移動並重建 floor／Far／Mid／Near wall view；本輪把這三個 renderer-driving 欄位提升為 `game.State` 的可保存狀態。
- remake game JSON version 從 `2` 升為 `3`，新增 `dungeon_x`、`dungeon_y`、`dungeon_direction`。
- `F5` 由 `State.SavePartyFile` 寫入目前位置；`F9`／啟動載入會恢復它，Ebiten GEO preview 以恢復後座標重新生成 floor，wall traversal 讀取恢復後方向。
- version `1`／`2` 舊檔仍可載入；沒有新欄位時安全回到 `(8,8)`、方向 `0`。超出 16×16／八方向範圍的 version 3 值也回到同一預設。

## 可沿用的 Gold Box contract

Area／map record 的原始二進位欄位與 remake renderer camera 不應混成同一層。各作品的 save adapter 可把已驗證的 dungeon camera state 映射到共用 `DungeonX/Y/Direction`，缺欄位時使用安全 defaults；完整 DOS `SAVGAM?.DAT` slot、Area1／Area2 寫回與 ECL 導出的真實 facing 仍待反組譯。

## 驗證

`internal/save` codec round-trip、`internal/game` F5/F9 adapter round-trip、`go test ./...` 與兩個 command build 覆蓋本輪邊界。
