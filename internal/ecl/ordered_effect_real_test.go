package ecl

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealTreasureCombatCandidatesPreservePendingRewardAndResumePC(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()

	tests := []struct {
		name           string
		member         string
		blockID        uint8
		start          int
		combatPC       int
		resumePC       int
		memory         map[uint16]uint16
		wantCoins      [7]uint16
		wantItemBlock  uint16
		wantSpawnID    uint8
		wantSpawnCount uint8
	}{
		{
			name: "ECL3 block 15 dynamic camp reward", member: "ECL3.DAX", blockID: 0x15,
			start: 0x0536, combatPC: 0x056E, resumePC: 0x0578,
			memory:    map[uint16]uint16{0x7F7A: 0x21, 0x7F7B: 2, 0x7F7D: 0x22, 0x7F7E: 0, 0x7F79: 1, 0x4C07: 1, 0x7F81: 3, 0x7F80: 0xFF},
			wantCoins: [7]uint16{0, 0, 0, 0, 0, 4, 3}, wantItemBlock: 0xFF,
			wantSpawnID: 0x21, wantSpawnCount: 2,
		},
		{
			name: "ECL4 block 25 fixed platinum reward", member: "ECL4.DAX", blockID: 0x25,
			start: 0x1288, combatPC: 0x12A3, resumePC: 0x1534,
			memory:    map[uint16]uint16{0x7F79: 3},
			wantCoins: [7]uint16{0, 0, 0, 0, 30, 0, 0}, wantItemBlock: 0xFF,
			wantSpawnID: 0x23, wantSpawnCount: 3,
		},
		{
			name: "ECL6 block 45 dynamic camp reward", member: "ECL6.DAX", blockID: 0x45,
			start: 0x0522, combatPC: 0x056B, resumePC: 0x0575,
			memory:    map[uint16]uint16{0x7F7A: 0x41, 0x7F7B: 2, 0x7F7D: 0x42, 0x7F7E: 0, 0x7F79: 1, 0x4C07: 1, 0x7F81: 2, 0x7F80: 0xFF},
			wantCoins: [7]uint16{0, 0, 0, 0, 0, 4, 2}, wantItemBlock: 0xFF,
			wantSpawnID: 0x41, wantSpawnCount: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blocks, parseErr := dax.Parse(realZipMember(t, archive, test.member))
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			var data []byte
			for _, block := range blocks {
				if block.Entry.ID == test.blockID {
					data = block.Data
					break
				}
			}
			if data == nil {
				t.Fatalf("%s block %#x is absent", test.member, test.blockID)
			}
			session, sessionErr := NewBlockSession(map[uint8][]byte{test.blockID: data}, test.blockID)
			if sessionErr != nil {
				t.Fatal(sessionErr)
			}
			for address, value := range test.memory {
				session.SetMemoryValue(address, value)
			}

			result, runErr := session.RunFrom(test.start, 100, nil)
			if runErr != nil {
				t.Fatal(runErr)
			}
			if !result.CombatRequested || result.PC != test.combatPC {
				t.Fatalf("combat=%v pc=%#x, want true/%#x", result.CombatRequested, result.PC, test.combatPC)
			}
			if len(result.TreasureRequests) != 1 || result.TreasureRequests[0].Coins != test.wantCoins || result.TreasureRequests[0].ItemBlock != test.wantItemBlock {
				t.Fatalf("treasure=%+v, want coins=%v item=%#x", result.TreasureRequests, test.wantCoins, test.wantItemBlock)
			}
			if len(result.MonsterSpawns) == 0 || result.MonsterSpawns[0].MonsterID != test.wantSpawnID || result.MonsterSpawns[0].Count != test.wantSpawnCount {
				t.Fatalf("spawns=%+v, want first id=%#x count=%d", result.MonsterSpawns, test.wantSpawnID, test.wantSpawnCount)
			}

			resumed, resumeErr := session.RunFrom(result.PC, 100, nil)
			if resumeErr != nil {
				t.Fatal(resumeErr)
			}
			if test.blockID == 0x25 {
				if resumed.PictureRequested || len(resumed.CallAddresses) != 1 || resumed.CallAddresses[0] != 0x2E10 || !resumed.Exited || resumed.PC != test.resumePC {
					t.Fatalf("resumed=%+v, want clear-picture, CALL 2E10, EXIT and pc %#x", resumed, test.resumePC)
				}
			} else {
				if !resumed.Exited || resumed.PC != test.resumePC {
					t.Fatalf("resumed exited=%v pc=%#x, want true/%#x", resumed.Exited, resumed.PC, test.resumePC)
				}
			}
		})
	}
}

// realBlockData returns one decoded block of one DAX member, or skips when the
// original image is unavailable.
func realBlockData(t *testing.T, member string, blockID uint8) []byte {
	t.Helper()
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, parseErr := dax.Parse(realZipMember(t, archive, member))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	for _, block := range blocks {
		if block.Entry.ID == blockID {
			return block.Data
		}
	}
	t.Fatalf("%s block %#x is absent", member, blockID)
	return nil
}

// NEWECL is a terminator, not a fallthrough opcode: the original handler loads
// the replacement block, resets the interpreter PC to the code base and raises
// both stop flags, so the byte after it is never executed (spec 1104).
//
// ECL4 block 0x25 is the case that proves it matters. Its per-turn tail is
// `DESTROY ITEMS, DESTROY ITEMS, SAVE, NEWECL`; the two SAVE instructions that
// follow in the payload are the pre_camp lifecycle entry, whose offset is
// exactly the byte after NEWECL. Treating NEWECL as fallthrough merged two
// unrelated runs into one ordered-effect candidate.
func TestRealNewEclStopsBeforeTheFollowingLifecycleEntry(t *testing.T) {
	data := realBlockData(t, "ECL4.DAX", 0x25)

	result, err := RunSubset(data, 0x021F, 100)
	if err != nil {
		t.Fatal(err)
	}
	if result.NewECLBlockID == nil || *result.NewECLBlockID != 0x50 {
		t.Fatalf("NewECLBlockID=%v, want 0x50", result.NewECLBlockID)
	}
	if result.PC != 0x022E {
		t.Fatalf("stopped at %#x, want %#x", result.PC, 0x022E)
	}
	// DESTROY ITEMS, DESTROY ITEMS, SAVE, NEWECL. A fifth step would mean the
	// pre_camp entry at 0x022E ran as part of this transaction.
	if result.Steps != 4 {
		t.Fatalf("executed %d instructions, want the 4 that precede and include NEWECL",
			result.Steps)
	}
	if len(result.DestroyItemIDs) != 2 || result.DestroyItemIDs[0] != 0x61 || result.DestroyItemIDs[1] != 0x60 {
		t.Fatalf("destroy-item ids=%v, want [0x61 0x60]", result.DestroyItemIDs)
	}
}

// The `PICTURE -> CALL 2E10h -> SETUP MONSTER` shape is the P0-C candidate. In
// the original, PICTURE only marks the screen dirty and CALL 2E10h is the one
// point that flushes the five dirty flags, so the picture change is committed
// before the combat setup that follows it (spec 1104 §四). Pin that order with
// two bounded runs rather than one aggregate result, because an aggregate
// cannot distinguish "before" from "after".
func TestRealPictureCommitCallPrecedesCombatSetup(t *testing.T) {
	data := realBlockData(t, "ECL2.DAX", 0x02)

	// PRINT, PRINT, SAVE, PICTURE 0xFF.
	before, err := RunSubset(data, 0x02CB, 4)
	if err == nil {
		t.Fatal("a four-step budget must stop before the commit call")
	}
	if len(before.CallAddresses) != 0 {
		t.Fatalf("call addresses=%v before the commit call, want none", before.CallAddresses)
	}
	if before.MonsterSetup != nil {
		t.Fatalf("monster setup=%+v before the commit call, want none", before.MonsterSetup)
	}
	// PICTURE 0xFF is the close-window branch, not a picture load.
	if before.PictureRequested {
		t.Fatalf("PICTURE 0xFF must not request a picture block, got %#x", before.PictureBlock)
	}

	// ... CALL 2E10h, SAVE, SETUP MONSTER.
	after, err := RunSubset(data, 0x02CB, 7)
	if err == nil {
		t.Fatal("a seven-step budget must stop at the GOSUB that follows")
	}
	if len(after.CallAddresses) != 1 || after.CallAddresses[0] != 0x2E10 {
		t.Fatalf("call addresses=%v, want exactly [0x2E10]", after.CallAddresses)
	}
	if len(after.CallRequests) != 1 || after.CallRequests[0].PC != 0x030A {
		t.Fatalf("call requests=%+v, want one at %#x", after.CallRequests, 0x030A)
	}
	if after.MonsterSetup == nil {
		t.Fatal("SETUP MONSTER after the commit call produced no setup")
	}
}

// CLEARMONSTERS frees the monster chain; it does not undo SETUP MONSTER.
// Four corpus sites run SETUP MONSTER first and then CLEARMONSTERS before the
// fight, so dropping the setup there costs the enemy sprite (spec 1104 §四).
func TestRealClearMonstersKeepsEarlierMonsterSetup(t *testing.T) {
	tests := []struct {
		member  string
		blockID uint8
		start   int
	}{
		{"ECL3.DAX", 0x11, 0x114E},
		{"ECL3.DAX", 0x12, 0x06BF},
		{"ECL4.DAX", 0x21, 0x05B5},
		{"ECL5.DAX", 0x32, 0x0774},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s/%#x", test.member, test.blockID), func(t *testing.T) {
			data := realBlockData(t, test.member, test.blockID)
			result, err := RunSubset(data, test.start, 200)
			if err != nil {
				t.Fatal(err)
			}
			if !result.CombatRequested {
				t.Fatalf("run from %#x did not reach the combat boundary", test.start)
			}
			if result.MonsterSetup == nil {
				t.Fatal("CLEARMONSTERS discarded the SETUP MONSTER that preceded it")
			}
			if len(result.MonsterSpawns) == 0 {
				t.Fatal("no monster spawned after CLEARMONSTERS")
			}
		})
	}
}
