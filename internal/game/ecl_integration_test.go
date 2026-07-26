package game

import (
	"archive/zip"
	"io"
	"path/filepath"
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
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "COMBAT" {
		t.Fatalf("real ECL path did not preserve combat boundary: mode=%v event=%q", state.Mode, state.OriginalEvent)
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
