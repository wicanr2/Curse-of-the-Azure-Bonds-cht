package gfx

import (
	"archive/zip"
	"io"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestParsePieceSetFromOriginalArea2Image(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("original image is not available: %v", err)
		}
		t.Fatal(err)
	}
	defer archive.Close()
	readMember := func(name string) []byte {
		for _, file := range archive.File {
			if file.Name != name {
				continue
			}
			reader, openErr := file.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			defer reader.Close()
			data, readErr := io.ReadAll(reader)
			if readErr != nil {
				t.Fatal(readErr)
			}
			return data
		}
		t.Fatalf("archive member %s is missing", name)
		return nil
	}
	wallBlocks, err := dax.Parse(readMember("WALLDEF2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	symbolBlocks, err := dax.Parse(readMember("8X8D2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for setID, selector := range map[uint8]uint8{1: 1, 2: 2, 3: 3} {
		set, parseErr := ParsePieceSet(setID, selector, wallBlocks, symbolBlocks)
		if parseErr != nil {
			t.Fatalf("piece set %d: %v", setID, parseErr)
		}
		if len(set.WallDefs) != 1 || len(set.SymbolBlockIDs) != 1 || set.SymbolBlockIDs[0] != selector {
			t.Fatalf("piece set %d metadata = walls %d symbols %v", setID, len(set.WallDefs), set.SymbolBlockIDs)
		}
	}
}

func testSymbolPicture() []byte {
	data := make([]byte, 17+32)
	data[0] = 8 // height
	data[2] = 1 // width units: 8 pixels
	data[8] = 1 // one symbol item
	for index := 17; index < len(data); index++ {
		data[index] = 0x12
	}
	return data
}

func testWallData(records int) []byte {
	data := make([]byte, records*5*156)
	for record := 0; record < records; record++ {
		data[record*5*156] = byte(record + 1)
	}
	return data
}

func TestParsePieceSetMapsSingleWallRecordToSelector(t *testing.T) {
	set, err := ParsePieceSet(1, 7,
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: testWallData(1)}},
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: testSymbolPicture()}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if set.SetID != 1 || set.Selector != 7 || len(set.WallDefs) != 1 || set.SymbolBlockIDs[0] != 7 {
		t.Fatalf("unexpected piece set metadata: %+v", set)
	}
	if got := set.Symbols[7].ItemCount; got != 1 {
		t.Fatalf("symbol item count = %d, want 1", got)
	}
}

func TestParsePieceSetUsesReferenceMultiRecordSymbolIDs(t *testing.T) {
	set, err := ParsePieceSet(2, 7,
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: testWallData(2)}},
		[]dax.Block{
			{Entry: dax.Entry{ID: 71}, Data: testSymbolPicture()},
			{Entry: dax.Entry{ID: 72}, Data: testSymbolPicture()},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := set.SymbolBlockIDs, []uint8{71, 72}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("symbol block IDs = %v, want %v", got, want)
	}
}

func TestMergePicturesUsesTransparentAndORLayers(t *testing.T) {
	destination := Picture{WidthUnits: 1, HeightUnits: 2, ItemCount: 1, Pixels: []uint8{1, 16, 4, 16, 16, 2, 8, 16}}
	source := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{16, 2, 8, 16, 16, 16, 16, 16}}
	merged, err := MergePictures(destination, source)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint8{1, 2, 12, 16, 16, 2, 8, 16, 16, 2, 8, 16, 16, 16, 16, 16}
	for i, got := range merged.Pixels {
		if got != want[i] {
			t.Fatalf("pixel %d = %d, want %d", i, got, want[i])
		}
	}
}

func TestMergePicturesRejectsWiderSource(t *testing.T) {
	_, err := MergePictures(Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{16, 16, 16, 16, 16, 16, 16, 16}}, Picture{WidthUnits: 2, HeightUnits: 1, ItemCount: 1, Pixels: make([]uint8, 16)})
	if err == nil {
		t.Fatal("expected dimension error")
	}
}

func TestMergePicturesAtAppliesPixelOffset(t *testing.T) {
	destination := Picture{WidthUnits: 1, HeightUnits: 2, ItemCount: 1, Pixels: make([]uint8, 16)}
	for index := range destination.Pixels {
		destination.Pixels[index] = 16
	}
	source := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{1, 2, 3, 4, 5, 6, 7, 8}}
	merged, err := MergePicturesAt(destination, source, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range source.Pixels {
		if got := merged.Pixels[8+index]; got != want {
			t.Fatalf("offset pixel %d=%d, want %d", index, got, want)
		}
	}
}

func TestFlipHorizontalPreservesIndexedPixels(t *testing.T) {
	picture := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{1, 2, 3, 4, 5, 6, 7, 8}}
	flipped := picture.FlipHorizontal()
	want := []uint8{8, 7, 6, 5, 4, 3, 2, 1}
	for i, got := range flipped.Pixels {
		if got != want[i] {
			t.Fatalf("pixel %d = %d, want %d", i, got, want[i])
		}
	}
}
