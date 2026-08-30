package ecl

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

// TestRealECLCorpusHasNoUnknownReachableCommands is the corpus gate for the
// parser/control-flow layer. It walks all five lifecycle entries of every
// original DOS ECL block and requires every statically reachable instruction
// to have a command-table entry. This does not promote a renderer or external
// routine side effect to "complete"; those remain separate adapter tests.
func TestRealECLCorpusHasNoUnknownReachableCommands(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	blocksSeen, entriesSeen, instructionsSeen := 0, 0, 0
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, err := dax.Parse(realZipMember(t, archive, member))
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			points, _, err := EntryPoints(block.Data, 5)
			if err != nil {
				t.Fatalf("%s block 0x%02X entry points: %v", member, block.Entry.ID, err)
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-CodeAddressBase)
			}
			graph, err := TraceGraph(block.Data, starts, len(block.Data)*8)
			if err != nil {
				t.Fatalf("%s block 0x%02X graph: %v", member, block.Entry.ID, err)
			}
			blocksSeen++
			entriesSeen += len(points)
			instructionsSeen += len(graph.Instructions)
			for _, instruction := range graph.Instructions {
				if _, ok := KnownCommands[instruction.Command.Opcode]; !ok {
					t.Fatalf("%s block 0x%02X reached unknown opcode 0x%02X at +0x%04X",
						member, block.Entry.ID, instruction.Command.Opcode, instruction.Offset)
				}
			}
		}
	}
	if blocksSeen != 25 || entriesSeen != 125 || instructionsSeen == 0 {
		t.Fatalf("ECL corpus blocks=%d entries=%d instructions=%d, want 25 blocks, 125 entries and nonzero instructions",
			blocksSeen, entriesSeen, instructionsSeen)
	}
}

func TestKnownECLCommandTableCoversEveryCoABOpcodeByte(t *testing.T) {
	for opcode := byte(0); opcode <= 0x40; opcode++ {
		if _, ok := KnownCommands[opcode]; !ok {
			t.Fatalf("missing ECL command metadata for opcode 0x%02X", opcode)
		}
	}
}

func TestRealECLCorpusDoesNotReachUnusedUnsupportedCommands(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()
	unused := map[byte]string{
		0x0F: "INPUT NUMBER",
		0x1F: "UNKNOWN_1F",
		0x23: "SURPRISE",
	}
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, err := dax.Parse(realZipMember(t, archive, member))
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			points, _, err := EntryPoints(block.Data, 5)
			if err != nil {
				t.Fatalf("%s block 0x%02X entry points: %v", member, block.Entry.ID, err)
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-CodeAddressBase)
			}
			graph, err := TraceGraph(block.Data, starts, len(block.Data)*8)
			if err != nil {
				t.Fatalf("%s block 0x%02X graph: %v", member, block.Entry.ID, err)
			}
			for _, instruction := range graph.Instructions {
				if name, found := unused[instruction.Command.Opcode]; found {
					t.Fatalf("%s block 0x%02X reaches CoAB-unused %s opcode 0x%02X at +0x%04X",
						member, block.Entry.ID, name, instruction.Command.Opcode, instruction.Offset)
				}
			}
		}
	}
}
