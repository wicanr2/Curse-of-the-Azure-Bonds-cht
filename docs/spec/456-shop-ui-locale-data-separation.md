# 456：商店服務玩家介面資料分離

狀態：`READY`

## 範圍

商店主選單、購買、查看、取出／集中／分配金幣、估價、販售、鑑定、角色與
物品選擇、價格／報價格式、錯誤與完成訊息，均改由正式 locale stable IDs
解析。Go 保留商品 `ItemRecord`、價格、typed coins、報價狀態與交易規則，
不再保存這批繁中 UI 文案。

三個場所用到的 STORE 標籤、提爾佛頓旅店指向商店的 ECL fragment，以及小型
魔法商店文字 fallback 也改為 stable ID，避免同一商店概念仍從旁路硬編碼。

## 測試與玩家路徑

- coverage 測試驗證 63 個商店 UI／格式／錯誤 stable IDs。
- 所有商店 UI 測試載入正式 `zh-TW.json`，選單與格式期望在執行時由 ID 取得。
- 商品價格、角色摘要、取款金額、估價、確認、鑑定與查看結果不再複製顯示字串。
- 真實 Weaponers of Cormyr 路徑從 GEO2 terrain `84h`、原始 PICTURE／ECL
  service boundary 進入商店，驗證主選單與商品 prompt 的 stable ID，購買後
  續跑並返回 `(2,12)`。

## 可量測結果

exact baseline 從 1,169 降為 1,100 occurrences，共移除 69 次：

| 分類 | 第 455 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 375 | 372 | -3 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 630 | 564 | -66 |
| 合計 | 1,169 | 1,100 | -69 |

## 明確未完成

`monster.ChineseName` 的物品 base name／name-number 組合仍是 Go 中文 catalog；
本輪只完成商店 UI 與格式資料分離，不能宣稱物品名稱已 JSON 化。後續必須以
stable item type／name-number IDs 搬移，並讓商店、裝備、戰利品與測試共用
同一份 game-pack／locale 真相來源。
