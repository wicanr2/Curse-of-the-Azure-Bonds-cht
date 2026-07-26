package gfx

import (
	"archive/zip"
	"io"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestParsePictureExpandsPackedNibblesAndMask(t *testing.T) {
	data := make([]byte, 17+8)
	data[0], data[1] = 2, 0 // height 2
	data[2], data[3] = 1, 0 // width units 1 => 8 pixels
	data[8] = 1
	data[17], data[18], data[19], data[20] = 0x12, 0xF0, 0x34, 0x56
	picture, err := ParsePicture(data, true, 0xF)
	if err != nil {
		t.Fatal(err)
	}
	if picture.Width() != 8 || picture.Height() != 2 || picture.ItemCount != 1 {
		t.Fatalf("picture=%+v", picture)
	}
	want := []uint8{1, 2, 16, 0, 3, 4, 5, 6}
	for x, value := range want {
		got, ok := picture.Pixel(0, x, 0)
		if !ok || got != value {
			t.Fatalf("pixel x=%d got=%d ok=%t want=%d", x, got, ok, value)
		}
	}
}

func TestOriginalTileAndSymbolPicturesParse(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()
	for _, member := range []string{"TILES.DAX", "8X8D2.DAX", "8X8D3.DAX", "8X8D4.DAX", "8X8D5.DAX", "8X8D6.DAX"} {
		data := readMember(t, archive, member)
		blocks, parseErr := dax.Parse(data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			picture, parseErr := ParsePicture(block.Data, false, 0)
			if parseErr != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, parseErr)
			}
			if picture.Width() == 0 || picture.Height() == 0 || len(picture.Pixels) != picture.ItemSize()*int(picture.ItemCount) {
				t.Fatalf("%s block 0x%02X invalid picture=%+v", member, block.Entry.ID, picture)
			}
		}
	}
}

func TestOriginalWallDefinitionsParse(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()
	for _, member := range []string{"WALLDEF2.DAX", "WALLDEF3.DAX", "WALLDEF4.DAX", "WALLDEF5.DAX", "WALLDEF6.DAX"} {
		blocks, parseErr := dax.Parse(readMember(t, archive, member))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			if _, parseErr := ParseWallDefs(block.Data); parseErr != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, parseErr)
			}
		}
	}
}

func readMember(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}
