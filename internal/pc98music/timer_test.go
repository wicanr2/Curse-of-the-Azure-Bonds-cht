package pc98music

import "testing"

func TestPC98TimerBClockCyclesUsesObservedClockConfiguration(t *testing.T) {
	cycles, err := PC98TimerBClockCycles(0xBA)
	if err != nil {
		t.Fatal(err)
	}
	if cycles != 80_640 {
		t.Fatalf("Timer B cycles=%d, want 80640", cycles)
	}

	accumulator, err := NewPC98TimerBSampleAccumulator(48_000)
	if err != nil {
		t.Fatal(err)
	}
	samples, err := accumulator.Advance(0xBA, PC98YM2203DefaultPrescale)
	if err != nil {
		t.Fatal(err)
	}
	if samples != 969 {
		t.Fatalf("first 0xBA period samples=%d, want 969", samples)
	}
}
