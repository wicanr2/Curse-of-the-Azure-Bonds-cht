package game

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
	s.pendingMusicEvents = append(s.pendingMusicEvents, MusicEvent{
		Action:  "play",
		TrackID: binding.TrackID,
	})
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
