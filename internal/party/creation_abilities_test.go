package party

import "testing"

// fakeAbilityRoll 用原作的參數：3d6 + 1，六次取最大（spec 1103 §二）。
type fakeAbilityRoll struct {
	adjust [6]int
	known  bool
}

func (fakeAbilityRoll) AbilityRollSpec() (int, int, int, int, bool) {
	return 3, 6, 1, 6, true
}

func (f fakeAbilityRoll) RaceAbilityAdjustments(int) ([6]int, bool) {
	return f.adjust, !f.known
}

// 值域：3d6 ＝ 3..18，加 1 之後 4..19，再加種族調整。
func TestCreationAbilitiesStayWithinRolledRange(t *testing.T) {
	// 半獸人：力量 ＋1、體質 ＋1、魅力 −2（spec 1103 §三）。
	tables := fakeAbilityRoll{adjust: [6]int{1, 0, 0, 0, 1, -2}}
	for seed := int64(0); seed < 300; seed++ {
		values, err := RollCreationAbilities(tables, 6, seed)
		if err != nil {
			t.Fatal(err)
		}
		for index, value := range values {
			low, high := 4+tables.adjust[index], 19+tables.adjust[index]
			if value < low || value > high {
				t.Fatalf("seed %d 屬性 %d ＝ %d，超出 %d..%d",
					seed, index, value, low, high)
			}
		}
	}
}

// 六次取最大讓分佈明顯偏高：純 3d6+1 的平均是 11.5，取最大之後應該高得多。
func TestCreationAbilitiesFavourTheBestOfSix(t *testing.T) {
	tables := fakeAbilityRoll{}
	total := 0
	const rounds = 400
	for seed := int64(0); seed < rounds; seed++ {
		values, err := RollCreationAbilities(tables, 7, seed)
		if err != nil {
			t.Fatal(err)
		}
		for _, value := range values {
			total += value
		}
	}
	average := float64(total) / float64(rounds*6)
	if average < 14 {
		t.Fatalf("六次取最大的平均應該明顯高於 11.5，得到 %.2f", average)
	}
}

// 種族調整是常數位移：同一個種子換成有調整的種族，六個值整組平移。
func TestCreationAbilitiesApplyRaceAdjustmentAsConstantShift(t *testing.T) {
	plain := fakeAbilityRoll{}
	// 矮人：體質 ＋1、魅力 −1。
	dwarf := fakeAbilityRoll{adjust: [6]int{0, 0, 0, 0, 1, -1}}
	for seed := int64(0); seed < 200; seed++ {
		base, err := RollCreationAbilities(plain, 7, seed)
		if err != nil {
			t.Fatal(err)
		}
		shifted, err := RollCreationAbilities(dwarf, 1, seed)
		if err != nil {
			t.Fatal(err)
		}
		for index := range base {
			want := base[index] + dwarf.adjust[index]
			if shifted[index] != want {
				t.Fatalf("seed %d 屬性 %d：得到 %d，應為 %d",
					seed, index, shifted[index], want)
			}
		}
	}
}

func TestCreationAbilitiesRejectUnknownRace(t *testing.T) {
	if _, err := RollCreationAbilities(fakeAbilityRoll{known: true}, 99, 1); err == nil {
		t.Fatal("查不到種族調整時應該回錯誤")
	}
}
