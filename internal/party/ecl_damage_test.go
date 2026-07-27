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
