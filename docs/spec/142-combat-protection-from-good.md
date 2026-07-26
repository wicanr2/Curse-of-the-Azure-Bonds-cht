# 第一百四十二輪：combat Protection from Good 與職業 spell identity

## RuleBook 證據

本地 RuleBook 確認 Protection from Good：對 good attackers，目標 AC 與 saving throws 提高 2；spell table 記為 `Both`、range `T`、area `1`、duration `3r/lvl`。

同一份一級法術表也顯示職業分表：牧師的 Protection from Good 是第 7 個法術，而魔法師的 Magic Missile 也是第 7 個法術。spell ID 不能脫離 caster class 當成全域唯一值。

## 實作結果

- 牧師 spell ID `7` 按 `G` 進入 party touch target selection；魔法師 spell ID `7` 仍按 `S` 進入 enemy target selection。
- Begin／Cancel／Enter transaction 與兩者各自的 slot lookup 都依 caster class 分流。
- fighter 保存 `Good` 與 `ProtectedFromGood`／`ProtectionGoodRounds`；攻擊解析只有 attacker `Good=true` 時才套用 AC +2。
- duration 為 `3 × caster level` 回合，回合開始時倒數並自動解除。

## 明確 boundary

目前 `MON*CHA` 沒有已驗證的 alignment 欄位，因此一般 ECL monster 不會被猜成 good；只有資料／測試明確標記 `Good=true` 才觸發。saving throw +2、完整 alignment import、spell overlap／dispel 仍待後續 rules／DOS adapter。
