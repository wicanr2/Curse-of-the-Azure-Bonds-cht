package party

import (
	"encoding/binary"
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// DOS player/creature records use a shared spell layout. These constants are
// deliberately limited to fields documented by the public CoAB format notes;
// the rest of the record is still left to a future parser.
const (
	DOSMemorizedSpellsOffset = 0x01E
	DOSMemorizedSpellsEnd    = 0x072 // exclusive; last documented byte is 0x071
	DOSKnownSpellsOffset     = 0x079
	DOSKnownSpellsEnd        = 0x0DD // exclusive; last documented byte is 0x0DC
	DOSSavingThrowsOffset    = 0x0DF
	DOSSavingThrowsEnd       = 0x0E4 // exclusive; five saveVerse bytes
	DOSThiefSkillsOffset     = 0x0EA
	DOSThiefSkillsEnd        = 0x0F2 // exclusive; open-locks is index 1
	DOSPlayerRecordSize      = 0x1A6 // last documented byte is current movement at 0x1A5
)

// PatchDOSPlayerRecord updates only fields whose offsets are currently
// documented by the CoAB record parser. Unknown bytes are copied unchanged.
// This is used by the SAVGAM slot writer; it is not a full Player serializer.
func PatchDOSPlayerRecord(data []byte, character Character) ([]byte, error) {
	if len(data) < DOSPlayerRecordSize {
		return nil, fmt.Errorf("DOS player record is %d bytes; need at least 0x%X", len(data), DOSPlayerRecordSize)
	}
	name := []byte(character.Name)
	if len(name) < 1 || len(name) > 15 {
		return nil, fmt.Errorf("DOS player name must be 1..15 bytes, got %d", len(name))
	}
	if character.HitPoints < 0 || character.HitPoints > 255 || character.MaxHitPoints < 0 || character.MaxHitPoints > 255 {
		return nil, fmt.Errorf("DOS HP must fit one byte: current=%d max=%d", character.HitPoints, character.MaxHitPoints)
	}
	out := append([]byte(nil), data...)
	for i := 0; i < 16; i++ {
		out[i] = 0
	}
	out[0] = byte(len(name))
	copy(out[1:], name)
	out[0x10] = uint8(character.Abilities.Strength)
	out[0x11] = uint8(character.Abilities.StrengthFull)
	out[0x1C] = uint8(character.Abilities.StrengthExceptional)
	out[0x12] = uint8(character.Abilities.Intelligence)
	out[0x14] = uint8(character.Abilities.Wisdom)
	out[0x16] = uint8(character.Abilities.Dexterity)
	out[0x18] = uint8(character.Abilities.Constitution)
	out[0x1A] = uint8(character.Abilities.Charisma)
	out[0x78] = uint8(character.MaxHitPoints)
	if character.ClassLevels != [8]uint8{} {
		copy(out[0x109:0x111], character.ClassLevels[:])
	}
	binary.LittleEndian.PutUint16(out[0x76:0x78], uint16(character.Age))
	if len(out) > 0x1A4 {
		out[0x1A4] = uint8(character.HitPoints)
	}
	binary.LittleEndian.PutUint16(out[0x101:0x103], character.Gold)
	binary.LittleEndian.PutUint16(out[0x105:0x107], character.Gems)
	binary.LittleEndian.PutUint16(out[0x107:0x109], character.Jewelry)
	out[0x141], out[0x142], out[0x143], out[0x144] = character.IconHeadBlock, character.IconWeaponBlock, character.IconID, character.IconSize
	for i := DOSMemorizedSpellsOffset; i < DOSMemorizedSpellsEnd; i++ {
		out[i] = 0
	}
	if len(character.SpellSlots) > DOSMemorizedSpellsEnd-DOSMemorizedSpellsOffset {
		return nil, fmt.Errorf("DOS memorized spell slots exceed %d", DOSMemorizedSpellsEnd-DOSMemorizedSpellsOffset)
	}
	copy(out[DOSMemorizedSpellsOffset:DOSMemorizedSpellsEnd], character.SpellSlots)
	for i := DOSKnownSpellsOffset; i < DOSKnownSpellsEnd; i++ {
		out[i] = 0
	}
	for _, spellID := range character.KnownSpells {
		if spellID == 0 || int(spellID) > DOSKnownSpellsEnd-DOSKnownSpellsOffset {
			return nil, fmt.Errorf("DOS known spell ID %d is outside 1..%d", spellID, DOSKnownSpellsEnd-DOSKnownSpellsOffset)
		}
		out[DOSKnownSpellsOffset+int(spellID)-1] = 1
	}
	if len(character.ThiefSkills) > DOSThiefSkillsEnd-DOSThiefSkillsOffset {
		return nil, fmt.Errorf("DOS thief skill count exceeds %d", DOSThiefSkillsEnd-DOSThiefSkillsOffset)
	}
	for i := DOSThiefSkillsOffset; i < DOSThiefSkillsEnd; i++ {
		out[i] = 0
	}
	copy(out[DOSThiefSkillsOffset:DOSThiefSkillsEnd], character.ThiefSkills)
	if len(character.SavingThrows) > DOSSavingThrowsEnd-DOSSavingThrowsOffset {
		return nil, fmt.Errorf("DOS saving throw count exceeds %d", DOSSavingThrowsEnd-DOSSavingThrowsOffset)
	}
	copy(out[DOSSavingThrowsOffset:DOSSavingThrowsEnd], character.SavingThrows)
	out[0x186] = byte(character.SavingThrowBonus)
	out[0xF7] = character.ControlMorale
	return out, nil
}

// DOSPlayerSpellRecord is the verified spell subset of a DOS .SAV/.GUY
// creature record. MemorizedSpells preserves slot order and omits empty slots;
// KnownSpells contains one-based spell IDs whose known flag is set.
type DOSPlayerSpellRecord struct {
	MemorizedSpells []uint8
	KnownSpells     []uint8
}

// DOSPlayerRecord is the verified, fixed-offset subset needed to project a
// DOS player into the remake. Multi-class raw class IDs are projected to a
// primary class while ClassLevels and MulticlassLevel remain lossless.
type DOSPlayerRecord struct {
	ID               string
	Name             string
	Race             Race
	Class            Class
	RawRace          uint8
	RawClass         uint8
	Abilities        Abilities
	Level            int
	MaxHitPoints     int
	CurrentHitPoints int
	Age              int16
	ControlMorale    uint8
	IconHead         uint8
	IconWeapon       uint8
	IconID           uint8
	IconSize         uint8
	Gold             uint16
	Gems             uint16
	Jewelry          uint16
	MemorizedSpells  []uint8
	KnownSpells      []uint8
	ItemsPointer     uint32
	EffectsPointer   uint32
	ThiefSkills      []uint8
	SavingThrows     []uint8
	SavingThrowBonus int8
	ClassLevels      [8]uint8
	MulticlassLevel  uint8
	Inventory        []monster.ItemRecord
	Effects          []monster.AffectRecord
}

// DOSPlayerFiles is the decomposed character bundle documented by the
// original game: the .SAV/.GUY record is required, while .FX and .SWG are
// optional sidecar streams.
type DOSPlayerFiles struct {
	Record    []byte
	Effects   []byte
	Inventory []byte
}

// ParseDOSPlayerFiles imports one character from the three original sidecar
// files. It deliberately stops before SAVGAM*.DAT/container parsing, whose
// address space and area payload are a separate format boundary.
func ParseDOSPlayerFiles(id string, files DOSPlayerFiles) (Character, error) {
	record, err := ParseDOSPlayerRecord(files.Record, id)
	if err != nil {
		return Character{}, err
	}
	if files.Effects != nil {
		if err := record.ApplyEffects(files.Effects); err != nil {
			return Character{}, err
		}
	}
	if files.Inventory != nil {
		if err := record.ApplyInventory(files.Inventory); err != nil {
			return Character{}, err
		}
	}
	return record.Character()
}

// ParseDOSPlayerRecord decodes the documented fixed portion of a decompressed
// .SAV/.GUY player record. Only single-class races/classes represented by the
// current remake Character model are accepted; raw offsets for inventory and
// effects are not silently interpreted.
func ParseDOSPlayerRecord(data []byte, id string) (DOSPlayerRecord, error) {
	return parseDOSPlayerRecord(data, id, false)
}

// ParseDOSNPCRecord accepts MON*CHA Player records whose class_id can be
// stale while exactly one ClassLevel slot identifies the class used by
// ReclacClassBonuses. Ordinary player save imports remain strict.
func ParseDOSNPCRecord(data []byte, id string) (DOSPlayerRecord, error) {
	return parseDOSPlayerRecord(data, id, true)
}

func parseDOSPlayerRecord(data []byte, id string, inferNPCClass bool) (DOSPlayerRecord, error) {
	if len(data) < DOSPlayerRecordSize {
		return DOSPlayerRecord{}, fmt.Errorf("DOS player record is %d bytes; need at least 0x%X", len(data), DOSPlayerRecordSize)
	}
	if id == "" {
		return DOSPlayerRecord{}, fmt.Errorf("DOS player record ID is required")
	}
	nameLength := int(data[0])
	if nameLength < 1 || nameLength > 15 {
		return DOSPlayerRecord{}, fmt.Errorf("DOS player name length %d is outside 1..15", nameLength)
	}
	rawRace, err := parseDOSRace(data[0x74])
	if err != nil {
		return DOSPlayerRecord{}, err
	}
	rawClass, err := parseDOSClass(data[0x75])
	if err != nil {
		return DOSPlayerRecord{}, err
	}
	var classLevels [8]uint8
	copy(classLevels[:], data[0x109:0x111])
	level := int(data[0x109+classLevelOffset(data[0x75])])
	if data[0x75] >= 8 && data[0x75] <= 16 {
		level = int(data[0xE6])
		if level == 0 {
			for _, classLevel := range classLevels {
				if int(classLevel) > level {
					level = int(classLevel)
				}
			}
		}
	}
	if level < 1 && inferNPCClass {
		slot := -1
		for index, classLevel := range classLevels {
			if classLevel == 0 {
				continue
			}
			if slot >= 0 {
				return DOSPlayerRecord{}, fmt.Errorf("DOS NPC has ambiguous class levels %v", classLevels)
			}
			slot = index
			level = int(classLevel)
		}
		if slot >= 0 {
			switch slot {
			case 0:
				rawClass = ClassCleric
			case 2:
				rawClass = ClassFighter
			case 3:
				rawClass = ClassPaladin
			case 4:
				rawClass = ClassRanger
			case 5:
				rawClass = ClassMagicUser
			case 6:
				rawClass = ClassThief
			default:
				return DOSPlayerRecord{}, fmt.Errorf("DOS NPC class-level slot %d is unsupported", slot)
			}
		}
	}
	if level < 1 {
		return DOSPlayerRecord{}, fmt.Errorf("DOS player class 0x%02X has no current level", rawClass)
	}
	spells, err := ParseDOSPlayerSpellRecord(data)
	if err != nil {
		return DOSPlayerRecord{}, err
	}
	return DOSPlayerRecord{
		ID: id, Name: string(data[1 : 1+nameLength]), Race: rawRace, Class: rawClass,
		RawRace: data[0x74], RawClass: data[0x75],
		Abilities: Abilities{
			Strength: int(data[0x10]), StrengthFull: int(data[0x11]), StrengthExceptional: int(data[0x1C]),
			Intelligence: int(data[0x12]), Wisdom: int(data[0x14]),
			Dexterity: int(data[0x16]), Constitution: int(data[0x18]), Charisma: int(data[0x1A]),
		},
		Level: level, MaxHitPoints: int(data[0x78]), CurrentHitPoints: int(data[0x1A4]),
		Age:           int16(binary.LittleEndian.Uint16(data[0x76:0x78])),
		ControlMorale: data[0xF7],
		IconHead:      data[0x141], IconWeapon: data[0x142], IconID: data[0x143], IconSize: data[0x144],
		Gold:             binary.LittleEndian.Uint16(data[0x101:0x103]),
		Gems:             binary.LittleEndian.Uint16(data[0x105:0x107]),
		Jewelry:          binary.LittleEndian.Uint16(data[0x107:0x109]),
		ItemsPointer:     binary.LittleEndian.Uint32(data[0x14D:0x151]),
		EffectsPointer:   binary.LittleEndian.Uint32(data[0x0F2:0x0F6]),
		ThiefSkills:      append([]uint8(nil), data[DOSThiefSkillsOffset:DOSThiefSkillsEnd]...),
		SavingThrows:     append([]uint8(nil), data[DOSSavingThrowsOffset:DOSSavingThrowsEnd]...),
		SavingThrowBonus: int8(data[0x186]),
		MemorizedSpells:  spells.MemorizedSpells, KnownSpells: spells.KnownSpells,
		ClassLevels: classLevels, MulticlassLevel: data[0xE6],
	}, nil
}

// Character projects the verified player fields into the current party model.
// It keeps the original current/max HP and icon values so the combat renderer
// can use imported data immediately.
func (r DOSPlayerRecord) Character() (Character, error) {
	character := Character{
		ID: r.ID, Name: r.Name, Race: r.Race, Class: r.Class, Abilities: r.Abilities,
		RawClassID: r.RawClass,
		Level:      r.Level, Age: r.Age, HitPoints: r.CurrentHitPoints, MaxHitPoints: r.MaxHitPoints,
		NPC: r.ControlMorale >= 0x80, ControlMorale: r.ControlMorale,
		ClassLevels: r.ClassLevels,
		Gold:        r.Gold, Gems: r.Gems, Jewelry: r.Jewelry,
		IconHeadBlock: r.IconHead, IconWeaponBlock: r.IconWeapon, IconID: r.IconID, IconSize: r.IconSize,
		Equipment:        append([]monster.ItemRecord(nil), r.Inventory...),
		Effects:          append([]monster.AffectRecord(nil), r.Effects...),
		SpellSlots:       append([]uint8(nil), r.MemorizedSpells...),
		KnownSpells:      append([]uint8(nil), r.KnownSpells...),
		ThiefSkills:      append([]uint8(nil), r.ThiefSkills...),
		SavingThrows:     append([]uint8(nil), r.SavingThrows...),
		SavingThrowBonus: r.SavingThrowBonus,
	}
	if err := character.Validate(); err != nil {
		return Character{}, err
	}
	return character, nil
}

// ApplyInventory decodes a DOS .SWG item stream and attaches it to the
// already parsed player record. The stream is a sequence of documented 0x3F
// byte item records; pointer resolution belongs to the outer save/container
// loader and is not guessed here.
func (r *DOSPlayerRecord) ApplyInventory(data []byte) error {
	if r == nil {
		return fmt.Errorf("cannot apply inventory to nil DOS player record")
	}
	items, err := monster.ParseItems(data)
	if err != nil {
		return err
	}
	r.Inventory = append(r.Inventory[:0], items...)
	return nil
}

// ApplyEffects decodes the external DOS .FX stream. Each effect is the
// documented 9-byte record; gameplay application remains a later rules layer.
func (r *DOSPlayerRecord) ApplyEffects(data []byte) error {
	if r == nil {
		return fmt.Errorf("cannot apply effects to nil DOS player record")
	}
	effects, err := monster.ParseAffects(data)
	if err != nil {
		return err
	}
	r.Effects = append(r.Effects[:0], effects...)
	return nil
}

func parseDOSRace(raw uint8) (Race, error) {
	switch raw {
	case 1:
		return RaceDwarf, nil
	case 2:
		return RaceElf, nil
	case 3:
		return RaceGnome, nil
	case 4:
		return RaceHalfElf, nil
	case 5:
		return RaceHalfling, nil
	case 6:
		return RaceHalfOrc, nil
	case 7:
		return RaceHuman, nil
	default:
		return 0, fmt.Errorf("unsupported DOS player race 0x%02X", raw)
	}
}

func parseDOSClass(raw uint8) (Class, error) {
	switch raw {
	case 0:
		return ClassCleric, nil
	case 2:
		return ClassFighter, nil
	case 3:
		return ClassPaladin, nil
	case 4:
		return ClassRanger, nil
	case 5:
		return ClassMagicUser, nil
	case 6:
		return ClassThief, nil
	case 8, 9, 10, 11, 12:
		return ClassCleric, nil
	case 13, 14, 15:
		return ClassFighter, nil
	case 16:
		return ClassMagicUser, nil
	default:
		return 0, fmt.Errorf("unsupported DOS single-class value 0x%02X", raw)
	}
}

func classLevelOffset(raw uint8) int {
	switch raw {
	case 0:
		return 0
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	default:
		return 0
	}
}

// ParseDOSPlayerSpellRecord parses only spell fields from an already
// decompressed player/creature record. It intentionally rejects truncated
// records instead of guessing at a format variant.
func ParseDOSPlayerSpellRecord(data []byte) (DOSPlayerSpellRecord, error) {
	if len(data) < DOSKnownSpellsEnd {
		return DOSPlayerSpellRecord{}, fmt.Errorf("DOS player record is %d bytes; need at least 0x%X", len(data), DOSKnownSpellsEnd)
	}

	result := DOSPlayerSpellRecord{
		MemorizedSpells: make([]uint8, 0, DOSMemorizedSpellsEnd-DOSMemorizedSpellsOffset),
		KnownSpells:     make([]uint8, 0, DOSKnownSpellsEnd-DOSKnownSpellsOffset),
	}
	for _, spellID := range data[DOSMemorizedSpellsOffset:DOSMemorizedSpellsEnd] {
		if spellID != 0 {
			result.MemorizedSpells = append(result.MemorizedSpells, spellID)
		}
	}
	for offset, known := range data[DOSKnownSpellsOffset:DOSKnownSpellsEnd] {
		if known != 0 {
			// The documented table is numbered from spell 1, while the
			// record stores the known flag at the corresponding offset.
			result.KnownSpells = append(result.KnownSpells, uint8(offset+1))
		}
	}
	return result, nil
}

// ApplyDOSSpellRecord replaces the remake character's ordered spell slots
// with the non-empty memorized slots from a verified DOS record.
func (c *Character) ApplyDOSSpellRecord(data []byte) error {
	if c == nil {
		return fmt.Errorf("cannot apply DOS spell record to nil character")
	}
	record, err := ParseDOSPlayerSpellRecord(data)
	if err != nil {
		return err
	}
	c.SpellSlots = append(c.SpellSlots[:0], record.MemorizedSpells...)
	c.KnownSpells = append(c.KnownSpells[:0], record.KnownSpells...)
	return nil
}

// ApplyDOSInventory replaces the remake equipment list with a decoded .SWG
// item stream. It is separate from the player pointer because .SAV/.GUY and
// .SWG are different files/regions in the original format.
func (c *Character) ApplyDOSInventory(data []byte) error {
	if c == nil {
		return fmt.Errorf("cannot apply inventory to nil character")
	}
	items, err := monster.ParseItems(data)
	if err != nil {
		return err
	}
	c.Equipment = append(c.Equipment[:0], items...)
	return nil
}

// ApplyDOSEffects replaces the preserved effect list from a DOS .FX stream.
// It does not change combat stats; callers can later interpret each effect
// through the game-specific AD&D rules layer.
func (c *Character) ApplyDOSEffects(data []byte) error {
	if c == nil {
		return fmt.Errorf("cannot apply effects to nil character")
	}
	effects, err := monster.ParseAffects(data)
	if err != nil {
		return err
	}
	c.Effects = append(c.Effects[:0], effects...)
	return nil
}
