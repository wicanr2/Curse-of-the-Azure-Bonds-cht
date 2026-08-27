# 傷害／治療類法術的骰子

由 `scripts/spell_damage_table.py` 產生，判讀見 [spec 1124](../spec/1124-spell-damage-dice.md)。不要手改。

屬性表的 `+0Ah` 為 0 的法術，傷害不在表裡而在各自的 handler。數字取自**收尾的呼叫點**（傷害走 `sub_F06`、治療走 `<overlay-23 entry#22>`），不是取自擲骰——燃燒之手沒擲骰，治療中傷擲完還加 1。

形狀欄位：`entry9`／`entry10` 是**骰數、面數、加值全是立即數**的那一種，數字可以直接用；`computed` 是有一段由程式算出來（火球 `等級d6`、魔法飛彈 `(等級＋1) div 2` 顆、寒冰錐 `等級d4 ＋ 等級`、電擊之握 `1d8 ＋ 等級`），要人去讀那一支；`flat` 是不擲骰的固定值；`other_finish` 是有擲骰但收尾不是那兩支（豁免即死、反擊、對抗）。

| 編號 | 名稱 | 環 | 職業 | 收尾 | 形狀 | 值 | 屬性旗標 | handler |
|---:|---|---:|---|---|---|---|---|---|
| 3 | Cure Light Wounds | 1 | cleric | heal | entry9 | `1d8` | — | `overlay-22 entry#27` |
| 4 | Cause Light Wounds | 1 | cleric | damage | entry10 | `1d8` | `08h` | `overlay-22 entry#28` |
| 9 | Burning Hands | 1 | magic-user | damage | computed | （不擲骰） | `09h` | `overlay-22 entry#32` |
| 15 | Magic Missile | 1 | magic-user | damage | computed | `?d4` | `08h` | `overlay-22 entry#37` |
| 20 | Shocking Grasp | 1 | magic-user | damage | computed | `1d8` | `0Ch` | `overlay-22 entry#39` |
| 31 | Knock（紮營） | 2 | magic-user | damage | flat | `0` | `00h` | `overlay-22 entry#48` |
| 36 | Animate Dead | 7 |  | none | other_finish | （不擲骰） | — | `overlay-22 entry#53` |
| 37 | Cure Blindness | 3 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#54` |
| 39 | Cure Disease（紮營） | 3 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#57` |
| 41 | Dispel Magic | 3 | cleric | none | other_finish | `1d100` | — | `overlay-22 entry#59` |
| 43 | Remove Curse | 3 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#9` |
| 46 | Dispel Magic | 3 | magic-user | none | other_finish | `1d100` | — | `overlay-22 entry#59` |
| 47 | Fireball | 3 | magic-user | damage | computed | `1d3`、`?d6` | `09h` | `overlay-22 entry#64` |
| 51 | Lightning Bolt | 3 | magic-user | none | other_finish | `?d6` | — | `overlay-22 entry#69` |
| 56 | Restoration | 7 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#71` |
| 58 | Cure Serious Wounds | 4 | cleric | heal | entry9 | `2d8 ＋ 1` | — | `overlay-22 entry#73` |
| 60 | （無名：物品效果 `3Ch`，spec 1169） | 6 |  | none | other_finish | `1d6` | — | `overlay-22 entry#75` |
| 64 | （無名：物品效果 `40h`，spec 1169） | 6 |  | damage | computed | `1d3`、`?d6` | `09h` | `overlay-22 entry#64` |
| 65 | （無名：物品效果 `41h`，spec 1169） | 6 |  | damage | entry10 | `2d4 ＋ 2` | `08h` | `overlay-22 entry#79` |
| 66 | Cause Serious Wounds | 4 | cleric | damage | entry10 | `2d8 ＋ 1` | `08h` | `overlay-22 entry#80` |
| 67 | Neutralize Poison（紮營） | 4 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#81` |
| 68 | Poison | 4 | cleric | damage | flat | `0` | `08h` | `overlay-22 entry#82` |
| 71 | Cure Critical Wounds | 5 | cleric | heal | entry9 | `3d8 ＋ 3` | — | `overlay-22 entry#84` |
| 72 | Cause Critical Wounds | 5 | cleric | damage | entry10 | `3d8 ＋ 3` | `08h` | `overlay-22 entry#85` |
| 74 | Flame Strike | 5 | cleric | damage | entry10 | `6d8` | `09h` | `overlay-22 entry#87` |
| 75 | Raise Dead（紮營） | 5 | cleric | none | other_finish | （不擲骰） | — | `overlay-22 entry#88` |
| 76 | Slay Living | 5 | cleric | none | other_finish | `2d8` | — | `overlay-22 entry#89` |
| 83 | Dimension Door | 4 | magic-user | none | other_finish | （不擲骰） | — | `overlay-22 entry#95` |
| 85 | Fire Shield | 4 | magic-user | none | other_finish | `1d10` | — | `overlay-22 entry#97` |
| 87 | Ice Storm | 4 | magic-user | damage | entry10 | `3d10` | `0Ah` | `overlay-22 entry#99` |
| 89 | Remove Curse | 4 | magic-user | none | other_finish | （不擲骰） | — | `overlay-22 entry#9` |
| 91 | Cloud Kill | 5 | magic-user | none | other_finish | （不擲骰） | — | `overlay-22 entry#101` |
| 92 | Cone of Cold | 5 | magic-user | damage | computed | `?d4` | `0Ah` | `overlay-22 entry#102` |
| 98 | （無名：物品效果 `62h`，spec 1169） | 6 |  | none | other_finish | `6d6` | — | `overlay-22 entry#107` |
| 99 | （無名：物品效果 `63h`，spec 1169） | 6 |  | heal | entry9 | `2d4 ＋ 2` | — | `overlay-22 entry#108` |
| 100 | Bestow Curse | 4 | magic-user | damage | flat | `0` | `00h` | `overlay-22 entry#62` |
