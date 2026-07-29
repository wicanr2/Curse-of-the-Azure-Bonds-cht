package main

import (
	"image"
	"image/color"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
)

func TestChromaKeyTopLeftRestoresMaskedCombatSprite(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	key := color.NRGBA{R: 255, G: 82, B: 82, A: 255}
	source.SetNRGBA(0, 0, key)
	source.SetNRGBA(1, 0, key)
	source.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 255, B: 255, A: 255})
	got := chromaKeyTopLeft(source)
	if _, _, _, alpha := got.At(1, 0).RGBA(); alpha != 0 {
		t.Fatalf("key pixel alpha=%#x", alpha)
	}
	if _, _, _, alpha := got.At(0, 1).RGBA(); alpha != 0xFFFF {
		t.Fatalf("sprite pixel alpha=%#x", alpha)
	}
}

func TestDrawCombatVisualRendersMissileTravelAndImpact(t *testing.T) {
	event := combat.VisualEvent{
		Kind: combat.VisualMissile,
		From: combat.TilePoint{X: 1, Y: 3},
		To:   combat.TilePoint{X: 5, Y: 3},
		Hit:  true,
	}
	fromX, fromY, toX, toY, x, y := combatVisualPoint(event, combat.VisualFrame{
		Phase: combat.VisualTravel, Progress: 0.5,
	}, combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{}, false))
	if fromX != 264 || fromY != 168 || toX != 72 || toY != 168 || x != 168 || y != 168 {
		t.Fatalf("travel points=(%v,%v)->(%v,%v) at (%v,%v)", fromX, fromY, toX, toY, x, y)
	}
}

func TestCombatProjectileDirectionUsesOriginalClockwiseOctants(t *testing.T) {
	tests := []struct {
		dx, dy int
		want   int
	}{
		{0, -4, 0}, {4, -4, 1}, {4, 0, 2}, {4, 4, 3},
		{0, 4, 4}, {-4, 4, 5}, {-4, 0, 6}, {-4, -4, 7},
	}
	for _, test := range tests {
		if got := combatProjectileDirection(test.dx, test.dy); got != test.want {
			t.Fatalf("direction(%d,%d)=%d, want %d", test.dx, test.dy, got, test.want)
		}
	}
}

func TestCombatArrowSpriteUsesCOMSPRDirectionBlocks(t *testing.T) {
	camera := combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{}, false)
	event := combat.VisualEvent{
		From: combat.TilePoint{X: 5, Y: 2},
		To:   combat.TilePoint{X: 1, Y: 2},
	}
	if key, flip := combatArrowSprite(event, camera); key != "comspr-block-02-item-00.png" || flip {
		t.Fatalf("east arrow=(%q,%v)", key, flip)
	}
	event.From, event.To = event.To, event.From
	if key, flip := combatArrowSprite(event, camera); key != "comspr-block-82-item-00.png" || flip {
		t.Fatalf("west arrow=(%q,%v)", key, flip)
	}
}

func TestCombatMagicMissileCyclesOriginalFourCOMSPRFrames(t *testing.T) {
	event := combat.VisualEvent{
		From: combat.TilePoint{X: 0, Y: 0},
		To:   combat.TilePoint{X: 4, Y: 0},
	}
	want := []struct {
		key  string
		flip bool
	}{
		{"comspr-block-05-item-00.png", false},
		{"comspr-block-05-item-00.png", true},
		{"comspr-block-85-item-00.png", true},
		{"comspr-block-85-item-00.png", false},
	}
	for index, expected := range want {
		frame := combat.VisualFrame{Phase: combat.VisualTravel, Progress: float64(index) / 12}
		key, flip := combatMagicMissileSprite(event, frame)
		if key != expected.key || flip != expected.flip {
			t.Fatalf("frame %d=(%q,%v), want (%q,%v)", index, key, flip, expected.key, expected.flip)
		}
	}
}

func TestCombatMagicImpactCyclesOriginalCOMSPRHitFrames(t *testing.T) {
	want := []struct {
		key  string
		flip bool
	}{
		{"comspr-block-0A-item-00.png", false},
		{"comspr-block-0A-item-00.png", true},
		{"comspr-block-8A-item-00.png", true},
		{"comspr-block-8A-item-00.png", false},
	}
	for index, expected := range want {
		key, flip := combatMagicImpactSprite(combat.VisualFrame{
			Phase: combat.VisualImpact, Progress: float64(index) / 4,
		})
		if key != expected.key || flip != expected.flip {
			t.Fatalf("impact %d=(%q,%v), want (%q,%v)", index, key, flip, expected.key, expected.flip)
		}
	}
}

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

func TestCombatMovementTerrainEntryMapsFallbackAndReferenceCoordinates(t *testing.T) {
	dungeon := &mapdata.DungeonFloor{}
	dungeon.Tiles[7][18] = 26 // BackgroundTiles table: cost 2, RANDCOM graphic 0x22.
	fallback, ok := combatMovementTerrainEntry("DUNGCOM", dungeon, mapdata.WildernessFloor{}, 0, 0, false, 0, 0)
	if !ok || fallback.MoveCost != 2 || fallback.TileIndex != 0x22 {
		t.Fatalf("fallback dungeon terrain=%+v ok=%v", fallback, ok)
	}
	reference, ok := combatMovementTerrainEntry("DUNGCOM", dungeon, mapdata.WildernessFloor{}, 0, 0, true, 18, 7)
	if !ok || reference != fallback {
		t.Fatalf("reference dungeon terrain=%+v ok=%v, want %+v", reference, ok, fallback)
	}

	wilderness := mapdata.GenerateWilderness(0, 1)
	got, ok := combatMovementTerrainEntry("WILDCOM", nil, wilderness, 25, 12, false, 3, 3)
	want, wantOK := wilderness.Entry(25, 12)
	if ok != wantOK || got != want {
		t.Fatalf("fallback wilderness terrain=(%+v,%v), want (%+v,%v)", got, ok, want, wantOK)
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

func TestWallStampNativePositionMatchesReferenceThreeCellOrigin(t *testing.T) {
	x, y, ok := wallStampNativePosition(0, 0)
	if !ok || x != 24 || y != 24 {
		t.Fatalf("origin=(%d,%d,%v), want (24,24,true)", x, y, ok)
	}
	x, y, ok = wallStampNativePosition(10, 10)
	if !ok || x != 104 || y != 104 {
		t.Fatalf("last=(%d,%d,%v), want (104,104,true)", x, y, ok)
	}
	if _, _, ok := wallStampNativePosition(0, -1); ok {
		t.Fatal("negative reference column was not clipped")
	}
	if _, _, ok := wallStampNativePosition(11, 0); ok {
		t.Fatal("reference row 11 was not clipped")
	}
}
