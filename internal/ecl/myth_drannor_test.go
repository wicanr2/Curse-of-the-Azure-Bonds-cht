package ecl

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealStandingStoneRevealsTyranthraxusAndMythDrannor(t *testing.T) {
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
	session.SetMemoryValue(0x4C5B, 0xFF)
	session.SetMemoryValue(0x4C5A, 1)

	result, err := session.RunFrom(0x01C4, 1000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.WaitingForMenu || !strings.Contains(strings.Join(result.Text, " "), "STANDING STONES") {
		t.Fatalf("initial result=%+v, want Standing Stones prompt", result)
	}

	press := uint16(0)
	reveal, err := session.ResumeInteractiveSelectionSeed(
		1000, &press, nil, 1, PartyContext{},
	)
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
	for address, want := range map[uint16]uint16{
		0x4C59: 1,
		0x4C5A: 1,
		0x4C5B: 0xFF,
	} {
		if got := mustMemory(t, session, address); got != want {
			t.Fatalf("memory[0x%04X]=0x%X, want 0x%X", address, got, want)
		}
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
