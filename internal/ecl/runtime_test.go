package ecl

import "testing"

func TestRunSubsetPrintsPackedTextAndStops(t *testing.T) {
	// payload: PRINT [packed "HI"] ; EXIT
	block := []byte{0, 0, 0x11, 0x80, 0x02, 0x20, 0x92, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("text=%q, want [HI]", result.Text)
	}
	if result.Steps != 2 {
		t.Fatalf("steps=%d, want 2", result.Steps)
	}
}

func TestRunSubsetIFSkipsOneCompleteCommand(t *testing.T) {
	// COMPARE 1, 2; IF =; PRINT "NO"; PRINT "YES"; EXIT.
	block := []byte{0, 0,
		0x03, 0x00, 1, 0x00, 2,
		0x16,
		0x11, 0x80, 0x02, 0xD8, 0x00,
		0x11, 0x80, 0x02, 0xD8, 0x00,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 {
		t.Fatalf("text=%q, want one skipped-branch result", result.Text)
	}
}
