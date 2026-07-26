# 第一百四十三輪：encounter FLEE／PARLAY

## RuleBook 證據

Encounter Menu 包含 `COMBAT／WAIT／FLEE／ADVANCE/PARLAY`；FLEE 會讓隊伍撤退，PARLAY 會先選擇談判者／策略。Parlay Menu 的五個策略是 `HAUGHTY／SLY／MEEK／NICE／ABUSIVE`。

## 實作結果

- ECL encounter menu 的 `FLEE` 選項現在進入繁中事件訊息，按 Enter 返回荒野。
- `PARLAY` 現在開啟五項繁中策略選單；選擇後顯示保留原始 tactic identity 的事件訊息，再按 Enter 返回荒野。
- `localizeOption` 保存英文原始 token 與繁中顯示分離，後續 ECL／對話 script 可使用同一組 token 接回。

## 明確 boundary

本輪不猜怪物速度、追擊機率、退後格位、speaker selection、reaction／combat continuation 或完整 encounter conversation script。RuleBook 已確認 menu framing 與 tactic vocabulary，但這些 runtime 結果仍需 ECL／資料層證據。
