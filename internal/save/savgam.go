package save

import (
	"encoding/binary"
	"fmt"
)

// These sizes are the fixed prefix written by the reference SaveGame/loadSaveGame
// pair in ovr017.cs. The player files that follow this prefix are a separate
// side-effect stream and are intentionally not guessed by this codec.
const (
	SAVGAMGameAreaSize      = 1
	SAVGAMArea1Size         = 0x800
	SAVGAMArea2Size         = 0x800
	SAVGAMRuntimeStateSize  = 0x400
	SAVGAMECLMemorySize     = 0x1E00
	SAVGAMMapStateSize      = 5
	SAVGAMStateBytes        = 2
	SAVGAMSetBlockBytes     = 12
	SAVGAMPartyCountSize    = 1
	SAVGAMCharacterRefSize  = 0x29
	SAVGAMCharacterRefCount = 8
	SAVGAMCharacterRefsSize = SAVGAMCharacterRefSize * SAVGAMCharacterRefCount
	SAVGAMFixedPrefixSize   = SAVGAMGameAreaSize + SAVGAMArea1Size + SAVGAMArea2Size + SAVGAMRuntimeStateSize + SAVGAMECLMemorySize + SAVGAMMapStateSize + SAVGAMStateBytes + SAVGAMSetBlockBytes + SAVGAMPartyCountSize + SAVGAMCharacterRefsSize
)

// SAVGAMSetBlock is one of the three little-endian block/set pairs in SAVGAM.
type SAVGAMSetBlock struct {
	BlockID uint16
	SetID   uint16
}

// SAVGAMContainer is the fixed, proven prefix of a reference SAVGAM?.DAT
// slot. Raw slices are retained because Area/ECL field meanings are still
// being decoded and must not be replaced with invented structs.
type SAVGAMContainer struct {
	GameArea uint8
	Area1    []byte
	Area2    []byte
	Runtime  []byte
	ECL      []byte

	MapPosX       int8
	MapPosY       int8
	MapDirection  uint8
	MapWallType   uint8
	MapWallRoof   uint8
	LastGameState uint8
	GameState     uint8
	SetBlocks     [3]SAVGAMSetBlock
	PartyCount    uint8
	// CharacterRefs are fixed 0x29-byte CHRDAT name records. Unused records
	// are still present in the prefix and should normally remain zero-filled.
	CharacterRefs [SAVGAMCharacterRefCount][]byte
}

// EncodeSAVGAM serializes the fixed prefix in the exact reference order.
func EncodeSAVGAM(c SAVGAMContainer) ([]byte, error) {
	if err := validateSAVGAM(c); err != nil {
		return nil, err
	}
	data := make([]byte, SAVGAMFixedPrefixSize)
	offset := 0
	data[offset] = c.GameArea
	offset += SAVGAMGameAreaSize
	for _, part := range [][]byte{c.Area1, c.Area2, c.Runtime, c.ECL} {
		copy(data[offset:], part)
		offset += len(part)
	}
	data[offset] = byte(c.MapPosX)
	data[offset+1] = byte(c.MapPosY)
	data[offset+2] = c.MapDirection
	data[offset+3] = c.MapWallType
	data[offset+4] = c.MapWallRoof
	offset += SAVGAMMapStateSize
	data[offset] = c.LastGameState
	data[offset+1] = c.GameState
	offset += SAVGAMStateBytes
	for _, setBlock := range c.SetBlocks {
		binary.LittleEndian.PutUint16(data[offset:], setBlock.BlockID)
		binary.LittleEndian.PutUint16(data[offset+2:], setBlock.SetID)
		offset += 4
	}
	data[offset] = c.PartyCount
	offset += SAVGAMPartyCountSize
	for _, ref := range c.CharacterRefs {
		copy(data[offset:offset+SAVGAMCharacterRefSize], ref)
		offset += SAVGAMCharacterRefSize
	}
	return data, nil
}

// DecodeSAVGAM decodes the fixed prefix and ignores bytes after it. The
// ignored suffix is where SaveGame writes individual CHRDAT player files.
func DecodeSAVGAM(data []byte) (SAVGAMContainer, error) {
	if len(data) < SAVGAMFixedPrefixSize {
		return SAVGAMContainer{}, fmt.Errorf("SAVGAM prefix truncated: got %d bytes, need %d", len(data), SAVGAMFixedPrefixSize)
	}
	c := SAVGAMContainer{}
	offset := 0
	c.GameArea = data[offset]
	offset += SAVGAMGameAreaSize
	c.Area1 = append([]byte(nil), data[offset:offset+SAVGAMArea1Size]...)
	offset += SAVGAMArea1Size
	c.Area2 = append([]byte(nil), data[offset:offset+SAVGAMArea2Size]...)
	offset += SAVGAMArea2Size
	c.Runtime = append([]byte(nil), data[offset:offset+SAVGAMRuntimeStateSize]...)
	offset += SAVGAMRuntimeStateSize
	c.ECL = append([]byte(nil), data[offset:offset+SAVGAMECLMemorySize]...)
	offset += SAVGAMECLMemorySize
	c.MapPosX = int8(data[offset])
	c.MapPosY = int8(data[offset+1])
	c.MapDirection = data[offset+2]
	c.MapWallType = data[offset+3]
	c.MapWallRoof = data[offset+4]
	offset += SAVGAMMapStateSize
	c.LastGameState = data[offset]
	c.GameState = data[offset+1]
	offset += SAVGAMStateBytes
	for i := range c.SetBlocks {
		c.SetBlocks[i] = SAVGAMSetBlock{
			BlockID: binary.LittleEndian.Uint16(data[offset:]),
			SetID:   binary.LittleEndian.Uint16(data[offset+2:]),
		}
		offset += 4
	}
	c.PartyCount = data[offset]
	offset += SAVGAMPartyCountSize
	if c.PartyCount > SAVGAMCharacterRefCount {
		return SAVGAMContainer{}, fmt.Errorf("SAVGAM party count %d exceeds %d", c.PartyCount, SAVGAMCharacterRefCount)
	}
	for i := range c.CharacterRefs {
		ref := data[offset : offset+SAVGAMCharacterRefSize]
		c.CharacterRefs[i] = append([]byte(nil), ref...)
		offset += SAVGAMCharacterRefSize
	}
	return c, nil
}

func validateSAVGAM(c SAVGAMContainer) error {
	if len(c.Area1) != SAVGAMArea1Size {
		return fmt.Errorf("SAVGAM Area1 must be %d bytes, got %d", SAVGAMArea1Size, len(c.Area1))
	}
	if len(c.Area2) != SAVGAMArea2Size {
		return fmt.Errorf("SAVGAM Area2 must be %d bytes, got %d", SAVGAMArea2Size, len(c.Area2))
	}
	if len(c.Runtime) != SAVGAMRuntimeStateSize {
		return fmt.Errorf("SAVGAM runtime state must be %d bytes, got %d", SAVGAMRuntimeStateSize, len(c.Runtime))
	}
	if len(c.ECL) != SAVGAMECLMemorySize {
		return fmt.Errorf("SAVGAM ECL memory must be %d bytes, got %d", SAVGAMECLMemorySize, len(c.ECL))
	}
	if c.PartyCount > SAVGAMCharacterRefCount {
		return fmt.Errorf("SAVGAM party count %d exceeds %d", c.PartyCount, SAVGAMCharacterRefCount)
	}
	for i, ref := range c.CharacterRefs {
		if len(ref) > SAVGAMCharacterRefSize {
			return fmt.Errorf("SAVGAM character ref %d exceeds %d bytes", i, SAVGAMCharacterRefSize)
		}
	}
	return nil
}
