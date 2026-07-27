package ecl

import (
	"archive/zip"
	"io"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

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
	result, err := RunSubset(block.Data, 0x14, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NPCIDs) != 1 || result.NPCIDs[0] != 0x55 || result.Steps != 5 || result.PC != 0x23 {
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
