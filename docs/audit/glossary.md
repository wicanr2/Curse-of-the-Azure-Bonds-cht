# 譯名一致性稽核

由 `cmd/glossary-audit` 產生，不要手改。表在 [`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)。

## 掃描範圍

| 檔案 | 字串數 |
|---|---:|
| `gamepack/pack/20-locale.zh-TW.json` | 1580 |
| `assets/locale/zh-TW.json` | 1142 |
| `internal/tooltext/messages/zh-TW.json` | 170 |

## 結果

| 項目 | 數量 |
|---|---:|
| 詞條 | 139 |
| **不一致** | **0** |

## 逐詞條

### 人物

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Akabar` | 阿卡巴 | 10 | — |
| `Akabar Bel Akash` | 阿卡巴・貝爾・阿卡什 | 4 | — |
| `Alaterian` | 阿拉特里安 | 1 | — |
| `Alias` | 愛麗雅絲 | 41 | — |
| `Azoun` | 阿祖恩 | 5 | — |
| `Belinda` | 貝琳達 | 1 | — |
| `Daemir` | 黛米爾 | 2 | — |
| `Dexam` | 德克薩姆 | 11 | — |
| `Dimswart` | 迪姆斯沃特 | 16 | — |
| `Dracandros` | 德拉坎德羅斯 | 20 | 德拉坎卓斯 |
| `Dragonbait` | 龍餌 | 29 | — |
| `Elminster` | 伊爾明斯特 | 5 | — |
| `Filani` | 菲拉妮 | 8 | — |
| `Flamed One` | 烈焰之主 | 11 | 烈焰者、燃燒者 |
| `Fzoul Chembryl` | 弗佐爾・錢布瑞爾 | 6 | 弗佐爾．錢布瑞爾 |
| `Gharri` | 加里 | 7 | 嘉莉 |
| `Imperceptor` | 總督察長 | 4 | — |
| `Kith` | 姬絲 | 1 | — |
| `Lilian` | 莉莉安 | 1 | — |
| `Mogion` | 摩貢 | 15 | — |
| `Myrixelets` | 米里塞勒特斯 | 1 | — |
| `Nacacia` | 娜卡西亞 | 9 | — |
| `Nameless One` | 無名者 | 3 | — |
| `Olive Ruskettle` | 奧莉芙・魯斯克特爾 | 4 | 拉斯凱托 |
| `Silk` | 絲綢 | 7 | — |
| `Sis` | 西絲 | 1 | — |
| `Tarsus` | 塔瑟斯 | 1 | — |
| `Tyranthraxus` | 提朗瑟克斯 | 21 | 泰蘭索斯 |
| `Vangerdahast` | 凡格達海斯 | 2 | — |

### 地名

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Ashabenford` | 阿沙本福德 | 8 | — |
| `Cormyr` | 科米爾 | 9 | — |
| `Dagger Falls` | 匕首瀑布 | 12 | — |
| `Dalelands` | 谷地諸邦 | 3 | — |
| `Dark Shrine` | 幽暗神殿 | 3 | — |
| `Essembra` | 艾森布拉 | 7 | — |
| `Great Glacier` | 大冰川 | 1 | — |
| `Hap` | 哈普 | 15 | — |
| `Haptooth` | 哈普圖斯 | 4 | — |
| `Hillsfar` | 希爾斯法 | 10 | — |
| `Mulmaster` | 穆爾馬斯特 | 6 | — |
| `Myth Drannor` | 迷斯卓諾 | 18 | — |
| `Pit of Moander` | 摩安德之坑 | 5 | — |
| `Shadowdale` | 暗影谷 | 19 | — |
| `Standing Stones` | 立石群 | 8 | — |
| `Teshwave` | 特什維夫 | 13 | — |
| `Tilverton` | 提爾佛頓 | 23 | — |
| `Voonlar` | 溫拉爾 | 10 | 弗恩拉、沃恩拉 |
| `Yulash` | 尤拉什 | 23 | — |
| `Zhentil Keep` | 散提爾堡 | 46 | 散塔林堡 |

### 怪物（由 combatant_name_rules 匯入）

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `AKABAR BEL AKAS` | 阿卡巴・貝爾・阿卡什 | 4 | — |
| `ALIAS` | 愛麗雅絲 | 41 | — |
| `ANHKHEG` | 掘地蟲 | 1 | — |
| `BAR PATRON` | 酒館客人 | 1 | — |
| `BEHOLDER` | 眼魔 | 17 | — |
| `BIT O' MOANDER` | 摩安德殘軀 | 2 | — |
| `BLACK DRAGON` | 黑龍 | 18 | — |
| `BUGBEAR` | 熊地精 | 4 | — |
| `CENTAUR` | 半人馬 | 5 | — |
| `CROCODILE` | 鱷魚 | 4 | — |
| `CULTIST` | 摩安德教徒 | 6 | — |
| `CYNTHIA` | 辛西亞 | 1 | — |
| `DARK ELF CLERIC` | 黑暗精靈牧師 | 2 | — |
| `DARK ELF LORD` | 黑暗精靈領主 | 1 | — |
| `DARK ELF MAGE` | 黑暗精靈法師 | 1 | — |
| `DISPLACER BEAST` | 移位獸 | 3 | — |
| `DK ELF FIGHTER` | 黑暗精靈戰士 | 1 | — |
| `DRACANDROS` | 德拉坎德羅斯 | 20 | — |
| `DRACOLICH` | 龍巫妖 | 2 | — |
| `DRAGONBAIT` | 龍餌 | 29 | — |
| `EFREETI` | 伊弗利特 | 10 | — |
| `ETTIN` | 雙頭巨魔 | 1 | — |
| `FIGHTER` | 戰士 | 34 | — |
| `FIGHTING DOG` | 鬥犬 | 1 | — |
| `FIRE KNIFE` | 火刀 | 40 | — |
| `GIANT SLUG` | 巨鼻涕蟲 | 1 | — |
| `GIANT SPIDER` | 巨蜘蛛 | 1 | — |
| `GRENDEL` | 格倫戴爾 | 1 | — |
| `GRIFFON` | 獅鷲 | 4 | — |
| `HELL HOUND` | 地獄犬 | 9 | — |
| `HIGH PRIEST` | 大祭司 | 6 | — |
| `HIPPOGRIFF` | 鷹馬 | 1 | — |
| `HOODED MEDUSA` | 兜帽梅杜莎 | 1 | — |
| `KNIGHT` | 騎士 | 8 | — |
| `LG VEGEPYGMY` | 大型菌人 | 1 | — |
| `LIZARD MAN` | 蜥蜴人 | 4 | — |
| `LOOTER` | 掠奪者 | 3 | — |
| `MAGE` | 法師 | 57 | — |
| `MANTICORE` | 蠍尾獅 | 3 | — |
| `MARGOYLE` | 大石像鬼 | 1 | — |
| `MINOTAUR` | 牛頭人 | 5 | — |
| `MOGION` | 摩貢 | 15 | — |
| `MONKEY` | 猴子 | 3 | — |
| `NEO-OTYUGH` | 新奧提尤格 | 1 | — |
| `OGRE` | 食人魔 | 5 | — |
| `OTYUGH` | 奧提尤格 | 13 | — |
| `OWL BEAR` | 梟熊 | 5 | — |
| `PHASE SPIDER` | 相位蜘蛛 | 2 | — |
| `PRIEST OF BANE` | 班恩祭司 | 2 | — |
| `RAKSHASA` | 羅剎 | 17 | — |
| `RED PLUME` | 紅羽衛 | 31 | — |
| `ROYAL GUARD` | 皇家衛兵 | 8 | — |
| `RUSTLE` | 拉梭 | 1 | — |
| `SALAMANDER` | 火蜥蜴 | 6 | — |
| `SHAMBLING MOUND` | 蔓生怪 | 16 | — |
| `SM VEGEPYGMY` | 小型菌人 | 1 | — |
| `THIEF` | 盜賊 | 24 | — |
| `THRI-KREEN` | 斯瑞克林 | 3 | — |
| `TROLL` | 巨魔 | 8 | — |
| `TYRANTHRAXUS` | 提朗瑟克斯 | 21 | — |
| `WORG` | 座狼 | 4 | — |
| `WYVERN` | 雙足飛龍 | 1 | — |
| `ZHENTIL CLERIC` | 散提爾堡牧師 | 1 | — |
| `ZHENTIL FIGHTER` | 散提爾堡戰士 | 1 | — |
| `ZHENTIL MAGE` | 散提爾堡法師 | 1 | — |
| `ZHENTRIM CLERIC` | 散塔林牧師 | 1 | — |
| `ZHENTRIM FGHTR` | 散塔林戰士 | 2 | — |
| `ZHENTRIM MAGE` | 散塔林法師 | 1 | — |

### 敘述中出現的怪物

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Displacer Beast` | 移位獸 | 3 | — |
| `Otyugh` | 奧提尤格 | 13 | — |
| `Thri-Kreen` | 斯瑞克林 | 3 | — |
| `bugbear` | 熊地精 | 4 | — |
| `phase spider` | 相位蜘蛛 | 2 | — |
| `warg` | 座狼 | 4 | — |

### 神器與關鍵物品

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Amulet of Lathander` | 洛山達護符 | 6 | — |
| `Gauntlet of Moander` | 摩安德護手 | 4 | — |
| `Helm of Dragons` | 龍盔 | 5 | 龍之頭盔 |
| `Pool of Radiance` | 光芒之池 | 8 | — |
| `azure bonds` | 枷印 | 51 | 蔚藍枷、青色印記 |

### 神祇

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Bane` | 班恩 | 24 | 貝恩 |
| `Gond` | 剛德 | 9 | — |
| `Lathander` | 洛山達 | 6 | — |
| `Moander` | 摩安德 | 38 | — |

### 組織

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Black Network` | 黑網 | 2 | — |
| `Fire Knives` | 火刀 | 40 | 火焰匕首 |
| `Harper` | 豎琴手 | 3 | — |
| `Red Plume` | 紅羽衛 | 31 | 紅羽戰士 |
| `Red Wizards of Thay` | 塞爾紅袍法師 | 1 | — |
| `Swanmays` | 天鵝少女團 | 2 | — |
| `Zhentarim` | 散塔林會 | 18 | — |
