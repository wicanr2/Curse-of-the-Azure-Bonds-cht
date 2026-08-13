package eclcatalog

import (
	"bytes"
	"strings"
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

func TestReviewLedgerFailsClosedAndAttachesByStableID(t *testing.T) {
	catalog := Catalog{Members: []Member{{Name: "ECL3.DAX", Blocks: []Block{{
		ID: "0x15", OrderedEffectCandidates: []OrderedEffectCandidate{{
			ID: "ECL3.DAX/0x15/0x050A-0x0578",
		}},
	}}}}}
	data := []byte(`{"format_version":1,"reviews":{"ECL3.DAX/0x15/0x050A-0x0578":{"status":"covered","confidence":"exact","spec_refs":["258-treasure-combat-continuation.md"],"note":"covered"}}}`)
	if err := ApplyReviewLedger(&catalog, data); err != nil {
		t.Fatal(err)
	}
	got := catalog.Members[0].Blocks[0].OrderedEffectCandidates[0].Review
	if got == nil || got.Status != "covered" || got.Confidence != "exact" {
		t.Fatalf("review=%+v", got)
	}
	bad := []byte(`{"format_version":1,"reviews":{"stale":{"status":"covered","confidence":"exact","spec_refs":["x.md"],"note":"bad"}}}`)
	if err := ApplyReviewLedger(&catalog, bad); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("stale review error=%v", err)
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
