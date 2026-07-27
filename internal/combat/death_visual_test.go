package combat

import (
	"testing"
	"time"
)

func TestDeathOverlayFrameRunsNineAlternatingCycles(t *testing.T) {
	for phase := int64(0); phase < 9; phase++ {
		frame, active := DeathOverlayFrame(time.Duration(phase) * DeathOverlayPhaseDuration)
		if !active || frame != uint8(phase%2) {
			t.Fatalf("phase=%d frame=%d active=%v", phase, frame, active)
		}
	}
	if _, active := DeathOverlayFrame(9 * DeathOverlayPhaseDuration); active {
		t.Fatal("death overlay remained active after nine cycles")
	}
}
