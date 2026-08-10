package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func programResult(id uint8) ecl.RunResult {
	return ecl.RunResult{ProgramExit: true, ProgramIDs: []uint8{id}}
}

func TestProgramCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"program_main_menu", "program_party_killed_prompt", "program_return_title",
		"program_party_killed_message", "program_victory_prompt", "program_victory_save",
		"program_end_without_save", "program_victory_message",
		"program_victory_save_requested", "program_adventure_ended",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == key {
			t.Fatalf("PROGRAM locale ID %q is absent", key)
		}
	}
}

func TestApplyECLProgramStartMenu(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.eclBlock = []byte{1}

	handled, err := state.applyECLProgram(programResult(0))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 0: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeTitle || state.session != nil || len(state.eclBlock) != 0 {
		t.Fatalf("PROGRAM 0 did not reset to title: mode=%v block=%d", state.Mode, len(state.eclBlock))
	}
	if state.Message != catalog.Text("program_main_menu", "") {
		t.Fatalf("PROGRAM 0 message=%q", state.Message)
	}
}

func TestApplyECLProgramTrainingHall(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
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
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)

	handled, err := state.applyECLProgram(programResult(3))
	if err != nil || !handled || !state.PartyKilled() {
		t.Fatalf("PROGRAM 3: handled=%v killed=%v err=%v", handled, state.PartyKilled(), err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 {
		t.Fatalf("PROGRAM 3 did not expose terminal menu: mode=%v choices=%v", state.Mode, state.Choices)
	}
	if state.Prompt != catalog.Text("program_party_killed_prompt", "") ||
		state.Choices[0] != catalog.Text("program_return_title", "") ||
		state.Message != catalog.Text("program_party_killed_message", "") {
		t.Fatalf("PROGRAM 3 localized terminal=%q choices=%v message=%q",
			state.Prompt, state.Choices, state.Message)
	}
	if got, want := state.ConsumeSoundEvents(), []SoundEvent{SoundCrash}; !reflect.DeepEqual(got, want) {
		t.Fatalf("PROGRAM 3 sound events=%v, want %v", got, want)
	}
	if err := state.Select(0); err != nil || state.Mode != ModeTitle {
		t.Fatalf("PROGRAM 3 title choice: mode=%v err=%v", state.Mode, err)
	}
	if state.Message != catalog.Text("program_adventure_ended", "") {
		t.Fatalf("PROGRAM 3 title message=%q", state.Message)
	}
}

func TestApplyECLProgramVictoryHealsAndRequestsSave(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
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
	if state.Prompt != catalog.Text("program_victory_prompt", "") ||
		len(state.Choices) != 2 || state.Choices[0] != catalog.Text("program_victory_save", "") ||
		state.Choices[1] != catalog.Text("program_end_without_save", "") ||
		state.Message != catalog.Text("program_victory_message", "") {
		t.Fatalf("PROGRAM 8 localized terminal=%q choices=%v message=%q",
			state.Prompt, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("victory save choice: %v", err)
	}
	if state.Mode != ModeTitle || !state.ConsumeSaveRequest() {
		t.Fatalf("victory save did not return title/request save: mode=%v", state.Mode)
	}
	if state.Message != catalog.Text("program_victory_save_requested", "") {
		t.Fatalf("victory save title message=%q", state.Message)
	}
}

func TestApplyECLProgramVictoryCanEndWithoutSave(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{ID: "hero", HitPoints: 1, MaxHitPoints: 1}}

	handled, err := state.applyECLProgram(programResult(8))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 8: handled=%v err=%v", handled, err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeTitle || state.ConsumeSaveRequest() ||
		state.Message != catalog.Text("program_adventure_ended", "") {
		t.Fatalf("victory no-save mode=%v requested=%v message=%q",
			state.Mode, state.saveRequested, state.Message)
	}
}

func TestApplyECLProgramCamp(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeEvent
	handled, err := state.applyECLProgram(programResult(9))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 9: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeWilderness || !state.campMenu {
		t.Fatalf("PROGRAM 9 did not enter camp: mode=%v camp=%v", state.Mode, state.campMenu)
	}
}
