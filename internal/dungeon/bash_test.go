package dungeon

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestBashDoorUsesDetailThreeStrengthTable(t *testing.T) {
	roster := party.Roster{
		{Abilities: party.Abilities{StrengthFull: 18, StrengthExceptional: 90}},
		{Abilities: party.Abilities{StrengthFull: 23}},
	}
	seenSides := []int{}
	result := BashDoor(roster, 3, func(sides int) int {
		seenSides = append(seenSides, sides)
		if sides == 6 {
			return 5
		}
		return sides
	})
	if !result.Attempted || !result.Opened || result.CharacterIndex != 1 {
		t.Fatalf("result=%#v sides=%v", result, seenSides)
	}
	if len(seenSides) != 1 || seenSides[0] != 6 {
		t.Fatalf("die calls=%v", seenSides)
	}
}

func TestBashDoorStrength18NormalBranchConsumesExtraRoll(t *testing.T) {
	calls := 0
	result := BashDoor(party.Roster{{Abilities: party.Abilities{StrengthFull: 18, StrengthExceptional: 50}}}, 2, func(int) int {
		calls++
		return 6
	})
	if !result.Opened || calls != 1 {
		t.Fatalf("result=%#v calls=%d", result, calls)
	}
}

func TestBashDoorStrength25NeedsNoDie(t *testing.T) {
	calls := 0
	result := BashDoor(party.Roster{{Abilities: party.Abilities{StrengthFull: 25}}}, 2, func(int) int { calls++; return 20 })
	if !result.Opened || calls != 0 {
		t.Fatalf("result=%#v calls=%d", result, calls)
	}
}
