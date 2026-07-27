package dungeon

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"

// BashResult records one reference bash_door transaction.
type BashResult struct {
	Attempted      bool
	Opened         bool
	CharacterIndex int
}

// BashDoor reproduces the verified strength branches in ovr015.bash_door.
// roll receives the die size and returns one result in 1..sides. Detail 3 is
// the unpickable-door branch; all other details use the ordinary table. The
// original routine does not check health before trying each character.
func BashDoor(roster party.Roster, detail uint8, roll func(sides int) int) BashResult {
	result := BashResult{Attempted: true, CharacterIndex: -1}
	if roll == nil {
		return result
	}
	for index, character := range roster {
		strength := character.Abilities.StrengthFull
		if strength == 0 {
			strength = character.Abilities.Strength
		}
		exceptional := character.Abilities.StrengthExceptional
		worked := false
		if detail == 3 {
			switch {
			case strength == 18 && exceptional >= 91 && exceptional <= 99:
				worked = roll(6) == 1
			case strength == 18 && exceptional == 100:
				worked = roll(6) <= 2
			case strength == 19 || strength == 20:
				worked = roll(6) <= 3
			case strength == 21 || strength == 22:
				worked = roll(6) <= 4
			case strength == 23:
				worked = roll(6) <= 5
			case strength == 24:
				worked = roll(8) <= 7
			case strength == 25:
				worked = true
			}
		} else {
			switch {
			case strength >= 3 && strength <= 7:
				worked = roll(6) == 1
			case strength >= 8 && strength <= 15:
				worked = roll(6) <= 2
			case strength == 15 || strength == 17:
				worked = roll(6) <= 3
			case strength == 18 && exceptional >= 0 && exceptional <= 50:
				worked = true
				_ = roll(6) // reference rolls after setting bash_worked
			case strength == 18 && exceptional >= 51 && exceptional <= 99:
				worked = roll(6) <= 4
			case strength == 18 && exceptional == 100:
				worked = roll(6) <= 5
			case strength == 19 || strength == 20:
				worked = roll(8) <= 7
			case strength == 21:
				worked = roll(10) <= 9
			case strength == 22 || strength == 23:
				worked = roll(12) <= 11
			case strength == 24:
				worked = roll(20) <= 19
			case strength == 25:
				worked = true
			}
		}
		if worked {
			result.Opened = true
			result.CharacterIndex = index
			return result
		}
	}
	return result
}
