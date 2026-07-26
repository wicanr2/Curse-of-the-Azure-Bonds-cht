# 第八十八輪：reference combat placement formula

狀態：`READY`（限 `try_place_combatant` 座標公式 adapter）

## 已確認

reference `ovr011.try_place_combatant(arg_0, arg_2, arg_4, arg_6, arg_8, ...)` 在候選格通過 occupancy／ground checks 後寫入：

```text
pos.x = candidateColumn + teamX * 6 + groupRow * 5 + 22
pos.y = candidateRow + teamY * 5 + 10
```

同一 routine 的 `CombatMap.size` 由 player `field_DE & 7` 先設定；screen position 之後才由 camera top-left 計算。

## 本輪成果

- `combat.ReferencePlacement` 封裝上述公式並加入 regression test。
- 既有 `Fighter.CombatX/Y/Size` 可承接此 adapter 的輸出；沒有真實 placement inputs 時仍使用 formation fallback。

## 邊界

尚未解碼 `team_start_x/y`、`team_direction`、occupancy table `unk_1AB1C`、ground tile validity 與 `field_DE` 的完整角色來源，因此尚未把公式強行套到目前 ECL encounter。

## 驗證

```sh
go test ./...
```
