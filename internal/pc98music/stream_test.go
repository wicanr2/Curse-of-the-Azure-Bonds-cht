package pc98music

import (
	"strings"
	"testing"
)

func TestDecodeStreamStructureFamiliesAndWidths(t *testing.T) {
	fm, err := DecodeStreamStructure(0, []byte{
		0x20, 0x04,
		0x85, 0x10,
		0x90,
		0xA1, 0x34, 0x12,
		0xB0, 0x28, 0x01,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fm) != 5 || fm[0].Name != "note" || fm[3].Name != "call" ||
		len(fm[3].Operands) != 2 || fm[4].Name != "extension_b0" {
		t.Fatalf("FM commands=%+v", fm)
	}
	psg, err := DecodeStreamStructure(3, []byte{0x80, 0x08, 0x91, 0x92, 0xA3, 0x02, 0xA4})
	if err != nil {
		t.Fatal(err)
	}
	if len(psg) != 5 || psg[1].Name != "mode_91" ||
		psg[2].Name != "mode_92" {
		t.Fatalf("PSG commands=%+v", psg)
	}
	timing, err := DecodeStreamStructure(6, []byte{
		0xA0, 0x34, 0x12,
		0x85, 0x09,
		0x80, 0x03,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(timing) != 4 || timing[0].Name != "ignored" ||
		timing[1].Name != "note" || timing[2].Name != "parameter_85" ||
		timing[3].Name != "rest" {
		t.Fatalf("timing commands=%+v", timing)
	}
}

func TestTimingMachineDoesNotExecuteFMPSGControlFlow(t *testing.T) {
	data := make([]byte, 0x110)
	copy(data[0x100:], []byte{0xA0, 0x01, 0x04})
	machine, err := NewSequenceMachine(6, 0x100, 0x103)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := machine.NextTimed(data, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 2 || commands[0].Name != "ignored" ||
		commands[1].Name != "note" || commands[1].Opcode != 0x01 {
		t.Fatalf("timing commands=%+v", commands)
	}
}

func TestTimingMachineAllowsReadThroughPastDeclaredRecord(t *testing.T) {
	data := make([]byte, 0x106)
	copy(data[0x100:], []byte{0x20, 0x04, 0xA0, 0xA4})
	copy(data[0x104:], []byte{0x30, 0x08})
	machine, _ := NewSequenceMachine(6, 0x100, len(data))
	if _, err := machine.NextTimed(data, 4); err != nil {
		t.Fatal(err)
	}
	commands, err := machine.NextTimed(data, 4)
	if err != nil || len(commands) != 3 || commands[2].Offset != 0x104 {
		t.Fatalf("tail=%+v err=%v", commands, err)
	}
}

func TestSequenceMachineExecutesCountedLoopAndCall(t *testing.T) {
	data := make([]byte, 0x130)
	copy(data[0x100:], []byte{
		0xA3, 0x02,
		0x20, 0x04,
		0xA4,
		0xA1, 0x10, 0x01,
		0x80, 0x03,
	})
	copy(data[0x110:], []byte{0x85, 0x22, 0x30, 0x02, 0xA2})
	machine, err := NewSequenceMachine(0, 0x100, 0x115)
	if err != nil {
		t.Fatal(err)
	}
	first, err := machine.NextTimed(data, 32)
	if err != nil || len(first) != 2 || first[1].Name != "note" {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	second, err := machine.NextTimed(data, 32)
	if err != nil || len(second) != 2 || second[1].Offset != 0x102 {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	third, err := machine.NextTimed(data, 32)
	if err != nil || len(third) != 4 || third[0].Name != "loop_end" ||
		third[1].Name != "call" || third[2].Name != "parameter_85" ||
		third[3].Offset != 0x112 {
		t.Fatalf("third=%+v err=%v", third, err)
	}
	fourth, err := machine.NextTimed(data, 32)
	if err != nil || len(fourth) != 2 || fourth[0].Name != "return" ||
		fourth[1].Name != "rest" {
		t.Fatalf("fourth=%+v err=%v", fourth, err)
	}
}

func TestSequenceMachineRejectsBadTargetsAndIgnoresStackUnderflow(t *testing.T) {
	data := make([]byte, 0x110)
	copy(data[0x100:], []byte{0xA0, 0x00, 0x02})
	machine, _ := NewSequenceMachine(0, 0x100, 0x103)
	if _, err := machine.NextTimed(data, 4); err == nil ||
		!strings.Contains(err.Error(), "outside sequence") {
		t.Fatalf("bad target err=%v", err)
	}
	copy(data[0x100:], []byte{0xA2, 0xA4, 0x20, 0x04})
	machine, _ = NewSequenceMachine(0, 0x100, 0x104)
	commands, err := machine.NextTimed(data, 4)
	if err != nil || len(commands) != 3 || commands[2].Name != "note" {
		t.Fatalf("underflow commands=%+v err=%v", commands, err)
	}
}

func TestDecodeStreamStructureRejectsFamilyMismatchAndTruncation(t *testing.T) {
	if _, err := DecodeStreamStructure(3, []byte{0x90}); err == nil ||
		!strings.Contains(err.Error(), "unknown opcode") {
		t.Fatalf("PSG accepted FM-only opcode: %v", err)
	}
	if _, err := DecodeStreamStructure(0, []byte{0xA1, 0x34}); err == nil ||
		!strings.Contains(err.Error(), "needs 2 operand bytes") {
		t.Fatalf("truncated call err=%v", err)
	}
	if _, err := DecodeStreamStructure(7, nil); err == nil ||
		!strings.Contains(err.Error(), "outside 0..6") {
		t.Fatalf("channel bounds err=%v", err)
	}
}
