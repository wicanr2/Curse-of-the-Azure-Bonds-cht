# 第一百七十一輪：unlocked door movement

狀態：`READY`（限 GEO door detail movement／raw unlock mutation）

## 反組譯證據

- reference `ovr015.TryStepForward`／`locked_door` 以目前 facing 的 `WallDoorFlagsGet` 分流：solid/open path 可繼續，detail `2`／`3` 的 door 要先經 door action。
- `MapSetDoorUnlocked` 將目前方向 `x3` 設為 `1`；bash／pick 成功後也會對相鄰 cell 的 opposite direction 呼叫一次，形成雙側 doorway state。
- skill checks、Knock spell、strength／thief roll 與 `can_bash/pick/knock_door` flags 都在上層 routine，不屬 GEO parser。

## 實作

- `geo.Grid.CanMoveDungeonWrapped`：無 wall 或 detail `1` 可通行；detail `2/3` 阻擋。原有 strict `CanMove`／wrapped `CanMoveWrapped` 仍保留 raw wall semantics。
- `(*geo.Grid).UnlockDoorWrapped`：對目前 cell 與相鄰 cell opposite side 設定 detail `1`，交由上層決定何時呼叫。
- dungeon preview movement 改用 dungeon-specific method，因此已解鎖 doorway 不再被 generic solid-wall guard 錯誤阻擋。

## 明確 boundary

本輪沒有提供玩家 door menu，也沒有猜測 bash／pick／knock 成功率、角色技能、法術消耗、時間、ECL continuation 或 door graphic mutation；目前是可重用 raw movement／mutation adapter。

## 驗證

GEO regression 覆蓋 locked detail `2` 阻擋、unlock 後 detail 雙側寫入與 movement 放行；Docker gate 覆蓋完整測試與 Ebiten build。
