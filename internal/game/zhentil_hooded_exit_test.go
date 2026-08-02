package game

import (
	"slices"
	"strings"
	"testing"
)

func TestRealZhentilHoodedWomanReachesBeholderCave(t *testing.T) {
	state, grid := newRealZhentilShrineState(t)

	// Establish the original Dimswart escort flag without creating a fighter.
	setZhentilShrineCell(state, grid, 2, 14, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}

	setZhentilShrineCell(state, grid, 4, 12, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 39 ||
		state.Message != requireGamePackText(t, state, "zhentil.hooded_offer") {
		t.Fatalf("hooded offer mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(state.Choices, "/"); got != "是/否" ||
		state.Message != requireGamePackText(t, state, "zhentil.hooded_follow") {
		t.Fatalf("hooded follow choices=%q message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		state.Message != requireGamePackText(t, state, "zhentil.fzoul_interrupts") {
		t.Fatalf("Fzoul interruption picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "zhentil.fzoul_retreats") {
		t.Fatalf("Fzoul retreat message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}

	if state.session.CurrentBlockID() != 0x22 || !state.PictureRequested ||
		state.PictureBlock != 40 || state.Message != requireGamePackText(t, state, "dexam.arrival") {
		t.Fatalf("Dexam arrival block=0x%02x picture=%v/%d message=%q",
			state.session.CurrentBlockID(), state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "dexam.journal_30") {
		t.Fatalf("Dexam journal prompt=%q", state.Message)
	}
	for _, id := range []string{"journal.30.1", "journal.30.2"} {
		if !slices.Contains(state.JournalPages, requireGamePackText(t, state, id)) {
			t.Fatalf("Journal 30 missing %s: %q", id, state.JournalPages)
		}
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(state.Choices, "/"); got != "戰鬥/等待/撤退/接近/談判" ||
		state.Message != requireGamePackText(t, state, "dexam.amulet_choice") {
		t.Fatalf("Dexam encounter choices=%q message=%q", state.Choices, state.Message)
	}

	// The original script interrupts every tactical choice with Fzoul's arrival.
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		state.Message != requireGamePackText(t, state, "dexam.fzoul_journal_7") {
		t.Fatalf("Fzoul arrival picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	for _, id := range []string{"journal.7.1", "journal.7.2"} {
		if !slices.Contains(state.JournalPages, requireGamePackText(t, state, id)) {
			t.Fatalf("Journal 7 missing %s: %q", id, state.JournalPages)
		}
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "dexam.kills_fzoul") {
		t.Fatalf("Fzoul death message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		state.Message != requireGamePackText(t, state, "dexam.fzoul_bond_fades") {
		t.Fatalf("bond fade picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 40 ||
		state.Message != requireGamePackText(t, state, "dexam.kill_order") {
		t.Fatalf("Dexam order picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "dexam.amulet_rises") {
		t.Fatalf("amulet departure message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "dexam.altar_melee") {
		t.Fatalf("altar melee message=%q", state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x22 ||
		state.GeoMapSet != 4 || state.GeoMapBlock != 0x25 ||
		state.DungeonX != 4 || state.DungeonY != 5 || state.DungeonDirection != 0 {
		t.Fatalf("beholder cave handoff mode=%v block=0x%02x geo=%d/%02x coords=%d,%d,%d message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.Message)
	}
}
