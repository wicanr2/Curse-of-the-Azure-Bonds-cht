package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

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
	if !fighter.HasPartyIcon || fighter.PartyHeadBlock != 0 || fighter.PartyBodyBlock != 0 || fighter.PartyIconSize != 2 {
		t.Fatalf("fighter icon=%#v", fighter)
	}
}

func TestFighterWithEquipmentAppliesReadiedWeaponAndArmor(t *testing.T) {
	data := make([]byte, monster.BaseItemHeaderSize+3*monster.BaseItemRecordSize)
	// type 0: one d6 weapon with a +1 item bonus
	data[monster.BaseItemHeaderSize+2] = 1
	data[monster.BaseItemHeaderSize+3] = 6
	data[monster.BaseItemHeaderSize+9] = 1
	data[monster.BaseItemHeaderSize+10] = 6
	data[monster.BaseItemHeaderSize+monster.BaseItemRecordSize+0] = 2
	// type 1: body armor with packed AC improvement 2
	data[monster.BaseItemHeaderSize+monster.BaseItemRecordSize+6] = 180
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	character := validCharacter()
	character.Equipment = []monster.ItemRecord{
		{Type: 0, Plus: 1, Readied: true},
		{Type: 1, Readied: true},
		{Type: 2, Readied: false},
	}
	fighter, err := character.FighterWithEquipment(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AttackBonus != 4 || fighter.DamageDiceCount != 1 || fighter.DamageDiceSides != 6 || fighter.DamageBonus != 1 || fighter.ArmorClass != 7 {
		t.Fatalf("equipped fighter=%#v", fighter)
	}
}

func TestEquipItemEnforcesClassSlotsHandsAndRings(t *testing.T) {
	data := make([]byte, monster.BaseItemHeaderSize+7*monster.BaseItemRecordSize)
	setBase := func(index int, slot, hands, mask uint8) {
		offset := monster.BaseItemHeaderSize + index*monster.BaseItemRecordSize
		data[offset] = slot
		data[offset+1] = hands
		data[offset+13] = mask
	}
	setBase(0, 0, 1, 8) // main-hand weapon
	setBase(1, 1, 1, 8) // off-hand item
	setBase(2, 0, 2, 8) // two-handed weapon
	setBase(3, 9, 0, 8)
	setBase(4, 9, 0, 8)
	setBase(5, 9, 0, 8)
	setBase(6, 0, 1, 2) // cleric-only item
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	character := validCharacter()
	character.Equipment = make([]monster.ItemRecord, 7)
	for index := range character.Equipment {
		character.Equipment[index].Type = uint8(index)
	}
	if err := character.EquipItem(0, catalog); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(1, catalog); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(2, catalog); err == nil {
		t.Fatal("two-handed weapon should conflict with off-hand")
	}
	if err := character.UnequipItem(1); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(2, catalog); err == nil {
		t.Fatal("two-handed weapon should conflict with main hand")
	}
	if err := character.UnequipItem(0); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(2, catalog); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(3, catalog); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(4, catalog); err != nil {
		t.Fatal(err)
	}
	if err := character.EquipItem(5, catalog); err == nil {
		t.Fatal("third ring should be rejected")
	}
	if err := character.EquipItem(6, catalog); err == nil {
		t.Fatal("fighter should reject cleric-only item")
	}
}

func TestUnequipCursedAndRemoveItemQuantityRules(t *testing.T) {
	character := validCharacter()
	character.Equipment = []monster.ItemRecord{
		{Type: 1, Count: 3},
		{Type: 2, Readied: true, Cursed: true},
		{Type: 3, Readied: true},
	}
	if err := character.UnequipItem(1); err == nil {
		t.Fatal("cursed readied item should remain locked")
	}
	if err := character.UnequipItem(2); err != nil {
		t.Fatal(err)
	}
	removed, err := character.RemoveItem(0)
	if err != nil || removed.Count != 1 || len(character.Equipment) != 3 || character.Equipment[0].Count != 2 {
		t.Fatalf("stack removal removed=%#v equipment=%#v err=%v", removed, character.Equipment, err)
	}
	if _, err := character.RemoveItem(1); err == nil {
		t.Fatal("readied item should not be removed")
	}
	if _, err := character.RemoveItem(0); err != nil {
		t.Fatal(err)
	}
	if _, err := character.RemoveItem(0); err != nil {
		t.Fatal(err)
	}
	if _, err := character.RemoveItem(1); err != nil {
		t.Fatal(err)
	}
	if len(character.Equipment) != 1 || character.Equipment[0].Type != 2 {
		t.Fatalf("non-stacking removal equipment=%#v", character.Equipment)
	}
}

func TestUseConsumableRemovesScrollAndDecrementsWandCharge(t *testing.T) {
	catalog, err := monster.ParseBaseItems(make([]byte, monster.BaseItemHeaderSize+80*monster.BaseItemRecordSize))
	if err != nil {
		t.Fatal(err)
	}
	character := validCharacter()
	character.Equipment = []monster.ItemRecord{
		{Type: 60, Count: 2, Affects: [3]uint8{3, 0x18, 0}},
		{Type: 78, Readied: true, Affects: [3]uint8{2, 0x5A, 0}},
	}
	use, err := character.UseConsumable(0, catalog)
	if err != nil || use.Kind != monster.ConsumableScroll || len(use.SpellIDs) != 2 || character.Equipment[0].Count != 1 {
		t.Fatalf("scroll use=%#v equipment=%#v err=%v", use, character.Equipment, err)
	}
	use, err = character.UseConsumable(1, catalog)
	if err != nil || use.Kind != monster.ConsumableCharged || use.ChargesBefore != 2 || use.ChargesAfter != 1 || character.Equipment[1].Affects[0] != 1 {
		t.Fatalf("wand use=%#v equipment=%#v err=%v", use, character.Equipment, err)
	}
	if _, err := character.UseConsumable(0, catalog); err != nil {
		t.Fatal(err)
	}
	if len(character.Equipment) != 1 || character.Equipment[0].Type != 78 {
		t.Fatalf("last scroll use equipment=%#v", character.Equipment)
	}
}

func TestRosterFindSpellReturnsFirstCharacterAndSlot(t *testing.T) {
	first := validCharacter()
	first.SpellSlots = []uint8{0x12, 0x24}
	second := validCharacter()
	second.ID = "p2"
	second.SpellSlots = []uint8{0x12}
	match, ok := (Roster{first, second}).FindSpell(0x12)
	if !ok || match.CharacterIndex != 0 || match.SlotIndex != 0 {
		t.Fatalf("match=%#v ok=%t", match, ok)
	}
	if _, ok := (Roster{first}).FindSpell(0x100); ok {
		t.Fatal("spell IDs above byte range should not match")
	}
	if _, ok := (Roster{first}).FindSpell(0x7F); ok {
		t.Fatal("unknown spell should not match")
	}
}

func TestDefaultIconSizeMatchesReferenceRaceSwitch(t *testing.T) {
	for _, test := range []struct {
		race Race
		want uint8
	}{
		{RaceDwarf, 1}, {RaceGnome, 1}, {RaceHalfling, 1},
		{RaceElf, 2}, {RaceHalfElf, 2}, {RaceHuman, 2},
	} {
		if got := DefaultIconSize(test.race); got != test.want {
			t.Errorf("DefaultIconSize(%v)=%d, want %d", test.race, got, test.want)
		}
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

func TestRollAbilitiesIsDeterministicAndWithin3d6(t *testing.T) {
	first := RollAbilities(42)
	second := RollAbilities(42)
	if first != second {
		t.Fatalf("same seed produced different abilities: %#v %#v", first, second)
	}
	for index := 0; index < 6; index++ {
		value, err := first.Value(index)
		if err != nil || value < 3 || value > 18 {
			t.Fatalf("ability[%d]=%d err=%v", index, value, err)
		}
	}
}
