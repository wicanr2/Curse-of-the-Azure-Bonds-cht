# 1237 — 世界目的地與 ECL 記憶體交易同步

狀態：`READY`（目的地選擇同步 `4C9B`／`4C9C`；連續路線回歸已建立）

## 問題與證據

無強化按鍵路徑在世界旅行後形成可操作但不前進的城內／城外循環。逐幀記錄補上
`Location`、`OriginalLocation` 與 `Area.CurrentCity` 後，確認 Go 狀態已選中
阿沙本福德（native value `2`），但 ECL 仍顯示上一站匕首瀑布的旅行方式、城邊
訊息與六項服務。

最小重現是讓 `4C9B` 保留匕首瀑布的 `3`，再由暗影谷目的地選單選
`ASHABENFORD`。舊實作只在後續 `TRAIL`／`ROAD`／`RIVER`／`WILDERNESS` 選擇
前寫 `4C9C`；目的地選擇本身沒有在 ECL resume 前提交 `4C9B`，因此狀態投影與
原版分派器讀到不同目的地。

## 修正

- 玩家選中 game-pack 宣告的世界目的地時，在同一次 `Select` 交易、ECL resume
  之前同步 `4C9B` 與 `4C9C`。
- ECL 路線可能重用 `4C9B`，resume 後仍依既有契約再次提交目的地。
- 路線方式的後續頁仍只恢復 arrival cell，不把目的地辨識擴大到一般事件選項。
- 按鍵追蹤輸出加入地點列舉值、中文名、原文名與城市索引，避免再以訊息文字
  反推狀態。

## 驗證與限制

- `TestRealShadowdaleToAshabenfordUsesTheSelectedTravelDispatcher`：刻意保留舊
  `4C9B=3` 後選阿沙本福德，ECL 必須顯示阿沙本福德旅行方式。
- 匕首瀑布離城／繼續旅行與匕首瀑布至特什維夫傭兵分支回歸均通過。
- 無強化按鍵長跑仍通過、原文 fallback 0；本規格只宣稱目的地交易一致，
  **不宣稱一般強度已連續通關**。世界路線的立即折返後續已由 spec 1238 在測試
  重放器處理，沒有再修改遊戲的旅行規則。
