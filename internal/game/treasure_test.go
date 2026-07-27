package game

import (
	"encoding/binary"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func TestParseTreasureItemBlocksScopesDAXBlocksByArea(t *testing.T) {
	item := make([]byte, monster.ItemRecordSize)
	item[0x2E] = 36
	item[0x39] = 1
	data := testDAXBlock(7, item)
	blocks, err := ParseTreasureItemBlocks(map[uint8][]byte{1: data, 2: data})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks[0x0107]) != 1 || len(blocks[0x0207]) != 1 || blocks[0x0107][0].Type != 36 {
		t.Fatalf("blocks=%#v", blocks)
	}
}

func testDAXBlock(id byte, decoded []byte) []byte {
	// One literal run: control count-1, followed by the fixed item bytes.
	data := make([]byte, 2+9+1+len(decoded))
	binary.LittleEndian.PutUint16(data[:2], 9)
	data[2] = id
	binary.LittleEndian.PutUint16(data[7:9], uint16(len(decoded)))
	binary.LittleEndian.PutUint16(data[9:11], uint16(len(decoded)+1))
	data[11] = byte(len(decoded) - 1)
	copy(data[12:], decoded)
	return data
}
