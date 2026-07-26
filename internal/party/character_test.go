package party

import "testing"

func validCharacter() Character {
	return Character{
		ID: "p1", Name: "英雄", Race: RaceHuman, Class: ClassFighter, Level: 1,
		Abilities: Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}
}

func TestValidHumanFighter(t *testing.T) {
	if err := validCharacter().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRaceClassRestrictions(t *testing.T) {
	character := validCharacter()
	character.Race = RaceDwarf
	character.Class = ClassMagicUser
	if err := character.Validate(); err == nil {
		t.Fatal("expected dwarf magic-user to be rejected")
	}
	character.Race = RaceElf
	character.Class = ClassMagicUser
	character.Abilities.Intelligence = 9
	if err := character.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRangerMinimums(t *testing.T) {
	character := validCharacter()
	character.Class = ClassRanger
	character.Abilities.Intelligence = 13
	character.Abilities.Wisdom = 14
	character.Abilities.Constitution = 14
	if err := character.Validate(); err != nil {
		t.Fatal(err)
	}
	character.Abilities.Wisdom = 13
	if err := character.Validate(); err == nil {
		t.Fatal("expected ranger wisdom minimum")
	}
}

func TestAbilityRange(t *testing.T) {
	character := validCharacter()
	character.Abilities.Charisma = 19
	if err := character.Validate(); err == nil {
		t.Fatal("expected ability range error")
	}
}

func TestRosterLimitAndDuplicateIDs(t *testing.T) {
	character := validCharacter()
	if err := (Roster{character, character}).Validate(); err == nil {
		t.Fatal("expected duplicate ID error")
	}
	roster := make(Roster, 7)
	for index := range roster {
		roster[index] = character
		roster[index].ID = string(rune('a' + index))
	}
	if err := roster.Validate(); err == nil {
		t.Fatal("expected roster size error")
	}
}

func TestStarterFighterProjection(t *testing.T) {
	fighter, err := validCharacter().Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if fighter.Side != 0 || fighter.MaxHitPoints != 10 || fighter.DamageDiceSides != 8 {
		t.Fatalf("fighter=%#v", fighter)
	}
}

func TestAbilityAdjustmentIsBounded(t *testing.T) {
	abilities := Abilities{Strength: 3, Intelligence: 10, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10}
	if err := abilities.Adjust(0, -1); err == nil {
		t.Fatal("expected lower bound error")
	}
	if err := abilities.Adjust(0, 1); err != nil || abilities.Strength != 4 {
		t.Fatalf("abilities=%#v err=%v", abilities, err)
	}
}
