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

func TestRealECL5AreaDepartureFindsAkabarAndDarkElfItems(t *testing.T) {
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
		runtime := NewRuntimeState(0x0014)
		runtime.Started = true
		runtime.Memory[0x4C60] = 1
		runtime.Memory[0x4C5E] = 1
		context := PartyContext{Members: []PartyMemberContext{
			{Name: "HERO", ItemTypes: []uint8{0x5E, 0x60, 0x61}},
			{Name: "AKABAR BEL AKAS", ControlMorale: 0xB2},
		}}
		result, runErr := runSubsetWithStateContext(block.Data, 0x0014, 500, nil, true, 1, runtime, &context)
		if runErr != nil {
			t.Fatal(runErr)
		}
		joined := strings.Join(result.Text, " ")
		if !result.WaitingForMenu ||
			!strings.Contains(joined, "YOUR HELP WAS INVALUABLE") ||
			len(result.DumpRequests) != 1 || !result.DumpRequests[0].Resolved {
			t.Fatalf("Akabar departure result=%+v text=%q", result, joined)
		}
		result, runErr = runSubsetWithStateContext(block.Data, 0x0014, 500, []uint16{0}, true, 1, runtime, &context)
		if runErr != nil {
			t.Fatal(runErr)
		}
		joined = strings.Join(result.Text, " ")
		if !result.WaitingForMenu || !strings.Contains(joined, "DECAY TO USELESSNESS") {
			t.Fatalf("dark-elf item decay result=%+v text=%q", result, joined)
		}
		return
	}
	t.Fatal("ECL5 block 0x30 is absent")
}

func TestRealECL1WorldDispatcherBuildsPostWizardDracolich(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL1.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID != 0x50 {
			continue
		}
		result, runErr := RunSubset(block.Data, 0x149A, 20)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !result.CombatRequested || len(result.MonsterSpawns) != 1 {
			t.Fatalf("post-wizard encounter result=%+v", result)
		}
		spawn := result.MonsterSpawns[0]
		if spawn.MonsterID != 0x3C || spawn.Count != 1 || spawn.IconBlock != 0x3C {
			t.Fatalf("post-wizard spawn=%+v, want MON5 dracolich 0x3C x1", spawn)
		}
		return
	}
	t.Fatal("ECL1 block 0x50 is absent")
}

func TestRealECL4CaveA2CannonContinuesToDeadElfHandler(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL4.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var caveBlock []byte
	for _, block := range blocks {
		if block.Entry.ID == 0x22 {
			caveBlock = block.Data
			break
		}
	}
	if len(caveBlock) == 0 {
		t.Fatal("ECL4 block 0x22 is absent")
	}

	// This is a continuation, not a direct-entry shortcut: terrain A2 first
	// shows three PRESS boundaries before falling through to the dead-elf
	// handler at +061B. Started preserves the supplied raw work memory.
	runtime := NewRuntimeState(0x050A)
	runtime.Started = true
	runtime.Memory[0xC04F] = 0xA2
	runtime.Memory[0x4C03] = 1
	wantPages := [][]string{
		{"A GUT WRENCHING JERK SLAMS YOUR STOMACH TO YOUR", "SPINE.  YOU FEEL AS IF YOU WERE PROPELLED FROM A", "CANNON."},
		{"YOU SEE THE WALLS AROUND BLUR TO A GRAY SMEAR."},
		{"YOU ARE SUDDENLY SLAMMED AGAINST A WALL."},
		{"YOU SEE THE REMAINS OF AN ELF FIGHTER.", "WHAT DO YOU DO?"},
	}
	wantPC := []int{0x14D5, 0x14D5, 0x14D5, 0x0675}
	var result RunResult
	for step, wantText := range wantPages {
		selections := []uint16(nil)
		if step > 0 {
			selections = []uint16{0} // PRESS BUTTON OR RETURN TO CONTINUE.
		}
		result, err = runSubsetWithState(caveBlock, 0x050A, 500, selections, true, 1, runtime)
		if err != nil {
			t.Fatalf("A2 continuation step %d: %v", step, err)
		}
		if !result.WaitingForMenu || result.PC != wantPC[step] || !reflect.DeepEqual(result.Text, wantText) {
			t.Fatalf("A2 continuation step %d result=%+v text=%q, want pc=+%04X text=%q", step, result, result.Text, wantPC[step], wantText)
		}
	}
	if !reflect.DeepEqual(result.CallAddresses, []uint16{0x2E10}) ||
		!reflect.DeepEqual(result.SaveWrites, []MemoryWrite{
			// ★ 執行序 4 是中間那條 `CALL`——`SaveWrites` 與 `CallRequests`
			// 共用同一條計數，所以這裡跳號是對的。
			{Address: 0xC04B, Value: 13, PC: 0x061B, Sequence: 1},
			{Address: 0xC04C, Value: 1, PC: 0x0621, Sequence: 2},
			{Address: 0xC04D, Value: 3, PC: 0x0627, Sequence: 3},
			{Address: 0x4C06, Value: 1, PC: 0x066B, Sequence: 5},
		}) {
		t.Fatalf("dead-elf position transaction calls=%#v writes=%+v", result.CallAddresses, result.SaveWrites)
	}
	for address, want := range map[uint16]uint16{0xC04B: 13, 0xC04C: 1, 0xC04D: 3, 0x4C03: 1, 0x4C06: 1} {
		if got := runtime.Memory[address]; got != want {
			t.Fatalf("dead-elf raw memory[%#x]=%#x, want %#x", address, got, want)
		}
	}
	if len(result.Menus) == 0 || !reflect.DeepEqual(result.Menus[len(result.Menus)-1].Options, []string{"EXAMINE REMAINS", "LEAVE"}) {
		t.Fatalf("dead-elf menu=%+v", result.Menus)
	}

	leave, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{1}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if leave.PC != 0x0E49 || leave.WaitingForMenu || len(leave.Text) != 0 || len(leave.SaveWrites) != 0 {
		t.Fatalf("dead-elf LEAVE result=%+v", leave)
	}
	if runtime.Memory[0x4C03] != 1 || runtime.Memory[0x4C06] != 1 {
		t.Fatalf("dead-elf LEAVE unexpectedly changed raw guard values: 4C03=%#x 4C06=%#x", runtime.Memory[0x4C03], runtime.Memory[0x4C06])
	}
}

func TestRealECL4CaveDeadElfPouchUnlocksJournal59AndRequestsPostCombat(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL4.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var caveBlock []byte
	for _, block := range blocks {
		if block.Entry.ID == 0x22 {
			caveBlock = block.Data
			break
		}
	}
	if len(caveBlock) == 0 {
		t.Fatal("ECL4 block 0x22 is absent")
	}

	runtime := NewRuntimeState(0x050A)
	runtime.Started = true
	runtime.Memory[0xC04F] = 0xA2
	runtime.Memory[0x4C03] = 1
	for step := 0; step < 4; step++ {
		selections := []uint16(nil)
		if step > 0 {
			selections = []uint16{0}
		}
		result, runErr := runSubsetWithState(caveBlock, 0x050A, 500, selections, true, 1, runtime)
		if runErr != nil {
			t.Fatalf("reach dead-elf step %d: %v", step, runErr)
		}
		if !result.WaitingForMenu {
			t.Fatalf("reach dead-elf step %d result=%+v", step, result)
		}
	}

	pouch, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if pouch.PC != 0x14D5 || !pouch.WaitingForMenu || !reflect.DeepEqual(pouch.Text, []string{
		"BURIED BENEATH THE MOLDERING CLOTHING,  YOU FIND A",
		"LEATHER POUCH.",
	}) {
		t.Fatalf("dead-elf pouch=%+v", pouch)
	}

	pouchMenu, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if pouchMenu.PC != 0x06D7 || !pouchMenu.WaitingForMenu || len(pouchMenu.Menus) == 0 ||
		!reflect.DeepEqual(pouchMenu.Menus[len(pouchMenu.Menus)-1].Options, []string{
			"PICK UP POUCH", "POKE AT POUCH", "CAST FIND TRAP", "LEAVE",
		}) {
		t.Fatalf("dead-elf pouch menu=%+v", pouchMenu)
	}

	gasTrap, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if gasTrap.PC != 0x14D5 || !gasTrap.WaitingForMenu ||
		!reflect.DeepEqual(gasTrap.Text, []string{"A GAS TRAP GOES OFF!"}) ||
		// `08A6h` 的 `SAVE 0 7F79` 起頭，接著 `08B0h` 的 `ADD 01 7F79 7F79`
		// 把整隊走一遍——算術也是寫入路徑，所以八次遞增都在裡面。
		!reflect.DeepEqual(gasTrap.SaveWrites, append([]MemoryWrite{
			{Address: 0x7F79, Value: 0, PC: 0x08A6, Sequence: 1},
		}, incrementingWrites(0x7F79, 0x08B0, 1, 8)...)) {
		t.Fatalf("dead-elf gas trap=%+v", gasTrap)
	}

	mapResult, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if mapResult.PC != 0x14D5 || !mapResult.WaitingForMenu ||
		!reflect.DeepEqual(mapResult.Text, []string{
			"YOU DISCOVER A MAP.  ON IT, YOU SEE DEXAMS ALTAR",
			"INDICATED AND A PATH THAT SEEMS TO LEAD OUTSIDE.",
			"YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59.",
		}) ||
		!reflect.DeepEqual(mapResult.DamageRequests, []DamageRequest{{Flags: 0xC0, DiceCount: 1, DiceSize: 12, SaveFlags: 3}}) ||
		!reflect.DeepEqual(mapResult.SaveWrites, []MemoryWrite{{Address: 0x4C07, Value: 0x80, PC: 0x09AC, Sequence: 1}}) {
		t.Fatalf("dead-elf Journal 59 map=%+v", mapResult)
	}
	if runtime.Memory[0x4C03] != 1 || runtime.Memory[0x4C07] != 0x80 {
		t.Fatalf("dead-elf map raw guards: 4C03=%#x 4C07=%#x", runtime.Memory[0x4C03], runtime.Memory[0x4C07])
	}

	combatResult, err := runSubsetWithState(caveBlock, 0x050A, 500, []uint16{0}, true, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	// ⚠ 這一處沒有擺怪，所以走的是 `24h` 的**第四支**：戰利品分配
	// （`overlay-05` ＝ POSTCOM），不是戰鬥（spec 1182）。
	if combatResult.PC != 0x089C || !combatResult.PostCombatRequested ||
		combatResult.CombatRequested || combatResult.WaitingForMenu ||
		len(combatResult.Text) != 0 || len(combatResult.SaveWrites) != 0 {
		t.Fatalf("dead-elf post-map combat=%+v", combatResult)
	}
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

// incrementingWrites 展開「同一條 `ADD 1` 在迴圈裡跑 count 次」的寫入序列。
func incrementingWrites(address uint16, pc int, firstValue uint16, count int) []MemoryWrite {
	writes := make([]MemoryWrite, 0, count)
	for index := 0; index < count; index++ {
		writes = append(writes, MemoryWrite{
			Address:  address,
			Value:    firstValue + uint16(index),
			PC:       pc,
			Sequence: index + 2,
		})
	}
	return writes
}
