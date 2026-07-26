# 第一百四十四輪：combat MOVE

## RuleBook 證據

Combat Menu 明確包含 `MOVE`；移動用來移動角色與攻擊，若離開敵人鄰接範圍，敵人可對角色背面免費攻擊。移動格數會受負重與護甲影響，離開戰場則另有速度／追擊規則。

## 實作結果

- 戰鬥中按 `M` 進入移動模式，方向鍵提交一格八方向中的水平／垂直移動；Esc 取消。
- `Battle.Move` 驗證 fighter 存活、單格 delta、CombatMap position 與存活 fighter occupancy。
- 成功移動會更新 `CombatX／CombatY`、清除移動模式、消耗當前 party turn，並沿用既有 enemy-turn advancement。
- 未施法時原本的 enemy target cursor 保持不變；移動模式不會誤觸施法。第 145 輪已補上離開敵人鄰接範圍的 free attack，第 146 輪再補上移入敵方格的攻擊 transaction。

## 明確 boundary

本輪仍未猜測地形／戰場邊界、負重／護甲 movement allowance、facing、離場 flee 與完整 movement animation；進入敵方格的攻擊已由第 146 輪接入 bounded transaction。
