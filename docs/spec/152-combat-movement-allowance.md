# 第一百五十二輪：armor movement allowance

## RuleBook／資料證據

RuleBook Armor List 給出護甲的每回合最大移動格數：皮甲 12、軟甲／鑲釘皮甲／環甲／鏈甲／帶甲 9、鱗甲／板條甲／板甲 6；負重過高時另可能限制到 3 格。MOVE 說明也確認移動格數會受護甲與負重影響。

## 實作結果

- `BaseItem.MovementAllowance` 依已確認 armor item types `50–58` 投影 table 上限，`Character.FighterWithEquipment` 對多件 readied armor 取較嚴格上限。
- `combat.Fighter.MovementAllowance` 接入 `State.BeginCombatMove`；每次方向鍵只消耗一格，尚有格數時維持 MOVE mode，耗盡才消耗 party turn。
- 舊 direct fighter／未帶 armor allowance 的資料以 1 格 fallback 維持 compatibility；移入敵格攻擊仍立即結束該 party turn。
- Ebiten MOVE 提示顯示剩餘格數，並新增 armor projection／多格 turn tests。

## 明確 boundary

本輪只接入已證實的護甲上限；尚未接入負重／金幣 encumbrance、地形 cost、障礙物／戰場邊界、離場 FLEE、facing、movement animation 與完整 DOS armor/item mapping。

