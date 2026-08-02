# 第 478 輪：手札穩定 ID 與存檔 v10

狀態：`READY`

## 問題與證據

- `NewState` 曾預先建立八頁開發用中文手札；它們不是原始 ECL 解鎖結果，會在
  新遊戲提前揭露劇情，違反第 461 輪確立的「事件發生後才可重讀」契約。
- `TextRule.journal_message_ids` 已是 game-pack 的權威身分，但 Engine
  `TextResult` 只回傳翻譯後文字；State 因此只能以中文頁首去重。
- remake 存檔 v9 沒有手札欄位。玩家讀檔後會遺失已解鎖條目；若直接保存中文
  全文，日後修訂翻譯或切換語系又會保存過期內容。

## READY 契約

1. Engine `TextResult` 同時回傳 `JournalMessageIDs` 與目前語系的
   `JournalPages`；前者是身分，後者只供當下顯示。
2. 新遊戲手札為空，只顯示正式 locale 的空白提示；不得建立摘要、攻略或開發
   說明假頁面。
3. ECL 事件以 `journal_message_ids` 精確去重及解鎖；不得以中文內容或 prefix
   判斷同一條目。
4. remake save v10 只保存有序、非空、不可重複的 message ID。讀檔時必須從
   目前 game-pack 與目前 locale 重新解析全文；未知 ID 失敗即關閉。
5. v1–v9 舊存檔沒有手札狀態，讀取後維持空手札；不得補回八頁假資料。
6. 頁碼與操作提示由 locale stable ID 提供；空手札不得顯示 `1 / 0`。

## 驗證

- Engine `go test ./...` 通過，並驗證泛用 pack 同時回傳 `journal` ID 與繁中頁面。
- CoAB 真實 ECL 新遊戲事件只解鎖 `journal.31`，其 UI index 為 0；菲拉妮事件
  接續增加 `journal.38.1..3`，不再依賴八頁假資料偏移。
- save v10 單元測試拒絕重複 ID，也拒絕把 v10 欄位偽裝成 v9。
- 跨語系 round-trip 先以繁中保存 `journal.31`，再以英文 State 讀取；ID 保持
  不變，頁面由正式英文 game-pack 重新解析，不沿用存檔時的中文。
- Go 漢字稽核由 378 降為 368：frontend 135→133、runtime 220→212、
  localization 23 不變。

本輪不改畫面幾何或素材，因此不新增 README 截圖；後續原版 UI 稽核仍須另外
逐畫面比對邊框、HEAD／BODY 舞台、第一人稱 viewport 與對話內框。
