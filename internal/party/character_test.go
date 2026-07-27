package party

import (
	"encoding/binary"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
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

func TestCharacterAdvanceEffects(t *testing.T) {
	character := validCharacter()
	character.Effects = []monster.AffectRecord{{Kind: 1, Duration: 3, Value: 3, Strength: 1}}
	if removed := character.AdvanceEffects(2); removed != 0 || character.Effects[0].Duration != 1 {
		t.Fatalf("effects after two minutes=%#v removed=%d", character.Effects, removed)
	}
	if removed := character.AdvanceEffects(1); removed != 1 || len(character.Effects) != 0 {
		t.Fatalf("effects after expiry=%#v removed=%d", character.Effects, removed)
	}
}

func TestShopBuySellAndIdentifyTransactions(t *testing.T) {
	character := validCharacter()
	character.Gold = 500
	item := monster.ItemRecord{Type: 36, Name: "長劍", Readied: true}
	if err := character.BuyItem(item, 125); err != nil {
		t.Fatal(err)
	}
	if character.Gold != 375 || len(character.Equipment) != 1 || character.Equipment[0].Readied {
		t.Fatalf("after buy character=%#v", character)
	}
	sold, err := character.SellItem(0, 75)
	if err != nil {
		t.Fatal(err)
	}
	if sold.Type != 36 || character.Gold != 450 || len(character.Equipment) != 0 {
		t.Fatalf("after sell character=%#v sold=%#v", character, sold)
	}
	if err := character.PayIdentifyFee(); err != nil {
		t.Fatal(err)
	}
	if character.Gold != 250 {
		t.Fatalf("after identify gold=%d, want 250", character.Gold)
	}
}

func TestShopTransactionsProtectReadiedAndGoldOverflow(t *testing.T) {
	character := validCharacter()
	character.Gold = ^uint16(0)
	character.Equipment = []monster.ItemRecord{{Type: 36, Readied: true}}
	if _, err := character.SellItem(0, 1); err == nil || len(character.Equipment) != 1 {
		t.Fatalf("readied/overflow sale should be rejected: character=%#v err=%v", character, err)
	}
	character.Gold = 0
	if err := character.BuyItem(monster.ItemRecord{Type: 36}, 1); err == nil {
		t.Fatal("buy with insufficient gold should be rejected")
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

func TestParseDOSPlayerSpellRecord(t *testing.T) {
	data := make([]byte, DOSKnownSpellsEnd)
	data[DOSMemorizedSpellsOffset] = 7
	data[DOSMemorizedSpellsOffset+1] = 0
	data[DOSMemorizedSpellsOffset+2] = 42
	data[DOSKnownSpellsOffset] = 1
	data[DOSKnownSpellsOffset+3] = 2
	data[DOSKnownSpellsEnd-1] = 1

	record, err := ParseDOSPlayerSpellRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := record.MemorizedSpells, []uint8{7, 42}; !sameUint8(got, want) {
		t.Fatalf("memorized spells=%v, want %v", got, want)
	}
	if got, want := record.KnownSpells, []uint8{1, 4, 100}; !sameUint8(got, want) {
		t.Fatalf("known spells=%v, want %v", got, want)
	}

	character := validCharacter()
	if err := character.ApplyDOSSpellRecord(data); err != nil {
		t.Fatal(err)
	}
	if got, want := character.SpellSlots, []uint8{7, 42}; !sameUint8(got, want) {
		t.Fatalf("character spell slots=%v, want %v", got, want)
	}
	if got, want := character.KnownSpells, []uint8{1, 4, 100}; !sameUint8(got, want) {
		t.Fatalf("character known spells=%v, want %v", got, want)
	}
}

func TestParseDOSPlayerSpellRecordRejectsTruncatedData(t *testing.T) {
	if _, err := ParseDOSPlayerSpellRecord(make([]byte, DOSKnownSpellsEnd-1)); err == nil {
		t.Fatal("expected truncated DOS record error")
	}
}

func TestParseDOSPlayerRecordProjectsDocumentedCharacterFields(t *testing.T) {
	data := make([]byte, DOSPlayerRecordSize)
	data[0] = 4
	copy(data[1:], []byte("ELLA"))
	data[0x10], data[0x11], data[0x12], data[0x14] = 16, 17, 15, 12
	data[0x1C] = 75
	data[0x16], data[0x18], data[0x1A] = 14, 13, 10
	data[0x74], data[0x75] = 7, 5 // human magic-user
	data[0x78], data[0x1A4] = 22, 18
	copy(data[DOSSavingThrowsOffset:DOSSavingThrowsEnd], []byte{14, 12, 10, 13, 11})
	data[0x186] = 0xFE // signed -2
	copy(data[DOSThiefSkillsOffset:DOSThiefSkillsEnd], []byte{12, 34, 56, 78, 90, 11, 22, 33})
	data[0x10E] = 4
	data[0x141], data[0x142], data[0x143], data[0x144] = 3, 4, 0x0A, 2
	binary.LittleEndian.PutUint16(data[0x101:0x103], 123)
	binary.LittleEndian.PutUint32(data[0x14D:0x151], 0x12345678)
	binary.LittleEndian.PutUint32(data[0x0F2:0x0F6], 0x87654321)
	data[DOSMemorizedSpellsOffset] = 15

	record, err := ParseDOSPlayerRecord(data, "wizard-1")
	if err != nil {
		t.Fatal(err)
	}
	if record.Name != "ELLA" || record.Level != 4 || record.Gold != 123 || record.CurrentHitPoints != 18 || record.IconHead != 3 || record.IconID != 0x0A || record.ItemsPointer != 0x12345678 || record.EffectsPointer != 0x87654321 || len(record.SavingThrows) != 5 || record.SavingThrows[2] != 10 || record.SavingThrowBonus != -2 {
		t.Fatalf("record=%#v", record)
	}
	itemData := make([]byte, monster.ItemRecordSize)
	itemData[0x2E], itemData[0x34], itemData[0x39] = 36, 1, 2
	copy(itemData[:4], []byte("SWORD"))
	if err := record.ApplyInventory(itemData); err != nil {
		t.Fatal(err)
	}
	effectData := []byte{0x27, 0x34, 0x12, 7, 1, 8, 9, 10, 11}
	if err := record.ApplyEffects(effectData); err != nil {
		t.Fatal(err)
	}
	character, err := record.Character()
	if err != nil {
		t.Fatal(err)
	}
	if character.HitPoints != 18 || character.MaxHitPoints != 22 || character.Gold != 123 || character.Abilities.StrengthFull != 17 || character.Abilities.StrengthExceptional != 75 || character.IconHeadBlock != 3 || character.IconID != 0x0A || character.SpellSlots[0] != 15 || character.OpenLocksSkill() != 34 || len(character.ThiefSkills) != 8 || len(character.SavingThrows) != 5 || character.SavingThrows[4] != 11 || character.SavingThrowBonus != -2 || len(character.Equipment) != 1 || character.Equipment[0].Type != 36 || character.Equipment[0].Readied != true || len(character.Effects) != 1 || character.Effects[0].Kind != 0x27 {
		t.Fatalf("character=%#v", character)
	}
	if err := character.ApplyDOSInventory(make([]byte, monster.ItemRecordSize-1)); err == nil {
		t.Fatal("expected malformed DOS inventory error")
	}
	if err := character.ApplyDOSEffects(make([]byte, monster.AffectRecordSize-1)); err == nil {
		t.Fatal("expected malformed DOS effects error")
	}
}

func TestParseDOSPlayerFilesBundlesOptionalSidecars(t *testing.T) {
	record := make([]byte, DOSPlayerRecordSize)
	record[0] = 4
	copy(record[1:], []byte("ELLA"))
	record[0x10], record[0x12], record[0x14] = 16, 15, 12
	record[0x16], record[0x18], record[0x1A] = 14, 13, 10
	record[0x74], record[0x75], record[0x78], record[0x1A4] = 7, 5, 22, 18
	record[0x10E] = 4
	character, err := ParseDOSPlayerFiles("ella-1", DOSPlayerFiles{
		Record:    record,
		Effects:   []byte{0x27, 0x34, 0x12, 7, 1, 8, 9, 10, 11},
		Inventory: make([]byte, monster.ItemRecordSize),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(character.Effects) != 1 || len(character.Equipment) != 1 {
		t.Fatalf("bundled character=%#v", character)
	}
}

func sameUint8(left, right []uint8) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
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

func TestFighterProjectionPreservesDOSIconID(t *testing.T) {
	character := validCharacter()
	character.IconID = 0x0A
	fighter, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if fighter.PartyIconID != 0x0A {
		t.Fatalf("fighter icon id=%#x, want 0x0A", fighter.PartyIconID)
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

func TestFighterWithEquipmentProjectsArmorMovementAllowance(t *testing.T) {
	items := make([]monster.BaseItem, 56)
	items[55] = monster.BaseItem{Type: 55, ACAdjustment: 183}
	fighter, err := validCharacterWithEquipment([]monster.ItemRecord{{Type: 55, Readied: true}}, monster.BaseItemCatalog{Items: items})
	if err != nil {
		t.Fatal(err)
	}
	if fighter.MovementAllowance != 9 {
		t.Fatalf("chain armor movement allowance=%d", fighter.MovementAllowance)
	}
}

func validCharacterWithEquipment(items []monster.ItemRecord, catalog monster.BaseItemCatalog) (combat.Fighter, error) {
	character := validCharacter()
	character.Equipment = items
	return character.FighterWithEquipment(catalog)
}

func TestFighterProjectionAppliesActiveBlessAndCurse(t *testing.T) {
	character := validCharacter()
	base, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	character.Effects = []monster.AffectRecord{
		{Kind: 0x01, Strength: 1, Active: true},
		{Kind: 0x02, Strength: 1, Active: false},
	}
	fighter, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AttackBonus != base.AttackBonus+1 {
		t.Fatalf("active bless projection attack=%d base=%d", fighter.AttackBonus, base.AttackBonus)
	}
	character.Effects[0].Active = false
	character.Effects[1].Active = true
	fighter, err = character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AttackBonus != base.AttackBonus-1 {
		t.Fatalf("active curse projection attack=%d base=%d", fighter.AttackBonus, base.AttackBonus)
	}
}

func TestFighterProjectionAppliesActiveBlindBestowCurseAndPrayer(t *testing.T) {
	character := validCharacter()
	base, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	character.Effects = []monster.AffectRecord{
		{Kind: 0x21, Active: true}, // Blind
		{Kind: 0x24, Active: true}, // Bestow Curse
		{Kind: 0x31, Active: true}, // friendly Prayer
	}
	fighter, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AttackBonus != base.AttackBonus-4-4+1 || fighter.ArmorClass != base.ArmorClass+4 {
		t.Fatalf("projected fighter=%#v base=%#v", fighter, base)
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

func TestConsumeAmmunitionUsesInjectedRawTypeMappingAtomically(t *testing.T) {
	character := validCharacter()
	character.Equipment = []monster.ItemRecord{
		{Type: 73, Count: 3},
		{Type: 28, Count: 2},
	}
	mapping := map[uint8][]uint8{11: {73}}
	if err := character.ConsumeAmmunition(11, 2, mapping); err != nil {
		t.Fatal(err)
	}
	if len(character.Equipment) != 2 || character.Equipment[0].Count != 1 || character.Equipment[1].Count != 2 {
		t.Fatalf("after arrow consumption=%+v", character.Equipment)
	}
	before := append([]monster.ItemRecord(nil), character.Equipment...)
	if err := character.ConsumeAmmunition(11, 2, mapping); err == nil {
		t.Fatal("expected insufficient arrow error")
	}
	if len(character.Equipment) != len(before) || character.Equipment[0].Count != before[0].Count {
		t.Fatalf("insufficient consumption mutated equipment=%+v before=%+v", character.Equipment, before)
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

func TestCombatIconBlocksAppliesSmallIconNamespaceOffset(t *testing.T) {
	character := validCharacter()
	character.Race = RaceDwarf
	character.IconHeadBlock, character.IconWeaponBlock = 3, 7
	if head, body := character.CombatIconBlocks(); head != 0x43 || body != 0x47 {
		t.Fatalf("small icon blocks=(%02X,%02X), want (43,47)", head, body)
	}
	character.IconSize = 2
	if head, body := character.CombatIconBlocks(); head != 3 || body != 7 {
		t.Fatalf("normal icon blocks=(%02X,%02X), want (03,07)", head, body)
	}
	if head, body := character.CombatIconBlocksFor(true); head != 0x83 || body != 0x87 {
		t.Fatalf("attack icon blocks=(%02X,%02X), want (83,87)", head, body)
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
