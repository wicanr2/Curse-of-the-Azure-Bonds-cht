// Package area contains the platform-neutral fields that the original
// Area1/Area2 loader exposes to map and ECL subsystems.
package area

type State struct {
	GameArea            uint8
	InDungeon           bool
	Current3DMapBlockID uint8
	CurrentCity         uint8
	LastXPos            int16
	LastYPos            int16
	LastECLBlockID      uint16
}

type LoadFilesEffect struct {
	GeoMapBlock *uint8
	BigPicture  bool
}

// ApplyLoadFiles mirrors the proven CMD_LoadFiles branch. Values are the
// three ECL operands in original order; operand 2 (index 2) is the GEO block
// selector when the area is a dungeon.
func (s *State) ApplyLoadFiles(values [3]uint16, lastDAXBlockID uint8) LoadFilesEffect {
	var effect LoadFilesEffect
	mapBlock := values[2]
	if s.InDungeon && mapBlock != 0xFF && mapBlock != 0x7F {
		block := uint8(mapBlock)
		s.Current3DMapBlockID = block
		effect.GeoMapBlock = &block
	}
	if !s.InDungeon && values[0] != 0xFF && lastDAXBlockID != 0x50 {
		effect.BigPicture = true
	}
	return effect
}
