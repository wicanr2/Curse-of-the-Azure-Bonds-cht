# 第一百四十八輪：combat VIEW

## RuleBook 證據

Combat Menu 的 `VIEW` 會顯示角色畫面與 View Menu；戰鬥中部分選項不可用。這是檢視資料的 read-only action，不應消耗 active character 的回合。

## 實作結果

- `State.BeginCombatView` 保存目前 party active fighter，`EndCombatView` 只關閉檢視；兩者都不改變 `combatTurnIndex`、HP 或 spell slots。
- `CombatViewLines` 提供可重用的繁中摘要：角色、生命、護甲等級與攻擊加值。
- Ebiten 以 `V` 開啟，`Enter`／`Esc` 返回戰鬥；檢視中不會誤觸攻擊、施法、選目標或移動。
- 角色摘要由 state/catalog 提供，renderer 只繪製，後續 Gold Box 遊戲可替換角色欄位而沿用 action boundary。

## 明確 boundary

本輪只接入已驗證的 read-only combat VIEW；未猜測原版 View Menu 的交易、物品、裝備、完整角色 record 欄位或戰鬥中的不可用 command 清單。Combat FLEE 仍需速度、戰場外移動與追擊資料證據。

