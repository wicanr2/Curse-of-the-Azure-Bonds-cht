package combat

import "time"

// VisualKind identifies choreography without coupling combat rules to a
// particular renderer or title asset pack.
type VisualKind uint8

const (
	VisualMelee VisualKind = iota + 1
	VisualMissile
	VisualMagicMissile
)

// VisualPhase is the player-visible ordering contract for one resolved action.
type VisualPhase uint8

const (
	VisualWindup VisualPhase = iota + 1
	VisualTravel
	VisualImpact
	VisualCommit
	VisualDeath
	VisualHandoff
)

const (
	VisualWindupDuration  = 140 * time.Millisecond
	VisualTravelDuration  = 220 * time.Millisecond
	VisualImpactDuration  = 140 * time.Millisecond
	VisualCommitDuration  = 120 * time.Millisecond
	VisualHandoffDuration = 100 * time.Millisecond
)

// VisualEvent is a renderer-neutral snapshot. Rule resolution may update the
// Battle immediately, but the frontend must present this ordered transaction
// before State advances to another actor.
type VisualEvent struct {
	Serial      uint64
	Kind        VisualKind
	ActorID     string
	TargetID    string
	From        TilePoint
	To          TilePoint
	Hit         bool
	Killed      bool
	Projectiles int
}

type VisualFrame struct {
	Phase    VisualPhase
	Progress float64
	Done     bool
}

func (event VisualEvent) Duration() time.Duration {
	duration := VisualWindupDuration + VisualTravelDuration + VisualImpactDuration +
		VisualCommitDuration + VisualHandoffDuration
	if event.Killed {
		duration += 9 * DeathOverlayPhaseDuration
	}
	return duration
}

// FrameAt converts elapsed wall-clock time into deterministic choreography.
// The renderer owns the clock; saved gameplay state remains clock-neutral.
func (event VisualEvent) FrameAt(elapsed time.Duration) VisualFrame {
	if elapsed < 0 {
		elapsed = 0
	}
	phases := []struct {
		phase    VisualPhase
		duration time.Duration
	}{
		{VisualWindup, VisualWindupDuration},
		{VisualTravel, VisualTravelDuration},
		{VisualImpact, VisualImpactDuration},
		{VisualCommit, VisualCommitDuration},
	}
	if event.Killed {
		phases = append(phases, struct {
			phase    VisualPhase
			duration time.Duration
		}{VisualDeath, 9 * DeathOverlayPhaseDuration})
	}
	phases = append(phases, struct {
		phase    VisualPhase
		duration time.Duration
	}{VisualHandoff, VisualHandoffDuration})
	for _, entry := range phases {
		if elapsed < entry.duration {
			return VisualFrame{
				Phase:    entry.phase,
				Progress: float64(elapsed) / float64(entry.duration),
			}
		}
		elapsed -= entry.duration
	}
	return VisualFrame{Phase: VisualHandoff, Progress: 1, Done: true}
}
