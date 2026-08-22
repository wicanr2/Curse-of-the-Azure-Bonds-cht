package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// MusicEvent 是規則層送給平台 adapter 的音樂意圖。TrackID 來自 game-pack，
// 因此規則與 ECL runtime 不需要知道檔案格式、解碼器或音訊裝置。
type MusicEvent struct {
	Action  string
	TrackID string
}

func (s *State) requestMusicForCurrentBlock(context string) {
	if s.dataPack == nil || s.session == nil {
		return
	}
	binding, found := s.dataPack.FindMusicBinding(s.session.CurrentBlockID(), context)
	if !found {
		return
	}
	event, changed := musicEventForTrack(s.activeMusicTrackID, binding.TrackID)
	if !changed {
		return
	}
	s.activeMusicTrackID = binding.TrackID
	s.musicPlaybackSnapshot = nil
	s.pendingMusicEvents = append(s.pendingMusicEvents, event)
}

// musicEventForTrack 決定「從這一首換到那一首」要發什麼。抽出來是為了**測得到**：
// 走 `requestMusicForCurrentBlock` 得先有一份 game-pack，而目前的 engine
// 根本不收「不放音樂」的宣告（見下），於是那條分支在整合測試裡永遠跑不到。
//
// ★ 空的 `TrackID` ＝**這裡不放音樂**，要把正在放的停掉，不是放一首叫「」的曲子。
// 原作的停止是在派曲處寫 `MUSICNUM := 255`（PC-98 常駐 `9451h`、初始化 `2F5Ah`），
// `MUSICSW := 0`（`2F46h`）是同一件事的開關版（spec 1192）。
//
// ⚠ **目前這條分支在正式資料上到不了**：engine 的 pack 驗證會把 `track_id` 是空的
// binding 擋掉（`music_bindings[i] references unknown track ""`），所以「這裡不放
// 音樂」在 pack 裡表達不出來。程式碼先擺著是因為少了它，空的 `TrackID` 會發成
// `play` 然後在 adapter 那裡查無此曲、**只留一行 log**：音樂繼續放著，看不出來。
// 要真的用上，得在共用 engine 那邊讓 pack 表達得出「停」。
func musicEventForTrack(previous, next string) (MusicEvent, bool) {
	if next == previous {
		return MusicEvent{}, false
	}
	if next == "" {
		return MusicEvent{Action: "stop"}, true
	}
	return MusicEvent{Action: "play", TrackID: next}, true
}

func (s *State) requestMusicForSignal(signal string, value uint16) {
	if s.dataPack == nil || s.session == nil {
		return
	}
	cue, found := s.dataPack.FindMusicCue(s.session.CurrentBlockID(), signal, value)
	if !found {
		return
	}
	s.requestMusicForCurrentBlock(cue.Context)
}

func (s *State) requestMusicIfBlockChanged(previous uint8) {
	if s.session != nil && s.session.CurrentBlockID() != previous {
		s.requestMusicForCurrentBlock("")
	}
}

// ConsumeMusicEvents 只轉交一次尚未處理的音樂意圖。平台 adapter 可在曲目素材
// 尚未解碼時忽略事件，但 deterministic runtime 仍能驗證原版選曲行為。
func (s *State) ConsumeMusicEvents() []MusicEvent {
	events := append([]MusicEvent(nil), s.pendingMusicEvents...)
	s.pendingMusicEvents = nil
	return events
}

// SetMusicPlaybackSnapshot attaches the platform player's exact synthesis
// continuation immediately before a remake JSON save.
func (s *State) SetMusicPlaybackSnapshot(snapshot *pc98music.TrackPCMStreamSnapshot) error {
	if snapshot != nil && s.activeMusicTrackID == "" {
		return fmt.Errorf("music playback snapshot has no active track ID")
	}
	if snapshot == nil {
		s.musicPlaybackSnapshot = nil
		return nil
	}
	copy := cloneTrackPCMStreamSnapshot(*snapshot)
	s.musicPlaybackSnapshot = &copy
	return nil
}

func cloneTrackPCMStreamSnapshot(snapshot pc98music.TrackPCMStreamSnapshot) pc98music.TrackPCMStreamSnapshot {
	snapshot.SynthState = append([]byte(nil), snapshot.SynthState...)
	snapshot.Pending = append([]byte(nil), snapshot.Pending...)
	for index := range snapshot.Playback.Channels {
		machine := &snapshot.Playback.Channels[index].Machine
		machine.CallStack = append([]int(nil), machine.CallStack...)
		machine.LoopStack = append([]pc98music.LoopSnapshot(nil), machine.LoopStack...)
	}
	return snapshot
}

func (s *State) musicSnapshot() *partySave.MusicSnapshot {
	if s.activeMusicTrackID == "" {
		return nil
	}
	result := &partySave.MusicSnapshot{TrackID: s.activeMusicTrackID}
	if s.musicPlaybackSnapshot != nil {
		copy := cloneTrackPCMStreamSnapshot(*s.musicPlaybackSnapshot)
		result.Stream = &copy
	}
	return result
}

// MusicPlaybackSnapshot returns the loaded continuation for the platform
// adapter. A nil stream means the stable track restarts normally when a local
// driver is available.
func (s *State) MusicPlaybackSnapshot() (string, *pc98music.TrackPCMStreamSnapshot, bool) {
	if s.activeMusicTrackID == "" {
		return "", nil, false
	}
	if s.musicPlaybackSnapshot == nil {
		return s.activeMusicTrackID, nil, true
	}
	copy := cloneTrackPCMStreamSnapshot(*s.musicPlaybackSnapshot)
	return s.activeMusicTrackID, &copy, true
}
