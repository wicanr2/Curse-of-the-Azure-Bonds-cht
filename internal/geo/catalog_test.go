package geo

import (
	"archive/zip"
	"io"
	"path/filepath"
	"testing"
)

func TestOriginalGEOCatalogPreservesAllMapIDs(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()
	catalog := NewCatalog()
	for set := uint8(2); set <= 6; set++ {
		name := "GEO" + string(rune('0'+set)) + ".DAX"
		data := readCatalogMember(t, archive, name)
		if err := catalog.AddDAX(set, data); err != nil {
			t.Fatal(err)
		}
	}
	if catalog.Len() != 16 {
		t.Fatalf("catalog has %d blocks, want 16 original GEO blocks", catalog.Len())
	}
	for _, ref := range catalog.Refs() {
		grid, ok := catalog.Lookup(ref)
		if !ok || grid.BlockID != ref.BlockID {
			t.Fatalf("lookup failed for GEO%d block 0x%02X", ref.Set, ref.BlockID)
		}
	}
}

func readCatalogMember(t *testing.T, archive *zip.ReadCloser, name string) []byte {
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
		closeErr := reader.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("read %s: read=%v close=%v", name, err, closeErr)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}
