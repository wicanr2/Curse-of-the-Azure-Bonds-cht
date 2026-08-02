package sound

import (
	"testing"
	"time"
)

func TestDurationToSampleFramesCeilInvertsEbitenPositionRounding(t *testing.T) {
	for _, frames := range []uint64{0, 1, 2, 127, 44_099, 44_100, 1_234_567} {
		position := time.Duration(frames) * time.Second / sampleRate
		if got := durationToSampleFramesCeil(position, sampleRate); got != frames {
			t.Fatalf("frames=%d position=%s restored=%d", frames, position, got)
		}
	}
}
