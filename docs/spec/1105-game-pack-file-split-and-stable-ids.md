# 1105 — game pack 分檔與 stable ID：先依種類切，區域切留給命名收斂之後

- 狀態：`READY`
- 證據等級：`exact`（現況數字現場量；引擎 schema 逐欄位列過）
- 對應 `ENG-02`（分檔）、`ENG-03`（stable ID 與 schema 版本）、`CHT-04`（熱鍵，見 §六）
- 前提決策：使用者 2026-08-16 定案「game pack 以分檔方式處理」

## 一、為什麼現在要切

`gamepack/events/pit-of-moander.json` 是單一 261 KB 檔。內容產出（`ENG-01`）會讓它
成長數倍，而它現在的形狀有三個具體問題：

| 問題 | 現況量測 |
|---|---|
| 翻譯與規則混在同一檔 | `locales` 佔 **55%**（134 KB），`text_rules` 佔 25%，其餘每項 ≤4% |
| 兩語系在同一個物件裡 | `en` 607 ＝ `zh-TW` 607，要看「這一輪翻了什麼」得在同一檔內對照 |
| 一次 review 的 diff 涵蓋整個作品 | 改一條提爾佛頓的文字，diff 落在與眼魔洞穴同一個檔 |

## ★★★ 二、切法：依**種類**，不是依區域

直覺的切法是依區域（`myth-drannor`、`tilverton`…），因為 `text_rules` 的 ID 已經
帶區域前綴。**實測否決了這個方案**：ID 命名空間目前不一致。

| 資料 | 命名 | 數量 |
|---|---|---|
| `text_rules` | 點號區域前綴（`wizard-tower.courtyard.entering`） | 369，22 個群組 |
| `option_rules` | 大多是 `ecl-option.*`（83）與 `option.*`（7），只有少數帶區域 | 113 |
| `events` | 點號前綴 | 4 |
| `locales` 的鍵 | **兩種慣例並存**：點號命名空間（`myth-drannor.*` 148、`journal.*` 48、`combatant.*` 23、`creation.*` 15、`music.*` 12）**加上約 100 個沒有命名空間的扁平鍵**（`tavern_whiskey`、`parlay_sly`、`spell_cleric_1`、`enter_city`…） |

⇒ 依區域切會要求先把那約 100 個扁平鍵改名並搬家。**那是另一件事，而且是不可逆的
大範圍改名**；把它綁在分檔上，等於讓一個機械動作卡在一個需要逐條判斷的決定上。

所以本規格切的是**種類**：

```
gamepack/pack/
    00-core.json        schema 標頭、presentation、search、character_creation、
                        combat_*、music_*、maps、combatant_name_rules
    10-content.json     text_rules、option_rules、events
    20-locale.en.json    locales.en
    20-locale.zh-TW.json locales.zh-TW
```

得到的三件事，都不需要改任何 ID：

- 翻譯的 diff 與規則的 diff 分開；每個語系一個檔，兩語系對齊變成**檔案層級**的比對。
- 內容（`10-content.json`）與機制資料（`00-core.json`）分開，符合 `AGENTS.md` §2 的界線。
- 檔名前綴的數字決定合併順序，順序固定 ⇒ 產物可重生。

⚠ **這是第一步，不是終點。** `10-content.json` 之後還要依區域再切；那一步的前提是
§四的命名收斂完成。本規格刻意不做，因為現在做會夾帶一次大改名。

## 三、合併契約（引擎側）

`engine.LoadPackParts(parts [][]byte) (*Pack, error)` 依序合併後**驗證一次**。
逐欄位規則：

| 欄位種類 | 規則 |
|---|---|
| 標頭（`$schema`／`schema_version`／`id`／`default_locale`） | 只能由一個 part 提供非零值；兩個 part 給了不同值即失敗 |
| 單例指標（`search`／`presentation`／`character_creation`） | 最多一個 part 提供；重複即失敗 |
| 清單（`text_rules`／`option_rules`／`events`／`maps`／`combat_*`／`music_*`／`combatant_name_rules`） | 依 part 順序串接 |
| `locales` | 逐語系合併；**同一個語系裡出現重複的鍵即失敗** |

★ **重複一律失敗，不做「後蓋前」。** 靜默覆蓋會讓「這條字串到底是哪一檔的」變成
要靠合併順序推理的問題，而那正是分檔要消除的東西。

★ 合併**之後**才驗證：現有 `Validate` 有跨欄位檢查（`message_id` 要有對應 locale
條目、`option_rule` 的 ID 與 source 不得重複…），逐檔驗會把合法的跨檔引用判成缺漏。

★ 單檔載入的 `LoadPack`／`LoadPackBytes` 不變，仍是 `DisallowUnknownFields`。
分檔只是多一個入口，不改既有契約。

## 四、stable ID：先立規範，既有的進 baseline

目標命名（新增一律照這個）：

```
<群組>.<地點或子系統>.<東西>
```

- 群組用連字號（`myth-drannor`、`fire-knife`、`wizard-tower`）。
- 全小寫，分隔用 `.`，群組內部用 `-`；**不用底線**。
- `message_id` 與引用它的規則 ID 共用同一個命名空間。

既有的約 100 個扁平鍵（`tavern_whiskey` 這一類）**不在本輪改名**。處置沿用本專案
已經驗證過的做法（`go-han-literals-baseline.json`）：把它們列進一份版本化 baseline，
新增不符合規範的鍵會讓測試變紅，既有的不變紅。

⇒ 「還有多少鍵沒收斂」因此隨時查得到，而且不會再增加。

## 五、本規格不宣稱

- 沒有宣稱區域分檔的最終切法；那要等命名收斂後另立規格。
- 沒有宣稱那約 100 個扁平鍵各自該歸哪個群組。
- 沒有宣稱 `schema_version` 需要升版：欄位語意一個都沒改，只是同一份資料換成多檔
  承載，合併後的 `Pack` 與合併前逐欄位相同（有測試釘住）。

## ★★★ 六、`CHT-04` 熱鍵：問題不在 `option_rule`

先前把「熱鍵沒有欄位」記在 `option_rule` 上。**那個位置是錯的**：`option_rule` 對應
的是原版 ECL 選單 token（`"ATTACK DRAGONS"`），而 ECL 選單是**按索引**選的
（`selections []uint16`），沒有字母熱鍵。

真正的缺口在**指令列**：

- 原版 DOS 沒有熱鍵表，`overlay-26:003D4h` 把玩家按的鍵大寫化後**逐字元掃整條
  選項行**，撞到相同字元就選那一項（spec 1060）。所以熱鍵必須是選項文字裡真的
  出現的大寫字母。
- remake 沒有沿用那條掃描（它不掃字串），但也沒有把繫結資料化：畫面用 locale
  字串顯示中文（`combat_menu_main` ＝「移動　查看　瞄準　使用　施法　快速　結束」），
  按鍵卻是散在前端的英文首字母常數（`ebiten.KeyM`／`KeyL`／`KeyC`／`KeyQ`…）。
- ⇒ **翻譯之後，畫面上看不出要按哪個鍵**，而且標籤與繫結沒有任何地方對得起來。

處置方向（本輪不做，避免在沒有消費端的情況下先加欄位）：把「指令 ID → 按鍵 →
`message_id`」做成一張表，繪製端與輸入端讀同一張，畫面顯示 `按鍵 ＋ 標籤`。
⚠ 在有消費端之前不要先往引擎 schema 加欄位——未被使用的欄位會變成猜測。

## 七、驗證

| 測試 | 釘住什麼 |
|---|---|
| `engine.TestLoadPackPartsMatchesSingleFile` | 分檔合併出來的 `Pack` 與原本的單檔逐欄位相同 |
| `engine.TestLoadPackPartsRejectsDuplicateLocaleKey` | 同語系重複鍵失敗，不是後蓋前 |
| `engine.TestLoadPackPartsRejectsConflictingHeader` | 兩個 part 給不同 `id`／`schema_version` 即失敗 |
| `engine.TestLoadPackPartsRejectsDuplicateSingleton` | `search`／`presentation`／`character_creation` 重複即失敗 |
| `gamepack.TestDefaultPackMatchesCommittedParts` | CoAB 的 `Default()` 走分檔路徑，且結果與合併前相同 |
| `gamepack.TestLocaleKeyNamingBaselineIsExact` | 新增不符合 §四規範的鍵會紅；既有扁平鍵在 baseline 內 |
