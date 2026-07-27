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
	if !result.Exited {
		t.Fatal("EXIT command did not expose lifecycle completion")
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

func TestRunSubsetAddNPCExposesIDAndContinuesToExit(t *testing.T) {
	block := []byte{0, 0, 0x36, 0x00, 0x55, 0x00, 0x64, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NPCIDs) != 1 || result.NPCIDs[0] != 0x55 ||
		len(result.NPCRequests) != 1 || result.NPCRequests[0] != (NPCRequest{ID: 0x55, Morale: 100}) ||
		result.Steps != 2 || result.PC != 6 {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetLoadPiecesExposesThreeSelectors(t *testing.T) {
	block := []byte{0, 0, 0x37, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil || !result.LoadPiecesRequested || result.LoadPieces != [3]uint16{1, 2, 3} {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if result.Steps != 2 || result.PC != 8 {
		t.Fatalf("steps=%d pc=%d, want two steps and exit at 8", result.Steps, result.PC)
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

func TestRunSubsetBitwiseWritesMemory(t *testing.T) {
	// AND 0xF0F0, 0x0FF0 -> 0x9000; OR 0x9000, 0x000F -> 0x9001.
	payload := []byte{
		0x2F, 0x02, 0xF0, 0xF0, 0x02, 0xF0, 0x0F, 0x02, 0x00, 0x90,
		0x30, 0x01, 0x00, 0x90, 0x00, 0x0F, 0x02, 0x01, 0x90,
		0x11, 0x01, 0x00, 0x90,
		0x11, 0x01, 0x01, 0x90,
		0x00,
	}
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 2 || result.Text[0] != "240" || result.Text[1] != "255" {
		t.Fatalf("text=%q, want [240 255]", result.Text)
	}
}

func TestRunSubsetBitwiseResultFeedsIfFlags(t *testing.T) {
	// AND 0, 0 compares the result with zero. IF <> must therefore skip the
	// first EXIT and reach SAVE/PRINT.
	block := []byte{0, 0,
		0x2F, 0x00, 0, 0x00, 0, 0x02, 0x00, 0x7B,
		0x17,
		0x00,
		0x09, 0x00, 7, 0x02, 0x01, 0x7B,
		0x11, 0x01, 0x01, 0x7B,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "7" {
		t.Fatalf("text=%q, want [7] after AND compare flags", result.Text)
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

func TestRunSubsetSaveTableWritesIndexedMemory(t *testing.T) {
	// SAVE 7 -> memory[0x9000]; SAVE TABLE memory[0x9000] to
	// memory[0x9100+2]; then print the indexed destination.
	payload := []byte{
		0x09, 0x00, 7, 0x02, 0x00, 0x90,
		0x35, 0x01, 0x00, 0x90, 0x02, 0x00, 0x91, 0x00, 2,
		0x11, 0x01, 0x02, 0x91,
		0x00,
	}
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

func TestRunSubsetReportsECLClockOperands(t *testing.T) {
	// ECL CLOCK timeStep=3, timeSlot=1; the reference command loads two
	// operands before calling step_game_time.
	block := []byte{0, 0, 0x34, 0, 3, 0, 1, 0}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ClockRequests) != 1 || result.ClockRequests[0] != (ClockRequest{TimeStep: 3, TimeSlot: 1}) {
		t.Fatalf("clock result=%+v", result.ClockRequests)
	}
}

func TestRunSubsetReportsPictureRequest(t *testing.T) {
	result, err := RunSubset([]byte{0, 0,
		0x09, 0x00, 3, 0x02, 0xE1, 0x7E,
		0x0E, 0x00, 0x1D,
		0x00,
	}, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result.PictureRequested || result.PictureBlock != 0x1D {
		t.Fatalf("result=%+v, want PIC block 0x1D", result)
	}
	if result.BigPictureRequested {
		t.Fatal("regular PIC was marked as BIGPIC")
	}
	if !result.PictureHeadBlockSet || result.PictureHeadBlock != 3 {
		t.Fatalf("picture head selector=%d,%v, want 3,true", result.PictureHeadBlock, result.PictureHeadBlockSet)
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

func TestRunSubsetRecordsExternalCallAndReturnsToECL(t *testing.T) {
	// CALL [0x2E10] is an external engine routine; the bounded VM exposes the
	// address and continues with the following localized text.
	block := []byte{0, 0,
		0x2D, 0x01, 0x10, 0x2E,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CallAddresses) != 1 || result.CallAddresses[0] != 0x2E10 {
		t.Fatalf("calls=%#v", result.CallAddresses)
	}
	if len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("text=%q, want [HI]", result.Text)
	}
}

func TestRunSubsetRecordsPrintReturnAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x33,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.PrintReturnCount != 1 || len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetRecordsLoadCharacterAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x0A, 0x01, 0x79, 0x7F,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.LoadCharacterAddresses) != 1 || result.LoadCharacterAddresses[0] != 0x7F79 || len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetDecodesLoadCharacterPlayerSelector(t *testing.T) {
	block := []byte{0, 0,
		0x0A, 0x02, 0x81, 0x00,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.LoadCharacterRequests) != 1 {
		t.Fatalf("requests=%#v", result.LoadCharacterRequests)
	}
	request := result.LoadCharacterRequests[0]
	if request.Address != 0x0081 || request.Value != 0x0081 || request.PlayerIndex != 1 || !request.HighBitSet {
		t.Fatalf("request=%#v", request)
	}
}

func TestRunSubsetLoadCharacterFeedsSelectedNameStringMemory(t *testing.T) {
	// LOAD CHARACTER 1; LOAD CHARACTER 0 (not found); COMPARE [0x7C00],
	// packed "HI"; IF =; GOTO success; PRINT "NO"; EXIT;
	// success: PRINT "YES"; EXIT. The failed lookup preserves the last player.
	block := []byte{0, 0,
		0x0A, 0x02, 0x01, 0x00,
		0x0A, 0x02, 0x00, 0x00,
		0x03, 0x81, 0x00, 0x7C, 0x80, 0x02, 0x20, 0x92,
		0x16,
		0x01, 0x02, 0x1C, 0x80,
		0x11, 0x80, 0x03, 0x38, 0xF0, 0x00,
		0x00,
		0x11, 0x80, 0x03, 0x64, 0x54, 0xC0,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 20, nil, 1, PartyContext{
		Members: []PartyMemberContext{{Name: "HI"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "YES" {
		t.Fatalf("text=%q, want selected-name comparison to print YES", result.Text)
	}
}

func TestRunSubsetRecordsFindItemQueryAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x32, 0x00, 0x5E,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FindItemIDs) != 1 || result.FindItemIDs[0] != 0x5E || len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetFindItemResolvesPartyInventoryCompareFlags(t *testing.T) {
	block := []byte{0, 0,
		0x32, 0x00, 0x5E,
		0x16,
		0x01, 0x02, 0x0F, 0x80,
		0x11, 0x80, 0x03, 0x38, 0xF0, 0x00,
		0x00,
		0x11, 0x80, 0x03, 0x64, 0x54, 0xC0,
		0x00,
	}
	for _, test := range []struct {
		name     string
		items    []uint8
		wantText string
		want     bool
	}{{name: "found", items: []uint8{0x5E}, wantText: "YES", want: true}, {name: "not found", wantText: "NO"}} {
		t.Run(test.name, func(t *testing.T) {
			result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 20, nil, 1, PartyContext{
				Members: []PartyMemberContext{{ItemTypes: test.items}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Text) != 1 || result.Text[0] != test.wantText || len(result.FindItemRequests) != 1 || !result.FindItemRequests[0].Resolved || result.FindItemRequests[0].Found != test.want {
				t.Fatalf("result=%+v, want text %q found %v", result, test.wantText, test.want)
			}
		})
	}
}

func TestRunSubsetDestroyItemsUpdatesWorkingInventoryQueries(t *testing.T) {
	block := []byte{0, 0,
		0x32, 0x00, 0x5E,
		0x40, 0x00, 0x5E,
		0x32, 0x00, 0x5E,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 10, nil, 1, PartyContext{
		Members: []PartyMemberContext{{ItemTypes: []uint8{0x5E}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FindItemRequests) != 2 || !result.FindItemRequests[0].Found || result.FindItemRequests[1].Found {
		t.Fatalf("requests=%#v", result.FindItemRequests)
	}
}

func TestRunSubsetFindSpecialUsesLoadCharacterSelection(t *testing.T) {
	block := []byte{0, 0,
		0x0A, 0x02, 0x01, 0x00,
		0x3F, 0x00, 0x27,
		0x16,
		0x01, 0x02, 0x13, 0x80,
		0x11, 0x80, 0x03, 0x38, 0xF0, 0x00,
		0x00,
		0x11, 0x80, 0x03, 0x64, 0x54, 0xC0,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 20, nil, 1, PartyContext{
		Members: []PartyMemberContext{{Effects: []uint8{0x27}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "YES" || len(result.FindSpecialRequests) != 1 || !result.FindSpecialRequests[0].Resolved || !result.FindSpecialRequests[0].Found || result.FindSpecialRequests[0].SelectedPlayerIndex != 0 {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetFindSpecialUsesWhoSelectionAfterResume(t *testing.T) {
	block := []byte{0, 0,
		0x39, 0x00, 0x00,
		0x3F, 0x00, 0x27,
		0x00,
	}
	context := PartyContext{Members: []PartyMemberContext{{}, {Effects: []uint8{0x27}}}}
	runtime := NewRuntimeState(0)
	paused, err := runSubsetWithStateContextAndWhoSelections(block, 0, 10, nil, nil, true, 1, runtime, &context)
	if err != nil {
		t.Fatal(err)
	}
	if !paused.WaitingForWho || runtime.PC != 0 {
		t.Fatalf("paused=%+v runtime=%+v", paused, runtime)
	}
	result, err := runSubsetWithStateContextAndWhoSelections(block, 0, 10, nil, []uint16{1}, true, 1, runtime, &context)
	if err != nil {
		t.Fatal(err)
	}
	if result.WaitingForWho || len(result.FindSpecialRequests) != 1 || !result.FindSpecialRequests[0].Found || result.FindSpecialRequests[0].SelectedPlayerIndex != 1 {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetDumpRemovesSelectedFromWorkingParty(t *testing.T) {
	block := []byte{0, 0,
		0x0A, 0x02, 0x02, 0x00,
		0x3E,
		0x3F, 0x00, 0x27,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 10, nil, 1, PartyContext{
		Members: []PartyMemberContext{{Name: "A", Effects: []uint8{0x27}}, {Name: "B", Effects: []uint8{0x2A}}, {Name: "C"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.DumpRequests) != 1 {
		t.Fatalf("dump requests=%#v", result.DumpRequests)
	}
	dump := result.DumpRequests[0]
	if !dump.Resolved || dump.SelectedPlayerIndex != 1 || !dump.NextSelectedPlayerSet || dump.NextSelectedPlayerIndex != 0 {
		t.Fatalf("dump=%#v", dump)
	}
	if len(result.FindSpecialRequests) != 1 || !result.FindSpecialRequests[0].Resolved || !result.FindSpecialRequests[0].Found || result.FindSpecialRequests[0].SelectedPlayerIndex != 0 {
		t.Fatalf("find special=%#v", result.FindSpecialRequests)
	}
}

func TestRunSubsetRecordsDestroyItemRequestAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x40, 0x00, 0x5E,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.DestroyItemIDs) != 1 || result.DestroyItemIDs[0] != 0x5E || len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunSubsetRecordsDamageOperandsAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x2E,
		0x00, 0x80,
		0x00, 0x01,
		0x00, 0x06,
		0x00, 0x01,
		0x00, 0x80,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := DamageRequest{Flags: 0x80, DiceCount: 1, DiceSize: 6, Bonus: 1, SaveFlags: 0x80}
	if len(result.DamageRequests) != 1 || result.DamageRequests[0] != want || len(result.Text) != 1 || result.Text[0] != "HI" {
		t.Fatalf("result=%+v, want damage=%+v and continuation", result, want)
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

func TestRunSubsetEmitsTreasureOperands(t *testing.T) {
	// TREASURE has eight decoded operands. The bounded runner must preserve
	// their raw values before continuing to EXIT.
	payload := []byte{0x27}
	for i := 1; i <= 8; i++ {
		payload = append(payload, 0x00, byte(i))
	}
	payload = append(payload, 0x00)
	result, err := RunSubset(append([]byte{0, 0}, payload...), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Steps != 2 || result.PC != len(payload) {
		t.Fatalf("result=%+v, want TREASURE then EXIT", result)
	}
	if len(result.TreasureRequests) != 1 || result.TreasureRequests[0].Coins[0] != 1 || result.TreasureRequests[0].Coins[6] != 7 || result.TreasureRequests[0].ItemBlock != 8 {
		t.Fatalf("treasure=%+v, want seven coin counts and item block", result.TreasureRequests)
	}
}

func TestRunSubsetEmitsRobRequestAndContinues(t *testing.T) {
	block := []byte{
		0, 0,
		0x28, 0x00, 0x01, 0x00, 0x32, 0x00, 0x00,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := RobRequest{AllParty: true, LossPercent: 50, SelectedPlayerIndex: -1}
	if len(result.RobRequests) != 1 || result.RobRequests[0] != want ||
		len(result.Text) != 1 || result.Text[0] != "HI" || !result.Exited {
		t.Fatalf("result=%+v, want ROB request %+v and continuation", result, want)
	}
}

func TestRunSubsetCombatDispatchesAreaShopService(t *testing.T) {
	block := []byte{
		0, 0,
		0x09, 0x00, 0x01, 0x01, 0x6C, 0x7F,
		0x09, 0x00, 0x10, 0x01, 0x6D, 0x7F,
		0x24,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	runtime := NewRuntimeState(0)
	result, err := runSubsetWithState(block, 0, 20, nil, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ShopRequested || result.ShopPriceScale != 0x10 ||
		result.CombatRequested || result.PC != 13 {
		t.Fatalf("shop dispatch result=%+v", result)
	}
	if runtime.Memory[0x7F6C] != 0 {
		t.Fatalf("EnterShop mirror=%d, want consumed zero", runtime.Memory[0x7F6C])
	}
	result, err = runSubsetWithState(block, 0, 20, nil, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "HI" || !result.Exited {
		t.Fatalf("shop continuation result=%+v", result)
	}
}

func TestRunSubsetCombatDispatchesAreaTempleService(t *testing.T) {
	block := []byte{
		0, 0,
		0x09, 0x00, 0x01, 0x01, 0xE2, 0x7E,
		0x24,
		0x11, 0x80, 0x02, 0x20, 0x92,
		0x00,
	}
	runtime := NewRuntimeState(0)
	result, err := runSubsetWithState(block, 0, 20, nil, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.TempleRequested || result.ShopRequested || result.CombatRequested || result.PC != 7 {
		t.Fatalf("temple dispatch result=%+v", result)
	}
	if runtime.Memory[0x7EE2] != 0 {
		t.Fatalf("EnterTemple mirror=%d, want consumed zero", runtime.Memory[0x7EE2])
	}
	result, err = runSubsetWithState(block, 0, 20, nil, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "HI" || !result.Exited {
		t.Fatalf("temple continuation result=%+v", result)
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

func TestRunSubsetExposesPartyRuleRequestsAndContinues(t *testing.T) {
	block := []byte{0, 0,
		0x1D, 0x01, 0x00, 0x90,
		0x22, 0x01, 0x01, 0x90, 0x01, 0x04, 0x90,
		0x00,
	}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.PartyStrengthRequests) != 1 || result.PartyStrengthRequests[0].Destination != 0x9000 || result.PartyStrengthRequests[0].Resolved {
		t.Fatalf("party strength requests=%#v", result.PartyStrengthRequests)
	}
	if len(result.PartySurpriseRequests) != 1 || result.PartySurpriseRequests[0] != (PartySurpriseRequest{RangerDestination: 0x9001, OtherDestination: 0x9004}) {
		t.Fatalf("party surprise requests=%#v", result.PartySurpriseRequests)
	}
	if result.Steps != 3 || result.PC == 0 {
		t.Fatalf("result=%+v, party commands should continue to EXIT", result)
	}
}

func TestRunSubsetWithPartyContextResolvesPartyCommands(t *testing.T) {
	block := []byte{0, 0,
		0x1D, 0x01, 0x00, 0x90,
		0x22, 0x01, 0x01, 0x90, 0x01, 0x04, 0x90,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 10, nil, 1, PartyContext{Members: []PartyMemberContext{{HitPoints: 20, HasRangerClass: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.PartyStrengthRequests) != 1 || !result.PartyStrengthRequests[0].Resolved || result.PartyStrengthRequests[0].Value != 2 {
		t.Fatalf("resolved strength=%#v", result.PartyStrengthRequests)
	}
	if len(result.PartySurpriseRequests) != 1 || !result.PartySurpriseRequests[0].Resolved || result.PartySurpriseRequests[0].RangerValue != 1 {
		t.Fatalf("resolved surprise=%#v", result.PartySurpriseRequests)
	}
}

func TestRunSubsetWithPartyContextResolvesCheckPartySkills(t *testing.T) {
	block := []byte{0, 0,
		0x1E,
		0x01, 0xA4, 0x80, // 0x80A4 - 0x7FFF = 0xA5: open-locks skill
		0x00, 0x00, // affect ID (unused by skill query)
		0x01, 0x00, 0x90,
		0x01, 0x01, 0x90,
		0x01, 0x02, 0x90,
		0x01, 0x03, 0x90,
		0x00,
	}
	result, err := RunSubsetInteractiveSeedWithPartyContext(block, 0, 10, nil, 1, PartyContext{Members: []PartyMemberContext{
		{ThiefSkills: [8]uint8{10}},
		{ThiefSkills: [8]uint8{20}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CheckPartyRequests) != 1 {
		t.Fatalf("check party requests=%#v", result.CheckPartyRequests)
	}
	request := result.CheckPartyRequests[0]
	if !request.Resolved || request.Minimum != 10 || request.Maximum != 20 || request.Average != 15 {
		t.Fatalf("check party result=%#v", request)
	}
}

func TestRunSubsetExposesWhoSelectionBoundary(t *testing.T) {
	result, err := RunSubset([]byte{0, 0, 0x39, 0x00, 0x00, 0x00}, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.WhoRequests) != 1 || result.WhoRequests[0].Prompt != "" || result.Steps != 2 {
		t.Fatalf("WHO result=%+v", result)
	}
}
