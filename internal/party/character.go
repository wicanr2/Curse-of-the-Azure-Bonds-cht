// Package party contains the data-neutral character-creation contract. It
// follows the six player races and class minimums observed in the CoAB rule
// book/reference rewrite; combat serialization remains a separate adapter.
package party

import "fmt"

type Race uint8

const (
	RaceDwarf Race = iota
	RaceElf
	RaceGnome
	RaceHalfElf
	RaceHalfling
	RaceHuman
)

type Class uint8

const (
	ClassCleric Class = iota
	ClassFighter
	ClassRanger
	ClassPaladin
	ClassMagicUser
	ClassThief
)

type Abilities struct {
	Strength     int
	Intelligence int
	Wisdom       int
	Dexterity    int
	Constitution int
	Charisma     int
}

type Character struct {
	ID        string
	Name      string
	Race      Race
	Class     Class
	Abilities Abilities
	Level     int
}

type Roster []Character

func (r Roster) Validate() error {
	if len(r) == 0 || len(r) > 6 {
		return fmt.Errorf("roster must contain 1..6 player characters")
	}
	seen := make(map[string]struct{}, len(r))
	for _, character := range r {
		if err := character.Validate(); err != nil {
			return err
		}
		if _, exists := seen[character.ID]; exists {
			return fmt.Errorf("duplicate character ID %q", character.ID)
		}
		seen[character.ID] = struct{}{}
	}
	return nil
}

func (r Race) String() string {
	return [...]string{"dwarf", "elf", "gnome", "half-elf", "halfling", "human"}[r]
}

func (c Class) String() string {
	return [...]string{"cleric", "fighter", "ranger", "paladin", "magic-user", "thief"}[c]
}

func (c Character) Validate() error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("character ID and name are required")
	}
	if c.Race > RaceHuman {
		return fmt.Errorf("unsupported race %d", c.Race)
	}
	if c.Class > ClassThief {
		return fmt.Errorf("unsupported class %d", c.Class)
	}
	if c.Level < 1 {
		return fmt.Errorf("character level must be positive")
	}
	values := []int{c.Abilities.Strength, c.Abilities.Intelligence, c.Abilities.Wisdom, c.Abilities.Dexterity, c.Abilities.Constitution, c.Abilities.Charisma}
	for index, value := range values {
		if value < 3 || value > 18 {
			return fmt.Errorf("ability %d=%d is outside 3..18", index, value)
		}
	}
	if !raceAllowsClass(c.Race, c.Class) {
		return fmt.Errorf("%s cannot be a %s", c.Race, c.Class)
	}
	minimums := []struct {
		ability string
		value   int
	}{
		{"wisdom", 9}, {"strength", 9}, {"strength", 13}, {"strength", 12}, {"intelligence", 9}, {"dexterity", 9},
	}
	minimum := minimums[c.Class]
	if abilityValue(c.Abilities, minimum.ability) < minimum.value {
		return fmt.Errorf("%s requires %s >= %d", c.Class, minimum.ability, minimum.value)
	}
	if c.Class == ClassRanger && (c.Abilities.Intelligence < 13 || c.Abilities.Wisdom < 14 || c.Abilities.Constitution < 14) {
		return fmt.Errorf("ranger requires intelligence >= 13, wisdom >= 14, constitution >= 14")
	}
	if c.Class == ClassPaladin && (c.Abilities.Wisdom < 13 || c.Abilities.Charisma < 17) {
		return fmt.Errorf("paladin requires wisdom >= 13 and charisma >= 17")
	}
	return nil
}

func raceAllowsClass(race Race, class Class) bool {
	switch race {
	case RaceHuman, RaceHalfElf:
		return true
	case RaceDwarf, RaceGnome, RaceHalfling:
		return class == ClassFighter || class == ClassThief
	case RaceElf:
		return class == ClassFighter || class == ClassMagicUser || class == ClassThief
	default:
		return false
	}
}

func abilityValue(abilities Abilities, name string) int {
	switch name {
	case "strength":
		return abilities.Strength
	case "intelligence":
		return abilities.Intelligence
	case "wisdom":
		return abilities.Wisdom
	case "dexterity":
		return abilities.Dexterity
	default:
		return 0
	}
}
