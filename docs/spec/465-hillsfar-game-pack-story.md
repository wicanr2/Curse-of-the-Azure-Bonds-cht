# 第四百六十五輪：希爾斯法城市與酒館事件資料化

狀態：`READY`

## 範圍

本輪把希爾斯法五個作品事件由 Go fallback 移入 CoAB game-pack：

- `hillsfar.fire-knives-ambush`
- `hillsfar.edge`
- `hillsfar.places`
- `hillsfar.dockside-bar`
- `hillsfar.red-plumes-spill-drinks`

每條規則保存原始 ECL source fragments 與 en／zh-TW 訊息；State 不再知道
希爾斯法、火刀、碼頭酒館或紅羽衛的中文內容。

## 正常玩家路徑

既有長回歸由 Standing Stone 正常 Journey 至 Hillsfar，依序驗證：

1. trail 上六名偽裝戰士的火刀伏擊與戰後抵達 `hillsfar.edge`。
2. 進城顯示 `hillsfar.places`，並播放 PC-98 城鎮 selector `06h`。
3. 進入 dockside bar，選擇休息後觸發 Red Plumes 打翻飲品的 YES／NO 分支。
4. 選 NO 後迎戰六名敵人；勝利返回同一酒館選單。
5. 離開酒館回 places，不重播城鎮音樂；離城回 edge 並播放荒野 selector `05h`。

## 驗證與邊界

- 五條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶手札頁。
- 完整 `gamepack`／`internal/game` 回歸通過，玩家路徑訊息由 stable ID 取得。
- Go 漢字 literal exact baseline 由 655 降至 651；`localization_debt`
  160→156，frontend 135、runtime 360 不變。dockside bar 舊分支只有 locale key，
  因此五個事件只減少四個漢字 literal。

本輪不代表 Hillsfar 所有酒類、傳聞、服務、紅羽衛分支、敵人能力、音效或整個
城市已完整還原。
