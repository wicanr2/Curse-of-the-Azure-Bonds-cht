# 第 428 輪：PC-98 手動 CAST 延遲與目標續跑

狀態：`READY`（限已接通法術的 casting-delay／目標 transaction）

## 結論

手動 CAST 與 Quick AI 都呼叫同一個 PC-98 `CASTCOMBATSPELL`，因此 raw
`CastingTime/3` 非零時都必須保存法術並重新插回同輪 scheduler；手動施法
不能繞過第 425 輪已證明的延遲。單體法術保存 stable combatant ID，區域與
直線法術則保存 32×16 戰鬥格點。重新選回施法者後，才以原目標執行既有法術
效果並消耗 memorized slot。

本輪接通目前已實作且 delay 非零的 Bless、Curse、Cure Light Wounds、Cause
Light Wounds、Protection From Evil／Good、Fireball、Lightning Bolt 與
Cloudkill。Magic Missile／Stinking Cloud 的整除結果為零，仍立即結算。

## 非破壞性證據與推論等級

沿用第 425 輪唯讀輸入與 hashes；本輪沒有修改原始 executable、overlay 或
既有 IDA database。

| 證據 | 結論 | 等級 |
|---|---|---|
| overlay 13 `2848h..2909h` | `CASTCOMBATSPELL` 依 `CastingTime/3` 決定立即呼叫或 Action handoff | `exact` |
| overlay 08 `0428h..046Fh` | pending Action 再次入列後清 spell byte並呼叫 `CASTSPELL` | `exact` |
| 手動 UI 與 Quick selector 的 typed call target | 兩條路徑共用 `CASTCOMBATSPELL` | `exact control flow` |
| Action target pointer 的完整 representation | 單體與格點目標需跨 handoff 保留 | `strong inference`；stable ID／座標是 remake typed projection |

原作 action target pointer 的實體生命週期仍未完整關閉，故不得把 remake 的
stable ID／`TargetX／TargetY` 欄位聲稱為原始記憶體 layout。

## remake mapping

- engine `combat/action.State` 保存 renderer-neutral `TargetID` 或
  `TargetX／TargetY／HasTargetPoint`；`Clear` 與 take 操作原子清除整筆交易。
- CoAB `Battle` 只負責將 transaction 放回同一 mutable initiative scheduler。
- CoAB `ConfirmCombatCast` 從正式 game-pack `combat_ai_spells.casting_time` 取得
  delay，不另寫一份法術時間表。
- pending resume 以 stable fighter ID 或格點恢復選擇，再呼叫同一份既有
  `CombatCastWithTerrain`，不複製法術效果。
- slot 仍在實際結算時消耗；中斷／死亡時是否消耗尚無完整 consumer，維持
  `hypothesis`，不宣稱 interruption fidelity。

## 驗證

- engine action：point `(17,9)` 經 delay transaction round-trip。
- Battle：point spell 經 scheduler handoff 後仍回傳同一法術與座標。
- State：手動 Bless 在可見 pending 階段不先套效果、不先消耗 slot，重新入列
  後才生效。
- State visual：手動 Fireball 選定 `(7,6)`，即使 delay=1 在同一呼叫內重新
  入列，最後 visual event 與效果仍使用 `(7,6)`。
- 第 425–427 輪 Quick Bless／Cure regressions 保留，避免共用 resolver 破壞
  Quick target transaction。

## 未完成邊界

- 受傷、死亡、沉默、麻痺等施法中斷條件、文字、slot 消耗時點與動畫／音效。
- 尚未實作法術的 CAST 效果與 Quick suitability。
- 原作 pointer 的 raw layout、save round-trip 與同格多 target ordering。
- Fireball／Lightning Bolt／Cloudkill 各自仍有第 354–357 輪列出的規則與
  動態演出缺口；本輪只關閉 casting handoff。
