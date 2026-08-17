# 玩家法術覆蓋（原作 100 筆 → remake）

由 `cmd/combat-spell-coverage-audit -output` 產生，不要手改。

- 分母是原作法術主表的 **100 筆**（`gamepack/rules/spell-table.json`，由 `cmd/spell-table-export` 從常駐資料段量出來，spec 1111）。
- **占位**：沒有名字的 13 筆，玩家取不到。**紮營**：`+0Bh = 0` 的 8 支，不會出現在戰鬥選單（spec 827）。
- `handler`／`visual`／`sound` 三欄是 game pack 宣告與 runtime callsite 的機器觀察，**不代表原版規則、時序或素材已還原**。
- 「資料」欄一律有值：施法時間、目標形狀、豁免、持續時間係數整張表都在，沒有 handler 的法術缺的是**效果**，不是資料。

## 摘要

| 分類 | 支數 |
|---|---:|
| 表的筆數 | 100 |
| ├ 占位（玩家取不到）| 13 |
| └ 玩家取得到 | 87 |
| 　├ 只能紮營施放 | 8 |
| 　└ **戰鬥可施放（真正的分母）** | **79** |
| 　　├ 職業模型不支援（德魯伊／表裡沒有職業）| 6 |
| 　　├ **可宣告（能收斂到 0 的分母）** | **73** |
| 　　├ game pack 已宣告 | 73 |
| 　　├ 其中 runtime handler 已觀察到 | 73 |
| 　　└ **尚未宣告** | **0** |
| 　　　├ 效果碼 remake 已判讀（只差宣告） | 0 |
| 　　　├ 效果碼 remake 還看不懂 | 0 |
| 　　　└ 傷害類（`+0Ah = 0`，骰數不在屬性表裡） | 0 |

全表用到 50 個相異效果碼，remake 判讀得了其中 20 個。
★ **記得上去不等於解讀得了**：`CastEffectSpell` 可以把任何碼寫進效果串列，
但只有已判讀的那幾個會改變戰鬥規則。上面兩列就是這條界線。

可宣告的分母已經歸零：戰鬥可施放而職業模型支援的法術全部宣告了。

## 逐支

| 編號 | 名稱 | 職業 | 環 | 資料 | handler | visual | sound |
|---:|---|---|---:|---|---|---|---|
| 1 | Bless | cleric | 1 | 10 節／半徑 2／效果 1 | observed | partial | observed |
| 2 | Curse | cleric | 1 | 10 節／半徑 2／效果 2 | observed | partial | observed |
| 3 | Cure Light Wounds | cleric | 1 | 5 節／1 個目標 | observed | partial | observed |
| 4 | Cause Light Wounds | cleric | 1 | 5 節／1 個目標 | observed | partial | observed |
| 5 | Detect Magic | cleric | 1 | 1 節／自己／效果 5 | observed | partial | observed |
| 6 | Protection From Evil | cleric | 1 | 4 節／1 個目標／效果 8 | observed | missing | observed |
| 7 | Protection from Good | cleric | 1 | 4 節／1 個目標／效果 9 | observed | missing | observed |
| 8 | Resist Cold | cleric | 1 | 10 節／1 個目標／效果 10 | observed | partial | observed |
| 9 | Burning Hands | magic-user | 1 | 1 節／1 個目標 | observed | partial | observed |
| 10 | Charm Person | magic-user | 1 | 1 節／1 個目標／豁免 spell／效果 11 | observed | partial | observed |
| 11 | Detect Magic | magic-user | 1 | 1 節／自己／效果 5 | observed | partial | observed |
| 12 | Enlarge | magic-user | 1 | 1 節／1 個目標／效果 12 | observed | partial | observed |
| 13 | Reduce | magic-user | 1 | 1 節／1 個目標／豁免 spell／效果 13 | observed | partial | observed |
| 14 | Friends | magic-user | 1 | 1 節／自己／效果 14 | 紮營法術 | — | — |
| 15 | Magic Missile | magic-user | 1 | 1 節／1 個目標 | observed | observed | observed_shared |
| 16 | Protection From Evil | magic-user | 1 | 1 節／1 個目標／效果 8 | observed | missing | observed |
| 17 | Protection From Good | magic-user | 1 | 1 節／1 個目標／效果 9 | observed | missing | observed |
| 18 | Read Magic | magic-user | 1 | 10 節／自己／效果 16 | 紮營法術 | — | — |
| 19 | Shield | magic-user | 1 | 1 節／自己／效果 17 | observed | partial | observed |
| 20 | Shocking Grasp | magic-user | 1 | 1 節／1 個目標 | observed | partial | observed |
| 21 | Sleep | magic-user | 1 | 1 節／半徑 1／效果 53 | observed | observed | observed |
| 22 | Find Traps | cleric | 2 | 5 節／自己／效果 19 | 紮營法術 | — | — |
| 23 | Hold Person | cleric | 2 | 5 節／3 個目標／豁免 spell／效果 52 | observed | partial | observed |
| 24 | Resist Fire | cleric | 2 | 5 節／1 個目標／效果 20 | observed | partial | observed |
| 25 | Silence, 15' Radius | cleric | 2 | 5 節／鎖定目標／豁免 spell／效果 21 | observed | partial | observed |
| 26 | Slow Poison | cleric | 2 | 1 節／1 個目標／效果 22 | observed | partial | observed |
| 27 | Snake Charm | cleric | 2 | 5 節／自己／效果 51 | observed | partial | observed |
| 28 | Spiritual Hammer | cleric | 2 | 5 節／自己／效果 23 | observed | partial | observed |
| 29 | Detect Invisibility | magic-user | 2 | 2 節／自己／效果 24 | observed | partial | observed |
| 30 | Invisibility | magic-user | 2 | 2 節／1 個目標／效果 25 | observed | partial | observed |
| 31 | Knock | magic-user | 2 | 1 節／自己 | 紮營法術 | — | — |
| 32 | Mirror Image | magic-user | 2 | 2 節／自己／效果 28 | observed | partial | observed |
| 33 | Ray of Enfeeblement | magic-user | 2 | 2 節／1 個目標／豁免 spell／效果 29 | observed | partial | observed |
| 34 | Stinking Cloud | magic-user | 2 | 2 節／半徑 1／豁免 paralysis-poison-death／效果 30 | observed | observed | observed |
| 35 | Strength | magic-user | 2 | 10 節／自己／效果 38 | 紮營法術 | — | — |
| 36 | Animate Dead |  | 7 | 0 節／半徑 0／豁免 spell | 未宣告 | — | — |
| 37 | Cure Blindness | cleric | 3 | 10 節／1 個目標 | observed | missing | observed |
| 38 | Cause Blindness | cleric | 3 | 10 節／1 個目標／豁免 spell／效果 33 | observed | partial | observed |
| 39 | Cure Disease | cleric | 3 | 100 節／自己 | 紮營法術 | — | — |
| 40 | Cause Disease | cleric | 3 | 100 節／1 個目標／豁免 spell／效果 34 | observed | partial | observed |
| 41 | Dispel Magic | cleric | 3 | 4 節／半徑 1 | observed | missing | observed |
| 42 | Prayer | cleric | 3 | 6 節／自己／效果 49 | observed | partial | observed |
| 43 | Remove Curse | cleric | 3 | 6 節／1 個目標 | observed | missing | observed |
| 44 | Bestow Curse | cleric | 3 | 6 節／1 個目標／豁免 spell／效果 36 | observed | partial | observed |
| 45 | Blink | magic-user | 3 | 1 節／自己／效果 37 | observed | partial | observed |
| 46 | Dispel Magic | magic-user | 3 | 3 節／半徑 1 | observed | missing | observed |
| 47 | Fireball | magic-user | 3 | 3 節／半徑 3／豁免 spell | observed | observed | observed |
| 48 | Haste | magic-user | 3 | 3 節／半徑 2／效果 39 | observed | partial | observed |
| 49 | Hold Person | magic-user | 3 | 3 節／4 個目標／豁免 spell／效果 52 | observed | partial | observed |
| 50 | Invisibility, 10' Radius | magic-user | 3 | 3 節／半徑 1／效果 25 | observed | partial | observed |
| 51 | Lightning Bolt | magic-user | 3 | 3 節／半徑 0／豁免 spell | observed | observed | observed |
| 52 | Protection From Evil, 10' Radius | magic-user | 3 | 3 節／1 個目標／效果 45 | observed | partial | observed |
| 53 | Protection From Good, 10' Radius | magic-user | 3 | 3 節／1 個目標／效果 46 | observed | partial | observed |
| 54 | Protection From Normal Missiles | magic-user | 3 | 3 節／1 個目標／效果 41 | observed | partial | observed |
| 55 | Slow | magic-user | 3 | 3 節／半徑 2／效果 42 | observed | partial | observed |
| 56 | Restoration | cleric | 7 | 6 節／1 個目標 | observed | missing | observed |
| 57 | （占位） | — | 6 | 0 節／自己／效果 39 | 取不到 | — | — |
| 58 | Cure Serious Wounds | cleric | 4 | 7 節／1 個目標 | observed | partial | observed |
| 59 | （占位） | — | 6 | 0 節／自己／效果 38 | 取不到 | — | — |
| 60 | （占位） | — | 6 | 0 節／1 個目標／豁免 spell | 取不到 | — | — |
| 61 | （占位） | — | 6 | 0 節／1 個目標／豁免 paralysis-poison-death／效果 52 | 取不到 | — | — |
| 62 | （占位） | — | 6 | 0 節／自己／效果 39 | 取不到 | — | — |
| 63 | （占位） | — | 6 | 0 節／4 個目標／效果 71 | 取不到 | — | — |
| 64 | （占位） | — | 6 | 0 節／半徑 3／豁免 spell | 取不到 | — | — |
| 65 | （占位） | — | 6 | 0 節／1 個目標 | 取不到 | — | — |
| 66 | Cause Serious Wounds | cleric | 4 | 7 節／1 個目標 | observed | partial | observed |
| 67 | Neutralize Poison | cleric | 4 | 7 節／自己 | 紮營法術 | — | — |
| 68 | Poison | cleric | 4 | 7 節／1 個目標／豁免 paralysis-poison-death | observed | missing | observed |
| 69 | Protection Evil, 10' Radius | cleric | 4 | 7 節／1 個目標／效果 45 | observed | partial | observed |
| 70 | Sticks to Snakes | cleric | 4 | 7 節／1 個目標／效果 3 | observed | partial | observed |
| 71 | Cure Critical Wounds | cleric | 5 | 8 節／1 個目標 | observed | partial | observed |
| 72 | Cause Critical Wounds | cleric | 5 | 8 節／1 個目標／豁免 spell | observed | partial | observed |
| 73 | Dispel Evil | cleric | 5 | 8 節／自己／效果 145 | observed | partial | observed |
| 74 | Flame Strike | cleric | 5 | 8 節／1 個目標／豁免 spell | observed | partial | observed |
| 75 | Raise Dead | cleric | 5 | 10 節／自己 | 紮營法術 | — | — |
| 76 | Slay Living | cleric | 5 | 10 節／1 個目標／豁免 spell | observed | missing | observed |
| 77 | Detect Magic | druid | 1 | 3 節／自己／效果 5 | 未宣告 | — | — |
| 78 | Entangle | druid | 1 | 3 節／半徑 3／豁免 spell／效果 136 | 未宣告 | — | — |
| 79 | Faerie Fire | druid | 1 | 3 節／逐一點選／豁免 spell／效果 7 | 未宣告 | — | — |
| 80 | Invisibility to Animals | druid | 1 | 4 節／1 個目標／效果 69 | 未宣告 | — | — |
| 81 | Charm Monsters | magic-user | 4 | 4 節／逐一點選／豁免 spell／效果 11 | observed | partial | observed |
| 82 | Confusion | magic-user | 4 | 4 節／半徑 3／豁免 spell／效果 35 | observed | partial | observed |
| 83 | Dimension Door | magic-user | 4 | 1 節／半徑 0 | observed | missing | observed |
| 84 | Fear | magic-user | 4 | 4 節／半徑 0／豁免 spell／效果 142 | observed | partial | observed |
| 85 | Fire Shield | magic-user | 4 | 4 節／自己 | observed | missing | observed |
| 86 | Fumble | magic-user | 4 | 4 節／1 個目標／豁免 spell／效果 27 | observed | partial | observed |
| 87 | Ice Storm | magic-user | 4 | 4 節／半徑 2 | observed | partial | observed |
| 88 | Minor Globe Of Invulnerability | magic-user | 4 | 4 節／自己／效果 63 | observed | partial | observed |
| 89 | Remove Curse | magic-user | 4 | 4 節／1 個目標 | observed | missing | observed |
| 90 | Animate Dead |  | 5 | 5 節／自己／效果 32 | 未宣告 | — | — |
| 91 | Cloud Kill | magic-user | 5 | 5 節／半徑 1 | observed | observed | observed |
| 92 | Cone of Cold | magic-user | 5 | 5 節／半徑 0／豁免 spell | observed | missing | observed |
| 93 | Feeblemind | magic-user | 5 | 5 節／1 個目標／豁免 spell／效果 68 | observed | partial | observed |
| 94 | Hold Monsters | magic-user | 5 | 5 節／4 個目標／豁免 spell／效果 52 | observed | partial | observed |
| 95 | （占位） | — | 6 | 10 節／自己／效果 73 | 取不到 | — | — |
| 96 | （占位） | — | 6 | 10 節／自己／效果 109 | 取不到 | — | — |
| 97 | （占位） | — | 6 | 0 節／自己／效果 25 | 取不到 | — | — |
| 98 | （占位） | — | 6 | 0 節／半徑 3／豁免 spell | 取不到 | — | — |
| 99 | （占位） | — | 6 | 0 節／自己 | 取不到 | — | — |
| 100 | Bestow Curse | magic-user | 4 | 4 節／1 個目標／豁免 spell | observed | missing | observed |
