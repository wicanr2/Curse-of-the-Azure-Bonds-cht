# 第一百一十七輪：繁中 CAMP Menu state

狀態：`READY`（限目前 remake 的 CAMP command boundary）

## 證據

繁中遊玩手冊的 CAMP 指令列為 `SAVE VIEW MAGIC REST ALTER FIX EXIT`。既有 `State.Camp()` 已驗證休息時恢復目前 party fighter HP，並以 `PROGRAM 9` 作為事件標記。

## 實作 contract

- 荒野選擇 `CAMP` 後進入繁中 CAMP Menu，保存原始 command 與 localized label 的分離。
- `REST` 重用既有安全休息 boundary；事件完成後回到 CAMP Menu。
- `EXIT` 回到荒野的 `ENTER CITY／JOURNEY ON／CAMP` 選單。
- `SAVE／VIEW／MAGIC／ALTER／FIX` 先以明確的繁中 placeholder event 呈現，Enter 後回到 CAMP Menu。
- 若 ECL 已提供 `CAMP` 事件 block，仍保留原有 ECL path，不由新 UI branch 覆寫。

## 驗證

`TestCampMenuRestAndExit` 覆蓋 CAMP 進入、REST return、REST count 與 EXIT return；`go test ./...` 在本輪通過。後續可在解出原版 SAVE、法術恢復、角色修改與修理 routine 後，逐項替換 placeholder service。
