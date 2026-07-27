package area

import (
	"encoding/binary"
	"testing"
)

func TestArea1CodecReadsKnownFieldsAndPreservesUnknownBytes(t *testing.T) {
	raw := make([]byte, SnapshotSize)
	raw[0x10] = 0xA5
	raw[area1MapBlock] = 0x42
	binary.LittleEndian.PutUint16(raw[area1Dungeon:], 1)
	lastX := int16(-12)
	lastY := int16(34)
	binary.LittleEndian.PutUint16(raw[area1LastX:], uint16(lastX))
	binary.LittleEndian.PutUint16(raw[area1LastY:], uint16(lastY))
	binary.LittleEndian.PutUint16(raw[area1LastECL:], 0x1234)
	raw[area1City] = 7
	binary.LittleEndian.PutUint16(raw[area1OutdoorSky:], 5)
	binary.LittleEndian.PutUint16(raw[area1IndoorSky:], 9)
	for index := 0; index < 7; index++ {
		binary.LittleEndian.PutUint16(raw[area1GameTime+index*2:], uint16(index+3))
	}

	state, err := DecodeArea1(raw)
	if err != nil {
		t.Fatal(err)
	}
	if state.Current3DMapBlockID != 0x42 || !state.InDungeon || state.LastXPos != -12 || state.LastYPos != 34 || state.LastECLBlockID != 0x1234 || state.CurrentCity != 7 || state.OutdoorSkyColor != 5 || state.IndoorSkyColor != 9 || state.GameTime != [7]uint16{3, 4, 5, 6, 7, 8, 9} {
		t.Fatalf("decoded state mismatch: %+v", state)
	}
	state.Current3DMapBlockID = 0x45
	state.InDungeon = false
	state.LastXPos = -99
	state.OutdoorSkyColor = 6
	state.IndoorSkyColor = 10
	state.GameTime[6] = 99
	encoded, err := EncodeArea1(state, raw)
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0x10] != 0xA5 || encoded[area1LastECL] != 0x34 || encoded[area1MapBlock] != 0x45 {
		t.Fatalf("unknown bytes were not preserved")
	}
	if got := int16(binary.LittleEndian.Uint16(encoded[area1LastX:])); got != -99 {
		t.Fatalf("last x = %d, want -99", got)
	}
	if got := binary.LittleEndian.Uint16(encoded[area1OutdoorSky:]); got != 6 || binary.LittleEndian.Uint16(encoded[area1IndoorSky:]) != 10 {
		t.Fatalf("sky colours=(%d,%d), want (6,10)", got, binary.LittleEndian.Uint16(encoded[area1IndoorSky:]))
	}
	if got := binary.LittleEndian.Uint16(encoded[area1GameTime+12:]); got != 99 {
		t.Fatalf("game time slot 6=%d, want 99", got)
	}
}

func TestArea2CodecAndRecordSizes(t *testing.T) {
	raw := make([]byte, SnapshotSize)
	raw[0x20] = 0xCC
	raw[area2GameArea] = 3
	raw[area2HeadBlock] = 0x12
	state, err := DecodeArea2(raw)
	if err != nil || state.GameArea != 3 || state.HeadBlockID != 0x12 {
		t.Fatalf("decode Area2: state=%+v err=%v", state, err)
	}
	state.GameArea = 9
	state.HeadBlockID = 0x34
	encoded, err := EncodeArea2(state, raw)
	if err != nil || encoded[area2GameArea] != 9 || encoded[area2HeadBlock] != 0x34 || encoded[0x20] != 0xCC {
		t.Fatalf("encode Area2 failed: err=%v", err)
	}
	if _, err := DecodeArea1(make([]byte, SnapshotSize-1)); err == nil {
		t.Fatal("short Area1 record accepted")
	}
	if _, err := EncodeArea2(state, make([]byte, SnapshotSize+1)); err == nil {
		t.Fatal("long Area2 record accepted")
	}
}
