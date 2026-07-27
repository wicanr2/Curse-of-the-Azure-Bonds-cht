package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func TestDestroyItemTypeRemovesReadiedAndStackedUnits(t *testing.T) {
	character := Character{Equipment: []monster.ItemRecord{
		{Type: 0x5E, Readied: true},
		{Type: 0x5E, Count: 3},
		{Type: 0x60, Count: 2},
	}}
	if destroyed := character.DestroyItemType(0x5E); destroyed != 4 {
		t.Fatalf("destroyed=%d, want 4", destroyed)
	}
	if len(character.Equipment) != 1 || character.Equipment[0].Type != 0x60 {
		t.Fatalf("equipment=%#v", character.Equipment)
	}
}
