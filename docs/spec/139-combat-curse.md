# 第一百三十九輪：combat Curse

## RuleBook 證據

本地 RuleBook 的一級牧師法術資料確認：Curse 將怪物 THAC0 降低 1；施法時與友方角色相鄰的怪物不受影響，且目標沒有 saving throw。spell table 將 Curse 記為 `Cmbt`、range `6`、area `5dia`、duration `6r`。

## 實作結果

- `CurseSpellID` 為 `2`，牧師回合按 `C` 進入敵方目標選擇。
- Begin／Cancel 不修改 slot；Enter confirmation 才消耗一個 memorized Curse。
- `Battle.CastCurse` 對未與存活 party fighter 八方向相鄰的敵人套用 `AttackBonus -1`、`Cursed=true`、`CurseRounds=6`。
- 每次 `StartRound` 遞減 Curse；第六次開始新回合時恢復原攻擊加值。
- 低階 direct API 缺少 position 時採「無法判定則不排除」fallback；沒有把未知位置猜成相鄰。

## 驗證與限制

測試覆蓋相鄰目標不生效、遠距目標 debuff、6 回合 expiry、slot transaction 與繁中 UI routing。range `6`、`5dia` 的完整地圖距離／area 選擇、saving throw（RuleBook 明確為無）以外的 spell overlap／dispel 仍待完整 CombatMap／ECL rules layer。
