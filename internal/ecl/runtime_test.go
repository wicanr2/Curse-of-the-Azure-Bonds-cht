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
		t.Fatalf("result=%+v, want one HI text", result)
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

func TestRunSubsetOnGotoSelectsTargetAndConsumesTargetList(t *testing.T) {
	// ON GOTO index=1, count=2, targets +0x14 and +0x16, then EXIT.
	payload := make([]byte, 23)
	payload[0] = 0x25
	payload[1], payload[2] = 0x00, 1
	payload[3], payload[4] = 0x00, 2
	payload[5], payload[6], payload[7] = 0x02, 0x14, 0x80
	payload[8], payload[9], payload[10] = 0x02, 0x16, 0x80
	payload[22] = 0x00
	block := append([]byte{0, 0}, payload...)
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Steps != 2 || result.PC != 23 {
		t.Fatalf("result=%+v, want two steps and stop at 23", result)
	}
}

func TestRunSubsetArithmeticWritesMemory(t *testing.T) {
	// MULTIPLY 3, 4 -> memory[0x9000], then PRINT memory[0x9000].
	payload := []byte{
		0x07, 0x00, 3, 0x00, 4, 0x02, 0x00, 0x90,
		0x11, 0x01, 0x00, 0x90,
		0x00,
	}
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "12" {
		t.Fatalf("text=%q, want [12]", result.Text)
	}
}

func TestRunSubsetGetTableReadsIndexedMemory(t *testing.T) {
	// GETTABLE memory[0x9000 + 1] -> memory[0x9100], then print it.
	payload := []byte{
		0x2A, 0x02, 0x00, 0x90, 0x00, 1, 0x02, 0x00, 0x91,
		0x11, 0x01, 0x00, 0x91,
		0x00,
	}
	// Seed memory through SAVE before GETTABLE.
	payload = append([]byte{0x09, 0x00, 7, 0x02, 0x01, 0x90}, payload...)
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "7" {
		t.Fatalf("text=%q, want [7]", result.Text)
	}
}

func TestRunSubsetCompareAndFeedsIf(t *testing.T) {
	// COMPARE AND (1,1) (2,2); IF =; PRINT "YES"; EXIT.
	payload := []byte{
		0x14,
		0x00, 1, 0x00, 1, 0x00, 2, 0x00, 2,
		0x16,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v, want one HI text", result)
	}
}

func TestRunSubsetHorizontalMenuExtractsOptions(t *testing.T) {
	// HORIZONTAL MENU memory[0x9000], two packed options, EXIT.
	payload := []byte{
		0x2B, 0x02, 0x00, 0x90, 0x00, 2,
		0x80, 0x02, 0x20, 0x92,
		0x80, 0x02, 0x0C, 0x32,
		0x00,
	}
	result, err := RunSubsetWithSelections(append([]byte{0, 0}, payload...), 0, 10, []uint16{1})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Menus) != 1 || len(result.Menus[0].Options) != 2 || result.Menus[0].Selected != 1 {
		t.Fatalf("menus=%+v, want two options with selected 1", result.Menus)
	}
}

func TestRunSubsetVerticalMenuExtractsPromptAndOptions(t *testing.T) {
	// VERTICAL MENU memory[0x9000], prompt "HI", two options, EXIT.
	payload := []byte{
		0x15, 0x02, 0x00, 0x90, 0x80, 0x02, 0x20, 0x92, 0x00, 2,
		0x80, 0x02, 0x20, 0x92,
		0x80, 0x02, 0x0C, 0x32,
		0x00,
	}
	result, err := RunSubsetWithSelections(append([]byte{0, 0}, payload...), 0, 10, []uint16{1})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Menus) != 1 || !result.Menus[0].Vertical || result.Menus[0].Prompt != "HI" || result.Menus[0].Selected != 1 {
		t.Fatalf("menus=%+v, want vertical HI menu selected 1", result.Menus)
	}
}

func TestRunSubsetInteractivePausesBeforeUnselectedMenu(t *testing.T) {
	payload := []byte{
		0x2B, 0x02, 0x00, 0x90, 0x00, 1,
		0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubsetInteractive(append([]byte{0, 0}, payload...), 0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.WaitingForMenu || len(result.Menus) != 1 || result.Menus[0].Options[0] != "HI" {
		t.Fatalf("result=%+v, want waiting menu", result)
	}
}

func TestRunSubsetAcceptsEmptyPackedString(t *testing.T) {
	block := []byte{0, 0, 0x11, 0x80, 0x00, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "" {
		t.Fatalf("text=%q, want one empty string", result.Text)
	}
}
