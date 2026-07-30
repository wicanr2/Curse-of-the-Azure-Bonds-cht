package pc98sfx

import "testing"

func TestPulseHalfCyclesUsesV30LoopExitTiming(t *testing.T) {
	got, err := PulseHalfCycles(1000, 5, 13, 30)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5_038 {
		t.Fatalf("cycles=%d, want 5038", got)
	}
}

func TestV30ProfileSeparatesRoutineEdgeOverheads(t *testing.T) {
	profile := V30PrefetchedProfile(8_000_000)
	tests := []struct {
		name     string
		overhead uint64
		want     uint64
	}{
		{name: "first gate on", overhead: profile.InitialGateOnOverheadCycles, want: 5_106},
		{name: "later gate on", overhead: profile.GateOnOverheadCycles, want: 5_038},
		{name: "non-final gate off", overhead: profile.GateOffOverheadCycles, want: 5_064},
		{name: "final gate off", overhead: profile.FinalGateOffOverheadCycles, want: 5_036},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := PulseHalfCycles(1000, profile.LoopTakenCycles,
				profile.LoopFinalCycles, test.overhead)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("cycles=%d, want %d", got, test.want)
			}
		})
	}
}

func TestRenderPCMProducesDeterministicDutyCycle(t *testing.T) {
	effect := Effect{Steps: []Step{{
		Kind:              StepPulse,
		FrequencyOrPeriod: 1000,
		PulseCount:        3,
	}}}
	profile := V30PrefetchedProfile(8_000_000)
	first, err := RenderPCM(effect, profile, 44_100)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RenderPCM(effect, profile, 44_100)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 168 {
		t.Fatalf("sample count=%d, want 168 with padded tail", len(first))
	}
	if len(second) != len(first) {
		t.Fatalf("second sample count=%d, want %d", len(second), len(first))
	}
	nonzero := 0
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("sample %d differs: %d != %d", index, first[index], second[index])
		}
		if first[index] != 0 {
			nonzero++
		}
	}
	if nonzero == 0 {
		t.Fatal("speaker PCM is silent")
	}
}

func TestRenderPCMIncludesProgramDelay(t *testing.T) {
	profile := V30PrefetchedProfile(8_000_000)
	got, err := RenderPCM(Effect{Steps: []Step{{
		Kind:              StepDelay,
		DelayMilliseconds: 5,
	}}}, profile, 44_100)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 221 {
		t.Fatalf("five-millisecond sample count=%d, want 221 with padded tail", len(got))
	}
}
