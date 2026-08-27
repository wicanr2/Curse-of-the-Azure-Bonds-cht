# 1212 — 戰鬥 CAST 選單與一般強度長跑

狀態：`READY`

## 原版契約

- spec 1015 已確認戰鬥回合的 `C` 會呼叫原版開始施法常式。
- spec 1105／1168 已確認主命令列包含「施法」，且法術項目是依角色狀態逐項
  組成，不是把 `C` 固定綁成某一支法術。
- 燃燒之手是全域 spell ID `9`；spec 1124／1125 已確認傷害等於施法者等級、
  火焰旗標為 `09h`。remake 的 game pack 與通用 `damage_dice` runtime 已具備
  這些規則，缺的是玩家可到達的正式選單入口。

## Remake 實作

- `State.CombatSpellChoices` 依目前行動角色的 `SpellSlots` 順序列出已宣告的
  戰鬥法術；相同法術的多個記憶槽只顯示一項，但真正施放仍只消耗一個槽。
- 前端 `C` 現在開啟 CAST 選單；方向鍵循環、Enter 選定、Escape 取消。選定後
  仍走既有 `BeginCombatCast`、目標選擇、施法延遲、中斷與 slot transaction。
- 舊的 `C = 詛咒術` 單一快捷鍵已移除。詛咒術仍可由正式 CAST 選單選取。

## 驗證

- `TestCombatSpellChoicesFollowActiveCasterSlotOrderAndFoldDuplicates`：鎖定槽順序、
  重複合併、未宣告法術排除與本地化名稱。
- `TestBurningHandsUsesItsOwnLocalizedName`：五級法師在相鄰敵人前由正式交易
  施放 spell ID 9、造成 5 點傷害，訊息必須顯示「燃燒之手」，不可沿用
  「造成輕傷」模板。
- `go test ./internal/game ./cmd/azure-bonds-game -count=1`：通過。
- 一般強度、既有路線、5000 幀按鍵長跑：319 格、127 句、ECL `0x01` 至
  `0x04`、0 句英文回退；皇家衛兵戰勝利。第二場八名火刀仍未撐過，之後在
  `GEO0x04` 西南區域循環，列為下一個正常玩家路徑瓶頸，不冒稱已通關。

## 邊界

- 本輪不改任何法術的原版施放資格；例如敵人貼身時祝福術可能不可施放。
- 燃燒之手已在正常按鍵長跑中於合法相鄰位置施放並消耗記憶槽；沒有以瞬移或
  直接呼叫製造玩家路徑證據。效果規則仍以 spec 1124／1125 為原版依據。
