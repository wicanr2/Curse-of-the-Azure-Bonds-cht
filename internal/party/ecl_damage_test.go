package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func TestApplyECLDamageSelectedTargetUsesDOSSaveVerse(t *testing.T) {
	roster := Roster{
		{ID: "first", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
		{ID: "second", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
	}
	request := ecl.DamageRequest{Flags: 0x80, DiceCount: 1, DiceSize: 6, Bonus: 1, SaveFlags: 0x81}
	outcomes, err := roster.ApplyECLDamage(request, 1, func(int) int { return 4 }, func(int) int { return 1 })
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 1 || outcomes[0].TargetIndex != 1 || outcomes[0].Applied != 5 || outcomes[0].Saved {
		t.Fatalf("outcomes=%#v", outcomes)
	}
	if roster[0].HitPoints != 10 || roster[1].HitPoints != 5 {
		t.Fatalf("roster HP=%d,%d", roster[0].HitPoints, roster[1].HitPoints)
	}
}

func TestApplyECLDamageWholePartyHonorsNaturalTwenty(t *testing.T) {
	roster := Roster{
		{ID: "first", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
		{ID: "second", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
	}
	request := ecl.DamageRequest{Flags: 0xC0, DiceCount: 1, DiceSize: 1, SaveFlags: 1}
	outcomes, err := roster.ApplyECLDamage(request, -1, func(int) int { return 1 }, func(int) int { return 20 })
	if err != nil || len(outcomes) != 2 || roster[0].HitPoints != 10 || roster[1].HitPoints != 10 {
		t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, roster, err)
	}
}

func TestApplyECLDamageRandomTargetsUseInjectedCanHitTarget(t *testing.T) {
	roster := Roster{
		{ID: "first", HitPoints: 10},
		{ID: "second", HitPoints: 10},
	}
	request := ecl.DamageRequest{Flags: 2, DiceCount: 1, DiceSize: 1, Bonus: 2, SaveFlags: 3}
	targetRolls := []int{2, 1}
	targetIndex := 0
	outcomes, err := roster.ApplyECLDamageWithHitResolver(request, -1, func(sides int) int {
		if sides == 1 {
			return 1
		}
		value := targetRolls[targetIndex]
		targetIndex++
		return value
	}, func(int) int { return 1 }, func(target Character, bonus int, rollDie func(int) int) (bool, error) {
		if target.ID == "first" && bonus != 3 {
			t.Fatalf("hit bonus=%d, want raw saveFlags 3", bonus)
		}
		return target.ID == "first", nil
	})
	if err != nil || len(outcomes) != 2 || outcomes[0].TargetIndex != 1 || outcomes[0].Hit || outcomes[1].TargetIndex != 0 || !outcomes[1].Hit || roster[0].HitPoints != 7 || roster[1].HitPoints != 10 {
		t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, roster, err)
	}
}
