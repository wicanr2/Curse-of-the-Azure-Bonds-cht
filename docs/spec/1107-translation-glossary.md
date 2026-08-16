# 1107 — 譯名一致性：一個專有名詞一種寫法

- 狀態：`READY`
- 證據等級：`exact`（九組不一致由 `internal/glossary` 在實際字串上量到，逐筆列於 §二）
- 表在 [`docs/knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)，
  稽核在 `internal/glossary` ＋ `cmd/glossary-audit`

## 一、為什麼需要一道閘

繁中字串散在三份 JSON：game pack 的 `20-locale.zh-TW.json`（645 條劇情與事件）、
UI 的 `assets/locale/zh-TW.json`（870 條選單與物品法術）、工具訊息
`internal/tooltext/messages/zh-TW.json`。三者之間沒有任何機制要求同一個人名寫成
同一個樣子，而內容產出（`ENG-01`）還會讓 game pack 那份成長數倍。

★ **實測不是零**：建表之前掃過一次，九組同名異譯，其中五組出現在手札長文裡——
那正是玩家會反覆重讀的地方。

## ★★★ 二、實際量到的九組不一致與處置

| # | 原文 | 原有寫法 | 定案 | 依據 |
|---:|---|---|---|---|
| 1 | `Bane` | 貝恩（5）／班恩（3） | **班恩** | 繁中《被遺忘的國度》慣用寫法；不採多數決 |
| 2 | `Zhentil Keep` | 散提爾堡（15）／散塔林堡（2） | **散提爾堡** | 散塔林（Zhentarim）是**組織**，散提爾堡是**城**，兩者不可混用 |
| 3 | `Dracandros` | 德拉坎德羅斯（15）／德拉坎卓斯（1） | **德拉坎德羅斯** | |
| 4 | `Flamed One` | 烈焰之主（7）／烈焰者（1）／燃燒者（1） | **烈焰之主** | |
| 5 | `Helm of Dragons` | 龍盔（3）／龍之頭盔（1） | **龍盔** | |
| 6 | `Olive Ruskettle` | 魯斯克特爾（2）／拉斯凱托（1） | **奧莉芙・魯斯克特爾** | 原作訪客簿寫 `O.RUSKETTLE` |
| 7 | 姓名分隔符 | `・`（4）／`．`（1） | **`・`（U+30FB）** | |
| 8 | `Zhentrim` | 散塔林會／黑網 | **散塔林會**；`Black Network` 才譯黑網 | 原文 `journal.38.3` 寫 `The Zhentrim`，別名只在 `38.2` 出現 |
| 9 | `Haptooth` | 哈普圖斯（2）／哈普（1 處誤用） | **哈普圖斯**；`Hap` 才是哈普 | 原作兩個詞都存在，不合併 |

★ 另修一處**原文沒有的專名**：`area5.depart-akabar-reluctant` 的原作文字是
`MUST FREE THIS TOWN`，繁中卻寫「解放哈普村」，英文釋義也寫成 `free Haptooth`。
兩邊都改回「這座城鎮」。⇒ **原文沒有的地名不要自己補**，補了就會製造出一組
無法用原文裁決的不一致。

## 三、閘的形狀

`internal/glossary.Run` 讀表 ＋ 掃三份繁中目錄，任何一條成立就紅：

| 代碼 | 條件 |
|---|---|
| `forbidden_variant` | 表列的禁用寫法出現在任何繁中字串裡 |
| `unused_term` | 表列的繁中寫法在資料裡一次都沒出現（防止表變成空殼） |
| `conflicting_rendering` | 同一個原文有兩種譯名（大寫折疊後比對） |
| `duplicate_rendering` | 同一個繁中寫法對到兩個不同原文 |
| `unsatisfiable_ban` | 禁用寫法是某個詞條譯名的一部分 |

★★★ **怪物名不進表**。它們的唯一來源是 game pack 的 `combatant_name_rules`
（spec 479），稽核時**自動匯入**成詞條。所以 `MOGION` 同時是人工詞條與
combatant 規則不算重複，而是跨檔交叉檢查的著力點——兩邊的譯名不一樣就
`conflicting_rendering`。人工同步兩份表的做法一開始就不要有。

★ **法術與物品名不進表**：`assets/locale/zh-TW.json` 裡 `spell_*` 58 個、
`item_*` 140 個，每個名字只有一個鍵，單一來源漂不掉。只有同時出現在劇情敘述裡的
關鍵物品（龍盔、摩安德護手、洛山達護符）才需要詞條。

### 三之一、禁用寫法的限制

閘只擋得住「這個字串在任何地方都不該出現」的變體。`Haptooth` 的誤用是
**哈普**——但那正是 `Hap` 的正確譯名，禁不得。`unsatisfiable_ban` 就是用來擋住
「把哈普加進禁用清單」這種會誤報的寫法。

⇒ **這一類要靠逐句對照原文，閘擋不住**。表的〈通用寫法〉一節把判準寫成規則
（原文有兩個名字就給兩個詞條、原文沒有的專名不要補），而不是寫成事件敘述。

## 四、明確不宣稱

- 沒有宣稱 68 個詞條就是全部專有名詞。表是從現有 645 條 game pack 訊息的英文
  釋義抽大寫詞、再逐個回頭核對繁中建起來的；`ENG-01` 每補一個區域都可能有新名字。
- 沒有宣稱既有譯名都貼近原作語感。本輪只解決**一致性**，沒有重譯任何名字
  （第 1 組換寫法是為了對齊領域慣例，不是語感偏好）。
- 沒有掃 `docs/`。手冊校訂稿（`docs/manual/journal/`）與攻略還沒納入閘，
  它們搬進 game pack 時才會被檢查。

## 五、回歸

| 測試 | 釘住什麼 |
|---|---|
| `glossary.TestGlossaryHasNoDrift` | 五種不一致一個都不存在 |
| `glossary.TestGlossaryTableIsWellFormed` | 表的每一列都完整，禁用寫法不自我矛盾 |
| `glossary.TestGlossaryScansEveryChineseCatalog` | 三份繁中目錄都在掃描範圍內，且都不是空的 |
| `glossary.TestCombatantNamesAreImported` | 怪物名的匯入路徑還在（斷掉會讓交叉檢查靜默失效） |

閘會紅已實測：把 `journal.7.2` 的班恩改回貝恩，`forbidden_variant` 立刻指出檔案與鍵。
