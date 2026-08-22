package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// registerSyncSession 給一個最小的 session，只用來看 `C04B`..`C04F` 有沒有被
// 回寫。block 內容不重要——這幾支路徑都不跑 bytecode。
func registerSyncSession(t *testing.T) *ecl.BlockSession {
	t.Helper()
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x01: {0, 0}}, 0x01)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func dungeonRegisters(t *testing.T, session *ecl.BlockSession) [3]uint16 {
	t.Helper()
	var out [3]uint16
	for index, address := range [...]uint16{0xC04B, 0xC04C, 0xC04D} {
		value, ok := session.MemoryValue(address)
		if !ok {
			t.Fatalf("暫存器 %04X 沒有被寫過", address)
		}
		out[index] = value
	}
	return out
}

// ★ 原作的引擎每次動到位置或朝向都寫地圖暫存器，三格因此永遠等於真實位置。
// 轉向是最容易漏掉的一條：它不跑 ECL，所以先前整條路徑都沒碰暫存器
// （spec 1172）。
func TestTurnDungeonSyncsTheFacingRegister(t *testing.T) {
	session := registerSyncSession(t)
	state := State{session: session, DungeonX: 4, DungeonY: 9, DungeonDirection: 2}
	state.Area.InDungeon = true

	state.TurnDungeon(2)
	if state.DungeonDirection != 4 {
		t.Fatalf("轉向後朝向 ＝ %d，want 4", state.DungeonDirection)
	}
	if value, ok := session.MemoryValue(0xC04D); !ok || value != 2 {
		t.Fatalf("C04D ＝ %d ok=%v，want 2", value, ok)
	}

	// 往回轉一格會繞回 6（西），折過的暫存器是 3。
	state.TurnDungeon(-6)
	if value, ok := session.MemoryValue(0xC04D); !ok || value != 3 {
		t.Fatalf("倒轉之後 C04D ＝ %d ok=%v，want 3", value, ok)
	}
}

// remake 存檔讀回來之後三格要跟隊伍位置一致：版本回退與範圍夾擠會改座標，
// 不回寫的話重畫會拿快照裡的舊值去投影（spec 1172）。
func TestLoadPartyFileSyncsTheDungeonRegisters(t *testing.T) {
	abilities := party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
		Dexterity: 12, Constitution: 14, Charisma: 10}
	// ⚠ 存檔端故意**不帶 session**：這一支要看的是「讀檔之後三格有沒有被回寫」，
	// 帶了 session 就變成在測快照還原，遮住真正想擋的洞。
	source := NewState(trainingTestCatalog(t))
	source.partyRoster = party.Roster{
		{ID: "p1", Name: "亞勇", Race: party.RaceHuman, Class: party.ClassFighter,
			Level: 3, Abilities: abilities},
	}
	source.Mode = ModeDungeon
	source.Area.InDungeon = true
	source.DungeonX, source.DungeonY, source.DungeonDirection = 9, 4, 6
	source.MapX, source.MapY = 9, 4
	path := filepath.Join(t.TempDir(), "party.json")
	if err := source.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}

	target := NewState(trainingTestCatalog(t))
	target.session = registerSyncSession(t)
	if err := target.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if target.DungeonX != 9 || target.DungeonY != 4 || target.DungeonDirection != 6 {
		t.Fatalf("讀檔後位置 ＝ (%d,%d,%d)，want (9,4,6)",
			target.DungeonX, target.DungeonY, target.DungeonDirection)
	}
	if got, want := dungeonRegisters(t, target.session), [3]uint16{9, 4, 3}; got != want {
		t.Fatalf("讀檔後三格 ＝ %v，want %v", got, want)
	}
}
