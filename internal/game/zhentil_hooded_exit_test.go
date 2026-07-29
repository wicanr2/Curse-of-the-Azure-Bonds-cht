package game

import (
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
		!strings.Contains(state.Message, "兜帽女子忽然現身") ||
		!strings.Contains(state.Message, "主人能幫助你們") {
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
		!strings.Contains(state.Message, "跟她走") {
		t.Fatalf("hooded follow choices=%q message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		!strings.Contains(state.Message, "奇異的微光") ||
		!strings.Contains(state.Message, "弗佐爾・錢布瑞爾") {
		t.Fatalf("Fzoul interruption picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "轉身衝出房間") {
		t.Fatalf("Fzoul retreat message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}

	if state.session.CurrentBlockID() != 0x22 || !state.PictureRequested ||
		state.PictureBlock != 40 || !strings.Contains(state.Message, "眼魔德克薩姆") ||
		!strings.Contains(state.Message, "披甲牛頭人") {
		t.Fatalf("Dexam arrival block=0x%02x picture=%v/%d message=%q",
			state.session.CurrentBlockID(), state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "手札第 30 條") {
		t.Fatalf("Dexam journal prompt=%q", state.Message)
	}
	journals := strings.Join(state.JournalPages, "\n")
	if !strings.Contains(journals, "手札條目 30（1/2）") ||
		!strings.Contains(journals, "總督察長") ||
		!strings.Contains(journals, "兩三個星期") {
		t.Fatalf("Journal 30 was not unlocked: %q", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(state.Choices, "/"); got != "戰鬥/等待/撤退/接近/談判" ||
		!strings.Contains(state.Message, "洛山達護符") {
		t.Fatalf("Dexam encounter choices=%q message=%q", state.Choices, state.Message)
	}

	// The original script interrupts every tactical choice with Fzoul's arrival.
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		!strings.Contains(state.Message, "手札第 7 條") {
		t.Fatalf("Fzoul arrival picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	journals = strings.Join(state.JournalPages, "\n")
	if !strings.Contains(journals, "手札條目 7（1/2）") ||
		!strings.Contains(journals, "貝恩霸權") ||
		!strings.Contains(journals, "只要我還活著") {
		t.Fatalf("Journal 7 was not unlocked: %q", journals)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "轟成一堆灰燼") {
		t.Fatalf("Fzoul death message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 33 ||
		!strings.Contains(state.Message, "弗佐爾的枷印消退") {
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
		!strings.Contains(state.Message, "殺了他們") {
		t.Fatalf("Dexam order picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "護符正飄向他") {
		t.Fatalf("amulet departure message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "混戰") {
		t.Fatalf("altar melee message=%q", state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x22 ||
		state.GeoMapSet != 4 || state.GeoMapBlock != 0x22 ||
		state.DungeonX != 4 || state.DungeonY != 5 || state.DungeonDirection != 0 {
		t.Fatalf("beholder cave handoff mode=%v block=0x%02x geo=%d/%02x coords=%d,%d,%d message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.Message)
	}
}
