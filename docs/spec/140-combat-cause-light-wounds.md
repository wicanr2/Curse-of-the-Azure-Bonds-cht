# 第一百四十輪：combat Cause Light Wounds

## RuleBook 證據

本地 RuleBook 的一級牧師法術資料確認：Cause Light Wounds 造成 `1–8 HP`，目標沒有 saving throw；spell table 將它記為 `Cmbt`、range `T`、area `1`、duration `-`。

## 實作結果

- `CauseLightWoundsSpellID` 為 `4`，牧師回合按 `W` 進入敵方 touch target selection。
- Begin／Cancel 不修改 slot；Enter confirmation 才消耗 memorized slot。
- `Battle.CastCauseLightWounds` 以 deterministic `1d8` 造成傷害，並封頂於目標現有 HP。
- 當 caster／target 都有 CombatMap position 時，只有八方向相鄰敵人可成為 touch target；缺少位置資料則保留 bounded fallback。
- 目標死亡時沿用既有 Battle status transition。

## 驗證與限制

測試覆蓋 1–8 傷害、touch range 拒絕、slot transaction 與遊戲狀態傷害同步。range `T` 的完整原版格位、施法動畫與其他 saving／status interaction 仍由後續 CombatMap／ECL rules layer 提供。
