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

// 角色檔名那 41 bytes 是 Turbo Pascal 的 string[40]：第一個位元組是長度。
// 少了它，原版會把第一個字元當成長度去組檔名（spec 1072；原版自己寫的
// BOB.GUY 開頭也是 03 'BOB'）。
func TestSAVGAMCharacterRefIsPascalString(t *testing.T) {
	ref, err := SAVGAMCharacterRef("CHRDATA1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ref) != SAVGAMCharacterRefSize {
		t.Fatalf("ref 是 %d bytes，want %d", len(ref), SAVGAMCharacterRefSize)
	}
	if ref[0] != 8 || string(ref[1:9]) != "CHRDATA1" {
		t.Fatalf("ref 開頭是 %v，want 08 CHRDATA1", ref[:9])
	}
	for _, tail := range ref[9:] {
		if tail != 0 {
			t.Fatalf("長度之外的位元組沒有清零：%v", ref)
		}
	}
	if _, err := SAVGAMCharacterRef(""); err == nil {
		t.Fatal("空名字應該被擋下來")
	}
	if _, err := SAVGAMCharacterRef(string(make([]byte, SAVGAMCharacterRefSize))); err == nil {
		t.Fatal("超過 40 字元應該被擋下來")
	}
}
