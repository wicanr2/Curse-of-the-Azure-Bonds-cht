package game

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// TestRealOverlandArrivalAndRouteGraphCoverage keeps the world-map contract
// honest at the boundary that can be checked without inventing city stories:
// every native location declared by the CoAB pack must be reachable through
// the directed adjacency graph, and ECL1's real arrival entry must project
// each native value into the corresponding world state. Individual city
// events remain owned by their ECL continuation tests.
func TestRealOverlandArrivalAndRouteGraphCoverage(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	allBlocks := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			allBlocks[block.Entry.ID] = block.Data
		}
	}

	probe := NewStateFromECLBlocks(trainingTestCatalog(t), allBlocks, 0x50)
	if probe.dataPack == nil {
		t.Fatal("CoAB game pack is unavailable")
	}
	overland, found := probe.dataPack.FindMapByKind("overland")
	if !found {
		t.Fatal("CoAB overland map declaration is unavailable")
	}
	if len(overland.Locations) != 14 {
		t.Fatalf("overland locations=%d, want 14", len(overland.Locations))
	}

	declared := make(map[uint8]bool, len(overland.Locations))
	for _, point := range overland.Locations {
		declared[point.Value] = true
	}
	for _, point := range overland.Locations {
		for _, destination := range point.Destinations {
			if !declared[destination] {
				t.Fatalf("location %q points to undeclared destination %d", point.ID, destination)
			}
		}
	}

	// The route table is directed in the game pack. Check that all declared
	// world locations are reachable from the normal opening location without
	// assuming that every road has a reverse edge.
	reachable := map[uint8]bool{0: true}
	for changed := true; changed; {
		changed = false
		for _, point := range overland.Locations {
			if !reachable[point.Value] {
				continue
			}
			for _, destination := range point.Destinations {
				if !reachable[destination] {
					reachable[destination] = true
					changed = true
				}
			}
		}
	}
	for _, point := range overland.Locations {
		if !reachable[point.Value] {
			t.Fatalf("world location %q (native value %d) is unreachable from Tilverton", point.ID, point.Value)
		}
	}

	// Each arrival is executed through the real ECL1 entry 1. The memory
	// values are the normal world-session preconditions used by the existing
	// new-game path; no location or menu text is injected into State.
	for _, point := range overland.Locations {
		point := point
		t.Run(point.ID, func(t *testing.T) {
			state := NewStateFromECLBlocks(trainingTestCatalog(t), allBlocks, 0x50)
			state.session.SetMemoryValue(0x4C59, 1)
			state.session.SetMemoryValue(0x4C5A, 1)
			state.session.SetMemoryValue(0x4C5B, 0xFF)
			if err := state.arriveAtWorldLocation(point.Value); err != nil {
				t.Fatalf("arrival native value %d: %v", point.Value, err)
			}
			if state.Area.CurrentCity != point.Value || state.Area.GameArea != 1 || state.Area.InDungeon {
				t.Fatalf("arrival %q projected area=%+v", point.ID, state.Area)
			}
			if got, ok := state.session.MemoryValue(0x4C9B); !ok || got != uint16(point.Value) {
				t.Fatalf("arrival %q current-location memory=%#x,%v, want %#x", point.ID, got, ok, point.Value)
			}
			if state.Location == LocationWilderness || state.LocationName == "" || state.OriginalLocation == "" {
				t.Fatalf("arrival %q did not project a world location: location=%v name=%q original=%q", point.ID, state.Location, state.LocationName, state.OriginalLocation)
			}
		})
	}
}
