package ecl

import (
	"archive/zip"
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
