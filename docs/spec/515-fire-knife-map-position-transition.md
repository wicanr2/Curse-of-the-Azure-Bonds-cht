# 第 515 輪：火刀戰後地圖位置轉移資料契約

狀態：`READY`（有界正常玩家切片；轉移來源仍是 `strong inference`）
日期：2026-08-09
## 目的與限制

本輪關閉第 514 輪留下的第一個座標輔助：火刀檢查站戰勝後，玩家按下
「繼續」時，remake 不再由測試直接把地城座標寫成 `(13,10)`。ECL 同一個
session 先續跑原始 block 3 的清理 helper，接著由 JSON game-pack 宣告一筆
作品中立的 `set_map_position`，State adapter 才把原生位置投影回 GEO2 block 3。

這不是「已找出 DOS 外部移動函式的 exact 實作」聲明。ECL work
`4BF0h／4BF1h` 與 DOS overlay 22 的 indexed operand `[di+4BF0h]` 尚未證明
屬於同一位址空間；兩者的 writer→projection→ECL consumer 完整橋接仍未閉合。
`(13,10)` 是由 raw ECL helper、
PC-98 `MOVEFORWARD` 對照、GEO 關閉狀態可行圖分析與既有事件路徑共同支持的
`strong inference`。下一個火刀據點 block 4 的 `(8,15)` handoff 仍未完成。

## 證據登錄

| 證據 | 位址／輸入基準 | 結果 | 等級 |
|---|---|---|---|
| DOS ECL 原始資料 | `ECL2.DAX` SHA-256 `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338`；`cmd/azure-bonds` typed decoder；ECL block 3 offset `+1B5Bh` | 清理後 branch 會呼叫 `2E10h`，但該次 runtime result 沒有 `C04Bh/C04Ch` `SAVE` | `exact` |
| DOS runtime continuation | `START.EXE` SHA-256 `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5`；`GAME.OVR` SHA-256 `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215`；Docker `coab-go-test:20260729`、Go `1.24.13` | 火刀勝利後按下 PRESS 的 `RunResult` 到 ECL block 3 `PC=1B5Bh`，含 `CALL 2E10h`、無 `SaveWrites`；原 State 仍為 `(1,8,S)` | `exact`（remake runtime observation） |
| PC-98 movement 對照 | `PC98-GAME.EXE` SHA-256 `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`；overlay 07 `MOVEFORWARD` `0062:1AEC`；`VARS3D` `0C29:A2A9..A2AD` | 原版 movement routine 以方向更新座標／地圖服務，最後寫 `POSCHANGE=1`；不能把 DOS `4BF0h` 直接等同 PC-98 位址 | `exact`（PC-98 bytes）；跨版本橋接 `strong inference` |
| GEO 靜態可行圖 | `GEO2.DAX` block 3；16×16 decoded grid | 在第三平面／門狀態未變更時，`(1,8)` 到 `(13,10)` 與 `(13,10)` 到 `(8,15)` 都沒有符合目前 movement predicate 的路徑；這不代表地圖永久不相連，`wall=09/detail=0` 的秘密門語意仍未知 | `exact`（關閉狀態 geometry）；外部 handoff 語意 `strong inference` |
| 目的格 | 同一 `GEO2.DAX` block 3 的 terrain／wall record | `(13,10,E)` 是 `terrain=83h`，wall `0Fh`；其後由既有 ECL 事件顯示 Myth Drannor 騎士 | `exact`（GEO／事件 observation）；跨 component 轉移 `strong inference` |

所有十六進位數字均保留其位址空間；ECL work address、file offset、PC-98
`segment:offset` 與 GEO cell 不可互換。此規格不把相同數字當作同一個 record。

## Engine／JSON 契約

Golden Box engine 新增作品中立的 `MapPositionTransition` 與
`set_map_position` action：

```json
{
  "type": "set_map_position",
  "mode": "dungeon",
  "position": {
    "map_kind": "first_person",
    "area_id": 2,
    "geometry_block": 3,
    "x": 13,
    "y": 10,
    "direction": 2,
    "wall_type": 15,
    "wall_roof": 131
  }
}
```

Engine 只驗證欄位形狀、非負座標與 cardinal direction，並把結果放在
`Runtime.MapPositions`；不解讀 CoAB 劇情、城市名稱、旗標意義或地圖大小。
State adapter 負責：

- 設定 `Area.GameArea`、`GeoMapSet／GeoMapBlock` 與 renderer map request；
- 更新 `DungeonX／DungeonY／DungeonDirection`、wall／roof registers；
- 將同一位置寫回 ECL `C04B..C04F`，保持後續地城 lifecycle 使用同一 session；
- 僅在沒有 `CombatRequested` 時套用資料事件，避免在戰鬥入口提前傳送。

CoAB event 的 predicate 目前使用戰後 runtime 觀察到的
`ECL block 3`、`7F72h=FDh` 與 `4C2Bh=1`；這是資料層的可回查條件，不是 Go
中的劇情 `if`。事件為 `once`，避免同一戰後 continuation 重播轉移。

## 驗證

Docker focused regression：

```text
角色建立 → Tilverton → 公會 → 下水道正常移動
→ 火刀檢查站 → 拒絕投降 → 五名 Fire Knife 戰鬥
→ 勝利 PRESS → JSON map-position transition
→ (13,10,E) → Myth Drannor 騎士事件
```

回歸同時確認拒絕投降仍先建立戰鬥，不能被戰後 predicate 提前攔截；事件文字、
裝備與規則沒有複製到測試，位置來自 game-pack event。後續仍須把騎士後的
外部地圖 handoff、block 4 `(8,15)`、完整戰鬥演出與整作正常玩家路徑閉合。
