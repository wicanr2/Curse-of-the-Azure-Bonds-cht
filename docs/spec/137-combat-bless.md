# 第一百三十七輪：combat Bless

## RuleBook 證據

本輪以本地 RuleBook 規則摘錄為依據：Bless 提升友方角色 THAC0 1；施法時鄰近怪物的角色不受影響。spell table 將 Bless 記為 `Both`、`6`、`5dia`、`6r`，本輪只接入已核對的一級戰鬥效果。

## 實作

- `BlessSpellID` 為 `1`。
- 牧師回合按 `B` 進入無目標施法確認；Enter 才消耗一個 memorized slot，Esc 不修改 slot。
- combat core 對所有存活隊友增加 `AttackBonus +1`，以 `Blessed` 狀態避免同一戰鬥重複疊加。
- 訊息與控制列已加入繁中 `B：祝福`。

## 明確 boundary

目前 CombatMap 尚未完成鄰近怪物判定，也沒有完整 spell duration／解除效果模型。因此本輪暫以所有存活隊友為 bounded effect，並在 core comment、測試與知識庫保留未完成項；後續解出 position／duration 後再替換 effect adapter，不改 UI transaction。

## 驗證

`internal/combat` 測試確認 Bless 只提升一次；`internal/game` 測試確認 Begin／Cancel 不扣 slot、Enter confirmation 扣 slot 並同步隊伍攻擊加值。
