# 455：酒館、飲品、傳聞與提爾佛頓事件資料分離

狀態：`READY`

## 範圍

本輪同時處理兩種酒館入口：作品中立的 BAR 傳聞服務，以及提爾佛頓原始 ECL
酒館事件。BAR 選單、六則預設傳聞、無新傳聞與離場訊息；ECL 的酒保、四種
飲品、特別客人、紫帶女子、建築側邊騷動、華麗匕首及手札 17 提示，均由
`assets/locale/zh-TW.json` stable IDs 取得。

`localizeOption`、`localizePrompt`、`localizeECLText` 與 `localizeECLLine` 仍
負責以原始英文 token／fragment 找出 ID，但不再把繁中譯文當 Go fallback。
這保留原 bytes 身分與 continuation，同時讓譯文只有一個正式資料來源。

## 測試與玩家路徑

- coverage 測試驗證 BAR、飲品、Tavern Tales、手札 17 與提爾佛頓 ECL 共
  41 個 IDs。
- BAR 單元測試從正式 locale 解析「聽傳聞」選項，傳聞內容仍可由場所注入。
- 提爾佛頓 terrain `88h` 正常路徑驗證 PICTURE／HEAD／BODY、動作選單、飲品、
  YES／NO、三段事件 continuation、遊戲內手札 17 解鎖及返回 `(6,10)`。
- Ashabenford Tale 28 與 Essembra Tale 60 的既有真實 ECL 路徑改用正式 locale，
  防止缺 key 時只顯示 stable ID。

## 可量測結果

exact baseline 從 1,223 降為 1,169 occurrences，共移除 54 次：

| 分類 | 第 454 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 409 | 375 | -34 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 650 | 630 | -20 |
| 合計 | 1,223 | 1,169 | -54 |

這不代表全遊戲酒館事件、所有 Tavern Tales 或飲酒規則已完整還原；只證明上述
已實作玩家路徑的繁中顯示不再依賴 Go 中文 literal。
