# 《青色枷的詛咒》譯名表

一個原文專有名詞只有一種繁中寫法。少數詞條在不同語境下有兩種都對的寫法，
用「／」列在同一格，第一個是正規寫法。這條規則由 `cmd/glossary-audit` 強制：
禁用寫法出現在任何繁中字串裡就紅，表中的繁中寫法在資料裡一次都沒出現也紅
（避免表變成沒人用的空殼）。掃描範圍是 `gamepack/pack/20-locale.zh-TW.json`、
`assets/locale/zh-TW.json`、`internal/tooltext/messages/zh-TW.json`。

**上場戰鬥的怪物名不寫在這裡**。它們的唯一來源是 game pack 的 `combatant_name_rules`
（spec 479），稽核時自動併入報告，不需要人工同步兩份。只在敘述裡出現、沒有 combatant
規則的怪物才進表（見最後一節）；哪天它們也上場了，兩邊寫法不一致會報
`conflicting_rendering`。

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
| `Tyranthraxus` | 提朗瑟克斯 | 泰蘭索斯 | 稱號見下一列 |
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
| `Gharri` | 加里 | 嘉莉 | 剛德的祭司 |
| `Akabar` | 阿卡巴 | — | |
| `Azoun` | 阿祖恩 | — | 國王 |
| `Vangerdahast` | 凡格達海斯 | — | 宮廷法師 |
| `Daemir` | 黛米爾 | — | 葬林幽魂公主 |
| `Nameless One` | 無名者 | — | |
| `Silk` | 絲綢 | — | 熔岩洞的黑暗精靈首領，自稱的化名 |
| `Myrixelets` | 米里塞勒特斯 | — | 手札 21 那封家書的署名 |
| `Sis` | 西絲 | — | 米里塞勒特斯家中的人 |
| `Kith` | 姬絲 | — | 手札 44 失蹤的天鵝少女 |
| `Belinda` | 貝琳達 | — | 同上 |
| `Akabar Bel Akash` | 阿卡巴・貝爾・阿卡什 | — | `Akabar` 的全名，兩個詞條並存；原文另作 `AKABAR BEL AKAS`（戰鬥員名截短） |
| `Alaterian` | 阿拉特里安 | — | 手札 21 那封信的收信人 |
| `Lilian` | 莉莉安 | — | 阿拉特里安的妻子 |
| `Tarsus` | 塔瑟斯 | — | 黑暗精靈洞穴裡承接傳送生意的人 |
| `Imperceptor` | 總督察長 | — | 穆爾馬斯特的班恩教階職稱 |

## 神祇

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Bane` | 班恩 | 貝恩 | `Banite` 併入本詞條 |
| `Moander` | 摩安德 | — | |
| `Lathander` | 洛山達 | — | |
| `Gond` | 剛德 | — | |

## 組織

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Fire Knives` | 火刀 | 火焰匕首 | 單複數同譯；原文另作 `FIRE KNIFE` |
| `Zhentarim` | 散塔林會 | — | 原文另作 `Zhentrim` |
| `Black Network` | 黑網 | — | 散塔林會的別名，只在原文用別名時使用 |
| `Red Plume` | 紅羽衛 | 紅羽戰士 | 希爾斯法軍隊；尤拉什圍城的一方 |
| `Red Wizards of Thay` | 塞爾紅袍法師 | — | |
| `Swanmays` | 天鵝少女團 | — | |
| `Harper` | 豎琴手 | — | 英文單複數同譯 |

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
| `Great Glacier` | 大冰川 | — | |
| `Dagger Falls` | 匕首瀑布 | — | |
| `Teshwave` | 特什維夫 | — | 以世界地圖的地名標籤為準 |
| `Dalelands` | 谷地諸邦 | — | |
| `Cormyr` | 科米爾 | — | |

## 神器與關鍵物品

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Helm of Dragons` | 龍盔 | 龍之頭盔 | |
| `Gauntlet of Moander` | 摩安德護手 | — | |
| `Amulet of Lathander` | 洛山達護符 | — | |
| `Pool of Radiance` | 光芒之池 | — | |
| `azure bonds` | 枷印／青色枷 | 蔚藍枷、青色印記 | 平常寫「枷印」；需要完整名稱時寫「青色枷」，沿用原中文版譯名（《青色枷的詛咒》） |

## 敘述中出現的怪物

沒有 `combatant_name_rules` 的怪物名放這裡。哪天它們真的上場，combatant 規則的譯名
與這裡不一致就會報 `conflicting_rendering`。

| 原文 | 繁中 | 禁用寫法 | 備註 |
|---|---|---|---|
| `Otyugh` | 奧提尤格 | — | 世界地圖遭遇與手札 21 |
| `Thri-Kreen` | 斯瑞克林 | — | 手札 45 |
| `phase spider` | 相位蜘蛛 | — | 手札 45 |
| `Displacer Beast` | 移位獸 | — | 手札 58 |
| `bugbear` | 熊地精 | — | 手札 36 |
| `warg` | 座狼 | — | 手札 36；原文另作 `WORG` |

## 例外：譯文可以不出現的名字

英文釋義提到名字、繁中卻沒有對應寫法時，稽核會報 `missing_rendering`。真的不該出現
的情況列在這裡，其他一律視為漏譯或誤譯。清單也會反向檢查：條件消失卻還留著的例外
會報 `stale_exception`。

| 訊息 ID | 原文 | 理由 |
|---|---|---|
| `wizard-tower.dracandros.freezes-party` | `Dracandros` | 譯文採原作的直接引語，沒有第三人稱敘述句 |
| `journal.38.3` | `Filani` | 承前一段用「她」，重複姓名反而不自然 |
