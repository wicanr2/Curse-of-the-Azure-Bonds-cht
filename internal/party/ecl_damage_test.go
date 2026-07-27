package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
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

func TestApplyECLDamageUsesImportedSavingThrowBonus(t *testing.T) {
	roster := Roster{{ID: "hero", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}, SavingThrowBonus: 2}}
	request := ecl.DamageRequest{Flags: 0x80, DiceCount: 1, DiceSize: 1, SaveFlags: 0x81}
	outcomes, err := roster.ApplyECLDamage(request, 0, func(int) int { return 1 }, func(int) int { return 10 })
	if err != nil || len(outcomes) != 1 || !outcomes[0].Saved || roster[0].HitPoints != 10 {
		t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, roster, err)
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

func TestCanHitECLDamageTargetAppliesInvisibilityRollPenalty(t *testing.T) {
	target := Character{Effects: []monster.AffectRecord{{Kind: 0x19, Active: true}}}
	hit, err := CanHitECLDamageTarget(target, 10, 0, func(int) int { return 14 })
	if err != nil || hit {
		t.Fatalf("invisible target hit=%t err=%v, want miss after -4", hit, err)
	}
	target.Effects[0].Active = false
	hit, err = CanHitECLDamageTarget(target, 10, 0, func(int) int { return 14 })
	if err != nil || !hit {
		t.Fatalf("inactive effect hit=%t err=%v, want hit", hit, err)
	}
}

func TestCanHitECLDamageTargetBlinkCanOverrideNaturalTwenty(t *testing.T) {
	target := Character{Effects: []monster.AffectRecord{{Kind: 0x25, Active: true}}}
	hit, err := CanHitECLDamageTargetWithContext(target, 10, 0, ECLHitContext{ActionDelay: 0}, func(int) int { return 20 })
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("blink with zero action delay should force the hit roll below zero")
	}
	hit, err = CanHitECLDamageTargetWithContext(target, 10, 0, ECLHitContext{ActionDelay: 1}, func(int) int { return 20 })
	if err != nil || !hit {
		t.Fatalf("blink after action delay hit=%t err=%v, want hit", hit, err)
	}
}

func TestCanHitECLDamageTargetDisplaceConsumesFirstHit(t *testing.T) {
	target := Character{Effects: []monster.AffectRecord{{Kind: 0x59, Active: true}}}
	hit, err := CanHitECLDamageTargetWithContext(target, 10, 0, ECLHitContext{CombatRound: 1}, func(int) int { return 20 })
	if err != nil {
		t.Fatal(err)
	}
	if hit || target.Effects[0].Data[0]&0x10 == 0 {
		t.Fatalf("first displaced hit=%t data=%02x, want miss and consumed bit", hit, target.Effects[0].Data[0])
	}
	hit, err = CanHitECLDamageTargetWithContext(target, 10, 0, ECLHitContext{CombatRound: 1}, func(int) int { return 20 })
	if err != nil || !hit {
		t.Fatalf("second displaced hit=%t err=%v, want hit after consumed bit", hit, err)
	}
}

func TestCanHitECLDamageDisplaceClearsBitAtRoundStartWhenRollIsZero(t *testing.T) {
	target := Character{Effects: []monster.AffectRecord{
		{Kind: 0x19, Active: true},
		{Kind: 0x59, Active: true, Data: [4]byte{0x10}},
	}}
	hit, err := CanHitECLDamageTargetWithContext(target, -1, 0, ECLHitContext{CombatRound: 0}, func(int) int { return 4 })
	if err != nil || !hit || target.Effects[1].Data[0]&0x10 != 0 {
		t.Fatalf("round-start displace hit=%t data=%02x err=%v, want hit and cleared bit", hit, target.Effects[1].Data[0], err)
	}
}

func TestApplyECLDamageProjectsReferenceDownAndDeathStates(t *testing.T) {
	tests := []struct {
		name   string
		damage int
		start  HealthStatus
		want   HealthStatus
	}{
		{name: "exact zero", damage: 10, want: HealthStatusUnconscious},
		{name: "small overkill", damage: 15, want: HealthStatusDying},
		{name: "large overkill", damage: 20, want: HealthStatusDead},
		{name: "animated exact zero", damage: 10, start: HealthStatusAnimated, want: HealthStatusDead},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			roster := Roster{{ID: "hero", HitPoints: 10, HealthStatus: test.start}}
			request := ecl.DamageRequest{Flags: 0x80, DiceCount: 1, DiceSize: uint16(test.damage), SaveFlags: 0x80}
			outcomes, err := roster.ApplyECLDamage(request, 0, func(int) int { return test.damage }, func(int) int { return 1 })
			if err != nil || len(outcomes) != 1 || outcomes[0].Applied != 10 || outcomes[0].Health != test.want || roster[0].HealthStatus != test.want || roster[0].HitPoints != 0 {
				t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, roster, err)
			}
		})
	}
}
