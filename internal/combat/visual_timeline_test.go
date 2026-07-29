package combat

import (
	"testing"
	"time"
)

func TestVisualTimelineOrdersActionAndOptionalDeath(t *testing.T) {
	event := VisualEvent{Kind: VisualMissile, Killed: true}
	checks := []struct {
		at    time.Duration
		phase VisualPhase
	}{
		{0, VisualWindup},
		{VisualWindupDuration, VisualTravel},
		{VisualWindupDuration + VisualTravelDuration, VisualImpact},
		{VisualWindupDuration + VisualTravelDuration + VisualImpactDuration, VisualCommit},
		{VisualWindupDuration + VisualTravelDuration + VisualImpactDuration + VisualCommitDuration, VisualDeath},
		{event.Duration() - VisualHandoffDuration, VisualHandoff},
	}
	for _, check := range checks {
		if got := event.FrameAt(check.at); got.Phase != check.phase || got.Done {
			t.Fatalf("FrameAt(%s)=%+v, want phase %v", check.at, got, check.phase)
		}
	}
	if got := event.FrameAt(event.Duration()); !got.Done {
		t.Fatalf("terminal frame=%+v", got)
	}
}

func TestVisualTimelineSkipsDeathForSurvivor(t *testing.T) {
	event := VisualEvent{Kind: VisualMagicMissile}
	deathStart := VisualWindupDuration + VisualTravelDuration +
		VisualImpactDuration + VisualCommitDuration
	if got := event.FrameAt(deathStart); got.Phase != VisualHandoff {
		t.Fatalf("nonlethal frame=%+v, want handoff", got)
	}
}
