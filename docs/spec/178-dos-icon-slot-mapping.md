# 第一百七十八輪：DOS CHEAD/CBODY icon slot mapping

狀態：`READY`

## 證據

reference `engine/ovr017.cs` 的 `LoadPlayerCombatIcon` 以 `sizeToken = { '\\0', 'S', 'T' }` 讀取角色 icon：`CHEAD{token}` 使用 `head_icon`、`CBODY{token}` 使用 `weapon_icon`；`ovr034.chead_cbody_comspr_icon` 在 token `T` 時將 block ID 加 `0x40`。`Player.icon_size` 的 reference comment 是 `1 small / 2 normal`。

## 本輪成果

- `Character.CombatIconBlocks()` 將 raw DOS slots 映射成實際 CHEAD／CBODY block；`icon_size=0` 依 race 使用既有 default。
- `Character.ToFighter` 使用 normalized block；Ebiten loader 載入 extracted CHEAD／CBODY PNG，缺少預合成 party PNG 時以 body→head 順序 on-demand 合成。
- 新增 small (`+0x40`)／normal regression；不改寫保存的 raw `IconHeadBlock`／`IconWeaponBlock` 欄位。
- direction-specific combat placement、color recolor、icon animation 與完整 `CombatIcon` runtime 仍是後續邊界。
