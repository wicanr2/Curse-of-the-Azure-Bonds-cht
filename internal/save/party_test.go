package save

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestPartyJSONRoundTrip(t *testing.T) {
	roster := party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	data, err := EncodeParty(roster)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeParty(data)
	if err != nil || len(got) != 1 || got[0].Name != "阿勇" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

func TestPartyJSONRejectsUnknownVersion(t *testing.T) {
	if _, err := DecodeParty([]byte(`{"version":99,"characters":[]}`)); err == nil {
		t.Fatal("expected version error")
	}
}
