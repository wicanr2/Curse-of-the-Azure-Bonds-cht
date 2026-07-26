# 第一百三十八輪：combat Bless adjacency／duration

## 反組譯／RuleBook 證據

本地 RuleBook 的 spell table 將 Bless duration 記為 `6r`；法術說明指出 Bless 提升友方 THAC0 1，且施法時與怪物相鄰的角色不受影響。這些欄位足以定義本輪的戰鬥 core contract。

## 實作結果

- `combat.Fighter` 保存 `Blessed` 與 `BlessRounds`。
- `Battle.CastBless` 以 CombatMap 的八方向相鄰判定排除鄰近存活怪物的隊友。
- 每次 `StartRound` 先消耗一回合 Bless duration；第六次開始新回合時撤回 `AttackBonus +1`。
- direct API 若 fighter 沒有位置資料，採無法判定則不排除的 bounded fallback，避免虛構座標。

## 驗證

`TestCastBlessSkipsAdjacentPartyAndExpiresAfterSixRounds` 覆蓋相鄰排除、遠距隊友加值、6 回合倒數與修正撤回；既有 game test 覆蓋 B／Enter／Esc 的 slot transaction。

## 尚未宣稱

原版更完整的 area diameter、地形遮蔽、spell dispel／重疊效果與戰鬥回合外時間仍待 rules／ECL layer；本輪只收斂 RuleBook 明確的 adjacency 與 `6r`。
