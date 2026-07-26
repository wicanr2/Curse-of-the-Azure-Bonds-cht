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
