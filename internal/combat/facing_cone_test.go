package combat

import "testing"

// 90° 扇形判斷（原作 `overlay-31:054Ah`，spec 1002）的八個方向逐一釘住。
// 起點固定 (10,10)，頂點是起點往該方向前進一格——**不是起點本身**，
// 所以邊界要用頂點去算，用起點算會差一格。
func TestFacingConeCoversNinetyDegreesAroundTheShiftedApex(t *testing.T) {
	const startX, startY = 10, 10
	cases := []struct {
		name      string
		direction uint8
		x, y      int
		want      bool
	}{
		// 方向 0（上），頂點 (10,9)：ty <= 9 − |tx − 10|
		{"上：正前方遠處", 0, 10, 4, true},
		{"上：右邊界剛好在線上", 0, 13, 6, true},
		{"上：右邊界外一格", 0, 14, 6, false},
		{"上：左邊界剛好在線上", 0, 7, 6, true},
		{"上：左邊界外一格", 0, 6, 6, false},
		{"上：正後方", 0, 10, 15, false},

		// 方向 4（下），頂點 (10,11)：ty >= 11 + |tx − 10|
		{"下：正前方遠處", 4, 10, 16, true},
		{"下：右邊界剛好在線上", 4, 13, 14, true},
		{"下：右邊界外一格", 4, 14, 14, false},
		{"下：正後方", 4, 10, 5, false},

		// 方向 2（右），頂點 (11,10)：tx >= 11 + |ty − 10|
		{"右：正前方遠處", 2, 15, 10, true},
		{"右：下邊界剛好在線上", 2, 15, 14, true},
		{"右：下邊界外一格", 2, 14, 14, false},
		{"右：正後方", 2, 5, 10, false},

		// 方向 6（左），頂點 (9,10)：tx <= 9 − |ty − 10|
		{"左：正前方遠處", 6, 5, 10, true},
		{"左：上邊界剛好在線上", 6, 5, 6, true},
		{"左：上邊界外一格", 6, 6, 6, false},
		{"左：正後方", 6, 15, 10, false},

		// 方向 1（右上），頂點 (11,9)：從正上到正右的 90°
		{"右上：頂點的正上方", 1, 11, 4, true},
		{"右上：頂點的正右方", 1, 16, 9, true},
		{"右上：正右方再往下一格", 1, 16, 10, false},
		{"右上：正上方再往左一格", 1, 10, 4, false},

		// 方向 3（右下），頂點 (11,11)：從正右到正下的 90°
		{"右下：頂點的正右方", 3, 16, 11, true},
		{"右下：頂點的正下方", 3, 11, 16, true},
		{"右下：正右方再往上一格", 3, 16, 10, false},

		// 方向 5（左下），頂點 (9,11)：從正下到正左的 90°
		{"左下：頂點的正下方", 5, 9, 16, true},
		{"左下：頂點的正左方", 5, 4, 11, true},
		{"左下：正左方再往上一格", 5, 4, 10, false},

		// 方向 7（左上），頂點 (9,9)：從正左到正上的 90°
		{"左上：頂點的正左方", 7, 4, 9, true},
		{"左上：頂點的正上方", 7, 9, 4, true},
		{"左上：正左方再往下一格", 7, 4, 10, false},
	}
	for _, test := range cases {
		got := InFacingCone(startX, startY, test.x, test.y, test.direction)
		if got != test.want {
			t.Errorf("%s：InFacingCone(%d,%d,%d,%d,%d) = %v，預期 %v",
				test.name, startX, startY, test.x, test.y, test.direction, got, test.want)
		}
	}
}

// 起點自己與正前方那一格在進 case 之前就先放行，所以八個方向都成立；
// 方向 8 與 0FFh（原地）則是無條件成立。
func TestFacingConeAlwaysAcceptsTheOriginAndTheCellAhead(t *testing.T) {
	for direction := uint8(0); direction < 8; direction++ {
		if !InFacingCone(10, 10, 10, 10, direction) {
			t.Errorf("方向 %d：起點自己應該成立", direction)
		}
		ahead := [8][2]int{{10, 9}, {11, 9}, {11, 10}, {11, 11}, {10, 11}, {9, 11}, {9, 10}, {9, 9}}
		if !InFacingCone(10, 10, ahead[direction][0], ahead[direction][1], direction) {
			t.Errorf("方向 %d：正前方那一格應該成立", direction)
		}
	}
	if !InFacingCone(10, 10, 0, 0, 8) || !InFacingCone(10, 10, 0, 0, 0xFF) {
		t.Fatal("方向 8 與 0FFh 是「原地」，一律成立")
	}
}

// 四個座標任一個出界就直接否，**不是夾到邊界**——戰場是 50×25。
func TestFacingConeRejectsAnyOutOfBoundsCoordinate(t *testing.T) {
	cases := [][4]int{
		{-1, 10, 10, 5}, {50, 10, 10, 5}, {10, -1, 10, 5}, {10, 25, 10, 5},
		{10, 10, -1, 5}, {10, 10, 50, 5}, {10, 10, 10, -1}, {10, 10, 10, 25},
	}
	for _, test := range cases {
		if InFacingCone(test[0], test[1], test[2], test[3], 0) {
			t.Errorf("(%d,%d)→(%d,%d) 有座標出界，應該回 false", test[0], test[1], test[2], test[3])
		}
	}
}

func opportunityBattle(t *testing.T, facing uint8, actionCount uint8, delay int) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 3, CombatY: 2},
	}, 1010)
	if err != nil {
		t.Fatal(err)
	}
	goblin := battle.fighters["goblin"]
	goblin.CombatFacing = facing
	goblin.CombatActionCount = actionCount
	goblin.CombatAction.Delay = delay
	battle.fighters["goblin"] = goblin
	return battle
}

// 離開接觸的機會攻擊要先過面向：朝向 −2..＋2 五個方向有一個把移動者圈進
// 扇形才打得到（spec 1010）。往西走的人只落在「朝向 6」那一支扇形裡，
// 所以朝向 6 打得到、朝向 2（背對）打不到。
func TestOpportunityAttackNeedsTheMoverInsideOneOfFiveFacings(t *testing.T) {
	facingWest := opportunityBattle(t, 6, 1, 0)
	result, err := facingWest.MoveWithFreeAttacks("hero", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FreeAttacks) != 1 {
		t.Fatalf("朝向 6 應該打得到，實際 %d 次", len(result.FreeAttacks))
	}

	facingEast := opportunityBattle(t, 2, 1, 0)
	result, err = facingEast.MoveWithFreeAttacks("hero", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FreeAttacks) != 0 {
		t.Fatalf("朝向 2 的五個方向不含 6，不該打得到，實際 %d 次", len(result.FreeAttacks))
	}
}

// 面向檢查有兩個旁路：打手的先攻還沒歸零（這一輪還沒輪到他），
// 或這一輪一次都還沒動過。兩者任一成立就**無條件打**，
// 連背對的朝向 2 也照打。
func TestOpportunityAttackSkipsTheFacingCheckOnBothBypasses(t *testing.T) {
	notYetActed := opportunityBattle(t, 2, 1, 5)
	result, err := notYetActed.MoveWithFreeAttacks("hero", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FreeAttacks) != 1 {
		t.Fatalf("先攻 > 0 應該跳過面向檢查，實際 %d 次", len(result.FreeAttacks))
	}

	noActionYet := opportunityBattle(t, 2, 0, 0)
	result, err = noActionYet.MoveWithFreeAttacks("hero", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FreeAttacks) != 1 {
		t.Fatalf("動作計數 0 應該跳過面向檢查，實際 %d 次", len(result.FreeAttacks))
	}
}
