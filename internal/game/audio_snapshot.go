package game

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"

// SetOneShotPlaybackSnapshot attaches the renderer's bounded audible
// continuation immediately before writing remake JSON.
func (s *State) SetOneShotPlaybackSnapshot(snapshot *audiostate.Snapshot) error {
	if snapshot == nil {
		s.oneShotPlaybackSnapshot = nil
		return nil
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	copy := audiostate.Clone(*snapshot)
	s.oneShotPlaybackSnapshot = &copy
	return nil
}

func (s *State) oneShotSnapshot() *audiostate.Snapshot {
	if s.oneShotPlaybackSnapshot == nil {
		return nil
	}
	copy := audiostate.Clone(*s.oneShotPlaybackSnapshot)
	return &copy
}

// OneShotPlaybackSnapshot returns a defensive copy for the platform adapter.
func (s *State) OneShotPlaybackSnapshot() (*audiostate.Snapshot, bool) {
	copy := s.oneShotSnapshot()
	return copy, copy != nil
}
