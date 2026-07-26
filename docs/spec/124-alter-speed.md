# 第一百二十四輪：ALTER SPEED message reveal

狀態：`READY`（限 remake 訊息顯示速度 preference）

## 證據

RuleBook 的 SPEED Menu 定義 `SLOWER／FASTER／EXIT`，用途是控制訊息顯示速度。本輪以 1–5 級 runtime preference 實作，預設第 3 級。

## 實作 contract

`CAMP → ALTER → SPEED` 可降低或提高速度，並立即回到設定選單顯示目前等級。Ebiten event message 依 `MessageSpeed()` 將 Unicode rune 逐字顯示；state core 不依賴 renderer，其他 UI adapter 可使用相同速度 contract。設定目前未寫入 versioned save。

## 驗證

`TestCampAlterSpeedAdjustsMessageRevealRate` 覆蓋預設值、較慢、較快與返回 ALTER Menu；`go test ./...` 已通過。
