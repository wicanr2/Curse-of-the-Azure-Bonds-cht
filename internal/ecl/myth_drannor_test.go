package ecl

import (
	"archive/zip"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealStandingStoneToMythDrannorBurialGlen(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()

	all := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(realZipMember(t, archive, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			all[block.Entry.ID] = block.Data
		}
	}
	session, err := NewBlockSession(all, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0x4C59, 1)
	session.SetMemoryValue(0x4C5A, 1)
	session.SetMemoryValue(0x4C5B, 0xFF)
	session.SetMemoryValue(0x4C9B, 4)

	result, err := session.RunFrom(0x01C4, 1000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.WaitingForMenu || !strings.Contains(strings.Join(result.Text, " "), "STANDING STONES") {
		t.Fatalf("initial result=%+v, want Standing Stones prompt", result)
	}

	press := uint16(0)
	reveal, err := session.ResumeInteractiveSelectionSeed(1000, &press, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	revealText := strings.Join(reveal.Text, " ")
	if !reveal.WaitingForMenu ||
		!strings.Contains(revealText, "REVEALING TYRANTHRAXUS") ||
		!strings.Contains(revealText, "MEET ME AT MYTH DRANNOR") {
		t.Fatalf("reveal=%+v text=%q", reveal, revealText)
	}
	if count := mustMemory(t, session, 0x7F79); count != 3 {
		t.Fatalf("removed-master count=%d, want 3", count)
	}

	actions, err := session.ResumeInteractiveSelectionSeed(1000, &press, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !actions.WaitingForMenu ||
		!reflect.DeepEqual(actions.Menus[len(actions.Menus)-1].Options, []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Standing Stone actions=%+v", actions.Menus)
	}
	journeyOn := uint16(1)
	destinations, err := session.ResumeInteractiveSelectionSeed(1000, &journeyOn, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	wantDestinations := []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR", "MYTH DRANNOR"}
	if !destinations.WaitingForMenu ||
		!reflect.DeepEqual(destinations.Menus[len(destinations.Menus)-1].Options, wantDestinations) {
		t.Fatalf("destinations=%+v, want %v", destinations.Menus, wantDestinations)
	}

	// ECL1's world adapter copies the destination bytes from the current
	// adjacency row into these option-indexed work cells.
	for index, value := range []uint16{2, 8, 11, 13} {
		session.SetMemoryValue(0x4C02+uint16(index), value)
	}
	mythDrannor := uint16(3)
	routes, err := session.ResumeInteractiveSelectionSeed(1000, &mythDrannor, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !routes.WaitingForMenu ||
		!strings.Contains(strings.Join(routes.Text, " "), "HOW WILL YOU GET TO MYTH DRANNOR") ||
		!reflect.DeepEqual(routes.Menus[len(routes.Menus)-1].Options, []string{"WILDERNESS", "EXIT"}) {
		t.Fatalf("Myth Drannor routes=%+v text=%q", routes.Menus, routes.Text)
	}

	wilderness := uint16(0)
	if _, err := session.ResumeInteractiveSelectionSeed(1000, &wilderness, nil, 1, PartyContext{}); err != nil {
		t.Fatal(err)
	}
	// The AREA wilderness loop supplies the destination through 0x4C9C when
	// the party reaches the city edge, then invokes SearchLocation.
	session.SetMemoryValue(0x4C9C, 13)
	edge, err := session.RunEntry(1, 1000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !edge.WaitingForMenu ||
		!strings.Contains(strings.Join(edge.Text, " "), "EDGE OF MYTH DRANNOR") ||
		!reflect.DeepEqual(edge.Menus[len(edge.Menus)-1].Options, []string{"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA"}) {
		t.Fatalf("Myth Drannor edge=%+v text=%q", edge.Menus, edge.Text)
	}

	enterCity := uint16(0)
	burialGlen, err := session.ResumeInteractiveSelectionSeed(1000, &enterCity, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x40 ||
		!burialGlen.Exited ||
		!burialGlen.LoadFilesRequested || burialGlen.LoadFiles != [3]uint16{0x40, 2, 0xFF} ||
		!burialGlen.LoadPiecesRequested || burialGlen.LoadPieces != [3]uint16{17, 18, 16} ||
		!strings.Contains(strings.Join(burialGlen.Text, " "), "TYRANTHRAXUS IS TO THE NORTH") {
		t.Fatalf("Burial Glen block=0x%02X result=%+v", session.CurrentBlockID(), burialGlen)
	}
}

func TestRealBurialGlenElfSpiritChoices(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL6.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	all := make(map[uint8][]byte)
	for _, block := range blocks {
		all[block.Entry.ID] = block.Data
	}
	tests := []struct {
		name      string
		selection uint16
		contains  []string
	}{
		{
			name:      "greet",
			selection: 0,
			contains:  []string{"THE SPIRIT TALKS OF THE GLEN", "JOURNAL ENTRY", "25. THEN, THE SPIRIT FADES"},
		},
		{
			name:      "flee",
			selection: 1,
			contains:  []string{"SO, YOU ARE SHEEP", "THEN YOU SHALL FEED ME", "THE SPIRIT FADES"},
		},
		{
			name:      "attack",
			selection: 2,
			contains:  []string{"THE SPIRIT DISAPPEARS"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session, err := NewBlockSession(all, 0x40)
			if err != nil {
				t.Fatal(err)
			}
			session.SetMemoryValue(0xC04B, 3)
			session.SetMemoryValue(0xC04C, 14)
			session.SetMemoryValue(0xC04D, 0)
			session.SetMemoryValue(0xC04F, 1)
			initial, err := session.RunEntry(1, 1000, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !initial.PictureRequested || initial.PictureBlock != 72 ||
				!initial.WaitingForMenu ||
				!reflect.DeepEqual(initial.Menus[len(initial.Menus)-1].Options, []string{"GREET", "FLEE", "ATTACK"}) ||
				!strings.Contains(strings.Join(initial.Text, " "), "AN ELFISH SPIRIT APPEARS AND GREETS YOU") {
				t.Fatalf("initial=%+v", initial)
			}
			result, err := session.ResumeInteractiveSelectionSeed(
				1000, &test.selection, nil, 1, PartyContext{},
			)
			if err != nil {
				t.Fatal(err)
			}
			joined := strings.Join(result.Text, " ")
			for _, fragment := range test.contains {
				if !strings.Contains(joined, fragment) {
					t.Fatalf("selection %d text=%q, missing %q", test.selection, joined, fragment)
				}
			}
		})
	}
}

func mustMemory(t *testing.T, session *BlockSession, address uint16) uint16 {
	t.Helper()
	value, ok := session.MemoryValue(address)
	if !ok {
		t.Fatalf("missing memory 0x%04X", address)
	}
	return value
}
