package ecl

import "strings"

// PackedTextCandidate preserves the byte offset of a readable-looking packed
// string so reverse-engineering tools can jump directly to its instruction.
type PackedTextCandidate struct {
	Offset int
	Text   string
}

// DecodePackedText decodes the six-bit text payload used by ECL string
// records. Zero values are padding and are omitted.
func DecodePackedText(payload []byte) string {
	var out strings.Builder
	state := 1
	var previous byte
	for _, current := range payload {
		appendValue := func(value byte) {
			if value == 0 {
				return
			}
			if value <= 0x1f {
				value += 0x40
			}
			out.WriteByte(value)
		}
		switch state {
		case 1:
			appendValue((current >> 2) & 0x3f)
			state = 2
		case 2:
			appendValue(((previous << 4) | (current >> 4)) & 0x3f)
			state = 3
		case 3:
			appendValue(((previous << 2) | (current >> 6)) & 0x3f)
			appendValue(current & 0x3f)
			state = 1
		}
		previous = current
	}
	return strings.TrimSpace(out.String())
}

// FindPackedTextCandidates returns readable-looking 0x80 length-prefixed
// strings. It intentionally retains the candidate nature: graphics and map
// payloads can contain the same byte marker by coincidence.
func FindPackedTextCandidates(data []byte) []string {
	positioned := FindPackedTextCandidatesAt(data)
	candidates := make([]string, 0, len(positioned))
	for _, candidate := range positioned {
		candidates = append(candidates, candidate.Text)
	}
	return candidates
}

// FindPackedTextCandidatesAt is the positioned form used by disassembly tools.
func FindPackedTextCandidatesAt(data []byte) []PackedTextCandidate {
	var candidates []PackedTextCandidate
	for i := 0; i+2 < len(data); i++ {
		if data[i] != 0x80 {
			continue
		}
		length := int(data[i+1])
		end := i + 2 + length
		if end > len(data) {
			continue
		}
		text := DecodePackedText(data[i+2 : end])
		if len(text) <= 3 || !hasLetter(text) || !strings.ContainsAny(text, " \t") {
			continue
		}
		candidates = append(candidates, PackedTextCandidate{Offset: i, Text: text})
	}
	return candidates
}

func hasLetter(text string) bool {
	for _, char := range text {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			return true
		}
	}
	return false
}
