# 532：開場與荒野選項的 JSON／引擎資料邊界

狀態：`READY`

## 目的

第 460 輪已將 ECL 選項建立為 game-pack `option_rules`，但開場、讀檔、紮營
返回與世界地圖預覽仍有三個選項直接呼叫 `catalog.Text("enter_city" …)`，
而 `initializeECL`／`Select` 還保留針對來源 token 的專用分支。這會讓 JSON
修改後，產品路徑與測試容易各自形成第二份文字真相。

本輪只修正資料邊界，不新增劇情、地圖、秘密門或規則語意。

## 已證實的資料契約

CoAB game-pack 已存在下列穩定規則，且 `en`／`zh-TW` 均有對應 locale：

| option rule ID | 原始來源 token | message ID |
|---|---|---|
| `ecl-option.enter-city` | `ENTER CITY` | `enter_city` |
| `ecl-option.journey-on` | `JOURNEY ON` | `journey_on` |
| `ecl-option.camp` | `CAMP` | `camp` |

`State` 現在統一先呼叫 engine 的 `Pack.LocalizeOption`。只有在沒有可用
game-pack rule 的資料中立／相容情境，才將來源 token 正規化成 locale key，
不建立 CoAB 專用中文 fallback。未知 token 仍原樣保留，方便發現資料缺口。

來源 token 是原版 ECL 的身份資料，不是玩家顯示文字；本輪沒有改寫原始 bytes
或把 token 命名成新的劇情規則。

## 接通範圍

- ECL 初始化選單與荒野邊界的 `ActionStart` fallback。
- 選擇 `ENTER CITY`／`JOURNEY ON`／`CAMP` 後的訊息。
- 角色建立完成、DOS／自訂存檔載入、隊伍檔載入後的荒野選單。
- 紮營返回與世界地圖預覽的相同三項選單。
- `TestLocalizedOpeningFlow` 以 stable `ecl-option.enter-city` 從實際 pack
  解析期望值；測試不再複製目前的繁中文字串。

## 驗證與證據等級

- `gamepack` option rule／locale 內容與 engine lookup：`exact`，因為資料檔、
  schema 與 resolver 直接閉合。
- 產品路徑移除專用 switch、受影響 Go 測試：`exact`，Docker 內執行
  `go test -modfile=workplace/coab-test.mod ./gamepack ./internal/game`，
  兩套件均通過。
- 本輪沒有把這項測試擴大成完整開場到結局、完整翻譯或完整戰鬥證據。

## 邊界與未完成

本輪沒有縮減 P0-1 DOS 外部地圖 handoff、P0-2 PC-98 MOVEPARTY／秘密門、
P0-3 後續 external routine 的缺口，也沒有新增 `(13,10)`→`(8,15)` 的正式
movement 規則。完整 ECL、全地圖、戰鬥演出、音效、存檔、全量翻譯與三平台
發行仍未完成。
