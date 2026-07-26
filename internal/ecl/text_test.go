package ecl

import "testing"

func packText(text string) []byte {
	values := make([]byte, 0, len(text))
	for i := 0; i < len(text); i++ {
		value := text[i]
		if value >= 'A' && value <= 'Z' {
			value -= 0x40
		}
		values = append(values, value)
	}
	for len(values)%4 != 0 {
		values = append(values, 0)
	}
	packed := make([]byte, 0, len(values)*3/4)
	for i := 0; i < len(values); i += 4 {
		packed = append(packed,
			(values[i]<<2)|(values[i+1]>>4),
			((values[i+1]&0x0f)<<4)|(values[i+2]>>2),
			((values[i+2]&0x03)<<6)|values[i+3])
	}
	return packed
}

func TestDecodePackedText(t *testing.T) {
	if got := DecodePackedText(packText("YOU ARE")); got != "YOU ARE" {
		t.Fatalf("got %q", got)
	}
}

func TestFindPackedTextCandidates(t *testing.T) {
	packed := packText("ENTER CITY")
	data := append([]byte{0x80, byte(len(packed))}, packed...)
	got := FindPackedTextCandidates(data)
	if len(got) != 1 || got[0] != "ENTER CITY" {
		t.Fatalf("got %#v", got)
	}
}

func TestFindRealECLCandidate(t *testing.T) {
	payload := []byte{0x64, 0xf5, 0x60, 0x05, 0x21, 0x60, 0x05, 0x48, 0x14, 0x20, 0x58, 0x05, 0x10, 0x71, 0x60, 0x3c, 0x68, 0x00}
	if got := DecodePackedText(payload); got != "YOU ARE AT THE EDGE OF" {
		t.Fatalf("got %q", got)
	}
}
