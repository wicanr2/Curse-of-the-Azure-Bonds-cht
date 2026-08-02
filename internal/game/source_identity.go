package game

import "slices"

// OriginalChoiceIndex finds an active option by the original ECL source
// token. Automation and tests must not branch on localized display text.
func (s *State) OriginalChoiceIndex(sources ...string) (int, bool) {
	for _, source := range sources {
		if index := slices.Index(s.currentOriginalChoices, source); index >= 0 {
			return index, true
		}
	}
	return 0, false
}

// MessageContainsGamePackText tests a displayed message through stable
// game-pack identity. The localized text is resolved at runtime so translation
// edits do not require a second control-flow literal.
func (s *State) MessageContainsGamePackText(messageID string) bool {
	if s.dataPack == nil {
		return false
	}
	value, found := s.dataPack.Text(messageID, s.catalog.Language)
	return found && value != "" && containsText(s.Message, value)
}

func containsText(value, fragment string) bool {
	if len(fragment) > len(value) {
		return false
	}
	for offset := 0; offset+len(fragment) <= len(value); offset++ {
		if value[offset:offset+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
