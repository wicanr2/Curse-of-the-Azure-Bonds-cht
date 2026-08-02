package game

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
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

func TestRemakeSaveRestoresDefensiveOneShotSnapshot(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, Abilities: party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 10, Dexterity: 11, Constitution: 12, Charisma: 10},
	}}
	want := audiostate.Snapshot{Version: audiostate.CurrentVersion, Enabled: true, OneShots: []audiostate.OneShot{{
		Backend: audiostate.BackendPC98Speaker, Key: "spell_hit", PositionFrames: 777,
	}}}
	if err := state.SetOneShotPlaybackSnapshot(&want); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "audio-save-v9.json")
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	got, ok := loaded.OneShotPlaybackSnapshot()
	if !ok || !reflect.DeepEqual(*got, want) {
		t.Fatalf("loaded one-shot snapshot=%+v ok=%v", got, ok)
	}
	got.OneShots[0].Key = "mutated"
	second, _ := loaded.OneShotPlaybackSnapshot()
	if second.OneShots[0].Key != "spell_hit" {
		t.Fatal("one-shot snapshot getter leaked mutable records")
	}
}

func TestRemakeSaveRestoresStableMusicIDAndSynthesisSnapshot(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, Abilities: party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 10, Dexterity: 11, Constitution: 12, Charisma: 10},
	}}
	state.activeMusicTrackID = "pc98-bgm-selector-05"
	want := pc98music.TrackPCMStreamSnapshot{
		Version: 1, Selector: 5, OutputSampleRate: 44_100,
		Playback:   pc98music.TrackPlaybackSnapshot{Version: 1},
		SynthState: []byte{1, 2, 3}, Pending: []byte{4, 5, 6, 7},
	}
	if err := state.SetMusicPlaybackSnapshot(&want); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "music-save-v8.json")
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	trackID, got, ok := loaded.MusicPlaybackSnapshot()
	if !ok || trackID != state.activeMusicTrackID || got == nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("loaded music track=%q snapshot=%+v ok=%v", trackID, got, ok)
	}
	got.SynthState[0] = 99
	_, second, _ := loaded.MusicPlaybackSnapshot()
	if second.SynthState[0] != 1 {
		t.Fatal("music snapshot getter leaked mutable synthesis bytes")
	}
}

func TestUnmappedECLBlocksDoNotGuessMusic(t *testing.T) {
	for _, block := range []uint8{0x30, 0x52} {
		state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
			block: {},
		}, block)
		state.requestMusicForCurrentBlock("")
		if got := state.ConsumeMusicEvents(); len(got) != 0 {
			t.Fatalf("block 0x%02X guessed music events %+v", block, got)
		}
	}
}

func TestPC98WorldTownMusicContextsUseProvenSceneRoles(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
		0x50: {},
	}, 0x50)
	state.requestMusicForCurrentBlock("")
	state.requestMusicForCurrentBlock("pc98-town-services-menu")
	want := []MusicEvent{
		{Action: "play", TrackID: "pc98-bgm-selector-05"},
		{Action: "play", TrackID: "pc98-bgm-selector-06"},
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("world/town music events=%+v, want %+v", got, want)
	}
	state.requestMusicForCurrentBlock("pc98-town-services-menu")
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("unchanged town context replayed music: %+v", got)
	}
}

func TestPC98PictureCueSwitchesWorldAndTownMusic(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
		0x50: {},
	}, 0x50)
	state.requestMusicForCurrentBlock("")
	state.requestMusicForSignal("picture", 80)
	state.requestMusicForSignal("picture", 80)
	state.requestMusicForSignal("picture", 121)
	want := []MusicEvent{
		{Action: "play", TrackID: "pc98-bgm-selector-05"},
		{Action: "play", TrackID: "pc98-bgm-selector-06"},
		{Action: "play", TrackID: "pc98-bgm-selector-05"},
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("picture cue music events=%+v, want %+v", got, want)
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
