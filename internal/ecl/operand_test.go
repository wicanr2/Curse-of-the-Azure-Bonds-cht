package ecl

import "testing"

func TestParseWordAndScalarOperands(t *testing.T) {
	// The byte at offset is skipped by the original ECL cursor convention.
	payload := []byte{0x88, 0x01, 0x6a, 0x80, 0x00, 0x2a, 0x99, 0x7f}
	operands, next, err := ParseOperands(payload, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if operands[0].Word != 0x806a || !operands[0].WordSet {
		t.Fatalf("first operand: %#v", operands[0])
	}
	if operands[1].Code != 0x00 || operands[1].Low != 0x2a || operands[1].WordSet {
		t.Fatalf("second operand: %#v", operands[1])
	}
	if next != 6 {
		t.Fatalf("next=%d, want 6", next)
	}
}

func TestRejectsTruncatedOperand(t *testing.T) {
	if _, _, err := ParseOperands([]byte{0, 1}, 0, 1); err == nil {
		t.Fatal("expected truncation error")
	}
}

func TestParsePackedOperandConsumesLengthPrefixedBytes(t *testing.T) {
	payload := []byte{0x11, 0x80, 0x03, 0xAA, 0xBB, 0xCC, 0x00}
	operands, next, err := ParseOperands(payload, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if operands[0].Code != 0x80 || string(operands[0].Packed) != string([]byte{0xAA, 0xBB, 0xCC}) {
		t.Fatalf("packed operand: %#v", operands[0])
	}
	if next != 6 {
		t.Fatalf("next=%d, want 6", next)
	}
}

func TestParseStringMemoryOperandConsumesWord(t *testing.T) {
	payload := []byte{0x11, 0x81, 0x34, 0x7C, 0x00}
	operands, next, err := ParseOperands(payload, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !operands[0].WordSet || operands[0].Word != 0x7C34 {
		t.Fatalf("string-memory operand: %#v", operands[0])
	}
	if next != 4 {
		t.Fatalf("next=%d, want 4", next)
	}
}

func TestFindSaveDestinationCandidatesChecksEveryByte(t *testing.T) {
	// Payload offset 0 also decodes as a longer PRINT candidate. A
	// resynchronizing linear scan would skip the real SAVE at offset 1.
	block := []byte{
		0, 0,
		0x11,
		0x09, 0x00, 0x01, 0x01, 0x08, 0x52,
	}
	instructions, err := FindSaveDestinationCandidates(block, 0x5208)
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != 1 || instructions[0].Offset != 1 ||
		instructions[0].Operands[0].Low != 1 {
		t.Fatalf("SAVE candidates=%+v", instructions)
	}
	if other, err := FindSaveDestinationCandidates(block, 0x5209); err != nil || len(other) != 0 {
		t.Fatalf("unexpected other destination=%+v err=%v", other, err)
	}
}
