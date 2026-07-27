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
