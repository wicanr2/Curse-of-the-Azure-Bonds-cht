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

func mustMemory(t *testing.T, session *BlockSession, address uint16) uint16 {
	t.Helper()
	value, ok := session.MemoryValue(address)
	if !ok {
		t.Fatalf("missing memory 0x%04X", address)
	}
	return value
}
