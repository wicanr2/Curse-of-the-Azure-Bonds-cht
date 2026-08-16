# 譯名一致性稽核

由 `cmd/glossary-audit` 產生，不要手改。表在 [`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)。

## 掃描範圍

| 檔案 | 字串數 |
|---|---:|
| `gamepack/pack/20-locale.zh-TW.json` | 675 |
| `assets/locale/zh-TW.json` | 870 |
| `internal/tooltext/messages/zh-TW.json` | 170 |

## 結果

| 項目 | 數量 |
|---|---:|
| 詞條 | 92 |
| **不一致** | **0** |

## 逐詞條

### 人物

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Akabar` | 阿卡巴 | 6 | — |
| `Akabar Bel Akash` | 阿卡巴・貝爾・阿卡什 | 3 | — |
| `Alaterian` | 阿拉特里安 | 1 | — |
| `Alias` | 愛麗雅絲 | 16 | — |
| `Azoun` | 阿祖恩 | 2 | — |
| `Belinda` | 貝琳達 | 1 | — |
| `Daemir` | 黛米爾 | 2 | — |
| `Dexam` | 德克薩姆 | 11 | — |
| `Dimswart` | 迪姆斯沃特 | 10 | — |
| `Dracandros` | 德拉坎德羅斯 | 18 | 德拉坎卓斯 |
| `Dragonbait` | 龍餌 | 12 | — |
| `Elminster` | 伊爾明斯特 | 3 | — |
| `Filani` | 菲拉妮 | 5 | — |
| `Flamed One` | 烈焰之主 | 9 | 烈焰者、燃燒者 |
| `Fzoul Chembryl` | 弗佐爾・錢布瑞爾 | 5 | 弗佐爾．錢布瑞爾 |
| `Gharri` | 加里 | 3 | 嘉莉 |
| `Imperceptor` | 總督察長 | 3 | — |
| `Kith` | 姬絲 | 1 | — |
| `Lilian` | 莉莉安 | 1 | — |
| `Mogion` | 摩貢 | 10 | — |
| `Myrixelets` | 米里塞勒特斯 | 1 | — |
| `Nacacia` | 娜卡西亞 | 5 | — |
| `Nameless One` | 無名者 | 3 | — |
| `Olive Ruskettle` | 奧莉芙・魯斯克特爾 | 3 | 拉斯凱托 |
| `Silk` | 絲綢 | 2 | — |
| `Sis` | 西絲 | 1 | — |
| `Tyranthraxus` | 提朗瑟克斯 | 16 | 泰蘭索斯 |
| `Vangerdahast` | 凡格達海斯 | 1 | — |

### 地名

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Ashabenford` | 阿沙本福德 | 6 | — |
| `Cormyr` | 科米爾 | 3 | — |
| `Dagger Falls` | 匕首瀑布 | 4 | — |
| `Dalelands` | 谷地諸邦 | 2 | — |
| `Dark Shrine` | 幽暗神殿 | 2 | — |
| `Essembra` | 艾森布拉 | 7 | — |
| `Great Glacier` | 大冰川 | 1 | — |
| `Hap` | 哈普 | 11 | — |
| `Haptooth` | 哈普圖斯 | 3 | — |
| `Hillsfar` | 希爾斯法 | 7 | — |
| `Mulmaster` | 穆爾馬斯特 | 2 | — |
| `Myth Drannor` | 迷斯卓諾 | 12 | — |
| `Pit of Moander` | 摩安德之坑 | 5 | — |
| `Shadowdale` | 暗影谷 | 11 | — |
| `Standing Stones` | 立石群 | 4 | — |
| `Teshwave` | 特什維夫 | 2 | — |
| `Tilverton` | 提爾佛頓 | 16 | — |
| `Yulash` | 尤拉什 | 14 | — |
| `Zhentil Keep` | 散提爾堡 | 20 | 散塔林堡 |

### 怪物（由 combatant_name_rules 匯入）

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `BEHOLDER` | 眼魔 | 3 | — |
| `BIT O' MOANDER` | 摩安德殘軀 | 2 | — |
| `BLACK DRAGON` | 黑龍 | 13 | — |
| `CULTIST` | 摩安德教徒 | 4 | — |
| `DARK ELF CLERIC` | 黑暗精靈牧師 | 1 | — |
| `DARK ELF MAGE` | 黑暗精靈法師 | 1 | — |
| `DK ELF FIGHTER` | 黑暗精靈戰士 | 1 | — |
| `DRACOLICH` | 龍巫妖 | 1 | — |
| `EFREETI` | 伊弗利特 | 4 | — |
| `FIGHTER` | 戰士 | 26 | — |
| `HIGH PRIEST` | 大祭司 | 4 | — |
| `HIPPOGRIFF` | 鷹馬 | 1 | — |
| `HOODED MEDUSA` | 兜帽梅杜莎 | 1 | — |
| `MINOTAUR` | 牛頭人 | 4 | — |
| `MOGION` | 摩貢 | 10 | — |
| `SALAMANDER` | 火蜥蜴 | 4 | — |
| `SHAMBLING MOUND` | 蔓生怪 | 4 | — |
| `ZHENTIL CLERIC` | 散提爾堡牧師 | 1 | — |
| `ZHENTIL FIGHTER` | 散提爾堡戰士 | 1 | — |
| `ZHENTIL MAGE` | 散提爾堡法師 | 1 | — |
| `ZHENTRIM CLERIC` | 散塔林牧師 | 1 | — |
| `ZHENTRIM FGHTR` | 散塔林戰士 | 2 | — |
| `ZHENTRIM MAGE` | 散塔林法師 | 1 | — |

### 敘述中出現的怪物

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Displacer Beast` | 移位獸 | 1 | — |
| `Otyugh` | 奧提尤格 | 2 | — |
| `Thri-Kreen` | 斯瑞克林 | 1 | — |
| `bugbear` | 熊地精 | 1 | — |
| `phase spider` | 相位蜘蛛 | 1 | — |
| `warg` | 座狼 | 1 | — |

### 神器與關鍵物品

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Amulet of Lathander` | 洛山達護符 | 5 | — |
| `Gauntlet of Moander` | 摩安德護手 | 3 | — |
| `Helm of Dragons` | 龍盔 | 5 | 龍之頭盔 |
| `Pool of Radiance` | 光芒之池 | 7 | — |
| `azure bonds` | 枷印 | 46 | 蔚藍枷、青色印記 |

### 神祇

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Bane` | 班恩 | 7 | 貝恩 |
| `Gond` | 剛德 | 5 | — |
| `Lathander` | 洛山達 | 5 | — |
| `Moander` | 摩安德 | 25 | — |

### 組織

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Black Network` | 黑網 | 2 | — |
| `Fire Knives` | 火刀 | 28 | 火焰匕首 |
| `Harper` | 豎琴手 | 3 | — |
| `Red Plume` | 紅羽衛 | 11 | 紅羽戰士 |
| `Red Wizards of Thay` | 塞爾紅袍法師 | 1 | — |
| `Swanmays` | 天鵝少女團 | 1 | — |
| `Zhentarim` | 散塔林會 | 4 | — |
