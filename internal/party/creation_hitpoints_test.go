package party

import "testing"

// fakeHitDice 用原作的實際數值，但不經過 game pack，讓測試只驗機制。
type fakeHitDice struct{}

func (fakeHitDice) HitDiceFor(slot int) (int, int, int, bool) {
	table := map[int][3]int{
		0: {1, 8, 10}, 1: {1, 8, 15}, 2: {1, 10, 10}, 3: {1, 10, 10},
		4: {2, 8, 11}, 5: {1, 4, 12}, 6: {1, 6, 11}, 7: {2, 4, 19},
	}
	entry, ok := table[slot]
	return entry[0], entry[1], entry[2], ok
}

func (fakeHitDice) ConstitutionHPBonus(constitution int) int {
	switch {
	case constitution <= 3:
		return -2
	case constitution <= 6:
		return -1
	case constitution <= 14:
		return 0
	case constitution == 15:
		return 1
	}
	return 2
}

func (fakeHitDice) FighterConstitutionHPBonus(classCombo, constitution int) int {
	// 只有單職戰士／聖騎士／遊俠（+75h ＝ 2／3／4）拿得到（spec 869）。
	if classCombo != 2 && classCombo != 3 && classCombo != 4 {
		return 0
	}
	if constitution < 17 {
		return 0
	}
	bonus := []int{1, 2, 3, 3, 4, 4, 4, 5, 5}
	index := constitution - 17
	if index >= len(bonus) {
		index = len(bonus) - 1
	}
	return bonus[index]
}

// 雙職角色的 HP 是兩個職業各擲一次再取平均（spec 1101 §二）。
// 骰子上限：戰士 1d10、盜賊 1d6，兩槽各拿一份體質加值。
func TestCreationHitPointsAveragesMultiClass(t *testing.T) {
	levels := [8]uint8{}
	levels[2], levels[6] = 1, 1 // 戰士／盜賊，職業組合 14

	for seed := int64(0); seed < 200; seed++ {
		points, err := RollCreationHitPoints(fakeHitDice{}, levels, 14, 16, seed)
		if err != nil {
			t.Fatal(err)
		}
		// 體質 16 ⇒ 每槽 +2，兩槽共 +4，除以 2 之後淨得 +2。
		// 未加成的總和落在 2..16，除以 2 之後是 1..8。
		if points.BaseMaxHitPoints < 1 || points.BaseMaxHitPoints > 8 {
			t.Fatalf("seed %d 基礎最大 HP=%d，超出 1..8", seed, points.BaseMaxHitPoints)
		}
		if points.MaxHitPoints < 3 || points.MaxHitPoints > 10 {
			t.Fatalf("seed %d 最大 HP=%d，超出 3..10", seed, points.MaxHitPoints)
		}
		if points.MaxHitPoints <= points.BaseMaxHitPoints {
			t.Fatalf("seed %d 體質加值沒進到最大 HP：max=%d base=%d",
				seed, points.MaxHitPoints, points.BaseMaxHitPoints)
		}
	}
}

// 戰士系的額外體質加值看的是職業組合編號而不是職業槽，
// 所以「戰士／盜賊」雖然有戰士槽也拿不到（spec 869）。
func TestCreationHitPointsFighterBonusIsSingleClassOnly(t *testing.T) {
	single := [8]uint8{}
	single[2] = 1
	multi := [8]uint8{}
	multi[2], multi[6] = 1, 1

	// 體質 18：一般加值 +2，單職戰士再多 +2。
	soloMin, soloMax := 1+4, 10+4
	for seed := int64(0); seed < 200; seed++ {
		solo, err := RollCreationHitPoints(fakeHitDice{}, single, 2, 18, seed)
		if err != nil {
			t.Fatal(err)
		}
		if solo.MaxHitPoints < soloMin || solo.MaxHitPoints > soloMax {
			t.Fatalf("seed %d 單職戰士 HP=%d，超出 %d..%d",
				seed, solo.MaxHitPoints, soloMin, soloMax)
		}
		combo, err := RollCreationHitPoints(fakeHitDice{}, multi, 14, 18, seed)
		if err != nil {
			t.Fatal(err)
		}
		// 多職只有一般加值：每槽 +2，兩槽共 +4，除以 2 ⇒ 淨 +2。
		if combo.MaxHitPoints > combo.BaseMaxHitPoints+2 {
			t.Fatalf("seed %d 多職拿到了戰士系加值：max=%d base=%d",
				seed, combo.MaxHitPoints, combo.BaseMaxHitPoints)
		}
	}
}

// 遊俠一級擲 2d8，而且體質加值總和會被乘 2（spec 869 的原作行為）。
func TestCreationHitPointsRangerDoublesConstitutionBonus(t *testing.T) {
	levels := [8]uint8{}
	levels[4] = 1

	for seed := int64(0); seed < 200; seed++ {
		points, err := RollCreationHitPoints(fakeHitDice{}, levels, 4, 18, seed)
		if err != nil {
			t.Fatal(err)
		}
		// 2d8 ⇒ 2..16；體質 18 給一般 +2、戰士系再 +2，總共 +4 再乘 2 ＝ +8。
		if points.BaseMaxHitPoints < 2 || points.BaseMaxHitPoints > 16 {
			t.Fatalf("seed %d 2d8 的結果=%d", seed, points.BaseMaxHitPoints)
		}
		if points.MaxHitPoints != points.BaseMaxHitPoints+8 {
			t.Fatalf("seed %d 遊俠加值不是 +8：max=%d base=%d",
				seed, points.MaxHitPoints, points.BaseMaxHitPoints)
		}
	}
}

// 體質低到讓加值變負時，原作的下限是 1 而不是 0 或負數（spec 1101 §一）。
func TestCreationHitPointsFloorsAtOne(t *testing.T) {
	levels := [8]uint8{}
	levels[5] = 1 // 法師 1d4

	for seed := int64(0); seed < 400; seed++ {
		points, err := RollCreationHitPoints(fakeHitDice{}, levels, 5, 3, seed)
		if err != nil {
			t.Fatal(err)
		}
		if points.MaxHitPoints < 1 {
			t.Fatalf("seed %d 最大 HP=%d，低於下限 1", seed, points.MaxHitPoints)
		}
	}
}

// 沒有任何職業槽有等級時原作會除以零；本套件回錯誤而不是恐慌。
func TestCreationHitPointsRejectsClasslessCharacter(t *testing.T) {
	if _, err := RollCreationHitPoints(fakeHitDice{}, [8]uint8{}, 0, 12, 1); err == nil {
		t.Fatal("零職業槽應該回錯誤")
	}
}
