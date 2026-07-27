package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestECLDestroyItemsMutatesPersistentRoster(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "dark-elf", Equipment: []monster.ItemRecord{
		{Type: 0x5E, Readied: true},
		{Type: 0x23},
	}}}
	state.applyECLInventorySignals(ecl.RunResult{DestroyItemIDs: []uint16{0x5E}})
	if len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 0x23 {
		t.Fatalf("roster=%#v", state.partyRoster[0].Equipment)
	}
}

func TestECLTreasureResolvesMoneyAndItemBlock(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "fighter", Name: "戰士"}}
	state.SetTreasureItemBlocks(map[uint16][]monster.ItemRecord{
		0x02: {{Type: 36, Name: "長劍", Value: 100}},
	})
	state.applyECLTreasureSignals(ecl.RunResult{TreasureRequests: []ecl.TreasureRequest{{
		Coins:     [7]uint16{1, 2, 0, 2, 0, 3, 4},
		ItemBlock: 0x02,
	}}})
	if err := state.ResolveTreasureRequests(); err != nil {
		t.Fatal(err)
	}
	if state.MoneyPool() != 2 {
		t.Fatalf("money pool=%d, want 2 GP", state.MoneyPool())
	}
	if gems, jewelry := state.TreasurePool(); gems != 3 || jewelry != 4 {
		t.Fatalf("treasure pool=(%d,%d), want (3,4)", gems, jewelry)
	}
	if len(state.PendingTreasureItems()) != 1 {
		t.Fatalf("pending items=%#v", state.PendingTreasureItems())
	}
	if err := state.TakeTreasureItem(0, 0); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 36 {
		t.Fatalf("equipment=%#v", state.partyRoster[0].Equipment)
	}
}
