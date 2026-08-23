package main

import (
	"image"
	"image/color"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

func TestCombatSpeedElapsedMatchesReferenceMultiplier(t *testing.T) {
	second := time.Second
	if got := combatSpeedElapsed(second, 4); got != second {
		t.Fatalf("default speed elapsed=%v want %v", got, second)
	}
	if fast, slow := combatSpeedElapsed(second, 0), combatSpeedElapsed(second, 9); fast <= second || slow >= second {
		t.Fatalf("speed ordering fast=%v default=%v slow=%v", fast, second, slow)
	}
	if got := combatSpeedElapsed(second, 255); got != combatSpeedElapsed(second, 9) {
		t.Fatalf("out-of-range speed was not clamped: %v", got)
	}
}

func TestCombatVisualResumeElapsedKeepsSavedBaseAtEverySpeed(t *testing.T) {
	base := 700 * time.Millisecond
	if got := combatVisualResumeElapsed(base, 0, 4); got != base {
		t.Fatalf("zero-delta resume=%v want=%v", got, base)
	}
	if got, want := combatVisualResumeElapsed(base, time.Second, 4), base+time.Second; got != want {
		t.Fatalf("default-speed resume=%v want=%v", got, want)
	}
	if fast, slow := combatVisualResumeElapsed(base, time.Second, 0), combatVisualResumeElapsed(base, time.Second, 9); fast <= slow {
		t.Fatalf("resume speed ordering fast=%v slow=%v", fast, slow)
	}
}

func TestImageCoverTransformUsesGlobalDestinationOrigin(t *testing.T) {
	scale, x, y := imageCoverTransform(88, 88, image.Rect(48, 48, 224, 224))
	if scale != 2 || x != 48 || y != 48 {
		t.Fatalf("transform=(%v,%v,%v), want (2,48,48)", scale, x, y)
	}
}

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

func TestCombatVisualPointUsesAreaCenterThenOrderedImpactTarget(t *testing.T) {
	event := combat.VisualEvent{
		From: combat.TilePoint{X: 1, Y: 1},
		To:   combat.TilePoint{X: 4, Y: 3},
		Impacts: []combat.VisualImpactTarget{
			{TargetID: "orc-1", To: combat.TilePoint{X: 3, Y: 2}, Hit: true},
			{TargetID: "orc-2", To: combat.TilePoint{X: 5, Y: 4}, Hit: true},
		},
	}
	camera := combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{}, false)
	_, _, travelToX, travelToY, _, _ := combatVisualPoint(event, combat.VisualFrame{
		Phase: combat.VisualTravel, ImpactIndex: -1, Progress: 1,
	}, camera)
	if travelToX != 120 || travelToY != 168 {
		t.Fatalf("travel target=(%v,%v), want area center (120,168)", travelToX, travelToY)
	}
	_, _, impactToX, impactToY, _, _ := combatVisualPoint(event, combat.VisualFrame{
		Phase: combat.VisualImpact, ImpactIndex: 1,
	}, camera)
	if impactToX != 72 || impactToY != 216 {
		t.Fatalf("impact target=(%v,%v), want second target (72,216)", impactToX, impactToY)
	}
}

func TestCombatVisualPointUsesCurrentLineSegment(t *testing.T) {
	event := combat.VisualEvent{
		From: combat.TilePoint{X: 1, Y: 1},
		To:   combat.TilePoint{X: 3, Y: 1},
		Segments: []combat.VisualPathSegment{
			{From: combat.TilePoint{X: 3, Y: 1}, To: combat.TilePoint{X: 5, Y: 1}},
		},
	}
	fromX, fromY, toX, toY, x, y := combatVisualPoint(event, combat.VisualFrame{
		Phase: combat.VisualSegmentTravel, SegmentIndex: 0, Progress: 0.5,
	}, combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{}, false))
	if fromX != 168 || fromY != 72 || toX != 72 || toY != 72 || x != 120 || y != 72 {
		t.Fatalf("line points=(%v,%v)->(%v,%v) at (%v,%v)", fromX, fromY, toX, toY, x, y)
	}
}

func TestCombatVisualPreservesEachKilledTargetUntilItsDeathPhase(t *testing.T) {
	event := combat.VisualEvent{
		Impacts: []combat.VisualImpactTarget{
			{TargetID: "orc-1", To: combat.TilePoint{X: 4, Y: 2}, Hit: true, Killed: true},
			{TargetID: "orc-2", To: combat.TilePoint{X: 5, Y: 3}, Hit: true, Killed: true},
		},
	}
	if _, ok := combatVisualPreservedImpact(event, combat.VisualFrame{
		Phase: combat.VisualTravel, ImpactIndex: -1,
	}, "orc-2"); !ok {
		t.Fatal("second target was removed before area travel")
	}
	if _, ok := combatVisualPreservedImpact(event, combat.VisualFrame{
		Phase: combat.VisualDeath, ImpactIndex: 0,
	}, "orc-1"); ok {
		t.Fatal("first target survived into its death phase")
	}
	if _, ok := combatVisualPreservedImpact(event, combat.VisualFrame{
		Phase: combat.VisualDeath, ImpactIndex: 0,
	}, "orc-2"); !ok {
		t.Fatal("second target was removed during first target death")
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
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	definition, found := pack.FindCombatVisual("missile", "travel")
	if !found {
		t.Fatal("missing missile/travel combat visual")
	}
	camera := combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{}, false)
	event := combat.VisualEvent{
		From: combat.TilePoint{X: 5, Y: 2},
		To:   combat.TilePoint{X: 1, Y: 2},
	}
	if key, flip := combatArrowSprite(definition, event, camera); key != "comspr-block-02-item-00.png" || flip {
		t.Fatalf("east arrow=(%q,%v)", key, flip)
	}
	event.From, event.To = event.To, event.From
	if key, flip := combatArrowSprite(definition, event, camera); key != "comspr-block-82-item-00.png" || flip {
		t.Fatalf("west arrow=(%q,%v)", key, flip)
	}
}

func TestCombatMagicMissileCyclesOriginalFourCOMSPRFrames(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	definition, found := pack.FindCombatVisual("magic_missile", "travel")
	if !found {
		t.Fatal("missing magic_missile/travel combat visual")
	}
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
		key, flip := combatMagicMissileSprite(definition, event, frame)
		if key != expected.key || flip != expected.flip {
			t.Fatalf("frame %d=(%q,%v), want (%q,%v)", index, key, flip, expected.key, expected.flip)
		}
	}
}

func TestCombatMagicImpactCyclesOriginalCOMSPRHitFrames(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	definition, found := pack.FindCombatVisual("magic_missile", "impact")
	if !found {
		t.Fatal("missing magic_missile/impact combat visual")
	}
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
		key, flip := combatMagicImpactSprite(definition, combat.VisualFrame{
			Phase: combat.VisualImpact, Progress: float64(index) / 4,
		})
		if key != expected.key || flip != expected.flip {
			t.Fatalf("impact %d=(%q,%v), want (%q,%v)", index, key, flip, expected.key, expected.flip)
		}
	}
}

func TestCombatLightningLineCyclesOriginalElectricalFrames(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	definition, found := pack.FindCombatVisual("lightning_bolt", "line")
	if !found {
		t.Fatal("missing lightning_bolt/line combat visual")
	}
	want := []struct {
		key  string
		flip bool
	}{
		{"comspr-block-06-item-00.png", false},
		{"comspr-block-06-item-00.png", true},
		{"comspr-block-86-item-00.png", true},
		{"comspr-block-86-item-00.png", false},
	}
	for index, expected := range want {
		key, flip := combatPathSequenceSprite(
			definition,
			combat.TilePoint{X: 2, Y: 1},
			combat.TilePoint{X: 6, Y: 1},
			float64(index)/12,
		)
		if key != expected.key || flip != expected.flip {
			t.Fatalf("electrical frame %d=(%q,%v), want (%q,%v)", index, key, flip, expected.key, expected.flip)
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

func TestWrapTextLinesByWidthUsesFaceAdvance(t *testing.T) {
	face := basicfont.Face7x13
	cell := font.MeasureString(face, "A").Ceil()
	got := wrapTextLinesByWidth("ABCD", face, cell*2, 3)
	want := []string{"AB", "CD"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wrapTextLinesByWidth()=%q, want %q", got, want)
	}
}

func TestJournalDisplayPagesPreserveEverySourceEntryWithoutClipping(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	entry := pack.Locales["zh-TW"]["journal.48"]
	pages := journalDisplayPages([]string{entry}, "", basicfont.Face7x13, font.MeasureString(basicfont.Face7x13, "AAAAAA").Ceil(), 3)
	if len(pages) < 2 {
		t.Fatalf("long stable journal entry produced %d display page(s)", len(pages))
	}
	got := strings.ReplaceAll(strings.Join(pages, ""), "\n", "")
	want := strings.ReplaceAll(entry, "\n", "")
	if got != want {
		t.Fatal("display pagination dropped or duplicated journal text")
	}
}

func TestSleepTwinkleUsesGeneratedVisualBinding(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	definition, found := pack.FindCombatVisual("sleep", "impact")
	if !found || definition.Generator != "pc98.twinkle" || len(definition.Frames) != 0 {
		t.Fatalf("sleep visual=%+v found=%v", definition, found)
	}
}

func TestFitTextToWidthKeepsSingleLineInsideDeclaredRegion(t *testing.T) {
	face := basicfont.Face7x13
	maxWidth := font.MeasureString(face, "ABC").Ceil()
	got := fitTextToWidth("ABCDEFGHI", face, maxWidth)
	if got == "ABCDEFGHI" || font.MeasureString(face, got).Ceil() > maxWidth {
		t.Fatalf("fitTextToWidth()=%q width=%d max=%d", got, font.MeasureString(face, got).Ceil(), maxWidth)
	}
}

func TestCombatTerrainEntryUsesWildernessCameraCenter(t *testing.T) {
	floor := mapdata.GenerateWilderness(0, 1)
	got, ok := combatMovementTerrainEntry("WILDCOM", nil, floor, 25, 12, false, 3, 3)
	want, wantOK := floor.Entry(25, 12)
	if ok != wantOK || got != want {
		t.Fatalf("center entry=(%#v,%v), want (%#v,%v)", got, ok, want, wantOK)
	}
}

func TestCombatTerrainEntryDoesNotTreatRANDCOMAsFloor(t *testing.T) {
	if _, ok := combatMovementTerrainEntry("RANDCOM", nil, mapdata.WildernessFloor{}, 0, 0, false, 0, 0); ok {
		t.Fatal("RANDCOM unexpectedly returned a full-floor entry")
	}
}

// 地形層與戰鬥員層必須是同一條座標路徑的兩端。先前地形直接拿畫面欄列當
// 地圖座標，於是人站在空地、腳下卻畫出牆。
func TestCombatMapTileForScreenInvertsFighterAnchor(t *testing.T) {
	for _, camera := range []combat.CombatCamera{
		combat.NewCombatCamera(combat.TilePoint{}, combat.TilePoint{X: 3, Y: 3}, false),
		combat.NewCombatCamera(combat.TilePoint{X: 0, Y: 0}, combat.TilePoint{X: 3, Y: 3}, true),
		combat.NewCombatCamera(combat.TilePoint{X: 9, Y: 4}, combat.TilePoint{X: 3, Y: 3}, true),
	} {
		for _, tile := range []combat.TilePoint{{X: 0, Y: 0}, {X: 5, Y: 2}, {X: 9, Y: 6}} {
			anchor := mirroredCombatAnchor(camera.Apply(tile))
			if got := combatMapTileForScreen(camera, anchor.X, anchor.Y); got != tile {
				t.Fatalf("camera=%+v tile=%+v anchor=%+v round trip=%+v", camera, tile, anchor, got)
			}
		}
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

func TestCombatScanTacticalMapPreservesRawTDAndTDEFBytes(t *testing.T) {
	dungeon := &mapdata.DungeonFloor{}
	for y := 7; y < 23; y++ {
		for x := 18; x < 50; x++ {
			dungeon.Tiles[y][x] = 22
		}
	}
	app := app{state: &game.State{}, dungeonFloor: dungeon, combatTerrainMode: "DUNGCOM"}

	got, err := app.combatScanTacticalMap()
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != 32 || got.Height != 16 {
		t.Fatalf("fallback tactical dimensions=%dx%d, want 32x16", got.Width, got.Height)
	}
	if got.Tiles[0] != 22 {
		t.Fatalf("local TD=%d, want raw floor ID 22", got.Tiles[0])
	}
	want := mapdata.BackgroundTiles[22]
	definition := got.Definitions[21]
	if definition.HT != want.MoveCost || definition.LOS != want.Height || definition.SYM != want.Field || definition.Raw3 != want.TileIndex {
		t.Fatalf("TDEF[21]=%+v, want BackgroundTiles[22]=%+v", definition, want)
	}
}

func TestCombatScanTacticalMapFailsClosedOnUnknownTD(t *testing.T) {
	dungeon := &mapdata.DungeonFloor{}
	for y := 7; y < 23; y++ {
		for x := 18; x < 50; x++ {
			dungeon.Tiles[y][x] = 22
		}
	}
	dungeon.Tiles[7][18] = 66
	app := app{state: &game.State{}, dungeonFloor: dungeon, combatTerrainMode: "DUNGCOM"}
	if _, err := app.combatScanTacticalMap(); err == nil {
		t.Fatal("unknown TD unexpectedly accepted")
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

func TestCombatCameraFocusFollowsActiveFighterOnLargeMap(t *testing.T) {
	active := combat.Fighter{HasCombatPosition: true, CombatX: 19, CombatY: 11}
	if got := combatCameraFocus(active, true, 0, 31, 0, 15); got != (combat.TilePoint{X: 19, Y: 11}) {
		t.Fatalf("active camera focus=%+v", got)
	}
	if got := combatCameraFocus(combat.Fighter{}, false, 0, 31, 0, 15); got != (combat.TilePoint{X: 16, Y: 8}) {
		t.Fatalf("fallback camera focus=%+v", got)
	}
}

func TestCombatPreviewFocusFindsOriginalBossSpriteWithoutMovingIt(t *testing.T) {
	fighters := []combat.Fighter{
		{Side: combat.SideEnemy, SpriteBlock: 0x45, HasCombatPosition: true, CombatX: 12, CombatY: 4},
		{Side: combat.SideEnemy, SpriteBlock: 0x47, HasCombatPosition: true, CombatX: 24, CombatY: 9},
	}
	if got, ok := combatPreviewFocus(fighters, 0x47); !ok || got != (combat.TilePoint{X: 24, Y: 9}) {
		t.Fatalf("boss preview focus=%+v ok=%v", got, ok)
	}
	if _, ok := combatPreviewFocus(fighters, 0); ok {
		t.Fatal("disabled preview camera unexpectedly selected a fighter")
	}
}

// 牆面素材的 EGA index 13 是透明鍵不是顏色（spec 1131）：不遮罩就會在天空與
// 牆角畫出桃紅色塊。
func TestMaskWallSymbolsReplacesMagentaWithTransparentIndex(t *testing.T) {
	piece := gfx.PieceSet{
		SetID:   1,
		Symbols: map[uint8]gfx.Picture{7: {ItemCount: 1, Pixels: []uint8{13, 4, 13, 0}}},
	}
	maskWallSymbols(piece)
	got := piece.Symbols[7].Pixels
	want := []uint8{wallSymbolTransparentIndex, 4, wallSymbolTransparentIndex, 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("masked pixels=%v, want %v", got, want)
	}
}

// wallLayoutRawStamps 要回**每一個非零格**，而且 `SymbolID` 是 WALLDEF 的原始
// 編號——符號組留給 `symbolImageForGlobalID` 照編號去判，與這一格出自第幾段無關。
//
// ★ 這是原作的形狀：`sub_39F` 從一塊平的三槽緩衝取位元組，原封不動交給
// `PUT8X8SYMBOL`，而後者只比編號落在哪一段（spec 1185）。
//
// ⚠ 先前這裡是兩條路徑：引擎照「這一格出自哪一筆記錄」選符號區塊，選不到的才
// 回退。那條規則會把多段牆面組自己的編號解到別的牆面組去，而每個索引仍然落在
// 合法圖片裡 ⇒ 牆畫在對的位置、用錯的圖、不報錯。
func TestWallLayoutRawStampsReturnsEveryCellWithItsRawID(t *testing.T) {
	piece := gfx.PieceSet{
		SetID:          1,
		WallDefs:       make([]gfx.WallDef, 1),
		SymbolSetIDs:   []uint8{1},
		SymbolBlockIDs: []uint8{1},
		Symbols:        map[uint8]gfx.Picture{1: {ItemCount: 70}},
	}
	// 牆位 0 是 1 欄 × 2 列，資料落在 156 bytes 的第 0..1 格。
	piece.WallDefs[0].Rows[0][0] = 1  // 天空格，第 0 段（共用）
	piece.WallDefs[0].Rows[0][1] = 76 // 一般牆磚，第 1 段
	call := gfx.WallLayoutCall{WallType: 1, Layout: 0, RowStart: 4, ColStart: 5}

	stamps := wallLayoutRawStamps(piece, call)
	if len(stamps) != 2 {
		t.Fatalf("stamps=%d，want 2（兩格都要回來）：%+v", len(stamps), stamps)
	}
	if stamps[0].SymbolID != 1 || stamps[0].Row != 4 || stamps[0].Column != 5 {
		t.Errorf("第一格＝%+v，want 編號 1 在 (列 4, 欄 5)", stamps[0])
	}
	if stamps[1].SymbolID != 76 || stamps[1].Row != 5 {
		t.Errorf("第二格＝%+v，want 編號 76 在第二列", stamps[1])
	}
}

// ⚠ 編號落在**別段**的格子一樣要回來。這一支就是提爾佛頓 `0Eh` 近端右側牆
// 那兩格的形狀：記錄掛第 3 段，WALLDEF 卻引用第 2 段的 `185`。舊的兩條路徑
// 設計會讓引擎那一條把它算成 `185 − 186 = −1` 而整格靜默消失。
func TestWallLayoutRawStampsKeepsCrossSegmentSymbols(t *testing.T) {
	piece := gfx.PieceSet{
		SetID:          3,
		WallDefs:       make([]gfx.WallDef, 1),
		SymbolSetIDs:   []uint8{3},
		SymbolBlockIDs: []uint8{3},
		Symbols:        map[uint8]gfx.Picture{3: {ItemCount: 70}},
	}
	piece.WallDefs[0].Rows[0][0] = 185 // 第 2 段的最後一項
	piece.WallDefs[0].Rows[0][1] = 200 // 第 3 段
	call := gfx.WallLayoutCall{WallType: 11, Layout: 0, RowStart: 0, ColStart: 0}

	stamps := wallLayoutRawStamps(piece, call)
	if len(stamps) != 2 {
		t.Fatalf("stamps=%+v，want 兩格都回來", stamps)
	}
	if stamps[0].SymbolID != 185 || stamps[1].SymbolID != 200 {
		t.Errorf("編號＝%d／%d，want 185／200（原始編號，不做任何段內換算）",
			stamps[0].SymbolID, stamps[1].SymbolID)
	}
}

// ★ 段由**編號本身**決定，不是由「這一筆 WALLDEF 掛哪一段」決定。
//
// ⚠ 這一條是提爾佛頓 `0Eh` 近端右側牆少畫兩格的成因：那面牆的 WALLDEF 引用
// 編號 `185`，那是**第 2 段的最後一項**（基準 `74h` ＝ 116），而那一筆記錄掛的
// 是第 3 段（基準 `BAh` ＝ 186）⇒ `185 − 186 = −1` ⇒ `WallSymbolItem` 回
// `ok=false` ⇒ 整格靜默消失。原版與 remake 因此在那條斜邊差 48 個取樣點，
// 而牆面其餘部分完全相同（spec 1185）。
func TestWallSymbolSegmentIsChosenByTheIDNotTheRecord(t *testing.T) {
	for _, item := range []struct {
		id      uint8
		segment int
		index   int
	}{
		{1, 0, 1},    // 共用段的第一項
		{45, 0, 45},  // 共用段的最後一項
		{46, 1, 0},   // 第 1 段的第一項（基準 2Eh）
		{115, 1, 69}, // 第 1 段的最後一項
		{116, 2, 0},  // 第 2 段的第一項（基準 74h）
		{185, 2, 69}, // ⚠ 就是那一格：第 2 段的最後一項
		{186, 3, 0},  // 第 3 段的第一項（基準 BAh）
		{255, 3, 69}, // 第 3 段的最後一項
	} {
		segment, index := wallSymbolSegment(item.id)
		if segment != item.segment || index != item.index {
			t.Errorf("編號 %d ＝ 第 %d 段第 %d 項，want 第 %d 段第 %d 項",
				item.id, segment, index, item.segment, item.index)
		}
	}
}
