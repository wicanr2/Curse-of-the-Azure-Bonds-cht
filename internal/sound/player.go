package sound

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

const sampleRate = 44100

// Player is an optional renderer-side sound adapter. It loads only the WAV
// assets proven by the reference seg044 resource table and ignores unmapped
// sound IDs just as the reference engine does.
type Player struct {
	context     *audio.Context
	players     map[ID]*audio.Player
	pc98Players map[Event]*audio.Player
	enabled     bool
	musicPlayer *audio.Player
	musicStream *pc98music.TrackPCMStream
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
	players := make(map[Event]*audio.Player)
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
	player := &Player{context: context, players: make(map[ID]*audio.Player), enabled: true}
	var loadErrors []error
	for _, id := range []ID{Missile, MagicHit, Death, Sound5, Hit, Miss, Step, Sound10, Start} {
		name, _ := AssetName(id)
		data, err := os.ReadFile(filepath.Join(assetDir, name))
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}
		stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
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
	player, err := p.context.NewPlayer(stream)
	if err != nil {
		_ = stream.Close()
		return err
	}
	p.musicStream = stream
	p.musicPlayer = player
	if p.enabled {
		player.Play()
	}
	return nil
}

// StopMusic stops and releases the active PC-98 stream.
func (p *Player) StopMusic() {
	if p == nil {
		return
	}
	if p.musicPlayer != nil {
		p.musicPlayer.Pause()
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
