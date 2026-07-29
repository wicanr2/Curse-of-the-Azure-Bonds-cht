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

func TestVisualTimelinePresentsAreaImpactsAndDeathsInOrder(t *testing.T) {
	event := VisualEvent{
		Kind: VisualMagicMissile,
		Impacts: []VisualImpactTarget{
			{TargetID: "orc-1", To: TilePoint{X: 4, Y: 2}, Hit: true},
			{TargetID: "orc-2", To: TilePoint{X: 5, Y: 3}, Hit: true, Killed: true},
			{TargetID: "orc-3", To: TilePoint{X: 6, Y: 4}, Hit: true},
		},
	}
	firstImpact := VisualWindupDuration + VisualTravelDuration
	secondImpact := firstImpact + VisualImpactDuration + VisualCommitDuration
	secondDeath := secondImpact + VisualImpactDuration + VisualCommitDuration
	thirdImpact := secondDeath + 9*DeathOverlayPhaseDuration

	checks := []struct {
		at          time.Duration
		phase       VisualPhase
		impactIndex int
		targetID    string
	}{
		{firstImpact, VisualImpact, 0, "orc-1"},
		{secondImpact, VisualImpact, 1, "orc-2"},
		{secondDeath, VisualDeath, 1, "orc-2"},
		{thirdImpact, VisualImpact, 2, "orc-3"},
	}
	for _, check := range checks {
		frame := event.FrameAt(check.at)
		if frame.Phase != check.phase || frame.ImpactIndex != check.impactIndex {
			t.Fatalf("FrameAt(%s)=%+v, want phase %v impact %d", check.at, frame, check.phase, check.impactIndex)
		}
		impact, ok := event.Impact(frame)
		if !ok || impact.TargetID != check.targetID {
			t.Fatalf("Impact(%+v)=(%+v,%v), want %q", frame, impact, ok, check.targetID)
		}
	}
	if got := event.FrameAt(event.Duration() - VisualHandoffDuration); got.Phase != VisualHandoff || got.ImpactIndex != -1 {
		t.Fatalf("handoff frame=%+v", got)
	}
}

func TestVisualTimelineLegacyTargetIsOneImpact(t *testing.T) {
	event := VisualEvent{
		TargetID: "ogre",
		To:       TilePoint{X: 5, Y: 4},
		Hit:      true,
		Killed:   true,
	}
	frame := event.FrameAt(VisualWindupDuration + VisualTravelDuration)
	impact, ok := event.Impact(frame)
	if !ok || impact.TargetID != event.TargetID || impact.To != event.To ||
		impact.Hit != event.Hit || impact.Killed != event.Killed {
		t.Fatalf("legacy impact=(%+v,%v), event=%+v", impact, ok, event)
	}
}

func TestVisualTimelineInterleavesLeadingImpactAndLineSegments(t *testing.T) {
	event := VisualEvent{
		Kind: VisualLineSpell,
		Impacts: []VisualImpactTarget{
			{TargetID: "target", To: TilePoint{X: 2, Y: 1}, Hit: true},
			{TargetID: "far", To: TilePoint{X: 4, Y: 1}, Hit: true},
		},
		TravelImpacts: 1,
		Segments: []VisualPathSegment{
			{From: TilePoint{X: 2, Y: 1}, To: TilePoint{X: 4, Y: 1}, HasImpact: true, ImpactIndex: 1},
			{From: TilePoint{X: 4, Y: 1}, To: TilePoint{X: 6, Y: 1}},
		},
	}
	leadingImpact := VisualWindupDuration + VisualTravelDuration
	firstSegment := leadingImpact + VisualImpactDuration + VisualCommitDuration
	secondImpact := firstSegment + VisualTravelDuration
	secondSegment := secondImpact + VisualImpactDuration + VisualCommitDuration
	checks := []struct {
		at           time.Duration
		phase        VisualPhase
		impactIndex  int
		segmentIndex int
		resolved     int
	}{
		{leadingImpact, VisualImpact, 0, -1, 0},
		{firstSegment, VisualSegmentTravel, -1, 0, 1},
		{secondImpact, VisualImpact, 1, 0, 1},
		{secondSegment, VisualSegmentTravel, -1, 1, 2},
	}
	for _, check := range checks {
		frame := event.FrameAt(check.at)
		if frame.Phase != check.phase || frame.ImpactIndex != check.impactIndex ||
			frame.SegmentIndex != check.segmentIndex || frame.ResolvedImpacts != check.resolved {
			t.Fatalf("FrameAt(%s)=%+v", check.at, frame)
		}
	}
}
