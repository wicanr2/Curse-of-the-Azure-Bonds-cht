package pc98music

import "github.com/wicanr2/golden-box-remake-engine/audio/ym2203"

const (
	PC98YM2203ClockHz         = uint64(3_993_600)
	PC98YM2203DefaultPrescale = uint32(6)
)

// PC98TimerBClockCycles maps this title's observed default-prescale YM2203
// Timer B data to one complete count period.
func PC98TimerBClockCycles(value byte) (uint64, error) {
	return ym2203.TimerBClockCycles(value, PC98YM2203DefaultPrescale)
}

// NewPC98TimerBSampleAccumulator creates the title adapter for the exact
// 3,993,600 Hz clock declared by the original PC-98 S98 runtime trace.
func NewPC98TimerBSampleAccumulator(
	sampleRate uint64,
) (*ym2203.TimerBSampleAccumulator, error) {
	return ym2203.NewTimerBSampleAccumulator(PC98YM2203ClockHz, sampleRate)
}
