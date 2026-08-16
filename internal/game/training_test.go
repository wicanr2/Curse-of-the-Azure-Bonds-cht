package game

import (
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func trainingTestCatalog(t *testing.T) locale.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestTrainingCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	uiKeys := []string{
		"training_select_character", "training_character_summary", "training_exit",
		"training_requires_healthy", "training_insufficient_gold", "training_insufficient_experience",
		"training_confirm_prompt", "training_confirm", "training_cancel", "training_success",
		"training_hp_increase", "training_dual_class_no_hp", "training_select_spell",
		"training_learned_spell_suffix", "class_cleric", "class_fighter", "class_ranger",
		"class_paladin", "class_magic_user", "class_thief", "class_unknown", "spell_unknown",
	}
	for _, key := range uiKeys {
		if got := catalog.Text(key, ""); got == "" {
			t.Fatalf("training locale ID %q is absent", key)
		}
	}
	seenSpellIDs := make(map[uint8]bool, len(trainingSpells))
	for _, spell := range trainingSpells {
		if seenSpellIDs[spell.ID] {
			t.Fatalf("duplicate training spell ID 0x%02X", spell.ID)
		}
		seenSpellIDs[spell.ID] = true
		if got := catalog.Text(spell.Key, ""); got == "" {
			t.Fatalf("training spell 0x%02X locale ID %q is absent", spell.ID, spell.Key)
		}
	}
}

func TestTrainingHallLevelsFighterAndChargesCharacter(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
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
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		Name: "亞倫", Class: party.ClassFighter, Level: 1, Experience: 2000,
		Platinum: 400, HealthStatus: party.HealthStatusOK,
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != catalog.Text("training_insufficient_experience", "") {
		t.Fatalf("mode=%v message=%q", state.Mode, state.Message)
	}
}

func TestTrainingUsesFixedHPAfterMaximumHitDice(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
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
	state := NewState(trainingTestCatalog(t))
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
	state := NewState(trainingTestCatalog(t))
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

func TestMagicUserTrainingChoosesOneNewKnownSpell(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		Name: "米拉", Race: party.RaceHuman, Class: party.ClassMagicUser, Level: 1,
		ClassLevels: [8]uint8{5: 1}, Experience: 2501,
		HitPoints: 4, MaxHitPoints: 4, Platinum: 400,
		HealthStatus: party.HealthStatusOK,
		Abilities:    party.Abilities{Constitution: 12},
		KnownSpells:  []uint8{9, 10, 11, 12, 13, 14},
	}}
	state.enterTrainingMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.trainingSpellMenu || state.Mode != ModeWilderness || len(state.Choices) == 0 {
		t.Fatalf("spell choice menu=%v mode=%v choices=%v", state.trainingSpellMenu, state.Mode, state.Choices)
	}
	magicMissile := -1
	for index, choice := range state.Choices {
		if choice == catalog.Text("spell_magic_user_15", "") {
			magicMissile = index
			break
		}
	}
	if magicMissile < 0 {
		t.Fatalf("magic missile absent from choices %v", state.Choices)
	}
	if err := state.Select(magicMissile); err != nil {
		t.Fatal(err)
	}
	got := state.partyRoster[0]
	if got.Level != 2 || !containsSpell(got.KnownSpells, 15) || state.Mode != ModeEvent {
		t.Fatalf("trained mage level=%d known=%v mode=%v", got.Level, got.KnownSpells, state.Mode)
	}
}

func TestNinthLevelRangerChoosesFromDruidAndMagicUserSpells(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		Name: "凱拉", Race: party.RaceHuman, Class: party.ClassRanger, Level: 8,
		ClassLevels: [8]uint8{4: 8}, Experience: 225001,
		HitPoints: 50, MaxHitPoints: 50, Platinum: 400,
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
	seenDruid, seenMagicUser := false, false
	for _, original := range state.currentOriginalChoices {
		if original == "TRAIN_SPELL_77" {
			seenDruid = true
		}
		if original == "TRAIN_SPELL_9" {
			seenMagicUser = true
		}
	}
	if !state.trainingSpellMenu || !seenDruid || !seenMagicUser {
		t.Fatalf("ranger spell menu=%v druid=%v magic-user=%v choices=%v",
			state.trainingSpellMenu, seenDruid, seenMagicUser, state.currentOriginalChoices)
	}
}

func containsSpell(spells []uint8, want uint8) bool {
	for _, spellID := range spells {
		if spellID == want {
			return true
		}
	}
	return false
}
