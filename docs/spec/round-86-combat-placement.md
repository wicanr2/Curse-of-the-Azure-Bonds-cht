# 第八十六輪：combat tile placement 與八方向 delta

狀態：`READY`（限目前戰鬥 formation 與 direction contract）

## 已確認

- reference `MapDirectionDelta` 順序為 `N, NE, E, SE, S, SW, W, NW`，座標分別為 `(0,-1)`、`(1,-1)`、`(1,0)`、`(1,1)`、`(0,1)`、`(-1,1)`、`(-1,0)`、`(-1,-1)`。
- reference `CombatMap` 保存 combatant tile position，`screenPos = pos - mapScreenTopLeft`；icon draw 再以 tile 座標轉成畫面位置。
- reference icon direction 使用同一個 0–7 direction，CombatIcon 在 direction `>3` 選水平翻轉圖。

## 本輪成果

- 新增 `internal/combat.DirectionDelta` 與 `FormationTile`，並以 regression 固定八方向順序。
- Ebiten combat renderer 改由 formation tile 計算 party／enemy sprite、姓名與 HP 座標，不再使用單排固定欄位。
- 目前 formation 仍是 deterministic fallback；未來 ECL／combat map record 可替換 tile positions 而不改 renderer contract。

## 邊界

尚未解碼真實 `CombatMap.pos`、size／occupied tile、camera `mapScreenTopLeft` 與完整移動／碰撞；本輪不是宣稱八方向戰鬥規則全部完成。

## 驗證

```sh
go test ./...
```
