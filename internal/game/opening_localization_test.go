package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

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
	got := localizeECLText(locale.Catalog{}, original)
	for _, want := range []string{"提爾佛頓", "五個青色", "符印下了詛咒", "金屬枷鎖", "被遺忘的國度", "重新掌握自己的命運", "致命敵人"} {
		if !strings.Contains(got, want) {
			t.Fatalf("localized opening %q does not contain %q", got, want)
		}
	}
	for _, english := range original {
		if strings.Contains(got, english) {
			t.Fatalf("localized opening retained English line %q", english)
		}
	}
}
