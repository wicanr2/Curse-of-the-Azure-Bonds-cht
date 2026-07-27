# 第一百七十三輪：dungeon pick-lock transaction

狀態：`READY`

## 證據

公開 CoAB reference `engine/ovr015.cs` 的 `pick_lock()` 依隊伍順序檢查角色：角色必須是原版 `Status.okey`，再以一次 d100 roll 與 `player.thief_skills[1]` 比較，條件是 `roll <= skill`。第一次成功即停止；成功後由 caller 對目前門與反向鄰接面呼叫 `MapSetDoorUnlocked`。函式無論成功或失敗都把 `can_pick_door` 設成 false，因此一次嘗試必須消耗 pick opportunity。

同一檔案的 `RemoveKnockSpell()` 依隊伍順序找到第一個 `Spells.knock`，清除一枚 memorized spell。reference `Classes/Spells.cs` 將 Knock 定義為 `0x1F`。

## 本輪邊界

- `internal/dungeon.PickLock` 接受注入 d100，保留「每位隊員都消費一次 roll、健康、隊伍順序、inclusive roll、失敗仍已嘗試」；`HitPoints <= 0` 視為目前資料模型中的非健康角色。
- `internal/dungeon.ConsumeSpell` 實作第一個 spell-slot 的移除，`KnockSpellID` 固定為 `0x1F`。
- 尚未把結果接入 dungeon UI、door mutation、Knock menu 或 bash；caller 仍需依目前 `WallDoorFlags` detail 與 `UnlockDoorWrapped` 完成 transaction。
- 不重算 thief skill，也不以 Dexterity 或 local class 猜測缺失的 DOS 百分比。

## Regression

`internal/dungeon/door_test.go` 覆蓋死亡角色跳過、隊伍順序、`roll == skill` 成功、失敗消耗嘗試，以及 Knock 第一個 slot 消耗。
