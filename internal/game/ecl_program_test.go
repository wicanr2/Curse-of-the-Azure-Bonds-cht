package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func programResult(id uint8) ecl.RunResult {
	return ecl.RunResult{ProgramExit: true, ProgramIDs: []uint8{id}}
}

func TestApplyECLProgramStartMenu(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.Mode = ModeWilderness
	state.eclBlock = []byte{1}

	handled, err := state.applyECLProgram(programResult(0))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 0: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeTitle || state.session != nil || len(state.eclBlock) != 0 {
		t.Fatalf("PROGRAM 0 did not reset to title: mode=%v block=%d", state.Mode, len(state.eclBlock))
	}
}

func TestApplyECLProgramTrainingHall(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.Mode = ModeEvent
	state.eventReturnMode = ModeDungeon
	state.DungeonWallRoof = 0x8C
	state.partyRoster = party.Roster{{Name: "亞倫", Class: party.ClassFighter, Level: 1}}

	handled, err := state.applyECLProgram(programResult(0))
	if err != nil || !handled {
		t.Fatalf("training PROGRAM 0: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeWilderness || !state.trainingMenu || len(state.Choices) != 2 {
		t.Fatalf("training hall not entered: mode=%v menu=%v choices=%v", state.Mode, state.trainingMenu, state.Choices)
	}
}

func TestHallChoiceRecognizesTrainingProgramContext(t *testing.T) {
	if !isTrainingProgramChoice("HALL", programResult(0)) {
		t.Fatal("HALL PROGRAM 0 was not recognized as training")
	}
	if isTrainingProgramChoice("BAR", programResult(0)) ||
		isTrainingProgramChoice("HALL", programResult(3)) {
		t.Fatal("unrelated program context was misclassified as training")
	}
}

func TestApplyECLProgramPartyKilled(t *testing.T) {
	state := NewState(locale.Catalog{})

	handled, err := state.applyECLProgram(programResult(3))
	if err != nil || !handled || !state.PartyKilled() {
		t.Fatalf("PROGRAM 3: handled=%v killed=%v err=%v", handled, state.PartyKilled(), err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 {
		t.Fatalf("PROGRAM 3 did not expose terminal menu: mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil || state.Mode != ModeTitle {
		t.Fatalf("PROGRAM 3 title choice: mode=%v err=%v", state.Mode, err)
	}
}

func TestApplyECLProgramVictoryHealsAndRequestsSave(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.partyRoster = party.Roster{{
		ID: "hero", HitPoints: 0, MaxHitPoints: 12,
		HealthStatus: party.HealthStatusDying, Bleeding: 4,
	}}
	state.party = []combat.Fighter{{
		ID: "hero", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 12,
		DeathOverlay: true, DownedCorpse: true,
	}}

	handled, err := state.applyECLProgram(programResult(8))
	if err != nil || !handled || !state.GameWon() {
		t.Fatalf("PROGRAM 8: handled=%v won=%v err=%v", handled, state.GameWon(), err)
	}
	character, fighter := state.partyRoster[0], state.party[0]
	if character.HitPoints != 12 || character.HealthStatus != party.HealthStatusOK || character.Bleeding != 0 {
		t.Fatalf("PROGRAM 8 roster recovery=%+v", character)
	}
	if fighter.HitPoints != 12 || fighter.DeathOverlay || fighter.DownedCorpse {
		t.Fatalf("PROGRAM 8 fighter recovery=%+v", fighter)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("victory save choice: %v", err)
	}
	if state.Mode != ModeTitle || !state.ConsumeSaveRequest() {
		t.Fatalf("victory save did not return title/request save: mode=%v", state.Mode)
	}
}

func TestApplyECLProgramCamp(t *testing.T) {
	state := NewState(locale.Catalog{})
	state.Mode = ModeEvent
	handled, err := state.applyECLProgram(programResult(9))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 9: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeWilderness || !state.campMenu {
		t.Fatalf("PROGRAM 9 did not enter camp: mode=%v camp=%v", state.Mode, state.campMenu)
	}
}
