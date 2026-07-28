package main

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
)

func TestWrapTextLinesUsesUnicodeRunesAndLineLimit(t *testing.T) {
	got := wrapTextLines("熔岩池中有火蜥蜴\n第二段", 4, 3)
	want := []string{"熔岩池中", "有火蜥蜴", "第二段"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wrapTextLines()=%q, want %q", got, want)
	}
}

func TestCombatTerrainEntryUsesWildernessCameraCenter(t *testing.T) {
	floor := mapdata.GenerateWilderness(0, 1)
	got, ok := combatTerrainEntry("WILDCOM", nil, floor, 25, 12, 3, 3)
	want, wantOK := floor.Entry(25, 12)
	if ok != wantOK || got != want {
		t.Fatalf("center entry=(%#v,%v), want (%#v,%v)", got, ok, want, wantOK)
	}
}

func TestCombatTerrainEntryDoesNotTreatRANDCOMAsFloor(t *testing.T) {
	if _, ok := combatTerrainEntry("RANDCOM", nil, mapdata.WildernessFloor{}, 0, 0, 0, 0); ok {
		t.Fatal("RANDCOM unexpectedly returned a full-floor entry")
	}
}

func TestCombatTerrainLayersDecodeGlobalRANDCOMNamespace(t *testing.T) {
	got := combatTerrainLayers("DUNGCOM", mapdata.BackgroundTile{TileIndex: 0x22})
	want := []combatTerrainLayer{
		{Atlas: "DUNGCOM", Index: 0x16},
		{Atlas: "RANDCOM", Index: 0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("table layers=%+v, want %+v", got, want)
	}
	got = combatTerrainLayers("DUNGCOM", mapdata.BackgroundTile{TileIndex: 0x27})
	if len(got) != 2 || got[1] != (combatTerrainLayer{Atlas: "RANDCOM", Index: 5}) {
		t.Fatalf("last RANDCOM layers=%+v", got)
	}
}

func TestCombatTerrainLayersKeepAtlasBoundsSeparate(t *testing.T) {
	if got := combatTerrainLayers("DUNGCOM", mapdata.BackgroundTile{TileIndex: 24}); !reflect.DeepEqual(got, []combatTerrainLayer{{Atlas: "DUNGCOM", Index: 24}}) {
		t.Fatalf("last DUNGCOM layer=%+v", got)
	}
	if got := combatTerrainLayers("WILDCOM", mapdata.BackgroundTile{TileIndex: 33}); !reflect.DeepEqual(got, []combatTerrainLayer{{Atlas: "WILDCOM", Index: 33}}) {
		t.Fatalf("last WILDCOM layer=%+v", got)
	}
	if got := combatTerrainLayers("DUNGCOM", mapdata.BackgroundTile{TileIndex: 0x21}); got != nil {
		t.Fatalf("namespace gap unexpectedly resolved: %+v", got)
	}
}

func TestMirroredCombatPlacementKeepsOriginalCPICAnchor(t *testing.T) {
	got := mirroredCombatAnchor(combat.TilePoint{X: 0, Y: 2})
	if got != (combat.TilePoint{X: 6, Y: 2}) {
		t.Fatalf("mirrored anchor=%+v, want (6,2)", got)
	}
}

func TestWrapTextLinesRejectsInvalidLayout(t *testing.T) {
	if got := wrapTextLines("中文", 0, 2); got != nil {
		t.Fatalf("zero-width wrap=%q, want nil", got)
	}
	if got := wrapTextLines("中文", 2, 0); got != nil {
		t.Fatalf("zero-line wrap=%q, want nil", got)
	}
}

func TestSelectCombatTerrainNameUsesDungeonStateWithoutAreaHeuristic(t *testing.T) {
	if got := selectCombatTerrainName(true, ""); got != "DUNGCOM" {
		t.Fatalf("dungeon terrain=%q, want DUNGCOM", got)
	}
	if got := selectCombatTerrainName(false, ""); got != "WILDCOM" {
		t.Fatalf("wilderness terrain=%q, want WILDCOM", got)
	}
	if got := selectCombatTerrainName(false, "RANDCOM"); got != "RANDCOM" {
		t.Fatalf("override terrain=%q, want RANDCOM", got)
	}
}
