# 第五百二十九輪：玩家戰鬥法術資料契約

狀態：`READY`（限目前已實作的玩家施法入口與 JSON／engine 契約）
日期：2026-08-10

## 目的

本輪把玩家在戰鬥中選擇法術時所需的「法術身份、施法者類別、目標交易、
施法時間與顯示訊息」從 CoAB `State` 的固定支援清單移到 Golden Box
engine 的 `combat_player_spells` 與 CoAB game pack。效果本身仍由既有
CoAB adapter／Battle helper 執行；本規格不宣稱新增或完整還原 AD&D 法術規則。

## 契約

每筆 `combat_player_spells` 必須包含：

| 欄位 | 層級 | 意義 |
|---|---|---|
| `id` | game pack stable ID | 作品內可長期引用的法術資料身份，不是顯示名稱 |
| `spell_id` | 原始 spell table identity | 保留 DOS／PC-98 原始數字；不可用中文名稱取代 |
| `caster_class` | title-owned token | 作品的施法者類別投影；engine 不猜 AD&D 類別語意 |
| `target_mode` | engine vocabulary | `none`、`enemy`、`party_member`、`area_point`、`line` |
| `behavior` | title-owned behavior token | 交給作品 adapter 選擇已有的效果 helper；未知值 fail-closed |
| `message_id` | cross-locale stable ID | 由 pack 的 `en`／`zh-TW` locale 提供顯示名稱 |
| `casting_time` | 原始／pack timing value | 施法延遲交易使用；若同時有 Quick AI metadata，兩者必須一致 |

`combat_ai_spells` 仍只負責 Quick priority、`cast_on`、`min_range` 與同一原始
法術的 AI metadata；它不是玩家 CAST 的顯示或目標契約。engine 會檢查同一
`spell_id` 的 `casting_time` 不可在兩份資料中漂移。

## CoAB 本輪資料

目前已接入原本已有 runtime helper 與回歸測試的 12 個法術：

- Cleric：`1／2／3／4／6／7`（祝福、詛咒、治療／造成輕傷、防護邪惡／善良）。
- Magic-user：`15／21／34／47／51／91`（魔法飛彈、睡眠術、惡臭之雲、火球術、
  閃電束、死雲術）。

這份清單只代表目前 remake adapter 已有的 12 個效果入口，不代表 CoAB 的
全法術表已完成；未宣告法術會在 CAST／pending spell 入口拒絕執行。

## 證據與推論等級

- `exact`（remake contract）：engine `Pack.Validate` 的欄位、stable ID／
  `spell_id` 唯一性、locale completeness、target mode enum、AI／player
  casting time 一致性，以及 `FindCombatPlayerSpell` 查詢都有測試。
- `exact`（CoAB integration）：玩家 CAST、Quick CAST、延遲法術、區域／直線
  目標移動與訊息 label 都從該 pack 查詢；受影響的 `gamepack`、`internal/ecl`、
  `internal/game` 測試在 Docker 通過。
- `strong inference`（資料映射）：`behavior` token 對應既有 CoAB helper，
  是目前 remake 對已觀察效果的 adapter 映射；它不是單靠 token 就證明 DOS／
  PC-98 的完整公式、亂數、動畫或音效。
- `unknown`：尚未接入的法術、完整法術記憶／消耗規則、所有原版 spell record
  projection、逐幀演出與硬體音效時序。

## 驗證規則

產品測試只能以 `id`、`spell_id`、`message_id` 與當次 locale resolver 取得
期望值，不得把「魔法飛彈」等目前文字複製成 Go 測試常數。改翻譯只應影響
內容校對；資料繫結、目標模式與施法交易測試不能因此失效。

本輪正式 gate：

```text
Docker: go test -mod=mod ./gamepack ./internal/ecl ./internal/game
Docker: go test ./...                 (golden-box-remake-engine)
```

root 驗證使用暫時 local `replace` 指向本地 engine 工作樹，沒有修改 root 的
模組邊界；提交前會把 root `go.mod` 鎖到已推送的 engine commit。

## 不在本輪宣稱的內容

本契約不會把 JSON 宣告數量寫成完整可通關、不會把 `behavior` 名稱當成原版
反組譯的完整證明，也不會取代 P0-1／P0-2／P0-3 地圖 handoff、完整 ECL、全量
戰鬥 AI、全翻譯、音樂音效與三平台 release 工作。下一個逆向工作仍以
`docs/knowledge/golden-box-reverse-engineering-worklist.md` 的正常玩家阻塞點為準。
