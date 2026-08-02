# 第四百六十一輪：遊戲內手札資料分離

狀態：`READY`

## 範圍與證據

本輪處理 `State.unlockJournalEntries` 原有的 15 組作品專用判斷。觸發來源仍是
原始 ECL 顯示片段；英文頁面以使用者提供的 Adventurer's Journal PDF OCR
協助轉錄，繁中頁面沿用既有遊戲內摘要並改由 stable ID 保存。OCR 只作定位與
轉錄輔助，不取代 PDF 與原始 ECL 的來源地位。

## 資料 contract

- CoAB `text_rules` 保存 `all_contains`、事件 `message_id` 與有序
  `journal_message_ids`。
- en／zh-TW locale 保存事件提示與 23 頁手札；runtime 只把匹配結果附加到
  遊戲內手札，不知道條目編號、劇情人物或中文內容。
- `journal.3.1` 這類 ID 是可編輯顯示資料的穩定身分；測試由目前 game-pack
  解析期望頁面，不複製繁中句子。
- 不把手冊全部預先解鎖。只有 ECL 事件實際觸發時，對應頁面才可在遊戲內重讀。

## 驗證

- 15 個觸發在 en／zh-TW 均命中指定 rule ID、事件訊息與正確頁數。
- 提爾佛頓旅店 31、菲拉妮 38、酒館匕首 17、高階祭司 19、公會地圖 4、
  火刀定身房 26、辦公室 9／29、王室 53／54、尤拉什 22／52、愛麗雅絲 3
  與摩安德神殿地圖 20 的既有真實 ECL 整合路徑均讀取 game-pack 頁面。
- Go 漢字 literal exact baseline 由 797 降至 722：`localization_debt`
  302→227，frontend 135、runtime 360 不變。

## 完成邊界

這證明上述 15 個既有觸發與 23 頁已資料化，不代表原手冊全部條目、所有地圖
圖像或全部 ECL 劇情文字已完成。剩餘 `localizeECLText／localizeECLLine`
fallback 仍要逐條以原始 bytes／玩家路徑驗證後移入 game-pack。
