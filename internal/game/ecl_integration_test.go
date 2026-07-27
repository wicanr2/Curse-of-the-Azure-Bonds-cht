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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
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

func TestRealECL3CallRedrawReachesStateAdapter(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	blocks, err := dax.Parse(zipData(t, image, "ECL3.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var block []byte
	for _, candidate := range blocks {
		if candidate.Entry.ID == 16 {
			block = candidate.Data
			break
		}
	}
	if len(block) == 0 {
		t.Fatal("ECL3 block 16 is absent")
	}
	result, err := ecl.RunSubsetInteractive(block, 0x0198, 40, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CallAddresses) != 1 || result.CallAddresses[0] != 0x2E10 {
		t.Fatalf("real CALL addresses=%#v, want [0x2E10]", result.CallAddresses)
	}
	state := State{}
	state.applyECLCallSignals(result)
	if got := state.ConsumeECLCallRequests(); len(got) != 1 || got[0] != 0x2E10 {
		t.Fatalf("State CALL requests=%#v", got)
	}
}

func TestRealECL1AddNPCBuildsThreePartyMembers(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclBlocks, err := dax.Parse(zipData(t, image, "ECL1.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var opening []byte
	for _, block := range eclBlocks {
		if block.Entry.ID == 0x52 {
			opening = block.Data
			break
		}
	}
	if len(opening) == 0 {
		t.Fatal("ECL1 block 0x52 is absent")
	}
	result, err := ecl.RunSubset(opening, 0x14, 100)
	if err != nil {
		t.Fatal(err)
	}

	records := make(map[uint8]monster.Record)
	monsterBlocks, err := dax.Parse(zipData(t, image, "MON1CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}
	affects := make(map[uint8][]monster.AffectRecord)
	affectBlocks, err := dax.Parse(zipData(t, image, "MON1SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range affectBlocks {
		parsed, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		affects[block.Entry.ID] = parsed
	}
	items := make(map[uint8][]monster.ItemRecord)
	itemBlocks, err := dax.Parse(zipData(t, image, "MON1ITM.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range itemBlocks {
		parsed, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		items[block.Entry.ID] = parsed
	}

	session, err := ecl.NewBlockSession(map[uint8][]byte{0x52: opening}, 0x52)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		session:     session,
		partyRoster: party.Roster{{ID: "pc", Name: "玩家", IconID: 0, HitPoints: 10, MaxHitPoints: 10}},
		party:       []combat.Fighter{{ID: "pc", Name: "玩家", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10}},
	}
	state.SetMonsterRecordsForECL(1, records)
	state.SetMonsterAffectsForECL(1, affects)
	state.SetMonsterItemsForECL(1, items)
	if err := state.applyECLNPCSignals(result); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster) != 4 || len(state.party) != 4 {
		t.Fatalf("party sizes roster=%d fighters=%d", len(state.partyRoster), len(state.party))
	}
	for index, want := range []string{"RUSTLE", "CYNTHIA", "GRENDEL"} {
		character := state.partyRoster[index+1]
		if character.Name != want || !character.NPC || character.ControlMorale != 0xB2 || character.IconID != uint8(index+1) {
			t.Fatalf("NPC %d=%+v", index, character)
		}
		if state.party[index+1].Name != want || state.party[index+1].Side != combat.SideParty {
			t.Fatalf("NPC fighter %d=%+v", index, state.party[index+1])
		}
	}
	if state.whoSelectedIndex != 3 {
		t.Fatalf("selected index=%d, want last NPC", state.whoSelectedIndex)
	}
	state.Mode = ModeEvent
	state.PictureRequested = true
	state.pendingPictureCombat = &result
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("picture continuation did not enter real opening combat: mode=%v event=%q", state.Mode, state.OriginalEvent)
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
