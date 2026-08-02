# 第四百六十二輪：開場敘事資料分離

狀態：`READY`

## 範圍與來源

本輪把兩組開場來源移出 `localizeECLLine`：十一行遊戲概要，以及 ECL1
block `01h` 新遊戲醒來事件的五行原文。來源身分仍是原始英文片段；CoAB
game-pack 新增 en／zh-TW 訊息，State 只呼叫中立 `MatchText`。

新遊戲實際有兩個 boundary，不能合成一個畫面：

1. `YOU AWAKEN...` 與 `ALL YOUR GEAR IS GONE...` 在 PRESS pause 顯示。
2. 選擇繼續後，PICTURE boundary 顯示 `ADDING TO YOUR DISQUIET...` 至
   `IDENTICALLY MARKED.`。

因此資料採 `opening.new-game-awakening` 與 `opening.new-game-marks` 兩個
stable IDs。`opening.curse-summary` 保存另一組十一行開場概要；目前由 raw
source oracle 與 pack coverage 證明，尚未把它擴大宣稱為正常 title 玩家路徑。

## 同輪清理

第 461 輪的 15 個手札規則會優先於舊 `localizeECLText／Line`。本輪移除其中
已被新規則完整遮蔽的 12 組段落 fallback 與 10 組逐行 fragment，並刪除不再
使用的 UI locale 複本。這些文字仍存在於 CoAB game-pack，不是刪除翻譯。

## 驗證與邊界

- 三條 opening rules 在 en／zh-TW 均回傳指定 rule ID 與非空訊息。
- title→角色建立→ECL1 block `01h` 正常路徑逐一驗證兩個新遊戲 boundary，
  最後仍返回提爾佛頓 `(7,13,W)`。
- 第 461 輪涉及的提爾佛頓、尤拉什、火刀與摩安德之坑整合測試全數通過，
  證明刪除 shadowed fallback 未改變事件結果。
- Go 漢字 literal exact baseline 由 722 降至 688；`localization_debt`
  227→193，frontend 135、runtime 360 不變。

本輪不代表所有開場動畫、ECL 劇情或翻譯已完成。剩餘 fallback 必須依真正
pause／picture／menu／combat boundary 分批資料化，不能只依原文換行切割。
