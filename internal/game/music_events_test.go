package game

import (
	"reflect"
	"testing"
)

func TestActiveECLBlockRequestsPC98MusicSelector(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
		0x01: {},
	}, 0x01)
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("construction requested music before gameplay: %+v", got)
	}
	state.requestMusicForCurrentBlock("")
	want := []MusicEvent{{
		Action:  "play",
		TrackID: "pc98-bgm-selector-03",
	}}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("music events=%+v, want %+v", got, want)
	}
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("music events were not consumed: %+v", got)
	}
}

func TestUnmappedAndContextualECLBlocksDoNotGuessMusic(t *testing.T) {
	for _, block := range []uint8{0x30, 0x50, 0x52} {
		state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
			block: {},
		}, block)
		state.requestMusicForCurrentBlock("")
		if got := state.ConsumeMusicEvents(); len(got) != 0 {
			t.Fatalf("block 0x%02X guessed music events %+v", block, got)
		}
	}
}

func TestECLBlockTransitionRequestsDestinationMusicOnce(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
		0x01: {},
		0x02: {},
	}, 0x01)
	if err := state.session.Switch(0x02); err != nil {
		t.Fatal(err)
	}
	state.requestMusicIfBlockChanged(0x01)
	want := []MusicEvent{{
		Action:  "play",
		TrackID: "pc98-bgm-selector-09",
	}}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("transition music events=%+v, want %+v", got, want)
	}
	state.requestMusicIfBlockChanged(0x02)
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("unchanged block replayed music: %+v", got)
	}
}
