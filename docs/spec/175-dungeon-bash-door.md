# 第一百七十五輪：dungeon bash door

狀態：`READY`

## 證據

公開 CoAB reference `engine/ovr015.cs` 的 `bash_door()` 依隊伍順序逐位嘗試，沒有先檢查 `health_status`。detail 3 使用 unpickable-door strength table：18/91–99 以 d6==1、18/100 以 d6<=2、19–20 d6<=3、21–22 d6<=4、23 d6<=5、24 d8<=7、25 直接成功。其他 locked-door branch 依 strength 使用 d6/d8/d10/d12/d20；18/0–50 先將 `bash_worked` 設為 true 仍額外擲 d6，這個看似不尋常的行為保留作 replay evidence。

## 本輪成果與邊界

- `Abilities`／DOS parser 保存 `Str.full` 與 `Str00.cur`；舊 remake save 沒有新欄位時 fallback 到既有 Strength。
- `internal/dungeon.BashDoor` 接受注入骰子，保留隊伍順序、die size、inclusive threshold 與 reference 的 18/0–50 extra roll。
- dungeon preview 新增 `B`：detail 2/3 可撞門，成功才呼叫 GEO 雙側 `UnlockDoorWrapped`。
- 尚未完成完整 `locked_door` menu、bash failure messaging、damage／time side effects、door graphics 或劇情流程 integration；本輪只完成已證實的門判定與 unlock transaction。

## Regression

`internal/dungeon/bash_test.go` 覆蓋 detail 3 strength table、18/0–50 extra roll 與 Strength 25 不擲骰；DOS character regression 覆蓋兩個 imported strength 欄位。
