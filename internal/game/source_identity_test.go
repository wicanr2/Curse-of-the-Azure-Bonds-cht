package game

import "testing"

func TestSourceIdentityDoesNotDependOnLocalizedDisplayText(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.currentOriginalChoices = []string{"WAIT", "ATTACK WIZARD", "PARLAY_SLY"}
	state.Choices = []string{"TRANSLATION A", "TRANSLATION B", "TRANSLATION C"}

	if index, found := state.OriginalChoiceIndex("ATTACK WIZARD"); !found || index != 1 {
		t.Fatalf("original choice index=%d found=%v", index, found)
	}
	if index, found := state.OriginalChoiceIndex("MISSING", "PARLAY_SLY"); !found || index != 2 {
		t.Fatalf("fallback original choice index=%d found=%v", index, found)
	}
	if _, found := state.OriginalChoiceIndex("TRANSLATION B"); found {
		t.Fatal("localized display text must not resolve as source identity")
	}

	messageID := "wizard-tower.dragons-convinced"
	message, found := state.dataPack.Text(messageID, state.catalog.Language)
	if !found {
		t.Fatalf("game-pack message %q unavailable", messageID)
	}
	state.Message = "PREFIX " + message + " SUFFIX"
	if !state.MessageContainsGamePackText(messageID) {
		t.Fatalf("stable message identity did not resolve %q", messageID)
	}
	if state.MessageContainsGamePackText("missing.message") {
		t.Fatal("missing message identity unexpectedly matched")
	}
}
