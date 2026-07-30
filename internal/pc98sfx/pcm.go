package pc98sfx

import (
	"fmt"

	"github.com/wicanr2/golden-box-remake-engine/audio/cyclepcm"
)

// TimingProfile describes the cycle assumptions used to reconstruct the
// software-driven PC-98 speaker waveform. It is intentionally explicit:
// GAME.EXE proves the loop, but the selected machine clock and bus/prefetch
// timing remain a playback profile until calibrated against original audio.
type TimingProfile struct {
	ClockHz                     uint64
	LoopTakenCycles             uint64
	LoopFinalCycles             uint64
	InitialGateOnOverheadCycles uint64
	GateOnOverheadCycles        uint64
	GateOffOverheadCycles       uint64
	FinalGateOffOverheadCycles  uint64
	Amplitude                   int16
}

// V30PrefetchedProfile uses NEC's documented V30 BCWZ/LOOP timing (5 clocks
// while taken, 13 clocks on exit) and the instruction-path overhead reconstructed
// from GAME.EXE under the no-wait, prefetched-instruction assumption.
func V30PrefetchedProfile(clockHz uint64) TimingProfile {
	return TimingProfile{
		ClockHz:                     clockHz,
		LoopTakenCycles:             5,
		LoopFinalCycles:             13,
		InitialGateOnOverheadCycles: 98,
		GateOnOverheadCycles:        30,
		GateOffOverheadCycles:       56,
		FinalGateOffOverheadCycles:  28,
		Amplitude:                   6000,
	}
}

// PulseHalfCycles returns one busy-loop half-wave including its surrounding
// instruction path. A loop count of zero is invalid for the recovered program.
func PulseHalfCycles(count uint16, loopTaken, loopFinal, overhead uint64) (uint64, error) {
	if count == 0 {
		return 0, fmt.Errorf("PC-98 speaker loop count must be nonzero")
	}
	return uint64(count-1)*loopTaken + loopFinal + overhead, nil
}

// RenderPCM reconstructs one effect as signed 16-bit mono PCM.
func RenderPCM(effect Effect, profile TimingProfile, sampleRate uint64) ([]int16, error) {
	if profile.ClockHz == 0 || profile.LoopTakenCycles == 0 ||
		profile.LoopFinalCycles == 0 || profile.Amplitude == 0 {
		return nil, fmt.Errorf("PC-98 speaker timing profile is incomplete")
	}
	renderer, err := cyclepcm.NewRenderer(profile.ClockHz, sampleRate)
	if err != nil {
		return nil, err
	}
	var output []int16
	appendSegments := func(segments []cyclepcm.Segment) error {
		samples, renderErr := renderer.Render(segments)
		if renderErr == nil {
			output = append(output, samples...)
		}
		return renderErr
	}
	for _, step := range effect.Steps {
		switch step.Kind {
		case StepPulse:
			for pulse := uint16(0); pulse < step.PulseCount; pulse++ {
				onOverhead := profile.GateOnOverheadCycles
				if pulse == 0 {
					onOverhead = profile.InitialGateOnOverheadCycles
				}
				onCycles, cycleErr := PulseHalfCycles(
					step.FrequencyOrPeriod,
					profile.LoopTakenCycles,
					profile.LoopFinalCycles,
					onOverhead,
				)
				if cycleErr != nil {
					return nil, cycleErr
				}
				offOverhead := profile.GateOffOverheadCycles
				if pulse+1 == step.PulseCount {
					offOverhead = profile.FinalGateOffOverheadCycles
				}
				offCycles, cycleErr := PulseHalfCycles(
					step.FrequencyOrPeriod,
					profile.LoopTakenCycles,
					profile.LoopFinalCycles,
					offOverhead,
				)
				if cycleErr != nil {
					return nil, cycleErr
				}
				if err := appendSegments([]cyclepcm.Segment{
					{Cycles: onCycles, Level: profile.Amplitude},
					{Cycles: offCycles, Level: 0},
				}); err != nil {
					return nil, err
				}
			}
		case StepDelay:
			delayCycles := profile.ClockHz * uint64(step.DelayMilliseconds) / 1000
			if err := appendSegments([]cyclepcm.Segment{{Cycles: delayCycles}}); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown PC-98 speaker step %q", step.Kind)
		}
	}
	output = append(output, renderer.Flush()...)
	return output, nil
}
