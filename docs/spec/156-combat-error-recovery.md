# 第一百五十六輪：combat error recovery

## 問題

State 的 combat action API 以 `error` 表示非法目標、彈藥不足、相鄰 missile 限制與未完成 action。Ebiten input adapter 若把這些玩家可修正的錯誤直接返回 `RunGame`，主迴圈可能結束，玩家無法留在戰鬥畫面重新選擇。

## 實作結果

- `combat.ErrAdjacentMissileTarget` 提供可辨識的規則錯誤 sentinel。
- `State.ReportCombatError` 將相鄰 missile 錯誤與一般 combat error 轉成繁中 `combatMessage`，不改變 Battle、HP、彈藥或 turn。
- Ebiten combat input 透過 `combatAction` 攔截所有玩家 combat action errors；錯誤顯示後仍留在目前狀態，下一次輸入可以重新操作。
- 正常 action 的成功／勝負 transition 不改變，非玩家啟動錯誤仍可由啟動器正常回報。

## 明確 boundary

本輪只處理 combat input 的可恢復錯誤呈現，不改變 ranged 距離、line-of-sight、enemy AI、DOS save 或完整錯誤 logging；通用錯誤細節仍保留於訊息字串，後續可再建立更完整的 locale error catalog。
