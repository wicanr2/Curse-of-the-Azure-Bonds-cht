package main

import (
	"encoding/binary"
	"testing"
)

func TestSoundFXSelectorTableAndAddressMapping(t *testing.T) {
	t.Parallel()

	table := soundFXSelectorTable()
	if len(table) != 34 {
		t.Fatalf("table bytes=%d, want 34", len(table))
	}
	for index, want := range []int{
		255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	} {
		raw := int(binary.LittleEndian.Uint16(table[index*2:]))
		if raw != want {
			t.Fatalf("table[%d]=%d, want %d", index, raw, want)
		}
		address := uint16(0x4838 + index*2)
		got, ok := soundFXSelector(address)
		if !ok || got != want {
			t.Fatalf("soundFXSelector(%04X)=(%d,%v), want (%d,true)", address, got, ok, want)
		}
	}
	for _, address := range []uint16{0x4837, 0x4839, 0x485A} {
		if _, ok := soundFXSelector(address); ok {
			t.Fatalf("accepted out-of-table DS:%04X", address)
		}
	}
}
