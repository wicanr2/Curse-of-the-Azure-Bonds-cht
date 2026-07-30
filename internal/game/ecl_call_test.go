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

func TestApplyECLCallSignalsRedrawProjectsSameBlockDungeonRegisters(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         12,
		DungeonY:         9,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x42,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1096, BlockID: 0x42},
		},
		SaveWrites: []ecl.MemoryWrite{
			{Address: 0xC04B, Value: 11, PC: 0x1084, BlockID: 0x42},
			{Address: 0xC04C, Value: 10, PC: 0x108A, BlockID: 0x42},
			{Address: 0xC04D, Value: 2, PC: 0x1090, BlockID: 0x42},
		},
	})

	if state.DungeonX != 11 || state.DungeonY != 10 ||
		state.DungeonDirection != 4 || state.MapX != 11 || state.MapY != 10 {
		t.Fatalf("redraw position=(%d,%d,%d) map=(%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.MapX, state.MapY)
	}
}

func TestApplyECLCallSignalsRedrawIgnoresStaleDungeonRegisters(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         3,
		DungeonY:         4,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0x2E10}})

	if state.DungeonX != 3 || state.DungeonY != 4 || state.DungeonDirection != 6 {
		t.Fatalf("stale redraw changed position=(%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}

func TestApplyECLCallSignalsRedrawIgnoresPriorBlockTransaction(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
		0x43: {0, 0},
	}, 0x43)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         3,
		DungeonY:         4,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x43,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1096, BlockID: 0x42},
		},
		SaveWrites: []ecl.MemoryWrite{
			{Address: 0xC04B, Value: 11, PC: 0x1084, BlockID: 0x42},
			{Address: 0xC04C, Value: 10, PC: 0x108A, BlockID: 0x42},
			{Address: 0xC04D, Value: 2, PC: 0x1090, BlockID: 0x42},
		},
	})

	if state.DungeonX != 3 || state.DungeonY != 4 || state.DungeonDirection != 6 {
		t.Fatalf("prior-block redraw changed position=(%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}

func TestApplyECLCallSignalsRedrawProjectsOnlyFreshRegisters(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 10)
	session.SetMemoryValue(0xC04C, 2)
	session.SetMemoryValue(0xC04D, 0)
	state := State{
		session:          session,
		DungeonX:         9,
		DungeonY:         2,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x42,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x13DB, BlockID: 0x42},
		},
		SaveWrites: []ecl.MemoryWrite{
			{Address: 0xC04B, Value: 10, PC: 0x13CF, BlockID: 0x42},
			{Address: 0xC04D, Value: 0, PC: 0x13D5, BlockID: 0x42},
		},
	})

	if state.DungeonX != 10 || state.DungeonY != 2 ||
		state.DungeonDirection != 0 || state.MapX != 10 || state.MapY != 2 {
		t.Fatalf("partial redraw position=(%d,%d,%d) map=(%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.MapX, state.MapY)
	}
}

func TestApplyECLCallSignalsRedrawRequiresFreshDirectionCommit(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x01: {0, 0},
	}, 0x01)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 0)
	session.SetMemoryValue(0xC04C, 0)
	state := State{
		session:          session,
		DungeonX:         6,
		DungeonY:         5,
		DungeonDirection: 2,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x01,
		SessionEndBlockID:    0x01,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1452, BlockID: 0x01},
		},
		SaveWrites: []ecl.MemoryWrite{
			{Address: 0xC04B, Value: 0, PC: 0x1444, BlockID: 0x01},
			{Address: 0xC04C, Value: 0, PC: 0x144B, BlockID: 0x01},
		},
	})

	if state.DungeonX != 6 || state.DungeonY != 5 || state.DungeonDirection != 2 {
		t.Fatalf("scratch coordinates changed position=(%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}
