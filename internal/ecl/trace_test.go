package ecl

import "testing"

func TestTraceStopsAtKnownExit(t *testing.T) {
	block := []byte{0x88, 0x13, 0x01, 0x01, 0x34, 0x12, 0x11, 0x00, 0x05, 0x00}
	trace, err := Trace(block, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(trace) != 3 || trace[0].Command.Name != "GOTO" || trace[1].Command.Name != "PRINT" || trace[2].Command.Name != "EXIT" {
		t.Fatalf("trace=%#v", trace)
	}
	if trace[0].Operands[0].Word != 0x1234 {
		t.Fatalf("goto operand=%#v", trace[0].Operands[0])
	}
}

func TestTraceStopsSafelyAtUnknownOpcode(t *testing.T) {
	trace, err := Trace([]byte{0, 0, 0x7F}, 20)
	if err == nil || len(trace) != 0 {
		t.Fatalf("trace=%#v err=%v", trace, err)
	}
}
