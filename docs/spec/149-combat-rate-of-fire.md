# 第一百四十九輪：weapon RateOfFire attacks

## RuleBook／資料證據

RuleBook 的 Combat Fighting 說明：弓每回合可攻擊兩次，飛鏢每回合可攻擊三次；武器表中的 `*`／`#` 另要求箭／弩矢彈藥。原始 `ITEMS` descriptor 的 byte `0x05` 是 RateOfFire，資料以二倍值保存：弓類 raw `4` 對應 2 次，飛鏢 raw `6` 對應 3 次。

## 實作結果

- `BaseItem.RateOfFire` 經 `ItemRecord.Effect` 投影為 `EquipmentEffect.AttacksPerTurn`，再由 `Character.FighterWithEquipment` 寫入 `combat.Fighter`。
- `Battle.AttackSequence` 以 `AttacksPerTurn` 執行 deterministic attack transaction；零值維持舊資料的一次攻擊，目標死亡時停止該 target sequence。
- `State.CombatAct` 若目標在多次攻擊中倒下，將剩餘攻擊轉向 target cursor 的下一個存活敵人；多次結果以繁中摘要顯示，仍只消耗一次 party turn。
- 未帶 equipment 的舊 party JSON 與近戰 raw RateOfFire `0` 行為維持一次攻擊。

## 明確 boundary

本輪只接入已由 ITEMS／RuleBook 證實的 weapon RateOfFire；第 150 輪已另建立 raw ammunition requirement 與注入式 inventory consumption。戰士／聖騎士／遊俠高等級額外攻擊、近戰 sweep、back stab、range／line-of-sight 與完整 Aim Menu 仍待完成。
