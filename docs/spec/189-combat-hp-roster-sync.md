# 第一百八十九輪：combat HP roster synchronization

狀態：`READY`

## 問題

戰鬥核心持有可變的 `combat.Fighter.HitPoints`，而 CAMP、remake JSON 與 DOS SAVGAM writer 使用 `partyRoster`。若只同步 renderer-facing `party` slice，戰鬥受傷會在返回冒險後消失。

## 實作

`finishCombat` 呼叫的 `syncPartyFromBattle` 現在依 fighter ID 更新 `partyRoster.HitPoints`／`MaxHitPoints`，同時維持 `party` mirror。角色名稱、財寶、裝備與法術等 roster 欄位不由 combat fighter 反推，避免遺失 DOS／中文化資料。

## 驗證

回歸測試以 enemy attack 改變 battle HP，結束戰鬥後同時檢查 `PartyFighters` 與 `partyRoster` 的 HP 相同；全套 Go tests 與兩個 CLI build 在 Docker 通過。

## 邊界

其他戰鬥狀態（效果持續時間、彈藥、位置與完整 AD&D death/retreat side effects）仍由各自的 rules adapter 處理。
