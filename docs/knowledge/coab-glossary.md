# 《青色枷的詛咒》譯名表

一個原文專有名詞只有一種繁中寫法。這條規則由 `cmd/glossary-audit` 強制：
禁用寫法出現在任何繁中字串裡就紅，表中的繁中寫法在資料裡一次都沒出現也紅
（避免表變成沒人用的空殼）。掃描範圍是 `gamepack/pack/20-locale.zh-TW.json`、
`assets/locale/zh-TW.json`、`internal/tooltext/messages/zh-TW.json`。

**怪物名不寫在這裡**。它們的唯一來源是 game pack 的 `combatant_name_rules`
（spec 479），稽核時自動併入報告，不需要人工同步兩份。

**法術與物品名不寫在這裡**。它們在 `assets/locale/zh-TW.json` 各自只有一個鍵
（`spell_*` 58 個、`item_*` 140 個），單一來源本來就漂不掉；只有同時出現在劇情
敘述裡的關鍵物品才進表（龍盔、摩安德護手、洛山達護符）。

## 通用寫法

| 項目 | 規則 |
|---|---|
| 人名的姓名分隔 | 用 `・`（U+30FB）。不要用 `．`、`·`、`-` |
| 稱號與本名 | 本名優先；稱號只在原文用稱號時才用（`Tyranthraxus, the Flamed One` ＝ 烈焰之主提朗瑟克斯） |
| 原文自己有兩個名字 | 分別給不同的繁中，不要合併。`Hap` 與 `Haptooth` 是兩個詞條 |
| 原文沒有的專名 | 不要自己加。原文寫 `THIS TOWN` 就譯「這座城鎮」，不要補上地名 |

## 人物

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Tyranthraxus` | 提朗瑟克斯 | — | 稱號見下一列 |
| `Flamed One` | 烈焰之主 | 烈焰者、燃燒者 | 提朗瑟克斯的稱號 |
| `Dracandros` | 德拉坎德羅斯 | 德拉坎卓斯 | 塞爾紅袍法師 |
| `Fzoul Chembryl` | 弗佐爾・錢布瑞爾 | 弗佐爾．錢布瑞爾 | |
| `Mogion` | 摩貢 | — | 摩安德大祭司 |
| `Dexam` | 德克薩姆 | — | |
| `Alias` | 愛麗雅絲 | — | |
| `Dragonbait` | 龍餌 | — | 蜥蜴人聖武士 |
| `Olive Ruskettle` | 奧莉芙・魯斯克特爾 | 拉斯凱托 | |
| `Dimswart` | 迪姆斯沃特 | — | 賢者 |
| `Elminster` | 伊爾明斯特 | — | |
| `Filani` | 菲拉妮 | — | 提爾佛頓賢者 |
| `Nacacia` | 娜卡西亞 | — | 公主 |
| `Akabar` | 阿卡巴 | — | |
| `Azoun` | 阿祖恩 | — | 國王 |
| `Vangerdahast` | 凡格達海斯 | — | 宮廷法師 |
| `Daemir` | 黛米爾 | — | 葬林幽魂公主 |
| `Nameless One` | 無名者 | — | |
| `Imperceptor` | 總督察長 | — | 穆爾馬斯特的班恩教階職稱 |

## 神祇

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Bane` | 班恩 | 貝恩 | `Banite` 併入本詞條 |
| `Moander` | 摩安德 | — | |
| `Lathander` | 洛山達 | — | |

## 組織

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Fire Knives` | 火刀 | — | 單複數同譯 |
| `Zhentarim` | 散塔林會 | — | 原文另作 `Zhentrim` |
| `Black Network` | 黑網 | — | 散塔林會的別名，只在原文用別名時使用 |
| `Red Wizards of Thay` | 塞爾紅袍法師 | — | |

## 地名

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Tilverton` | 提爾佛頓 | — | |
| `Zhentil Keep` | 散提爾堡 | 散塔林堡 | 城；與組織「散塔林會」是兩回事 |
| `Yulash` | 尤拉什 | — | |
| `Myth Drannor` | 迷斯卓諾 | — | |
| `Shadowdale` | 暗影谷 | — | |
| `Hillsfar` | 希爾斯法 | — | |
| `Essembra` | 艾森布拉 | — | |
| `Ashabenford` | 阿沙本福德 | — | |
| `Mulmaster` | 穆爾馬斯特 | — | |
| `Hap` | 哈普 | — | 世界地圖的目的地名 |
| `Haptooth` | 哈普圖斯 | — | 巫師塔下的村莊 |
| `Pit of Moander` | 摩安德之坑 | — | |
| `Dark Shrine` | 幽暗神殿 | — | |
| `Standing Stones` | 立石群 | — | |

## 神器與關鍵物品

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Helm of Dragons` | 龍盔 | 龍之頭盔 | |
| `Gauntlet of Moander` | 摩安德護手 | — | |
| `Amulet of Lathander` | 洛山達護符 | — | |
| `Pool of Radiance` | 光芒之池 | — | |
| `azure bonds` | 枷印 | — | 完整敘述時作「蔚藍枷」 |
