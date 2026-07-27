# 第五十二輪：ECL ENCOUNTER MENU

狀態：READY（bounded encounter-menu bridge）

## 已完成

- 實作 `ENCOUNTER MENU`（opcode `0x29`、14 operands）cursor framing。
- 解析最大接近距離、memory destination、五個 action mappings 與三個文字 operands。
- interactive runner 提供 `COMBAT／WAIT／FLEE／ADVANCE`；最大距離為零時第四項顯示 `PARLAY`。
- 缺少 selection 時回傳 `WaitingForMenu`。第 308 輪依 reference 修正：
  五個 script values 是 behavior modes；COMBAT 對 modes 0/1/3/4 解析為
  destination `1`，再以 `EncounterActions` 保存結果。mode 2 與其他選項仍需
  movement／distance context。
- regression 覆蓋 pause 與 selection mapping。

## 明確邊界

- 這只完成 VM／menu bridge及已證實的 COMBAT behavior modes；接近距離的地圖碰撞、
  mode 2 group movement、WAIT／ADVANCE loop、PARLAY／FLEE 規則仍未完成。
- 不把 `COMBAT` action mapping 當作已建立 Battle；仍需真實 `LOAD MONSTER` 或外部遭遇資料。
