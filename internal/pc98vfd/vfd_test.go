package pc98vfd

import (
	"encoding/binary"
	"testing"
)

func TestParsePreservesMissingSector(t *testing.T) {
	const dataOffset = 0x300
	data := make([]byte, dataOffset+1024)
	copy(data, signature[:])
	entry := data[descriptorOffset : descriptorOffset+descriptorSize]
	entry[0], entry[1], entry[2], entry[3] = 0, 0, 1, 3
	copy(entry[4:8], []byte{0, 0, 1, 1})
	binary.LittleEndian.PutUint32(entry[8:12], ^uint32(0))

	image, err := Parse(data, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := image.MissingSectors(); len(got) != 1 {
		t.Fatalf("missing sectors = %d, want 1", len(got))
	}
	if got := image.Sectors[0].Size(); got != 1024 {
		t.Fatalf("sector size = %d, want 1024", got)
	}
}

func TestParseReadsPayload(t *testing.T) {
	const dataOffset = 0x300
	data := make([]byte, dataOffset+1024)
	copy(data, signature[:])
	entry := data[descriptorOffset : descriptorOffset+descriptorSize]
	entry[0], entry[1], entry[2], entry[3] = 0, 0, 1, 3
	binary.LittleEndian.PutUint32(entry[8:12], dataOffset)
	data[dataOffset] = 0xA5

	image, err := Parse(data, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if image.Sectors[0].Missing() {
		t.Fatal("payload sector reported missing")
	}
	if got := image.Sectors[0].Data[0]; got != 0xA5 {
		t.Fatalf("payload byte = %02x, want a5", got)
	}
}

func TestParseRejectsUnexpectedGeometry(t *testing.T) {
	data := make([]byte, 0x400)
	copy(data, signature[:])
	entry := data[descriptorOffset : descriptorOffset+descriptorSize]
	entry[0], entry[1], entry[2], entry[3] = 1, 0, 1, 3
	binary.LittleEndian.PutUint32(entry[8:12], 0x300)

	if _, err := Parse(data, 1, 1, 1); err == nil {
		t.Fatal("Parse accepted a descriptor with the wrong cylinder")
	}
}
