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
