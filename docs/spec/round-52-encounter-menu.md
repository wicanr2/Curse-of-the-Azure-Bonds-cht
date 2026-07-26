# 第五十二輪：ECL ENCOUNTER MENU

狀態：READY（bounded encounter-menu bridge）

## 已完成

- 實作 `ENCOUNTER MENU`（opcode `0x29`、14 operands）cursor framing。
- 解析最大接近距離、memory destination、五個 action mappings 與三個文字 operands。
- interactive runner 提供 `COMBAT／WAIT／FLEE／ADVANCE`；最大距離為零時第四項顯示 `PARLAY`。
- 缺少 selection 時回傳 `WaitingForMenu`；selection 會寫入 ECL memory destination，並以 `EncounterActions` 保存實際 mapping。
- regression 覆蓋 pause 與 selection mapping。

## 明確邊界

- 這只完成 VM／menu bridge；接近距離的地圖碰撞、外部 encounter routine、monster roster、PARLAY／FLEE 規則與戰場仍未完成。
- 不把 `COMBAT` action mapping 當作已建立 Battle；仍需真實 `LOAD MONSTER` 或外部遭遇資料。
