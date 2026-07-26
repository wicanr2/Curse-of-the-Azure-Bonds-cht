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
