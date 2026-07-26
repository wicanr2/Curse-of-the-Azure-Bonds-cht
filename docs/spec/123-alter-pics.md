# 第一百二十三輪：ALTER PICS preferences

狀態：`READY`（限 remake renderer preference boundary）

## 證據

RuleBook 的 PICS Menu 定義 `MONSTERS ON/OFF` 與 `ANIMATIONS ON/OFF`；前者控制遭遇特寫圖片，後者控制動畫。

## 實作 contract

`CAMP → ALTER → PICS` 顯示兩個目前狀態並可各自切換，設定保留在 `game.State`。圖片關閉時 ECL picture request 仍完成事件轉移，但不進入圖片 renderer；動畫關閉時事件圖片與戰鬥 SPRIT renderer 固定使用首幀。Ebiten 只透過 `PicturesEnabled()`／`AnimationsEnabled()` 讀取設定。

設定目前屬於 runtime remake preference，尚未寫入 versioned save；原版設定欄位與完整 config container 仍待反組譯。

## 驗證

`TestCampAlterPicsTogglesRendererPreferences` 覆蓋進入 PICS、兩個 toggle、狀態 label 與返回 ALTER Menu；`go test ./...` 已通過。
