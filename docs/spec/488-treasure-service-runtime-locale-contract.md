# 第 488 輪：財寶服務執行期文字資料契約

狀態：`READY`

## 範圍與既有證據

本輪不改 TREASURE opcode、金錢換算、ITEM block、隨機物品、戰後延遲或 ECL
continuation。這些仍由既有原始 bytes 與正常玩家路徑規格約束；本輪只移除
「財寶服務邊界 → 財寶列表 → 指定角色收取／取消／略過」的 Go 中文副本。

## 資料契約

- 正式 locale 提供 `treasure_prompt`、`treasure_exit`、`treasure_ready`、
  `treasure_take_prompt`、`treasure_cancel`、`treasure_taken`、
  `treasure_skipped` 與 `treasure_assets_pending`。
- State 保存 `TREASURE_ITEM_n`、`TREASURE_CHARACTER_n`、`TREASURE_CANCEL`、
  `TREASURE_EXIT` 等控制 identity；翻譯文字不參與選擇與 continuation。
- 缺 ITEM 素材的錯誤物件是 runtime format argument，不是翻譯內容的一部分。
- 物品名稱仍由正式 item catalog 解析；本輪沒有另造財寶物品名稱表。

## 驗證

- 正式 catalog coverage 驗證八個 stable ID 均存在。
- 財寶 menu 測試覆蓋列表、角色選擇、取消、收取、略過與裝備 mutation；所有
  期望文字都在執行時從正式 catalog 解析。
- 火刀辦公室真實 ECL／正常地城路徑改用正式 catalog，並驗證財寶 prompt 與
  exit choice；原 3,000 gold、3 gems、2 jewelry、ITEM block `82h` 及搜尋
  一次性行為不變。
- 聚焦 `./internal/game` 與 Docker／Xvfb／`--network none` 正式全套 gate
  通過；漢字稽核 `173→164`，移除九筆 `runtime_ui_debt`。其餘 164 筆與
  完整遊戲仍未完成。
