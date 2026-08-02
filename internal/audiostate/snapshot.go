// Package audiostate defines renderer-neutral persistence records for active
// one-shot players. Asset lookup and backend playback stay in platform code.
package audiostate

import (
	"fmt"
	"strconv"
)

type Backend string

const (
	BackendDOSWAV      Backend = "dos_wav"
	BackendPC98Speaker Backend = "pc98_speaker"
)

const MaxActiveOneShots = 64

type OneShot struct {
	Backend        Backend `json:"backend"`
	Key            string  `json:"key"`
	PositionFrames uint64  `json:"position_frames"`
}

type Snapshot struct {
	Version  int       `json:"version"`
	Enabled  bool      `json:"enabled"`
	OneShots []OneShot `json:"one_shots,omitempty"`
}

const CurrentVersion = 1

func (snapshot Snapshot) Validate() error {
	if snapshot.Version != CurrentVersion {
		return fmt.Errorf("unsupported one-shot audio snapshot version %d", snapshot.Version)
	}
	if len(snapshot.OneShots) > MaxActiveOneShots {
		return fmt.Errorf("one-shot audio count %d exceeds %d", len(snapshot.OneShots), MaxActiveOneShots)
	}
	if !snapshot.Enabled && len(snapshot.OneShots) != 0 {
		return fmt.Errorf("disabled audio snapshot contains active one-shots")
	}
	seen := make(map[string]bool, len(snapshot.OneShots))
	for index, shot := range snapshot.OneShots {
		if shot.Backend != BackendDOSWAV && shot.Backend != BackendPC98Speaker {
			return fmt.Errorf("one-shot %d has unknown backend %q", index, shot.Backend)
		}
		if shot.Key == "" || len(shot.Key) > 64 {
			return fmt.Errorf("one-shot %d key length %d outside 1..64", index, len(shot.Key))
		}
		if shot.Backend == BackendDOSWAV {
			selector, err := strconv.ParseUint(shot.Key, 10, 8)
			if err != nil || strconv.FormatUint(selector, 10) != shot.Key {
				return fmt.Errorf("one-shot %d DOS selector %q is not canonical uint8", index, shot.Key)
			}
		}
		identity := string(shot.Backend) + "\x00" + shot.Key
		if seen[identity] {
			return fmt.Errorf("one-shot %d duplicates %s/%s", index, shot.Backend, shot.Key)
		}
		seen[identity] = true
	}
	return nil
}

func Clone(snapshot Snapshot) Snapshot {
	snapshot.OneShots = append([]OneShot(nil), snapshot.OneShots...)
	return snapshot
}
