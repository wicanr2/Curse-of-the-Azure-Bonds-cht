// Package party contains the data-neutral character-creation contract. It
// follows the six player races and class minimums observed in the CoAB rule
// book/reference rewrite; combat serialization remains a separate adapter.
package party

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

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

// RollAbilities reproduces the rule-book's six ability generation shape: six
// independent 3d6 rolls. A caller-provided seed keeps tests and replays
// deterministic without coupling the party package to UI randomness.
func RollAbilities(seed int64) Abilities {
	rng := rand.New(rand.NewSource(seed))
	roll := func() int {
		total := 0
		for i := 0; i < 3; i++ {
			total += rng.Intn(6) + 1
		}
		return total
	}
	return Abilities{Strength: roll(), Intelligence: roll(), Wisdom: roll(), Dexterity: roll(), Constitution: roll(), Charisma: roll()}
}

func (a Abilities) Value(index int) (int, error) {
	values := [...]int{a.Strength, a.Intelligence, a.Wisdom, a.Dexterity, a.Constitution, a.Charisma}
	if index < 0 || index >= len(values) {
		return 0, fmt.Errorf("ability index %d is out of range", index)
	}
	return values[index], nil
}

func (a *Abilities) Adjust(index, delta int) error {
	if delta == 0 {
		return nil
	}
	value, err := a.Value(index)
	if err != nil {
		return err
	}
	if value+delta < 3 || value+delta > 18 {
		return fmt.Errorf("ability value must remain between 3 and 18")
	}
	value += delta
	switch index {
	case 0:
		a.Strength = value
	case 1:
		a.Intelligence = value
	case 2:
		a.Wisdom = value
	case 3:
		a.Dexterity = value
	case 4:
		a.Constitution = value
	case 5:
		a.Charisma = value
	}
	return nil
}

type Character struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Race      Race      `json:"race"`
	Class     Class     `json:"class"`
	Abilities Abilities `json:"abilities"`
	Level     int       `json:"level"`
	// IconSize follows Player.icon_size: 1 is small and 2 is normal.
	// Zero means derive the original default from Race for old save files.
	IconSize  uint8                `json:"icon_size,omitempty"`
	Equipment []monster.ItemRecord `json:"equipment,omitempty"`
}

type Roster []Character

// ItemClassBit returns the class usability bit used by the original ITEMS
// table. The values deliberately follow the DOS table rather than the local
// enum order, which starts cleric at zero.
func ItemClassBit(class Class) (uint8, bool) {
	switch class {
	case ClassMagicUser:
		return 1, true
	case ClassCleric:
		return 2, true
	case ClassThief:
		return 4, true
	case ClassFighter:
		return 8, true
	case ClassPaladin:
		return 64, true
	case ClassRanger:
		return 128, true
	default:
		return 0, false
	}
}

// CanEquip validates one inventory index against the reference class mask
// and the currently readied equipment slots. It does not mutate the party.
func (c Character) CanEquip(index int, catalog monster.BaseItemCatalog) error {
	if index < 0 || index >= len(c.Equipment) {
		return fmt.Errorf("equipment index %d is out of range", index)
	}
	item := c.Equipment[index]
	base, ok := catalog.Lookup(item.Type)
	if !ok {
		return fmt.Errorf("item type 0x%02X is outside base catalog", item.Type)
	}
	classBit, ok := ItemClassBit(c.Class)
	if !ok || !base.UsableByMask(classBit) {
		return fmt.Errorf("%s cannot equip item type 0x%02X", c.Class, item.Type)
	}
	if item.Readied {
		return nil
	}
	rings := 0
	for otherIndex, other := range c.Equipment {
		if otherIndex == index || !other.Readied {
			continue
		}
		otherBase, otherOK := catalog.Lookup(other.Type)
		if !otherOK {
			return fmt.Errorf("readied item type 0x%02X is outside base catalog", other.Type)
		}
		if base.Slot == 9 && otherBase.Slot == 9 {
			rings++
			continue
		}
		if base.Slot <= 9 && base.Slot == otherBase.Slot {
			return fmt.Errorf("equipment slot %d is already occupied", base.Slot)
		}
		if base.Slot == 0 && base.HandsRequired >= 2 && otherBase.Slot == 1 {
			return fmt.Errorf("two-handed item conflicts with off-hand equipment")
		}
		if base.Slot == 1 && otherBase.Slot == 0 && otherBase.HandsRequired >= 2 {
			return fmt.Errorf("off-hand equipment conflicts with two-handed weapon")
		}
	}
	if base.Slot == 9 && rings >= 2 {
		return fmt.Errorf("at most two rings may be readied")
	}
	return nil
}

// EquipItem marks an inventory item readied after CanEquip validation.
func (c *Character) EquipItem(index int, catalog monster.BaseItemCatalog) error {
	if c == nil {
		return fmt.Errorf("cannot equip item on nil character")
	}
	if err := c.CanEquip(index, catalog); err != nil {
		return err
	}
	c.Equipment[index].Readied = true
	return nil
}

// UnequipItem clears the readied state without changing the inventory record.
func (c *Character) UnequipItem(index int) error {
	if c == nil {
		return fmt.Errorf("cannot unequip item on nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return fmt.Errorf("equipment index %d is out of range", index)
	}
	if c.Equipment[index].Readied && c.Equipment[index].Cursed {
		return fmt.Errorf("cursed item type 0x%02X cannot be unequipped", c.Equipment[index].Type)
	}
	c.Equipment[index].Readied = false
	return nil
}

// RemoveItem removes one inventory unit. Count zero is the DOS encoding for
// a non-stacking single item; a positive count is decremented in place. A
// readied item must be unequipped first so shop/treasure mutations cannot
// silently invalidate the equipment projection.
func (c *Character) RemoveItem(index int) (monster.ItemRecord, error) {
	if c == nil {
		return monster.ItemRecord{}, fmt.Errorf("cannot remove item from nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return monster.ItemRecord{}, fmt.Errorf("equipment index %d is out of range", index)
	}
	item := c.Equipment[index]
	if item.Readied {
		return monster.ItemRecord{}, fmt.Errorf("readied item type 0x%02X must be unequipped first", item.Type)
	}
	removed := item
	removed.Count = 1
	if item.Count > 1 {
		c.Equipment[index].Count--
		return removed, nil
	}
	c.Equipment = append(c.Equipment[:index], c.Equipment[index+1:]...)
	return removed, nil
}

// UseConsumable consumes one scroll/potion or one charge from a wand. The
// returned signal is intentionally renderer- and rules-neutral; a later
// spell/effect system decides what EffectID or SpellIDs do.
func (c *Character) UseConsumable(index int, catalog monster.BaseItemCatalog) (monster.ConsumableUse, error) {
	if c == nil {
		return monster.ConsumableUse{}, fmt.Errorf("cannot use item on nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return monster.ConsumableUse{}, fmt.Errorf("equipment index %d is out of range", index)
	}
	use, err := c.Equipment[index].DecodeConsumable(catalog)
	if err != nil {
		return monster.ConsumableUse{}, err
	}
	if use.Kind == monster.ConsumableCharged {
		if use.ChargesBefore <= 0 {
			return monster.ConsumableUse{}, fmt.Errorf("item type 0x%02X has no charges", c.Equipment[index].Type)
		}
		c.Equipment[index].Affects[0]--
		use.ChargesAfter = int(c.Equipment[index].Affects[0])
		return use, nil
	}
	if _, err := c.RemoveItem(index); err != nil {
		return monster.ConsumableUse{}, err
	}
	return use, nil
}

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

// Fighter returns the starter combat projection used by the character
// creation slice. The full AD&D level, equipment and spell tables remain a
// separate data source; these defaults are intentionally deterministic.
func (c Character) Fighter() (combat.Fighter, error) {
	if err := c.Validate(); err != nil {
		return combat.Fighter{}, err
	}
	hitDie, damageSides := 6, 4
	specialist := c.Class == ClassFighter || c.Class == ClassRanger || c.Class == ClassPaladin
	switch c.Class {
	case ClassCleric:
		hitDie, damageSides = 8, 6
	case ClassFighter, ClassRanger, ClassPaladin:
		hitDie, damageSides = 10, 8
	case ClassThief:
		hitDie, damageSides = 6, 6
	}
	constitutionBonus := (c.Abilities.Constitution - 14) / 2
	hitPoints := hitDie + constitutionBonus
	if hitPoints < 1 {
		hitPoints = 1
	}
	attackBonus := (c.Abilities.Strength - 10) / 2
	if !specialist {
		attackBonus = (c.Abilities.Dexterity - 10) / 2
	}
	armorClass := 10 - (c.Abilities.Dexterity-10)/2
	iconSize := c.IconSize
	if iconSize == 0 {
		iconSize = DefaultIconSize(c.Race)
	}
	return combat.Fighter{
		ID: c.ID, Name: c.Name, Side: combat.SideParty,
		HasPartyIcon: true, PartyHeadBlock: 0, PartyBodyBlock: 0, PartyIconSize: iconSize,
		HitPoints: hitPoints, MaxHitPoints: hitPoints, ArmorClass: armorClass,
		AttackBonus: attackBonus, DamageDiceCount: 1, DamageDiceSides: damageSides,
		InitiativeBonus: (c.Abilities.Dexterity - 10) / 2,
	}, nil
}

// FighterWithEquipment applies readied base weapon/armor effects from a
// decoded ITEMS catalog. It is an additive adapter for remake party JSON;
// magical effect stacks and the original DOS inventory record remain outside
// this boundary.
func (c Character) FighterWithEquipment(catalog monster.BaseItemCatalog) (combat.Fighter, error) {
	fighter, err := c.Fighter()
	if err != nil {
		return combat.Fighter{}, err
	}
	hasWeapon := false
	for _, item := range c.Equipment {
		if !item.Readied {
			continue
		}
		effect, effectErr := item.Effect(catalog, false)
		if effectErr != nil {
			return combat.Fighter{}, effectErr
		}
		fighter.ArmorClass -= effect.ArmorClassImprovement
		if effect.DamageDiceCount > 0 && !hasWeapon {
			fighter.AttackBonus += effect.AttackBonus
			fighter.DamageDiceCount = effect.DamageDiceCount
			fighter.DamageDiceSides = effect.DamageDiceSides
			fighter.DamageBonus = effect.DamageBonus
			hasWeapon = true
		}
	}
	return fighter, nil
}

// DefaultIconSize matches the original character creation race switch.
func DefaultIconSize(r Race) uint8 {
	switch r {
	case RaceDwarf, RaceGnome, RaceHalfling:
		return 1
	default:
		return 2
	}
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
