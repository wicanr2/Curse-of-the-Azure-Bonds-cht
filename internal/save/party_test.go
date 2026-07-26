package save

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
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

func TestGameJSONRoundTripRestoresAdventureState(t *testing.T) {
	roster := party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	areaState := area.State{GameArea: 4, InDungeon: true, Current3DMapBlockID: 0x25, CurrentCity: 2, LastXPos: -3, LastYPos: 7, LastECLBlockID: 0x1234}
	data, err := EncodeGame(roster, areaState, 3, 1, 12, 8)
	if err != nil {
		t.Fatal(err)
	}
	file, err := DecodeGame(data)
	if err != nil {
		t.Fatal(err)
	}
	if file.Area != areaState || file.Mode != 3 || file.Location != 1 || file.MapX != 12 || file.MapY != 8 {
		t.Fatalf("decoded game state=%+v", file)
	}
}

func TestDecodeGameAcceptsLegacyPartySave(t *testing.T) {
	file, err := DecodeGame([]byte(`{"version":1,"characters":[{"id":"p1","name":"阿勇","race":5,"class":1,"level":1,"abilities":{"strength":16,"intelligence":10,"wisdom":10,"dexterity":12,"constitution":14,"charisma":10}}]}`))
	if err != nil || file.Version != 1 {
		t.Fatalf("legacy decode file=%+v err=%v", file, err)
	}
}

func TestPartyJSONRejectsUnknownVersion(t *testing.T) {
	if _, err := DecodeParty([]byte(`{"version":99,"characters":[]}`)); err == nil {
		t.Fatal("expected version error")
	}
}
