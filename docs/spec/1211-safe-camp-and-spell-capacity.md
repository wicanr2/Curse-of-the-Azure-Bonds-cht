# 1211 — 提爾佛頓安全紮營與高等級法術容量

狀態：`READY`（安全營地路徑與一至十二級牧師／法師、九至十二級聖武士容量；
不宣稱一般強度已通過盜賊公會後的八人火刀連戰）

## 已證實差異

- 武器店正式 continuation 回到 GEO2/0x01 `(3,12)`；該格的 ECL 休息遭遇是
  `1 小時／100%`，所以在此記憶法術必定被皇家衛兵中斷。
- `(7,13)` 的 PreCamp 是 `0/0`。正常按鍵路徑現在用方向鍵沿 GEO 通路
  `E,E,E,S,E` 回到安全格，再走 `CAMP → MAGIC → MEMORIZE → REST`。
- `firstLevelMemorizedCapacity` 原本對所有新施法者固定回傳一格，沒有使用角色
  記錄已存在的 `SpellCastCount[3][5]`。這使五級牧師／法師仍顯示 `1/1`。
- spec 809／810 已以 DOS 與 PC-98 逐位元組一致的增量表證實容量。重算現在補上
  牧師與聖武士列，並保留既有法師／遊俠列；牧師智慧 13..18 的六段額外法術
  依原程式門檻套用。營地的一級容量直接讀 `SpellCastCount`，舊 JSON 才退回一格。

## 正常玩家路徑結果

- 五級牧師在本次確定性角色資料上顯示 `2/5`，休息後帶入祝福與治療輕傷；
  五級法師顯示 `1/4`，休息事件明確回報完成兩名角色的法術記憶。
- 牧師用正式 `B`／Enter／施法延遲流程施放祝福；QUICK 狀態之間以正式 Space
  收回手動控制，下一場不再同步跳到勝敗。
- 一般強度仍在 ECL `0x02` 的八名火刀連戰失利。這一場不是 spec 295 的
  ECL `0x03` 五名火刀檢查哨，不能把敵人數直接改成五。後續要驗 QUICK 戰術、
  起始法術效果與正常玩家可用操作。

## 回歸

- `TestRecalculateTrainingSpellCountsUsesExactLevelFiveTables`
- `TestCampMagicNewClericCanMemorizeFromClericalList`
- `TestKeysDriveARealSessionFromTheTitle`（`COAB_KEY_BOOST=0`）
