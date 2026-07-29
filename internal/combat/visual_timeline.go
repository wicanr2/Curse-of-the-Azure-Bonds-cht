package combat

import "time"

// VisualKind identifies choreography without coupling combat rules to a
// particular renderer or title asset pack.
type VisualKind uint8

const (
	VisualMelee VisualKind = iota + 1
	VisualMissile
	VisualMagicMissile
	VisualAreaSpell
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
	Effect      string
	ActorID     string
	TargetID    string
	From        TilePoint
	To          TilePoint
	Hit         bool
	Killed      bool
	Projectiles int
	Impacts     []VisualImpactTarget
}

// VisualImpactTarget describes one resolved target in an ordered area or chained
// effect. Single-target callers may keep using VisualEvent's legacy target
// fields; Impacts is the title-neutral extension used by effects such as an
// area spell which presents each affected combatant in sequence.
type VisualImpactTarget struct {
	TargetID string
	To       TilePoint
	Hit      bool
	Killed   bool
}

type VisualFrame struct {
	Phase       VisualPhase
	ImpactIndex int
	Progress    float64
	Done        bool
}

func (event VisualEvent) Duration() time.Duration {
	duration := VisualWindupDuration + VisualTravelDuration + VisualHandoffDuration
	for _, impact := range event.visualImpacts() {
		duration += VisualImpactDuration + VisualCommitDuration
		if impact.Killed {
			duration += 9 * DeathOverlayPhaseDuration
		}
	}
	return duration
}

// FrameAt converts elapsed wall-clock time into deterministic choreography.
// The renderer owns the clock; saved gameplay state remains clock-neutral.
func (event VisualEvent) FrameAt(elapsed time.Duration) VisualFrame {
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed < VisualWindupDuration {
		return visualFrame(VisualWindup, -1, elapsed, VisualWindupDuration)
	}
	elapsed -= VisualWindupDuration
	if elapsed < VisualTravelDuration {
		return visualFrame(VisualTravel, -1, elapsed, VisualTravelDuration)
	}
	elapsed -= VisualTravelDuration
	for index, impact := range event.visualImpacts() {
		if elapsed < VisualImpactDuration {
			return visualFrame(VisualImpact, index, elapsed, VisualImpactDuration)
		}
		elapsed -= VisualImpactDuration
		if elapsed < VisualCommitDuration {
			return visualFrame(VisualCommit, index, elapsed, VisualCommitDuration)
		}
		elapsed -= VisualCommitDuration
		if impact.Killed {
			deathDuration := 9 * DeathOverlayPhaseDuration
			if elapsed < deathDuration {
				return visualFrame(VisualDeath, index, elapsed, deathDuration)
			}
			elapsed -= deathDuration
		}
	}
	if elapsed < VisualHandoffDuration {
		return visualFrame(VisualHandoff, -1, elapsed, VisualHandoffDuration)
	}
	return VisualFrame{Phase: VisualHandoff, Progress: 1, Done: true}
}

// Impact returns the target associated with frame. During windup, travel and
// handoff no target impact is active.
func (event VisualEvent) Impact(frame VisualFrame) (VisualImpactTarget, bool) {
	return event.ImpactAt(frame.ImpactIndex)
}

func (event VisualEvent) ImpactAt(index int) (VisualImpactTarget, bool) {
	impacts := event.visualImpacts()
	if index < 0 || index >= len(impacts) {
		return VisualImpactTarget{}, false
	}
	return impacts[index], true
}

func (event VisualEvent) ImpactCount() int { return len(event.visualImpacts()) }

func (event VisualEvent) visualImpacts() []VisualImpactTarget {
	if event.Impacts != nil {
		return event.Impacts
	}
	return []VisualImpactTarget{{
		TargetID: event.TargetID,
		To:       event.To,
		Hit:      event.Hit,
		Killed:   event.Killed,
	}}
}

func visualFrame(phase VisualPhase, impactIndex int, elapsed, duration time.Duration) VisualFrame {
	return VisualFrame{
		Phase:       phase,
		ImpactIndex: impactIndex,
		Progress:    float64(elapsed) / float64(duration),
	}
}
