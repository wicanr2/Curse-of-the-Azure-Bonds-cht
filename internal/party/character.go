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
	Strength int
	// StrengthFull and StrengthExceptional preserve DOS Str.full and
	// Str00.cur when imported; Strength remains the legacy normalized value.
	StrengthFull        int
	StrengthExceptional int
	Intelligence        int
	Wisdom              int
	Dexterity           int
	Constitution        int
	Charisma            int
}

// HealthStatus mirrors the reference player's combat health states while
// keeping the existing zero-value JSON saves valid.
type HealthStatus uint8

const (
	HealthStatusOK HealthStatus = iota
	HealthStatusAnimated
	HealthStatusUnconscious
	HealthStatusDying
	HealthStatusDead
)

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
	IconSize        uint8                  `json:"icon_size,omitempty"`
	IconHeadBlock   uint8                  `json:"icon_head,omitempty"`
	IconWeaponBlock uint8                  `json:"icon_weapon,omitempty"`
	IconID          uint8                  `json:"icon_id,omitempty"`
	HitPoints       int                    `json:"hit_points,omitempty"`
	MaxHitPoints    int                    `json:"max_hit_points,omitempty"`
	HealthStatus    HealthStatus           `json:"health_status,omitempty"`
	Bleeding        int                    `json:"bleeding,omitempty"`
	Gold            uint16                 `json:"gold,omitempty"`
	Gems            uint16                 `json:"gems,omitempty"`
	Jewelry         uint16                 `json:"jewelry,omitempty"`
	Equipment       []monster.ItemRecord   `json:"equipment,omitempty"`
	Effects         []monster.AffectRecord `json:"effects,omitempty"`
	// SpellSlots is the data-neutral ordered spell-slot list used by ECL
	// SPELL resolution. It is optional until original DOS player offsets are
	// decoded; old saves therefore remain valid.
	SpellSlots []uint8 `json:"spell_slots,omitempty"`
	// KnownSpells preserves the imported spell-book flags separately from
	// SpellSlots, which represents spells currently memorized for use.
	KnownSpells []uint8 `json:"known_spells,omitempty"`
	// SavingThrows preserves the original DOS saveVerse order. The raw
	// five-byte order is retained so ECL DAMAGE adapters can apply the correct
	// saving-throw type.
	SavingThrows []uint8 `json:"saving_throws,omitempty"`
	// SavingThrowBonus preserves DOS player field_186, including signed item or
	// effect-derived bonus already present in an imported record.
	SavingThrowBonus int8 `json:"saving_throw_bonus,omitempty"`
	// ThiefSkills preserves the eight DOS thief percentages; index 1 is
	// open-locks and remains in original order.
	ThiefSkills []uint8 `json:"thief_skills,omitempty"`
}

const (
	DeathDamageFire = 0x01
	DeathDamageAcid = 0x10
)

// DeathEffectContext contains the external state required by the verified
// CheckAffectsEffect(Death) handlers. Damage type is deliberately explicit
// because ECL DAMAGE's five operands do not encode it.
type DeathEffectContext struct {
	DamageFlags       uint8
	DamageFlagsKnown  bool
	CombatHealAllowed bool
	RollDie           func(int) int
}

// ApplyDeathEffects projects the verified affect_63 and troll_fire_or_acid
// death handlers. Unknown Death effects remain preserved for later adapters.
func (c *Character) ApplyDeathEffects(context DeathEffectContext) error {
	if c == nil {
		return nil
	}
	if context.RollDie == nil {
		return fmt.Errorf("death effects require an injected die roller")
	}
	for index := 0; index < len(c.Effects); index++ {
		effect := c.Effects[index]
		if !effect.Active {
			continue
		}
		switch effect.Kind {
		case 0x63: // affect_63: death-time recovery
			heal := 0
			if c.HealthStatus == HealthStatusDying && c.Bleeding < 6 {
				heal = 6 - c.Bleeding
			} else if c.HealthStatus == HealthStatusUnconscious {
				heal = 6
			}
			if heal > 0 && context.CombatHealAllowed {
				c.HealthStatus = HealthStatusOK
				c.HitPoints = heal
				c.Bleeding = 0
				roll := context.RollDie(4)
				if roll < 1 || roll > 4 {
					return fmt.Errorf("death recovery die roll %d is outside 1..4", roll)
				}
				regeneration := uint16(roll + 1)
				c.Effects = append(c.Effects, monster.AffectRecord{Kind: 0x5F, Value: regeneration, Duration: regeneration, Strength: 0xFF, Active: true})
				c.Effects = append(c.Effects[:index], c.Effects[index+1:]...)
				index--
			}
		case 0x64: // troll_fire_or_acid
			if context.DamageFlagsKnown && context.DamageFlags&(DeathDamageFire|DeathDamageAcid) == 0 {
				total := 0
				for rollIndex := 0; rollIndex < 3; rollIndex++ {
					roll := context.RollDie(6)
					if roll < 1 || roll > 6 {
						return fmt.Errorf("troll regeneration die roll %d is outside 1..6", roll)
					}
					total += roll
				}
				c.Effects = append(c.Effects, monster.AffectRecord{Kind: 0x66, Value: uint16(total), Duration: uint16(total), Strength: 0xFF, Active: true})
			}
		}
	}
	return nil
}

// combatAffectKinds is the verified RemoveCombatAffects table from the DOS
// engine. Unknown and persistent non-combat effects remain untouched.
var combatAffectKinds = map[uint8]struct{}{
	0x07: {}, 0x0B: {}, 0x0D: {}, 0x15: {}, 0x17: {}, 0x1E: {}, 0x1F: {}, 0x20: {},
	0x33: {}, 0x34: {}, 0x35: {}, 0x3A: {}, 0x3B: {}, 0x5F: {}, 0x62: {}, 0x88: {},
	0x89: {}, 0x8B: {}, 0x90: {},
}

// RemoveCombatAffects removes only the effects listed by the reference
// RemoveCombatAffects routine and reports how many records were removed.
func (c *Character) RemoveCombatAffects() int {
	if c == nil || len(c.Effects) == 0 {
		return 0
	}
	kept := make([]monster.AffectRecord, 0, len(c.Effects))
	removed := 0
	for _, effect := range c.Effects {
		if _, ok := combatAffectKinds[effect.Kind]; ok {
			removed++
			continue
		}
		kept = append(kept, effect)
	}
	c.Effects = kept
	return removed
}

// AdvanceEffects consumes imported DOS effect durations without applying
// effect-specific AD&D modifiers. The removed count lets game state surface
// a localized message when a later rules layer wants to do so.
func (c *Character) AdvanceEffects(minutes uint16) int {
	if c == nil || minutes == 0 {
		return 0
	}
	before := len(c.Effects)
	c.Effects = monster.AdvanceAffects(c.Effects, minutes)
	return before - len(c.Effects)
}

// OpenLocksSkill returns the imported DOS open-locks percentage. Zero means
// no verified skill data is available, not an automatic failed roll.
func (c Character) OpenLocksSkill() uint8 {
	if len(c.ThiefSkills) <= 1 {
		return 0
	}
	return c.ThiefSkills[1]
}

type Roster []Character

type SpellMatch struct {
	CharacterIndex int
	SlotIndex      int
}

// FindSpell searches party characters in marching order, then each character's
// ordered spell slots. This matches ECL SPELL's first-match contract while
// keeping the source of spell slots replaceable by a DOS record decoder.
func (r Roster) FindSpell(spellID uint16) (SpellMatch, bool) {
	if spellID > 0xFF {
		return SpellMatch{}, false
	}
	for characterIndex, character := range r {
		for slotIndex, knownSpell := range character.SpellSlots {
			if knownSpell == uint8(spellID) {
				return SpellMatch{CharacterIndex: characterIndex, SlotIndex: slotIndex}, true
			}
		}
	}
	return SpellMatch{}, false
}

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

// DestroyItemType applies an explicit ECL DESTROY ITEMS effect. Unlike the
// player-facing RemoveItem operation, this effect may remove readied items:
// the original event can destroy equipment while it is worn. It removes all
// matching records and returns the number of item units destroyed.
func (c *Character) DestroyItemType(itemType uint8) int {
	if c == nil {
		return 0
	}
	updated := make([]monster.ItemRecord, 0, len(c.Equipment))
	destroyed := 0
	for _, item := range c.Equipment {
		if item.Type != itemType {
			updated = append(updated, item)
			continue
		}
		count := int(item.Count)
		if count == 0 {
			count = 1
		}
		destroyed += count
	}
	c.Equipment = updated
	return destroyed
}

// ConsumeAmmunition removes shots for a weapon's raw ITEMS AmmunitionType.
// The raw code and inventory item type use different namespaces in CoAB, so
// the caller must inject the verified mapping for the current game's data.
// The mutation is atomic: insufficient mapped ammunition leaves equipment
// unchanged.
func (c *Character) ConsumeAmmunition(ammunitionType uint8, shots int, itemTypes map[uint8][]uint8) error {
	if c == nil {
		return fmt.Errorf("cannot consume ammunition from nil character")
	}
	if ammunitionType == 0 {
		return fmt.Errorf("ammunition type is not required")
	}
	if shots <= 0 {
		return fmt.Errorf("ammunition shots must be positive")
	}
	candidates, ok := itemTypes[ammunitionType]
	if !ok || len(candidates) == 0 {
		return fmt.Errorf("no inventory mapping for ammunition type %d", ammunitionType)
	}
	allowed := make(map[uint8]bool, len(candidates))
	for _, itemType := range candidates {
		allowed[itemType] = true
	}
	available := 0
	for _, item := range c.Equipment {
		if !allowed[item.Type] {
			continue
		}
		if item.Count == 0 {
			available++
		} else {
			available += int(item.Count)
		}
	}
	if available < shots {
		return fmt.Errorf("ammunition type %d has %d shots; need %d", ammunitionType, available, shots)
	}
	remaining := shots
	updated := make([]monster.ItemRecord, 0, len(c.Equipment))
	for _, item := range c.Equipment {
		if !allowed[item.Type] || remaining == 0 {
			updated = append(updated, item)
			continue
		}
		count := 1
		if item.Count > 0 {
			count = int(item.Count)
		}
		consume := count
		if consume > remaining {
			consume = remaining
		}
		remaining -= consume
		if item.Count > 0 {
			item.Count -= uint8(consume)
			if item.Count > 0 {
				updated = append(updated, item)
			}
		}
	}
	c.Equipment = updated
	return nil
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
	hitPoints, maxHitPoints := hitPoints, hitPoints
	if c.MaxHitPoints > 0 {
		maxHitPoints = c.MaxHitPoints
		hitPoints = c.HitPoints
	}
	if c.HitPoints > 0 && c.HitPoints <= maxHitPoints {
		hitPoints = c.HitPoints
	}
	headBlock, bodyBlock := c.CombatIconBlocks()
	fighter := combat.Fighter{
		ID: c.ID, Name: c.Name, Side: combat.SideParty,
		HasPartyIcon: true, PartyHeadBlock: headBlock, PartyBodyBlock: bodyBlock, PartyIconID: c.IconID, PartyIconSize: iconSize,
		HitPoints: hitPoints, MaxHitPoints: maxHitPoints, ArmorClass: armorClass,
		AttackBonus: attackBonus, DamageDiceCount: 1, DamageDiceSides: damageSides,
		InitiativeBonus: (c.Abilities.Dexterity - 10) / 2,
	}
	return c.applyKnownEffects(fighter), nil
}

// applyKnownEffects projects only unconditional attack and AC modifiers documented
// for the current CoAB rules slice. Effects requiring target alignment,
// morale, saving throws, or status transitions remain in the rules layer.
func (c Character) applyKnownEffects(fighter combat.Fighter) combat.Fighter {
	for _, effect := range c.Effects {
		if !effect.Active {
			continue
		}
		switch effect.Kind {
		case 0x01: // Bless: -1 THAC0, equivalent to +1 attack bonus.
			fighter.AttackBonus++
		case 0x02: // Curse: +1 THAC0, equivalent to -1 attack bonus.
			fighter.AttackBonus--
		case 0x21: // Blind: -4 attack and +4 AC (worse AC).
			fighter.AttackBonus -= 4
			fighter.ArmorClass += 4
		case 0x24: // Bestow Curse: -4 attack.
			fighter.AttackBonus -= 4
		case 0x31: // Prayer: friendly effect gives -1 THAC0, or +1 attack.
			fighter.AttackBonus++
		}
	}
	return fighter
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
		if effect.MovementAllowance > 0 && (fighter.MovementAllowance == 0 || effect.MovementAllowance < fighter.MovementAllowance) {
			fighter.MovementAllowance = effect.MovementAllowance
		}
		if effect.DamageDiceCount > 0 && !hasWeapon {
			fighter.AttackBonus += effect.AttackBonus
			fighter.DamageDiceCount = effect.DamageDiceCount
			fighter.DamageDiceSides = effect.DamageDiceSides
			fighter.DamageBonus = effect.DamageBonus
			fighter.AttacksPerTurn = effect.AttacksPerTurn
			fighter.AmmunitionType = effect.AmmunitionType
			fighter.WeaponRange = effect.WeaponRange
			fighter.MissileWeapon = effect.MissileWeapon
			fighter.ThrownWeapon = effect.ThrownWeapon
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

// CombatIconBlocks maps raw DOS head_icon/weapon_icon slots to the actual
// CHEAD/CBODY blocks selected by LoadPlayerCombatIcon. Small icons use the T
// files, whose namespace is the raw slot plus 0x40.
func (c Character) CombatIconBlocks() (head, body uint8) {
	return c.CombatIconBlocksFor(false)
}

// CombatIconBlocksFor adds the reference LoadIcons attack namespace when the
// combat icon is in its attack state. The DAX loader uses normal_id+0x80.
func (c Character) CombatIconBlocksFor(attack bool) (head, body uint8) {
	head, body = c.IconHeadBlock, c.IconWeaponBlock
	size := c.IconSize
	if size == 0 {
		size = DefaultIconSize(c.Race)
	}
	if size == 1 {
		if head < 0x40 {
			head += 0x40
		}
		if body < 0x40 {
			body += 0x40
		}
	}
	if attack {
		head += 0x80
		body += 0x80
	}
	return head, body
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
