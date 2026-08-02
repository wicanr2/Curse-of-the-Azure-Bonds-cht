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
	if binding.TrackID == s.activeMusicTrackID {
		return
	}
	s.activeMusicTrackID = binding.TrackID
	s.musicPlaybackSnapshot = nil
	s.pendingMusicEvents = append(s.pendingMusicEvents, MusicEvent{
		Action:  "play",
		TrackID: binding.TrackID,
	})
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
