# 第一百七十九輪：DOS icon attack layer

狀態：`READY`

## 證據

reference `CombatIcon.LoadIcons` 載入 normal block 與 `normal_id+0x80` 的 attack block；`GetIcon` 只依 `direction > 3` 選擇水平翻轉副本。這不是另一個任意 sprite family，也不能用 normal CHEAD/CBODY 代替攻擊姿態。

## 本輪成果

- `Character.CombatIconBlocksFor(true)` 映射 normal block 到 attack `+0x80` namespace。
- Ebiten on-demand CHEAD/CBODY composition 在 `IconAttack` 時使用 attack blocks；normal 仍使用 normalized small／normal slots。
- 新增 normal／attack block regression；完整 direction-specific placement、recolor 與 animated CombatIcon cache 仍待完成。
