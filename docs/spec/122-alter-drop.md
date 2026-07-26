# 第一百二十二輪：ALTER DROP confirmation

狀態：`READY`（限 remake party 的永久移除 transaction）

## 證據

RuleBook 明確說明 ALTER DROP 會從隊伍移除角色並從 saved game disk 刪除，且不可恢復。本輪將這個不可逆操作做成二次確認。

## 實作 contract

`CAMP → ALTER → DROP` 先列出角色，選取後顯示警告與 `確認移除／取消`。只有確認才從 `partyRoster` 移除角色，並依 stable character ID 同步移除 combat fighter；取消不改變任何角色資料。最後一名角色不可移除。

目前使用 remake save 的 roster snapshot；原版 save disk slot／file side effects 仍由後續 SAVGAM container adapter 處理。

## 驗證

`TestCampAlterDropRequiresConfirmationAndRemovesCharacter` 覆蓋角色選擇、取消、確認、roster／fighter 同步與返回 ALTER Menu；`go test ./...` 已通過。
