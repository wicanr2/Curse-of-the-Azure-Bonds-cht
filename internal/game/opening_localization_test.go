package game

import "testing"

func TestOpeningCurseTextIsLocalizedLineForLine(t *testing.T) {
	original := []string{
		"ON YOUR WAY TO THE TOWN OF TILVERTON YOU ARE",
		"AMBUSHED, CAPTURED, AND KNOCKED UNCONSCIOUS. WHEN",
		"YOU AWAKE YOUR PARTY HAS BEEN CURSED WITH FIVE AZURE",
		"SYMBOLS.",
		"THE SYMBOLS ENSNARE YOUR WILL LIKE METAL BONDS.",
		"AND WHEN THE BONDS GLOW YOU MUST DO AS THEY COMMAND.",
		"YOUR ONLY HOPE IS TO SEARCH THE FORGOTTEN REALMS",
		"FOR THE MEMBERS OF THE ALLIANCE WHO CREATED THE BONDS",
		"AND REGAIN CONTROL OF YOUR OWN DESTINY.",
		"NOWHERE IN THE REALMS IS COMPLETELY SAFE. EVEN",
		"THE MOST PEACEFUL SCENE CAN HIDE A DEADLY FOE.",
	}
	state := NewState(testCatalog())
	got := state.localizeECLText(original)
	want := requireGamePackText(t, &state, "opening.curse-summary")
	if got != want {
		t.Fatalf("localized opening=%q, want game-pack message %q", got, want)
	}
}

func TestActualNewGameTextIsLocalized(t *testing.T) {
	awakening := []string{
		"YOU AWAKEN IN A SMALL ROOM. LOOKING AROUND, YOU NOTICE",
		"THAT ALL YOUR GEAR IS GONE, AS IS YOUR MEMORY OF RECENT EVENTS.",
	}
	marks := []string{
		"ADDING TO YOUR DISQUIET, YOU NOTICE THAT YOUR SWORD ARM",
		"HAS BEEN SOMEHOW IMPRINTED WITH STRANGE PATTERNS. THE REST",
		"OF YOUR PARTY ARE IDENTICALLY MARKED.",
	}
	state := NewState(testCatalog())
	for messageID, original := range map[string][]string{
		"opening.new-game-awakening": awakening,
		"opening.new-game-marks":     marks,
	} {
		got := state.localizeECLText(original)
		want := requireGamePackText(t, &state, messageID)
		if got != want {
			t.Fatalf("localized new game=%q, want game-pack message %q", got, want)
		}
	}
}
