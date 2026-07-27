# 第二百八十九輪：訓練升級選法術

狀態：`READY`

## 證據

- DOS Player 的 `spellCastCount[3,5]` 緊接 `hit_point_rolled @ 0x12C`，位於
  `0x12D..0x13B`；三列是 cleric、druid、magic-user，五欄是法術等級 1..5。
- `ReclacClassBonuses → sub_6A00F` 依 magic-user／ranger class level 重算容量。
  magic-user 使用 `MU_spell_lvl_learn`；ranger 8 級後使用 `unk_1A758`，可同時取得
  druid 與 magic-user 容量。
- `BuildSpellList(SpellLoc.choose)` 只加入法術等級 1..5、對應 class/level 容量大於
  0、可學且尚未知的法術。`train_player` 在 magic-user 升級或 ranger 新等級大於 8
  時要求選一個，並寫入 spell book。

## 實作

Character／JSON／DOS patch 保存 3×5 容量。訓練重算 magic-user／ranger 容量後，
依 reference spell class／level metadata 和 KnownSpells 建立候選；選單不提供取消，
選定一個繁中法術後加入 KnownSpells，再顯示升級結果。魔法師與 9 級遊俠分別有
regression，後者必須同時看見 druid 與 magic-user 候選。

現有 spell metadata 只收錄 CoAB `choose` 會用到的 1..5 級 magic-user 與 ranger
druid 法術；monster／cleric 法術不會錯誤出現在升級選單。
