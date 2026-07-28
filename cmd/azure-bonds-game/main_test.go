package main

import (
	"reflect"
	"testing"

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
