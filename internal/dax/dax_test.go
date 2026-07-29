package dax

import (
	"encoding/binary"
	"testing"
)

func TestParseAndDecodeRLE(t *testing.T) {
	data := make([]byte, 2+9+5)
	binary.LittleEndian.PutUint16(data, 9)
	data[2] = 7
	binary.LittleEndian.PutUint16(data[7:9], 5)
	binary.LittleEndian.PutUint16(data[9:11], 5)
	copy(data[11:], []byte{0xfd, 'A', 0x01, 'B', 'C'})
	blocks, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Entry.ID != 7 {
		t.Fatalf("unexpected blocks: %#v", blocks)
	}
	if got := string(blocks[0].Data); got != "AAABC" {
		t.Fatalf("decoded %q, want AAABC", got)
	}
}

func TestRejectsShortLiteral(t *testing.T) {
	data := make([]byte, 2+9+2)
	binary.LittleEndian.PutUint16(data, 9)
	data[7] = 3
	data[11] = 2
	data[12] = 'A'
	if _, err := Parse(data); err == nil {
		t.Fatal("expected truncated literal error")
	}
}

func TestParsePC98RawBlock(t *testing.T) {
	data := make([]byte, 2+9+4)
	binary.LittleEndian.PutUint16(data, 9)
	data[2] = 0x50
	binary.LittleEndian.PutUint16(data[7:9], 3)
	binary.LittleEndian.PutUint16(data[9:11], 4)
	copy(data[11:], []byte{0xFF, 'E', 'C', 'L'})

	blocks, err := ParsePC98(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Entry.ID != 0x50 ||
		string(blocks[0].Data) != "ECL" {
		t.Fatalf("unexpected PC-98 raw blocks: %#v", blocks)
	}
}

func TestParsePC98CompressedBlock(t *testing.T) {
	// 05 is the codec flag and '.' is the shared fill byte:
	// literal ABC, five '.', then four Z.
	packed := []byte{
		0x05, '.',
		0x03, 'A', 'B', 'C',
		0xC3,
		0x81, 'Z',
	}
	data := make([]byte, 2+9+len(packed))
	binary.LittleEndian.PutUint16(data, 9)
	data[2] = 0x51
	binary.LittleEndian.PutUint16(data[7:9], 12)
	binary.LittleEndian.PutUint16(data[9:11], uint16(len(packed)))
	copy(data[11:], packed)

	blocks, err := ParsePC98(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(blocks[0].Data); got != "ABC.....ZZZZ" {
		t.Fatalf("decoded PC-98 block %q", got)
	}
}

func TestParsePC98RejectsDecodedOverflow(t *testing.T) {
	data := make([]byte, 2+9+3)
	binary.LittleEndian.PutUint16(data, 9)
	data[2] = 1
	binary.LittleEndian.PutUint16(data[7:9], 1)
	binary.LittleEndian.PutUint16(data[9:11], 3)
	copy(data[11:], []byte{0x05, 0, 0xC0})
	if _, err := ParsePC98(data); err == nil {
		t.Fatal("ParsePC98 accepted a repeat run beyond decoded size")
	}
}
