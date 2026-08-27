package party

import "testing"

func TestStartingHitPointsAccumulatesFinalFighterLevels(t *testing.T) {
	levels := [8]uint8{}
	levels[2] = 5
	for seed := int64(0); seed < 200; seed++ {
		points, err := RollStartingHitPoints(fakeHitDice{}, levels, 2, 18, seed)
		if err != nil {
			t.Fatal(err)
		}
		// 五級戰士各擲一次 1d10、每級拿體質 18 的 +4。
		if points.BaseMaxHitPoints < 5 || points.BaseMaxHitPoints > 50 {
			t.Fatalf("seed %d 基礎 HP=%d，超出五級戰士 5..50", seed, points.BaseMaxHitPoints)
		}
		if points.MaxHitPoints != points.BaseMaxHitPoints+20 {
			t.Fatalf("seed %d 體質加值不是 5×4：max=%d base=%d",
				seed, points.MaxHitPoints, points.BaseMaxHitPoints)
		}
	}
}

func TestStartingHitPointsLevelOneMatchesCreationHelper(t *testing.T) {
	levels := [8]uint8{}
	levels[5] = 1
	for seed := int64(0); seed < 50; seed++ {
		starting, err := RollStartingHitPoints(fakeHitDice{}, levels, 5, 16, seed)
		if err != nil {
			t.Fatal(err)
		}
		creation, err := RollCreationHitPoints(fakeHitDice{}, levels, 5, 16, seed)
		if err != nil {
			t.Fatal(err)
		}
		if starting != creation {
			t.Fatalf("seed %d 一級生命週期分歧：starting=%+v creation=%+v", seed, starting, creation)
		}
	}
}

func TestStartingHitPointsIsDeterministic(t *testing.T) {
	levels := [8]uint8{}
	levels[4] = 5
	first, err := RollStartingHitPoints(fakeHitDice{}, levels, 4, 17, 42)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RollStartingHitPoints(fakeHitDice{}, levels, 4, 17, 42)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("同種子結果不一致：%+v != %+v", first, second)
	}
}
