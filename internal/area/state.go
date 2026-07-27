// Package area contains the platform-neutral fields that the original
// Area1/Area2 loader exposes to map and ECL subsystems.
package area

type State struct {
	GameArea uint8
	// HeadBlockID mirrors Area2.HeadBlockId at 0x5C2. 0xFF means PICTURE
	// uses PIC/BIGPIC instead of the HEAD/BODY scene branch.
	HeadBlockID         uint8
	InDungeon           bool
	Current3DMapBlockID uint8
	CurrentCity         uint8
	LastXPos            int16
	LastYPos            int16
	LastECLBlockID      uint16
	OutdoorSkyColor     uint16
	IndoorSkyColor      uint16
	// GameTime mirrors the seven Area1 words at 0x18C..0x198 in reference
	// order, allowing the DOS SAVGAM clock to round-trip.
	GameTime [7]uint16
}

type LoadFilesEffect struct {
	GeoMapBlock *uint8
	BigPicture  bool
}

// ApplyLoadFiles mirrors the proven CMD_LoadFiles branch. Values are the
// three ECL operands in original order. Reference CMD_LoadFiles names operand
// 1 var_3 and uses it as the GEO block in a dungeon; operand 3 (var_1) drives
// the outdoor BIGPIC reload branch.
func (s *State) ApplyLoadFiles(values [3]uint16, lastDAXBlockID uint8) LoadFilesEffect {
	var effect LoadFilesEffect
	mapBlock := values[0]
	if s.InDungeon && mapBlock != 0xFF && mapBlock != 0x7F {
		block := uint8(mapBlock)
		s.Current3DMapBlockID = block
		effect.GeoMapBlock = &block
	}
	if !s.InDungeon && values[2] != 0xFF && lastDAXBlockID != 0x50 {
		effect.BigPicture = true
	}
	return effect
}
