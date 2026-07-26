# 第一百一十七輪：繁中 CAMP Menu state

狀態：`READY`（限目前 remake 的 CAMP command boundary）

## 證據

繁中遊玩手冊的 CAMP 指令列為 `SAVE VIEW MAGIC REST ALTER FIX EXIT`。`State.Camp()` 現在將 `PROGRAM 9` 接到 CAMP Menu；`REST` 的自然恢復另由 REST Menu service 處理，不在進入 CAMP 時瞬間補滿 HP。

## 實作 contract

- 荒野選擇 `CAMP` 後進入繁中 CAMP Menu，保存原始 command 與 localized label 的分離。
- `REST` 重用既有安全休息 boundary；事件完成後回到 CAMP Menu。
- `EXIT` 回到荒野的 `ENTER CITY／JOURNEY ON／CAMP` 選單。
- `VIEW／MAGIC／SAVE` 已接入目前 remake 的只讀／spell-slot／party-save boundary；`ALTER → ORDER／DROP／PICS／SPEED／ICON` 已接入 party reorder／confirmed removal／renderer preferences／message reveal／player sprite selection；`FIX` 已接入目前可證實的 Cure Light Wounds healing boundary。原版施法時間、遊戲時間與中斷規則仍未完成。
- 若 ECL 已提供 `CAMP` 事件 block，仍保留原有 ECL path，不由新 UI branch 覆寫。

## 驗證

`TestCampMenuRestAndExit`、`TestCampOpensMenuWithoutInstantHealing`、`TestCampRestNaturallyHealsOneHPPer24Hours` 與 `TestCampFixUsesMemorizedCureLightWounds` 覆蓋 CAMP 進入、REST menu／return、自然治療、FIX 與 EXIT；`go test ./...` 在本輪通過。後續仍可在解出原版 SAVE、完整法術恢復、角色修改與 REST／FIX 的時間／中斷 routine 後，逐項替換窄 service boundary。
