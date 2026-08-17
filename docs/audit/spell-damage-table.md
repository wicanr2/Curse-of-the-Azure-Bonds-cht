# 傷害／治療類法術的骰子

由 `scripts/spell_damage_table.py` 產生，判讀見 [spec 1124](../spec/1124-spell-damage-dice.md)。不要手改。

屬性表的 `+0Ah` 為 0 的法術，傷害不在表裡而在各自的 handler。
**骰數 0 代表用施法者等級當骰數**（魔法飛彈 `等級d4`）。

`ambiguous` 是 handler 裡擲了不只一次骰——**第一次不一定是傷害**，要人去讀那一支才知道哪一次是。只取第一次會產生看起來有答案但是錯的列。

| 編號 | 名稱 | 環 | 職業 | 形狀 | 骰 | handler |
|---:|---|---:|---|---|---|---|
| 3 | Cure Light Wounds | 1 | cleric | entry9 | `1d8` | `overlay-22 entry#27` |
| 4 | Cause Light Wounds | 1 | cleric | entry10 | `1d8` | `overlay-22 entry#28` |
| 9 | Burning Hands | 1 | magic-user | unread | — | `overlay-22 entry#32` |
| 15 | Magic Missile | 1 | magic-user | entry10 | `等級d4` | `overlay-22 entry#37` |
| 20 | Shocking Grasp | 1 | magic-user | entry10 | `1d8` | `overlay-22 entry#39` |
| 31 | Knock（紮營） | 2 | magic-user | unread | — | `overlay-22 entry#48` |
| 36 | Animate Dead | 7 |  | unread | — | `overlay-22 entry#53` |
| 37 | Cure Blindness | 3 | cleric | unread | — | `overlay-22 entry#54` |
| 39 | Cure Disease（紮營） | 3 | cleric | unread | — | `overlay-22 entry#57` |
| 41 | Dispel Magic | 3 | cleric | entry9 | `1d100` | `overlay-22 entry#59` |
| 43 | Remove Curse | 3 | cleric | unread | — | `overlay-22 entry#9` |
| 46 | Dispel Magic | 3 | magic-user | entry9 | `1d100` | `overlay-22 entry#59` |
| 47 | Fireball | 3 | magic-user | ambiguous | `1d3`、`等級0d6` | `overlay-22 entry#64` |
| 51 | Lightning Bolt | 3 | magic-user | unread | — | `overlay-22 entry#69` |
| 56 | Restoration | 7 | cleric | unread | — | `overlay-22 entry#71` |
| 58 | Cure Serious Wounds | 4 | cleric | entry9 | `2d8` | `overlay-22 entry#73` |
| 66 | Cause Serious Wounds | 4 | cleric | entry10 | `2d8` | `overlay-22 entry#80` |
| 67 | Neutralize Poison（紮營） | 4 | cleric | unread | — | `overlay-22 entry#81` |
| 68 | Poison | 4 | cleric | unread | — | `overlay-22 entry#82` |
| 71 | Cure Critical Wounds | 5 | cleric | entry9 | `3d8` | `overlay-22 entry#84` |
| 72 | Cause Critical Wounds | 5 | cleric | entry10 | `3d8` | `overlay-22 entry#85` |
| 74 | Flame Strike | 5 | cleric | entry10 | `6d8` | `overlay-22 entry#87` |
| 75 | Raise Dead（紮營） | 5 | cleric | unread | — | `overlay-22 entry#88` |
| 76 | Slay Living | 5 | cleric | entry10 | `2d8` | `overlay-22 entry#89` |
| 83 | Dimension Door | 4 | magic-user | unread | — | `overlay-22 entry#95` |
| 85 | Fire Shield | 4 | magic-user | entry9 | `1d10` | `overlay-22 entry#97` |
| 87 | Ice Storm | 4 | magic-user | entry10 | `3d10` | `overlay-22 entry#99` |
| 89 | Remove Curse | 4 | magic-user | unread | — | `overlay-22 entry#9` |
| 91 | Cloud Kill | 5 | magic-user | unread | — | `overlay-22 entry#101` |
| 92 | Cone of Cold | 5 | magic-user | entry10 | `等級d4` | `overlay-22 entry#102` |
| 100 | Bestow Curse | 4 | magic-user | unread | — | `overlay-22 entry#62` |
