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

func TestRunSubsetReportsNewECLBlock(t *testing.T) {
	block := []byte{0, 0, 0x20, 0x00, 0x51, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.NewECLBlockID == nil || *result.NewECLBlockID != 0x51 {
		t.Fatalf("result=%+v, want NEWECL 0x51", result)
	}
}

func TestRunSubsetReportsCombatRequest(t *testing.T) {
	result, err := RunSubset([]byte{0, 0, 0x24, 0x00}, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result.CombatRequested || result.Steps != 1 || result.PC != 1 {
		t.Fatalf("result=%+v, want one combat request", result)
	}
}

func TestRunSubsetReportsPictureRequest(t *testing.T) {
	result, err := RunSubset([]byte{0, 0, 0x0E, 0x00, 0x1D, 0x00}, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result.PictureRequested || result.PictureBlock != 0x1D {
		t.Fatalf("result=%+v, want PIC block 0x1D", result)
	}
	if result.BigPictureRequested {
		t.Fatal("regular PIC was marked as BIGPIC")
	}
	big, err := RunSubset([]byte{0, 0, 0x0E, 0x00, 0x78, 0x00}, 0, 10)
	if err != nil || !big.BigPictureRequested || big.PictureBlock != 0x78 {
		t.Fatalf("big picture result=%+v err=%v", big, err)
	}
}

func TestRunSubsetCarriesEncounterDescriptorsToCombat(t *testing.T) {
	// SETUP MONSTER 4,2,4; LOAD MONSTER 0x56,10,0x56; COMBAT.
	block := []byte{0, 0,
		0x0C, 0, 4, 0, 2, 0, 4,
		0x0B, 0, 0x56, 0, 10, 0, 0x56,
		0x24,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result.CombatRequested || result.MonsterSetup == nil || len(result.MonsterSpawns) != 1 {
		t.Fatalf("result=%+v", result)
	}
	if result.MonsterSetup.SpriteID != 4 || result.MonsterSpawns[0].MonsterID != 0x56 || result.MonsterSpawns[0].Count != 10 {
		t.Fatalf("encounter=%+v setup=%+v", result.MonsterSpawns, result.MonsterSetup)
	}
}

func TestRunSubsetRecordsProgramAndContinuesBoundedTrace(t *testing.T) {
	// PROGRAM 9; the external routine is an explicit VM boundary.
	result, err := RunSubset([]byte{0, 0, 0x38, 0, 9, 0x00}, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ProgramIDs) != 1 || result.ProgramIDs[0] != 9 || !result.ProgramExit || result.Steps != 1 {
		t.Fatalf("result=%+v", result)
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

func TestRunSubsetConsumesTreasureOperandsAsBoundedNoOp(t *testing.T) {
	// TREASURE has eight decoded operands. The bounded runner must consume
	// them before continuing to EXIT, without inventing inventory effects.
	payload := []byte{0x27}
	for i := 0; i < 8; i++ {
		payload = append(payload, 0x00, 0x00)
	}
	payload = append(payload, 0x00)
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Steps != 2 || result.PC != len(payload) {
		t.Fatalf("result=%+v, want TREASURE then EXIT", result)
	}
}

func TestRunSubsetEmitsSpellAndProtectionSignals(t *testing.T) {
	payload := []byte{
		0x3B, 0x00, 0x12, 0x01, 0x00, 0x7C, 0x01, 0x01, 0x7C,
		0x3C, 0x01, 0x02, 0x7C,
		0x00,
	}
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.SpellSearches) != 1 {
		t.Fatalf("spell searches=%#v", result.SpellSearches)
	}
	spell := result.SpellSearches[0]
	if spell.SpellID != 0x12 || spell.SpellSlotAddress != 0x7C00 || spell.CharacterAddress != 0x7C01 {
		t.Fatalf("spell=%#v", spell)
	}
	if len(result.ProtectionRequests) != 1 || result.ProtectionRequests[0] != 0x7C02 {
		t.Fatalf("protection=%#v", result.ProtectionRequests)
	}
	if result.Steps != 3 {
		t.Fatalf("steps=%d, want 3", result.Steps)
	}
}

func TestRunSubsetRandomUsesInclusiveRangeAndSeed(t *testing.T) {
	// RANDOM 3, 0x0100; SAVE 0x0100, 0x0101; EXIT.
	block := []byte{0, 0, 0x08, 0x00, 0x03, 0x01, 0x00, 0x01, 0x09, 0x01, 0x00, 0x01, 0x01, 0x01, 0x01, 0x00}
	first, err := RunSubsetWithSelectionsSeed(block, 0, 8, nil, 99)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RunSubsetWithSelectionsSeed(block, 0, 8, nil, 99)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.RandomValues) != 1 || first.RandomValues[0] > 3 || len(second.RandomValues) != 1 || first.RandomValues[0] != second.RandomValues[0] {
		t.Fatalf("seeded random values=%v,%v", first.RandomValues, second.RandomValues)
	}
	if first.Steps != second.Steps || first.PC != second.PC {
		t.Fatalf("seeded runs diverged: first=%+v second=%+v", first, second)
	}
}

func TestRunSubsetEncounterMenuPausesAndMapsSelection(t *testing.T) {
	payload := []byte{0x29}
	appendScalar := func(value byte) { payload = append(payload, 0x00, value) }
	appendWord := func(value uint16) { payload = append(payload, 0x01, byte(value), byte(value>>8)) }
	appendPacked := func(value []byte) {
		payload = append(payload, 0x80, byte(len(value)))
		payload = append(payload, value...)
	}
	appendScalar(8)
	appendScalar(2)
	appendScalar(3)
	appendWord(0x0200)
	appendScalar(1)
	appendScalar(2)
	appendScalar(3)
	appendScalar(4)
	appendScalar(5)
	appendPacked([]byte{0x00})
	appendPacked([]byte{0x00})
	appendPacked([]byte{0x00})
	appendScalar(0)
	appendScalar(0)
	payload = append(payload, 0x00)
	block := append([]byte{0, 0}, payload...)
	waiting, err := RunSubsetInteractive(block, 0, 8, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !waiting.WaitingForMenu || len(waiting.Menus) != 1 || waiting.Menus[0].Options[3] != "ADVANCE" {
		t.Fatalf("waiting result=%+v", waiting)
	}
	selected, err := RunSubsetInteractive(block, 0, 8, []uint16{3})
	if err != nil {
		t.Fatal(err)
	}
	if selected.WaitingForMenu || len(selected.EncounterActions) != 1 || selected.EncounterActions[0] != 4 {
		t.Fatalf("selected result=%+v", selected)
	}
}

func TestRuntimeStateResumesAtPausedMenu(t *testing.T) {
	block := []byte{0, 0,
		0x2B, 0x02, 0x00, 0x90, 0x00, 0x01,
		0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	runtime := NewRuntimeState(0)
	first, err := runSubsetWithState(block, 0, 8, nil, true, 1, runtime)
	if err != nil || !first.WaitingForMenu {
		t.Fatalf("first run=%+v err=%v, want menu pause", first, err)
	}
	if runtime.PC != 0 || !runtime.Started {
		t.Fatalf("runtime=%+v, want resumable menu PC 0", runtime)
	}
	second, err := runSubsetWithState(block, 0, 8, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if second.WaitingForMenu || len(second.Menus) != 1 || second.Menus[0].Selected != 0 {
		t.Fatalf("second run=%+v, want selected menu to continue", second)
	}
	if runtime.Memory[0x9000] != 0 {
		t.Fatalf("runtime memory=%#v, want selected value retained", runtime.Memory)
	}
}
