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
)

const sampleRate = 44100

// Player is an optional renderer-side sound adapter. It loads only the WAV
// assets proven by the reference seg044 resource table and ignores unmapped
// sound IDs just as the reference engine does.
type Player struct {
	context     *audio.Context
	players     map[ID]*audio.Player
	enabled     bool
	musicPlayer *audio.Player
	musicStream *pc98music.TrackPCMStream
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
		for _, player := range p.players {
			player.Pause()
		}
		return
	}
	player, ok := p.players[id]
	if !ok {
		return
	}
	_ = player.Rewind()
	player.Play()
}
