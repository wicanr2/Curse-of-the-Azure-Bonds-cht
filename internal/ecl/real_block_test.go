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
