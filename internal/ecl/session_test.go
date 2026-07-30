package ecl

import "testing"

func sessionBlock(entry uint16) []byte {
	block := make([]byte, 2+24)
	for i := 0; i < 5; i++ {
		pos := 2 + i*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, byte(entry), byte(entry>>8)
	}
	return block
}

func TestBlockSessionSwitchAndInitialEntry(t *testing.T) {
	session, err := NewBlockSession(map[uint8][]byte{0x50: sessionBlock(0x8014), 0x51: sessionBlock(0x8020)}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := session.InitialEntry(); got != 0x14 {
		t.Fatalf("initial entry=%#x, want 0x14", got)
	}
	id := uint8(0x51)
	if err := session.ApplyResult(RunResult{NewECLBlockID: &id}); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x51 {
		t.Fatalf("current block=%#x, want 0x51", session.CurrentBlockID())
	}
	if got, _ := session.InitialEntry(); got != 0x20 {
		t.Fatalf("switched entry=%#x, want 0x20", got)
	}
}

func TestBlockSessionRejectsUnavailableSwitch(t *testing.T) {
	session, err := NewBlockSession(map[uint8][]byte{0x50: sessionBlock(0x8014)}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Switch(0x51); err == nil {
		t.Fatal("expected unavailable block error")
	}
}

func TestBlockSessionMapsPayloadIntoReferenceCodeMemory(t *testing.T) {
	payload := []byte{
		0x2A, 0x02, 0x10, 0x80, 0x00, 0, 0x02, 0x00, 0x7B,
		0x11, 0x01, 0x00, 0x7B,
		0x00, 0, 0, 7,
	}
	session, err := NewBlockSession(map[uint8][]byte{1: append([]byte{0, 0}, payload...)}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.RunFrom(0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "7" {
		t.Fatalf("text=%q, want GETTABLE byte from ECL address 0x8010", result.Text)
	}
}

func TestBlockSessionFirstRunFromPreservesSeededMemory(t *testing.T) {
	block := append([]byte{0, 0},
		0x11, 0x01, 0x00, 0x90,
		0x00,
	)
	session, err := NewBlockSession(map[uint8][]byte{1: block}, 1)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0x9000, 7)

	result, err := session.RunFrom(0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Text) != 1 || result.Text[0] != "7" {
		t.Fatalf("text=%q, want first RunFrom to read seeded memory", result.Text)
	}
	if value, ok := session.MemoryValue(0x9000); !ok || value != 7 {
		t.Fatalf("shared memory[0x9000]=%d,%v, want 7,true", value, ok)
	}
}

func TestBlockSessionRunInteractiveFollowsNewECL(t *testing.T) {
	first := sessionBlock(0x8014)
	// Replace initial entry code at payload +0x14 with NEWECL 0x51.
	first[2+0x14] = 0x20
	first[2+0x15], first[2+0x16] = 0x00, 0x51
	second := sessionBlock(0x8014)
	second[2+0x14] = 0x2B
	// Keep the synthetic second block bounded; the session must switch before
	// returning its result even though the menu framing is intentionally minimal.
	session, err := NewBlockSession(map[uint8][]byte{0x50: first, 0x51: second}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.RunInteractive(20, nil)
	if err == nil {
		// The second block is expected to stop safely on its synthetic menu;
		// either way the block transition must already have happened.
	}
	if session.CurrentBlockID() != 0x51 {
		t.Fatalf("current block=%#x result=%+v err=%v", session.CurrentBlockID(), result, err)
	}
}

func TestBlockSessionResumesMenuWithCumulativeSelections(t *testing.T) {
	block := append([]byte{0, 0},
		0x2B, 0x02, 0x00, 0x90, 0x00, 0x02,
		0x80, 0x02, 0x20, 0x92,
		0x80, 0x02, 0x0C, 0x32,
		0x00,
	)
	session, err := NewBlockSession(map[uint8][]byte{0x50: block}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	first, err := session.RunFrom(0, 8, nil)
	if err != nil || !first.WaitingForMenu {
		t.Fatalf("first run=%+v err=%v, want menu pause", first, err)
	}
	second, err := session.RunFrom(0, 8, []uint16{1})
	if err != nil {
		t.Fatal(err)
	}
	if second.WaitingForMenu || len(second.Menus) != 1 || second.Menus[0].Selected != 1 {
		t.Fatalf("second run=%+v, want resumed selection 1", second)
	}
}

func TestBlockSessionResumesMenuWithIncrementalSelection(t *testing.T) {
	code := []byte{
		0x2B, 0x02, 0x00, 0x90, 0x00, 0x02,
		0x80, 0x02, 0x20, 0x92,
		0x80, 0x02, 0x0C, 0x32,
		0x00,
	}
	block := make([]byte, 2+0x14+len(code))
	for i := 0; i < 5; i++ {
		pos := 2 + i*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, 0x14, 0x80
	}
	copy(block[2+0x14:], code)
	session, err := NewBlockSession(map[uint8][]byte{0x50: block}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	first, err := session.RunInteractiveSeedWithPartyContext(8, nil, 1, PartyContext{})
	if err != nil || !first.WaitingForMenu {
		t.Fatalf("first run=%+v err=%v, want menu pause", first, err)
	}
	selection := uint16(1)
	second, err := session.ResumeInteractiveSelectionSeed(8, &selection, nil, 1, PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if second.WaitingForMenu || len(second.Menus) != 1 || second.Menus[0].Selected != 1 {
		t.Fatalf("second run=%+v, want resumed incremental selection 1", second)
	}
}

func TestBlockSessionCarriesMemoryAcrossNewECL(t *testing.T) {
	first := append([]byte{0, 0,
		0x09, 0x00, 7, 0x02, 0x00, 0x7B,
		0x20, 0x00, 0x51,
	}, make([]byte, 32)...)
	second := append([]byte{0, 0}, make([]byte, 32)...)
	for i := 0; i < 5; i++ {
		pos := 2 + i*4
		second[pos+1], second[pos+2], second[pos+3] = 0x02, 0x14, 0x80
	}
	second[2+0x14] = 0x11
	second[2+0x15], second[2+0x16], second[2+0x17] = 0x01, 0x00, 0x7B
	second[2+0x18] = 0x00
	session, err := NewBlockSession(map[uint8][]byte{0x50: first, 0x51: second}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.RunFrom(0, 20, nil)
	if err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x51 || len(result.Text) != 1 || result.Text[0] != "7" {
		t.Fatalf("result=%+v block=%#x, want target block to print carried memory", result, session.CurrentBlockID())
	}
	if value, ok := session.MemoryValue(0x7B00); !ok || value != 7 {
		t.Fatalf("shared memory[0x7B00]=%d,%v, want 7,true", value, ok)
	}
}

func TestBlockSessionRunEntryRestartsLifecycleWithSharedMemory(t *testing.T) {
	block := sessionBlock(0x8014)
	block = append(block, make([]byte, 32)...)
	// entry 0 saves 7 then exits; initial entry saves 9 then exits.
	block[4], block[5] = 0x20, 0x80
	copy(block[2+0x14:], []byte{0x09, 0x00, 9, 0x02, 0x00, 0x90, 0x00})
	copy(block[2+0x20:], []byte{0x09, 0x00, 7, 0x02, 0x01, 0x90, 0x00})
	session, err := NewBlockSession(map[uint8][]byte{1: block}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.RunInteractive(20, nil); err != nil {
		t.Fatal(err)
	}
	result, err := session.RunEntry(0, 20, nil)
	if err != nil {
		t.Fatal(err)
	}
	initial, initialOK := session.MemoryValue(0x9000)
	turn, turnOK := session.MemoryValue(0x9001)
	if !result.Exited || !initialOK || initial != 9 || !turnOK || turn != 7 {
		t.Fatalf("result=%+v memory=(%d,%v),(%d,%v)", result, initial, initialOK, turn, turnOK)
	}
}

func TestBlockSessionAggregatesSignalsAcrossNewECL(t *testing.T) {
	first := sessionBlock(0x8014)
	first[2] = 0x20
	first[3], first[4] = 0x00, 0x51
	second := sessionBlock(0x8014)
	second = append(second, make([]byte, 12)...)
	second[2+0x14] = 0x0E
	second[2+0x15], second[2+0x16] = 0x00, 0x1D
	second[2+0x17] = 0x20
	second[2+0x18], second[2+0x19] = 0x00, 0x52
	third := sessionBlock(0x8014)
	third = append(third, make([]byte, 16)...)
	third[2+0x14] = 0x21
	copy(third[2+0x15:], []byte{0x00, 0x12, 0x00, 0x34, 0x00, 0x10, 0x00})
	firstResult, firstErr := RunSubset(first, 0, 40)
	if firstErr != nil || firstResult.NewECLBlockID == nil {
		t.Fatalf("first synthetic block=%x result=%+v err=%v", first, firstResult, firstErr)
	}
	session, err := NewBlockSession(map[uint8][]byte{0x50: first, 0x51: second, 0x52: third}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.RunFrom(0, 40, nil)
	if err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x52 || !result.PictureRequested || result.PictureBlock != 0x1D || !result.LoadFilesRequested || result.LoadFiles != [3]uint16{0x12, 0x34, 0x10} {
		t.Fatalf("cross-block signals=%+v current=0x%02X", result, session.CurrentBlockID())
	}
}

func TestBlockSessionCarriesDumpedWorkingPartyAcrossNewECL(t *testing.T) {
	first := append(sessionBlock(0x8014), make([]byte, 16)...)
	copy(first[2+0x14:], []byte{
		0x0A, 0x02, 0x01, 0x00,
		0x3E,
		0x20, 0x00, 0x51,
	})
	second := append(sessionBlock(0x8014), make([]byte, 8)...)
	copy(second[2+0x14:], []byte{
		0x3F, 0x00, 0x27,
		0x00,
	})
	session, err := NewBlockSession(map[uint8][]byte{0x50: first, 0x51: second}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	context := PartyContext{Members: []PartyMemberContext{{Name: "A", Effects: []uint8{0x27}}, {Name: "B"}}}
	result, err := session.RunInteractiveSeedWithPartyContext(20, nil, 1, context)
	if err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x51 || len(result.DumpRequests) != 1 || !result.DumpRequests[0].Resolved || len(result.FindSpecialRequests) != 1 || !result.FindSpecialRequests[0].Found {
		t.Fatalf("result=%+v current=0x%02X", result, session.CurrentBlockID())
	}
	if len(context.Members) != 2 || context.Members[0].Name != "A" || context.Members[1].Name != "B" {
		t.Fatalf("caller context was mutated: %#v", context)
	}
}
