package ecl

import "testing"

func TestRunSubsetExposesLoadFilesMapSelector(t *testing.T) {
	// LOAD FILES 0x12, 0x34, 0x10; EXIT. The third value is the original
	// GEO block selector used by CMD_LoadFiles when in a dungeon.
	payload := []byte{
		0x21,
		0x00, 0x12,
		0x00, 0x34,
		0x00, 0x10,
		0x00,
	}
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result.LoadFilesRequested || result.LoadFiles != [3]uint16{0x12, 0x34, 0x10} {
		t.Fatalf("load result=%+v, want three decoded selectors", result)
	}
}
