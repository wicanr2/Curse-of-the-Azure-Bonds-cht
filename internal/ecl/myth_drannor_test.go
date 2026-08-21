package ecl

import (
	"archive/zip"
	"fmt"
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

func TestRealBurialGlenPrincessDaemirChoices(t *testing.T) {
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
	t.Run("combat initialization projects signed party modifier", func(t *testing.T) {
		for _, value := range []uint16{0x02, 0xFE} {
			session, sessionErr := NewBlockSession(all, 0x40)
			if sessionErr != nil {
				t.Fatal(sessionErr)
			}
			session.SetMemoryValue(0x4CBB, value)
			session.SetMemoryValue(0xC04B, 0)
			session.SetMemoryValue(0xC04C, 0)
			session.SetMemoryValue(0xC04D, 0)
			session.SetMemoryValue(0xC04F, 0)
			if _, runErr := session.RunEntry(1, 1000, nil); runErr != nil {
				t.Fatal(runErr)
			}
			if got := mustMemory(t, session, 0x7F71); got != value {
				t.Fatalf("SAVE [4CBB] -> [7F71] got=%02x want=%02x", got, value)
			}
		}
	})
	newSession := func(t *testing.T, approval uint16) *BlockSession {
		t.Helper()
		session, err := NewBlockSession(all, 0x40)
		if err != nil {
			t.Fatal(err)
		}
		session.SetMemoryValue(0xC04B, 13)
		session.SetMemoryValue(0xC04C, 14)
		session.SetMemoryValue(0xC04D, 4)
		session.SetMemoryValue(0xC04F, 0x03)
		session.SetMemoryValue(0x4CBA, approval)
		return session
	}
	runPrompt := func(t *testing.T, session *BlockSession, wantFragment string) RunResult {
		t.Helper()
		initial, err := session.RunEntry(1, 1000, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !initial.PictureRequested || initial.PictureBlock != 72 ||
			!initial.WaitingForMenu ||
			!reflect.DeepEqual(initial.Menus[len(initial.Menus)-1].Options,
				[]string{"ACCEPT", "REJECT", "KILL", "FLEE"}) ||
			!strings.Contains(strings.Join(initial.Text, " "), "PRINCESS DAEMIR") ||
			!strings.Contains(strings.Join(initial.Text, " "), wantFragment) ||
			mustMemory(t, session, 0x4CC0) != 1 {
			t.Fatalf("initial=%+v approval=%d", initial, mustMemory(t, session, 0x4CBA))
		}
		return initial
	}

	t.Run("positive acceptance grants blessing cells", func(t *testing.T) {
		session := newSession(t, 0x80)
		runPrompt(t, session, "ACCEPT MY BLESSING")
		accept := uint16(0)
		result, err := session.ResumeInteractiveSelectionSeed(1000, &accept, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(result.Text, " "), "GO FORTH WITH MY BLESSING") ||
			mustMemory(t, session, 0x4CBA) != 0x85 ||
			mustMemory(t, session, 0x4CBB) != 0x02 {
			t.Fatalf("result=%+v approval=%02x modifier=%02x",
				result, mustMemory(t, session, 0x4CBA), mustMemory(t, session, 0x4CBB))
		}
	})

	t.Run("despoiler acceptance restores neutral without blessing", func(t *testing.T) {
		session := newSession(t, 0x7F)
		runPrompt(t, session, "ACCEPT MY FORGIVENESS")
		accept := uint16(0)
		result, err := session.ResumeInteractiveSelectionSeed(1000, &accept, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(result.Text, " "), "YOU ARE FORGIVEN") ||
			mustMemory(t, session, 0x4CBA) != 0x80 {
			t.Fatalf("result=%+v approval=%02x", result, mustMemory(t, session, 0x4CBA))
		}
		if _, ok := session.MemoryValue(0x4CBB); ok {
			t.Fatalf("forgiveness unexpectedly wrote 4CBB=%02x", mustMemory(t, session, 0x4CBB))
		}
	})

	for _, selection := range []uint16{1, 2} {
		name := "reject"
		if selection == 2 {
			name = "kill"
		}
		t.Run(name+" shares the weapon curse branch", func(t *testing.T) {
			session := newSession(t, 0x80)
			runPrompt(t, session, "ACCEPT MY BLESSING")
			result, err := session.ResumeInteractiveSelectionSeed(1000, &selection, nil, 1, PartyContext{})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(strings.Join(result.Text, " "), "YOUR WEAPONS WILL TWIST IN YOUR HANDS") ||
				mustMemory(t, session, 0x4CBA) != 0x76 ||
				mustMemory(t, session, 0x4CBB) != 0xFE {
				t.Fatalf("result=%+v approval=%02x modifier=%02x",
					result, mustMemory(t, session, 0x4CBA), mustMemory(t, session, 0x4CBB))
			}
		})
	}

	t.Run("flee exits without changing approval or modifier", func(t *testing.T) {
		session := newSession(t, 0x80)
		runPrompt(t, session, "ACCEPT MY BLESSING")
		flee := uint16(3)
		result, err := session.ResumeInteractiveSelectionSeed(1000, &flee, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Exited || mustMemory(t, session, 0x4CBA) != 0x80 {
			t.Fatalf("result=%+v approval=%02x", result, mustMemory(t, session, 0x4CBA))
		}
		if _, ok := session.MemoryValue(0x4CBB); ok {
			t.Fatalf("flee unexpectedly wrote 4CBB=%02x", mustMemory(t, session, 0x4CBB))
		}
	})
}

func TestRealBurialGlenRedWebChoicesAndCombatContinuation(t *testing.T) {
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
	newSession := func(t *testing.T) *BlockSession {
		t.Helper()
		session, err := NewBlockSession(all, 0x40)
		if err != nil {
			t.Fatal(err)
		}
		session.SetMemoryValue(0xC04B, 6)
		session.SetMemoryValue(0xC04C, 14)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, 0x82)
		// The normal path reaches the web after resolving terrain 0x01's
		// elven-spirit event, which sets this chapter flag.
		session.SetMemoryValue(0x4CBE, 1)
		initial, err := session.RunEntry(1, 2000, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !initial.WaitingForMenu ||
			!reflect.DeepEqual(initial.Menus[len(initial.Menus)-1].Options,
				[]string{"ENTER IT", "SPEAK", "HACK IT", "RETREAT"}) ||
			!strings.Contains(strings.Join(initial.Text, " "), "A RED WEB STRETCHES ACROSS THE PASSAGE") {
			t.Fatalf("initial=%+v", initial)
		}
		return session
	}

	t.Run("speak accepts bounded string then returns to web", func(t *testing.T) {
		session := newSession(t)
		speak := uint16(1)
		prompt, err := session.ResumeInteractiveSelectionSeed(2000, &speak, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !prompt.WaitingForString || len(prompt.StringInputRequests) != 1 ||
			prompt.StringInputRequests[0].MaxLength != 8 ||
			prompt.StringInputRequests[0].Destination != 0x7F79 ||
			!strings.Contains(strings.Join(prompt.Text, " "), "WHAT WORD DO YOU SAY") {
			t.Fatalf("prompt=%+v", prompt)
		}
		word := "Krrkik"
		brighter, err := session.ResumeInteractiveInputSeed(
			2000, nil, nil, &word, 1, PartyContext{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if brighter.WaitingForString || !brighter.WaitingForMenu ||
			!strings.Contains(strings.Join(brighter.Text, " "), "THE WEB GLOWS MORE BRIGHTLY") {
			t.Fatalf("brighter=%+v", brighter)
		}
		press := uint16(0)
		again, err := session.ResumeInteractiveSelectionSeed(2000, &press, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !again.WaitingForMenu ||
			!reflect.DeepEqual(again.Menus[len(again.Menus)-1].Options,
				[]string{"ENTER IT", "SPEAK", "HACK IT", "RETREAT"}) {
			t.Fatalf("again=%+v", again)
		}
	})

	t.Run("enter chains spiders then rakshasa and marks web complete", func(t *testing.T) {
		session := newSession(t)
		enter := uint16(0)
		spiders, err := session.ResumeInteractiveSelectionSeed(2000, &enter, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !spiders.CombatRequested ||
			!reflect.DeepEqual(spiders.MonsterSpawns, []MonsterSpawn{{MonsterID: 0x42, Count: 4, IconBlock: 0x41}}) ||
			!strings.Contains(strings.Join(spiders.Text, " "), "STUCK FAST") {
			t.Fatalf("spiders=%+v", spiders)
		}
		rakshasa, err := session.ResumeInteractiveSelectionSeed(2000, nil, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !rakshasa.PictureRequested || rakshasa.PictureBlock != 72 ||
			!rakshasa.CombatRequested ||
			!reflect.DeepEqual(rakshasa.MonsterSpawns, []MonsterSpawn{{MonsterID: 0x43, Count: 1, IconBlock: 0x43}}) ||
			!strings.Contains(strings.Join(rakshasa.Text, " "), "A RAKSHASA") {
			t.Fatalf("rakshasa=%+v", rakshasa)
		}
		done, err := session.ResumeInteractiveSelectionSeed(2000, nil, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !done.Exited || mustMemory(t, session, 0x4CBF) != 1 ||
			!strings.Contains(strings.Join(done.Text, " "), "EVENTUALLY FREE YOURSELF") {
			t.Fatalf("done=%+v", done)
		}
	})

	t.Run("hack draws spiders without rakshasa", func(t *testing.T) {
		session := newSession(t)
		hack := uint16(2)
		strike, err := session.ResumeInteractiveSelectionSeed(2000, &hack, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !strike.WaitingForMenu ||
			!strings.Contains(strings.Join(strike.Text, " "), "WIRE SNARES") {
			t.Fatalf("strike=%+v", strike)
		}
		press := uint16(0)
		spiders, err := session.ResumeInteractiveSelectionSeed(2000, &press, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !spiders.CombatRequested ||
			!reflect.DeepEqual(spiders.MonsterSpawns, []MonsterSpawn{{MonsterID: 0x42, Count: 4, IconBlock: 0x41}}) ||
			!strings.Contains(strings.Join(spiders.Text, " "), "SPIDERS INVESTIGATE THE NOISE") {
			t.Fatalf("spiders=%+v", spiders)
		}
	})

	t.Run("retreat exits without combat", func(t *testing.T) {
		session := newSession(t)
		retreat := uint16(3)
		result, err := session.ResumeInteractiveSelectionSeed(2000, &retreat, nil, 1, PartyContext{})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Exited || result.CombatRequested || result.WaitingForString {
			t.Fatalf("retreat=%+v", result)
		}
	})
}

func TestRealBurialGlenPhaseSpidersFromSolidWalls(t *testing.T) {
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
	session, err := NewBlockSession(all, 0x40)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 12)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 0)
	session.SetMemoryValue(0xC04F, 0x93)
	session.SetMemoryValue(0x4CCD, 0)

	prompt, err := session.RunEntry(1, 2000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !prompt.WaitingForMenu ||
		!strings.Contains(strings.Join(prompt.Text, " "), "SPIDERS COME OUT OF THE SOLID WALLS") ||
		mustMemory(t, session, 0x4CCD) != 0 {
		t.Fatalf("prompt=%+v", prompt)
	}
	press := uint16(0)
	fight, err := session.ResumeInteractiveSelectionSeed(2000, &press, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !fight.CombatRequested ||
		!reflect.DeepEqual(fight.MonsterSpawns,
			[]MonsterSpawn{{MonsterID: 0x41, Count: 10, IconBlock: 0x41}}) ||
		mustMemory(t, session, 0x7F82) != 8 ||
		mustMemory(t, session, 0x4C01) != 10 {
		t.Fatalf("fight=%+v", fight)
	}
	done, err := session.ResumeInteractiveSelectionSeed(2000, nil, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !done.Exited || mustMemory(t, session, 0x4CCD) != 1 {
		t.Fatalf("done=%+v", done)
	}

	revisit, err := session.RunEntry(1, 2000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !revisit.Exited || revisit.CombatRequested || revisit.WaitingForMenu {
		t.Fatalf("revisit=%+v", revisit)
	}

	glowingSession, err := NewBlockSession(all, 0x40)
	if err != nil {
		t.Fatal(err)
	}
	glowingSession.SetMemoryValue(0xC04B, 14)
	glowingSession.SetMemoryValue(0xC04C, 8)
	glowingSession.SetMemoryValue(0xC04D, 0)
	glowingSession.SetMemoryValue(0xC04F, 0x94)
	glowingSession.SetMemoryValue(0x4CCE, 0)
	glowingPrompt, err := glowingSession.RunEntry(1, 2000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !glowingPrompt.WaitingForMenu ||
		!strings.Contains(strings.Join(glowingPrompt.Text, " "),
			"GLOWING SPIDERS SKITTER FORWARD AT YOUR APPROACH") {
		t.Fatalf("glowing prompt=%+v", glowingPrompt)
	}
	glowingFight, err := glowingSession.ResumeInteractiveSelectionSeed(
		2000, &press, nil, 1, PartyContext{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !glowingFight.CombatRequested ||
		!reflect.DeepEqual(glowingFight.MonsterSpawns,
			[]MonsterSpawn{{MonsterID: 0x41, Count: 8, IconBlock: 0x41}}) ||
		mustMemory(t, glowingSession, 0x7F82) != 9 ||
		mustMemory(t, glowingSession, 0x4C01) != 8 {
		t.Fatalf("glowing fight=%+v", glowingFight)
	}
	glowingDone, err := glowingSession.ResumeInteractiveSelectionSeed(
		2000, nil, nil, 1, PartyContext{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !glowingDone.Exited || mustMemory(t, glowingSession, 0x4CCE) != 1 {
		t.Fatalf("glowing done=%+v", glowingDone)
	}
}

func TestRealBurialGlenPhaseSpiderBonePileBranches(t *testing.T) {
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

	runToBonesMenu := func(t *testing.T) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 15)
		session.SetMemoryValue(0xC04C, 8)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, 0x95)
		session.SetMemoryValue(0x4CCF, 0)
		session.SetMemoryValue(0x4CBA, 0x80)

		prompt, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "),
				"SPIDERS HAVE GATHERED A PILE OF BONES HERE") {
			t.Fatalf("prompt=%+v", prompt)
		}
		press := uint16(0)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x41, Count: 6, IconBlock: 0x41}}) ||
			mustMemory(t, session, 0x7F82) != 10 ||
			mustMemory(t, session, 0x4C01) != 6 {
			t.Fatalf("fight=%+v", fight)
		}
		menu, runErr := session.ResumeInteractiveSelectionSeed(
			3000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !menu.WaitingForMenu ||
			!reflect.DeepEqual(menu.Menus[len(menu.Menus)-1].Options,
				[]string{"LOOT", "REPLACE IN CRYPTS", "IGNORE"}) ||
			!strings.Contains(strings.Join(menu.Text, " "), "WHAT DO YOU DO WITH THE BONES") ||
			mustMemory(t, session, 0x4CCF) != 1 {
			t.Fatalf("menu=%+v", menu)
		}
		return session
	}

	t.Run("loot lowers approval and emits original treasure", func(t *testing.T) {
		session := runToBonesMenu(t)
		loot := uint16(0)
		result, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &loot, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if len(result.TreasureRequests) != 1 ||
			result.TreasureRequests[0] != (TreasureRequest{
				Coins:     [7]uint16{0, 0, 0, 0, 0, 1, 0},
				ItemBlock: 0xFF,
			}) ||
			mustMemory(t, session, 0x4CBA) != 0x7F {
			t.Fatalf("loot=%+v approval=%02x", result, mustMemory(t, session, 0x4CBA))
		}
	})

	t.Run("replace raises approval", func(t *testing.T) {
		session := runToBonesMenu(t)
		replace := uint16(1)
		result, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &replace, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !result.Exited || len(result.TreasureRequests) != 0 ||
			mustMemory(t, session, 0x4CBA) != 0x81 {
			t.Fatalf("replace=%+v approval=%02x", result, mustMemory(t, session, 0x4CBA))
		}
	})

	t.Run("ignore preserves approval", func(t *testing.T) {
		session := runToBonesMenu(t)
		ignore := uint16(2)
		result, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &ignore, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !result.Exited || len(result.TreasureRequests) != 0 ||
			mustMemory(t, session, 0x4CBA) != 0x80 {
			t.Fatalf("ignore=%+v approval=%02x", result, mustMemory(t, session, 0x4CBA))
		}
	})

	revisit := runToBonesMenu(t)
	ignore := uint16(2)
	if _, err := revisit.ResumeInteractiveSelectionSeed(3000, &ignore, nil, 1, PartyContext{}); err != nil {
		t.Fatal(err)
	}
	again, err := revisit.RunEntry(1, 3000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !again.Exited || again.CombatRequested || again.WaitingForMenu {
		t.Fatalf("revisit=%+v", again)
	}
}

func TestRealBurialGlenThriKreenDefenseFlagsAndWaves(t *testing.T) {
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
	newSession := func(t *testing.T, terrain uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 0)
		session.SetMemoryValue(0xC04C, 0)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, terrain)
		return session
	}

	t.Run("entrance force is immediate and marks 4CC8", func(t *testing.T) {
		session := newSession(t, 0x8E)
		fight, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x40, Count: 12, IconBlock: 0x40}}) ||
			!strings.Contains(strings.Join(fight.Text, " "), "BAR YOUR ENTRANCE") ||
			mustMemory(t, session, 0x7F82) != 1 ||
			mustMemory(t, session, 0x4C02) != 12 {
			t.Fatalf("entrance fight=%+v", fight)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			3000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !done.Exited || mustMemory(t, session, 0x4CC8) != 1 {
			t.Fatalf("entrance done=%+v", done)
		}
		revisit, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !revisit.Exited || revisit.CombatRequested {
			t.Fatalf("entrance revisit=%+v", revisit)
		}
	})

	t.Run("inner guards pause then mark 4CC9", func(t *testing.T) {
		session := newSession(t, 0x8F)
		prompt, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "), "GUARDS HERE PREPARE FOR COMBAT") {
			t.Fatalf("guards prompt=%+v", prompt)
		}
		press := uint16(0)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x40, Count: 6, IconBlock: 0x40}}) ||
			mustMemory(t, session, 0x7F82) != 2 ||
			mustMemory(t, session, 0x4C02) != 6 {
			t.Fatalf("guards fight=%+v", fight)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			3000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !done.Exited || mustMemory(t, session, 0x4CC9) != 1 {
			t.Fatalf("guards done=%+v", done)
		}
	})

	runBivouacStart := func(t *testing.T, session *BlockSession) RunResult {
		t.Helper()
		prompt, runErr := session.RunEntry(1, 5000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "), "PREPARE TO MAKE A STAND") {
			t.Fatalf("bivouac prompt=%+v", prompt)
		}
		press := uint16(0)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			5000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x40, Count: 12, IconBlock: 0x40}}) ||
			mustMemory(t, session, 0x7F82) != 3 ||
			mustMemory(t, session, 0x4C02) != 12 {
			t.Fatalf("bivouac fight=%+v", fight)
		}
		return fight
	}
	wantTreasure := TreasureRequest{
		Coins:     [7]uint16{0, 0, 0, 2000, 1500, 4, 6},
		ItemBlock: 0x81,
	}

	t.Run("uncleared guards produce all three waves", func(t *testing.T) {
		session := newSession(t, 0x90)
		runBivouacStart(t, session)
		reinforcements, runErr := session.ResumeInteractiveSelectionSeed(
			5000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !reinforcements.WaitingForMenu ||
			!strings.Contains(strings.Join(reinforcements.Text, " "),
				"OTHER THRI-KREEN RESPOND TO THE NOISE") {
			t.Fatalf("reinforcements=%+v", reinforcements)
		}
		press := uint16(0)
		secondFight, runErr := session.ResumeInteractiveSelectionSeed(
			5000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !secondFight.CombatRequested ||
			!reflect.DeepEqual(secondFight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x40, Count: 6, IconBlock: 0x40}}) ||
			mustMemory(t, session, 0x7F82) != 5 {
			t.Fatalf("second fight=%+v", secondFight)
		}
		stragglers, runErr := session.ResumeInteractiveSelectionSeed(
			5000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !stragglers.WaitingForMenu ||
			!strings.Contains(strings.Join(stragglers.Text, " "), "A FEW MORE STRAGGLE IN") {
			t.Fatalf("stragglers=%+v", stragglers)
		}
		thirdFight, runErr := session.ResumeInteractiveSelectionSeed(
			5000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !thirdFight.CombatRequested ||
			!reflect.DeepEqual(thirdFight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x40, Count: 6, IconBlock: 0x40}}) ||
			mustMemory(t, session, 0x7F82) != 6 {
			t.Fatalf("third fight=%+v", thirdFight)
		}
		loot, runErr := session.ResumeInteractiveSelectionSeed(
			5000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !loot.CombatRequested ||
			!reflect.DeepEqual(loot.TreasureRequests, []TreasureRequest{wantTreasure}) ||
			!strings.Contains(strings.Join(loot.Text, " "), "YOU GATHER UP SOME VALUABLES") ||
			mustMemory(t, session, 0x4CC8) != 1 ||
			mustMemory(t, session, 0x4CC9) != 1 ||
			mustMemory(t, session, 0x4CCA) != 1 {
			t.Fatalf("loot=%+v", loot)
		}
	})

	t.Run("cleared outer guards suppress matching reinforcements", func(t *testing.T) {
		session := newSession(t, 0x90)
		session.SetMemoryValue(0x4CC8, 1)
		session.SetMemoryValue(0x4CC9, 1)
		runBivouacStart(t, session)
		loot, runErr := session.ResumeInteractiveSelectionSeed(
			5000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !loot.CombatRequested || loot.WaitingForMenu ||
			!reflect.DeepEqual(loot.TreasureRequests, []TreasureRequest{wantTreasure}) ||
			!strings.Contains(strings.Join(loot.Text, " "), "YOU GATHER UP SOME VALUABLES") ||
			mustMemory(t, session, 0x4CCA) != 1 {
			t.Fatalf("suppressed reinforcement loot=%+v", loot)
		}
	})
}

func TestRealBurialGlenSpiderMausoleumApprovalBranches(t *testing.T) {
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
	newSession := func(t *testing.T, terrain, approval uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 0)
		session.SetMemoryValue(0xC04C, 0)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, terrain)
		session.SetMemoryValue(0x4CBA, approval)
		return session
	}

	t.Run("inhabited mausoleum has eight giant spiders", func(t *testing.T) {
		session := newSession(t, 0x91, 0x80)
		prompt, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "),
				"WEBS FESTOON THIS MAUSOLEUM") {
			t.Fatalf("mausoleum prompt=%+v", prompt)
		}
		press := uint16(0)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			3000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x42, Count: 8, IconBlock: 0x41}}) ||
			mustMemory(t, session, 0x7F82) != 7 ||
			mustMemory(t, session, 0x4C00) != 8 {
			t.Fatalf("mausoleum fight=%+v", fight)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			3000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !done.Exited || mustMemory(t, session, 0x4CCB) != 1 {
			t.Fatalf("mausoleum done=%+v", done)
		}
		revisit, runErr := session.RunEntry(1, 3000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !revisit.Exited || revisit.CombatRequested || revisit.WaitingForMenu {
			t.Fatalf("mausoleum revisit=%+v", revisit)
		}
	})

	runFunnelStart := func(t *testing.T, session *BlockSession) RunResult {
		t.Helper()
		prompt, runErr := session.RunEntry(1, 4000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "), "YOU SEE A FUNNEL OF WEBS") {
			t.Fatalf("funnel prompt=%+v", prompt)
		}
		return prompt
	}

	t.Run("friendly spirits warn and no leaves event repeatable", func(t *testing.T) {
		session := newSession(t, 0x92, 0x80)
		runFunnelStart(t, session)
		press := uint16(0)
		warning, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !warning.WaitingForMenu ||
			!reflect.DeepEqual(warning.Menus[len(warning.Menus)-1].Options,
				[]string{"YES", "NO"}) ||
			!strings.Contains(strings.Join(warning.Text, " "), "GUARD THEIR NEST FIERCELY") {
			t.Fatalf("warning=%+v", warning)
		}
		no := uint16(1)
		retreat, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !retreat.Exited || retreat.CombatRequested {
			t.Fatalf("retreat=%+v", retreat)
		}
		if _, found := session.MemoryValue(0x4CCC); found {
			t.Fatalf("NO unexpectedly set 4CCC=%02x", mustMemory(t, session, 0x4CCC))
		}
		revisit, runErr := session.RunEntry(1, 4000, nil)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !revisit.WaitingForMenu ||
			!strings.Contains(strings.Join(revisit.Text, " "), "YOU SEE A FUNNEL OF WEBS") {
			t.Fatalf("retreat revisit=%+v", revisit)
		}
	})

	t.Run("yes marks nest before four-spider combat", func(t *testing.T) {
		session := newSession(t, 0x92, 0x80)
		runFunnelStart(t, session)
		press := uint16(0)
		warning, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !warning.WaitingForMenu {
			t.Fatalf("warning=%+v err=%v", warning, runErr)
		}
		yes := uint16(0)
		eggs, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !eggs.WaitingForMenu ||
			!strings.Contains(strings.Join(eggs.Text, " "), "YOU CAN SEE SOME EGGS") ||
			mustMemory(t, session, 0x4CCC) != 1 {
			t.Fatalf("eggs=%+v", eggs)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x42, Count: 4, IconBlock: 0x41}}) ||
			mustMemory(t, session, 0x7F70) != 2 ||
			mustMemory(t, session, 0x7F82) != 0 ||
			mustMemory(t, session, 0x4C00) != 4 {
			t.Fatalf("nest fight=%+v", fight)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			4000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !done.Exited {
			t.Fatalf("nest done=%+v", done)
		}
	})

	t.Run("low approval receives no warning or yes-no menu", func(t *testing.T) {
		session := newSession(t, 0x92, 0x7F)
		runFunnelStart(t, session)
		press := uint16(0)
		eggs, runErr := session.ResumeInteractiveSelectionSeed(
			4000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !eggs.WaitingForMenu ||
			!strings.Contains(strings.Join(eggs.Text, " "), "YOU CAN SEE SOME EGGS") ||
			strings.Contains(strings.Join(eggs.Text, " "), "DO YOU CONTINUE") ||
			reflect.DeepEqual(eggs.Menus[len(eggs.Menus)-1].Options,
				[]string{"YES", "NO"}) ||
			mustMemory(t, session, 0x4CCC) != 1 {
			t.Fatalf("low-approval eggs=%+v", eggs)
		}
	})
}

func TestRealBurialGlenElvenCourtApprovalBranches(t *testing.T) {
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
	newSession := func(t *testing.T, terrain, approval uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 3)
		session.SetMemoryValue(0xC04C, 1)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, terrain)
		session.SetMemoryValue(0x4CBA, approval)
		return session
	}
	press := uint16(0)

	t.Run("ghostly doorway yes enters and no leaves", func(t *testing.T) {
		for choice := uint16(0); choice < 2; choice++ {
			session := newSession(t, 0x08, 0x80)
			intro, runErr := session.RunEntry(1, 7000, nil)
			if runErr != nil || intro.PictureBlock != 72 || !intro.WaitingForMenu ||
				!strings.Contains(strings.Join(intro.Text, " "), "A CRUSHED THRI-KREEN") {
				t.Fatalf("doorway intro choice=%d result=%+v err=%v", choice, intro, runErr)
			}
			prompt, runErr := session.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
			if runErr != nil || !prompt.WaitingForMenu ||
				!reflect.DeepEqual(prompt.Menus[len(prompt.Menus)-1].Options, []string{"YES", "NO"}) {
				t.Fatalf("doorway prompt choice=%d result=%+v err=%v", choice, prompt, runErr)
			}
			done, runErr := session.ResumeInteractiveSelectionSeed(7000, &choice, nil, 1, PartyContext{})
			if runErr != nil || !done.Exited {
				t.Fatalf("doorway choice=%d result=%+v err=%v", choice, done, runErr)
			}
			if choice == 0 {
				if !strings.Contains(strings.Join(done.Text, " "), "ENTER AND MEET OUR QUEEN") ||
					mustMemory(t, session, 0xC04B) != 4 ||
					mustMemory(t, session, 0xC04C) != 2 ||
					mustMemory(t, session, 0xC04D) != 3 {
					t.Fatalf("doorway YES result=%+v", done)
				}
			}
		}
	})

	t.Run("armor stairs preserve or reduce approval", func(t *testing.T) {
		for choice := uint16(0); choice < 4; choice++ {
			session := newSession(t, 0x89, 0x80)
			intro, runErr := session.RunEntry(1, 7000, nil)
			if runErr != nil {
				t.Fatal(runErr)
			}
			menu, runErr := session.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
			if runErr != nil || !menu.WaitingForMenu ||
				!reflect.DeepEqual(menu.Menus[len(menu.Menus)-1].Options,
					[]string{"GO UPSTAIRS", "TAKE ARMOR", "ATTACK", "RETREAT"}) {
				t.Fatalf("choice %d intro=%+v menu=%+v err=%v", choice, intro, menu, runErr)
			}
			branch, runErr := session.ResumeInteractiveSelectionSeed(7000, &choice, nil, 1, PartyContext{})
			if runErr != nil {
				t.Fatal(runErr)
			}
			switch choice {
			case 0:
				if !branch.WaitingForMenu ||
					!strings.Contains(strings.Join(branch.Text, " "), "ARMOR SEEMS TO BOW") {
					t.Fatalf("upstairs=%+v", branch)
				}
				done, resumeErr := session.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
				if resumeErr != nil || !done.Exited ||
					mustMemory(t, session, 0x4CBA) != 0x81 ||
					mustMemory(t, session, 0xC04B) != 2 ||
					mustMemory(t, session, 0xC04C) != 1 ||
					mustMemory(t, session, 0xC04D) != 3 {
					t.Fatalf("upstairs continuation=%+v err=%v", done, resumeErr)
				}
			case 1, 2:
				if !branch.WaitingForMenu ||
					!strings.Contains(strings.Join(branch.Text, " "), "CRUMBLES INTO RUSTY FLAKES") {
					t.Fatalf("hostile armor choice=%d result=%+v", choice, branch)
				}
				done, resumeErr := session.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
				if resumeErr != nil || !done.Exited || mustMemory(t, session, 0x4CBA) != 0x7E {
					t.Fatalf("hostile armor continuation choice=%d result=%+v approval=%02x err=%v",
						choice, done, mustMemory(t, session, 0x4CBA), resumeErr)
				}
			case 3:
				if !branch.Exited || mustMemory(t, session, 0x4CBA) != 0x80 {
					t.Fatalf("retreat=%+v approval=%02x", branch, mustMemory(t, session, 0x4CBA))
				}
				if _, found := session.MemoryValue(0x4CC4); found {
					t.Fatal("retreat unexpectedly consumed terrain 89h")
				}
				continue
			}
			if mustMemory(t, session, 0x4CC4) != 1 {
				t.Fatalf("choice %d did not set 4CC4", choice)
			}
		}
	})

	t.Run("court greeting or fourteen-monster punishment", func(t *testing.T) {
		friendly := newSession(t, 0x8A, 0x80)
		result, runErr := friendly.RunEntry(1, 7000, nil)
		if runErr != nil || !result.WaitingForMenu ||
			!strings.Contains(strings.Join(result.Text, " "), "COURT GIVES GREETINGS") ||
			mustMemory(t, friendly, 0x4CC5) != 1 {
			t.Fatalf("friendly=%+v err=%v", result, runErr)
		}

		hostile := newSession(t, 0x8A, 0x7F)
		result, runErr = hostile.RunEntry(1, 7000, nil)
		if runErr != nil || !result.WaitingForMenu ||
			!strings.Contains(strings.Join(result.Text, " "), "DESPOILERS SHALL FIGHT DESPOILERS") {
			t.Fatalf("hostile prompt=%+v err=%v", result, runErr)
		}
		fight, runErr := hostile.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x42, Count: 6, IconBlock: 0x41},
				{MonsterID: 0x41, Count: 4, IconBlock: 0x41},
				{MonsterID: 0x40, Count: 4, IconBlock: 0x40},
			}) {
			t.Fatalf("hostile fight=%+v err=%v", fight, runErr)
		}
	})

	t.Run("queen rewards defenders or collapses tower", func(t *testing.T) {
		friendly := newSession(t, 0x8B, 0x80)
		result, runErr := friendly.RunEntry(1, 7000, nil)
		if runErr != nil || result.PictureBlock != 72 || !result.WaitingForMenu ||
			!strings.Contains(strings.Join(result.Text, " "), "MY HEART REJOICES") ||
			mustMemory(t, friendly, 0x4CC6) != 1 {
			t.Fatalf("friendly queen=%+v err=%v", result, runErr)
		}
		reward, runErr := friendly.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
		if runErr != nil || len(reward.TreasureRequests) != 1 ||
			reward.TreasureRequests[0].Coins != [7]uint16{0, 0, 0, 0, 0, 12, 8} ||
			reward.TreasureRequests[0].ItemBlock != 0x41 {
			t.Fatalf("friendly reward=%+v err=%v", reward, runErr)
		}

		for choice := uint16(0); choice < 2; choice++ {
			hostile := newSession(t, 0x8B, 0x7F)
			offer, offerErr := hostile.RunEntry(1, 7000, nil)
			if offerErr != nil || !offer.WaitingForMenu ||
				!reflect.DeepEqual(offer.Menus[len(offer.Menus)-1].Options, []string{"YES", "NO"}) ||
				mustMemory(t, hostile, 0x4CBA) != 0x7A ||
				mustMemory(t, hostile, 0x4CC6) != 1 {
				t.Fatalf("hostile offer choice=%d result=%+v err=%v", choice, offer, offerErr)
			}
			next, resumeErr := hostile.ResumeInteractiveSelectionSeed(7000, &choice, nil, 1, PartyContext{})
			if resumeErr != nil {
				t.Fatal(resumeErr)
			}
			if choice == 0 {
				if len(next.TreasureRequests) != 1 ||
					next.TreasureRequests[0].Coins != [7]uint16{0, 0, 0, 0, 0, 4, 2} ||
					next.TreasureRequests[0].ItemBlock != 0x40 {
					t.Fatalf("hostile accepted reward=%+v", next)
				}
			} else if !next.WaitingForMenu ||
				!strings.Contains(strings.Join(next.Text, " "), "THEN LIE WITH US") {
				t.Fatalf("hostile refusal=%+v", next)
			}
			collapse, collapseErr := hostile.ResumeInteractiveSelectionSeed(7000, &press, nil, 1, PartyContext{})
			if collapseErr != nil || !collapse.Exited ||
				mustMemory(t, hostile, 0xC04B) != 5 ||
				mustMemory(t, hostile, 0xC04C) != 2 ||
				mustMemory(t, hostile, 0xC04D) != 3 {
				t.Fatalf("hostile collapse=%+v err=%v", collapse, collapseErr)
			}
		}
	})
}

func TestRealBurialGlenRedPlumeTrapBranches(t *testing.T) {
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
	newSession := func(t *testing.T, approval uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 13)
		session.SetMemoryValue(0xC04C, 6)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, 0x05)
		session.SetMemoryValue(0x4CBA, approval)
		return session
	}
	resume := func(t *testing.T, session *BlockSession, choice uint16) RunResult {
		t.Helper()
		result, runErr := session.ResumeInteractiveSelectionSeed(
			10000, &choice, nil, 1, PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		return result
	}
	press := uint16(0)

	t.Run("journal offer and three decisions", func(t *testing.T) {
		for decision := uint16(0); decision < 3; decision++ {
			session := newSession(t, 0x85)
			intro, runErr := session.RunEntry(1, 10000, nil)
			if runErr != nil || !intro.WaitingForMenu ||
				!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
					[]string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) ||
				// ⚠ 三句旁白依距離挑（spec 1144）。這一處的距離上限是 2，所以演的是
				// 第三句（遠距）；`HE MAKES A GESTURE OF FRIENDSHIP` 是距離 0 那一句。
				intro.Menus[len(intro.Menus)-1].Prompt != "YOU SPOT A LONE RED PLUME" {
				t.Fatalf("decision=%d intro=%+v err=%v", decision, intro, runErr)
			}
			journal := resume(t, session, 1)
			if !journal.WaitingForMenu ||
				!strings.Contains(strings.Join(journal.Text, " "), "JOURNAL ENTRY 33") ||
				!reflect.DeepEqual(journal.Menus[len(journal.Menus)-1].Options,
					[]string{"AGREE", "REFUSE PAYMENT", "DISAGREE"}) ||
				mustMemory(t, session, 0x4CC2) != 1 {
				t.Fatalf("decision=%d journal=%+v", decision, journal)
			}
			branch := resume(t, session, decision)
			if decision < 2 {
				if !branch.WaitingForMenu ||
					!strings.Contains(strings.Join(branch.Text, " "), "FOLLOW ME") {
					t.Fatalf("decision=%d branch=%+v", decision, branch)
				}
			} else if !branch.WaitingForMenu ||
				!strings.Contains(strings.Join(branch.Text, " "), "SCREAM PIERCES THE AIR") {
				t.Fatalf("decision=%d branch=%+v", decision, branch)
			}
		}
	})

	t.Run("agree warning two arrows and seven monsters", func(t *testing.T) {
		session := newSession(t, 0x85)
		if _, runErr := session.RunEntry(1, 10000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		resume(t, session, 1)
		resume(t, session, 0)
		prompt := resume(t, session, press)
		if !prompt.WaitingForMenu ||
			!reflect.DeepEqual(prompt.Menus[len(prompt.Menus)-1].Options, []string{"YES", "NO"}) {
			t.Fatalf("continue prompt=%+v", prompt)
		}
		reveal := resume(t, session, 0)
		if reveal.PictureBlock != 0x43 ||
			!strings.Contains(strings.Join(reveal.Text, " "), "HIS SHAPE CHANGES") {
			t.Fatalf("reveal=%+v", reveal)
		}
		damage := resume(t, session, press)
		if !reflect.DeepEqual(damage.DamageRequests, []DamageRequest{{
			Flags: 2, DiceCount: 1, DiceSize: 6, Bonus: 6, SaveFlags: 0x35,
		}}) {
			t.Fatalf("damage=%+v", damage)
		}
		fight := resume(t, session, press)
		if !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x41, Count: 6, IconBlock: 0x41},
				{MonsterID: 0x49, Count: 1, IconBlock: 0x43},
			}) {
			t.Fatalf("fight=%+v", fight)
		}
	})

	t.Run("warning and investigation can both be declined", func(t *testing.T) {
		agree := newSession(t, 0x85)
		agree.RunEntry(1, 10000, nil)
		resume(t, agree, 1)
		resume(t, agree, 0)
		resume(t, agree, press)
		if done := resume(t, agree, 1); !done.Exited || done.CombatRequested {
			t.Fatalf("declined trap=%+v", done)
		}

		disagree := newSession(t, 0x85)
		disagree.RunEntry(1, 10000, nil)
		resume(t, disagree, 1)
		resume(t, disagree, 2)
		if done := resume(t, disagree, 1); !done.Exited || done.CombatRequested {
			t.Fatalf("declined investigation=%+v", done)
		}
	})
}

func TestRealBurialGlenJournal56AndMoreRuinsExit(t *testing.T) {
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
	newSession := func(t *testing.T) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		return session
	}

	t.Run("terrain 0C is one-shot and wait or parlay unlock journal 56", func(t *testing.T) {
		for _, choice := range []uint16{1, 3} {
			session := newSession(t)
			session.SetMemoryValue(0xC04B, 4)
			session.SetMemoryValue(0xC04C, 8)
			session.SetMemoryValue(0xC04D, 3)
			session.SetMemoryValue(0xC04F, 0x0C)
			intro, runErr := session.RunEntry(1, 10000, nil)
			if runErr != nil || !intro.WaitingForMenu ||
				intro.Menus[len(intro.Menus)-1].Prompt !=
					"A FIGURE APPEARS FROM THE SHADOWS. 'HAIL BONDED ONES!'" ||
				!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
					[]string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) ||
				mustMemory(t, session, 0x4CC7) != 1 {
				t.Fatalf("choice=%d intro=%+v err=%v", choice, intro, runErr)
			}
			journal, runErr := session.ResumeInteractiveSelectionSeed(
				10000, &choice, nil, 1, PartyContext{},
			)
			if runErr != nil || !journal.WaitingForMenu ||
				!strings.Contains(strings.Join(journal.Text, " "), "JOURNAL ENTRY 56") ||
				!strings.Contains(strings.Join(journal.Text, " "), "HURRY ON") {
				t.Fatalf("choice=%d journal=%+v err=%v", choice, journal, runErr)
			}
			press := uint16(0)
			done, runErr := session.ResumeInteractiveSelectionSeed(
				10000, &press, nil, 1, PartyContext{},
			)
			if runErr != nil || !done.Exited {
				t.Fatalf("choice=%d done=%+v err=%v", choice, done, runErr)
			}
			revisit, runErr := session.RunEntry(1, 10000, nil)
			if runErr != nil || !revisit.Exited || revisit.WaitingForMenu ||
				len(revisit.Text) != 0 {
				t.Fatalf("choice=%d revisit=%+v err=%v", choice, revisit, runErr)
			}
		}
	})

	t.Run("east boundary selects path woods or turn back", func(t *testing.T) {
		for choice := uint16(0); choice < 3; choice++ {
			session := newSession(t)
			session.SetMemoryValue(0xC04B, 15)
			session.SetMemoryValue(0xC04C, 6)
			session.SetMemoryValue(0xC04D, 1)
			session.SetMemoryValue(0xC04F, 0)
			session.SetMemoryValue(0x7ED5, 1)
			intro, runErr := session.RunEntry(0, 10000, nil)
			if runErr != nil || !intro.WaitingForMenu ||
				!strings.Contains(strings.Join(intro.Text, " "), "HEADING TOWARD MORE RUINS") ||
				!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
					[]string{"PATH", "WOODS", "TURN BACK"}) {
				t.Fatalf("choice=%d intro=%+v err=%v", choice, intro, runErr)
			}
			done, runErr := session.ResumeInteractiveSelectionSeed(
				10000, &choice, nil, 1, PartyContext{},
			)
			if runErr != nil || !done.Exited {
				t.Fatalf("choice=%d done=%+v err=%v", choice, done, runErr)
			}
			if choice == 2 {
				if session.CurrentBlockID() != 0x40 ||
					mustMemory(t, session, 0x7EC9) != 0xFF {
					t.Fatalf("turn-back block=%02x result=%+v", session.CurrentBlockID(), done)
				}
				continue
			}
			wantY := uint16(12)
			if choice == 1 {
				wantY = 6
			}
			if session.CurrentBlockID() != 0x42 ||
				mustMemory(t, session, 0xC04B) != 0 ||
				mustMemory(t, session, 0xC04C) != wantY ||
				!strings.Contains(strings.Join(done.Text, " "), "HELM OF DRAGONS REPORTS") {
				t.Fatalf("choice=%d block=%02x result=%+v coords=(%d,%d)",
					choice, session.CurrentBlockID(), done,
					mustMemory(t, session, 0xC04B), mustMemory(t, session, 0xC04C))
			}
		}
	})
}

func TestRealOuterRuinsTirsheyaAlliance(t *testing.T) {
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
	session, err := NewBlockSession(all, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 1)
	session.SetMemoryValue(0xC04C, 12)
	session.SetMemoryValue(0xC04D, 1)
	session.SetMemoryValue(0xC04F, 0x01)

	intro, err := session.RunEntry(1, 30000, nil)
	if err != nil || !intro.WaitingForMenu ||
		!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
			[]string{"WAIT", "ATTACK", "FLEE"}) ||
		!strings.Contains(strings.Join(intro.Text, " "), "A RAKSHASA WITH MATTED FUR") ||
		mustMemory(t, session, 0x4CD0) != 1 {
		t.Fatalf("intro=%+v err=%v 4CD0=%d", intro, err, mustMemory(t, session, 0x4CD0))
	}
	wait := uint16(0)
	tale, err := session.ResumeInteractiveSelectionSeed(
		30000, &wait, nil, 1, PartyContext{},
	)
	if err != nil || !tale.WaitingForMenu ||
		!reflect.DeepEqual(tale.Menus[len(tale.Menus)-1].Options,
			[]string{"YES", "NO"}) ||
		!strings.Contains(strings.Join(tale.Text, " "), "JOURNAL ENTRY 5") {
		t.Fatalf("tale=%+v err=%v", tale, err)
	}
	yes := uint16(0)
	guards, err := session.ResumeInteractiveSelectionSeed(
		30000, &yes, nil, 1, PartyContext{},
	)
	if err != nil || !guards.WaitingForMenu ||
		!strings.Contains(strings.Join(guards.Text, " "), "THE ENTRANCE IS GUARDED") {
		t.Fatalf("guards=%+v err=%v", guards, err)
	}
	firstFight, err := session.ResumeInteractiveSelectionSeed(
		30000, &yes, nil, 1, PartyContext{},
	)
	if err != nil || !firstFight.CombatRequested ||
		!reflect.DeepEqual(firstFight.MonsterSpawns, []MonsterSpawn{
			{MonsterID: 0x44, Count: 5, IconBlock: 0x44},
			{MonsterID: 0x45, Count: 5, IconBlock: 0x45},
		}) {
		t.Fatalf("first fight=%+v err=%v", firstFight, err)
	}

	arrival, err := session.ResumeInteractiveSelectionSeed(
		30000, nil, nil, 1, PartyContext{},
	)
	if err != nil || !arrival.WaitingForMenu ||
		!strings.Contains(strings.Join(arrival.Text, " "), "ANOTHER RAKSHASA") {
		t.Fatalf("arrival=%+v err=%v", arrival, err)
	}
	press := uint16(0)
	ultimatum, err := session.ResumeInteractiveSelectionSeed(
		30000, &press, nil, 1, PartyContext{},
	)
	if err != nil || !ultimatum.WaitingForMenu ||
		!reflect.DeepEqual(ultimatum.Menus[len(ultimatum.Menus)-1].Options,
			[]string{"TIRSHEYA", "BEYRHA", "FLEE"}) {
		t.Fatalf("ultimatum=%+v err=%v", ultimatum, err)
	}
	attackBeyrha := uint16(1)
	threat, err := session.ResumeInteractiveSelectionSeed(
		30000, &attackBeyrha, nil, 1, PartyContext{},
	)
	if err != nil || !threat.WaitingForMenu ||
		!strings.Contains(strings.Join(threat.Text, " "), "THANK YOU AT DINNER") {
		t.Fatalf("threat=%+v err=%v", threat, err)
	}
	secondFight, err := session.ResumeInteractiveSelectionSeed(
		30000, &press, nil, 1, PartyContext{},
	)
	if err != nil || !secondFight.CombatRequested ||
		!reflect.DeepEqual(secondFight.MonsterSpawns, []MonsterSpawn{
			{MonsterID: 0x43, Count: 1, IconBlock: 0x43, PartyMask: 1},
			{MonsterID: 0x44, Count: 6, IconBlock: 0x44},
			{MonsterID: 0x45, Count: 6, IconBlock: 0x45},
		}) ||
		!reflect.DeepEqual(secondFight.NPCRequests,
			[]NPCRequest{{ID: 0x43, Morale: 100}}) ||
		!reflect.DeepEqual(secondFight.CombatTeamWrites,
			[]CombatTeamWrite{{TeamListIndex: 8, Value: 0x80}}) {
		t.Fatalf("second fight=%+v err=%v", secondFight, err)
	}
	done, err := session.ResumeInteractiveSelectionSeed(
		30000, nil, nil, 1, PartyContext{},
	)
	if err != nil || !done.Exited || len(done.DumpRequests) != 1 ||
		mustMemory(t, session, 0x4CD1) != 1 {
		t.Fatalf("done=%+v err=%v 4CD1=%d", done, err, mustMemory(t, session, 0x4CD1))
	}
	revisit, err := session.RunEntry(1, 30000, nil)
	if err != nil || !revisit.Exited || revisit.WaitingForMenu ||
		revisit.CombatRequested || len(revisit.Text) != 0 {
		t.Fatalf("revisit=%+v err=%v", revisit, err)
	}
}

func TestRealOuterRuinsStorehouse(t *testing.T) {
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
	newSession := func() *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x42)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0x4CD1, 0)
		session.SetMemoryValue(0x4CD2, 0)
		return session
	}

	t.Run("entrance flee or fight converges on cleared flag", func(t *testing.T) {
		fleeSession := newSession()
		fleeSession.SetMemoryValue(0xC04F, 0x02)
		intro, runErr := fleeSession.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "GUARD THE ENTRANCE") ||
			!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
				[]string{"YES", "NO"}) {
			t.Fatalf("entrance intro=%+v err=%v", intro, runErr)
		}
		yes := uint16(0)
		fled, runErr := fleeSession.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !fled.Exited || mustMemory(t, fleeSession, 0x4CD1) != 0 {
			t.Fatalf("fled=%+v err=%v 4CD1=%d",
				fled, runErr, mustMemory(t, fleeSession, 0x4CD1))
		}

		fightSession := newSession()
		fightSession.SetMemoryValue(0xC04F, 0x02)
		if _, runErr = fightSession.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		no := uint16(1)
		fight, runErr := fightSession.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x44, Count: 6, IconBlock: 0x44},
				{MonsterID: 0x45, Count: 6, IconBlock: 0x45},
			}) {
			t.Fatalf("entrance fight=%+v err=%v", fight, runErr)
		}
		done, runErr := fightSession.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited || mustMemory(t, fightSession, 0x4CD1) != 1 {
			t.Fatalf("entrance done=%+v err=%v 4CD1=%d",
				done, runErr, mustMemory(t, fightSession, 0x4CD1))
		}
		revisit, runErr := fightSession.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || revisit.WaitingForMenu ||
			revisit.CombatRequested || len(revisit.Text) != 0 {
			t.Fatalf("entrance revisit=%+v err=%v", revisit, runErr)
		}
	})

	t.Run("active search yields exact treasure once", func(t *testing.T) {
		session := newSession()
		session.SetMemoryValue(0xC04F, 0x83)
		ordinary, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !ordinary.Exited ||
			!strings.Contains(strings.Join(ordinary.Text, " "), "FOODSTUFFS") ||
			mustMemory(t, session, 0x4CD2) != 0 {
			t.Fatalf("ordinary storehouse=%+v err=%v 4CD2=%d",
				ordinary, runErr, mustMemory(t, session, 0x4CD2))
		}

		session.SetMemoryValue(0x7ECA, 1)
		search, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !search.CombatRequested ||
			!strings.Contains(strings.Join(search.Text, " "), "FEW VALUABLES") ||
			!reflect.DeepEqual(search.TreasureRequests, []TreasureRequest{{
				Coins:     [7]uint16{0, 0, 0, 2000, 1500, 8, 8},
				ItemBlock: 0x82,
			}}) {
			t.Fatalf("storehouse search=%+v err=%v", search, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited || mustMemory(t, session, 0x4CD2) != 1 {
			t.Fatalf("storehouse treasure continuation=%+v err=%v 4CD2=%d",
				done, runErr, mustMemory(t, session, 0x4CD2))
		}
		repeat, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !repeat.Exited || repeat.CombatRequested ||
			len(repeat.Text) != 0 || len(repeat.TreasureRequests) != 0 {
			t.Fatalf("repeated storehouse search=%+v err=%v", repeat, runErr)
		}
	})
}

func TestRealOuterRuinsFugitiveAndCache(t *testing.T) {
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
	newSession := func() *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x42)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		for address := uint16(0x4CD3); address <= 0x4CD5; address++ {
			session.SetMemoryValue(address, 0)
		}
		return session
	}

	t.Run("rescue earns dying clue", func(t *testing.T) {
		session := newSession()
		session.SetMemoryValue(0xC04F, 0x04)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "DO YOU GO TO THE RESCUE") ||
			!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
				[]string{"YES", "NO"}) ||
			mustMemory(t, session, 0x4CD3) != 1 {
			t.Fatalf("fugitive intro=%+v err=%v 4CD3=%d",
				intro, runErr, mustMemory(t, session, 0x4CD3))
		}
		yes := uint16(0)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns,
				[]MonsterSpawn{{MonsterID: 0x44, Count: 6, IconBlock: 0x44}}) {
			t.Fatalf("rescue fight=%+v err=%v", fight, runErr)
		}
		clue, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !clue.PictureRequested ||
			clue.PictureBlock != 0x40 || !clue.PictureHeadBlockSet ||
			clue.PictureHeadBlock != 0x40 || !clue.WaitingForMenu ||
			!strings.Contains(strings.Join(clue.Text, " "), "RUINED BUILDING") {
			t.Fatalf("dying clue=%+v err=%v", clue, runErr)
		}
		press := uint16(0)
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited ||
			mustMemory(t, session, 0x4CD4) != 1 ||
			mustMemory(t, session, 0x4CD5) != 1 {
			t.Fatalf("rescue done=%+v err=%v flags=%d/%d",
				done, runErr, mustMemory(t, session, 0x4CD4),
				mustMemory(t, session, 0x4CD5))
		}
	})

	t.Run("decline may attack or leave remains", func(t *testing.T) {
		for attack := uint16(0); attack < 2; attack++ {
			session := newSession()
			session.SetMemoryValue(0xC04F, 0x04)
			if _, runErr := session.RunEntry(1, 30000, nil); runErr != nil {
				t.Fatal(runErr)
			}
			no := uint16(1)
			killed, runErr := session.ResumeInteractiveSelectionSeed(
				30000, &no, nil, 1, PartyContext{},
			)
			if runErr != nil || !killed.WaitingForMenu ||
				!strings.Contains(strings.Join(killed.Text, " "), "TEAR HIM TO SHREDS") {
				t.Fatalf("attack=%d killed=%+v err=%v", attack, killed, runErr)
			}
			result, runErr := session.ResumeInteractiveSelectionSeed(
				30000, &attack, nil, 1, PartyContext{},
			)
			if runErr != nil {
				t.Fatal(runErr)
			}
			if attack == 0 {
				if !result.CombatRequested ||
					!reflect.DeepEqual(result.MonsterSpawns,
						[]MonsterSpawn{{MonsterID: 0x44, Count: 6, IconBlock: 0x44}}) {
					t.Fatalf("decline then attack=%+v", result)
				}
				result, runErr = session.ResumeInteractiveSelectionSeed(
					30000, nil, nil, 1, PartyContext{},
				)
			} else if !result.WaitingForMenu ||
				!strings.Contains(strings.Join(result.Text, " "), "LEAVE THE REMAINS") {
				t.Fatalf("decline and leave=%+v", result)
			}
			if runErr != nil || !result.WaitingForMenu {
				t.Fatalf("attack=%d remains=%+v err=%v", attack, result, runErr)
			}
			press := uint16(0)
			done, runErr := session.ResumeInteractiveSelectionSeed(
				30000, &press, nil, 1, PartyContext{},
			)
			if runErr != nil || !done.Exited || mustMemory(t, session, 0x4CD5) != 0 {
				t.Fatalf("attack=%d done=%+v err=%v 4CD5=%d",
					attack, done, runErr, mustMemory(t, session, 0x4CD5))
			}
		}
	})

	t.Run("remains are one shot", func(t *testing.T) {
		session := newSession()
		session.SetMemoryValue(0xC04F, 0x05)
		remains, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !remains.WaitingForMenu ||
			!strings.Contains(strings.Join(remains.Text, " "), "NOTHING OF VALUE") ||
			mustMemory(t, session, 0x4CD4) != 1 {
			t.Fatalf("remains=%+v err=%v 4CD4=%d",
				remains, runErr, mustMemory(t, session, 0x4CD4))
		}
		press := uint16(0)
		if _, runErr = session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		); runErr != nil {
			t.Fatal(runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
			t.Fatalf("remains revisit=%+v err=%v", revisit, runErr)
		}
	})

	t.Run("active search consumes cache clue", func(t *testing.T) {
		session := newSession()
		session.SetMemoryValue(0xC04F, 0x06)
		session.SetMemoryValue(0x4CD5, 1)
		ordinary, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !ordinary.Exited || len(ordinary.Text) != 0 {
			t.Fatalf("ordinary cache cell=%+v err=%v", ordinary, runErr)
		}
		session.SetMemoryValue(0x7ECA, 1)
		found, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !found.WaitingForMenu ||
			!strings.Contains(strings.Join(found.Text, " "), "LOCATE A CACHE") {
			t.Fatalf("cache found=%+v err=%v", found, runErr)
		}
		press := uint16(0)
		treasure, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !treasure.CombatRequested ||
			!reflect.DeepEqual(treasure.TreasureRequests, []TreasureRequest{{
				Coins:     [7]uint16{0, 0, 1, 0, 0, 0, 0},
				ItemBlock: 0x43,
			}}) {
			t.Fatalf("cache treasure=%+v err=%v", treasure, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited || mustMemory(t, session, 0x4CD5) != 0 {
			t.Fatalf("cache done=%+v err=%v 4CD5=%d",
				done, runErr, mustMemory(t, session, 0x4CD5))
		}
		repeat, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !repeat.Exited || repeat.CombatRequested ||
			len(repeat.TreasureRequests) != 0 {
			t.Fatalf("cache repeat=%+v err=%v", repeat, runErr)
		}
	})
}

func TestRealOuterRuinsNamelessAndBrushAmbush(t *testing.T) {
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
	newSession := func(terrain uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x42)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, terrain)
		return session
	}
	press := uint16(0)

	t.Run("Nameless warning is one shot with exact portrait", func(t *testing.T) {
		session := newSession(0x07)
		warning, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !warning.WaitingForMenu ||
			!warning.PictureRequested || warning.PictureBlock != 0x46 ||
			!warning.PictureHeadBlockSet || warning.PictureHeadBlock != 0x43 ||
			!strings.Contains(strings.Join(warning.Text, " "),
				"NAMELESS SLIDES OUT OF THE SHADOWS") ||
			mustMemory(t, session, 0x4C06) != 1 {
			t.Fatalf("Nameless warning=%+v err=%v 4C06=%d",
				warning, runErr, mustMemory(t, session, 0x4C06))
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("Nameless continuation=%+v err=%v", done, runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || revisit.WaitingForMenu ||
			len(revisit.Text) != 0 {
			t.Fatalf("Nameless revisit=%+v err=%v", revisit, runErr)
		}
	})

	t.Run("declining rescue leaves the victim to the hound", func(t *testing.T) {
		session := newSession(0x08)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!reflect.DeepEqual(intro.Menus[len(intro.Menus)-1].Options,
				[]string{"YES", "NO"}) ||
			!strings.Contains(strings.Join(intro.Text, " "), "SMALL CHILD") ||
			mustMemory(t, session, 0x4CD6) != 1 {
			t.Fatalf("brush intro=%+v err=%v 4CD6=%d",
				intro, runErr, mustMemory(t, session, 0x4CD6))
		}
		no := uint16(1)
		victim, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !victim.WaitingForMenu ||
			!strings.Contains(strings.Join(victim.Text, " "),
				"SOMETHING BLOODY IN ITS MOUTH") {
			t.Fatalf("brush decline=%+v err=%v", victim, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited || done.CombatRequested {
			t.Fatalf("brush decline continuation=%+v err=%v", done, runErr)
		}
	})

	t.Run("rescue triggers rocks and exact three-group ambush", func(t *testing.T) {
		session := newSession(0x08)
		if _, runErr := session.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		yes := uint16(0)
		rescue, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !rescue.WaitingForMenu ||
			!strings.Contains(strings.Join(rescue.Text, " "), "CUT IT DOWN") ||
			!reflect.DeepEqual(rescue.CallAddresses, []uint16{0x2E10}) {
			t.Fatalf("brush rescue=%+v err=%v", rescue, runErr)
		}
		ambush, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !ambush.WaitingForMenu ||
			!strings.Contains(strings.Join(ambush.Text, " "),
				"SUCH KINDNESS SHOULD BE REWARDED") {
			t.Fatalf("brush ambush=%+v err=%v", ambush, runErr)
		}
		rocks, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !rocks.WaitingForMenu ||
			!reflect.DeepEqual(rocks.DamageRequests, []DamageRequest{{
				Flags: 0x0C, DiceCount: 2, DiceSize: 8, Bonus: 0, SaveFlags: 0x34,
			}}) ||
			!strings.Contains(strings.Join(rocks.Text, " "),
				"MONSTERS THEN LEAP DOWN TO ATTACK") {
			t.Fatalf("brush rocks=%+v err=%v", rocks, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x44, Count: 5, IconBlock: 0x44},
				{MonsterID: 0x45, Count: 5, IconBlock: 0x45},
				{MonsterID: 0x43, Count: 1, IconBlock: 0x43},
			}) {
			t.Fatalf("brush fight=%+v err=%v", fight, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("brush victory continuation=%+v err=%v", done, runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
			t.Fatalf("brush revisit=%+v err=%v", revisit, runErr)
		}
	})

	t.Run("bloodstained brush remains descriptive", func(t *testing.T) {
		session := newSession(0x09)
		for visit := 0; visit < 2; visit++ {
			result, runErr := session.RunEntry(1, 30000, nil)
			if runErr != nil || !result.Exited || result.WaitingForMenu ||
				!strings.Contains(strings.Join(result.Text, " "),
					"BLOODSTAINS MARK THE LEAVES") {
				t.Fatalf("bloodstains visit=%d result=%+v err=%v",
					visit, result, runErr)
			}
		}
	})
}

func TestRealOuterRuinsRakshasaRoomsAndSewer(t *testing.T) {
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
	newSession := func(terrain uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x42)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 9)
		session.SetMemoryValue(0xC04C, 2)
		session.SetMemoryValue(0xC04D, 3)
		session.SetMemoryValue(0xC04F, terrain)
		return session
	}
	press := uint16(0)

	t.Run("margoyle doorway is optional but attack springs exact trap", func(t *testing.T) {
		decline := newSession(0x0B)
		intro, runErr := decline.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "),
				"TWO MARGOYLES ARE TORTURING A SMALL ANIMAL") {
			t.Fatalf("doorway intro=%+v err=%v", intro, runErr)
		}
		no := uint16(1)
		left, runErr := decline.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !left.Exited || left.CombatRequested {
			t.Fatalf("doorway decline=%+v err=%v", left, runErr)
		}
		repeat, runErr := decline.RunEntry(1, 30000, nil)
		if runErr != nil || !repeat.WaitingForMenu ||
			!strings.Contains(strings.Join(repeat.Text, " "), "TWO MARGOYLES") {
			t.Fatalf("doorway decline repeat=%+v err=%v", repeat, runErr)
		}

		session := newSession(0x0B)
		if _, runErr = session.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		yes := uint16(0)
		collapse, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !collapse.WaitingForMenu ||
			!strings.Contains(strings.Join(collapse.Text, " "),
				"THE DOORWAY COLLAPSES ONTO YOU") ||
			!reflect.DeepEqual(collapse.CallAddresses, []uint16{0x2E10}) ||
			mustMemory(t, session, 0x4CD8) != 1 ||
			mustMemory(t, session, 0xC04B) != 10 ||
			mustMemory(t, session, 0xC04C) != 2 ||
			mustMemory(t, session, 0xC04D) != 0 {
			t.Fatalf("doorway collapse=%+v err=%v", collapse, runErr)
		}
		buried, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !buried.WaitingForMenu ||
			!reflect.DeepEqual(buried.DamageRequests, []DamageRequest{{
				Flags: 0xC0, DiceCount: 3, DiceSize: 10, Bonus: 0, SaveFlags: 1,
			}}) ||
			!strings.Contains(strings.Join(buried.Text, " "), "DO YOU PLAY DEAD") {
			t.Fatalf("doorway buried=%+v err=%v", buried, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!strings.Contains(strings.Join(fight.Text, " "), "YOU SURPRISE HIM") ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{{
				MonsterID: 0x43, Count: 1, IconBlock: 0x43,
			}}) {
			t.Fatalf("doorway fight=%+v err=%v", fight, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("doorway victory=%+v err=%v", done, runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
			t.Fatalf("doorway revisit=%+v err=%v", revisit, runErr)
		}

		retreat := newSession(0x0B)
		if _, runErr = retreat.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		if _, runErr = retreat.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		); runErr != nil {
			t.Fatal(runErr)
		}
		if _, runErr = retreat.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		); runErr != nil {
			t.Fatal(runErr)
		}
		escaped, runErr := retreat.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !escaped.WaitingForMenu || escaped.CombatRequested ||
			!strings.Contains(strings.Join(escaped.Text, " "),
				"COLLAPSE FAILED TO FINISH YOU OFF") {
			t.Fatalf("doorway rakshasa retreat=%+v err=%v", escaped, runErr)
		}
	})

	t.Run("sewer grate branches and enters block 43 kitchen", func(t *testing.T) {
		session := newSession(0x0C)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "),
				"A LONE MARGOYLE SKITTERS AWAY") ||
			mustMemory(t, session, 0x4CD9) != 1 {
			t.Fatalf("sewer intro=%+v err=%v", intro, runErr)
		}
		yes := uint16(0)
		fled, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !fled.WaitingForMenu ||
			!strings.Contains(strings.Join(fled.Text, " "), "RUNS DOWN THROUGH THE SEWER") {
			t.Fatalf("sewer margoyle flee=%+v err=%v", fled, runErr)
		}
		grate, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !grate.WaitingForMenu ||
			!strings.Contains(strings.Join(grate.Text, " "), "DO YOU WANT TO ENTER") {
			t.Fatalf("sewer grate=%+v err=%v", grate, runErr)
		}
		warning, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !warning.WaitingForMenu ||
			!strings.Contains(strings.Join(warning.Text, " "), "GREAT DANGER LIES BEFORE YOU") {
			t.Fatalf("sewer warning=%+v err=%v", warning, runErr)
		}
		kitchen, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || session.CurrentBlockID() != 0x43 ||
			!kitchen.WaitingForMenu ||
			!kitchen.LoadFilesRequested ||
			kitchen.LoadFiles != [3]uint16{0x43, 2, 0xFF} ||
			!kitchen.LoadPiecesRequested ||
			kitchen.LoadPieces != [3]uint16{13, 16, 3} ||
			!strings.Contains(strings.Join(kitchen.Text, " "),
				"THE SEWER ENDS IN A DARKENED KITCHEN") {
			t.Fatalf("sewer kitchen block=%02x result=%+v err=%v",
				session.CurrentBlockID(), kitchen, runErr)
		}

		stay := newSession(0x0C)
		if _, runErr = stay.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		no := uint16(1)
		killed, runErr := stay.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !killed.WaitingForMenu ||
			!strings.Contains(strings.Join(killed.Text, " "), "SLAUGHTER IT") {
			t.Fatalf("sewer kill=%+v err=%v", killed, runErr)
		}
		if _, runErr = stay.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		); runErr != nil {
			t.Fatal(runErr)
		}
		declined, runErr := stay.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !declined.Exited || stay.CurrentBlockID() != 0x42 {
			t.Fatalf("sewer entry decline block=%02x result=%+v err=%v",
				stay.CurrentBlockID(), declined, runErr)
		}
	})

	t.Run("gambling room has exact defenders and treasure", func(t *testing.T) {
		session := newSession(0x8A)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "),
				"RAKSHASAS GAMBLING WITH DICE") ||
			mustMemory(t, session, 0x4CD7) != 1 {
			t.Fatalf("gambling intro=%+v err=%v", intro, runErr)
		}
		rise, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !rise.WaitingForMenu ||
			!strings.Contains(strings.Join(rise.Text, " "), "DO YOU FLEE") {
			t.Fatalf("gambling rise=%+v err=%v", rise, runErr)
		}
		no := uint16(1)
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x45, Count: 8, IconBlock: 0x45},
				{MonsterID: 0x43, Count: 6, IconBlock: 0x43},
			}) {
			t.Fatalf("gambling fight=%+v err=%v", fight, runErr)
		}
		treasure, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !treasure.CombatRequested ||
			!strings.Contains(strings.Join(treasure.Text, " "), "GATHER THE VALUABLES") ||
			!reflect.DeepEqual(treasure.TreasureRequests, []TreasureRequest{{
				Coins:     [7]uint16{0, 0, 0, 1200, 2000, 15, 9},
				ItemBlock: 0x81,
			}}) {
			t.Fatalf("gambling treasure=%+v err=%v", treasure, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("gambling done=%+v err=%v", done, runErr)
		}
	})

	t.Run("only haughty parlay unlocks journal 57", func(t *testing.T) {
		haughty := newSession(0x8D)
		intro, runErr := haughty.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			intro.Menus[len(intro.Menus)-1].Prompt !=
				"A RAKSHASA RESIDES HERE IN SPLENDOR" ||
			mustMemory(t, haughty, 0x4CDA) != 1 {
			t.Fatalf("rakshasa residence=%+v err=%v", intro, runErr)
		}
		parlay := uint16(3)
		tactics, runErr := haughty.ResumeInteractiveSelectionSeed(
			30000, &parlay, nil, 1, PartyContext{},
		)
		if runErr != nil || !tactics.WaitingForMenu ||
			!reflect.DeepEqual(tactics.Menus[len(tactics.Menus)-1].Options,
				[]string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK",
					"PARLAY_NICE", "PARLAY_ABUSIVE"}) {
			t.Fatalf("rakshasa tactics=%+v err=%v", tactics, runErr)
		}
		story, runErr := haughty.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !story.WaitingForMenu ||
			!strings.Contains(strings.Join(story.Text, " "), "JOURNAL ENTRY 57") ||
			mustMemory(t, haughty, 0x4CBD) != 1 {
			t.Fatalf("rakshasa haughty=%+v err=%v", story, runErr)
		}

		for tactic := uint16(1); tactic < 5; tactic++ {
			session := newSession(0x8D)
			if _, runErr = session.RunEntry(1, 30000, nil); runErr != nil {
				t.Fatal(runErr)
			}
			if _, runErr = session.ResumeInteractiveSelectionSeed(
				30000, &parlay, nil, 1, PartyContext{},
			); runErr != nil {
				t.Fatal(runErr)
			}
			fight, fightErr := session.ResumeInteractiveSelectionSeed(
				30000, &tactic, nil, 1, PartyContext{},
			)
			if fightErr != nil || !fight.CombatRequested ||
				!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
					{MonsterID: 0x44, Count: 5, IconBlock: 0x44},
					{MonsterID: 0x45, Count: 5, IconBlock: 0x45},
					{MonsterID: 0x43, Count: 6, IconBlock: 0x43},
				}) {
				t.Fatalf("rakshasa tactic=%d fight=%+v err=%v",
					tactic, fight, fightErr)
			}
		}
	})
}

func TestRealInnerRuinsKitchenOfficeAndBedroom(t *testing.T) {
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
	newSession := func(terrain uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x43)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, terrain)
		return session
	}
	press := uint16(0)

	for _, room := range []struct {
		name      string
		terrain   uint16
		flag      uint16
		fragments []string
	}{
		{
			name: "office", terrain: 0x8B, flag: 0x4C05,
			fragments: []string{"CONVERTED TO AN OFFICE", "ICONS OF BANE"},
		},
		{
			name: "kitchen", terrain: 0x8C, flag: 0x4C06,
			fragments: []string{"COME INTO THE KITCHEN", "NO THREAT TO YOU"},
		},
	} {
		t.Run(room.name+" is descriptive and one-shot", func(t *testing.T) {
			session := newSession(room.terrain)
			intro, runErr := session.RunEntry(1, 30000, nil)
			joined := strings.Join(intro.Text, " ")
			if runErr != nil || !intro.WaitingForMenu ||
				mustMemory(t, session, room.flag) != 1 {
				t.Fatalf("%s intro=%+v err=%v", room.name, intro, runErr)
			}
			for _, fragment := range room.fragments {
				if !strings.Contains(joined, fragment) {
					t.Fatalf("%s text=%q lacks %q", room.name, joined, fragment)
				}
			}
			done, runErr := session.ResumeInteractiveSelectionSeed(
				30000, &press, nil, 1, PartyContext{},
			)
			if runErr != nil || !done.Exited {
				t.Fatalf("%s done=%+v err=%v", room.name, done, runErr)
			}
			revisit, runErr := session.RunEntry(1, 30000, nil)
			if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
				t.Fatalf("%s revisit=%+v err=%v", room.name, revisit, runErr)
			}
		})
	}

	t.Run("bedroom loot is optional but the room is one-shot", func(t *testing.T) {
		session := newSession(0x8A)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "),
				"DO YOU WANT TO LOOT THE ROOM") ||
			mustMemory(t, session, 0x4C04) != 1 {
			t.Fatalf("bedroom intro=%+v err=%v", intro, runErr)
		}
		yes := uint16(0)
		loot, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &yes, nil, 1, PartyContext{},
		)
		if runErr != nil || !loot.CombatRequested ||
			len(loot.MonsterSpawns) != 0 ||
			!reflect.DeepEqual(loot.TreasureRequests, []TreasureRequest{{
				Coins:     [7]uint16{0, 0, 0, 5000, 5000, 12, 15},
				ItemBlock: 0xFF,
			}}) {
			t.Fatalf("bedroom loot=%+v err=%v", loot, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("bedroom service=%+v err=%v", done, runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
			t.Fatalf("bedroom revisit=%+v err=%v", revisit, runErr)
		}

		decline := newSession(0x8A)
		if _, runErr = decline.RunEntry(1, 30000, nil); runErr != nil {
			t.Fatal(runErr)
		}
		no := uint16(1)
		left, runErr := decline.ResumeInteractiveSelectionSeed(
			30000, &no, nil, 1, PartyContext{},
		)
		if runErr != nil || !left.Exited ||
			left.CombatRequested || len(left.TreasureRequests) != 0 ||
			mustMemory(t, decline, 0x4C04) != 1 {
			t.Fatalf("bedroom decline=%+v err=%v", left, runErr)
		}
	})
}

func TestRealInnerRuinsKennelStatuaryAndChapel(t *testing.T) {
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
	newSession := func(terrain uint16) *BlockSession {
		t.Helper()
		session, sessionErr := NewBlockSession(all, 0x43)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, terrain)
		return session
	}
	press := uint16(0)

	t.Run("kennel has ten hell hounds and the original extra pause", func(t *testing.T) {
		session := newSession(0x87)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "CONVERTED TO A KENNEL") ||
			mustMemory(t, session, 0x4C01) != 1 {
			t.Fatalf("kennel intro=%+v err=%v", intro, runErr)
		}
		blank, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !blank.WaitingForMenu ||
			strings.TrimSpace(strings.Join(blank.Text, " ")) != "" ||
			blank.CombatRequested {
			t.Fatalf("kennel blank pause=%+v err=%v", blank, runErr)
		}
		warning, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !warning.WaitingForMenu ||
			!strings.Contains(strings.Join(warning.Text, " "),
				"MINIONS OF TYRANTHRAXUS RUSH TO ATTACK YOU") {
			t.Fatalf("kennel warning=%+v err=%v", warning, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{{
				MonsterID: 0x44, Count: 10, IconBlock: 0x44,
			}}) {
			t.Fatalf("kennel fight=%+v err=%v", fight, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("kennel victory=%+v err=%v", done, runErr)
		}
		revisit, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
			t.Fatalf("kennel revisit=%+v err=%v", revisit, runErr)
		}
	})

	t.Run("statuary has ten margoyles", func(t *testing.T) {
		session := newSession(0x88)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "STATUES BEGIN TO MOVE") ||
			mustMemory(t, session, 0x4C02) != 1 {
			t.Fatalf("statuary intro=%+v err=%v", intro, runErr)
		}
		warning, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !warning.WaitingForMenu ||
			!strings.Contains(strings.Join(warning.Text, " "),
				"MINIONS OF TYRANTHRAXUS RUSH TO ATTACK YOU") {
			t.Fatalf("statuary warning=%+v err=%v", warning, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{{
				MonsterID: 0x45, Count: 10, IconBlock: 0x45,
			}}) {
			t.Fatalf("statuary fight=%+v err=%v", fight, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("statuary victory=%+v err=%v", done, runErr)
		}
	})

	t.Run("chapel has one high priest and four priests of Bane", func(t *testing.T) {
		session := newSession(0x89)
		intro, runErr := session.RunEntry(1, 30000, nil)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "PRIVATE CHAPEL") ||
			mustMemory(t, session, 0x4C03) != 1 {
			t.Fatalf("chapel intro=%+v err=%v", intro, runErr)
		}
		speech, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		joined := strings.Join(speech.Text, " ")
		if runErr != nil || !speech.WaitingForMenu ||
			!strings.Contains(joined, "TYRANTHRAXUS' GRAND TOOLS") ||
			!strings.Contains(joined, "OTHER PRIESTS SLIP UP BESIDE HIM") {
			t.Fatalf("chapel speech=%+v err=%v", speech, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(
			30000, &press, nil, 1, PartyContext{},
		)
		if runErr != nil || !fight.CombatRequested ||
			!reflect.DeepEqual(fight.MonsterSpawns, []MonsterSpawn{
				{MonsterID: 0x48, Count: 1, IconBlock: 0x48},
				{MonsterID: 0x46, Count: 4, IconBlock: 0x46},
			}) {
			t.Fatalf("chapel fight=%+v err=%v", fight, runErr)
		}
		done, runErr := session.ResumeInteractiveSelectionSeed(
			30000, nil, nil, 1, PartyContext{},
		)
		if runErr != nil || !done.Exited {
			t.Fatalf("chapel victory=%+v err=%v", done, runErr)
		}
	})
}

func TestRealInnerRuinsTyranthraxusAndNamelessFinalRitual(t *testing.T) {
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

	for _, allied := range []bool{false, true} {
		t.Run(fmt.Sprintf("rakshasa-alliance-%v", allied), func(t *testing.T) {
			session, sessionErr := NewBlockSession(all, 0x43)
			if sessionErr != nil {
				t.Fatal(sessionErr)
			}
			for address, value := range map[uint16]uint16{
				0xC04F: 0x83,
				0x4C59: 1,
				0x4C5A: 1,
				0x4C5B: 0xFF,
			} {
				session.SetMemoryValue(address, value)
			}
			if allied {
				session.SetMemoryValue(0x4CBD, 1)
			}
			context := PartyContext{Members: []PartyMemberContext{{Name: "HERO"}}}
			press := uint16(0)
			result, runErr := session.RunEntrySeedWithPartyContext(
				1, 30000, nil, nil, 1, context,
			)
			if runErr != nil {
				t.Fatal(runErr)
			}
			expected := []struct {
				fragment string
				picture  int
			}{
				{"UNABLE TO CONTROL YOURSELF", -1},
				{"THROUGH THE BONDS", -1},
				{"JOURNAL ENTRY 48", 0x47},
				{"HANDS OVER THE THREE ARTIFACTS", 0x46},
				{"DISPOSE OF THESE UNPLEASANT ITEMS", 0x4C},
				{"ARCS EACH ARTIFACT INTO THE POOL", -1},
				{"PARCHMENT WITH THE PHRASE TO RELEASE YOUR BONDS", 0x47},
				{"COMPLETE THE FINAL SPELL", -1},
				{"REVEALING HIMSELF AS NAMELESS", 0x46},
				{"STRIKES NAMELESS DOWN WITH A SINGLE BLOW", 0x47},
				{"BOND'S CONTROL FADE", -1},
				{"PARTY RETRIEVES THE ARTIFACTS", -1},
				{"MINIONS OF TYRANTHRAXUS RUSH TO ATTACK YOU", -1},
			}
			for index, want := range expected {
				joined := strings.Join(result.Text, " ")
				if !result.WaitingForMenu || !strings.Contains(joined, want.fragment) {
					t.Fatalf("stage %d result=%+v, want %q", index, result, want.fragment)
				}
				if want.picture >= 0 && (!result.PictureRequested ||
					int(result.PictureBlock) != want.picture) {
					t.Fatalf("stage %d picture=%v/%02X, want %02X",
						index, result.PictureRequested, result.PictureBlock, want.picture)
				}
				result, runErr = session.ResumeInteractiveSelectionSeed(
					30000, &press, nil, 1, context,
				)
				if runErr != nil {
					t.Fatalf("resume stage %d: %v", index, runErr)
				}
			}
			wantSpawns := []MonsterSpawn{
				{MonsterID: 0x48, Count: 2, IconBlock: 0x48},
				{MonsterID: 0x44, Count: 6, IconBlock: 0x44},
				{MonsterID: 0x45, Count: 6, IconBlock: 0x45},
			}
			if !result.CombatRequested || !reflect.DeepEqual(result.MonsterSpawns, wantSpawns) {
				t.Fatalf("ritual combat=%+v, want %+v", result, wantSpawns)
			}
			if mustMemory(t, session, 0x4C00) != 1 {
				t.Fatalf("ritual completion 4C00=%04X", mustMemory(t, session, 0x4C00))
			}
			done, runErr := session.ResumeInteractiveSelectionSeed(
				30000, nil, nil, 1, context,
			)
			if runErr != nil || !done.Exited {
				t.Fatalf("ritual victory=%+v err=%v", done, runErr)
			}
			revisit, runErr := session.RunEntry(1, 30000, nil)
			if runErr != nil || !revisit.Exited || len(revisit.Text) != 0 {
				t.Fatalf("ritual revisit=%+v err=%v", revisit, runErr)
			}
		})
	}
}

func TestRealInnerRuinsUpperFloorAndTyranthraxusFinalBattle(t *testing.T) {
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
	newSession := func(terrain uint16) *BlockSession {
		session, sessionErr := NewBlockSession(all, 0x43)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		for address, value := range map[uint16]uint16{
			0xC04F: terrain, 0x4C00: 1, 0x4C01: 1, 0x4C02: 1, 0x4C03: 1,
			0x4C59: 1, 0x4C5A: 1, 0x4C5B: 0xFF,
		} {
			session.SetMemoryValue(address, value)
		}
		return session
	}
	press := uint16(0)
	context := PartyContext{Members: []PartyMemberContext{{Name: "HERO"}}}

	t.Run("magic circle has one high priest and three priests of Bane", func(t *testing.T) {
		session := newSession(0x90)
		intro, runErr := session.RunEntrySeedWithPartyContext(1, 30000, nil, nil, 1, context)
		if runErr != nil || !intro.WaitingForMenu ||
			!strings.Contains(strings.Join(intro.Text, " "), "BLUE LIGHTNING ARCS AT YOU") ||
			mustMemory(t, session, 0x4C0A) != 1 {
			t.Fatalf("magic-circle intro=%+v err=%v", intro, runErr)
		}
		disrupted, runErr := session.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
		if runErr != nil || !disrupted.WaitingForMenu ||
			!strings.Contains(strings.Join(disrupted.Text, " "), "DISRUPTED THE CEREMONY") {
			t.Fatalf("magic-circle disrupted=%+v err=%v", disrupted, runErr)
		}
		attack, runErr := session.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
		if runErr != nil || !attack.WaitingForMenu ||
			!strings.Contains(strings.Join(attack.Text, " "), "MINIONS OF TYRANTHRAXUS") {
			t.Fatalf("magic-circle attack=%+v err=%v", attack, runErr)
		}
		fight, runErr := session.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
		want := []MonsterSpawn{
			{MonsterID: 0x48, Count: 1, IconBlock: 0x48},
			{MonsterID: 0x46, Count: 3, IconBlock: 0x46},
		}
		if runErr != nil || !fight.CombatRequested || !reflect.DeepEqual(fight.MonsterSpawns, want) {
			t.Fatalf("magic-circle fight=%+v err=%v", fight, runErr)
		}
	})

	for _, room := range []struct {
		terrain uint16
		text    string
		flag    uint16
	}{
		{0x91, "FOOD STOREROOM", 0x4C0B},
		{0x92, "MOULDERING BOOKS", 0},
		{0x95, "OLD BIERS AND CASKETS", 0x4C0F},
		{0x96, "STENCH OF PRESERVING FLUIDS", 0x4C10},
	} {
		t.Run(fmt.Sprintf("room-%02X", room.terrain), func(t *testing.T) {
			session := newSession(room.terrain)
			result, runErr := session.RunEntrySeedWithPartyContext(1, 30000, nil, nil, 1, context)
			if runErr != nil || !result.WaitingForMenu ||
				!strings.Contains(strings.Join(result.Text, " "), room.text) {
				t.Fatalf("room %02X result=%+v err=%v", room.terrain, result, runErr)
			}
			if room.flag != 0 && mustMemory(t, session, room.flag) != 1 {
				t.Fatalf("room %02X flag %04X=%04X", room.terrain, room.flag, mustMemory(t, session, room.flag))
			}
		})
	}

	t.Run("stairs connect first and second floor", func(t *testing.T) {
		up := newSession(0x97)
		prompt, runErr := up.RunEntrySeedWithPartyContext(1, 30000, nil, nil, 1, context)
		if runErr != nil || !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "), "STAIRS LEAD UP HERE") {
			t.Fatalf("stairs-up prompt=%+v err=%v", prompt, runErr)
		}
		moved, runErr := up.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
		if runErr != nil || !moved.Exited || mustMemory(t, up, 0xC04B) != 2 ||
			mustMemory(t, up, 0xC04C) != 5 || mustMemory(t, up, 0xC04D) != 0 {
			t.Fatalf("stairs-up moved=%+v err=%v", moved, runErr)
		}
		down := newSession(0x98)
		prompt, runErr = down.RunEntrySeedWithPartyContext(1, 30000, nil, nil, 1, context)
		if runErr != nil || !prompt.WaitingForMenu ||
			!strings.Contains(strings.Join(prompt.Text, " "), "STAIRS LEAD DOWN HERE") {
			t.Fatalf("stairs-down prompt=%+v err=%v", prompt, runErr)
		}
		moved, runErr = down.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
		if runErr != nil || !moved.Exited || mustMemory(t, down, 0xC04B) != 10 ||
			mustMemory(t, down, 0xC04C) != 7 || mustMemory(t, down, 0xC04D) != 2 {
			t.Fatalf("stairs-down moved=%+v err=%v", moved, runErr)
		}
	})

	for _, allied := range []bool{false, true} {
		t.Run(fmt.Sprintf("final-battle-allies-%v", allied), func(t *testing.T) {
			session := newSession(0x9A)
			if allied {
				session.SetMemoryValue(0x4CBD, 1)
				session.SetMemoryValue(0x4CC7, 1)
			}
			result, runErr := session.RunEntrySeedWithPartyContext(1, 30000, nil, nil, 1, context)
			for _, fragment := range []string{
				"POWER OF YOUR BONDS HAS RETURNED",
				"GREAT FORCE OF WILL",
				"THAT AMULET WILL LET YOU SCRATCH ME",
			} {
				if runErr != nil || !result.WaitingForMenu ||
					!strings.Contains(strings.Join(result.Text, " "), fragment) {
					t.Fatalf("final %q result=%+v err=%v", fragment, result, runErr)
				}
				result, runErr = session.ResumeInteractiveSelectionSeed(30000, &press, nil, 1, context)
			}
			want := []MonsterSpawn{
				{MonsterID: 0x45, Count: 28, IconBlock: 0x45},
				{MonsterID: 0x47, Count: 1, IconBlock: 0x47},
				{MonsterID: 0x48, Count: 8, IconBlock: 0x48},
			}
			if runErr != nil || !result.CombatRequested || !reflect.DeepEqual(result.MonsterSpawns, want) {
				t.Fatalf("final battle=%+v err=%v", result, runErr)
			}
			victory, runErr := session.ResumeInteractiveSelectionSeed(30000, nil, nil, 1, context)
			if runErr != nil || !victory.ProgramExit || !reflect.DeepEqual(victory.ProgramIDs, []uint8{8}) {
				t.Fatalf("final victory=%+v err=%v", victory, runErr)
			}
		})
	}
}

func TestRealBurialGlenRandomStreamSnapshotContinuation(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	blocks, err := dax.Parse(realZipMember(t, archive, "ECL6.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	all := make(map[uint8][]byte, len(blocks))
	for _, block := range blocks {
		all[block.Entry.ID] = block.Data
	}
	newSession := func() *BlockSession {
		session, sessionErr := NewBlockSession(all, 0x40)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04B, 6)
		session.SetMemoryValue(0xC04C, 12)
		session.SetMemoryValue(0xC04D, 0)
		session.SetMemoryValue(0xC04F, 0x04)
		return session
	}

	const seed = int64(0x408)
	original := newSession()
	first, err := original.RunEntrySeedWithPartyContext(1, 4000, nil, nil, seed, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.RandomValues) == 0 {
		t.Fatalf("real terrain 04h did not consume random: %+v", first)
	}
	snapshot, err := original.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// A real movement transaction clears the per-step guard before the next
	// terrain invocation. Preserve all other shared chapter state.
	original.SetMemoryValue(0x7F81, 0)
	want, err := original.RunEntrySeedWithPartyContext(1, 4000, nil, nil, seed, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}

	restored := newSession()
	if err := restored.RestoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	restored.SetMemoryValue(0x7F81, 0)
	got, err := restored.RunEntrySeedWithPartyContext(1, 4000, nil, nil, seed, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.RandomValues, want.RandomValues) ||
		!reflect.DeepEqual(got.MonsterSpawns, want.MonsterSpawns) ||
		!reflect.DeepEqual(got.Text, want.Text) {
		t.Fatalf("restored real event random=%v spawns=%v text=%q; want random=%v spawns=%v text=%q",
			got.RandomValues, got.MonsterSpawns, got.Text,
			want.RandomValues, want.MonsterSpawns, want.Text)
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
