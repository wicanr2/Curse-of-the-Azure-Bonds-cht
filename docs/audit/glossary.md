# 譯名一致性稽核

由 `cmd/glossary-audit` 產生，不要手改。表在 [`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)。

## 掃描範圍

| 檔案 | 字串數 |
|---|---:|
| `gamepack/pack/20-locale.zh-TW.json` | 645 |
| `assets/locale/zh-TW.json` | 870 |
| `internal/tooltext/messages/zh-TW.json` | 170 |

## 結果

| 項目 | 數量 |
|---|---:|
| 詞條 | 68 |
| **不一致** | **0** |

## 逐詞條

### 人物

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Akabar` | 阿卡巴 | 5 | — |
| `Alias` | 愛麗雅絲 | 16 | — |
| `Azoun` | 阿祖恩 | 1 | — |
| `Daemir` | 黛米爾 | 2 | — |
| `Dexam` | 德克薩姆 | 11 | — |
| `Dimswart` | 迪姆斯沃特 | 10 | — |
| `Dracandros` | 德拉坎德羅斯 | 15 | 德拉坎卓斯 |
| `Dragonbait` | 龍餌 | 12 | — |
| `Elminster` | 伊爾明斯特 | 3 | — |
| `Filani` | 菲拉妮 | 5 | — |
| `Flamed One` | 烈焰之主 | 7 | 烈焰者、燃燒者 |
| `Fzoul Chembryl` | 弗佐爾・錢布瑞爾 | 5 | 弗佐爾．錢布瑞爾 |
| `Imperceptor` | 總督察長 | 3 | — |
| `Mogion` | 摩貢 | 10 | — |
| `Nacacia` | 娜卡西亞 | 4 | — |
| `Nameless One` | 無名者 | 3 | — |
| `Olive Ruskettle` | 奧莉芙・魯斯克特爾 | 3 | 拉斯凱托 |
| `Tyranthraxus` | 提朗瑟克斯 | 15 | — |
| `Vangerdahast` | 凡格達海斯 | 1 | — |

### 地名

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Ashabenford` | 阿沙本福德 | 6 | — |
| `Dark Shrine` | 幽暗神殿 | 2 | — |
| `Essembra` | 艾森布拉 | 6 | — |
| `Hap` | 哈普 | 9 | — |
| `Haptooth` | 哈普圖斯 | 3 | — |
| `Hillsfar` | 希爾斯法 | 7 | — |
| `Mulmaster` | 穆爾馬斯特 | 2 | — |
| `Myth Drannor` | 迷斯卓諾 | 9 | — |
| `Pit of Moander` | 摩安德之坑 | 4 | — |
| `Shadowdale` | 暗影谷 | 11 | — |
| `Standing Stones` | 立石群 | 3 | — |
| `Tilverton` | 提爾佛頓 | 11 | — |
| `Yulash` | 尤拉什 | 14 | — |
| `Zhentil Keep` | 散提爾堡 | 17 | 散塔林堡 |

### 怪物（由 combatant_name_rules 匯入）

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `BEHOLDER` | 眼魔 | 3 | — |
| `BIT O' MOANDER` | 摩安德殘軀 | 2 | — |
| `BLACK DRAGON` | 黑龍 | 11 | — |
| `CULTIST` | 摩安德教徒 | 4 | — |
| `DARK ELF CLERIC` | 黑暗精靈牧師 | 1 | — |
| `DARK ELF MAGE` | 黑暗精靈法師 | 1 | — |
| `DK ELF FIGHTER` | 黑暗精靈戰士 | 1 | — |
| `DRACOLICH` | 龍巫妖 | 1 | — |
| `EFREETI` | 伊弗利特 | 4 | — |
| `FIGHTER` | 戰士 | 29 | — |
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

### 神器與關鍵物品

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Amulet of Lathander` | 洛山達護符 | 4 | — |
| `Gauntlet of Moander` | 摩安德護手 | 2 | — |
| `Helm of Dragons` | 龍盔 | 4 | 龍之頭盔 |
| `Pool of Radiance` | 光芒之池 | 7 | — |
| `azure bonds` | 枷印 | 42 | — |

### 神祇

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Bane` | 班恩 | 6 | 貝恩 |
| `Lathander` | 洛山達 | 4 | — |
| `Moander` | 摩安德 | 23 | — |

### 組織

| 原文 | 繁中 | 出現次數 | 禁用寫法 |
|---|---|---:|---|
| `Black Network` | 黑網 | 1 | — |
| `Fire Knives` | 火刀 | 25 | — |
| `Red Wizards of Thay` | 塞爾紅袍法師 | 1 | — |
| `Zhentarim` | 散塔林會 | 3 | — |
