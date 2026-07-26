# 第一百五十一輪：combat DONE

## RuleBook 證據

Combat Menu 明列 `DONE`；它讓目前角色完成本回合，不應被當成攻擊或施法。完成 party action 後，遊戲再進入既有的 enemy-turn／下一個 party-turn 流程。

## 實作結果

- `State.CombatDone` 驗證目前是存活 party turn，設定繁中「結束回合」訊息，遞增 `combatTurnIndex`，並重用 `advanceCombatToParty`。
- `D` 鍵接入 Ebiten combat input；VIEW／MOVE／CAST selection 中不會誤觸 DONE。
- DONE 不消耗彈藥、不呼叫 `Battle.Attack`、不改變敵人 HP；它只消耗當前 party turn。

## 明確 boundary

本輪未猜測 RuleBook 其他 Combat Menu command：`AIM`、`USE`、`TURN`、`QUICK` 與完整 `VIEW` menu；DONE 也不代表 hold／delay action，該語意需進一步反組譯。

