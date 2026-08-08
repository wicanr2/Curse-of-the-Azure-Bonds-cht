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
	VisualLineSpell
	VisualTwinkle
)

// VisualPhase is the player-visible ordering contract for one resolved action.
type VisualPhase uint8

const (
	VisualWindup VisualPhase = iota + 1
	VisualTravel
	VisualSegmentTravel
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
	// VisualTwinkleDuration is the exact PC-98 delay budget at GAMESPEED=4:
	// (4+1) outer passes * 4 frames * (4*18ms). Drawing overhead is excluded.
	VisualTwinkleDuration = 1440 * time.Millisecond
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
	// TravelImpacts is the number of ordered impacts presented immediately
	// after the primary From→To travel. Segments then interleave their own
	// travel with optional impacts. A nil Segments slice preserves the legacy
	// behavior in which every impact follows primary travel.
	TravelImpacts int
	Segments      []VisualPathSegment
	// PersistentAreaID associates a reveal animation with the already
	// resolved rules object. Renderers can hide only that new area during
	// travel without suppressing older overlapping effects.
	PersistentAreaID uint64
}

// VisualPathSegment is one ordered continuation or reflected part of an
// effect path. HasImpact binds the segment end to Impacts[ImpactIndex].
type VisualPathSegment struct {
	From        TilePoint
	To          TilePoint
	HasImpact   bool
	ImpactIndex int
}

// VisualImpactTarget describes one resolved target in an ordered area or chained
// effect. Single-target callers may keep using VisualEvent's legacy target
// fields; Impacts is the title-neutral extension used by effects such as an
// area spell which presents each affected combatant in sequence.
type VisualImpactTarget struct {
	TargetID  string
	To        TilePoint
	Hit       bool
	Killed    bool
	Damage    int
	Saved     bool
	Resisted  bool
	Protected bool
}

type VisualFrame struct {
	Phase        VisualPhase
	ImpactIndex  int
	SegmentIndex int
	// ResolvedImpacts counts impact transactions whose commit/death phases
	// have fully completed before this frame.
	ResolvedImpacts int
	Progress        float64
	Done            bool
}

func (event VisualEvent) Duration() time.Duration {
	if event.Kind == VisualTwinkle {
		return time.Duration(len(event.visualImpacts()))*VisualTwinkleDuration + VisualHandoffDuration
	}
	duration := VisualWindupDuration + VisualTravelDuration + VisualHandoffDuration
	if event.Segments == nil {
		for _, impact := range event.visualImpacts() {
			duration += visualImpactDuration(impact)
		}
		return duration
	}
	impacts := event.visualImpacts()
	for index := 0; index < event.leadingImpactCount(); index++ {
		duration += visualImpactDuration(impacts[index])
	}
	for _, segment := range event.Segments {
		duration += VisualTravelDuration
		if segment.HasImpact && segment.ImpactIndex >= event.leadingImpactCount() &&
			segment.ImpactIndex < len(impacts) {
			duration += visualImpactDuration(impacts[segment.ImpactIndex])
		}
	}
	return duration
}

// FrameAt converts an elapsed timeline position into deterministic
// choreography. The game State commits that position so save/load can resume
// the same frame; renderers only provide monotonic clock deltas.
func (event VisualEvent) FrameAt(elapsed time.Duration) VisualFrame {
	if elapsed < 0 {
		elapsed = 0
	}
	if event.Kind == VisualTwinkle {
		impacts := event.visualImpacts()
		for index := range impacts {
			if elapsed < VisualTwinkleDuration {
				return visualFrame(VisualImpact, index, -1, index, elapsed, VisualTwinkleDuration)
			}
			elapsed -= VisualTwinkleDuration
		}
		if elapsed < VisualHandoffDuration {
			return visualFrame(VisualHandoff, -1, -1, len(impacts), elapsed, VisualHandoffDuration)
		}
		return VisualFrame{Phase: VisualHandoff, ImpactIndex: -1, SegmentIndex: -1, ResolvedImpacts: len(impacts), Progress: 1, Done: true}
	}
	if elapsed < VisualWindupDuration {
		return visualFrame(VisualWindup, -1, -1, 0, elapsed, VisualWindupDuration)
	}
	elapsed -= VisualWindupDuration
	if elapsed < VisualTravelDuration {
		return visualFrame(VisualTravel, -1, -1, 0, elapsed, VisualTravelDuration)
	}
	elapsed -= VisualTravelDuration
	impacts := event.visualImpacts()
	if event.Segments == nil {
		for index, impact := range impacts {
			if frame, remaining, ok := impactFrameAt(index, -1, index, impact, elapsed); ok {
				return frame
			} else {
				elapsed = remaining
			}
		}
		if elapsed < VisualHandoffDuration {
			return visualFrame(VisualHandoff, -1, -1, len(impacts), elapsed, VisualHandoffDuration)
		}
		return VisualFrame{Phase: VisualHandoff, ImpactIndex: -1, SegmentIndex: -1, ResolvedImpacts: len(impacts), Progress: 1, Done: true}
	}

	resolved := 0
	for index := 0; index < event.leadingImpactCount(); index++ {
		if frame, remaining, ok := impactFrameAt(index, -1, resolved, impacts[index], elapsed); ok {
			return frame
		} else {
			elapsed = remaining
			resolved++
		}
	}
	for segmentIndex, segment := range event.Segments {
		if elapsed < VisualTravelDuration {
			return visualFrame(VisualSegmentTravel, -1, segmentIndex, resolved, elapsed, VisualTravelDuration)
		}
		elapsed -= VisualTravelDuration
		if segment.HasImpact && segment.ImpactIndex >= resolved && segment.ImpactIndex < len(impacts) {
			if frame, remaining, ok := impactFrameAt(segment.ImpactIndex, segmentIndex, resolved, impacts[segment.ImpactIndex], elapsed); ok {
				return frame
			} else {
				elapsed = remaining
				resolved++
			}
		}
	}
	if elapsed < VisualHandoffDuration {
		return visualFrame(VisualHandoff, -1, -1, resolved, elapsed, VisualHandoffDuration)
	}
	return VisualFrame{Phase: VisualHandoff, ImpactIndex: -1, SegmentIndex: -1, ResolvedImpacts: resolved, Progress: 1, Done: true}
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

func (event VisualEvent) leadingImpactCount() int {
	count := event.TravelImpacts
	if count < 0 {
		return 0
	}
	if count > len(event.visualImpacts()) {
		return len(event.visualImpacts())
	}
	return count
}

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

func visualImpactDuration(impact VisualImpactTarget) time.Duration {
	duration := VisualImpactDuration + VisualCommitDuration
	if impact.Killed {
		duration += 9 * DeathOverlayPhaseDuration
	}
	return duration
}

func impactFrameAt(impactIndex, segmentIndex, resolved int, impact VisualImpactTarget, elapsed time.Duration) (VisualFrame, time.Duration, bool) {
	if elapsed < VisualImpactDuration {
		return visualFrame(VisualImpact, impactIndex, segmentIndex, resolved, elapsed, VisualImpactDuration), elapsed, true
	}
	elapsed -= VisualImpactDuration
	if elapsed < VisualCommitDuration {
		return visualFrame(VisualCommit, impactIndex, segmentIndex, resolved, elapsed, VisualCommitDuration), elapsed, true
	}
	elapsed -= VisualCommitDuration
	if impact.Killed {
		deathDuration := 9 * DeathOverlayPhaseDuration
		if elapsed < deathDuration {
			return visualFrame(VisualDeath, impactIndex, segmentIndex, resolved, elapsed, deathDuration), elapsed, true
		}
		elapsed -= deathDuration
	}
	return VisualFrame{}, elapsed, false
}

func visualFrame(phase VisualPhase, impactIndex, segmentIndex, resolvedImpacts int, elapsed, duration time.Duration) VisualFrame {
	return VisualFrame{
		Phase:           phase,
		ImpactIndex:     impactIndex,
		SegmentIndex:    segmentIndex,
		ResolvedImpacts: resolvedImpacts,
		Progress:        float64(elapsed) / float64(duration),
	}
}
