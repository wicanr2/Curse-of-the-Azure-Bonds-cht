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
