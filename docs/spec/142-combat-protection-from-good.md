# 第一百四十二輪：combat Protection from Good 與舊職業 spell identity

狀態：`SUPERSEDED`（效果邊界仍有效；spell identity 由 spec 423 取代）

## RuleBook 證據

本地 RuleBook 確認 Protection from Good：對 good attackers，目標 AC 與 saving throws 提高 2；spell table 記為 `Both`、range `T`、area `1`、duration `3r/lvl`。

同一份手冊一級法術表雖各自列出職業內順序，但這不能證明 Player record 採
class-local ID。PC-98 原始 spell table 與 Quick consumer 已證明牧師 Protection
From Good 是全域 `07h`，魔法師 Magic Missile 是全域 `0Fh`；本段舊推論無效。

## 實作結果

- 牧師 spell ID `07h` 按 `G` 進入 party touch target selection；魔法師
  Magic Missile 由 spec 423 修正為 `0Fh`，按 `S` 進入 enemy target selection。
- Begin／Cancel／Enter transaction 與 slot lookup 現依全域 ID 分流；caster
  class 只負責職業資格，不再替同一 raw ID 猜法術身分。
- fighter 保存 `Good` 與 `ProtectedFromGood`／`ProtectionGoodRounds`；攻擊解析只有 attacker `Good=true` 時才套用 AC +2。
- duration 為 `3 × caster level` 回合，回合開始時倒數並自動解除。

## 明確 boundary

目前 `MON*CHA` 沒有已驗證的 alignment 欄位，因此一般 ECL monster 不會被猜成 good；只有資料／測試明確標記 `Good=true` 才觸發。saving throw +2、完整 alignment import、spell overlap／dispel 仍待後續 rules／DOS adapter。
