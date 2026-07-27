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

func TestActualNewGameTextIsLocalized(t *testing.T) {
	original := []string{
		"YOU AWAKEN IN A SMALL ROOM. LOOKING AROUND, YOU NOTICE",
		"THAT ALL YOUR GEAR IS GONE, AS IS YOUR MEMORY OF RECENT EVENTS.",
		"ADDING TO YOUR DISQUIET, YOU NOTICE THAT YOUR SWORD ARM",
		"HAS BEEN SOMEHOW IMPRINTED WITH STRANGE PATTERNS. THE REST",
		"OF YOUR PARTY ARE IDENTICALLY MARKED.",
	}
	got := localizeECLText(locale.Catalog{}, original)
	for _, want := range []string{"小房間", "所有裝備都不見了", "持劍的手臂", "奇異圖紋", "相同的印記"} {
		if !strings.Contains(got, want) {
			t.Fatalf("localized new game %q does not contain %q", got, want)
		}
	}
}
