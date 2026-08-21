package game

import (
	"archive/zip"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestRealPlayerPathStandingStoneToBurialGlen(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	all := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			all[block.Entry.ID] = block.Data
		}
	}
	state := NewStateFromECLBlocks(trainingTestCatalog(t), all, 0x50)
	hero := combat.Fighter{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
		AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1,
		DamageBonus: 100, AttacksPerTurn: 8, InitiativeBonus: 100,
		SavingThrows: []uint8{1, 1, 1, 1, 1},
	}
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassMagicUser,
		Level: 10, HitPoints: 999, MaxHitPoints: 999,
		ClassLevels:  [8]uint8{0: 10, 5: 10},
		SpellSlots:   []uint8{MagicMissileSpellID, CureLightWoundsSpellID},
		SavingThrows: []uint8{1, 1, 1, 1, 1},
		Abilities: party.Abilities{
			Strength: 18, Intelligence: 18, Wisdom: 18,
			Dexterity: 18, Constitution: 18, Charisma: 18,
		},
	}}
	monsterBlocks, err := dax.Parse(zipData(t, image, "MON6CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	monsterRecords := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monsterRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(6, monsterRecords)
	affectBlocks, err := dax.Parse(zipData(t, image, "MON6SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	monsterAffects := make(map[uint8][]monster.AffectRecord, len(affectBlocks))
	for _, block := range affectBlocks {
		affects, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monsterAffects[block.Entry.ID] = affects
	}
	state.SetMonsterAffectsForECL(6, monsterAffects)
	treasureBlocks, err := ParseTreasureItemBlocks(map[uint8][]byte{
		6: zipData(t, image, "ITEM6.DAX"),
	})
	if err != nil {
		t.Fatal(err)
	}
	state.SetTreasureItemBlocks(treasureBlocks)
	state.session.SetMemoryValue(0x4C59, 1)
	state.session.SetMemoryValue(0x4C5A, 1)
	state.session.SetMemoryValue(0x4C5B, 0xFF)

	// AREA wilderness arrival supplies current-city value 4 and invokes the
	// ordinary ECL SearchLocation lifecycle.
	if err := state.arriveAtWorldLocation(4); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationStandingStone ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Standing Stone arrival location=%v originals=%v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.tyranthraxus-reveal") {
		t.Fatalf("Tyranthraxus reveal message=%q", state.Message)
	}
	wantStandingActions := []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}
	for step := 0; step < 6 && !reflect.DeepEqual(state.currentOriginalChoices, wantStandingActions); step++ {
		switch {
		case state.Mode == ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}):
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected Standing Stone continuation mode=%v originals=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, wantStandingActions) {
		t.Fatalf("Standing Stone mode=%v actions=%v message=%q original=%q",
			state.Mode, state.currentOriginalChoices, state.Message, state.OriginalEvent)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	wantDestinations := []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR", "MYTH DRANNOR"}
	if !reflect.DeepEqual(state.currentOriginalChoices, wantDestinations) {
		t.Fatalf("Standing Stone destinations=%v, want %v", state.currentOriginalChoices, wantDestinations)
	}
	for address, want := range map[uint16]uint16{
		0x4C02: 2,
		0x4C03: 8,
		0x4C04: 11,
		0x4C05: 13,
	} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("route selector memory[0x%04X]=%d,%v want %d,true", address, got, ok, want)
		}
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"WILDERNESS", "EXIT"}) {
		t.Fatalf("Myth Drannor route prompt=%q originals=%v", state.Prompt, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeMap || !state.pendingWorldTravel || state.pendingWorldDestination != 13 {
		t.Fatalf("world travel mode=%v pending=%v destination=%d",
			state.Mode, state.pendingWorldTravel, state.pendingWorldDestination)
	}

	// Enter completes this bounded wilderness travel slice and hands current
	// city 13 back to the same ECL session.
	if err := state.EnterPlaces(); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationMythDrannor ||
		state.Message != gamePackText(t, state, "myth-drannor.edge") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA"}) {
		t.Fatalf("Myth Drannor edge location=%v originals=%v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x40 ||
		state.Area.GameArea != 6 || !state.Area.InDungeon ||
		state.GeoMapSet != 6 || state.GeoMapBlock != 0x40 ||
		state.Message != gamePackText(t, state, "myth-drannor.helm-north") {
		t.Fatalf("Burial Glen mode=%v block=0x%02X area=%+v geo=%d/0x%02X message=%q",
			state.Mode, state.session.CurrentBlockID(), state.Area,
			state.GeoMapSet, state.GeoMapBlock, state.Message)
	}
	for step := 0; step < 4 && state.Mode != ModeDungeon; step++ {
		switch {
		case state.Mode == ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case state.Mode == ModeWilderness &&
			reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}):
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected Burial Glen entry continuation mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Burial Glen continuation mode=%v, want dungeon", state.Mode)
	}
	if state.DungeonX != 2 || state.DungeonY != 15 || state.DungeonDirection != 2 {
		t.Fatalf("Burial Glen spawn=(%d,%d,%d), want (2,15,2)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}

	blocks, err := dax.Parse(zipData(t, image, "GEO6.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var burialGlen geo.Grid
	found := false
	for _, block := range blocks {
		if block.Entry.ID == 0x40 {
			burialGlen, err = geo.Parse(block.Entry.ID, block.Data)
			if err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("GEO6 block 0x40 not found")
	}
	if !burialGlen.CanMoveDungeonWrapped(2, 15, 2) {
		t.Fatal("Burial Glen spawn cannot move east to (3,15)")
	}
	state.SetDungeonGeometryView(3, 15, 2)
	state.DungeonWallRoof = burialGlen.CellWrapped(3, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("quiet Burial Glen step mode=%v, want dungeon", state.Mode)
	}
	if !burialGlen.CanMoveDungeonWrapped(3, 15, 0) {
		t.Fatal("Burial Glen (3,15) cannot move north to spirit at (3,14)")
	}
	state.SetDungeonGeometryView(3, 14, 0)
	state.DungeonWallRoof = burialGlen.CellWrapped(3, 14).Terrain
	if state.DungeonWallRoof != 0x01 {
		t.Fatalf("Burial Glen spirit terrain=%02x, want 01", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 72 ||
		state.Message != gamePackText(t, state, "myth-drannor.elf-spirit.greeting") {
		t.Fatalf("elf spirit mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"GREET", "FLEE", "ATTACK"}) {
		t.Fatalf("elf spirit choices=%v", state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.elf-spirit.journal-25") {
		t.Fatalf("elf spirit greeting result=%q", state.Message)
	}
	wantJournal := gamePackText(t, state, "journal.25")
	found = false
	for _, page := range state.JournalPages {
		found = found || page == wantJournal
	}
	if !found {
		t.Fatalf("Journal 25 not unlocked: %v", state.JournalPages)
	}
	for step := 0; step < 4 && state.Mode != ModeDungeon; step++ {
		switch {
		case state.Mode == ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case state.Mode == ModeWilderness &&
			reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}):
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected post-spirit continuation mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-spirit mode=%v, want dungeon", state.Mode)
	}

	// The same GEO6 block has a passable eastward route from the spirit at
	// (3,14) to terrain 0x82 at (6,14); exercise each normal lifecycle rather
	// than jumping directly to the event program counter.
	for x := 4; x <= 6; x++ {
		if !burialGlen.CanMoveDungeonWrapped(x-1, 14, 2) {
			t.Fatalf("Burial Glen route (%d,14)->(%d,14) is not passable", x-1, x)
		}
		state.SetDungeonGeometryView(x, 14, 2)
		state.DungeonWallRoof = burialGlen.CellWrapped(x, 14).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
	}
	if state.DungeonWallRoof != 0x82 ||
		state.Message != gamePackText(t, state, "myth-drannor.red-web") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER IT", "SPEAK", "HACK IT", "RETREAT"}) {
		t.Fatalf("red web terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !state.ECLStringEditing() || state.ECLStringMaxLength() != 8 ||
		state.Message != gamePackText(t, state, "myth-drannor.red-web.word") {
		t.Fatalf("red web input editing=%v max=%d message=%q",
			state.ECLStringEditing(), state.ECLStringMaxLength(), state.Message)
	}
	if err := state.AppendECLString([]rune("Krrkik")); err != nil {
		t.Fatal(err)
	}
	if err := state.SubmitECLString(); err != nil {
		t.Fatal(err)
	}
	if state.ECLStringEditing() ||
		state.Message != gamePackText(t, state, "myth-drannor.red-web.brighter") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("red web answer editing=%v choices=%v message=%q",
			state.ECLStringEditing(), state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.red-web") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER IT", "SPEAK", "HACK IT", "RETREAT"}) {
		t.Fatalf("red web resumed choices=%v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.CombatStatus() != combat.StatusActive ||
		len(state.livingBySide(combat.SideEnemy)) != 4 {
		t.Fatalf("spider combat mode=%v status=%v enemies=%#v message=%q",
			state.Mode, state.CombatStatus(), state.livingBySide(combat.SideEnemy), state.Message)
	}
	for _, enemy := range state.livingBySide(combat.SideEnemy) {
		if enemy.Name != monsterRecords[0x42].Name {
			t.Fatalf("spider enemy=%q, want source record %q", enemy.Name, monsterRecords[0x42].Name)
		}
	}
	// Save version 7 at a real campaign combat boundary, then rebuild a fresh
	// State from the player-supplied image. The rest of this test must continue
	// through spider victory, the same ECL session's rakshasa handoff, second
	// victory and completion flag using only the loaded state.
	beforeBattle, err := state.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	beforeSession, err := state.session.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	savePath := filepath.Join(t.TempDir(), "myth-drannor-red-web-combat.json")
	if err := state.SavePartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	loaded := NewStateFromECLBlocks(trainingTestCatalog(t), all, 0x50)
	loaded.SetMonsterRecordsForECL(6, monsterRecords)
	loaded.SetMonsterAffectsForECL(6, monsterAffects)
	loaded.SetTreasureItemBlocks(treasureBlocks)
	if err := loaded.LoadPartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	afterBattle, err := loaded.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	afterSession, err := loaded.session.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterBattle, beforeBattle) ||
		!reflect.DeepEqual(afterSession, beforeSession) ||
		loaded.Location != LocationMythDrannor || loaded.Mode != ModeCombat {
		t.Fatalf("red-web campaign restore battleEqual=%v sessionEqual=%v location=%v mode=%v",
			reflect.DeepEqual(afterBattle, beforeBattle),
			reflect.DeepEqual(afterSession, beforeSession), loaded.Location, loaded.Mode)
	}
	state = loaded
	// Exercise ALT+Q semantics from the normal Standing Stone -> GEO -> red-web
	// route, not from a direct-entry battle fixture. Headless mode intentionally
	// runs the delegated actions synchronously; the focused visual regression
	// separately proves that the Ebiten adapter yields and accepts Space.
	if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
		t.Fatalf("normal-path ALT+M enabled=%v err=%v", enabled, err)
	}
	if err := state.CombatQuickAll(); err != nil {
		t.Fatal(err)
	}
	for action := 0; action < 64 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 72 ||
		state.Message != gamePackText(t, state, "myth-drannor.red-web.rakshasa") {
		t.Fatalf("rakshasa reveal mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if !reflect.DeepEqual(state.partyRoster[0].SpellSlots, []uint8{CureLightWoundsSpellID}) {
		t.Fatalf("normal-path Quick Magic did not consume only global spell 0x%02X: %v",
			MagicMissileSpellID, state.partyRoster[0].SpellSlots)
	}
	if changed := state.CombatManualControl(); changed != 1 {
		t.Fatalf("normal-path Space recovery changed=%d want 1", changed)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.CombatStatus() != combat.StatusActive ||
		len(state.livingBySide(combat.SideEnemy)) != 1 ||
		state.livingBySide(combat.SideEnemy)[0].Name != monsterRecords[0x43].Name {
		t.Fatalf("rakshasa combat mode=%v status=%v enemies=%#v",
			state.Mode, state.CombatStatus(), state.livingBySide(combat.SideEnemy))
	}
	for action := 0; action < 32 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent ||
		state.Message != gamePackText(t, state, "myth-drannor.red-web.free") {
		t.Fatalf("post-rakshasa mode=%v message=%q status=%v",
			state.Mode, state.Message, state.CombatStatus())
	}
	if value, ok := state.session.MemoryValue(0x4CBF); !ok || value != 1 {
		t.Fatalf("red-web completion memory=%d,%v want 1,true", value, ok)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-red-web mode=%v, want dungeon", state.Mode)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("completed red web retriggered mode=%v message=%q", state.Mode, state.Message)
	}

	// Continue north through the original GEO door to terrain 0x04 at
	// (6,12). This is the next normally reachable Burial Glen event, not a
	// direct ECL entry or injected encounter.
	for y := 13; y >= 12; y-- {
		if !burialGlen.CanMoveDungeonWrapped(6, y+1, 0) {
			t.Fatalf("Burial Glen route (6,%d)->(6,%d) is not passable", y+1, y)
		}
		state.SetDungeonGeometryView(6, y, 0)
		state.DungeonWallRoof = burialGlen.CellWrapped(6, y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
	}
	for attempt := 0; attempt < 8 && state.Mode == ModeDungeon; attempt++ {
		for _, y := range []int{13, 12} {
			direction := uint8(4)
			if y == 12 {
				direction = 0
			}
			state.SetDungeonGeometryView(6, y, direction)
			state.DungeonWallRoof = burialGlen.CellWrapped(6, y).Terrain
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				if state.Mode != ModeDungeon {
					t.Fatalf("random encounter flee mode=%v choices=%v message=%q",
						state.Mode, state.currentOriginalChoices, state.Message)
				}
			}
			if state.Mode != ModeDungeon {
				break
			}
		}
	}
	if state.DungeonWallRoof != 0x04 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.grave.thri-kreen") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		graveCount, graveCountSet := state.session.MemoryValue(0x4CC1)
		stepGuard, stepGuardSet := state.session.MemoryValue(0x7F81)
		t.Fatalf("grave encounter terrain=%02x mode=%v status=%v message=%q choices=%v enemies=%#v grave-count=%d,%v step-guard=%d,%v",
			state.DungeonWallRoof, state.Mode, state.CombatStatus(), state.Message,
			state.currentOriginalChoices, state.livingBySide(combat.SideEnemy),
			graveCount, graveCountSet, stepGuard, stepGuardSet)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.CombatStatus() != combat.StatusActive ||
		len(state.livingBySide(combat.SideEnemy)) == 0 {
		t.Fatalf("grave combat mode=%v status=%v enemies=%#v message=%q",
			state.Mode, state.CombatStatus(), state.livingBySide(combat.SideEnemy), state.Message)
	}
	for action := 0; action < 64 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Message != gamePackText(t, state, "myth-drannor.grave.skeleton") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"LOOT GRAVE", "REBURY SKELETON", "GO"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "myth-drannor.grave.loot"),
			gamePackText(t, state, "myth-drannor.grave.rebury"),
			gamePackText(t, state, "option.go"),
		}) {
		t.Fatalf("grave menu mode=%v choices=%v original=%v message=%q round=%d enemies=%#v",
			state.Mode, state.Choices, state.currentOriginalChoices, state.Message,
			state.battle.Round(), state.livingBySide(combat.SideEnemy))
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	graveCount, graveCountSet := state.session.MemoryValue(0x4CC1)
	stepGuard, stepGuardSet := state.session.MemoryValue(0x7F81)
	spiritApproval, spiritApprovalSet := state.session.MemoryValue(0x4CBA)
	if state.Mode != ModeDungeon || !graveCountSet || graveCount != 1 ||
		!stepGuardSet || stepGuard != 1 ||
		// ECL6 initializes 4CBAh to biased neutral 0x80. Reburial adds one
		// raw point; the later loot branch subtracts it again.
		!spiritApprovalSet || spiritApproval != 0x81 {
		t.Fatalf("rebury result mode=%v grave-count=%d,%v step-guard=%d,%v spirit-approval=%d,%v",
			state.Mode, graveCount, graveCountSet, stepGuard, stepGuardSet,
			spiritApproval, spiritApprovalSet)
	}

	for attempt := 0; attempt < 8 && state.Mode == ModeDungeon; attempt++ {
		for _, y := range []int{13, 12} {
			direction := uint8(4)
			if y == 12 {
				direction = 0
			}
			state.SetDungeonGeometryView(6, y, direction)
			state.DungeonWallRoof = burialGlen.CellWrapped(6, y).Terrain
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				if state.Mode != ModeDungeon {
					t.Fatalf("second random encounter flee mode=%v choices=%v message=%q",
						state.Mode, state.currentOriginalChoices, state.Message)
				}
			}
			if state.Mode != ModeDungeon {
				break
			}
		}
	}
	if state.Message != gamePackText(t, state, "myth-drannor.grave.thri-kreen") {
		t.Fatalf("second grave event mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	for action := 0; action < 64 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"LOOT GRAVE", "REBURY SKELETON", "GO"}) {
		t.Fatalf("second grave menu mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	graveCount, graveCountSet = state.session.MemoryValue(0x4CC1)
	stepGuard, stepGuardSet = state.session.MemoryValue(0x7F81)
	spiritApproval, spiritApprovalSet = state.session.MemoryValue(0x4CBA)
	gems, jewelry := state.TreasurePool()
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"TREASURE_EXIT"}) ||
		!graveCountSet || graveCount != 2 ||
		!stepGuardSet || stepGuard != 1 ||
		!spiritApprovalSet || spiritApproval != 0x80 ||
		gems != 0 || jewelry != 1 {
		t.Fatalf("loot result mode=%v treasure-menu=%v choices=%v original=%v message=%q grave-count=%d,%v step-guard=%d,%v spirit-approval=%d,%v treasure=%d/%d pending=%d",
			state.Mode, state.treasureMenu, state.Choices, state.currentOriginalChoices,
			state.Message, graveCount, graveCountSet, stepGuard, stepGuardSet,
			spiritApproval, spiritApprovalSet, gems, jewelry, len(state.pendingTreasure))
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-loot continuation mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}

	// The original GEO has a nine-step passable route from the grave at
	// (6,12) to Princess Daemir's terrain 0x03 at (13,14). Walk every cell
	// through the normal dungeon lifecycle; random encounters may interrupt
	// the route, but fleeing resumes the same ECL session and coordinate.
	type dungeonStep struct {
		x         int
		y         int
		direction uint8
	}
	daemirRoute := []dungeonStep{
		{x: 7, y: 12, direction: 2},
		{x: 8, y: 12, direction: 2},
		{x: 9, y: 12, direction: 2},
		{x: 10, y: 12, direction: 2},
		{x: 11, y: 12, direction: 2},
		{x: 12, y: 12, direction: 2},
		{x: 13, y: 12, direction: 2},
		{x: 13, y: 13, direction: 4},
		{x: 13, y: 14, direction: 4},
	}
	previousX, previousY := 6, 12
	for _, step := range daemirRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("Burial Glen route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x03 || state.Mode != ModeEvent ||
		!state.PictureRequested || state.PictureBlock != 72 ||
		state.Message != gamePackText(t, state, "myth-drannor.daemir.offer") {
		t.Fatalf("Daemir arrival terrain=%02x mode=%v picture=%v/%d message=%q choices=%v",
			state.DungeonWallRoof, state.Mode, state.PictureRequested, state.PictureBlock,
			state.Message, state.currentOriginalChoices)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ACCEPT", "REJECT", "KILL", "FLEE"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "option.accept"),
			gamePackText(t, state, "option.reject"),
			gamePackText(t, state, "option.kill"),
			gamePackText(t, state, "option.flee"),
		}) {
		t.Fatalf("Daemir choices=%v originals=%v", state.Choices, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	visitedDaemir, visitedDaemirSet := state.session.MemoryValue(0x4CC0)
	spiritApproval, spiritApprovalSet = state.session.MemoryValue(0x4CBA)
	weaponModifier, weaponModifierSet := state.session.MemoryValue(0x4CBB)
	if state.Message != gamePackText(t, state, "myth-drannor.daemir.blessing") ||
		!visitedDaemirSet || visitedDaemir != 1 ||
		!spiritApprovalSet || spiritApproval != 0x85 ||
		!weaponModifierSet || weaponModifier != 0x02 {
		t.Fatalf("Daemir blessing message=%q visited=%d,%v approval=%d,%v modifier=%02x,%v",
			state.Message, visitedDaemir, visitedDaemirSet,
			spiritApproval, spiritApprovalSet, weaponModifier, weaponModifierSet)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-Daemir continuation mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("visited Daemir terrain retriggered mode=%v message=%q", state.Mode, state.Message)
	}
	if projected, found := state.session.MemoryValue(0x7F71); !found || projected != 0x02 {
		t.Fatalf("Daemir blessing was not projected to party combat work: 7F71=%02x,%v", projected, found)
	}

	// Continue through the original wrapped GEO from Daemir to the nearest
	// uncompleted special terrain. The first terrain 0x93 is at (12,10):
	// ECL6 masks it to selector 0x13 and dispatches to payload +0x195E.
	wallSpiderRoute := []dungeonStep{
		{x: 13, y: 13, direction: 0},
		{x: 13, y: 12, direction: 0},
		{x: 12, y: 12, direction: 6},
		{x: 12, y: 11, direction: 0},
		{x: 12, y: 10, direction: 0},
	}
	previousX, previousY = 13, 14
	for _, step := range wallSpiderRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("post-Daemir route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x93 || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.phase-spider-wall") {
		t.Fatalf("wall-spider arrival terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 10 {
		t.Fatalf("wall-spider combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("wall-spider victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("wall-spider continuation mode=%v message=%q", state.Mode, state.Message)
	}
	wallSpiderCleared, wallSpiderClearedSet := state.session.MemoryValue(0x4CCD)
	if !wallSpiderClearedSet || wallSpiderCleared != 1 {
		t.Fatalf("wall-spider completion 4CCD=%d,%v", wallSpiderCleared, wallSpiderClearedSet)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.CombatActive() {
		t.Fatalf("cleared wall-spider terrain retriggered mode=%v active=%v",
			state.Mode, state.CombatActive())
	}

	// Continue four more legal steps to terrain 0x94. It is the second entry
	// in the same original spider defence cluster and uses independent flag
	// 4CCEh with eight PHASE SPIDER combatants.
	glowingSpiderRoute := []dungeonStep{
		{x: 12, y: 9, direction: 0},
		{x: 13, y: 9, direction: 2},
		{x: 14, y: 9, direction: 2},
		{x: 14, y: 8, direction: 0},
	}
	previousX, previousY = 12, 10
	for _, step := range glowingSpiderRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("wall-to-glowing-spider route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x94 || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.phase-spider-glowing") {
		t.Fatalf("glowing-spider arrival terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 8 {
		t.Fatalf("glowing-spider combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("glowing-spider victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("glowing-spider continuation mode=%v message=%q", state.Mode, state.Message)
	}
	glowingSpiderCleared, glowingSpiderClearedSet := state.session.MemoryValue(0x4CCE)
	if !glowingSpiderClearedSet || glowingSpiderCleared != 1 {
		t.Fatalf("glowing-spider completion 4CCE=%d,%v",
			glowingSpiderCleared, glowingSpiderClearedSet)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.CombatActive() {
		t.Fatalf("cleared glowing-spider terrain retriggered mode=%v active=%v",
			state.Mode, state.CombatActive())
	}

	// Returning south through (14,9) reaches terrain 0x95 at (14,10). It
	// completes the original three-part spider defence and continues into a
	// post-victory bone-pile decision rather than ending at generic combat.
	approvalBeforeBones, approvalBeforeBonesSet := state.session.MemoryValue(0x4CBA)
	if !approvalBeforeBonesSet {
		t.Fatal("Daemir approval is unavailable before the bone-pile event")
	}
	gemsBeforeBones, jewelryBeforeBones := state.TreasurePool()
	pendingItemsBeforeBones := len(state.PendingTreasureItems())
	boneSpiderRoute := []dungeonStep{
		{x: 14, y: 9, direction: 4},
		{x: 14, y: 10, direction: 4},
	}
	previousX, previousY = 14, 8
	for _, step := range boneSpiderRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("glowing-to-bone-spider route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x95 || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.phase-spider-bones") {
		t.Fatalf("bone-spider arrival terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 6 {
		t.Fatalf("bone-spider combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeWilderness {
		t.Fatalf("bone-spider victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"LOOT", "REPLACE IN CRYPTS", "IGNORE"}) ||
		state.Message != gamePackText(t, state, "myth-drannor.bones.prompt") ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "myth-drannor.bones.loot"),
			gamePackText(t, state, "myth-drannor.bones.replace"),
			gamePackText(t, state, "myth-drannor.bones.ignore"),
		}) {
		t.Fatalf("bone-pile menu mode=%v choices=%v original=%v message=%q",
			state.Mode, state.Choices, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	bonesCleared, bonesClearedSet := state.session.MemoryValue(0x4CCF)
	approvalAfterBones, approvalAfterBonesSet := state.session.MemoryValue(0x4CBA)
	gemsAfterBones, jewelryAfterBones := state.TreasurePool()
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		!bonesClearedSet || bonesCleared != 1 ||
		!approvalAfterBonesSet || uint8(approvalAfterBones) != uint8(approvalBeforeBones-1) ||
		gemsAfterBones != gemsBeforeBones+1 ||
		jewelryAfterBones != jewelryBeforeBones ||
		len(state.PendingTreasureItems()) != pendingItemsBeforeBones {
		t.Fatalf("bone loot mode=%v treasure-menu=%v cleared=%d,%v approval=%02x→%02x,%v gems=%d→%d jewelry=%d→%d items=%d→%d",
			state.Mode, state.treasureMenu, bonesCleared, bonesClearedSet,
			approvalBeforeBones, approvalAfterBones, approvalAfterBonesSet,
			gemsBeforeBones, gemsAfterBones,
			jewelryBeforeBones, jewelryAfterBones,
			pendingItemsBeforeBones, len(state.PendingTreasureItems()))
	}
	for state.treasureMenu {
		if err := state.Select(len(state.Choices) - 1); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-bone-loot continuation mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.CombatActive() {
		t.Fatalf("cleared bone-spider terrain retriggered mode=%v active=%v",
			state.Mode, state.CombatActive())
	}

	// Continue along the shortest legal GEO route to the thri-kreen defence
	// cluster. Terrain 0x8E attacks immediately with twelve guards and marks
	// 4CC8h, including the duplicate 0x8E cell later on this route.
	entranceGuardRoute := []dungeonStep{
		{x: 14, y: 9, direction: 0},
		{x: 13, y: 9, direction: 6},
		{x: 12, y: 9, direction: 6},
		{x: 12, y: 8, direction: 0},
		{x: 12, y: 7, direction: 0},
		{x: 11, y: 7, direction: 6},
		{x: 10, y: 7, direction: 6},
	}
	previousX, previousY = 14, 10
	for _, step := range entranceGuardRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("bone-pile-to-entrance-guard route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if !(step.x == 10 && step.y == 7) &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
				if err := state.Select(0); err != nil {
					t.Fatal(err)
				}
				if reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
					if err := state.Select(2); err != nil {
						t.Fatal(err)
					}
				}
				continue
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		if !(step.x == 10 && step.y == 7) && state.Mode != ModeDungeon {
			t.Fatalf("quiet entrance-guard route cell (%d,%d) mode=%v choices=%v message=%q terrain=%02x",
				step.x, step.y, state.Mode, state.currentOriginalChoices,
				state.Message, state.DungeonWallRoof)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8E || state.Mode != ModeCombat ||
		!state.CombatActive() || len(state.CombatTargets()) != 12 ||
		state.Message != gamePackText(t, state, "myth-drannor.thri-kreen.entrance") {
		t.Fatalf("entrance guards terrain=%02x mode=%v active=%v targets=%d message=%q",
			state.DungeonWallRoof, state.Mode, state.CombatActive(),
			len(state.CombatTargets()), state.Message)
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("entrance-guard victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if entranceCleared, found := state.session.MemoryValue(0x4CC8); state.Mode != ModeDungeon || !found || entranceCleared != 1 {
		t.Fatalf("entrance-guard continuation mode=%v 4CC8=%d,%v",
			state.Mode, entranceCleared, found)
	}

	// The route to terrain 0x8F crosses the second 0x8E cell at (9,8).
	// Because 4CC8h is already set, it must not create another battle.
	innerGuardRoute := []dungeonStep{
		{x: 10, y: 8, direction: 4},
		{x: 9, y: 8, direction: 6},
		{x: 9, y: 9, direction: 4},
	}
	previousX, previousY = 10, 7
	for _, step := range innerGuardRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("entrance-to-inner-guard route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if !(step.x == 9 && step.y == 9) &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
				if err := state.Select(0); err != nil {
					t.Fatal(err)
				}
				if reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
					if err := state.Select(2); err != nil {
						t.Fatal(err)
					}
				}
				continue
			}
			if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		if step.x == 9 && step.y == 8 && state.Mode != ModeDungeon {
			t.Fatalf("cleared duplicate terrain 0x8E retriggered mode=%v message=%q",
				state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8F || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.thri-kreen.guards") {
		t.Fatalf("inner guards terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 6 {
		t.Fatalf("inner-guard combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("inner-guard victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if innerCleared, found := state.session.MemoryValue(0x4CC9); state.Mode != ModeDungeon || !found || innerCleared != 1 {
		t.Fatalf("inner-guard continuation mode=%v 4CC9=%d,%v",
			state.Mode, innerCleared, found)
	}

	// One step west reaches terrain 0x90. Since the normal path has already
	// set 4CC8h and 4CC9h, the source ECL suppresses both six-warrior
	// reinforcement waves after the twelve-warrior bivouac battle.
	if !burialGlen.CanMoveDungeonWrapped(9, 9, 6) {
		t.Fatal("inner guards cannot move west to the bivouac")
	}
	moneyBeforeBivouac := state.MoneyPool()
	gemsBeforeBivouac, jewelryBeforeBivouac := state.TreasurePool()
	itemsBeforeBivouac := len(state.PendingTreasureItems())
	state.SetDungeonGeometryView(8, 9, 6)
	state.DungeonWallRoof = burialGlen.CellWrapped(8, 9).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x90 || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.thri-kreen.bivouac") {
		t.Fatalf("bivouac arrival terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 12 {
		t.Fatalf("bivouac combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	bivouacCleared, bivouacClearedSet := state.session.MemoryValue(0x4CCA)
	gemsAfterBivouac, jewelryAfterBivouac := state.TreasurePool()
	if state.CombatStatus() != combat.StatusPartyWon ||
		state.Mode != ModeWilderness || !state.treasureMenu ||
		state.Message != gamePackText(t, state, "myth-drannor.thri-kreen.valuables") ||
		!bivouacClearedSet || bivouacCleared != 1 ||
		state.MoneyPool() != moneyBeforeBivouac+9500 ||
		gemsAfterBivouac != gemsBeforeBivouac+4 ||
		jewelryAfterBivouac != jewelryBeforeBivouac+6 ||
		len(state.PendingTreasureItems()) != itemsBeforeBivouac+1 {
		t.Fatalf("bivouac loot status=%v mode=%v menu=%v message=%q 4CCA=%d,%v money=%d→%d gems=%d→%d jewelry=%d→%d items=%d→%d",
			state.CombatStatus(), state.Mode, state.treasureMenu, state.Message,
			bivouacCleared, bivouacClearedSet,
			moneyBeforeBivouac, state.MoneyPool(),
			gemsBeforeBivouac, gemsAfterBivouac,
			jewelryBeforeBivouac, jewelryAfterBivouac,
			itemsBeforeBivouac, len(state.PendingTreasureItems()))
	}
	for state.treasureMenu {
		if err := state.Select(len(state.Choices) - 1); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("post-bivouac continuation mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.CombatActive() {
		t.Fatalf("cleared bivouac retriggered mode=%v active=%v",
			state.Mode, state.CombatActive())
	}

	// Detour west to the one-shot terrain 0Ch figure before entering the
	// spider mausoleums. This keeps Journal 56 on the ordinary GEO/player
	// path rather than granting it through a direct ECL entry.
	clanFigureRoute := []dungeonStep{
		{x: 7, y: 9, direction: 6},
		{x: 7, y: 8, direction: 0},
		{x: 7, y: 7, direction: 0},
		{x: 8, y: 7, direction: 2},
		{x: 8, y: 6, direction: 0},
		{x: 7, y: 6, direction: 6},
		{x: 6, y: 6, direction: 6},
		{x: 6, y: 7, direction: 4},
		{x: 6, y: 8, direction: 4},
		{x: 5, y: 8, direction: 6},
		{x: 4, y: 8, direction: 6},
	}
	previousX, previousY = 8, 9
	for index, step := range clanFigureRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("clan-figure route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if index < len(clanFigureRoute)-1 &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		if index < len(clanFigureRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("quiet clan-figure route cell (%d,%d) mode=%v choices=%v message=%q",
				step.x, step.y, state.Mode, state.currentOriginalChoices, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x0C || state.Mode != ModeWilderness ||
		state.Prompt != gamePackText(t, state, "myth-drannor.clan-figure.greeting") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
		t.Fatalf("clan figure terrain=%02x mode=%v prompt=%q choices=%v",
			state.DungeonWallRoof, state.Mode, state.Prompt, state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.clan-figure.journal") {
		t.Fatalf("clan figure journal message=%q", state.Message)
	}
	if journals := state.JournalPages; len(journals) == 0 ||
		journals[len(journals)-1] != gamePackText(t, state, "journal.56") {
		t.Fatalf("Journal 56 was not unlocked from game-pack: %v", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("clan figure continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("completed clan figure retriggered mode=%v prompt=%q",
			state.Mode, state.Prompt)
	}

	// Return to the bivouac cell without crossing a geometry boundary, then
	// resume the already-proven mausoleum route.
	clanFigureReturnRoute := []dungeonStep{
		{x: 5, y: 8, direction: 2},
		{x: 6, y: 8, direction: 2},
		{x: 6, y: 7, direction: 0},
		{x: 6, y: 6, direction: 0},
		{x: 7, y: 6, direction: 2},
		{x: 8, y: 6, direction: 2},
		{x: 8, y: 7, direction: 4},
		{x: 7, y: 7, direction: 6},
		{x: 7, y: 8, direction: 4},
		{x: 7, y: 9, direction: 4},
		{x: 8, y: 9, direction: 2},
	}
	previousX, previousY = 4, 8
	for _, step := range clanFigureReturnRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("clan-figure return (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("clan-figure return cell (%d,%d) mode=%v choices=%v message=%q",
				step.x, step.y, state.Mode, state.currentOriginalChoices, state.Message)
		}
		previousX, previousY = step.x, step.y
	}

	// Continue from the bivouac to terrain 0x91 at (9,2). The legal route
	// deliberately crosses already-cleared 0x8F and 0x8E cells; neither may
	// replay while random encounters still use their normal pause→FLEE path.
	spiderMausoleumRoute := []dungeonStep{
		{x: 9, y: 9, direction: 2},
		{x: 9, y: 8, direction: 0},
		{x: 9, y: 7, direction: 0},
		{x: 10, y: 7, direction: 2},
		{x: 11, y: 7, direction: 2},
		{x: 11, y: 6, direction: 0},
		{x: 11, y: 5, direction: 0},
		{x: 11, y: 4, direction: 0},
		{x: 11, y: 3, direction: 0},
		{x: 10, y: 3, direction: 6},
		{x: 9, y: 3, direction: 6},
		{x: 9, y: 2, direction: 0},
	}
	previousX, previousY = 8, 9
	for _, step := range spiderMausoleumRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("bivouac-to-spider-mausoleum route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if state.Mode == ModeWilderness &&
				!(step.x == 9 && step.y == 2) &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
				if err := state.Select(0); err != nil {
					t.Fatal(err)
				}
				if reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
					if err := state.Select(2); err != nil {
						t.Fatal(err)
					}
				}
				continue
			}
			if reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		if !(step.x == 9 && step.y == 2) && state.Mode != ModeDungeon {
			t.Fatalf("quiet spider-mausoleum route cell (%d,%d) mode=%v choices=%v message=%q terrain=%02x",
				step.x, step.y, state.Mode, state.currentOriginalChoices,
				state.Message, state.DungeonWallRoof)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x91 || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != gamePackText(t, state, "myth-drannor.spider-mausoleum") {
		t.Fatalf("spider mausoleum terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 8 {
		t.Fatalf("spider-mausoleum combat mode=%v active=%v targets=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("spider-mausoleum victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if mausoleumCleared, found := state.session.MemoryValue(0x4CCB); state.Mode != ModeDungeon || !found || mausoleumCleared != 1 {
		t.Fatalf("spider-mausoleum continuation mode=%v 4CCB=%d,%v",
			state.Mode, mausoleumCleared, found)
	}

	// Terrain 0x92 is two steps away at (10,1). The party still has
	// approval >=80h, so the spirit warning and YES/NO boundary must appear.
	funnelRoute := []dungeonStep{
		{x: 9, y: 1, direction: 0},
		{x: 10, y: 1, direction: 2},
	}
	previousX, previousY = 9, 2
	for _, step := range funnelRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("mausoleum-to-funnel route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if state.Mode == ModeWilderness &&
				!(step.x == 10 && step.y == 1) &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
				if err := state.Select(0); err != nil {
					t.Fatal(err)
				}
				if reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
					if err := state.Select(2); err != nil {
						t.Fatal(err)
					}
				}
				continue
			}
			if reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		previousX, previousY = step.x, step.y
	}
	if approval, found := state.session.MemoryValue(0x4CBA); !found || approval < 0x80 {
		t.Fatalf("funnel path approval=%02x,%v, want warning-eligible", approval, found)
	}
	if state.DungeonWallRoof != 0x92 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.spider-funnel") {
		t.Fatalf("spider funnel terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "option.yes"),
			gamePackText(t, state, "option.no"),
		}) ||
		state.Message != gamePackText(t, state, "myth-drannor.spider-warning") {
		t.Fatalf("spider warning mode=%v choices=%v original=%v message=%q",
			state.Mode, state.Choices, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("spider warning NO mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if _, found := state.session.MemoryValue(0x4CCC); found {
		t.Fatalf("spider warning NO unexpectedly set 4CCC")
	}

	// Re-enter and accept. The source writes 4CCC before combat and projects
	// +2 into enemy-side attack-roll work 7F70h.
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.spider-funnel") {
		t.Fatalf("spider funnel revisit message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.spider-eggs") {
		t.Fatalf("spider eggs mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if nestMarked, found := state.session.MemoryValue(0x4CCC); !found || nestMarked != 1 {
		t.Fatalf("spider nest pre-combat 4CCC=%d,%v", nestMarked, found)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatTargets()) != 4 ||
		state.battle.SideAttackRollModifier(combat.SideEnemy) != 2 {
		t.Fatalf("spider-nest combat mode=%v active=%v targets=%d enemy-modifier=%d",
			state.Mode, state.CombatActive(), len(state.CombatTargets()),
			state.battle.SideAttackRollModifier(combat.SideEnemy))
	}
	for turn := 0; turn < 4 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("spider-nest victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("spider-nest continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.CombatActive() {
		t.Fatalf("cleared spider nest retriggered mode=%v active=%v",
			state.Mode, state.CombatActive())
	}

	// Continue west through the six-cell terrain 96h corridor to the elven
	// court entrance at terrain 89h. Terrain 96h has GEO geometry only and
	// must remain quiet; already-cleared 91h may not replay.
	courtEntranceRoute := []dungeonStep{
		{x: 9, y: 1, direction: 6},
		{x: 9, y: 2, direction: 4},
		{x: 9, y: 3, direction: 4},
		{x: 8, y: 3, direction: 6},
		{x: 7, y: 3, direction: 6},
		{x: 7, y: 2, direction: 0},
		{x: 6, y: 2, direction: 6},
		{x: 5, y: 2, direction: 6},
	}
	previousX, previousY = 10, 1
	for _, step := range courtEntranceRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("spider-nest-to-court route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		for attempt := 0; attempt < 8; attempt++ {
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatal(err)
			}
			if step.x == 5 && step.y == 2 {
				break
			}
			if state.Mode == ModeWilderness &&
				reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
				if err := state.Select(0); err != nil {
					t.Fatal(err)
				}
				if reflect.DeepEqual(state.currentOriginalChoices,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
					if err := state.Select(2); err != nil {
						t.Fatal(err)
					}
				}
				continue
			}
			if reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) {
				if err := state.Select(2); err != nil {
					t.Fatal(err)
				}
				continue
			}
			break
		}
		if step.x == 5 && step.y == 2 {
			break
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("quiet court route cell (%d,%d) mode=%v choices=%v message=%q terrain=%02x",
				step.x, step.y, state.Mode, state.currentOriginalChoices,
				state.Message, state.DungeonWallRoof)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x08 || state.Mode != ModeEvent ||
		state.PictureBlock != 72 ||
		state.Message != gamePackText(t, state, "myth-drannor.court.entry") {
		t.Fatalf("court entry terrain=%02x mode=%v picture=%d message=%q",
			state.DungeonWallRoof, state.Mode, state.PictureBlock, state.Message)
	}
	for page := 0; page < 4 &&
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}); page++ {
		switch state.Mode {
		case ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case ModeWilderness:
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected court entry continuation mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) ||
		state.Message != gamePackText(t, state, "myth-drannor.court.enter") {
		t.Fatalf("court enter choices=%v original=%v message=%q",
			state.Choices, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent ||
		state.Message != gamePackText(t, state, "myth-drannor.court.welcome") {
		t.Fatalf("court welcome mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("court welcome continuation mode=%v", state.Mode)
	}

	// YES moves the party inside to (4,2,S); two legal steps reach 89h.
	previousX, previousY = 4, 2
	for _, step := range []dungeonStep{
		{x: 4, y: 1, direction: 0},
		{x: 3, y: 1, direction: 6},
	} {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("court vestibule route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x89 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.court.armor") {
		t.Fatalf("court armor terrain=%02x mode=%v choices=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.currentOriginalChoices, state.Message)
	}
	armorApprovalBefore, armorApprovalFound := state.session.MemoryValue(0x4CBA)
	if !armorApprovalFound {
		t.Fatal("court armor has no approval value")
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"GO UPSTAIRS", "TAKE ARMOR", "ATTACK", "RETREAT"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "myth-drannor.court.go-upstairs"),
			gamePackText(t, state, "myth-drannor.court.take-armor"),
			state.catalog.Text("attack", "攻擊"),
			state.catalog.Text("retreat", "撤退"),
		}) {
		t.Fatalf("court armor choices=%v original=%v", state.Choices, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.court.armor-bows") {
		t.Fatalf("court armor pass message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("court armor continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if approval, found := state.session.MemoryValue(0x4CBA); !found ||
		approval != armorApprovalBefore+1 {
		t.Fatalf("court armor approval=%02x,%v, want %02x+1",
			approval, found, armorApprovalBefore)
	}

	// GO UPSTAIRS places the party at (2,1,W). Walk around the inner stair
	// to terrain 8Ah, then south to Queen Daemir's spirit at terrain 8Bh.
	courtRoute := []dungeonStep{
		{x: 2, y: 2, direction: 4},
		{x: 1, y: 2, direction: 6},
	}
	previousX, previousY = 2, 1
	for _, step := range courtRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("inner-court route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8A || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.court.greeting") {
		t.Fatalf("court greeting terrain=%02x mode=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("court greeting continuation mode=%v", state.Mode)
	}

	if !burialGlen.CanMoveDungeonWrapped(1, 2, 4) {
		t.Fatal("court greeting-to-queen route is not passable")
	}
	gemsBeforeCourtReward, jewelryBeforeCourtReward := state.TreasurePool()
	itemsBeforeCourtReward := len(state.PendingTreasureItems())
	state.SetDungeonGeometryView(1, 3, 4)
	state.DungeonWallRoof = burialGlen.CellWrapped(1, 3).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x8B || state.Mode != ModeEvent ||
		state.PictureBlock != 72 ||
		state.Message != gamePackText(t, state, "myth-drannor.court.reward") {
		t.Fatalf("court reward terrain=%02x mode=%v picture=%d message=%q",
			state.DungeonWallRoof, state.Mode, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !state.treasureMenu && state.Mode == ModeWilderness &&
		reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	for step := 0; step < 4 && !state.treasureMenu; step++ {
		switch state.Mode {
		case ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case ModeWilderness:
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected court reward continuation mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	gemsAfterCourtReward, jewelryAfterCourtReward := state.TreasurePool()
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		len(state.currentOriginalChoices) == 0 ||
		state.currentOriginalChoices[len(state.currentOriginalChoices)-1] != "TREASURE_EXIT" ||
		gemsAfterCourtReward != gemsBeforeCourtReward+12 ||
		jewelryAfterCourtReward != jewelryBeforeCourtReward+8 ||
		len(state.PendingTreasureItems()) != itemsBeforeCourtReward+6 {
		t.Fatalf("court reward treasure mode=%v choices=%v original=%v pending=%d message=%q",
			state.Mode, state.Choices, state.currentOriginalChoices,
			len(state.PendingTreasureItems()), state.Message)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent ||
		state.Message != gamePackText(t, state, "myth-drannor.court.farewell") {
		t.Fatalf("court farewell mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode == ModeWilderness &&
		reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("court completed mode=%v message=%q", state.Mode, state.Message)
	}
	for _, address := range []uint16{0x4CC4, 0x4CC5, 0x4CC6} {
		if value, found := state.session.MemoryValue(address); !found || value != 1 {
			t.Fatalf("court completion flag %04x=%d,%v", address, value, found)
		}
	}

	// Leave the court by its real doorway and cross the northern ruins to
	// terrain 05h. The court is a side chamber, not a chapter exit.
	redPlumeRoute := []dungeonStep{
		{x: 1, y: 2, direction: 0},
		{x: 1, y: 1, direction: 0},
		{x: 2, y: 1, direction: 2},
		{x: 3, y: 1, direction: 2},
		{x: 4, y: 1, direction: 2},
		{x: 4, y: 2, direction: 4},
		{x: 5, y: 2, direction: 2},
		{x: 6, y: 2, direction: 2},
		{x: 7, y: 2, direction: 2},
		{x: 7, y: 3, direction: 4},
		{x: 8, y: 3, direction: 2},
		{x: 9, y: 3, direction: 2},
		{x: 10, y: 3, direction: 2},
		{x: 11, y: 3, direction: 2},
		{x: 12, y: 3, direction: 2},
		{x: 12, y: 4, direction: 4},
		{x: 12, y: 5, direction: 4},
		{x: 12, y: 6, direction: 4},
		{x: 13, y: 6, direction: 2},
	}
	previousX, previousY = 1, 3
	for index, step := range redPlumeRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("red-plume route %d (%d,%d)->(%d,%d) is not passable",
				index, previousX, previousY, step.x, step.y)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(redPlumeRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(redPlumeRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("red-plume route %d interrupted mode=%v terrain=%02x message=%q",
				index, state.Mode, state.DungeonWallRoof, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	// ⚠ 旁白依距離挑（spec 1144）：這一處的距離上限是 2，抵達時演的是遠距那一句。
	if state.DungeonWallRoof != 0x05 || state.Mode != ModeWilderness ||
		state.Prompt != gamePackText(t, state, "myth-drannor.lone-red-plume-spotted") {
		t.Fatalf("red-plume arrival terrain=%02x mode=%v prompt=%q choices=%v",
			state.DungeonWallRoof, state.Mode, state.Prompt, state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.red-plume.journal") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"AGREE", "REFUSE PAYMENT", "DISAGREE"}) {
		t.Fatalf("red-plume journal mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	if journals := state.JournalPages; len(journals) == 0 ||
		journals[len(journals)-1] != gamePackText(t, state, "journal.33") {
		t.Fatalf("Journal 33 was not unlocked from game-pack: %v", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.red-plume.warning") {
		t.Fatalf("red-plume warning=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("red-plume continue choices=%v", state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.PictureBlock != 0x43 ||
		state.Message != gamePackText(t, state, "myth-drannor.red-plume.reveal") {
		t.Fatalf("red-plume reveal picture=%d message=%q", state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	// 兩支箭是 `2Eh` 的「連打 N 下」形式（旗標 bit 7 清空 ⇒ 整個 byte 是次數，
	// spec 1152）。正式路徑現在自己就會結算它，不需要測試另外呼叫一次。
	if len(state.pendingDamageRequests) != 0 {
		t.Fatalf("red-plume 兩支箭沒有在正式路徑結算：%#v", state.pendingDamageRequests)
	}
	for step := 0; step < 4 && state.Mode == ModeWilderness &&
		reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}); step++ {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 7 {
		t.Fatalf("red-plume combat mode=%v active=%v enemies=%d message=%q choices=%v pending-damage=%v",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideEnemy)), state.Message,
			state.currentOriginalChoices, state.pendingDamageRequests)
	}
	arrowHP := state.partyRoster[0].HitPoints
	if arrowHP >= state.partyRoster[0].MaxHitPoints {
		t.Fatalf("red-plume arrows did not create a real Cure target: hp=%d/%d",
			arrowHP, state.partyRoster[0].MaxHitPoints)
	}
	if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
		t.Fatalf("red-plume ALT+M enabled=%v err=%v", enabled, err)
	}
	if err := state.CombatQuick(); err != nil {
		t.Fatal(err)
	}
	pendingHero, _ := state.fighter("hero")
	if pendingHero.CombatAction.SpellID != CureLightWoundsSpellID ||
		pendingHero.CombatAction.TargetID != "hero" {
		t.Fatalf("red-plume Quick Cure pending=%+v slots=%v",
			pendingHero.CombatAction, state.partyRoster[0].SpellSlots)
	}
	state.CombatManualControl()
	for action := 0; action < 8 && state.Mode == ModeCombat; action++ {
		pendingHero, _ = state.fighter("hero")
		if pendingHero.CombatAction.SpellID == 0 {
			break
		}
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	pendingHero, _ = state.fighter("hero")
	if pendingHero.CombatAction.SpellID != 0 || pendingHero.CombatAction.TargetID != "" ||
		len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("red-plume normal-path Quick Cure did not consume its pending action: action=%+v slots=%v hp=%d->%d",
			pendingHero.CombatAction, state.partyRoster[0].SpellSlots, arrowHP, state.partyRoster[0].HitPoints)
	}
	for action := 0; action < 96 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.grave.skeleton") {
		t.Fatalf("red-plume victory mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("red-plume grave GO mode=%v message=%q", state.Mode, state.Message)
	}

	// The mainline leaves Burial Glen by walking east to the real wrapped
	// boundary. A normal movement attempt at x=15 invokes entry 0; it is not
	// a terrain event or a hard-coded coordinate switch.
	burialExitRoute := []dungeonStep{
		{x: 14, y: 6, direction: 2},
		{x: 15, y: 6, direction: 2},
	}
	previousX, previousY = 13, 6
	for _, step := range burialExitRoute {
		if !burialGlen.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("Burial Glen exit route (%d,%d)->(%d,%d) is not passable",
				previousX, previousY, step.x, step.y)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = burialGlen.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("Burial Glen exit route cell (%d,%d) mode=%v choices=%v",
				step.x, step.y, state.Mode, state.currentOriginalChoices)
		}
		previousX, previousY = step.x, step.y
	}
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.more-ruins") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PATH", "WOODS", "TURN BACK"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "myth-drannor.path"),
			gamePackText(t, state, "myth-drannor.woods"),
			gamePackText(t, state, "myth-drannor.turn-back"),
		}) {
		t.Fatalf("more-ruins menu mode=%v message=%q choices=%v originals=%v",
			state.Mode, state.Message, state.Choices, state.currentOriginalChoices)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x40 {
		t.Fatalf("more-ruins TURN BACK mode=%v block=%02x",
			state.Mode, state.session.CurrentBlockID())
	}
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x42 ||
		state.DungeonX != 0 || state.DungeonY != 12 ||
		state.Message != gamePackText(t, state, "myth-drannor.helm-north") {
		t.Fatalf("more-ruins PATH block=%02x position=(%d,%d) mode=%v message=%q",
			state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.Mode, state.Message)
	}
	for step := 0; step < 4 && state.Mode != ModeDungeon; step++ {
		switch {
		case state.Mode == ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case state.Mode == ModeWilderness &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}):
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected outer-ruins entry mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	var outerRuins geo.Grid
	found = false
	for _, block := range blocks {
		if block.Entry.ID == 0x42 {
			outerRuins, err = geo.Parse(block.Entry.ID, block.Data)
			if err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("GEO6 block 0x42 not found")
	}
	if state.Mode != ModeDungeon ||
		!outerRuins.CanMoveDungeonWrapped(0, 12, 2) {
		t.Fatalf("outer-ruins spawn mode=%v cannot move east", state.Mode)
	}
	state.SetDungeonGeometryView(1, 12, 2)
	state.DungeonWallRoof = outerRuins.CellWrapped(1, 12).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x01 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.greeting") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"WAIT", "ATTACK", "FLEE"}) {
		t.Fatalf("Tirsheya intro terrain=%02x mode=%v message=%q choices=%v",
			state.DungeonWallRoof, state.Mode, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.tale") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("Tirsheya tale message=%q choices=%v",
			state.Message, state.currentOriginalChoices)
	}
	if journals := state.JournalPages; len(journals) == 0 ||
		journals[len(journals)-1] != gamePackText(t, state, "journal.5") {
		t.Fatalf("Journal 5 was not unlocked from game-pack: %v", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.guards") {
		t.Fatalf("Tirsheya guard message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 10 {
		t.Fatalf("Tirsheya first combat mode=%v active=%v enemies=%d",
			state.Mode, state.CombatActive(), len(state.livingBySide(combat.SideEnemy)))
	}
	for action := 0; action < 128 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.beyrha-arrives") {
		t.Fatalf("Beyrha arrival mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.ultimatum") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"TIRSHEYA", "BEYRHA", "FLEE"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			gamePackText(t, state, "myth-drannor.outer.tirsheya.name"),
			gamePackText(t, state, "myth-drannor.outer.beyrha.name"),
			gamePackText(t, state, "option.flee"),
		}) {
		t.Fatalf("Beyrha ultimatum mode=%v message=%q choices=%v originals=%v",
			state.Mode, state.Message, state.Choices, state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.tirsheya.attack-beyrha") {
		t.Fatalf("attack Beyrha message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 12 ||
		len(state.livingBySide(combat.SideParty)) != 2 {
		t.Fatalf("Beyrha combat mode=%v active=%v party=%d enemies=%d",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideParty)),
			len(state.livingBySide(combat.SideEnemy)))
	}
	temporaryAllies := make([]combat.Fighter, 0, 1)
	for _, fighter := range state.livingBySide(combat.SideParty) {
		if fighter.TemporaryAlly {
			temporaryAllies = append(temporaryAllies, fighter)
		}
	}
	if len(temporaryAllies) != 1 || !temporaryAllies[0].QuickFight ||
		temporaryAllies[0].Name != monsterRecords[0x43].Name ||
		len(state.partyRoster) != 1 {
		t.Fatalf("Beyrha temporary ally=%+v persistent roster=%+v",
			temporaryAllies, state.partyRoster)
	}
	beforeAllyBattle, err := state.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	beforeAllySession, err := state.session.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	allySavePath := filepath.Join(t.TempDir(), "outer-ruins-temporary-ally-combat.json")
	if err := state.SavePartyFile(allySavePath); err != nil {
		t.Fatal(err)
	}
	allyLoaded := NewStateFromECLBlocks(trainingTestCatalog(t), all, 0x50)
	allyLoaded.SetMonsterRecordsForECL(6, monsterRecords)
	allyLoaded.SetMonsterAffectsForECL(6, monsterAffects)
	allyLoaded.SetTreasureItemBlocks(treasureBlocks)
	if err := allyLoaded.LoadPartyFile(allySavePath); err != nil {
		t.Fatal(err)
	}
	afterAllyBattle, err := allyLoaded.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	afterAllySession, err := allyLoaded.session.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	loadedTemporaryAllies := make([]combat.Fighter, 0, 1)
	for _, fighter := range allyLoaded.livingBySide(combat.SideParty) {
		if fighter.TemporaryAlly {
			loadedTemporaryAllies = append(loadedTemporaryAllies, fighter)
		}
	}
	if !reflect.DeepEqual(afterAllyBattle, beforeAllyBattle) ||
		!reflect.DeepEqual(afterAllySession, beforeAllySession) ||
		!reflect.DeepEqual(loadedTemporaryAllies, temporaryAllies) ||
		len(allyLoaded.partyRoster) != 1 {
		t.Fatalf("temporary-ally restore battleEqual=%v sessionEqual=%v allies=%+v roster=%+v",
			reflect.DeepEqual(afterAllyBattle, beforeAllyBattle),
			reflect.DeepEqual(afterAllySession, beforeAllySession),
			loadedTemporaryAllies, allyLoaded.partyRoster)
	}
	state = allyLoaded
	for action := 0; action < 160 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	persistentParty := state.PartyFighters()
	if completed, ok := state.session.MemoryValue(0x4CD1); !ok || completed != 1 ||
		state.Mode != ModeDungeon || len(state.partyRoster) != 1 ||
		len(persistentParty) != 1 || persistentParty[0].TemporaryAlly {
		t.Fatalf("Tirsheya alliance completion mode=%v party=%d 4CD1=%d,%v message=%q",
			state.Mode, len(state.partyRoster), completed, ok, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("completed Tirsheya event retriggered mode=%v message=%q",
			state.Mode, state.Message)
	}
	if searched, ok := state.session.MemoryValue(0x4CD2); ok && searched != 0 {
		t.Fatalf("storehouse search flag was set before entering: 4CD2=%d", searched)
	}

	storehouseRoute := []struct {
		x, y      int
		direction int
	}{
		{2, 12, 2},
		{3, 12, 2},
		{3, 13, 4},
		{3, 14, 4},
	}
	previousX, previousY = 1, 12
	for _, step := range storehouseRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, step.direction) {
			t.Fatalf("illegal storehouse route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, uint8(step.direction))
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("storehouse route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x02 {
		t.Fatalf("storehouse entrance terrain=%02x, want 02", state.DungeonWallRoof)
	}
	if !outerRuins.CanMoveDungeonWrapped(3, 14, 6) {
		t.Fatal("storehouse entrance cannot legally move west into the building")
	}
	state.SetDungeonGeometryView(2, 14, 6)
	state.DungeonWallRoof = outerRuins.CellWrapped(2, 14).Terrain
	if state.DungeonX != 2 || state.DungeonY != 14 {
		t.Fatalf("storehouse geometry adapter position=(%d,%d)", state.DungeonX, state.DungeonY)
	}
	if transient, ok := state.session.MemoryValue(0x7F81); !ok || transient != 0 {
		t.Fatalf("storehouse movement did not clear transient 7F81=%d,%v", transient, ok)
	}
	if boundary, ok := state.session.MemoryValue(0x7ED5); !ok || boundary != 0 {
		t.Fatalf("storehouse movement did not clear boundary flag 7ED5=%d,%v", boundary, ok)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.DungeonWallRoof != 0x83 ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.storehouse.supplies") {
		searched, ok := state.session.MemoryValue(0x4CD2)
		t.Fatalf("storehouse interior block=%02x mode=%v terrain=%02x message=%q 4CD2=%d,%v",
			state.session.CurrentBlockID(), state.Mode, state.DungeonWallRoof,
			state.Message, searched, ok)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("storehouse description return mode=%v", state.Mode)
	}
	beforeMoney := state.MoneyPool()
	beforeGems, beforeJewelry := state.TreasurePool()
	beforeItems := len(state.PendingTreasureItems())
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.storehouse.valuables") ||
		state.MoneyPool() != beforeMoney+9500 {
		t.Fatalf("storehouse treasure mode=%v menu=%v money=%d/%d message=%q",
			state.Mode, state.treasureMenu, state.MoneyPool(), beforeMoney,
			state.Message)
	}
	storehouseGems, storehouseJewelry := state.TreasurePool()
	if storehouseGems != beforeGems+8 || storehouseJewelry != beforeJewelry+8 ||
		len(state.PendingTreasureItems()) != beforeItems+2 {
		t.Fatalf("storehouse treasure gems=%d/%d jewelry=%d/%d items=%d/%d",
			storehouseGems, beforeGems, storehouseJewelry, beforeJewelry,
			len(state.PendingTreasureItems()), beforeItems)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("storehouse treasure return mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	repeatGems, repeatJewelry := state.TreasurePool()
	if state.Mode != ModeDungeon || state.MoneyPool() != beforeMoney+9500 ||
		repeatGems != storehouseGems || repeatJewelry != storehouseJewelry ||
		len(state.PendingTreasureItems()) != 0 {
		t.Fatalf("repeated storehouse search mode=%v money=%d gems=%d jewelry=%d items=%d",
			state.Mode, state.MoneyPool(), repeatGems, repeatJewelry,
			len(state.PendingTreasureItems()))
	}

	fugitiveRoute := []struct {
		x, y      int
		direction int
	}{
		{3, 14, 2},
		{3, 13, 0},
		{3, 12, 0},
		{3, 11, 0},
		{3, 10, 0},
		{3, 9, 0},
		{3, 8, 0},
		{3, 7, 0},
	}
	previousX, previousY = 2, 14
	for index, step := range fugitiveRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, step.direction) {
			t.Fatalf("illegal fugitive route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, uint8(step.direction))
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(fugitiveRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("fugitive route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x04 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.fugitive.intro") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("fugitive intro mode=%v terrain=%02x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 6 {
		t.Fatalf("fugitive rescue combat mode=%v active=%v enemies=%d",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideEnemy)))
	}
	for action := 0; action < 96 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent || !state.PictureRequested ||
		state.PictureBlock != 0x40 || state.SceneHeadBlock != 0x40 ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.fugitive.clue") {
		t.Fatalf("fugitive clue mode=%v picture=%v/%02x head=%02x message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("fugitive clue continuation mode=%v choices=%v",
			state.Mode, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("fugitive clue return mode=%v message=%q", state.Mode, state.Message)
	}
	for address, want := range map[uint16]uint16{
		0x4CD3: 1,
		0x4CD4: 1,
		0x4CD5: 1,
	} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("fugitive memory[%04X]=%d,%v want %d,true",
				address, got, ok, want)
		}
	}

	cacheRoute := []struct {
		x, y      int
		direction int
	}{
		{3, 6, 0},
		{2, 6, 6},
		{1, 6, 6},
		{0, 6, 6},
		{15, 6, 6},
		{15, 5, 0},
		{15, 4, 0},
		{15, 3, 0},
		{14, 3, 6},
	}
	previousX, previousY = 3, 7
	for _, step := range cacheRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, step.direction) {
			t.Fatalf("illegal cache route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, uint8(step.direction))
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("cache route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x06 {
		t.Fatalf("cache terrain=%02x, want 06", state.DungeonWallRoof)
	}
	cacheMoney := state.MoneyPool()
	cacheRemainder := state.MoneyPoolCopperRemainder()
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.fugitive.cache") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("cache discovery mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	if state.pendingTreasureMessage != state.Message {
		t.Fatalf("cache pending treasure message=%q, want %q",
			state.pendingTreasureMessage, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	cacheItems := state.PendingTreasureItems()
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.fugitive.cache") ||
		state.MoneyPool() != cacheMoney ||
		state.MoneyPoolCopperRemainder() != cacheRemainder+100 ||
		len(cacheItems) != 3 ||
		// 順序是 ITEM 區塊順序的**反序**：原作把每一件都前插到 `DS:6F8Ch`
		// 這條單鏈的最前面（`overlay-02:1C7Ch`），而顯示端從鏈頭走
		// （`overlay-05:0CF5h` 的 `PRINTITEMNAME` 迴圈）。spec 1151。
		cacheItems[0].Type != 0x24 || cacheItems[0].Plus != 5 ||
		cacheItems[1].Type != 0x41 || cacheItems[1].Plus != 1 ||
		cacheItems[2].Type != 0x3F || cacheItems[2].Plus != 2 {
		t.Fatalf("cache treasure mode=%v menu=%v money=%d/%d remainder=%d/%d items=%+v message=%q",
			state.Mode, state.treasureMenu, state.MoneyPool(), cacheMoney,
			state.MoneyPoolCopperRemainder(), cacheRemainder, cacheItems,
			state.Message)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("cache treasure return mode=%v", state.Mode)
	}
	if clue, ok := state.session.MemoryValue(0x4CD5); !ok || clue != 0 {
		t.Fatalf("cache clue after treasure=%d,%v want 0,true", clue, ok)
	}
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.MoneyPool() != cacheMoney ||
		state.MoneyPoolCopperRemainder() != cacheRemainder+100 ||
		len(state.PendingTreasureItems()) != 0 {
		t.Fatalf("cache repeat mode=%v money=%d remainder=%d items=%d",
			state.Mode, state.MoneyPool(), state.MoneyPoolCopperRemainder(),
			len(state.PendingTreasureItems()))
	}

	namelessRoute := []dungeonStep{
		{x: 15, y: 3, direction: 2},
		{x: 15, y: 4, direction: 4},
		{x: 15, y: 5, direction: 4},
		{x: 15, y: 6, direction: 4},
		{x: 15, y: 7, direction: 4},
		{x: 14, y: 7, direction: 6},
		{x: 13, y: 7, direction: 6},
	}
	previousX, previousY = 14, 3
	for index, step := range namelessRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal Nameless route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(namelessRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(namelessRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("Nameless route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x07 || state.Mode != ModeEvent ||
		state.PictureBlock != 0x46 || state.SceneHeadBlock != 0x43 ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.nameless-warning") {
		t.Fatalf("Nameless warning mode=%v terrain=%02x picture=%02x/%02x message=%q",
			state.Mode, state.DungeonWallRoof, state.PictureBlock,
			state.SceneHeadBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Nameless continuation mode=%v choices=%v",
			state.Mode, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Nameless return mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Nameless warning retriggered mode=%v message=%q",
			state.Mode, state.Message)
	}

	brushRoute := []dungeonStep{
		{x: 13, y: 8, direction: 4},
		{x: 13, y: 9, direction: 4},
		{x: 12, y: 9, direction: 6},
	}
	previousX, previousY = 13, 7
	for index, step := range brushRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal brush route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(brushRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(brushRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("brush route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x08 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.brush-decoy") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("brush decoy mode=%v terrain=%02x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.brush-rescue") ||
		state.DungeonX != 11 || state.DungeonY != 10 ||
		state.DungeonDirection != 4 {
		t.Fatalf("brush rescue mode=%v position=(%d,%d,%d) message=%q choices=%v",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.brush-ambush") {
		t.Fatalf("brush ambush mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.brush-attack") ||
		len(state.pendingDamageRequests) != 0 {
		t.Fatalf("brush rocks mode=%v message=%q pending=%+v choices=%v",
			state.Mode, state.Message, state.pendingDamageRequests,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 11 {
		t.Fatalf("brush combat mode=%v active=%v enemies=%d message=%q",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideEnemy)), state.Message)
	}
	for action := 0; action < 192 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("brush victory mode=%v status=%v enemies=%d visual=%v message=%q",
			state.Mode, state.CombatStatus(),
			len(state.livingBySide(combat.SideEnemy)),
			state.CombatVisualPending(), state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("brush victory continuation mode=%v message=%q",
			state.Mode, state.Message)
	}
	if flag, ok := state.session.MemoryValue(0x4CD6); !ok || flag != 1 {
		t.Fatalf("brush completion 4CD6=%d,%v", flag, ok)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("brush ambush retriggered mode=%v message=%q",
			state.Mode, state.Message)
	}

	bloodstainRoute := []dungeonStep{
		{x: 11, y: 11, direction: 4},
	}
	previousX, previousY = 11, 10
	for index, step := range bloodstainRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal bloodstain route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(bloodstainRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(bloodstainRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("bloodstain route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x09 || state.Mode != ModeEvent ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.brush-bloodstains") {
		t.Fatalf("bloodstains mode=%v terrain=%02x message=%q",
			state.Mode, state.DungeonWallRoof, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("bloodstains return mode=%v", state.Mode)
	}

	rakshasaResidenceRoute := []dungeonStep{
		{x: 11, y: 10, direction: 0},
		{x: 12, y: 10, direction: 2},
		{x: 12, y: 9, direction: 0},
		{x: 12, y: 8, direction: 0},
		{x: 12, y: 7, direction: 0},
		{x: 11, y: 7, direction: 6},
		{x: 11, y: 6, direction: 0},
		{x: 10, y: 6, direction: 6},
		{x: 9, y: 6, direction: 6},
	}
	previousX, previousY = 11, 11
	for index, step := range rakshasaResidenceRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal rakshasa-residence route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(rakshasaResidenceRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(rakshasaResidenceRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("rakshasa-residence route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8D || state.Mode != ModeWilderness ||
		state.Prompt != gamePackText(t, state, "myth-drannor.outer.rakshasa-residence") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
		t.Fatalf("rakshasa residence mode=%v terrain=%02x prompt=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Prompt,
			state.currentOriginalChoices)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK",
			"PARLAY_NICE", "PARLAY_ABUSIVE"}) ||
		!reflect.DeepEqual(state.Choices, []string{
			state.catalog.Text("parlay_haughty", ""), state.catalog.Text("parlay_sly", ""),
			state.catalog.Text("parlay_meek", ""), state.catalog.Text("parlay_nice", ""),
			state.catalog.Text("parlay_abusive", ""),
		}) {
		t.Fatalf("rakshasa parlay originals=%v localized=%v", state.currentOriginalChoices, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.rakshasa-parlay") {
		t.Fatalf("rakshasa parlay mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	if journals := state.JournalPages; len(journals) == 0 ||
		journals[len(journals)-1] != gamePackText(t, state, "journal.57") {
		t.Fatalf("Journal 57 was not unlocked from game-pack: %v", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("rakshasa parlay return mode=%v message=%q", state.Mode, state.Message)
	}

	doorwayTrapRoute := []dungeonStep{
		{x: 10, y: 6, direction: 2},
		{x: 10, y: 5, direction: 0},
		{x: 10, y: 4, direction: 0},
		{x: 10, y: 3, direction: 0},
		{x: 9, y: 3, direction: 6},
		{x: 9, y: 2, direction: 0},
	}
	previousX, previousY = 9, 6
	for index, step := range doorwayTrapRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal doorway-trap route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(doorwayTrapRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(doorwayTrapRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("doorway-trap route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x0B || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.margoyle-trap") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("doorway trap mode=%v terrain=%02x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.margoyle-collapse") ||
		state.DungeonX != 10 || state.DungeonY != 2 || state.DungeonDirection != 0 {
		t.Fatalf("doorway collapse mode=%v position=(%d,%d,%d) message=%q",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.margoyle-rakshasa") ||
		len(state.pendingDamageRequests) != 0 {
		t.Fatalf("doorway buried mode=%v message=%q pending=%+v",
			state.Mode, state.Message, state.pendingDamageRequests)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.margoyle-surprise") ||
		len(state.livingBySide(combat.SideEnemy)) != 1 {
		t.Fatalf("doorway combat mode=%v active=%v enemies=%d message=%q",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideEnemy)), state.Message)
	}
	for action := 0; action < 64 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("doorway victory mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}

	gamblingRoute := []dungeonStep{
		{x: 9, y: 2, direction: 6},
		{x: 8, y: 2, direction: 6},
		{x: 7, y: 2, direction: 6},
		{x: 6, y: 2, direction: 6},
		{x: 5, y: 2, direction: 6},
		{x: 4, y: 2, direction: 6},
		{x: 3, y: 2, direction: 6},
		{x: 2, y: 2, direction: 6},
		{x: 1, y: 2, direction: 6},
		{x: 1, y: 3, direction: 4},
	}
	previousX, previousY = 10, 2
	for index, step := range gamblingRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal gambling-room route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(gamblingRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(gamblingRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("gambling-room route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8A || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.gambling-room") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("gambling room mode=%v terrain=%02x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.gambling-rise") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("gambling rise mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	gamblingMoney := state.MoneyPool()
	gamblingGems, gamblingJewelry := state.TreasurePool()
	gamblingItems := len(state.PendingTreasureItems())
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(state.livingBySide(combat.SideEnemy)) != 14 {
		t.Fatalf("gambling combat mode=%v active=%v enemies=%d",
			state.Mode, state.CombatActive(),
			len(state.livingBySide(combat.SideEnemy)))
	}
	for action := 0; action < 224 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		state.CombatStatus() != combat.StatusPartyWon ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.gambling-treasure") ||
		state.MoneyPool() != gamblingMoney+11200 {
		t.Fatalf("gambling treasure mode=%v menu=%v status=%v money=%d/%d message=%q",
			state.Mode, state.treasureMenu, state.CombatStatus(),
			state.MoneyPool(), gamblingMoney, state.Message)
	}
	afterGamblingGems, afterGamblingJewelry := state.TreasurePool()
	if afterGamblingGems != gamblingGems+15 ||
		afterGamblingJewelry != gamblingJewelry+9 ||
		len(state.PendingTreasureItems()) != gamblingItems+1 {
		t.Fatalf("gambling treasure gems=%d/%d jewelry=%d/%d items=%d/%d",
			afterGamblingGems, gamblingGems,
			afterGamblingJewelry, gamblingJewelry,
			len(state.PendingTreasureItems()), gamblingItems)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("gambling treasure return mode=%v message=%q", state.Mode, state.Message)
	}

	sewerRoute := []dungeonStep{
		{x: 1, y: 4, direction: 4},
		{x: 2, y: 4, direction: 2},
		{x: 3, y: 4, direction: 2},
		{x: 3, y: 5, direction: 4},
		{x: 3, y: 6, direction: 4},
		{x: 3, y: 7, direction: 4},
		{x: 4, y: 7, direction: 2},
		{x: 5, y: 7, direction: 2},
		{x: 5, y: 6, direction: 0},
	}
	previousX, previousY = 1, 3
	for index, step := range sewerRoute {
		if !outerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("illegal sewer route (%d,%d) direction=%d",
				previousX, previousY, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = outerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		for attempt := 0; index < len(sewerRoute)-1 && attempt < 8 &&
			reflect.DeepEqual(state.currentOriginalChoices,
				[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}); attempt++ {
			if err := state.Select(2); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(sewerRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("sewer route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x0C || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.outer.sewer-margoyle") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("sewer margoyle mode=%v terrain=%02x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.sewer-escape") {
		t.Fatalf("sewer escape message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.sewer-grate") {
		t.Fatalf("sewer grate message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.outer.sewer-warning") {
		t.Fatalf("sewer warning message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x43 ||
		state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.kitchen-arrival") ||
		state.GeoMapSet != 6 || state.GeoMapBlock != 0x43 ||
		state.DungeonX != 15 || state.DungeonY != 15 ||
		state.DungeonDirection != 0 {
		t.Fatalf("kitchen arrival block=%02x mode=%v geo=%d/%02x position=(%d,%d,%d) message=%q",
			state.session.CurrentBlockID(), state.Mode,
			state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("inner-ruins arrival continuation mode=%v message=%q",
			state.Mode, state.Message)
	}
	var innerRuins geo.Grid
	found = false
	for _, block := range blocks {
		if block.Entry.ID == 0x43 {
			innerRuins, err = geo.Parse(block.Entry.ID, block.Data)
			if err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("GEO6 block 0x43 not found")
	}

	innerKitchenRoute := []dungeonStep{
		{x: 15, y: 14, direction: 0},
		{x: 14, y: 14, direction: 6},
		{x: 13, y: 14, direction: 6},
	}
	previousX, previousY = 15, 15
	for _, step := range innerKitchenRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("inner kitchen route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	// 內城前導從下水道進來時就先唸過廚房那一段，並把 `4C07` 設成 1
	// （`ECL6/0x43:00EEh`）。走到廚房格時守衛 `COMPARE AND 4C06,0 4C07,0`
	// 因此不成立，那一格安靜，`4C06` 也還沒被動過。
	kitchenFlag, kitchenSeen := state.session.MemoryValue(0x4C06)
	arrival, arrivalOK := state.session.MemoryValue(0x4C07)
	if kitchenSeen || !arrivalOK || arrival != 1 || state.DungeonWallRoof != 0x8C ||
		state.Mode != ModeDungeon || state.Message != "" {
		t.Fatalf("inner kitchen flag=%d,%v arrival=%d,%v terrain=%02x mode=%v message=%q",
			kitchenFlag, kitchenSeen, arrival, arrivalOK, state.DungeonWallRoof, state.Mode, state.Message)
	}

	innerOfficeRoute := []dungeonStep{
		{x: 13, y: 13, direction: 0},
		{x: 13, y: 12, direction: 0},
		{x: 12, y: 12, direction: 6},
	}
	previousX, previousY = 13, 14
	for _, step := range innerOfficeRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("inner office route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	// 辦公室的一次性標記是內城自己那一份 `4C05`（spec 1162），別的段寫過同一格
	// 也蓋不到這裡，所以這一趟唸得出描述。
	if officeFlag, ok := state.session.MemoryValue(0x4C05); !ok ||
		officeFlag != 1 || state.DungeonWallRoof != 0x8B ||
		state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.office") {
		t.Fatalf("inner office flag=%d,%v terrain=%02x mode=%v message=%q",
			officeFlag, ok, state.DungeonWallRoof, state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}

	beforeBedroomMoney := state.MoneyPool()
	beforeBedroomGems, beforeBedroomJewelry := state.TreasurePool()
	previousX, previousY = 12, 12
	for _, step := range []dungeonStep{
		{x: 11, y: 12, direction: 6},
		{x: 10, y: 12, direction: 6},
	} {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("inner bedroom route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x8A || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.bedroom") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("inner bedroom terrain=%02x mode=%v message=%q choices=%v",
			state.DungeonWallRoof, state.Mode, state.Message,
			state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	afterBedroomGems, afterBedroomJewelry := state.TreasurePool()
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		state.MoneyPool() != beforeBedroomMoney+30000 ||
		afterBedroomGems != beforeBedroomGems+12 ||
		afterBedroomJewelry != beforeBedroomJewelry+15 {
		t.Fatalf("inner bedroom treasure mode=%v menu=%v money=%d/%d gems=%d/%d jewelry=%d/%d",
			state.Mode, state.treasureMenu,
			state.MoneyPool(), beforeBedroomMoney,
			afterBedroomGems, beforeBedroomGems,
			afterBedroomJewelry, beforeBedroomJewelry)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("inner bedroom treasure return mode=%v", state.Mode)
	}

	chapelRoute := []dungeonStep{
		{x: 9, y: 12, direction: 6},
		{x: 9, y: 13, direction: 4},
	}
	previousX, previousY = 10, 12
	for _, step := range chapelRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("inner chapel route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		previousX, previousY = step.x, step.y
	}
	if state.DungeonWallRoof != 0x89 || state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.chapel") {
		t.Fatalf("inner chapel terrain=%02x mode=%v message=%q",
			state.DungeonWallRoof, state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.inner.chapel-priest") {
		t.Fatalf("inner chapel priest message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	chapelEnemies := state.livingBySide(combat.SideEnemy)
	if state.Mode != ModeCombat || !state.CombatActive() ||
		len(chapelEnemies) != 5 ||
		chapelEnemies[0].SpriteBlock != 0x46 ||
		chapelEnemies[4].SpriteBlock != 0x48 {
		t.Fatalf("inner chapel combat mode=%v active=%v enemies=%+v",
			state.Mode, state.CombatActive(), chapelEnemies)
	}
	for action := 0; action < 96 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeEvent ||
		state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("inner chapel victory mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("inner chapel victory continuation mode=%v message=%q",
			state.Mode, state.Message)
	}

	// Return to the bedroom corridor and approach the west-wing choke point.
	// The next legal step after (7,10) is terrain 83h, one of the four tiles
	// that dispatch the Tyranthraxus/Nameless final ritual. Do not bypass it
	// merely to enter the kennel and statuary rooms behind that story gate.
	ritualGateApproach := []dungeonStep{
		{x: 9, y: 12, direction: 0},
		{x: 10, y: 12, direction: 2},
		{x: 10, y: 11, direction: 0},
		{x: 9, y: 11, direction: 6},
		{x: 8, y: 11, direction: 6},
		{x: 8, y: 10, direction: 0},
		{x: 7, y: 10, direction: 6},
	}
	previousX, previousY = 9, 13
	for _, step := range ritualGateApproach {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("ritual-gate route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("ritual-gate approach cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if !innerRuins.CanMoveDungeonWrapped(7, 10, 4) ||
		innerRuins.CellWrapped(7, 11).Terrain != 0x83 {
		t.Fatalf("ritual gate from (7,10) south passable=%v terrain=%02x",
			innerRuins.CanMoveDungeonWrapped(7, 10, 4),
			innerRuins.CellWrapped(7, 11).Terrain)
	}

	state.SetDungeonGeometryView(7, 11, 4)
	state.DungeonWallRoof = innerRuins.CellWrapped(7, 11).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	ritualMessages := []string{
		"myth-drannor.inner.ritual.arrival",
		"myth-drannor.inner.ritual.control",
		"myth-drannor.inner.ritual.journal",
		"myth-drannor.inner.ritual.hand-over",
		"myth-drannor.inner.ritual.dispose-order",
		"myth-drannor.inner.ritual.pool",
		"myth-drannor.inner.ritual.parchment",
		"myth-drannor.inner.ritual.final-spell",
		"myth-drannor.inner.ritual.nameless-reveal",
		"myth-drannor.inner.ritual.nameless-falls",
		"myth-drannor.inner.ritual.bonds-fade",
		"myth-drannor.inner.ritual.recover",
		"myth-drannor.inner.minions-attack",
	}
	ritualPictureStages := map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 9: true}
	for index, messageID := range ritualMessages {
		wantMode := ModeWilderness
		if ritualPictureStages[index] {
			wantMode = ModeEvent
		}
		if state.Mode != wantMode || state.Message != gamePackText(t, state, messageID) {
			t.Fatalf("ritual stage %d mode=%v message=%q, want %q",
				index, state.Mode, state.Message, gamePackText(t, state, messageID))
		}
		if index == 2 {
			wantJournal := gamePackText(t, state, "journal.48")
			found := false
			for _, page := range state.JournalPages {
				found = found || page == wantJournal
			}
			if !found {
				t.Fatalf("Journal 48 not unlocked from game-pack: %v", state.JournalPages)
			}
		}
		var advanceErr error
		if wantMode == ModeEvent {
			advanceErr = state.Continue()
			if advanceErr == nil {
				advanceErr = state.Select(0)
			}
		} else {
			advanceErr = state.Select(0)
		}
		if advanceErr != nil {
			t.Fatalf("ritual stage %d: %v", index, advanceErr)
		}
	}
	ritualEnemies := state.livingBySide(combat.SideEnemy)
	ritualCounts := make(map[uint8]int)
	for _, enemy := range ritualEnemies {
		ritualCounts[enemy.SpriteBlock]++
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		!reflect.DeepEqual(ritualCounts, map[uint8]int{0x44: 6, 0x45: 6, 0x48: 2}) {
		t.Fatalf("ritual combat mode=%v active=%v counts=%v enemies=%+v",
			state.Mode, state.CombatActive(), ritualCounts, ritualEnemies)
	}
	for action := 0; action < 256 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("ritual victory mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("ritual continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if completed, ok := state.session.MemoryValue(0x4C00); !ok || completed != 1 {
		t.Fatalf("ritual completion 4C00=%04X present=%v", completed, ok)
	}

	// The completed ritual makes the adjacent 84h/85h dispatch cells silent,
	// opening the normal GEO route to the two west-wing encounters from round 405.
	westRoute := []dungeonStep{
		{x: 6, y: 11, direction: 6},
		{x: 5, y: 11, direction: 6},
		{x: 5, y: 10, direction: 0},
		{x: 4, y: 10, direction: 6},
		{x: 4, y: 9, direction: 0},
	}
	previousX, previousY = 7, 11
	for index, step := range westRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("west route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if index < len(westRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("west route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.statuary") {
		t.Fatalf("statuary player path mode=%v terrain=%02x message=%q",
			state.Mode, state.DungeonWallRoof, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.inner.minions-attack") {
		t.Fatalf("statuary attack message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	statuaryEnemies := state.livingBySide(combat.SideEnemy)
	if state.Mode != ModeCombat || len(statuaryEnemies) != 10 {
		t.Fatalf("statuary combat mode=%v enemies=%+v", state.Mode, statuaryEnemies)
	}
	for _, enemy := range statuaryEnemies {
		if enemy.SpriteBlock != 0x45 {
			t.Fatalf("statuary enemy=%+v, want MARGOYLE 45h", enemy)
		}
	}
	for action := 0; action < 160 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("statuary victory mode=%v status=%v", state.Mode, state.CombatStatus())
	}
	if err := state.Continue(); err != nil || state.Mode != ModeDungeon {
		t.Fatalf("statuary continuation mode=%v err=%v", state.Mode, err)
	}

	kennelRoute := []dungeonStep{
		{x: 4, y: 10, direction: 4},
		{x: 3, y: 10, direction: 6},
		{x: 2, y: 10, direction: 6},
		{x: 1, y: 10, direction: 6},
		{x: 1, y: 9, direction: 0},
	}
	previousX, previousY = 4, 9
	for index, step := range kennelRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("kennel route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if index == 0 && state.Mode == ModeWilderness {
			if state.Message != gamePackText(t, state, "myth-drannor.inner.helm-northeast") {
				t.Fatalf("inner Helm direction message=%q", state.Message)
			}
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		}
		if index < len(kennelRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("kennel route cell (%d,%d) mode=%v message=%q",
				step.x, step.y, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.kennel") {
		t.Fatalf("kennel player path mode=%v terrain=%02x message=%q",
			state.Mode, state.DungeonWallRoof, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	// The original event deliberately has one empty PRESS pause.
	if state.Mode != ModeWilderness || state.Message != "" {
		t.Fatalf("kennel blank pause mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.inner.minions-attack") {
		t.Fatalf("kennel attack message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	kennelEnemies := state.livingBySide(combat.SideEnemy)
	if state.Mode != ModeCombat || len(kennelEnemies) != 10 {
		t.Fatalf("kennel combat mode=%v enemies=%+v", state.Mode, kennelEnemies)
	}
	for _, enemy := range kennelEnemies {
		if enemy.SpriteBlock != 0x44 {
			t.Fatalf("kennel enemy=%+v, want HELL HOUND 44h", enemy)
		}
	}
	for action := 0; action < 160 && state.Mode == ModeCombat; action++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
		if event, ok := state.CombatVisualEvent(); ok {
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("kennel victory mode=%v status=%v", state.Mode, state.CombatStatus())
	}
	if err := state.Continue(); err != nil || state.Mode != ModeDungeon {
		t.Fatalf("kennel continuation mode=%v err=%v", state.Mode, err)
	}
	if statuaryDone, _ := state.session.MemoryValue(0x4C02); statuaryDone != 1 {
		t.Fatalf("statuary completion 4C02=%04X", statuaryDone)
	}
	if kennelDone, _ := state.session.MemoryValue(0x4C01); kennelDone != 1 {
		t.Fatalf("kennel completion 4C01=%04X", kennelDone)
	}

	// Leave the west wing by a legal first-floor route, then use the original
	// 97h stair transaction to reach the second-floor spawn at (2,5,N).
	stairsRoute := []dungeonStep{
		{x: 1, y: 10, direction: 4},
		{x: 2, y: 10, direction: 2},
		{x: 3, y: 10, direction: 2},
		{x: 4, y: 10, direction: 2},
		{x: 5, y: 10, direction: 2},
		{x: 5, y: 11, direction: 4},
		{x: 6, y: 11, direction: 2},
		{x: 6, y: 10, direction: 0},
		{x: 6, y: 9, direction: 0},
		{x: 6, y: 8, direction: 0},
		{x: 7, y: 8, direction: 2},
		{x: 8, y: 8, direction: 2},
		{x: 8, y: 7, direction: 0},
		{x: 9, y: 7, direction: 2},
		{x: 10, y: 7, direction: 2},
	}
	previousX, previousY = 1, 9
	for index, step := range stairsRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("stairs route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if state.Mode == ModeWilderness && index < len(stairsRoute)-1 {
			if state.Message != "" &&
				state.Message != gamePackText(t, state, "myth-drannor.inner.helm-northeast") &&
				state.Message != gamePackText(t, state, "myth-drannor.inner.minions-attack") &&
				state.Message != gamePackText(t, state, "myth-drannor.inner.bonds-returning") {
				t.Fatalf("stairs route cell (%d,%d) mode=%v message=%q choices=%v originals=%v",
					step.x, step.y, state.Mode, state.Message, state.Choices, state.currentOriginalChoices)
			}
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
			if state.Mode == ModeCombat {
				for action := 0; action < 256 && state.Mode == ModeCombat; action++ {
					if err := state.CombatAct(); err != nil {
						t.Fatal(err)
					}
					if event, ok := state.CombatVisualEvent(); ok {
						if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
							t.Fatal(err)
						}
					}
				}
				if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
					t.Fatalf("stairs-route random combat mode=%v status=%v", state.Mode, state.CombatStatus())
				}
				if err := state.Continue(); err != nil {
					t.Fatal(err)
				}
			}
		}
		if index < len(stairsRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("stairs route cell (%d,%d) mode=%v message=%q",
				step.x, step.y, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	if state.Mode != ModeWilderness ||
		state.Message != gamePackText(t, state, "myth-drannor.inner.stairs-up") {
		t.Fatalf("stairs-up mode=%v terrain=%02x message=%q",
			state.Mode, state.DungeonWallRoof, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 2 || state.DungeonY != 5 ||
		state.DungeonDirection != 0 {
		t.Fatalf("stairs-up destination mode=%v pos=(%d,%d,%d) message=%q",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection, state.Message)
	}

	finalRoute := []dungeonStep{
		{x: 2, y: 4, direction: 0},
		{x: 2, y: 3, direction: 0},
		{x: 2, y: 2, direction: 0},
		{x: 2, y: 1, direction: 0},
		{x: 2, y: 0, direction: 0},
		{x: 3, y: 0, direction: 2},
		{x: 4, y: 0, direction: 2},
		{x: 5, y: 0, direction: 2},
		{x: 6, y: 0, direction: 2},
		{x: 6, y: 1, direction: 4},
	}
	previousX, previousY = 2, 5
	for index, step := range finalRoute {
		if !innerRuins.CanMoveDungeonWrapped(previousX, previousY, int(step.direction)) {
			t.Fatalf("final route (%d,%d)->(%d,%d) direction=%d is not passable",
				previousX, previousY, step.x, step.y, step.direction)
		}
		state.SetDungeonGeometryView(step.x, step.y, step.direction)
		state.DungeonWallRoof = innerRuins.CellWrapped(step.x, step.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if index < len(finalRoute)-1 && state.Mode == ModeWilderness {
			if state.Message != gamePackText(t, state, "myth-drannor.inner.minions-attack") &&
				state.Message != gamePackText(t, state, "myth-drannor.inner.bonds-returning") {
				t.Fatalf("final route boundary (%d,%d) message=%q choices=%v",
					step.x, step.y, state.Message, state.Choices)
			}
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
			if state.Mode == ModeCombat {
				for action := 0; action < 256 && state.Mode == ModeCombat; action++ {
					if err := state.CombatAct(); err != nil {
						t.Fatal(err)
					}
					if event, ok := state.CombatVisualEvent(); ok {
						if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
							t.Fatal(err)
						}
					}
				}
				if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
					t.Fatalf("final-route random combat mode=%v status=%v", state.Mode, state.CombatStatus())
				}
				if err := state.Continue(); err != nil {
					t.Fatal(err)
				}
			}
		}
		if index < len(finalRoute)-1 && state.Mode != ModeDungeon {
			t.Fatalf("final route cell (%d,%d) terrain=%02x mode=%v message=%q",
				step.x, step.y, state.DungeonWallRoof, state.Mode, state.Message)
		}
		previousX, previousY = step.x, step.y
	}
	state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 40 && y >= 0 && y < 25}
	})
	state.EnableCombatVisualTimeline(true)
	for _, messageID := range []string{
		"myth-drannor.inner.final-compulsion",
		"myth-drannor.inner.final-defiance",
		"myth-drannor.inner.final-amulet",
	} {
		if state.Mode != ModeWilderness || state.Message != gamePackText(t, state, messageID) {
			t.Fatalf("final confrontation mode=%v message=%q, want %q",
				state.Mode, state.Message, gamePackText(t, state, messageID))
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	finalEnemies := state.livingBySide(combat.SideEnemy)
	finalCounts := make(map[uint8]int)
	var tyranthraxus *combat.Fighter
	for _, enemy := range finalEnemies {
		finalCounts[enemy.SpriteBlock]++
		if enemy.SpriteBlock == 0x47 {
			copy := enemy
			tyranthraxus = &copy
		}
	}
	if state.Mode != ModeCombat || !state.CombatActive() ||
		!reflect.DeepEqual(finalCounts, map[uint8]int{0x45: 28, 0x47: 1, 0x48: 8}) {
		t.Fatalf("final combat mode=%v active=%v counts=%v enemies=%+v",
			state.Mode, state.CombatActive(), finalCounts, finalEnemies)
	}
	if tyranthraxus == nil || len(tyranthraxus.MonsterAffects) != 6 {
		t.Fatalf("Tyranthraxus MON6SPC effects=%+v", tyranthraxus)
	}
	detectInvisible := false
	magicResistance := false
	innateFireAttack := false
	for _, affect := range tyranthraxus.MonsterAffects {
		if !affect.Innate {
			t.Fatalf("Tyranthraxus effect was suppressed by raw byte 4: %+v", affect)
		}
		if affect.Kind == 0x18 {
			detectInvisible = true
		}
		if affect.Kind == 0x6A {
			magicResistance = true
		}
		if affect.Kind == 0x4F {
			innateFireAttack = true
		}
	}
	if !detectInvisible || !tyranthraxus.MonsterCanDetectInvisible() {
		t.Fatalf("Tyranthraxus detect-invisible projection=%+v", tyranthraxus.MonsterAffects)
	}
	effect19Target := combat.Fighter{MonsterAffects: []combat.MonsterAffect{{Kind: 0x19, Active: true}}}
	effect47Target := combat.Fighter{MonsterAffects: []combat.MonsterAffect{{Kind: 0x47, Active: true}}}
	if !effect19Target.VisibleTo(*tyranthraxus) || effect47Target.VisibleTo(*tyranthraxus) {
		t.Fatalf("Tyranthraxus real MON6 visibility projection: effect19=%v effect47=%v effects=%+v",
			effect19Target.VisibleTo(*tyranthraxus), effect47Target.VisibleTo(*tyranthraxus), tyranthraxus.MonsterAffects)
	}
	if base, ok := tyranthraxus.MonsterMagicResistanceBase(); !magicResistance || !ok || base != 15 {
		t.Fatalf("Tyranthraxus magic-resistance projection base=%d ok=%v effects=%+v",
			base, ok, tyranthraxus.MonsterAffects)
	}
	if !tyranthraxus.MonsterProtectedFromDamage(combat.DamageFlagFire) ||
		!tyranthraxus.MonsterProtectedFromDamage(combat.DamageFlagElectricity) {
		t.Fatalf("Tyranthraxus elemental protection projection=%+v", tyranthraxus.MonsterAffects)
	}
	if !innateFireAttack || len(tyranthraxus.MonsterPostHitAffects(1)) != 1 ||
		len(tyranthraxus.MonsterPostHitAffects(2)) != 1 || len(tyranthraxus.MonsterPostHitAffects(3)) != 0 {
		t.Fatalf("Tyranthraxus post-hit effect projection=%+v", tyranthraxus.MonsterAffects)
	}
	if !tyranthraxus.MonsterThrowsLightning() {
		t.Fatalf("Tyranthraxus effect 84 projection=%+v", tyranthraxus.MonsterAffects)
	}
	boundaryTarget := *tyranthraxus
	boundaryTarget.ID = "tyranthraxus-boundary"
	boundaryTarget.HitPoints = boundaryTarget.MaxHitPoints
	boundaryTarget.HasCombatPosition = true
	boundaryTarget.CombatX, boundaryTarget.CombatY = 4, 1
	// This direct boundary isolates the exact effect-87 elemental protection
	// consumer. Magic-resistance caller ordering for monster spell effects is a
	// separate evidence boundary and is intentionally not part of this oracle.
	boundaryTarget.MagicResistanceRules = nil
	boundaryCaster := combat.Fighter{
		ID: "boundary-mage", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
		HasCombatPosition: true, CombatX: 1, CombatY: 1,
		SavingThrows: []uint8{20, 20, 20, 20, 20},
	}
	boundaryBattle, err := combat.NewBattle([]combat.Fighter{boundaryCaster, boundaryTarget}, 412)
	if err != nil {
		t.Fatal(err)
	}
	fireResult, err := boundaryBattle.CastFireball(
		boundaryCaster.ID, combat.TilePoint{X: 4, Y: 1}, 3,
	)
	if err != nil || len(fireResult.Impacts) != 1 || !fireResult.Impacts[0].Protected ||
		fireResult.Impacts[0].Damage != 0 {
		t.Fatalf("real MON6 fire boundary result=%+v err=%v", fireResult, err)
	}
	lineResult, err := boundaryBattle.CastReflectingLineSpell(
		boundaryCaster.ID, 0x33, combat.TilePoint{X: 4, Y: 1}, 3,
		combat.ReflectingLineOptions{
			WeightedBudget: 8,
			DamageFlags:    combat.DamageFlagElectricity | combat.DamageFlagMagic,
		}, func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x <= 8 && y == 1}
		},
	)
	if err != nil || len(lineResult.Impacts) != 1 || !lineResult.Impacts[0].Protected ||
		lineResult.Impacts[0].Damage != 0 {
		t.Fatalf("real MON6 electricity boundary result=%+v err=%v", lineResult, err)
	}
	boundaryAttacker := *tyranthraxus
	boundaryAttacker.ID = "tyranthraxus-attack-boundary"
	boundaryAttacker.HitPoints = boundaryAttacker.MaxHitPoints
	boundaryAttacker.AttackBonus = 20
	boundaryAttacker.DamageDiceCount = 1
	boundaryAttacker.DamageDiceSides = 1
	boundaryAttacker.DamageBonus = 0
	boundaryVictim := combat.Fighter{
		ID: "boundary-hero", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
		ArmorClass: 10,
	}
	attackBattle, err := combat.NewBattle([]combat.Fighter{boundaryAttacker, boundaryVictim}, 414)
	if err != nil {
		t.Fatal(err)
	}
	attackResult, err := attackBattle.Attack(boundaryAttacker.ID, boundaryVictim.ID)
	if err != nil || !attackResult.Hit || attackResult.Damage != 1 || len(attackResult.Effects) != 1 {
		t.Fatalf("real MON6 post-hit result=%+v err=%v", attackResult, err)
	}
	fireEffect := attackResult.Effects[0]
	if fireEffect.Kind != 0x4F || fireEffect.DamageFlags != combat.DamageFlagFire|combat.DamageFlagMagic ||
		fireEffect.RolledDamage < 2 || fireEffect.RolledDamage > 20 ||
		fireEffect.Damage != fireEffect.RolledDamage || fireEffect.Protected ||
		attackResult.TargetHP != 99-fireEffect.Damage {
		t.Fatalf("real MON6 4F effect=%+v result=%+v", fireEffect, attackResult)
	}
	sawTyranthraxusLightning := false
	for action := 0; action < 1200 && state.Mode == ModeCombat; action++ {
		if event, ok := state.CombatVisualEvent(); ok {
			if event.ActorID == tyranthraxus.ID && event.Kind == combat.VisualLineSpell && event.Effect == "lightning_bolt" {
				sawTyranthraxusLightning = true
			}
			if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if !sawTyranthraxusLightning {
		t.Fatal("normal MON6 final-battle path did not schedule Tyranthraxus effect 84")
	}
	if state.CombatStatus() != combat.StatusPartyWon || !state.GameWon() ||
		state.Mode != ModeWilderness || state.OriginalEvent != "PROGRAM 8" {
		t.Fatalf("PROGRAM 8 game won=%v mode=%v event=%q message=%q choices=%v",
			state.GameWon(), state.Mode, state.OriginalEvent, state.Message, state.Choices)
	}
	// ★ 存檔詢問排在結局過場**之後**（spec 1154）。這條路徑也走一次五頁，
	// 順便證明過場在最終戰結束的那一刻就接得上、每一頁都有譯文。
	for page := 0; state.endingScene; page++ {
		if page >= len(endingSceneKeys) {
			t.Fatalf("結局過場翻不完，停在第 %d 頁", state.endingPageIndex)
		}
		if want := state.catalog.Text(endingSceneKeys[page], ""); want == "" || state.Message != want {
			t.Fatalf("結局第 %d 頁 ＝ %q，預期 %q", page+1, state.Message, want)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("結局第 %d 頁：%v", page+1, err)
		}
	}
	if state.Prompt != state.catalog.Text("program_victory_prompt", "") ||
		len(state.Choices) != 2 ||
		state.Choices[0] != state.catalog.Text("program_victory_save", "") ||
		state.Choices[1] != state.catalog.Text("program_end_without_save", "") ||
		state.Message != state.catalog.Text("program_victory_message", "") {
		t.Fatalf("normal final-battle PROGRAM 8 locale prompt=%q choices=%v message=%q",
			state.Prompt, state.Choices, state.Message)
	}

	boundaryHero := hero
	boundaryHero.AttackBonus = 23
	if err := state.StartCombat([]combat.Fighter{boundaryHero}, []combat.Fighter{{
		ID: "modifier-boundary", Name: "命中邊界", Side: combat.SideEnemy,
		HitPoints: 2, MaxHitPoints: 2, ArmorClass: 30,
		DamageDiceCount: 1, DamageDiceSides: 1,
	}}, 1); err != nil {
		t.Fatal(err)
	}
	if modifier := state.battle.SideAttackRollModifier(combat.SideParty); modifier != 2 {
		t.Fatalf("party battle attack-roll modifier=%d, want 2", modifier)
	}
	result, err := state.battle.ResolveAttack("hero", "modifier-boundary", 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	fighter, _ := state.battle.Fighter("hero")
	if !result.Hit || result.Total != 30 || fighter.AttackBonus != boundaryHero.AttackBonus {
		t.Fatalf("Daemir attack result=%+v fighter=%+v", result, fighter)
	}
}

func gamePackText(t *testing.T, state State, messageID string) string {
	t.Helper()
	if state.dataPack == nil {
		t.Fatal("game pack is unavailable")
	}
	text, ok := state.dataPack.Text(messageID, state.catalog.Language)
	if !ok {
		t.Fatalf("game pack message %q is unavailable for locale %q",
			messageID, state.catalog.Language)
	}
	return text
}
