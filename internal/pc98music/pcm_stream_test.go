//go:build cgo

package pc98music

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
)

func TestTrackPCMStreamProducesDeterministicStereo(t *testing.T) {
	render := func() ([]byte, string) {
		playback, initial := syntheticPlayback(t)
		renderer := newYM2203EventRenderer(nil)
		stream, err := newTrackPCMStream(playback, renderer, initial, 44_100)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		output := make([]byte, 4096)
		if _, err := io.ReadFull(stream, output); err != nil {
			t.Fatal(err)
		}
		nonzero := false
		for index := 0; index < len(output); index += 4 {
			if output[index] != 0 || output[index+1] != 0 {
				nonzero = true
			}
			if output[index] != output[index+2] ||
				output[index+1] != output[index+3] {
				t.Fatalf("frame %d is not dual mono", index/4)
			}
		}
		if !nonzero {
			t.Fatal("rendered track is silent")
		}
		sum := sha256.Sum256(output)
		return output, hex.EncodeToString(sum[:])
	}
	first, firstHash := render()
	second, secondHash := render()
	if firstHash != secondHash || len(first) != len(second) {
		t.Fatalf("render changed: %s != %s", firstHash, secondHash)
	}
	const expected = "4dba2117508462b1f49cb7f3c4d7b935519629655d454e8224c8fd604d263677"
	if firstHash != expected {
		t.Fatalf("PCM SHA-256=%s, want %s", firstHash, expected)
	}
}

func TestGameTrackPCMStreamStartsWithExactMSCPLAYSilence(t *testing.T) {
	playback, initial := syntheticPlayback(t)
	renderer := newYM2203EventRenderer(nil)
	stream, err := newTrackPCMStream(playback, renderer, initial, 44_100)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	stream.prependSilence(GameMusicStartDelay, 44_100)

	const silentFrames = 35_280
	silence := make([]byte, silentFrames*4)
	if _, err := io.ReadFull(stream, silence); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(silence, make([]byte, len(silence))) {
		t.Fatal("GAME.EXE MSCPLAY 800ms pre-roll is not silent")
	}
	if playback.tick != 0 {
		t.Fatalf("playback advanced to tick %d during transition silence", playback.tick)
	}

	audible := make([]byte, 4096)
	if _, err := io.ReadFull(stream, audible); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(audible, make([]byte, len(audible))) {
		return
	}
	t.Fatal("track remained silent after the exact 800ms transition")
}
