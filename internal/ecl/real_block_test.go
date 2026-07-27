package ecl

import (
	"archive/zip"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealAllInitializationEntriesReachSupportedBoundary(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocksSeen := 0
	entriesSeen := 0
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, parseErr := dax.Parse(realZipMember(t, archive, member))
		if parseErr != nil {
			t.Fatalf("%s: %v", member, parseErr)
		}
		blocksSeen += len(blocks)
		for _, block := range blocks {
			reports, smokeErr := SmokeInitializationEntries(block.Data, 5, 500, nil)
			if smokeErr != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, smokeErr)
			}
			entriesSeen += len(reports)
			for _, report := range reports {
				if report.Err != nil {
					t.Fatalf("%s block 0x%02X entry %d stopped at +0x%04X after %d steps: %v", member, block.Entry.ID, report.Index, report.Result.PC, report.Result.Steps, report.Err)
				}
			}
		}
	}
	if blocksSeen != 25 || entriesSeen != 125 {
		t.Fatalf("smoke coverage blocks=%d entries=%d, want 25 and 125", blocksSeen, entriesSeen)
	}
}

func TestRealECL5SunlightFindItemUsesPartyContext(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL5.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID != 0x30 {
			continue
		}
		result, runErr := RunSubsetInteractiveSeedWithPartyContext(block.Data, 0x14, 500, []uint16{0}, 1, PartyContext{
			Members: []PartyMemberContext{{Name: "HERO", ItemTypes: []uint8{0x5E}}},
		})
		if runErr != nil {
			t.Fatal(runErr)
		}
		if len(result.FindItemRequests) == 0 || !result.FindItemRequests[0].Resolved || !result.FindItemRequests[0].Found {
			t.Fatalf("find requests=%#v", result.FindItemRequests)
		}
		if !strings.Contains(strings.Join(result.Text, " "), "SUNLIGHT") {
			t.Fatalf("text=%q, want sunlight decay event", result.Text)
		}
		instruction, decodeErr := decodeInstruction(block.Data[2:], 0x20E)
		if decodeErr != nil || instruction.Command.Opcode != 0x3E {
			t.Fatalf("ECL5 Akabar leave instruction=%+v err=%v, want DUMP", instruction, decodeErr)
		}
		return
	}
	t.Fatal("ECL5 block 0x30 is absent")
}

func TestRealECL5WizardTowerDragonParlayMappings(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL5.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID != 0x33 {
			continue
		}
		for _, selection := range []uint16{0, 2} {
			result, runErr := RunSubsetInteractiveSeed(block.Data, 0x0452, 500, []uint16{selection, 0}, 1)
			if runErr != nil || !result.CombatRequested || len(result.MonsterSpawns) != 1 ||
				result.MonsterSpawns[0].MonsterID != 0x35 || result.MonsterSpawns[0].Count != 14 {
				t.Fatalf("wizard-tower outer selection %d result=%+v err=%v, want fourteen black dragons",
					selection, result, runErr)
			}
		}
		noEligibility := NewRuntimeState(0x0676)
		noEligibility.Memory[0x4C61] = 0
		skippedHeart, skippedErr := runSubsetWithState(block.Data, 0x0676, 500, nil, true, 1, noEligibility)
		if skippedErr != nil || len(skippedHeart.DamageRequests) != 0 ||
			strings.Contains(strings.Join(skippedHeart.Text, " "), "TAKE ONE OF THEIR HEARTS") ||
			!strings.Contains(strings.Join(skippedHeart.Text, " "), "REST SAFELY") {
			t.Fatalf("ineligible dragon-heart continuation=%+v err=%v", skippedHeart, skippedErr)
		}
		declineRuntime := NewRuntimeState(0x0676)
		declineRuntime.Memory[0x4C61] = 1
		declinedHeart, declinedErr := runSubsetWithState(block.Data, 0x0676, 500, []uint16{1}, true, 1, declineRuntime)
		if declinedErr != nil || len(declinedHeart.DamageRequests) != 0 ||
			declineRuntime.Memory[0x4C64] != 0 ||
			!strings.Contains(strings.Join(declinedHeart.Text, " "), "REST SAFELY") {
			t.Fatalf("declined dragon-heart continuation=%+v memory=%#v err=%v",
				declinedHeart, declineRuntime.Memory, declinedErr)
		}
		for tactic, want := range []uint16{1, 0, 0, 0, 1} {
			runtime := NewRuntimeState(0x05EA)
			result, runErr := runSubsetWithState(block.Data, 0x05EA, 500, []uint16{uint16(tactic)}, true, 1, runtime)
			if runErr != nil {
				t.Fatalf("tactic %d: %v", tactic, runErr)
			}
			if got := runtime.Memory[0x7F79]; got != want {
				t.Fatalf("tactic %d mapping=%d, want %d result=%+v", tactic, got, want, result)
			}
			if want == 1 {
				hostile, hostileErr := RunSubsetInteractiveSeed(block.Data, 0x05EA, 500, []uint16{uint16(tactic), 0}, 1)
				if hostileErr != nil || !hostile.CombatRequested || len(hostile.MonsterSpawns) != 1 ||
					hostile.MonsterSpawns[0].MonsterID != 0x35 || hostile.MonsterSpawns[0].Count != 14 {
					t.Fatalf("hostile tactic %d result=%+v err=%v, want fourteen black dragons",
						tactic, hostile, hostileErr)
				}
			}
			if want == 0 && !strings.Contains(strings.Join(result.Text, " "), "YOU HAVE CONVINCED US") {
				t.Fatalf("successful tactic %d text=%q", tactic, result.Text)
			}
		}
		return
	}
	t.Fatal("ECL5 block 0x33 is absent")
}

func TestRealECL5WizardTowerExitBranches(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL5.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID != 0x33 {
			continue
		}
		for name, test := range map[string]struct {
			selections []uint16
			blockID    uint8
			exited     bool
		}{
			"caves":   {selections: []uint16{0}, blockID: 0x32},
			"village": {selections: []uint16{1, 0}, blockID: 0x31},
			"depart":  {selections: []uint16{1, 1}, blockID: 0x30},
			"stay":    {selections: []uint16{2}, exited: true},
		} {
			runtime := NewRuntimeState(0x07EA)
			result, runErr := runSubsetWithState(block.Data, 0x07EA, 500, test.selections, true, 1, runtime)
			if runErr != nil {
				t.Fatalf("%s: %v", name, runErr)
			}
			if result.Exited != test.exited {
				t.Fatalf("%s exited=%v, want %v result=%+v", name, result.Exited, test.exited, result)
			}
			if test.blockID != 0 && (result.NewECLBlockID == nil || *result.NewECLBlockID != test.blockID) {
				t.Fatalf("%s block=%v, want %#x result=%+v", name, result.NewECLBlockID, test.blockID, result)
			}
			if name == "caves" && (runtime.Memory[0xC04B] != 6 ||
				runtime.Memory[0xC04C] != 15 || runtime.Memory[0xC04D] != 0) {
				t.Fatalf("caves position=%d,%d,%d, want 6,15,0",
					runtime.Memory[0xC04B], runtime.Memory[0xC04C], runtime.Memory[0xC04D])
			}
			if name == "stay" && len(result.Menus) != 1 {
				t.Fatalf("stay menus=%#v, want one roof menu", result.Menus)
			}
			if (name == "village" || name == "depart") && len(result.Menus) != 2 {
				t.Fatalf("%s menus=%#v, want roof and destination menus", name, result.Menus)
			}
		}
		return
	}
	t.Fatal("ECL5 block 0x33 is absent")
}

func realZipMember(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, entry := range archive.File {
		if entry.Name != name {
			continue
		}
		reader, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	t.Fatalf("%s is absent from original image", name)
	return nil
}

func TestRealECL1Block82AddNPCReachesExit(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	var eclData []byte
	for _, entry := range archive.File {
		if entry.Name != "ECL1.DAX" {
			continue
		}
		reader, openErr := entry.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		eclData, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if len(eclData) == 0 {
		t.Fatal("ECL1.DAX is absent from original image")
	}
	blocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	var block dax.Block
	found := false
	for _, candidate := range blocks {
		if candidate.Entry.ID == 0x52 {
			block, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("ECL1 block 0x52 is absent")
	}
	result, err := RunSubset(block.Data, 0x14, 100)
	if err != nil {
		t.Fatal(err)
	}
	wantNPCs := []NPCRequest{{ID: 0x55, Morale: 100}, {ID: 0x58, Morale: 100}, {ID: 0x5A, Morale: 100}}
	if !reflect.DeepEqual(result.NPCIDs, []uint16{0x55, 0x58, 0x5A}) ||
		!reflect.DeepEqual(result.NPCRequests, wantNPCs) ||
		!result.CombatRequested || result.DelayCount != 12 ||
		!reflect.DeepEqual(result.CallAddresses, []uint16{0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803, 0x6803}) ||
		result.Steps != 53 || result.PC != 0x236 {
		t.Fatalf("result=%+v", result)
	}
}

func TestRealECL2Block1LoadPiecesPrefix(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	var eclData []byte
	for _, entry := range archive.File {
		if entry.Name != "ECL2.DAX" {
			continue
		}
		reader, openErr := entry.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		eclData, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if len(eclData) == 0 {
		t.Fatal("ECL2.DAX is absent from original image")
	}
	blocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	var block dax.Block
	found := false
	for _, candidate := range blocks {
		if candidate.Entry.ID == 0x01 {
			block, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("ECL2 block 0x01 is absent")
	}
	result, err := RunSubset(block.Data, 0x14, 2)
	if err == nil || !result.LoadPiecesRequested || result.LoadPieces != [3]uint16{1, 2, 3} {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestRealECL1ToECL2NEWECLSwitch(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocksByID := make(map[uint8][]byte)
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX"} {
		var data []byte
		for _, entry := range archive.File {
			if entry.Name != member {
				continue
			}
			reader, openErr := entry.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			data, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		if len(data) == 0 {
			t.Fatalf("%s is absent from original image", member)
		}
		parsed, parseErr := dax.Parse(data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range parsed {
			if member == "ECL1.DAX" && block.Entry.ID != 0x50 {
				continue
			}
			if member == "ECL2.DAX" && block.Entry.ID != 3 {
				continue
			}
			blocksByID[block.Entry.ID] = block.Data
		}
	}
	session, err := NewBlockSession(blocksByID, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := session.RunFrom(0x5B5, 128, nil)
	if session.CurrentBlockID() != 3 {
		t.Fatalf("current block=0x%02X result=%+v err=%v, want ECL2 block 3", session.CurrentBlockID(), result, runErr)
	}
	// ECL2's target entry may stop at a later unsupported routine; the
	// transition itself must still be applied and remain observable.
	if runErr != nil && result.Steps == 0 {
		t.Fatalf("target transition produced no bounded result: result=%+v err=%v", result, runErr)
	}
}

func TestRealECL2Block3DamageOperandOrder(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	var eclData []byte
	for _, entry := range archive.File {
		if entry.Name != "ECL2.DAX" {
			continue
		}
		reader, openErr := entry.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		eclData, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if len(eclData) == 0 {
		t.Fatal("ECL2.DAX is absent from original image")
	}
	blocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID != 3 {
			continue
		}
		instruction, decodeErr := decodeInstruction(block.Data[2:], 0x1599)
		if decodeErr != nil {
			t.Fatal(decodeErr)
		}
		if instruction.Command.Opcode != 0x2E || len(instruction.Operands) != 5 {
			t.Fatalf("instruction=%+v, want DAMAGE with five operands", instruction)
		}
		want := []uint16{0xA0, 1, 6, 1, 0x80}
		for index, operand := range instruction.Operands {
			value, valueErr := operandValue(operand, nil)
			if valueErr != nil {
				t.Fatalf("operand %d: %v", index, valueErr)
			}
			if value != want[index] {
				t.Fatalf("operand %d=0x%X, want 0x%X", index, value, want[index])
			}
		}
		return
	}
	t.Fatal("ECL2 block 3 is absent")
}
