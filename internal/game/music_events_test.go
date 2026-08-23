package game

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiomap"
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
// 「停」的呼叫端是音樂開關，見 `TestMusicSwitchIsTheOnlyThingThatStopsMusic`。
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

// ★ 空的 `track_id` 不是「這裡不放音樂」，它什麼都不是。
//
// ⚠ 這條原本叫 `TestEnginePackCannotExpressStopYet`，寫著「engine 一旦鬆綁就把
// 不放音樂的 binding 補進 game-pack」。**那是錯的指示**：原作沒有這種資料——
// 派曲常式（`sub_18AA7`）查不到就 `ret`，音樂繼續放；會停的只有玩家關音樂。
// 所以 engine 收不收空的 `track_id` 從頭到尾都不是卡點，而照原本那句話做，
// 只會做出一個原作沒有的行為（spec 1192）。
//
// 留著這條是因為**空的 `track_id` 該被擋下來**：它是打錯字，不是一種宣告。
func TestEmptyTrackIDIsRejectedBecauseItMeansNothing(t *testing.T) {
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
		t.Fatal("空的 `track_id` 應該被擋下來——它是打錯字，不是「這裡不放音樂」")
	}
	if !strings.Contains(err.Error(), "unknown track") {
		t.Fatalf("擋下來的理由變了：%v", err)
	}
}

// ★ 音樂開關是 `stop` 的**唯一**來源。
//
// 原作沒有「這一段不放音樂」這種資料——派曲表查不到就維持現況（`sub_18AA7`
// 的 default 直接 ret）。真正會停的只有玩家把音樂關掉：`MUSICSW` 翻成 1，
// 派曲常式第一關就寫 `MUSICNUM := 255` 然後叫驅動程式停（spec 1192）。
func TestMusicSwitchIsTheOnlyThingThatStopsMusic(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)
	state.requestMusicForCurrentBlock("")
	if got := state.ConsumeMusicEvents(); len(got) != 1 || got[0].Action != "play" {
		t.Fatalf("一開始應該放這個場景的曲子：%+v", got)
	}

	state.ToggleMusicSwitch()
	if !state.MusicSwitchOff() {
		t.Fatal("翻一次應該是關")
	}
	want := []MusicEvent{{Action: "stop"}}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("關掉應該立刻停：%+v", got)
	}

	// 關著的時候換場景不該發出任何東西——原作在派曲常式第一關就 ret 了。
	state.requestMusicForCurrentBlock("")
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("關著的時候還在發音樂事件：%+v", got)
	}

	// 再開要**立刻**放回這個場景該放的那一首，不是等下一次換場景。
	state.ToggleMusicSwitch()
	if state.MusicSwitchOff() {
		t.Fatal("翻兩次應該回到開")
	}
	back := []MusicEvent{{Action: "play", TrackID: "pc98-bgm-selector-03"}}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, back) {
		t.Fatalf("再開應該放回原本那一首：%+v", got)
	}
}

// 關著的時候關第二次不該再發一個 `stop`：沒在放就沒有東西要停。
func TestTurningMusicOffTwiceOnlyStopsOnce(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)
	state.ToggleMusicSwitch() // 從沒放過 → 沒有東西要停
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("本來就沒在放，不該發 stop：%+v", got)
	}
	state.ToggleMusicSwitch() // 開 → 放
	if got := state.ConsumeMusicEvents(); len(got) != 1 || got[0].Action != "play" {
		t.Fatalf("打開應該放：%+v", got)
	}
	state.ToggleMusicSwitch() // 關 → 停
	if got := state.ConsumeMusicEvents(); len(got) != 1 || got[0].Action != "stop" {
		t.Fatalf("關掉應該停：%+v", got)
	}
	state.musicSwitchOff = false
	state.musicSwitchOff = true
	state.requestMusicForCurrentBlock("")
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("已經停了就不該再停一次：%+v", got)
	}
}

// ★ 音效與音樂是**兩個**開關。原作 BGM 走 INT 7Eh（`sub_18BDB`）根本不看
// `SOUNDTYPE`，所以 Ctrl+S 關掉音效時音樂照放；反過來 Ctrl+O 關掉音樂時
// 音效照響。綁在一起是很容易犯又不會被任何測試抓到的錯。
func TestSoundAndMusicSwitchesAreIndependent(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)

	state.ToggleSoundSwitch()
	if !state.SoundSwitchOff() {
		t.Fatal("翻一次應該是關")
	}
	state.pendingSoundEvents = []SoundEvent{SoundStep}
	if got := state.ConsumeSoundEvents(); len(got) != 0 {
		t.Fatalf("音效關著還出聲：%+v", got)
	}
	// 音效關著，音樂照常。
	state.requestMusicForCurrentBlock("")
	if got := state.ConsumeMusicEvents(); len(got) != 1 || got[0].Action != "play" {
		t.Fatalf("關音效不該影響音樂：%+v", got)
	}

	state.ToggleSoundSwitch()
	state.pendingSoundEvents = []SoundEvent{SoundStep}
	if got := state.ConsumeSoundEvents(); len(got) != 1 || got[0] != SoundStep {
		t.Fatalf("再開音效應該出聲：%+v", got)
	}
	// 音樂關著，音效照常。
	state.ToggleMusicSwitch()
	state.ConsumeMusicEvents()
	state.pendingSoundEvents = []SoundEvent{SoundStep}
	if got := state.ConsumeSoundEvents(); len(got) != 1 {
		t.Fatalf("關音樂不該影響音效：%+v", got)
	}
}

// ★ 開戰換曲：原作在這裡**不看場景**，`INITCOMBAT`（COMPREP）把曲號直接推給
// `MSCPLAY`。remake 用 `context: "pc98-combat"` 的 binding 表達，所以任何一段
// 開戰都會換到戰鬥曲，而戰鬥結束要換回那一段自己的曲子。
func TestCombatSwitchesToTheBattleThemeAndBack(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)
	state.requestMusicForCurrentBlock("")
	scene := state.ConsumeMusicEvents()
	if len(scene) != 1 || scene[0].Action != "play" {
		t.Fatalf("場景曲沒放：%+v", scene)
	}
	sceneTrack := scene[0].TrackID

	state.requestCombatMusic(0)
	battle := state.ConsumeMusicEvents()
	if len(battle) != 1 || battle[0].Action != "play" {
		t.Fatalf("開戰應該換曲：%+v", battle)
	}
	if battle[0].TrackID == sceneTrack {
		t.Fatalf("開戰換到同一首（%s）：戰鬥曲沒接上", battle[0].TrackID)
	}

	state.restoreSceneMusic()
	back := state.ConsumeMusicEvents()
	if len(back) != 1 || back[0].TrackID != sceneTrack {
		t.Fatalf("戰鬥結束應該換回場景曲 %s：%+v", sceneTrack, back)
	}
}

// ★ 開戰換曲**不挑段**：原作那一段不看 `CURRENTECL`，所以每一個有 ECL 的段
// 開戰都要換得到曲子。少一段就是那一段的戰鬥沒有音樂——**而那不會報錯**。
func TestEveryECLBlockHasCombatMusic(t *testing.T) {
	state := NewState(testCatalog())
	if state.dataPack == nil {
		t.Skip("沒有 game pack")
	}
	blocks := []uint8{1, 2, 3, 4, 16, 17, 18, 21, 32, 33, 34, 35, 37, 48, 49,
		50, 51, 53, 64, 66, 67, 69, 80, 81, 82}
	for _, block := range blocks {
		if _, found := state.dataPack.FindMusicBinding(block, combatMusicContext); !found {
			t.Errorf("段 0x%02X 開戰沒有戰鬥曲", block)
		}
	}
}

// TestMonsterSetCueSelectsTheDungeonTwoCombatTrack 釘住 `47h` 那個分岔真的接上了。
//
// ★ 原作的 `INITCOMBAT`（COMPREP，`overlay-10:1D8Dh`）是
// `cmp byte [LOADMONNUM], 47h` / `jnz`：相等時把曲目 `0Bh` 推給 `MSCPLAY`，
// 否則推 `07h`。**它不看場景**——戰鬥曲是由怪物組決定的（spec 1192）。
//
// ⚠ 是 `jnz`，也就是**恰好相等**才換曲，不是「大於等於」。
func TestMonsterSetCueSelectsTheDungeonTwoCombatTrack(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	blocks := 0
	for _, binding := range pack.MusicBindings {
		if binding.Context == "pc98-combat-dungeon-two" {
			blocks = len(binding.ECLBlocks)
		}
	}
	if blocks == 0 {
		t.Fatal("game pack 沒有 `pc98-combat-dungeon-two` 的曲目綁定")
	}
	cue, ok := pack.FindMusicCue(1, "monster_set", 0x47)
	if !ok || cue.Context != "pc98-combat-dungeon-two" {
		t.Fatalf("怪物組 47h 的 cue ＝ %+v（有 %v），預期 `pc98-combat-dungeon-two`", cue, ok)
	}
	if _, ok := pack.FindMusicCue(1, "monster_set", 0x46); ok {
		t.Error("怪物組 46h 也匹配到了：原作是 `jnz`，恰好相等才換曲")
	}
}

func TestTitleAndCreationHaveTheirOwnMusic(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("建構不該發音樂：%+v", got)
	}

	state.RequestTitleMusic()
	title := state.ConsumeMusicEvents()
	if len(title) != 1 || title[0].Action != "play" {
		t.Fatalf("標題曲沒放：%+v", title)
	}

	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	creation := state.ConsumeMusicEvents()
	if len(creation) != 1 || creation[0].Action != "play" {
		t.Fatalf("角色建立曲沒放：%+v", creation)
	}
	if creation[0].TrackID == title[0].TrackID {
		t.Fatalf("角色建立和標題放同一首（%s）", creation[0].TrackID)
	}
}

// ★ 事件驅動的換曲點都**不挑段**：每一個有 ECL 的段都要找得到那一首。
// 少一段就是那一段沒有那個場合的音樂——**而那不會報錯**。
func TestEventDrivenMusicCoversEveryECLBlock(t *testing.T) {
	state := NewState(testCatalog())
	if state.dataPack == nil {
		t.Skip("沒有 game pack")
	}
	blocks := []uint8{1, 2, 3, 4, 16, 17, 18, 21, 32, 33, 34, 35, 37, 48, 49,
		50, 51, 53, 64, 66, 67, 69, 80, 81, 82}
	for _, context := range []string{
		combatMusicContext, titleMusicContext, creationMusicContext,
	} {
		for _, block := range blocks {
			if _, found := state.dataPack.FindMusicBinding(block, context); !found {
				t.Errorf("段 0x%02X 沒有 %q 的曲子", block, context)
			}
		}
	}
}

// ★ 結局過場也有自己那一首（PC-98 overlay-18 `168Dh`：`MUSICNO := 0Ah` 之後
// 立刻 `MSCPLAY`，才開始印結局文字）。與開戰、開場、角色建立同一類——**不看
// `CURRENTECL`**，所以每一個有 ECL 的段都要綁得到。
//
// ⚠ 這條擋的是「打通關卻沒有結局曲」，而那**不會報錯**：沒有綁定時
// `requestMusicForCurrentBlock` 直接 return，結局照樣沿用戰鬥曲演完。
func TestEveryECLBlockHasEndingMusic(t *testing.T) {
	state := NewState(testCatalog())
	if state.dataPack == nil {
		t.Skip("沒有 game pack")
	}
	blocks := []uint8{1, 2, 3, 4, 16, 17, 18, 21, 32, 33, 34, 35, 37, 48, 49,
		50, 51, 53, 64, 66, 67, 69, 80, 81, 82}
	for _, block := range blocks {
		binding, found := state.dataPack.FindMusicBinding(block, endingMusicContext)
		if !found {
			t.Errorf("段 0x%02X 的結局沒有曲子", block)
			continue
		}
		// 曲目 10（`reference_selector` 10）＝ `pc98-bgm-selector-0a`。
		if binding.TrackID != "pc98-bgm-selector-0a" {
			t.Errorf("段 0x%02X 的結局曲是 %q，原作是曲目 10", block, binding.TrackID)
		}
	}
}

// TestEndingSceneSwitchesToTheEndingTrack 釘住換曲點的**位置**：原作在印第一頁
// 結局文字之前就換好了，所以 `beginEndingScene` 一跑就要發事件。
func TestEndingSceneSwitchesToTheEndingTrack(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{0x01: {}}, 0x01)
	state.ConsumeMusicEvents()
	state.requestCombatMusic(0x01)
	if got := state.ConsumeMusicEvents(); len(got) == 0 {
		t.Fatal("先放戰鬥曲，才看得出結局有沒有換曲")
	}

	state.beginEndingScene()
	events := state.ConsumeMusicEvents()
	if len(events) != 1 || events[0].Action != "play" ||
		events[0].TrackID != "pc98-bgm-selector-0a" {
		t.Fatalf("結局過場沒有換到結局曲：%+v", events)
	}
	if state.endingPageIndex != 0 || !state.endingScene {
		t.Fatalf("換曲要發生在第一頁之前：endingScene=%v page=%d",
			state.endingScene, state.endingPageIndex)
	}
}

// ★ 全滅也有自己那一首（POSTCOM，PC-98 overlay-05 `1955h`：印完
// 「モンスターはパーティーを全滅させ、喜んでいる。」之後 `MUSICNO := 2` 再
// `MSCPLAY`）。非全滅那一條走 `18F6h` 的 `jmp` 直接跳過那個換曲點——**兩條不是
// 同一首**，混用會讓打輸的時候響起場景曲。
func TestEveryECLBlockHasPartyWipeMusic(t *testing.T) {
	state := NewState(testCatalog())
	if state.dataPack == nil {
		t.Skip("沒有 game pack")
	}
	blocks := []uint8{1, 2, 3, 4, 16, 17, 18, 21, 32, 33, 34, 35, 37, 48, 49,
		50, 51, 53, 64, 66, 67, 69, 80, 81, 82}
	for _, block := range blocks {
		binding, found := state.dataPack.FindMusicBinding(block, partyWipeMusicContext)
		if !found {
			t.Errorf("段 0x%02X 的全滅沒有曲子", block)
			continue
		}
		if binding.TrackID != "pc98-bgm-selector-02" {
			t.Errorf("段 0x%02X 的全滅曲是 %q，原作是曲目 2", block, binding.TrackID)
		}
	}
}

// TestEndingAndPartyWipeTracksAreNotTheSceneTrack 釘住三個事件驅動換曲點彼此
// 不同，也都不等於場景曲——**綁錯到同一首不會報錯**，只會安靜地沿用。
func TestEndingAndPartyWipeTracksAreNotTheSceneTrack(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]string{}
	for _, context := range []string{
		combatMusicContext, endingMusicContext, partyWipeMusicContext,
	} {
		binding, found := pack.FindMusicBinding(0x01, context)
		if !found {
			t.Fatalf("段 0x01 沒有 %q 的綁定", context)
		}
		if previous, clash := seen[binding.TrackID]; clash {
			t.Errorf("%q 與 %q 綁到同一首 %s", context, previous, binding.TrackID)
		}
		seen[binding.TrackID] = context
	}
	scene, found := pack.FindMusicBinding(0x01, "")
	if found {
		if context, clash := seen[scene.TrackID]; clash {
			t.Errorf("段 0x01 的場景曲 %s 與 %q 同一首", scene.TrackID, context)
		}
	}
}

// ⚠ 這條的表在 `internal/audiomap`，不在測試檔裡：`cmd/music-change-points`
// 與 `cmd/remake-status` 用同一份，報表與測試才不會各自漂移。

// TestEveryOriginalMusicChangePointHasARemakeCounterpart 釘住 13／13。
//
// ⚠ 「接上了」不等於「在原作會發的那一刻發」。這條測的是**有沒有落點**，
// 時機要實機比對。
func TestEveryOriginalMusicChangePointHasARemakeCounterpart(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	if len(audiomap.ChangePoints) != 13 {
		t.Fatalf("原作是 13 個換曲點，表裡有 %d 個", len(audiomap.ChangePoints))
	}
	trackBySelector := map[int]string{}
	for _, track := range pack.MusicTracks {
		trackBySelector[int(track.ReferenceSelector)] = track.ID
	}
	bindings := make([]audiomap.Binding, 0, len(pack.MusicBindings))
	for _, binding := range pack.MusicBindings {
		bindings = append(bindings, audiomap.Binding{
			Context: binding.Context, TrackID: binding.TrackID})
	}
	for _, result := range audiomap.Resolve(trackBySelector, bindings) {
		if result.TrackID == "" {
			t.Errorf("%s（%s）選的曲目 %d 不在 pack 裡",
				result.Point.Site, result.Point.Event, result.Point.Selector)
			continue
		}
		if !result.Wired {
			t.Errorf("%s（%s）沒有落點：context %q ＋ 曲目 %d（%s）",
				result.Point.Site, result.Point.Event, result.Point.Context,
				result.Point.Selector, result.TrackID)
		}
	}
}
