package game

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"

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

// musicEventForTrack 是選曲／停止的決定點。三個方向都要釘：換曲、不變、停。
//
// ⚠ 「停」目前在正式資料上到不了——engine 的 pack 驗證擋掉 `track_id` 是空的
// binding。`TestEnginePackCannotExpressStopYet` 把那個限制釘住：它一旦鬆綁，
// 那條測試會紅，提醒把 binding 補上（spec 1192）。
func TestMusicEventForTrack(t *testing.T) {
	for _, item := range []struct {
		name           string
		previous, next string
		wantAction     string
		wantTrack      string
		wantChanged    bool
	}{
		{name: "換曲", previous: "a", next: "b", wantAction: "play", wantTrack: "b", wantChanged: true},
		{name: "同一首不重發", previous: "a", next: "a", wantChanged: false},
		{name: "空的代表停，不是放一首叫「」的曲子", previous: "a", next: "",
			wantAction: "stop", wantChanged: true},
		{name: "本來就沒在放也不用停", previous: "", next: "", wantChanged: false},
		{name: "從沒放到放", previous: "", next: "b", wantAction: "play", wantTrack: "b", wantChanged: true},
	} {
		t.Run(item.name, func(t *testing.T) {
			event, changed := musicEventForTrack(item.previous, item.next)
			if changed != item.wantChanged {
				t.Fatalf("changed ＝ %v，want %v", changed, item.wantChanged)
			}
			if !changed {
				return
			}
			if event.Action != item.wantAction || event.TrackID != item.wantTrack {
				t.Fatalf("得到 %+v，want action=%q track=%q",
					event, item.wantAction, item.wantTrack)
			}
		})
	}
}

// ★ 這條釘的是**別人家的限制**：共用 engine 的 pack 驗證不收「這裡不放音樂」的
// 宣告（`track_id` 空的 binding 會被判成 `references unknown track ""`）。
//
// ⚠ 所以 `musicEventForTrack` 的 `stop` 分支目前在正式資料上**到不了**。
// 這條測試一旦紅，代表 engine 那邊鬆綁了 —— 那時要做的是把「不放音樂」的
// binding 補進 game-pack，而不是把這條測試刪掉（spec 1192）。
func TestEnginePackCannotExpressStopYet(t *testing.T) {
	parts := map[string][]byte{"00-core.json": []byte(`{
	  "schema_version": 1,
	  "id": "music-lifecycle-test",
	  "default_locale": "en",
	  "locales": {"en": {"music-lifecycle-test.noop": "noop"}},
	  "music_tracks": [
	    {"id": "t1", "source_platform": "pc-9801", "reference_selector": 1, "driver_index": 0}
	  ],
	  "music_bindings": [
	    {"ecl_blocks": [1], "track_id": "t1"},
	    {"ecl_blocks": [2], "track_id": ""}
	  ]
	}`)}
	_, err := goldenbox.LoadPackPartsFS(
		func(name string) ([]byte, error) { return parts[name], nil },
		[]string{"00-core.json"})
	if err == nil {
		t.Fatal("engine 現在收得下空的 track_id 了：把「不放音樂」的 binding 補進 game-pack，" +
			"讓 `musicEventForTrack` 的 stop 分支真的用得到（spec 1192）")
	}
	if !strings.Contains(err.Error(), "unknown track") {
		t.Fatalf("擋下來的理由變了：%v", err)
	}
}
