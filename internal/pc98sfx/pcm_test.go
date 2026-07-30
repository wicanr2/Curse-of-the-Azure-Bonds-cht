package pc98sfx

import "testing"

func TestPulseHalfCyclesUsesV30LoopExitTiming(t *testing.T) {
	got, err := PulseHalfCycles(1000, 13, 5, 29)
	if err != nil {
		t.Fatal(err)
	}
	if got != 13_021 {
		t.Fatalf("cycles=%d, want 13021", got)
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
	if len(first) != 432 {
		t.Fatalf("sample count=%d, want 432 with padded tail", len(first))
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
