package dungeon

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestDoorMenuOptionsRespectDetailAndPartyCapabilities(t *testing.T) {
	roster := party.Roster{{ThiefSkills: []uint8{0, 35}, SpellSlots: []uint8{KnockSpellID}}}
	if got := DoorMenuOptionsFor(roster, 2); got != (DoorMenuOptions{Bash: true, Pick: true, Knock: true}) {
		t.Fatalf("detail2 options=%#v", got)
	}
	if got := DoorMenuOptionsFor(roster, 3); got != (DoorMenuOptions{Bash: true, Knock: true}) {
		t.Fatalf("detail3 options=%#v", got)
	}
	if got := DoorMenuOptionsFor(party.Roster{{HitPoints: 1}}, 1); got != (DoorMenuOptions{}) {
		t.Fatalf("unlocked options=%#v", got)
	}
}
