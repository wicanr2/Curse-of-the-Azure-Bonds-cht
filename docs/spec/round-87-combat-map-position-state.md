# 第八十七輪：CombatMap position／size state boundary

狀態：`READY`（限 Fighter position state 與目前 renderer adapter）

## 已確認

- reference `CombatantMap` 欄位為 `Point pos`、`int size`、衍生的 `screenPos`。
- reference placement routine 會先寫入 map position，並以 size 檢查 occupied tiles；camera 再將 `pos - mapScreenTopLeft` 轉為 screen position。

## 本輪成果

- `combat.Fighter` 增加 `HasCombatPosition`、`CombatX`、`CombatY`、`CombatSize`。
- `game.StartCombat` 保留外部提供的真實 position／size；缺少時才套用 `FormationTile` fallback。
- Ebiten renderer 優先使用 fighter 的 CombatMap position，讓未來 ECL／combat map decoder 可直接驅動畫面。
- regression 驗證外部 hero position `(4,3)` 保留，enemy fallback 仍建立有效 position。

## 邊界

尚未解碼 reference placement arguments、occupied size map、ground tile collision、camera `mapScreenTopLeft` 與 combat movement；目前 fallback formation 仍不是完整原版戰場生成。

## 驗證

```sh
go test ./...
```
