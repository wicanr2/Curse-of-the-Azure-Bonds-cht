# 1220：locale drift 稽核跟上分割 pack 與 tooltext 分層

狀態：`READY`（2026-08-27）

## 問題

`cmd/locale-drift-audit` 仍寫死已移除的
`gamepack/events/pit-of-moander.json`，因此對目前四份分割 pack 完全跑不起來。
改為讀分割 pack 後又揭出第二個過期假陽性：第 714 輪已把開發者文字移入
`internal/tooltext`，但 AST 掃描仍把 `tooltext.Text("h…")` 當成玩家 UI key。

## 修正

- 依 `gamepack/pack/*.json` 檔名順序讀入全部部件。
- 合併 `locales` 與 `option_rules`；同語系重複 key 立即失敗。
- 在每個部件的原始 JSON 上遞迴搜尋 `message_id`，不限於某一份內容檔。
- AST 玩家 UI 掃描明確排除 `tooltext.Text`；`h.*` 屬開發者／工具 catalog，
  不得塞進 `assets/locale/zh-TW.json`。
- 測試 fixture 改用分割 pack，並新增「`tooltext` 不污染玩家 locale」回歸。

## 當前結果

- UI catalog：1,162 鍵；靜態產品引用 459；缺鍵 0。
- game-pack locale：英文與繁中各 1,594 鍵；`message_id` 引用 1,495；
  英中內容完全相同 0。
- 產品 Go 檔 48 份；動態 `Text` 呼叫 53；違規 0。
- `cmd/glossary-audit -check`：139 條詞條，未解決不一致 0。

## 證據邊界

UI 的 703 個靜態孤立 key 與 game-pack 的 99 個孤立 key 不是刪除清單：
物品、法術、手札、音樂與建角類別大量透過動態組鍵或專用 catalog 消費。
本 gate 證明得了分割 pack 對稱、靜態引用不缺鍵、英中沒有整句照抄；
證明不了譯文語意、語氣、版面與原文逐句正確。
