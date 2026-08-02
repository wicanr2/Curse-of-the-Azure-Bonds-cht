# 459：角色建立模板與前端文字資料分離

狀態：`READY`

## 證據與範圍

第 249、250、251、259、261 與 278 輪已分別保存 `Gbl.RaceClasses`、半獸人、
起始年齡、DOS raw class／八槽 levels、18 個多職組合及正常新遊戲 ECL1
dispatch 證據。本輪不重新猜角色規則，而是把已證明的 22 個單職與 18 個多職
模板從 Go 搬入 CoAB game-pack JSON。

獨立 engine `f3c652a` 新增作品中立的 `character_creation.templates` schema：

- stable `id／display_id`；
- opaque `race_id／primary_class_id／raw_class_id`；
- 初始 level、固定八槽 class levels、固定六項 base abilities；
- duplicate、跨 locale、shape、level 與能力值範圍 validation。

Engine 不解讀 CoAB enum 或年齡表。CoAB adapter 才把 raw IDs 投影到
`party.Race／Class`、驗證 `StartingAgeSpecFor` 與 `Character.Validate`。任一模板
錯誤會阻止角色建立開啟，不再像舊 slice 一樣靜默 `continue` 而改變選項數。

## 前端與文字

模板預設名稱由 game-pack `display_id` 解析；建立標題、姓名輸入提示、能力名稱、
種族、單職名稱、選項格式、進度與快捷鍵說明由正式 `zh-TW.json` stable IDs
解析。Ebiten `drawCreation` 不再保存繁中表，測試也不複製模板譯名。

## 玩家路徑與驗證

- engine 全套 `go test ./...` 通過，schema 的 locale completeness 與 8／6 shape
  failure cases 有 regression。
- CoAB game-pack 驗證 40 筆模板、首末 raw IDs、levels、abilities 與 en／zh-TW
  display coverage。
- 正常 party-less production start 從 title 進角色建立、選第一筆 JSON 模板、
  完成建立並進 ECL1 global block `01h`；不是 direct-entry debug party。
- 建立角色的 remake party save/load round-trip 保留由 pack 解析出的名稱。

## 可量測結果

Go 漢字 literal exact baseline 由 953 降為 867 occurrences，共移除 86 次：

| 分類 | 第 458 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 372 | 372 | 0 |
| Ebiten 前端 UI | 164 | 135 | -29 |
| runtime／工具 UI | 417 | 360 | -57 |
| 合計 | 953 | 867 | -86 |

## 明確未完成

本輪沒有宣稱原版完整建角 UX 已完成。性別、alignment、exceptional strength、
原版擲骰／調整流程、裝備購買、dual-class、所有多職規則與 DOS 原生角色檔建立
仍是獨立缺口；現行畫面也仍須在發行前 UI 稽核中與 DOS／PC-98 原版重新比對。
