package combat

import "testing"

// 突襲遮罩（`Area2 + 596h`）在整個 DOS 與 PC-98 build 裡**只有讀與清 0**：
// 先攻擲骰讀它（`overlay-13 sub_0`），戰鬥主迴圈每輪把它清 0
// （`overlay-08 sub_F3`），沒有任何地方寫進非 0（spec 1136）。所以先攻的
// `−6` 那一支在這一作永遠走不到，CoAB 一律傳遮罩 0 是**有證據的值**。
//
// 這條測試釘住可觀察的後果：沒有 `−6`，`1d6 + 敏捷反應調整（−4..+5）`
// 再夾底到 1，值域就是 `1..11`。有人日後接上非 0 的遮罩時這裡會紅，
// 逼他先把「誰設這個遮罩」的證據補上，而不是照感覺補一個來源。
func TestInitiativeStaysInsideTheNoSurpriseRange(t *testing.T) {
	fighters := []Fighter{
		{ID: "dull", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, Dexterity: 3},
		{ID: "average", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, Dexterity: 10},
		{ID: "nimble", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, Dexterity: 18},
		{ID: "monster", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, Dexterity: 25},
	}
	battle, err := NewBattle(fighters, 1136)
	if err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 200; round++ {
		for _, entry := range battle.initializeRoundDelays() {
			if entry.ActionDelay < 1 || entry.ActionDelay > 11 {
				t.Fatalf("第 %d 輪 %s 的先攻是 %d，落在 1..11 之外——"+
					"只有突襲的 −6 會產生這種值，而原作沒有任何地方設那個遮罩",
					round, entry.ID, entry.ActionDelay)
			}
		}
	}
}
