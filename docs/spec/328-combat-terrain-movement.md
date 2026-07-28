# 第三百二十八輪：戰鬥地形移動成本與大型 footprint

狀態：`READY`

## 資料契約

`mapdata.BackgroundTile` 已保存 reference background table 的：

- `MoveCost`：移入該格消耗的 movement points。
- `0xFF`：不可通行 sentinel。
- `TileIndex`：DUNGCOM／WILDCOM／RANDCOM 圖像 lookup；它不是 movement cost。

Battle 不應 import mapdata 或 renderer。`combat.MovementTerrain` 因此是注入式
callback：給 CombatMap `(x,y)`，回傳正整數 cost 與 passable 狀態。

## Transaction

`MoveWithTerrainAndFreeAttacks` 的順序為：

1. 驗證 fighter、戰鬥狀態與單格 delta。
2. 對目的地完整 footprint 的每一格查 terrain。
3. 任一格不存在、`0xFF` 或 cost 小於 1，整個 move 拒絕且不改座標。
4. footprint 的 movement cost 取所有格最大值。
5. cost 大於本回合 remaining points 時拒絕且不改座標。
6. 通過後才進 occupancy／move-attack／free-attack 舊 transaction。
7. State 依 `MoveResult.MovementCost` 扣點，而非固定扣 1。

這使 2×2 dragon 不可讓左上格通過、其餘身體穿牆；cost-2 的桌椅／困難地形也
不能只消耗一點。

## Coordinate adapter

reference `try_place_combatant` 產生約 `x+22, y+10` 的 CombatMap 絕對座標；
prototype formation fallback 則是 viewport `0..6`。State 在 StartCombat 時保存
此次 battle 的 coordinate namespace，移動後不重新猜測：

- reference coordinates：直接查 50×25 background floor。
- fallback dungeon：viewport `(x,y)` 查 floor `(18+x,7+y)`。
- fallback wilderness：viewport `(x,y)` 查
  `(MapX+x-3, MapY+y-3)`。

Ebiten 只組合這個作品 adapter，合法性與扣點仍由 Battle／State transaction
負責；輸入錯誤沿既有 combat error path 顯示繁中訊息，不會退出程式。

## 驗收

- 2×2 footprint 的非左上目的格 impassable 時拒絕且座標不變。
- 2×1 footprint 取兩格最高 cost。
- remaining 1 不可進 cost 2，State 的位置、move mode 與 remaining 均保持。
- DUNGCOM fallback/reference 與 WILDCOM centered lookup 指向同一原始 entry。
- fallback formation 與 reference placement namespace 在 StartCombat 後固定。

## 邊界

本輪未宣稱已還原 diagonal cost、facing、zone of control、特殊飛行／穿牆、
完整 combat-map bounds 或 camera scroll animation。這些規則需要額外 reference
證據，不能從 `MoveCost` 自行推導。
