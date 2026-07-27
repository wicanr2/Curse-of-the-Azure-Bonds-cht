package game

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func TestRealECLJourneyReachesBattleWithLoadedParty(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclData := zipData(t, image, "ECL1.DAX")
	eclBlocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	eclByID := make(map[uint8][]byte, len(eclBlocks))
	for _, block := range eclBlocks {
		eclByID[block.Entry.ID] = block.Data
	}

	monsterData := zipData(t, image, "MON1CHA.DAX")
	monsterBlocks, err := dax.Parse(monsterData)
	if err != nil {
		t.Fatal(err)
	}
	records := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}

	state := NewStateFromECLBlocks(testCatalog(), eclByID, 0x51)
	debugSession, err := ecl.NewBlockSession(eclByID, 0x51)
	if err != nil {
		t.Fatal(err)
	}
	debugResult, debugErr := debugSession.RunInteractive(180, []uint16{1, 1})
	if debugErr != nil {
		t.Fatal(debugErr)
	}
	if !debugResult.CombatRequested || len(debugResult.MonsterSpawns) != 0 {
		t.Fatalf("real result=%+v, want COMBAT without spawn descriptors", debugResult)
	}
	state.SetMonsterRecords(records)
	if err := state.SetParty([]combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 5,
		AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 6,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	// The observed ECL1 path is JOURNEY ON, then STORE; the latter reaches
	// the first COMBAT command in block 0x51.
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.OriginalEvent != "PICTURE" || !state.PictureRequested {
		t.Fatalf("journey picture boundary=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "COMBAT" {
		t.Fatalf("real ECL path did not preserve combat boundary: mode=%v event=%q", state.Mode, state.OriginalEvent)
	}
}

func TestRealCrossDAXNEWECLReachesECL1Entry(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocksByID := make(map[uint8][]byte)
	for _, member := range []string{"ECL1.DAX", "ECL4.DAX", "ECL5.DAX"} {
		data := zipData(t, image, member)
		blocks, parseErr := dax.Parse(data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			blocksByID[block.Entry.ID] = block.Data
		}
	}
	session, err := ecl.NewBlockSession(blocksByID, 0x25)
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := session.RunFrom(555, 200, nil)
	if runErr != nil {
		t.Fatal(runErr)
	}
	if session.CurrentBlockID() != 0x50 {
		t.Fatalf("cross-DAX block=0x%02X, want ECL1 block 0x50; result=%+v", session.CurrentBlockID(), result)
	}
	if len(result.Text) == 0 || !strings.Contains(strings.Join(result.Text, " "), "YOU ARE AT THE EDGE OF TILVERTON") {
		t.Fatalf("cross-DAX text=%q, want ECL1 opening event", result.Text)
	}
	if !result.WaitingForMenu || len(result.Menus) != 1 || len(result.Menus[0].Options) != 3 {
		t.Fatalf("cross-DAX result=%+v, want opening menu pause", result)
	}
}

func TestRealECL2EncounterBuildsBattleFromMON2CHA(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclData := zipData(t, image, "ECL2.DAX")
	eclBlocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	var encounterBlock []byte
	for _, block := range eclBlocks {
		if block.Entry.ID == 3 {
			encounterBlock = block.Data
			break
		}
	}
	if len(encounterBlock) == 0 {
		t.Fatal("ECL2 block 3 is absent")
	}
	result, err := ecl.RunSubset(encounterBlock, 0x2B0, 128)
	if err != nil {
		t.Fatal(err)
	}
	if !result.CombatRequested || len(result.MonsterSpawns) != 2 {
		t.Fatalf("ECL2 encounter result=%+v, want COMBAT with two spawn descriptors", result)
	}

	monsterData := zipData(t, image, "MON2CHA.DAX")
	monsterBlocks, err := dax.Parse(monsterData)
	if err != nil {
		t.Fatal(err)
	}
	records := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
		ArmorClass: 5, AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 6, InitiativeBonus: 100,
	}}
	if err := state.StartEncounter(result, records, partyFighters, 37); err != nil {
		t.Fatal(err)
	}
	if !state.CombatActive() || len(state.CombatTargets()) == 0 {
		t.Fatalf("ECL2 battle was not created: fighters=%#v", state.CombatFighters())
	}
}

func zipData(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}
