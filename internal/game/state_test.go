package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func testCatalog() locale.Catalog {
	return locale.Catalog{Language: "zh-TW", Strings: map[string]string{
		"title": "青色枷的詛咒", "press_enter": "請按 Enter 繼續",
		"you_are_at_the_edge_of": "你已抵達邊界。", "enter_city": "進入城市", "journey_on": "繼續旅程",
	}}
}

func TestLocalizedOpeningFlow(t *testing.T) {
	state := NewState(testCatalog())
	if state.Title != "青色枷的詛咒" || state.Mode != ModeTitle {
		t.Fatalf("initial state=%#v", state)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Choices[0] != "進入城市" {
		t.Fatalf("opening state=%#v", state)
	}
	if err := state.Apply(ActionEnterCity); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "進入城市" {
		t.Fatalf("event state=%#v", state)
	}
}

func TestPictureResultBecomesResumableEvent(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"事件畫面"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x1D, 0x00}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 0x1D || state.OriginalEvent != "PICTURE" {
		t.Fatalf("picture state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.PictureRequested {
		t.Fatalf("picture continuation state=%#v", state)
	}
}

func TestPictureUsesHeadBodyBranchWhenHeadBlockIsPresent(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"人物"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x03, 0x00}
	state.SetSceneCharacter(0x01, 0x02)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.SceneCharacterRequested || state.SceneHeadBlock != 0x01 || state.SceneBodyBlock != 0x03 {
		t.Fatalf("scene character state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.SceneCharacterRequested {
		t.Fatal("scene character request was not cleared")
	}
}

func TestAreaStateHeadBlockDrivesPictureBranch(t *testing.T) {
	state := NewState(testCatalog())
	state.SetAreaState(area.State{GameArea: 2, HeadBlockID: 0x01})
	state.Mode = ModeWilderness
	state.Choices = []string{"人物"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x02, 0x00}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.SceneCharacterRequested || state.SceneHeadBlock != 0x01 || state.SceneBodyBlock != 0x02 {
		t.Fatalf("area-driven scene state=%#v", state)
	}
}

func TestRejectsWrongModeAction(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.Apply(ActionEnterCity); err == nil {
		t.Fatal("expected invalid action")
	}
}

func TestOpeningStateRecordsOriginalECLText(t *testing.T) {
	// 0x80 length + packed "YOU ARE AT THE EDGE OF" candidate.
	packed := []byte{0x64, 0xf5, 0x60, 0x05, 0x21, 0x60, 0x05, 0x48, 0x14, 0x20, 0x58, 0x05, 0x10, 0x71, 0x60, 0x3c, 0x68, 0x00}
	block := append([]byte{0, 0, 0x80, byte(len(packed))}, packed...)
	state := NewStateFromECL(testCatalog(), block)
	if state.OriginalOpening != "YOU ARE AT THE EDGE OF" {
		t.Fatalf("original=%q", state.OriginalOpening)
	}
}

func TestStateUsesECLInitialMenuChoices(t *testing.T) {
	block := make([]byte, 2+47)
	for i := 0; i < 5; i++ {
		pos := 2 + i*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, 0x14, 0x80
	}
	payload := block[2:]
	payload[20] = 0x2B
	payload[21], payload[22], payload[23] = 0x02, 0x00, 0x90
	payload[24], payload[25] = 0x00, 0x02
	// Packed "ENTER CITY" and "JOURNEY ON" from the original ECL string records.
	copy(payload[26:], []byte{0x80, 0x08, 0x14, 0xE5, 0x05, 0x4A, 0x00, 0xC9, 0x51, 0x90})
	copy(payload[36:], []byte{0x80, 0x08, 0x28, 0xF5, 0x52, 0x38, 0x56, 0x60, 0x3C, 0xE0})
	payload[46] = 0x00
	state := NewStateFromECL(testCatalog(), block)
	if len(state.OriginalChoices) != 2 || state.OriginalChoices[0] != "ENTER CITY" || state.Choices[1] != "繼續旅程" {
		t.Fatalf("state=%#v", state)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil || state.Mode != ModeEvent || state.Message != "繼續旅程" {
		t.Fatalf("selected state=%#v err=%v", state, err)
	}
}

func TestLocalizedCityMenuOptions(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["ashabenford"] = "阿沙本福德"
	catalog.Strings["dagger_falls"] = "匕首瀑布"
	if got := localizeOption(catalog, "DAGGER FALLS"); got != "匕首瀑布" {
		t.Fatalf("localized city=%q", got)
	}
}

func TestLocalizedEncounterMenuOptions(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["encounter_combat"] = "戰鬥"
	catalog.Strings["encounter_wait"] = "等待"
	catalog.Strings["encounter_flee"] = "撤退"
	catalog.Strings["encounter_advance"] = "接近"
	catalog.Strings["encounter_parlay"] = "談判"
	for original, want := range map[string]string{"COMBAT": "戰鬥", "WAIT": "等待", "FLEE": "撤退", "ADVANCE": "接近", "PARLAY": "談判"} {
		if got := localizeOption(catalog, original); got != want {
			t.Fatalf("%s localized as %q, want %q", original, got, want)
		}
	}
}

func TestLocationDefaultsToWilderness(t *testing.T) {
	state := NewState(testCatalog())
	if state.Location != LocationWilderness || state.LocationName != "Wilderness" {
		t.Fatalf("location=%v, want wilderness", state.Location)
	}
}

func TestCityMenuSelectionMapsAllLocalizedLocations(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["ashabenford"] = "阿沙本福德"
	catalog.Strings["dagger_falls"] = "匕首瀑布"
	for index, want := range []struct {
		location Location
		name     string
		original string
	}{{LocationShadowdale, "暗影谷", "SHADOWDALE"}, {LocationAshabenford, "阿沙本福德", "ASHABENFORD"}, {LocationDaggerFalls, "匕首瀑布", "DAGGER FALLS"}} {
		state := NewState(catalog)
		state.selectionSequence = []uint16{0, 0, 1, uint16(index)}
		state.applyCitySelection()
		if state.Location != want.location || state.LocationName != want.name || state.OriginalLocation != want.original || state.Area.CurrentCity != uint8(index) {
			t.Fatalf("city %d state=%#v want=%#v", index, state, want)
		}
	}
}

func TestECLLoadFilesTransfersGeoBlockRequest(t *testing.T) {
	state := NewState(testCatalog())
	state.SetInDungeon(true)
	state.applyGeoMapLoad(ecl.RunResult{LoadFilesRequested: true, LoadFiles: [3]uint16{0xFF, 0xFF, 0x10}})
	set, block, ok := state.ConsumeGeoMapRequest()
	if !ok || set != 2 || block != 0x10 {
		t.Fatalf("geo request=(%d,%#x,%v), want set 2 block 0x10", set, block, ok)
	}
	if _, _, ok := state.ConsumeGeoMapRequest(); ok {
		t.Fatal("geo map request should be consumed exactly once")
	}
}

func TestShadowdaleWildernessMapMovementAndExit(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["wilderness"] = "荒野"
	catalog.Strings["exit"] = "離開"
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Location = LocationShadowdale
	state.LocationName = "暗影谷"
	state.Choices = []string{"荒野", "離開"}
	state.currentOriginalChoices = []string{"WILDERNESS", "EXIT"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeMap || state.MapX != 0 || state.MapY != 0 {
		t.Fatalf("map entry state=%#v", state)
	}
	dx, dy := 0, 0
	for _, candidate := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
		if state.WildernessFloor.CanEnter(candidate[0], candidate[1]) {
			dx, dy = candidate[0], candidate[1]
			break
		}
	}
	if dx == 0 && dy == 0 {
		t.Fatal("generated origin has no passable neighbor")
	}
	if err := state.Move(dx, dy); err != nil {
		t.Fatal(err)
	}
	if state.MapX != dx || state.MapY != dy {
		t.Fatalf("map position=(%d,%d)", state.MapX, state.MapY)
	}
	if err := state.LeaveMap(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 3 {
		t.Fatalf("leave state=%#v", state)
	}
}

func TestShadowdalePlaceMenuAndEvents(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["shadowdale_map_prompt"] = "暗影谷荒野"
	catalog.Strings["what_place"] = "你在暗影谷。要去哪裡？"
	catalog.Strings["inn"] = "客棧"
	catalog.Strings["store"] = "商店"
	catalog.Strings["bar"] = "酒館"
	catalog.Strings["leave"] = "離開"
	catalog.Strings["inn_event"] = "暗影谷客棧事件"
	state := NewState(catalog)
	state.Location = LocationShadowdale
	state.Mode = ModeMap
	if err := state.EnterPlaces(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || len(state.Choices) != 4 || state.Choices[0] != "客棧" {
		t.Fatalf("place menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "暗影谷客棧事件" || state.OriginalEvent != "INN" {
		t.Fatalf("inn event=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModePlace {
		t.Fatalf("place continuation state=%#v err=%v", state, err)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeMap {
		t.Fatalf("leave continuation state=%#v err=%v", state, err)
	}
}

func TestStateExposesCombatEntryFromECL(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["combat_started"] = "戰鬥開始（戰鬥規則尚未完成）"
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Choices = []string{"遭遇"}
	state.currentOriginalChoices = []string{"ENCOUNTER"}
	state.eclBlock = []byte{0, 0, 0x24}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "戰鬥開始（戰鬥規則尚未完成）" || state.OriginalEvent != "COMBAT" {
		t.Fatalf("combat state=%#v", state)
	}
}

func TestECLCombatRequestStartsBattleWithConfiguredParty(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"遭遇"}
	state.currentOriginalChoices = []string{"ENCOUNTER"}
	state.eclBlock = []byte{0, 0, 0x0B, 0x00, 0x56, 0x00, 0x01, 0x00, 0x56, 0x24}
	state.eclStart = 0
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1}}
	if err := state.SetParty(party); err != nil {
		t.Fatal(err)
	}
	state.SetMonsterRecords(map[uint8]monster.Record{0x56: {Name: "BUGBEAR", MaxHitPoints: 2, HitPoints: 2, ArmorClass: 0, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1}})
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.CombatActive() || len(state.CombatTargets()) != 1 {
		t.Fatalf("combat state=%#v targets=%#v", state, state.CombatTargets())
	}
}

func TestPlayableCombatStateRunsPartyTurnAndVictory(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HasCombatPosition: true, CombatX: 4, CombatY: 3,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
	}}
	enemies := []combat.Fighter{{
		ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 1, MaxHitPoints: 1, ArmorClass: 0,
		AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
	}}
	if err := state.StartCombat(party, enemies, 7); err != nil {
		t.Fatal(err)
	}
	positions := state.CombatFighters()
	var hero, goblin combat.Fighter
	for _, fighter := range positions {
		if fighter.ID == "hero" {
			hero = fighter
		}
		if fighter.ID == "goblin" {
			goblin = fighter
		}
	}
	if len(positions) != 2 || !hero.HasCombatPosition || hero.CombatX != 4 || hero.CombatY != 3 || !goblin.HasCombatPosition {
		t.Fatalf("combat positions=%#v", positions)
	}
	if !state.CombatActive() || len(state.CombatTargets()) != 1 {
		t.Fatalf("combat state=%#v", state)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("combat result mode=%v status=%v message=%q", state.Mode, state.CombatStatus(), state.Message)
	}
}

func TestStartEncounterBuildsBattleFromECLAndMonsterRecord(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1}}
	records := map[uint8]monster.Record{0x56: {Name: "BUGBEAR", MaxHitPoints: 2, HitPoints: 2, ArmorClass: 0, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1}}
	result := ecl.RunResult{CombatRequested: true, MonsterSetup: &ecl.MonsterSetup{SpriteID: 0x09}, MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 0x56, Count: 1, IconBlock: 0x35}}}
	if err := state.StartEncounter(result, records, party, 11); err != nil {
		t.Fatal(err)
	}
	enemies := state.CombatTargets()
	if !state.CombatActive() || len(enemies) != 1 || enemies[0].Name != "BUGBEAR" || enemies[0].SpriteSet != state.Area.GameArea || enemies[0].SpriteBlock != 0x35 || enemies[0].AnimationBlock != 0x09 || !enemies[0].HasAnimation {
		t.Fatalf("state=%#v enemies=%#v", state, enemies)
	}
}

func TestCampBoundaryAndInGameJournal(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	if err := state.Camp(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.CampCount != 1 || state.OriginalEvent != "PROGRAM 9" {
		t.Fatalf("camp state=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("camp continuation mode=%v err=%v", state.Mode, err)
	}
	if err := state.OpenJournal(); err != nil || state.Mode != ModeJournal || state.JournalTitle == "" || state.JournalText == "" {
		t.Fatalf("journal state=%#v err=%v", state, err)
	}
	firstPage := state.JournalText
	if err := state.NextJournalPage(); err != nil || state.JournalPage != 1 || state.JournalText == firstPage {
		t.Fatalf("journal next page=%#v err=%v", state, err)
	}
	if err := state.PreviousJournalPage(); err != nil || state.JournalPage != 0 || state.JournalText != firstPage {
		t.Fatalf("journal previous page=%#v err=%v", state, err)
	}
	if err := state.CloseJournal(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("journal close mode=%v err=%v", state.Mode, err)
	}
}

func TestPartyPersistsThroughCampAndRestoresHitPoints(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 3, MaxHitPoints: 10}}
	if err := state.SetParty(party); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeWilderness
	if err := state.Camp(); err != nil {
		t.Fatal(err)
	}
	if got := state.PartyFighters()[0].HitPoints; got != 10 {
		t.Fatalf("restored hp=%d, want 10", got)
	}
}

func TestCharacterCreationBuildsPartyAndReturnsToWilderness(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil || state.Mode != ModeCharacterCreation {
		t.Fatalf("open creation mode=%v err=%v", state.Mode, err)
	}
	for index := range state.CreationOptions {
		if err := state.AddCreationCharacter(index); err != nil {
			t.Fatal(err)
		}
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.PartyFighters()) != 3 {
		t.Fatalf("finished state mode=%v party=%#v", state.Mode, state.PartyFighters())
	}
}

func TestCharacterCreationCustomNameSupportsUnicode(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCreationName(); err != nil {
		t.Fatal(err)
	}
	if err := state.AppendCreationName([]rune("阿勇")); err != nil {
		t.Fatal(err)
	}
	if err := state.CommitCreationName(); err != nil {
		t.Fatal(err)
	}
	if state.CreationOptions[0].Name != "阿勇" || state.CreationEditing {
		t.Fatalf("creation=%#v", state.CreationOptions[0])
	}
}

func TestCharacterCreationAdjustsAbilityValues(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.ToggleCreationAbilities(); err != nil {
		t.Fatal(err)
	}
	if err := state.AdjustCreationAbility(1); err != nil {
		t.Fatal(err)
	}
	if got, _ := state.CreationOptions[0].Abilities.Value(0); got != 17 {
		t.Fatalf("strength=%d, want 17", got)
	}
	if err := state.MoveCreationAbility(1); err != nil {
		t.Fatal(err)
	}
	if state.CreationAbility != 1 {
		t.Fatalf("ability cursor=%d", state.CreationAbility)
	}
}

func TestCharacterCreationRerollsAbilitiesWithSeed(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.RerollCreationAbilities(42); err != nil {
		t.Fatal(err)
	}
	first, _ := state.CreationOptions[0].Abilities.Value(0)
	if err := state.RerollCreationAbilities(42); err != nil {
		t.Fatal(err)
	}
	second, _ := state.CreationOptions[0].Abilities.Value(0)
	if first != second || first < 3 || first > 18 {
		t.Fatalf("reroll values=%d,%d", first, second)
	}
}

func TestPartySaveLoadRoundTrip(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.AddCreationCharacter(0); err != nil {
		t.Fatal(err)
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/party.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if len(loaded.PartyFighters()) != 1 || loaded.PartyFighters()[0].Name != "戰士" {
		t.Fatalf("loaded party=%#v", loaded.PartyFighters())
	}
}

func TestResolveSpellSearchUsesLoadedPartyRoster(t *testing.T) {
	state := NewState(testCatalog())
	first := party.Character{
		ID: "p1", Name: "施法者", Race: party.RaceHuman, Class: party.ClassMagicUser, Level: 1,
		Abilities:  party.Abilities{Strength: 10, Intelligence: 14, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10},
		SpellSlots: []uint8{0x12, 0x24},
	}
	second := first
	second.ID, second.Name = "p2", "另一位施法者"
	second.SpellSlots = []uint8{0x12}
	state.partyRoster = party.Roster{first, second}
	match, ok := state.ResolveSpellSearch(ecl.SpellSearch{SpellID: 0x12})
	if !ok || match.CharacterIndex != 0 || match.SlotIndex != 0 {
		t.Fatalf("spell match=%#v ok=%t", match, ok)
	}
}

func TestItemCatalogFeedsCharacterCreationFighterProjection(t *testing.T) {
	data := make([]byte, monster.BaseItemHeaderSize+monster.BaseItemRecordSize)
	data[monster.BaseItemHeaderSize+2] = 2
	data[monster.BaseItemHeaderSize+3] = 4
	data[monster.BaseItemHeaderSize+9] = 2
	data[monster.BaseItemHeaderSize+10] = 4
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	character := party.Character{
		ID: "p1", Name: "戰士", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
		Equipment: []monster.ItemRecord{{Type: 0, Plus: 1, Readied: true}},
	}
	state := NewState(testCatalog())
	state.SetItemCatalog(catalog)
	state.Mode = ModeCharacterCreation
	state.CreationRoster = party.Roster{character}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	fighters := state.PartyFighters()
	if len(fighters) != 1 || fighters[0].DamageDiceCount != 2 || fighters[0].DamageDiceSides != 4 || fighters[0].DamageBonus != 1 || fighters[0].AttackBonus != 4 {
		t.Fatalf("equipped creation fighter=%#v", fighters)
	}
}
