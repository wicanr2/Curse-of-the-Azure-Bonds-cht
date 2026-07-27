package sound

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

// Player is an optional renderer-side sound adapter. It loads only the WAV
// assets proven by the reference seg044 resource table and ignores unmapped
// sound IDs just as the reference engine does.
type Player struct {
	context *audio.Context
	players map[ID]*audio.Player
	enabled bool
}

// Load creates a player from the extracted asset directory.
func Load(assetDir string) (*Player, error) {
	context := audio.NewContext(sampleRate)
	player := &Player{context: context, players: make(map[ID]*audio.Player), enabled: true}
	for _, id := range []ID{Missile, MagicHit, Death, Sound5, Hit, Miss, Step, Sound10, Start} {
		name, _ := AssetName(id)
		data, err := os.ReadFile(filepath.Join(assetDir, name))
		if err != nil {
			return nil, err
		}
		stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		p, err := context.NewPlayer(stream)
		if err != nil {
			return nil, err
		}
		player.players[id] = p
	}
	return player, nil
}

// SetEnabled mirrors reference SetSound. Disabled playback leaves state and
// callers untouched.
func (p *Player) SetEnabled(enabled bool) {
	if p != nil {
		p.enabled = enabled
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
