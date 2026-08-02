# 457：物品名稱與修飾詞資料分離

狀態：`READY`

## 範圍

`ItemRecord.Type` 與三個 `NameNumbers` 仍是 DOS 原始格式的 typed fields；繁中
base name、name-number 修飾詞、未知物品格式、加值、詛咒與堆疊數量格式則移入
正式 `assets/locale/zh-TW.json`。stable key 採十六進位原始 ID：
`item_type_XX` 與 `item_name_XX`。

`monster.LocalizedItemName` 只負責作品中立的組字規則：由第三個名稱欄往第一個
名稱欄解析、遵守隱藏旗標、避免重複 base token，再套用加值、詛咒與弩矢數量。
未知 type 保留可見診斷；未知 name-number 不猜譯並略過。

## 消費端與測試

- 商店購買／販售／鑑定、角色裝備、戰利品與物品選單均使用 State 的同一份
  locale catalog。
- `cmd/azure-bonds` 的 ITEMS／MON*ITM 診斷新增 `-locale`，不再呼叫中文專用
  Go catalog。
- `internal/monster` 測試用合成 resolver 驗證 typed composition，不複製 CoAB
  中文 catalog；遊戲層測試由正式 JSON 動態取得物品顯示名稱。

## 可量測結果

Go 漢字 literal exact baseline 由 1,100 降為 974 occurrences，共移除 126 次：

| 分類 | 第 456 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 372 | 372 | 0 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 564 | 438 | -126 |
| 合計 | 1,100 | 974 | -126 |

## 明確未完成

`monster.ChineseAffectName` 仍保存效果名稱與未知效果格式的中文 Go catalog；它是
下一批獨立資料分離債務。本輪只證明物品名稱資料化，不代表所有物品效果、法術
說明、裝備規則或全遊戲玩家文字都已完成。
