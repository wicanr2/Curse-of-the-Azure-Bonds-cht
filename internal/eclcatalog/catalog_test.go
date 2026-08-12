package eclcatalog

import (
	"bytes"
	"testing"
)

func TestEffectKindsKeepBoundariesSeparate(t *testing.T) {
	tests := []struct {
		opcode byte
		want   []string
	}{
		{0x0E, []string{"presentation"}},
		{0x24, []string{"combat_boundary"}},
		{0x27, []string{"treasure_boundary"}},
		{0x2D, []string{"external_call"}},
		{0x38, []string{"program_boundary"}},
	}
	for _, test := range tests {
		got := effectKinds(test.opcode)
		if !equalStrings(got, test.want) {
			t.Fatalf("opcode %#x kinds=%v, want %v", test.opcode, got, test.want)
		}
	}
}

func TestOriginalCorpusCatalogIsDeterministic(t *testing.T) {
	first, err := BuildFile("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original archive unavailable: %v", err)
	}
	second, err := BuildFile("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := Encode(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := Encode(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("catalog output is not deterministic")
	}
	if !bytes.Equal(EncodeMarkdown(first), EncodeMarkdown(second)) {
		t.Fatal("catalog Markdown output is not deterministic")
	}
	if first.Summary.MemberCount != 6 || first.Summary.BlockCount != 25 ||
		first.Summary.LifecycleEntryCount != 125 ||
		first.Summary.UniqueReachableInstructionCount == 0 {
		t.Fatalf("unexpected summary: %+v", first.Summary)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
