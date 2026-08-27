package sound

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

func TestReferenceSoundAssetMapping(t *testing.T) {
	for _, test := range []struct {
		id   ID
		want string
	}{
		{Missile, "missle.ogg"}, {MagicHit, "magic_hit.ogg"}, {Death, "death.ogg"},
		{Sound5, "sound_5.ogg"}, {Hit, "hit.ogg"}, {Miss, "miss.ogg"},
		{Step, "step.ogg"}, {Sound10, "sound_10.ogg"}, {Start, "start_sound.ogg"},
	} {
		got, ok := AssetName(test.id)
		if !ok || got != test.want {
			t.Fatalf("sound %d asset=%q,%v want %q,true", test.id, got, ok, test.want)
		}
	}
	if _, ok := AssetName(Stop); ok {
		t.Fatal("stop must not resolve to an OGG asset")
	}
}

func TestDOSIDKeepsPlatformMappingSeparateFromSemanticEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event Event
		want  ID
	}{
		{"cast", Missile},
		{"arrow", Missile},
		{"miss", Miss},
		{"spell_hit", MagicHit},
		{"dead", Death},
		{"whistle", Sound5},
		{"hit", Hit},
		{"lightning", Lightning},
		{"swish", Miss},
		{"step", Step},
		{"fireball", Sound10}, {"combat", Combat}, {"crash", Crash},
		{"overture", Start},
	}
	for _, test := range tests {
		got, ok := DOSID(test.event)
		if !ok || got != test.want {
			t.Errorf("DOSID(%q)=(%d,%v), want (%d,true)", test.event, got, ok, test.want)
		}
	}
	for _, id := range []ID{Combat, Crash} {
		if _, ok := AssetName(id); ok {
			t.Fatalf("DOS selector %d has no extracted reference asset", id)
		}
	}
}

func TestStereoPCMDuplicatesSignedMonoSamples(t *testing.T) {
	got := stereoPCM([]int16{-32768, -1, 0, 32767})
	want := []byte{
		0x00, 0x80, 0x00, 0x80,
		0xFF, 0xFF, 0xFF, 0xFF,
		0x00, 0x00, 0x00, 0x00,
		0xFF, 0x7F, 0xFF, 0x7F,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("stereo PCM=% X, want % X", got, want)
	}
}

func TestReferenceOGGAssetsDecode(t *testing.T) {
	for _, id := range []ID{Missile, MagicHit, Death, Sound5, Hit, Miss, Step, Sound10, Start} {
		name, _ := AssetName(id)
		data, err := os.ReadFile(filepath.Join("..", "..", "assets", "audio", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if _, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data)); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
	}
}

func TestMusicOGGAssetsDecode(t *testing.T) {
	for selector := 1; selector <= 12; selector++ {
		name := fmt.Sprintf("pc98-bgm-selector-%02x.ogg", selector)
		data, err := os.ReadFile(filepath.Join("..", "..", "assets", "audio", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		if stream.Length() <= 0 {
			t.Fatalf("%s decoded to an empty stream", name)
		}
	}
}
