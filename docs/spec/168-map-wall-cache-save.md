# 第一百六十八輪：map wall cache save

狀態：`READY`（限原版 5-byte map state 的 wall cache → remake save）

## 反組譯證據

reference `ovr017.loadSaveGame` 讀取 5 bytes：

| byte | global 欄位 | 用途 |
|---:|---|---|
| 0 | `mapPosX` | dungeon map X |
| 1 | `mapPosY` | dungeon map Y |
| 2 | `mapDirection` | facing |
| 3 | `mapWallType` | 目前 facing wall type cache |
| 4 | `mapWallRoof` | `get_wall_x2(mapPosY,mapPosX)` cache |

`ovr008.MovePositionForward` 與 `ovr015` 的轉向／移動流程會在位置或方向改變後重算 byte 3；byte 4 來自 GEO `x2`，目前 parser 的對應原始欄位是 `Cell.Terrain`。這些 cache 不是完整 `SAVGAM?.DAT` container，但已是其明確的 map-state boundary。

## 實作

- remake save version 升至 `4`，新增 `dungeon_wall_type` 與 `dungeon_wall_roof`。
- `game.State` 在 preview wall preparation 時由 wrapped GEO 重算兩欄；F5 保存，F9／啟動載入 version 4 恢復。
- v1／v2 舊檔與 v3（沒有 wall cache）仍可載入，cache 安全回到 0；座標／方向越界時所有 dungeon state 回到 defaults。

## 明確 boundary

本輪沒有宣稱已解析原版 `SAVGAM?.DAT` header、slot、Area1／Area2、ECL memory 或 player records；也沒有把 cache 當成新的地圖規則。完整 container adapter 仍需實際 save bytes 與 file side effects 證據。

## 驗證

save codec 與 game F5/F9 round-trip 覆蓋 wall type／roof；wrapped GEO preview 仍以計算值更新 cache，並保留 version 3 相容路徑。
