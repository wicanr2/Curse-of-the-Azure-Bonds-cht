package ecl

import engineecl "github.com/wicanr2/golden-box-remake-engine/ecl"

// PackedTextCandidate is retained as a compatibility alias while the generic
// six-bit decoder lives in golden-box-remake-engine/ecl.
type PackedTextCandidate = engineecl.PackedTextCandidate

func DecodePackedText(payload []byte) string {
	return engineecl.DecodePackedText(payload)
}

func FindPackedTextCandidates(data []byte) []string {
	return engineecl.FindPackedTextCandidates(data)
}

func FindPackedTextCandidatesAt(data []byte) []PackedTextCandidate {
	return engineecl.FindPackedTextCandidatesAt(data)
}
