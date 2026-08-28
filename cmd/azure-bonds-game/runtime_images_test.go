package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeImageManifestIsCompleteAndSelfContained(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "runtime-images")
	catalog, err := loadRuntimeImageCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(catalog.manifest.Tiles), 48; got != want {
		t.Fatalf("tiles=%d, want %d", got, want)
	}
	combatCount := 0
	for _, source := range []string{"DUNGCOM", "WILDCOM", "RANDCOM"} {
		if len(catalog.manifest.Combat[source]) == 0 {
			t.Fatalf("combat source %s is empty", source)
		}
		combatCount += len(catalog.manifest.Combat[source])
	}
	if combatCount != 65 {
		t.Fatalf("combat images=%d, want 65", combatCount)
	}
	if got, want := len(catalog.manifest.Symbols), 1625; got != want {
		t.Fatalf("symbols=%d, want %d", got, want)
	}
	if got, want := len(catalog.manifest.Sky), 3; got != want {
		t.Fatalf("sky=%d, want %d", got, want)
	}
	if got, want := len(catalog.manifest.Walls), 20; got != want {
		t.Fatalf("wall definitions=%d, want %d", got, want)
	}
	all := append([]runtimeImageRecord{}, catalog.manifest.Tiles...)
	for _, records := range catalog.manifest.Combat {
		all = append(all, records...)
	}
	all = append(all, catalog.manifest.Symbols...)
	all = append(all, catalog.manifest.Sky...)
	for _, record := range all {
		if filepath.IsAbs(record.File) {
			t.Fatalf("manifest path must be relative: %q", record.File)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(record.File))); err != nil {
			t.Fatalf("manifest image %q: %v", record.File, err)
		}
	}
}
