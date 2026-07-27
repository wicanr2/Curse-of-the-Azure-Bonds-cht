package sound

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

func TestReferenceSoundAssetMapping(t *testing.T) {
	for _, test := range []struct {
		id   ID
		want string
	}{
		{Missile, "missle.wav"}, {MagicHit, "magic_hit.wav"}, {Death, "death.wav"},
		{Sound5, "sound_5.wav"}, {Hit, "hit.wav"}, {Miss, "miss.wav"},
		{Step, "step.wav"}, {Sound10, "sound_10.wav"}, {Start, "start_sound.wav"},
	} {
		got, ok := AssetName(test.id)
		if !ok || got != test.want {
			t.Fatalf("sound %d asset=%q,%v want %q,true", test.id, got, ok, test.want)
		}
	}
	if _, ok := AssetName(Stop); ok {
		t.Fatal("stop must not resolve to a WAV asset")
	}
}

func TestReferenceWAVAssetsDecode(t *testing.T) {
	for _, id := range []ID{Missile, MagicHit, Death, Sound5, Hit, Miss, Step, Sound10, Start} {
		name, _ := AssetName(id)
		data, err := os.ReadFile(filepath.Join("..", "..", "assets", "audio", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if _, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data)); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
	}
}
