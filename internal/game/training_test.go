package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestTrainingHallLevelsFighterAndChargesCharacter(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.fixSeed = 7
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "亞倫", Class: party.ClassFighter, Level: 1,
		ClassLevels: [8]uint8{2: 1}, Experience: 2001,
		HitPoints: 10, MaxHitPoints: 10, Platinum: 400,
		HealthStatus: party.HealthStatusOK,
		Abilities:    party.Abilities{Constitution: 16},
	}}

	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.trainingConfirmMenu || len(state.Choices) != 2 {
		t.Fatalf("confirmation menu=%v choices=%v", state.trainingConfirmMenu, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	got := state.partyRoster[0]
	if got.Level != 2 || got.ClassLevels[2] != 2 {
		t.Fatalf("trained levels=%d/%v", got.Level, got.ClassLevels)
	}
	if characterCoinGoldWorth(got) != 1000 {
		t.Fatalf("remaining money=%d GP, want 1000", characterCoinGoldWorth(got))
	}
	if got.MaxHitPoints <= 10 || got.HitPoints != got.MaxHitPoints {
		t.Fatalf("trained HP=%d/%d", got.HitPoints, got.MaxHitPoints)
	}
	if state.Mode != ModeEvent {
		t.Fatalf("mode=%v, want event result", state.Mode)
	}
}

func TestTrainingHallRejectsInsufficientExperience(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.partyRoster = party.Roster{{
		Name: "亞倫", Class: party.ClassFighter, Level: 1, Experience: 2000,
		Platinum: 400, HealthStatus: party.HealthStatusOK,
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "經驗值不足，現在還不能升級。" {
		t.Fatalf("mode=%v message=%q", state.Mode, state.Message)
	}
}

func TestTrainingUsesFixedHPAfterMaximumHitDice(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.partyRoster = party.Roster{{
		Name: "老兵", Race: party.RaceHuman, Class: party.ClassFighter, Level: 10,
		ClassLevels: [8]uint8{2: 10}, Experience: 750001,
		HitPoints: 70, MaxHitPoints: 70, Platinum: 400,
		HealthStatus: party.HealthStatusOK,
		Abilities:    party.Abilities{Constitution: 18},
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	got := state.partyRoster[0]
	if got.Level != 11 || got.MaxHitPoints != 73 {
		t.Fatalf("high-level fighter level=%d HP=%d, want 11 and 73", got.Level, got.MaxHitPoints)
	}
}

func TestTrainingEnforcesRaceClassLevelLimit(t *testing.T) {
	character := party.Character{
		Race: party.RaceDwarf, Class: party.ClassFighter, Level: 7,
		ClassLevels: [8]uint8{2: 7}, Experience: 125001,
		Abilities: party.Abilities{StrengthFull: 16},
	}
	if _, _, eligible := trainableClass(character, 0xFF); eligible {
		t.Fatal("dwarf fighter with strength 16 trained past reference level limit")
	}
	character.Abilities.StrengthFull = 18
	if _, level, eligible := trainableClass(character, 0xFF); !eligible || level != 7 {
		t.Fatalf("strong dwarf eligibility=%v level=%d, want level 7 trainable", eligible, level)
	}
}

func TestDualClassTrainingSuppressesHPUntilOldLevelExceeded(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.partyRoster = party.Roster{{
		Name: "轉職者", Race: party.RaceHuman, Class: party.ClassMagicUser, Level: 2,
		ClassLevels: [8]uint8{5: 2}, Experience: 5001,
		HitDice: 2, MulticlassLevel: 5,
		HitPoints: 12, MaxHitPoints: 12, Platinum: 400,
		HealthStatus: party.HealthStatusOK,
		Abilities:    party.Abilities{Constitution: 16},
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	got := state.partyRoster[0]
	if got.Level != 3 || got.HitDice != 3 || got.HitPoints != 12 || got.MaxHitPoints != 12 {
		t.Fatalf("dual-class gate level=%d hitDice=%d HP=%d/%d, want 3/3 and unchanged",
			got.Level, got.HitDice, got.HitPoints, got.MaxHitPoints)
	}
}

func TestDualClassTrainingRestoresHPAfterOldLevelExceeded(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.fixSeed = 11
	state.partyRoster = party.Roster{{
		Name: "轉職者", Race: party.RaceHuman, Class: party.ClassMagicUser, Level: 5,
		ClassLevels: [8]uint8{5: 5}, Experience: 40001,
		HitDice: 5, MulticlassLevel: 5,
		HitPoints: 12, MaxHitPoints: 12, Platinum: 400,
		HealthStatus: party.HealthStatusOK,
		Abilities:    party.Abilities{Constitution: 16},
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	got := state.partyRoster[0]
	if got.Level != 6 || got.HitDice != 6 || got.MaxHitPoints <= 12 {
		t.Fatalf("dual-class recovery level=%d hitDice=%d HP=%d/%d, want 6/6 and HP growth",
			got.Level, got.HitDice, got.HitPoints, got.MaxHitPoints)
	}
}
