package save

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
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

func TestGameJSONRoundTripRestoresDungeonViewState(t *testing.T) {
	roster := party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	data, err := EncodeGameWithDungeonState(roster, area.State{GameArea: 2, InDungeon: true}, 3, 1, 4, 5, 11, 6, 2, 7, 0x40)
	if err != nil {
		t.Fatal(err)
	}
	file, err := DecodeGame(data)
	if err != nil {
		t.Fatal(err)
	}
	if file.Version != CurrentGameVersion || file.DungeonX != 11 || file.DungeonY != 6 || file.DungeonDir != 2 || file.DungeonWallType != 7 || file.DungeonWallRoof != 0x40 {
		t.Fatalf("decoded dungeon state=%+v", file)
	}
}

func TestGameJSONRoundTripRestoresGameTime(t *testing.T) {
	roster := party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	wantClock := [7]uint16{2, 4, 1, 3, 5, 6, 7}
	data, err := EncodeGameWithTime(roster, area.State{}, 3, 1, 12, 8, 7, 13, 0, 0, 0, wantClock, 9)
	if err != nil {
		t.Fatal(err)
	}
	file, err := DecodeGame(data)
	if err != nil {
		t.Fatal(err)
	}
	if file.Version != CurrentGameVersion || file.GameTime != wantClock || file.GameAgeCycles != 9 {
		t.Fatalf("decoded game time=%+v, want clock=%v age=9", file, wantClock)
	}
}

func TestDecodeGameAcceptsLegacyPartySave(t *testing.T) {
	file, err := DecodeGame([]byte(`{"version":1,"characters":[{"id":"p1","name":"阿勇","race":5,"class":1,"level":1,"abilities":{"strength":16,"intelligence":10,"wisdom":10,"dexterity":12,"constitution":14,"charisma":10}}]}`))
	if err != nil || file.Version != 1 {
		t.Fatalf("legacy decode file=%+v err=%v", file, err)
	}
}

func TestDecodeGameAcceptsVersion3DungeonSave(t *testing.T) {
	file, err := DecodeGame([]byte(`{"version":3,"characters":[{"id":"p1","name":"阿勇","race":5,"class":1,"level":1,"abilities":{"strength":16,"intelligence":10,"wisdom":10,"dexterity":12,"constitution":14,"charisma":10}}],"dungeon_x":11,"dungeon_y":6,"dungeon_direction":2}`))
	if err != nil || file.Version != 3 || file.DungeonX != 11 || file.DungeonY != 6 || file.DungeonDir != 2 {
		t.Fatalf("version 3 decode file=%+v err=%v", file, err)
	}
	if file.DungeonWallType != 0 || file.DungeonWallRoof != 0 {
		t.Fatalf("version 3 wall cache=%+v, want zero defaults", file)
	}
}

func TestEncodeCurrentGameVersionCarriesECLSession(t *testing.T) {
	roster := party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	snapshot := ecl.SessionSnapshot{Version: 1, CurrentBlock: 0x42, PC: 123, Started: true}
	data, err := EncodeGameWithSession(roster, area.State{}, 3, 1, 0, 0, 7, 13, 0, 0, 0, [7]uint16{}, 0, &snapshot)
	if err != nil {
		t.Fatal(err)
	}
	file, err := DecodeGame(data)
	if err != nil {
		t.Fatal(err)
	}
	if file.Version != CurrentGameVersion || file.ECLSession == nil || file.ECLSession.CurrentBlock != 0x42 || file.ECLSession.PC != 123 {
		t.Fatalf("decoded game=%+v", file)
	}
}

func TestPartyJSONRejectsUnknownVersion(t *testing.T) {
	if _, err := DecodeParty([]byte(`{"version":99,"characters":[]}`)); err == nil {
		t.Fatal("expected version error")
	}
}

func TestDecodeGameRejectsCombatPayloadBeforeVersionSeven(t *testing.T) {
	data := []byte(`{"version":6,"characters":[{"id":"p1","name":"阿勇","race":5,"class":1,"level":1,"abilities":{"strength":16,"intelligence":10,"wisdom":10,"dexterity":12,"constitution":14,"charisma":10}}],"combat":{"battle":{"version":1}}}`)
	if _, err := DecodeGame(data); err == nil {
		t.Fatal("version 6 combat payload unexpectedly decoded")
	}
}
