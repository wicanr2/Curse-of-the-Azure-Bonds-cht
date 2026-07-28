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
	globalSymbolBlocks, err := dax.Parse(readMember("8X8D1.DAX"))
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
	foundArea := false
	for _, block := range globalSymbolBlocks {
		if block.Entry.ID != 0xCA {
			continue
		}
		foundArea = true
		picture, parseErr := ParsePicture(block.Data, true, 13)
		if parseErr != nil {
			t.Fatalf("AREA symbol block: %v", parseErr)
		}
		if picture.Width() != 8 || picture.Height() != 8 || picture.ItemCount != 40 {
			t.Fatalf("AREA symbols = %dx%d items=%d", picture.Width(), picture.Height(), picture.ItemCount)
		}
	}
	if !foundArea {
		t.Fatal("8X8D1 block 0xCA is missing")
	}
	skyBlocks, err := dax.Parse(readMember("SKY.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	expectedSky := map[uint8][2]int{
		250: {88, 16},
		251: {24, 24},
		252: {88, 48},
	}
	if len(skyBlocks) != len(expectedSky) {
		t.Fatalf("SKY blocks=%d, want %d", len(skyBlocks), len(expectedSky))
	}
	for _, block := range skyBlocks {
		picture, parseErr := ParsePicture(block.Data, true, 13)
		if parseErr != nil {
			t.Fatalf("SKY block 0x%02X: %v", block.Entry.ID, parseErr)
		}
		dimensions, ok := expectedSky[block.Entry.ID]
		if !ok || picture.Width() != dimensions[0] || picture.Height() != dimensions[1] || picture.ItemCount != 1 {
			t.Fatalf("SKY block 0x%02X = %dx%d items=%d", block.Entry.ID, picture.Width(), picture.Height(), picture.ItemCount)
		}
	}
}

func testSymbolPicture() []byte {
	return testSymbolPictureItems(1)
}

func testSymbolPictureItems(items int) []byte {
	data := make([]byte, 17+items*32)
	data[0] = 8 // height
	data[2] = 1 // width units: 8 pixels
	data[8] = byte(items)
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
	wallData := testWallData(2)
	wallData[0] = 0x2E
	wallData[5*156] = 0x2F
	set, err := ParsePieceSet(2, 7,
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: wallData}},
		[]dax.Block{
			{Entry: dax.Entry{ID: 71}, Data: testSymbolPictureItems(70)},
			{Entry: dax.Entry{ID: 72}, Data: testSymbolPictureItems(70)},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := set.SymbolBlockIDs, []uint8{71, 72}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("symbol block IDs = %v, want %v", got, want)
	}
	if got, want := set.WallDefs[0].Rows[0][0], uint8(0x74); got != want {
		t.Fatalf("first wall symbol = 0x%02X, want 0x%02X", got, want)
	}
	if got, want := set.WallDefs[1].Rows[0][0], uint8(0xBB); got != want {
		t.Fatalf("second wall symbol = 0x%02X, want 0x%02X", got, want)
	}
	if _, id, ok := set.WallSymbol(0, 0, 0); !ok || id != 0x74 {
		t.Fatalf("WallSymbol first record = id 0x%02X, ok %v", id, ok)
	}
}

func TestBuildWallLayoutUsesReferenceWindowShape(t *testing.T) {
	wallData := testWallData(1)
	wallData[0] = 0x2E
	set, err := ParsePieceSet(1, 7,
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: wallData}},
		[]dax.Block{{Entry: dax.Entry{ID: 7}, Data: testSymbolPictureItems(70)}},
	)
	if err != nil {
		t.Fatal(err)
	}
	stamps, err := BuildWallLayout(set, 1, 0, 4, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(stamps) != 1 || stamps[0].Row != 4 || stamps[0].Column != 5 || stamps[0].Item != 0 {
		t.Fatalf("stamps = %+v, want one front-wall stamp at (4,5), item 0", stamps)
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

func TestComposeHeadBodyGrowsCanvasAndOffsetsBody(t *testing.T) {
	head := Picture{WidthUnits: 1, HeightUnits: 2, ItemCount: 1, Pixels: filledPicturePixels(16, 1)}
	body := Picture{WidthUnits: 1, HeightUnits: 3, ItemCount: 1, Pixels: filledPicturePixels(24, 2)}
	composed, err := ComposeHeadBody(head, body, 1)
	if err != nil {
		t.Fatal(err)
	}
	if composed.Width() != 8 || composed.Height() != 4 {
		t.Fatalf("composed dimensions=%dx%d, want 8x4", composed.Width(), composed.Height())
	}
	if value, _ := composed.Pixel(0, 0, 0); value != 1 {
		t.Fatalf("HEAD pixel=%d, want 1 at origin", value)
	}
	if value, _ := composed.Pixel(0, 0, 1); value != 2 {
		t.Fatalf("overlap pixel=%d, want BODY 2 drawn over HEAD", value)
	}
	if value, _ := composed.Pixel(0, 0, 3); value != 2 {
		t.Fatalf("BODY pixel=%d, want 2 at y=3", value)
	}
}

func TestParseOriginalCombatTileSets(t *testing.T) {
	for _, test := range []struct {
		name  string
		count int
	}{
		{name: "DUNGCOM.DAX", count: 25},
		{name: "WILDCOM.DAX", count: 34},
		{name: "RANDCOM.DAX", count: 6},
	} {
		t.Run(test.name, func(t *testing.T) {
			blocks, err := dax.Parse(readOriginalMember(t, test.name))
			if err != nil {
				t.Fatal(err)
			}
			if len(blocks) != 1 {
				t.Fatalf("blocks=%d, want 1", len(blocks))
			}
			tiles, err := ParseCombatTiles(blocks[0].Data)
			if err != nil {
				t.Fatal(err)
			}
			if len(tiles.Tiles) != test.count {
				t.Fatalf("tiles=%d, want %d", len(tiles.Tiles), test.count)
			}
			if tiles.Tiles[0].Width() != 24 || tiles.Tiles[0].Height() != 24 {
				t.Fatalf("tile dimensions=%dx%d, want 24x24", tiles.Tiles[0].Width(), tiles.Tiles[0].Height())
			}
		})
	}
}

func readOriginalMember(t *testing.T, name string) []byte {
	t.Helper()
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("original image is not available: %v", err)
		}
		t.Fatal(err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	t.Fatalf("archive member %s is missing", name)
	return nil
}

func TestComposeHeadBodyTreatsBlackAsTransparentLayerPixel(t *testing.T) {
	head := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: filledPicturePixels(8, 3)}
	bodyPixels := filledPicturePixels(8, 0)
	bodyPixels[1] = 4
	body := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: bodyPixels}
	composed, err := ComposeHeadBody(head, body, 0)
	if err != nil {
		t.Fatal(err)
	}
	if value, _ := composed.Pixel(0, 0, 0); value != 3 {
		t.Fatalf("black BODY pixel erased HEAD: got %d want 3", value)
	}
	if value, _ := composed.Pixel(0, 1, 0); value != 4 {
		t.Fatalf("non-black BODY pixel=%d, want 4", value)
	}
}

func filledPicturePixels(size int, value uint8) []uint8 {
	pixels := make([]uint8, size)
	for index := range pixels {
		pixels[index] = value
	}
	return pixels
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
