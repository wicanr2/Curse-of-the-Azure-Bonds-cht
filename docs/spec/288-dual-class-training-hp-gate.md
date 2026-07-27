# 第二百八十八輪：dual-class 訓練 HP gate

狀態：`READY`

## 證據

- DOS Player record `HitDice` 位於 `0xE5`，`multiclassLevel` 位於 `0xE6`。
- dual-class 建立時 reference 保存舊職業等級至 `multiclassLevel`，並把 `HitDice`
  重設為 1。
- `ReclacClassBonuses` 在每次升級後以目前 active class levels 更新 `HitDice`。
- `train_player` 在提升職業等級、重算 class bonuses 後，若
  `HitDice <= multiclassLevel` 立即返回，不增加 rolled／current／maximum HP。

## 實作

`party.Character` 保存兩欄，DOS parser、JSON 與 raw-preserving patch 都必須 round-trip。
訓練提升 active class 後更新 `HitDice`；未超過舊職業等級時仍完成扣款與升級，但保持
current/max HP，超過後恢復一般擲骰／Constitution HP 成長。

reference 的升級選法術另依 `spellCastCount[class, spellLevel]` 篩選候選法術；
此欄位與忠實選單已由第 289 輪接續完成。
