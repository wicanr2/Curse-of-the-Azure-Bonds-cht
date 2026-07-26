# 第一百二十八輪：CAMP REST menu／自然恢復

狀態：`READY`（限 REST 時數、自然 HP 恢復與 CAMP menu boundary）

## 證據

RuleBook 的 CAMP 章節明確列出 `REST ADD SUBTRACT EXIT`；REST 會在選定法術後開始計時，每 24 個不間斷小時，受傷角色恢復 1 HP。休息可能被 random encounter 中斷；安全地點才適合長時間休息。

目前沒有完整 memorize menu、原始 game clock 或 encounter interruption data，因此本輪只接入可驗證的時間／自然恢復部分，並將 24 小時作為 deterministic add/subtract 單位。`SetRestHours` 提供 test／未來時間 adapter 的注入口。

## 實作 contract

- `CAMP`／ECL `PROGRAM 9` 只開啟 CAMP Menu，不修改 HP。
- `CAMP → REST` 進入 `開始休息／增加 24 小時／減少 24 小時／返回紮營選單`。
- 開始休息時，每位受傷角色恢復 `restHours / 24` HP，封頂 MaxHP，並以 stable character ID 同步 combat fighter；少於 24 小時不自然恢復。
- REST 完成後進入 event，Enter 回 CAMP Menu；取消不改 HP。spell memorization、time-of-day、random encounter interruption 與 safe-location rule 仍由後續 adapter 接入。

## 驗證

`TestCampMenuRestAndExit` 覆蓋 REST menu／開始／返回／EXIT；`TestCampOpensMenuWithoutInstantHealing` 防止 CAMP 直接補滿 HP；`TestCampRestNaturallyHealsOneHPPer24Hours` 驗證 48 小時恢復 2 HP 並同步 fighter。Docker `go test ./...` 已通過。
