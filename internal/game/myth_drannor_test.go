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
	state := NewStateFromECLBlocks(testCatalog(), all, 0x50)
	hero := combat.Fighter{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
		AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1,
		DamageBonus: 100, AttacksPerTurn: 8, InitiativeBonus: 100,
	}
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
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
