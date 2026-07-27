package geo

import (
	"archive/zip"
	"io"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestParseSyntheticPlanes(t *testing.T) {
	data := make([]byte, BlockSize)
	data[2] = 0xA5
	data[2+0x100] = 0x3C
	data[2+0x200] = 0x7E
	data[2+0x300] = 0xE4
	grid, err := Parse(1, data)
	if err != nil {
		t.Fatal(err)
	}
	cell, ok := grid.Cell(0, 0)
	if !ok || cell.WallDirections != [4]uint8{10, 5, 3, 12} || cell.Terrain != 0x7E || cell.DetailDirections != [4]uint8{0, 1, 2, 3} {
		t.Fatalf("cell=%+v, want decoded packed planes", cell)
	}
}

func TestCanMoveHonorsBothCellWallsAndBounds(t *testing.T) {
	var grid Grid
	grid.Cells[0][0].WallDirections[1] = 1
	if grid.CanMove(0, 0, 2) {
		t.Fatal("movement through current cell wall should be blocked")
	}
	grid.Cells[0][0].WallDirections[1] = 0
	grid.Cells[0][1].WallDirections[3] = 1
	if grid.CanMove(0, 0, 2) {
		t.Fatal("movement through neighbor cell wall should be blocked")
	}
	grid.Cells[0][1].WallDirections[3] = 0
	if !grid.CanMove(0, 0, 2) {
		t.Fatal("open adjacent cells should be traversable")
	}
	if grid.CanMove(0, 0, 6) {
		t.Fatal("movement outside grid should be blocked")
	}
}

func TestWrappedDungeonMovementCrossesMapEdge(t *testing.T) {
	var grid Grid
	grid.Cells[0][0].WallDirections[0] = 0
	grid.Cells[Height-1][0].WallDirections[2] = 0
	if !grid.CanMoveWrapped(0, 0, 0) {
		t.Fatal("open north edge should wrap to the last row")
	}
	grid.Cells[0][0].WallDirections[0] = 1
	if grid.CanMoveWrapped(0, 0, 0) {
		t.Fatal("current wrapped wall should block movement")
	}
	if got := WrapCoordinate(-1, Width); got != Width-1 {
		t.Fatalf("wrapped coordinate=%d, want %d", got, Width-1)
	}
}

func TestOriginalGEOBlocksHaveKnownShape(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()
	for _, member := range []string{"GEO2.DAX", "GEO3.DAX", "GEO4.DAX", "GEO5.DAX", "GEO6.DAX"} {
		data := zipMember(t, archive, member)
		blocks, parseErr := dax.Parse(data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if len(blocks) == 0 {
			t.Fatalf("%s has no blocks", member)
		}
		for _, block := range blocks {
			if len(block.Data) != BlockSize {
				t.Fatalf("%s block 0x%02X length=%d, want %d", member, block.Entry.ID, len(block.Data), BlockSize)
			}
			if _, parseErr := Parse(block.Entry.ID, block.Data); parseErr != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, parseErr)
			}
		}
	}
}

func zipMember(t *testing.T, archive *zip.ReadCloser, name string) []byte {
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
		if closeErr := reader.Close(); err != nil || closeErr != nil {
			t.Fatalf("read %s: read=%v close=%v", name, err, closeErr)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}
