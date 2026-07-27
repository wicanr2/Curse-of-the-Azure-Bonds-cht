package save

import (
	"bytes"
	"testing"
)

func TestSAVGAMFixedPrefixRoundTrip(t *testing.T) {
	c := SAVGAMContainer{
		GameArea: 3, Area1: filledBytes(SAVGAMArea1Size, 0x11), Area2: filledBytes(SAVGAMArea2Size, 0x22),
		Runtime: filledBytes(SAVGAMRuntimeStateSize, 0x33), ECL: filledBytes(SAVGAMECLMemorySize, 0x44),
		MapPosX: -7, MapPosY: 13, MapDirection: 6, MapWallType: 0x81, MapWallRoof: 0x09,
		LastGameState: 0x12, GameState: 0x34, PartyCount: 2,
		SetBlocks:     [3]SAVGAMSetBlock{{BlockID: 0x1234, SetID: 0x5678}, {BlockID: 1, SetID: 2}},
		CharacterRefs: [8][]byte{[]byte("ALICE"), []byte("BOB")},
	}
	encoded, err := EncodeSAVGAM(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != SAVGAMFixedPrefixSize {
		t.Fatalf("encoded size = %d, want %d", len(encoded), SAVGAMFixedPrefixSize)
	}
	decoded, err := DecodeSAVGAM(append(encoded, 0xAA, 0xBB))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.GameArea != c.GameArea || decoded.MapPosX != c.MapPosX || decoded.MapPosY != c.MapPosY || decoded.PartyCount != c.PartyCount || decoded.SetBlocks != c.SetBlocks {
		t.Fatalf("decoded scalar state differs: %#v", decoded)
	}
	if !bytes.Equal(decoded.Area1, c.Area1) || !bytes.Equal(decoded.Area2, c.Area2) || !bytes.Equal(decoded.Runtime, c.Runtime) || !bytes.Equal(decoded.ECL, c.ECL) || !bytes.Equal(decoded.CharacterRefs[0][:len("ALICE")], []byte("ALICE")) || decoded.CharacterRefs[0][len("ALICE")] != 0 {
		t.Fatal("decoded raw state differs")
	}
}

func TestSAVGAMRejectsMalformedPrefix(t *testing.T) {
	if _, err := DecodeSAVGAM(make([]byte, SAVGAMFixedPrefixSize-1)); err == nil {
		t.Fatal("expected truncated prefix error")
	}
	c := SAVGAMContainer{Area1: make([]byte, SAVGAMArea1Size), Area2: make([]byte, SAVGAMArea2Size), Runtime: make([]byte, SAVGAMRuntimeStateSize), ECL: make([]byte, SAVGAMECLMemorySize), PartyCount: 9}
	if _, err := EncodeSAVGAM(c); err == nil {
		t.Fatal("expected party count error")
	}
	c.PartyCount = 0
	c.CharacterRefs[0] = make([]byte, SAVGAMCharacterRefSize+1)
	if _, err := EncodeSAVGAM(c); err == nil {
		t.Fatal("expected character ref size error")
	}
}

func filledBytes(size int, value byte) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = value
	}
	return data
}
