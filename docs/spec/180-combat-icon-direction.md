# 第一百八十輪：combat icon direction

狀態：`READY`

## 證據

reference `ovr011.SetupCombatActions` 設定 `player.actions.direction = HalfDirToIso[gbl.mapDirection / 2]`，`HalfDirToIso = {7,2,3,6}`；敵方再加 4 modulo 8。`CombatIcon.GetIcon` 依 direction 是否大於 3 選水平翻轉 normal／attack 副本。

## 本輪成果

- `combat.IconDirectionForTeam` 保存 exact map-direction → icon-direction table，並處理 party／enemy opposite facing。
- `State.StartCombat` 依可注入的 `combatMapDirection` 設定所有 fighter 的 `IconDirection`；新增 validation setter 與 StartEncounter regression。
- map direction 尚未由完整 Area／ECL CombatMap loader 導入；目前 default 0，caller 可在 encounter adapter 提供實際方向。
