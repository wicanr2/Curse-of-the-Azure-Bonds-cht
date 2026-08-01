package game

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestGameSaveRestoresECLMemoryPCAndRandomContinuation(t *testing.T) {
	// RANDOM; SAVE; PROGRAM 9 is a real engine boundary. A second RANDOM after
	// loading must match uninterrupted execution, including the mutable word.
	program := []byte{
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x00, 0x70,
		0x09, 0x02, 0x34, 0x12, 0x01, 0x01, 0x70,
		0x38, 0x00, 0x09,
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x02, 0x70,
		0x00,
	}
	block := make([]byte, 2+20, 2+20+len(program))
	for index := 0; index < 5; index++ {
		offset := 2 + index*4
		block[offset], block[offset+1], block[offset+2], block[offset+3] = 0x01, 0x02, 0x14, 0x80
	}
	block = append(block, program...)
	blocks := map[uint8][]byte{1: block}
	state := NewStateFromECLBlocks(testCatalog(), blocks, 1)
	// Use a non-default seed. After loading, RunFrom passes its compatibility
	// default, but the restored session-owned stream must remain authoritative.
	state.session.ResetRandomSeed(77)
	hero := combat.Fighter{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10}
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 10, MaxHitPoints: 10,
		Abilities: party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10},
	}}
	boundary, err := state.session.RunFrom(20, 100, nil)
	if err != nil || !boundary.ProgramExit {
		t.Fatalf("boundary=%+v err=%v", boundary, err)
	}
	savePath := filepath.Join(t.TempDir(), "session-v6.json")
	if err := state.SavePartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	want, err := state.session.RunFrom(20, 100, nil)
	if err != nil {
		t.Fatal(err)
	}

	restored := NewStateFromECLBlocks(testCatalog(), blocks, 1)
	if err := restored.LoadPartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	got, err := restored.session.RunFrom(20, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.RandomValues, want.RandomValues) {
		t.Fatalf("restored random=%v, want %v", got.RandomValues, want.RandomValues)
	}
	if value, ok := restored.session.MemoryValue(0x7001); !ok || value != 0x1234 {
		t.Fatalf("restored memory[7001]=%04x,%v", value, ok)
	}
	if restored.session.CurrentBlockID() != 1 {
		t.Fatalf("restored block=%02x", restored.session.CurrentBlockID())
	}
}
