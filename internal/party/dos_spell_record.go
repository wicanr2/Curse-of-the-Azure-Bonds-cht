package party

import "fmt"

// DOS player/creature records use a shared spell layout. These constants are
// deliberately limited to fields documented by the public CoAB format notes;
// the rest of the record is still left to a future parser.
const (
	DOSMemorizedSpellsOffset = 0x01E
	DOSMemorizedSpellsEnd    = 0x072 // exclusive; last documented byte is 0x071
	DOSKnownSpellsOffset     = 0x079
	DOSKnownSpellsEnd        = 0x0DD // exclusive; last documented byte is 0x0DC
)

// DOSPlayerSpellRecord is the verified spell subset of a DOS .SAV/.GUY
// creature record. MemorizedSpells preserves slot order and omits empty slots;
// KnownSpells contains one-based spell IDs whose known flag is set.
type DOSPlayerSpellRecord struct {
	MemorizedSpells []uint8
	KnownSpells     []uint8
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
	return nil
}
