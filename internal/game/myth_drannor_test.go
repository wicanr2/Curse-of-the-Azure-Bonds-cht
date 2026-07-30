package game

import (
	"archive/zip"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealPlayerPathStandingStoneToBurialGlen(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	all := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			all[block.Entry.ID] = block.Data
		}
	}
	state := NewStateFromECLBlocks(testCatalog(), all, 0x50)
	state.session.SetMemoryValue(0x4C59, 1)
	state.session.SetMemoryValue(0x4C5A, 1)
	state.session.SetMemoryValue(0x4C5B, 0xFF)

	// AREA wilderness arrival supplies current-city value 4 and invokes the
	// ordinary ECL SearchLocation lifecycle.
	if err := state.arriveAtWorldLocation(4); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationStandingStone ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Standing Stone arrival location=%v originals=%v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != gamePackText(t, state, "myth-drannor.tyranthraxus-reveal") {
		t.Fatalf("Tyranthraxus reveal message=%q", state.Message)
	}
	wantStandingActions := []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}
	for step := 0; step < 6 && !reflect.DeepEqual(state.currentOriginalChoices, wantStandingActions); step++ {
		switch {
		case state.Mode == ModeEvent:
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		case reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}):
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected Standing Stone continuation mode=%v originals=%v",
				state.Mode, state.currentOriginalChoices)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, wantStandingActions) {
		t.Fatalf("Standing Stone mode=%v actions=%v message=%q original=%q",
			state.Mode, state.currentOriginalChoices, state.Message, state.OriginalEvent)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	wantDestinations := []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR", "MYTH DRANNOR"}
	if !reflect.DeepEqual(state.currentOriginalChoices, wantDestinations) {
		t.Fatalf("Standing Stone destinations=%v, want %v", state.currentOriginalChoices, wantDestinations)
	}
	for address, want := range map[uint16]uint16{
		0x4C02: 2,
		0x4C03: 8,
		0x4C04: 11,
		0x4C05: 13,
	} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("route selector memory[0x%04X]=%d,%v want %d,true", address, got, ok, want)
		}
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"WILDERNESS", "EXIT"}) {
		t.Fatalf("Myth Drannor route prompt=%q originals=%v", state.Prompt, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeMap || !state.pendingWorldTravel || state.pendingWorldDestination != 13 {
		t.Fatalf("world travel mode=%v pending=%v destination=%d",
			state.Mode, state.pendingWorldTravel, state.pendingWorldDestination)
	}

	// Enter completes this bounded wilderness travel slice and hands current
	// city 13 back to the same ECL session.
	if err := state.EnterPlaces(); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationMythDrannor ||
		state.Message != gamePackText(t, state, "myth-drannor.edge") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA"}) {
		t.Fatalf("Myth Drannor edge location=%v originals=%v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x40 ||
		state.Area.GameArea != 6 || !state.Area.InDungeon ||
		state.GeoMapSet != 6 || state.GeoMapBlock != 0x40 ||
		state.Message != gamePackText(t, state, "myth-drannor.helm-north") {
		t.Fatalf("Burial Glen mode=%v block=0x%02X area=%+v geo=%d/0x%02X message=%q",
			state.Mode, state.session.CurrentBlockID(), state.Area,
			state.GeoMapSet, state.GeoMapBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Burial Glen continuation mode=%v, want dungeon", state.Mode)
	}
}

func gamePackText(t *testing.T, state State, messageID string) string {
	t.Helper()
	if state.dataPack == nil {
		t.Fatal("game pack is unavailable")
	}
	text, ok := state.dataPack.Text(messageID, state.catalog.Language)
	if !ok {
		t.Fatalf("game pack message %q is unavailable for locale %q",
			messageID, state.catalog.Language)
	}
	return text
}
