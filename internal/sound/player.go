package sound

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

const sampleRate = 44100
const musicBufferBytes = sampleRate * 4 / 4 // 250 ms of stereo s16 PCM.

// Player is an optional renderer-side sound adapter. It loads only the OGG
// assets proven by the reference seg044 resource table and ignores unmapped
// sound IDs just as the reference engine does.
type Player struct {
	context     *audio.Context
	players     map[ID]oneShotPlayer
	pc98Players map[Event]oneShotPlayer
	enabled     bool
	musicPlayer *audio.Player
	musicStream *pc98music.TrackPCMStream
}

type oneShotPlayer interface {
	Rewind() error
	Play()
	Pause()
	IsPlaying() bool
	Position() time.Duration
	SetPosition(time.Duration) error
}

// LoadPC98Effects imports GAME.EXE's exact SOUNDFX program and prepares
// software-speaker one-shots. The selected CPU clock is explicit because the
// original busy-loop pitch changes with the PC-98 machine profile.
func (p *Player) LoadPC98Effects(game []byte, clockHz uint64) error {
	if p == nil || p.context == nil {
		return fmt.Errorf("sound player is unavailable")
	}
	effects, err := pc98sfx.Import(game)
	if err != nil {
		return err
	}
	players := make(map[Event]oneShotPlayer)
	for _, effect := range effects {
		if effect.Event == "" || effect.NoOp {
			continue
		}
		mono, renderErr := pc98sfx.RenderPCM(
			effect,
			pc98sfx.V30PrefetchedProfile(clockHz),
			sampleRate,
		)
		if renderErr != nil {
			return fmt.Errorf("%s: %w", effect.Symbol, renderErr)
		}
		stream := stereoPCM(mono)
		if len(stream) == 0 {
			continue
		}
		player, playerErr := p.context.NewPlayer(bytes.NewReader(stream))
		if playerErr != nil {
			return fmt.Errorf("%s: %w", effect.Symbol, playerErr)
		}
		players[Event(effect.Event)] = player
	}
	p.pc98Players = players
	return nil
}

func stereoPCM(mono []int16) []byte {
	stereo := make([]byte, len(mono)*4)
	for index, sample := range mono {
		stereo[index*4] = byte(sample)
		stereo[index*4+1] = byte(uint16(sample) >> 8)
		stereo[index*4+2] = byte(sample)
		stereo[index*4+3] = byte(uint16(sample) >> 8)
	}
	return stereo
}

// Load creates a player from the extracted asset directory.
func Load(assetDir string) (*Player, error) {
	context := audio.NewContext(sampleRate)
	player := &Player{context: context, players: make(map[ID]oneShotPlayer), enabled: true}
	var loadErrors []error
	for _, id := range []ID{Missile, MagicHit, Death, Sound5, Hit, Miss, Step, Sound10, Start} {
		name, _ := AssetName(id)
		data, err := os.ReadFile(filepath.Join(assetDir, name))
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}
		stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("%s: %w", name, err))
			continue
		}
		p, err := context.NewPlayer(stream)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("%s: %w", name, err))
			continue
		}
		player.players[id] = p
	}
	return player, errors.Join(loadErrors...)
}

// PlayOGGTrack replaces the current background stream with a pre-rendered
// OGG asset. Track IDs are stable game-pack IDs and map directly to filenames.
func (p *Player) PlayOGGTrack(assetDir, trackID string) error {
	if p == nil || p.context == nil {
		return fmt.Errorf("sound player is unavailable")
	}
	if trackID == "" || filepath.Base(trackID) != trackID {
		return fmt.Errorf("invalid music track ID %q", trackID)
	}
	data, err := os.ReadFile(filepath.Join(assetDir, trackID+".ogg"))
	if err != nil {
		return err
	}
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s.ogg: %w", trackID, err)
	}
	if stream.Length() <= 0 {
		return fmt.Errorf("%s.ogg has no decoded audio", trackID)
	}
	p.StopMusic()
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := p.context.NewPlayer(loop)
	if err != nil {
		return err
	}
	player.SetBufferSize(musicBufferBytes)
	p.musicPlayer = player
	if p.enabled {
		player.Play()
	}
	return nil
}

// PlayPC98Track replaces the current background stream with one selector
// rendered from the user's exact local MSCDRV.EXE.
func (p *Player) PlayPC98Track(driver []byte, selector int) error {
	if p == nil || p.context == nil {
		return fmt.Errorf("sound player is unavailable")
	}
	p.StopMusic()
	stream, err := pc98music.NewGameTrackPCMStream(driver, selector, sampleRate)
	if err != nil {
		return err
	}
	return p.installPC98MusicStream(stream)
}

func (p *Player) installPC98MusicStream(stream *pc98music.TrackPCMStream) error {
	player, err := p.context.NewPlayer(stream)
	if err != nil {
		_ = stream.Close()
		return err
	}
	player.SetBufferSize(musicBufferBytes)
	p.musicStream = stream
	p.musicPlayer = player
	if p.enabled {
		player.Play()
	}
	return nil
}

// SnapshotPC98Music captures from the sample frame reported as audible by
// Ebiten, not from the decoder's read-ahead position.
func (p *Player) SnapshotPC98Music() (*pc98music.TrackPCMStreamSnapshot, error) {
	if p == nil || p.musicPlayer == nil || p.musicStream == nil {
		return nil, nil
	}
	position := p.musicPlayer.Position()
	frames := durationToSampleFramesCeil(position, sampleRate)
	snapshot, err := p.musicStream.SnapshotAtFrame(frames)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func durationToSampleFramesCeil(position time.Duration, rate uint64) uint64 {
	if position <= 0 || rate == 0 {
		return 0
	}
	nanos := uint64(position)
	second := uint64(time.Second)
	whole := nanos / second * rate
	remainder := nanos % second
	return whole + (remainder*rate+second-1)/second
}

func sampleFramesToDurationFloor(frames, rate uint64) (time.Duration, error) {
	if frames == 0 {
		return 0, nil
	}
	if rate == 0 {
		return 0, fmt.Errorf("sample rate is zero")
	}
	seconds := frames / rate
	if seconds > uint64((time.Duration(1<<63-1))/time.Second) {
		return 0, fmt.Errorf("sample frame position %d overflows duration", frames)
	}
	nanos := seconds * uint64(time.Second)
	remainder := frames % rate
	if remainder != 0 {
		if remainder > ^uint64(0)/uint64(time.Second) {
			return 0, fmt.Errorf("sample rate %d cannot be converted safely", rate)
		}
		fraction := remainder * uint64(time.Second) / rate
		if nanos > uint64(1<<63-1)-fraction {
			return 0, fmt.Errorf("sample frame position %d overflows duration", frames)
		}
		nanos += fraction
	}
	return time.Duration(nanos), nil
}

// SnapshotOneShots records only players that are still active at the audible
// position. A second IsPlaying check prevents an effect that ended while its
// position was sampled from being resurrected by a later load.
func (p *Player) SnapshotOneShots() (audiostate.Snapshot, error) {
	snapshot := audiostate.Snapshot{Version: audiostate.CurrentVersion}
	if p == nil {
		return snapshot, fmt.Errorf("sound player is unavailable")
	}
	snapshot.Enabled = p.enabled
	if !p.enabled {
		return snapshot, snapshot.Validate()
	}
	for id, player := range p.players {
		if player.IsPlaying() {
			position := player.Position()
			if player.IsPlaying() {
				snapshot.OneShots = append(snapshot.OneShots, audiostate.OneShot{
					Backend: audiostate.BackendDOSWAV, Key: strconv.Itoa(int(id)),
					PositionFrames: durationToSampleFramesCeil(position, sampleRate),
				})
			}
		}
	}
	for event, player := range p.pc98Players {
		if player.IsPlaying() {
			position := player.Position()
			if player.IsPlaying() {
				snapshot.OneShots = append(snapshot.OneShots, audiostate.OneShot{
					Backend: audiostate.BackendPC98Speaker, Key: string(event),
					PositionFrames: durationToSampleFramesCeil(position, sampleRate),
				})
			}
		}
	}
	sort.Slice(snapshot.OneShots, func(i, j int) bool {
		if snapshot.OneShots[i].Backend != snapshot.OneShots[j].Backend {
			return snapshot.OneShots[i].Backend < snapshot.OneShots[j].Backend
		}
		return snapshot.OneShots[i].Key < snapshot.OneShots[j].Key
	})
	return snapshot, snapshot.Validate()
}

// RestoreOneShots fails closed: all identities and positions are resolved
// before playback begins, and any seek failure leaves every one-shot stopped.
func (p *Player) RestoreOneShots(snapshot audiostate.Snapshot) error {
	if p == nil {
		return fmt.Errorf("sound player is unavailable")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	p.stopOneShots()
	type resolvedShot struct {
		player   oneShotPlayer
		position time.Duration
	}
	resolved := make([]resolvedShot, 0, len(snapshot.OneShots))
	for index, shot := range snapshot.OneShots {
		position, err := sampleFramesToDurationFloor(shot.PositionFrames, sampleRate)
		if err != nil {
			return fmt.Errorf("one-shot %d position: %w", index, err)
		}
		var player oneShotPlayer
		switch shot.Backend {
		case audiostate.BackendDOSWAV:
			value, err := strconv.ParseUint(shot.Key, 10, 8)
			if err != nil {
				return fmt.Errorf("one-shot %d DOS selector %q: %w", index, shot.Key, err)
			}
			player = p.players[ID(value)]
		case audiostate.BackendPC98Speaker:
			if p.pc98Players == nil {
				return fmt.Errorf("one-shot %d PC-98 backend is not loaded", index)
			}
			player = p.pc98Players[Event(shot.Key)]
		}
		if player == nil {
			return fmt.Errorf("one-shot %d asset %s/%s is unavailable", index, shot.Backend, shot.Key)
		}
		resolved = append(resolved, resolvedShot{player: player, position: position})
	}
	p.enabled = snapshot.Enabled
	for index, shot := range resolved {
		if err := shot.player.SetPosition(shot.position); err != nil {
			p.stopOneShots()
			return fmt.Errorf("seek one-shot %d: %w", index, err)
		}
	}
	for _, shot := range resolved {
		shot.player.Play()
	}
	if p.musicPlayer != nil {
		if p.enabled {
			p.musicPlayer.Play()
		} else {
			p.musicPlayer.Pause()
		}
	}
	return nil
}

// StopOneShots prevents sounds from the pre-load state leaking across a load.
func (p *Player) StopOneShots() {
	if p != nil {
		p.stopOneShots()
	}
}

func (p *Player) RestorePC98Track(driver []byte, snapshot pc98music.TrackPCMStreamSnapshot) error {
	if p == nil || p.context == nil {
		return fmt.Errorf("sound player is unavailable")
	}
	p.StopMusic()
	stream, err := pc98music.RestoreGameTrackPCMStream(driver, snapshot)
	if err != nil {
		return err
	}
	return p.installPC98MusicStream(stream)
}

// StopMusic stops and releases the active PC-98 stream.
func (p *Player) StopMusic() {
	if p == nil {
		return
	}
	if p.musicPlayer != nil {
		p.musicPlayer.Pause()
		_ = p.musicPlayer.Close()
		p.musicPlayer = nil
	}
	if p.musicStream != nil {
		_ = p.musicStream.Close()
		p.musicStream = nil
	}
}

// SetEnabled mirrors reference SetSound. Disabled playback leaves state and
// callers untouched.
func (p *Player) SetEnabled(enabled bool) {
	if p != nil {
		p.enabled = enabled
		if !enabled {
			p.stopOneShots()
		}
		if p.musicPlayer != nil {
			if enabled {
				p.musicPlayer.Play()
			} else {
				p.musicPlayer.Pause()
			}
		}
	}
}

// Play starts a one-shot reference sample. Stop stops all currently loaded
// samples; no-op and unknown IDs are intentionally harmless.
func (p *Player) Play(id ID) {
	if p == nil || !p.enabled {
		return
	}
	if id == Stop {
		p.stopOneShots()
		return
	}
	player, ok := p.players[id]
	if !ok {
		return
	}
	_ = player.Rewind()
	player.Play()
}

// PlayEvent routes a semantic gameplay intent to the configured PC-98
// software-speaker effect, falling back to the DOS WAV resource adapter.
func (p *Player) PlayEvent(event Event) {
	if p == nil || !p.enabled {
		return
	}
	if event == "stop" {
		p.stopOneShots()
		return
	}
	if p.pc98Players != nil {
		player, ok := p.pc98Players[event]
		if !ok {
			return
		}
		_ = player.Rewind()
		player.Play()
		return
	}
	if id, ok := DOSID(event); ok {
		p.Play(id)
	}
}

func (p *Player) stopOneShots() {
	for _, player := range p.players {
		player.Pause()
	}
	for _, player := range p.pc98Players {
		player.Pause()
	}
}
