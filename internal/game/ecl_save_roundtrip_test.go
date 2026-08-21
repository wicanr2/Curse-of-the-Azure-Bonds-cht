package game

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestGameSaveRestoresECLMemoryPCAndRandomContinuation(t *testing.T) {
	// RANDOM; SAVE; PROGRAM 9 is a real engine boundary. A second RANDOM after
	// loading must match uninterrupted execution, including the mutable word.
	program := []byte{
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x00, 0x70,
		0x09, 0x02, 0x34, 0x12, 0x01, 0x01, 0x70,
		0x38, 0x00, 0x09,
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x02, 0x70,
		0x00,
	}
	block := make([]byte, 2+20, 2+20+len(program))
	for index := 0; index < 5; index++ {
		offset := 2 + index*4
		block[offset], block[offset+1], block[offset+2], block[offset+3] = 0x01, 0x02, 0x14, 0x80
	}
	block = append(block, program...)
	blocks := map[uint8][]byte{1: block}
	state := NewStateFromECLBlocks(testCatalog(), blocks, 1)
	// Use a non-default seed. After loading, RunFrom passes its compatibility
	// default, but the restored session-owned stream must remain authoritative.
	state.session.ResetRandomSeed(77)
	hero := combat.Fighter{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10}
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 10, MaxHitPoints: 10,
		Abilities: party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10},
	}}
	boundary, err := state.session.RunFrom(20, 100, nil)
	if err != nil || !boundary.ProgramExit {
		t.Fatalf("boundary=%+v err=%v", boundary, err)
	}
	savePath := filepath.Join(t.TempDir(), "session-v6.json")
	if err := state.SavePartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	want, err := state.session.RunFrom(20, 100, nil)
	if err != nil {
		t.Fatal(err)
	}

	restored := NewStateFromECLBlocks(testCatalog(), blocks, 1)
	if err := restored.LoadPartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	got, err := restored.session.RunFrom(20, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.RandomValues, want.RandomValues) {
		t.Fatalf("restored random=%v, want %v", got.RandomValues, want.RandomValues)
	}
	if value, ok := restored.session.MemoryValue(0x7001); !ok || value != 0x1234 {
		t.Fatalf("restored memory[7001]=%04x,%v", value, ok)
	}
	if restored.session.CurrentBlockID() != 1 {
		t.Fatalf("restored block=%02x", restored.session.CurrentBlockID())
	}
}

// 存檔要把「停在旁邊」那幾段的暫存也帶著（`4C00h`..`4C0Fh`，spec 1162）。
// 只釘 `internal/ecl` 那一層不夠：存檔是 JSON，欄位漏掉照樣編得出來，
// 症狀要到「讀檔之後走回上一段」才看得到。
func TestGameSaveKeepsParkedBlockScratchAcrossBlocks(t *testing.T) {
	exitOnly := func() []byte {
		block := make([]byte, 2+20)
		for index := 0; index < 5; index++ {
			offset := 2 + index*4
			block[offset], block[offset+1], block[offset+2], block[offset+3] = 0x01, 0x02, 0x14, 0x80
		}
		return append(block, 0x00)
	}
	blocks := map[uint8][]byte{0x10: exitOnly(), 0x20: exitOnly()}
	state := NewStateFromECLBlocks(testCatalog(), blocks, 0x10)
	hero := combat.Fighter{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10}
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 10, MaxHitPoints: 10,
		Abilities: party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10},
	}}
	// 第一段演過一次一次性事件，然後換段。
	state.session.SetMemoryValue(0x4C06, 7)
	if err := state.session.Switch(0x20); err != nil {
		t.Fatal(err)
	}
	state.session.SetMemoryValue(0x4C06, 3)

	savePath := filepath.Join(t.TempDir(), "parked-scratch.json")
	if err := state.SavePartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	restored := NewStateFromECLBlocks(testCatalog(), blocks, 0x10)
	if err := restored.LoadPartyFile(savePath); err != nil {
		t.Fatal(err)
	}
	if value, ok := restored.session.MemoryValue(0x4C06); !ok || value != 3 {
		t.Fatalf("讀檔後目前這一段的 4C06=%d,%v，want 3", value, ok)
	}
	if err := restored.session.Switch(0x10); err != nil {
		t.Fatal(err)
	}
	if value, ok := restored.session.MemoryValue(0x4C06); !ok || value != 7 {
		t.Fatalf("讀檔後走回 0x10 的 4C06=%d,%v，want 7", value, ok)
	}
}

// 原版版面的存檔（SAVGAM）前三塊就是 ECL 位址空間的區 0／1／2（spec 1163），
// 所以旗標要寫得進去也讀得回來。這是與原版存檔互通的那一半，和 remake 自己的
// `SavePartyFile` 是兩條路。
func TestSAVGAMPrefixCarriesECLBankMemory(t *testing.T) {
	exitOnly := func() []byte {
		block := make([]byte, 2+20)
		for index := 0; index < 5; index++ {
			offset := 2 + index*4
			block[offset], block[offset+1], block[offset+2], block[offset+3] = 0x01, 0x02, 0x14, 0x80
		}
		return append(block, 0x00)
	}
	blocks := map[uint8][]byte{0x10: exitOnly()}
	state := NewStateFromECLBlocks(testCatalog(), blocks, 0x10)
	state.partyRoster = party.Roster{{ID: "p1", Name: "阿勇"}}
	// 三個區各放一格：地圖本地旗標（區 0）、任務旗標（區 1）、區 2。
	state.Area.GameArea = 4
	state.session.SetMemoryValue(0x4C5A, 1)
	state.session.SetMemoryValue(0x7F79, 0x2A)
	state.session.SetMemoryValue(0x7A40, 0x1234)

	path := filepath.Join(t.TempDir(), "SAVGAM0.DAT")
	if err := state.SaveSAVGAMPrefix(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewStateFromECLBlocks(testCatalog(), blocks, 0x10)
	if err := loaded.LoadSAVGAMPrefix(path); err != nil {
		t.Fatal(err)
	}
	for address, want := range map[uint16]uint16{0x4C5A: 1, 0x7F79: 0x2A, 0x7A40: 0x1234} {
		if value, ok := loaded.session.MemoryValue(address); !ok || value != want {
			t.Fatalf("SAVGAM 讀回來的 memory[%#x]=%#x,%v，want %#x", address, value, ok, want)
		}
	}
	// ⚠ Area 編碼器管的那幾格由 `s.Area` 決定，不是由 VM 記憶體決定——兩邊是
	// 同一批位址（`7F12h` ＝ `Area2` 位移 `0x624`），remake 這一側以 `s.Area`
	// 為準，寫入順序就是這個意思。
	if value, ok := loaded.session.MemoryValue(0x7F12); !ok || value != uint16(loaded.Area.GameArea) {
		t.Fatalf("7F12=%#x,%v，want Area.GameArea=%#x", value, ok, loaded.Area.GameArea)
	}
	// 沒被 VM 碰過的格子不會多出鍵來——否則 `MemoryValue` 的第二個回傳值
	// 就永遠是 true，別處靠它分辨「沒寫過」的判斷會整批失效。
	if _, ok := loaded.session.MemoryValue(0x4C5B); ok {
		t.Fatal("SAVGAM 讀回來多出了沒有人寫過的格子")
	}
}
