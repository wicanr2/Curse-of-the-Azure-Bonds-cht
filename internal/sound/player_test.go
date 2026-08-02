package sound

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
)

func TestDurationToSampleFramesCeilInvertsEbitenPositionRounding(t *testing.T) {
	for _, frames := range []uint64{0, 1, 2, 127, 44_099, 44_100, 1_234_567} {
		position := time.Duration(frames) * time.Second / sampleRate
		if got := durationToSampleFramesCeil(position, sampleRate); got != frames {
			t.Fatalf("frames=%d position=%s restored=%d", frames, position, got)
		}
	}
}

type fakeOneShotPlayer struct {
	playing        bool
	position       time.Duration
	endAfterRead   bool
	setPositionErr error
	playCount      int
	pauseCount     int
}

func (p *fakeOneShotPlayer) Rewind() error   { p.position = 0; return nil }
func (p *fakeOneShotPlayer) Play()           { p.playing = true; p.playCount++ }
func (p *fakeOneShotPlayer) Pause()          { p.playing = false; p.pauseCount++ }
func (p *fakeOneShotPlayer) IsPlaying() bool { return p.playing }
func (p *fakeOneShotPlayer) Position() time.Duration {
	position := p.position
	if p.endAfterRead {
		p.playing = false
	}
	return position
}
func (p *fakeOneShotPlayer) SetPosition(position time.Duration) error {
	if p.setPositionErr != nil {
		return p.setPositionErr
	}
	p.position = position
	return nil
}

func TestSampleFrameDurationRoundTrip(t *testing.T) {
	for _, frames := range []uint64{0, 1, 2, 127, 44_099, 44_100, 1_234_567} {
		position, err := sampleFramesToDurationFloor(frames, sampleRate)
		if err != nil {
			t.Fatal(err)
		}
		if got := durationToSampleFramesCeil(position, sampleRate); got != frames {
			t.Fatalf("frames=%d position=%s restored=%d", frames, position, got)
		}
	}
	if _, err := sampleFramesToDurationFloor(^uint64(0), sampleRate); err == nil {
		t.Fatal("overflowing sample frame position was accepted")
	}
}

func TestOneShotSnapshotKeepsOnlyActiveAudiblePlayers(t *testing.T) {
	dos := &fakeOneShotPlayer{playing: true, position: time.Duration(123) * time.Second / sampleRate}
	pc98 := &fakeOneShotPlayer{playing: true, position: time.Duration(456) * time.Second / sampleRate}
	ended := &fakeOneShotPlayer{playing: true, position: time.Second, endAfterRead: true}
	player := &Player{
		enabled:     true,
		players:     map[ID]oneShotPlayer{Missile: dos, Death: ended},
		pc98Players: map[Event]oneShotPlayer{Event("spell_hit"): pc98},
	}
	got, err := player.SnapshotOneShots()
	if err != nil {
		t.Fatal(err)
	}
	want := audiostate.Snapshot{Version: audiostate.CurrentVersion, Enabled: true, OneShots: []audiostate.OneShot{
		{Backend: audiostate.BackendDOSWAV, Key: "2", PositionFrames: 123},
		{Backend: audiostate.BackendPC98Speaker, Key: "spell_hit", PositionFrames: 456},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot=%+v want=%+v", got, want)
	}
}

func TestOneShotRestoreResumesExactPC98FrameAndFailsClosed(t *testing.T) {
	effect := &fakeOneShotPlayer{playing: true}
	player := &Player{enabled: true, pc98Players: map[Event]oneShotPlayer{Event("spell_hit"): effect}}
	snapshot := audiostate.Snapshot{Version: audiostate.CurrentVersion, Enabled: true, OneShots: []audiostate.OneShot{{
		Backend: audiostate.BackendPC98Speaker, Key: "spell_hit", PositionFrames: 12_345,
	}}}
	if err := player.RestoreOneShots(snapshot); err != nil {
		t.Fatal(err)
	}
	if !effect.playing || effect.playCount != 1 || durationToSampleFramesCeil(effect.position, sampleRate) != 12_345 {
		t.Fatalf("restored effect=%+v", effect)
	}

	effect.setPositionErr = errors.New("seek failed")
	if err := player.RestoreOneShots(snapshot); err == nil {
		t.Fatal("seek failure was accepted")
	}
	if effect.playing {
		t.Fatal("failed restore left an effect playing")
	}
}

func TestOneShotRestoreRejectsUnavailableBackendAssetBeforePlayback(t *testing.T) {
	effect := &fakeOneShotPlayer{playing: true}
	player := &Player{enabled: true, pc98Players: map[Event]oneShotPlayer{Event("spell_hit"): effect}}
	snapshot := audiostate.Snapshot{Version: audiostate.CurrentVersion, Enabled: true, OneShots: []audiostate.OneShot{{
		Backend: audiostate.BackendDOSWAV, Key: "2", PositionFrames: 1,
	}}}
	if err := player.RestoreOneShots(snapshot); err == nil {
		t.Fatal("unavailable DOS snapshot asset was accepted")
	}
	if effect.playCount != 0 {
		t.Fatal("unavailable asset started playback")
	}
	if effect.playing {
		t.Fatal("unavailable asset left a pre-load effect playing")
	}
}

func TestOneShotSnapshotRestoresMixedLoadedBackends(t *testing.T) {
	dos := &fakeOneShotPlayer{playing: true, position: time.Duration(12) * time.Second / sampleRate}
	pc98 := &fakeOneShotPlayer{playing: true, position: time.Duration(34) * time.Second / sampleRate}
	player := &Player{
		enabled:     true,
		players:     map[ID]oneShotPlayer{Missile: dos},
		pc98Players: map[Event]oneShotPlayer{Event("spell_hit"): pc98},
	}
	snapshot, err := player.SnapshotOneShots()
	if err != nil {
		t.Fatal(err)
	}
	if err := player.RestoreOneShots(snapshot); err != nil {
		t.Fatal(err)
	}
	if !dos.playing || !pc98.playing || durationToSampleFramesCeil(dos.position, sampleRate) != 12 || durationToSampleFramesCeil(pc98.position, sampleRate) != 34 {
		t.Fatalf("mixed restore dos=%+v pc98=%+v", dos, pc98)
	}
}

func TestDisabledOneShotSnapshotStopsEffectsWithoutResurrection(t *testing.T) {
	effect := &fakeOneShotPlayer{playing: true}
	player := &Player{enabled: true, players: map[ID]oneShotPlayer{Missile: effect}}
	snapshot := audiostate.Snapshot{Version: audiostate.CurrentVersion, Enabled: false}
	if err := player.RestoreOneShots(snapshot); err != nil {
		t.Fatal(err)
	}
	if player.enabled || effect.playing || effect.playCount != 0 {
		t.Fatalf("disabled restore player.enabled=%v effect=%+v", player.enabled, effect)
	}
	got, err := player.SnapshotOneShots()
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled || len(got.OneShots) != 0 {
		t.Fatalf("disabled snapshot=%+v", got)
	}
}
