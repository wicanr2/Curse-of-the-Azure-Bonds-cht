package combat

import "time"

// DeathOverlayPhaseDuration is the renderer-neutral cadence used by the
// remake preview for the reference CombatantKilled flash.
const DeathOverlayPhaseDuration = 100 * time.Millisecond

// DeathOverlayFrame returns the alternating skull frame (0=attack, 1=normal)
// while the nine-cycle flash is active. Wall-clock timestamps stay in the
// frontend; the combat core only owns this deterministic lifecycle rule.
func DeathOverlayFrame(elapsed time.Duration) (frame uint8, active bool) {
	if elapsed < 0 {
		elapsed = 0
	}
	phase := elapsed / DeathOverlayPhaseDuration
	if phase >= 9 {
		return 0, false
	}
	return uint8(phase % 2), true
}
