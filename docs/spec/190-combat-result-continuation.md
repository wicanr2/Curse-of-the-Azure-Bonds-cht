# 第一百九十輪：combat result continuation

狀態：`READY`

## 問題

戰鬥結果是獨立的繁中 event screen。按 Enter 後若只切換 `ModeWilderness`，仍可能留下進入戰鬥前 ECL 的 options，導致下一次方向鍵／Enter 作用於 stale command。

## 實作

`restoreWildernessMenu` 統一建立三個已驗證的荒野入口：`ENTER CITY`、`JOURNEY ON`、`CAMP`，並提供繁中 labels 與 prompt。`Continue` 在沒有 CAMP／商店／酒館子選單時呼叫它；`leaveLocation` 也重用同一 helper。

## 驗證

回歸測試以可擊敗的 synthetic encounter 產生 combat result，按一次 `Continue` 後確認 mode 與 `currentOriginalChoices` 回到荒野主選單。

## 邊界

完整 encounter script 的戰鬥後 ECL block continuation 尚未完成；本輪只修正已建立 State event boundary 的 UI/input contract。
