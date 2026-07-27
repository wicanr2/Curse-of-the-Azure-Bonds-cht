# 第二百八十七輪：訓練規則與正式流程驗證

狀態：`READY`

## 新證據

公開 CoAB reference `ovr018.sub_509E0` 與 `get_con_hp_adj` 顯示：

- class hit-dice 上限依序是 cleric 10、fighter 10、paladin 10、ranger 11、
  magic-user 12、thief 11。
- 超過上限後不再擲骰；fighter／paladin 固定 +3 HP，cleric／ranger／thief +2，
  magic-user +1。Constitution bonus 只計入仍在 hit-dice 上限內的職業。
- 多職業 HP 由所有現有職業數量相除；martial Constitution bonus 依 primary class
  套用，ranger 第 1 級的 HP 與 Constitution adjustment 都加倍。
- `Limits.RaceClassLimit` 對 dwarf、elf、gnome、half-elf、halfling 的特定職業，
  依目前等級和 Strength／Intelligence 阻止繼續訓練。

## 實作與驗證

訓練 resolver 必須先套 race/class limit，再判斷 XP。HP resolver 在 hit-dice 上限前
維持擲兩次取高，達上限後使用 fixed gain，且採 reference 的跨職業 Constitution
計算。正式 real-image regression 必須由角色建立後的 ECL2／GEO2 `(5,2)` 開始，
驗證 PICTURE 4、中文問題、YES、場所限定 `PROGRAM 0`、角色選擇、確認、扣款、
等級／HP 成長及離開後回到同一地城格。

升級時 magic-user／高等 ranger 的互動式選法術，以及 dual-class
`HitDice <= multiclassLevel` HP gate 尚未具備完整資料模型，本輪不宣稱完成。
