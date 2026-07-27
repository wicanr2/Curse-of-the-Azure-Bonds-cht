package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func TestApplyECLCallSignalsMovesForwardAndWraps(t *testing.T) {
	tests := []struct {
		name      string
		x, y      int
		direction uint8
		wantX     int
		wantY     int
	}{
		{name: "north wraps", x: 7, y: 0, direction: 0, wantX: 7, wantY: 15},
		{name: "east wraps", x: 15, y: 6, direction: 2, wantX: 0, wantY: 6},
		{name: "south wraps", x: 7, y: 15, direction: 4, wantX: 7, wantY: 0},
		{name: "west wraps", x: 0, y: 6, direction: 6, wantX: 15, wantY: 6},
		{name: "odd direction unchanged", x: 7, y: 6, direction: 1, wantX: 7, wantY: 6},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := State{DungeonX: test.x, DungeonY: test.y, DungeonDirection: test.direction}
			state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0xC01E}})
			if state.DungeonX != test.wantX || state.DungeonY != test.wantY {
				t.Fatalf("position=(%d,%d), want (%d,%d)", state.DungeonX, state.DungeonY, test.wantX, test.wantY)
			}
			if got := state.ConsumeECLCallRequests(); !reflect.DeepEqual(got, []uint16{0xC01E}) {
				t.Fatalf("requests=%#v", got)
			}
			if got := state.ConsumeECLCallRequests(); len(got) != 0 {
				t.Fatalf("requests were not one-shot: %#v", got)
			}
		})
	}
}

func TestApplyECLCallSignalsPreservesOrderAndDefaultSound(t *testing.T) {
	state := State{}
	state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0x2E10, 0xB200, 0x9999}})

	if got, want := state.ConsumeECLCallRequests(), []uint16{0x2E10, 0xB200, 0x9999}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests=%#v, want %#v", got, want)
	}
	if got, want := state.ConsumeSoundEvents(), []SoundEvent{SoundStep}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sound=%#v, want %#v", got, want)
	}
}
