# 第一百四十一輪：combat Protection from Evil

## RuleBook 證據

本地 RuleBook 確認 Protection from Evil：對邪惡攻擊者，目標 AC 與 saving throws 提高 2；spell table 記為 `Both`、range `T`、area `1`、duration `3r/lvl`。

## 實作結果

- `ProtectionFromEvilSpellID` 為 `6`，牧師回合按 `P` 進入 party touch target selection；施法者本人可選為 self target。
- Begin／Cancel 不修改 slot；Enter confirmation 才消耗 memorized slot。
- fighter 新增 `Evil` 與 `ProtectedFromEvil`／`ProtectionEvilRounds` 狀態。
- `ResolveAttack` 只有在 attacker `Evil=true` 且 target 受防護時，才將 target AC 視為 `+2`；不改寫 base ArmorClass。
- duration 為 `3 × caster level` 回合，回合開始時倒數並自動解除。

## 明確 boundary

目前 `MON*CHA` 沒有已驗證的 alignment 欄位，因此一般 ECL monster 不會被猜成 evil；只有資料／測試明確標記 `Evil=true` 才觸發 AC 修正。saving throw engine、完整 alignment import、spell overlap／dispel 仍待後續 rules／DOS adapter。
