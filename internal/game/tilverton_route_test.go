package game

import (
	"archive/zip"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 提爾佛頓（ECL2 的 0x01／0x02／0x03）是**遊戲真正開始的地方**，而
// `TestRealNewGameRunsToTheEnding` 從世界地圖那個 hub（0x50）起跑，整章跳過。
// 可達性盤點因此顯示那 48 個分派索引「主線一格都沒踏到」。
//
// ★ `cmd/dungeon-walk-probe` 已經證明**走得進去**（冷走走得到 25 個索引），
// 所以那不是缺陷、是路線的選擇。這一條把選擇補上：真的走一遍那三段，
// 並套用與主線同一條語系不變量。
//
// ⚠ 這**不是**「從新遊戲玩到提爾佛頓」：進段是直接進的（那三段本來就是開場，
// 沒有更前面的東西）。它證明的是**段內**走得通、而且沿路的字都是中文。
//
// ⚠ 走法是廣度優先的幾何走訪，不是劇情路線：它不解謎、不觸發特定旗標，
// 所以「走到的格子」是下界。
func TestTilvertonRouteIsWalkableAndLocalized(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	catalog := geo.NewCatalog()
	for chapter := 2; chapter <= 6; chapter++ {
		if err := catalog.AddDAX(uint8(chapter),
			zipData(t, image, "GEO"+strconv.Itoa(chapter)+".DAX")); err != nil {
			t.Fatal(err)
		}
	}

	messages := map[string]bool{}
	totalCells := 0
	for _, id := range []string{"ECL2/0x01", "ECL2/0x02", "ECL2/0x03"} {
		seg, ok := segment.Lookup(id)
		if !ok {
			t.Fatalf("段註冊表裡沒有 %s", id)
		}
		t.Run(id, func(t *testing.T) {
			cells := walkTilvertonSegment(t, blocks, catalog, seg, messages)
			// ⚠ 正對照：走不到東西的話下面的語系檢查會**正確地通過**，
			// 因為它一句話都沒收到。先擋住這種假綠。
			if cells < 8 {
				t.Fatalf("%s 只走到 %d 格，這一段不可能這麼小", id, cells)
			}
			totalCells += cells
			t.Logf("%s 走到 %d 格", id, cells)
		})
	}
	if totalCells == 0 {
		return
	}

	// 把提爾佛頓走過的格子併進可達性導出（見 `exportCampaignVisitedCells`）。
	exportCampaignVisitedCells(t)

	t.Run("語系：提爾佛頓沿路沒有落回原文", func(t *testing.T) {
		if len(messages) < 10 {
			t.Fatalf("只記到 %d 句話，走三段不可能這麼少", len(messages))
		}
		var fallbacks []string
		for message := range messages {
			if campaignMessageHasHan(message) {
				continue
			}
			if !campaignMessageHasLatinWord(message) {
				continue
			}
			fallbacks = append(fallbacks, message)
		}
		sort.Strings(fallbacks)
		for _, message := range fallbacks {
			t.Errorf("落回原文：%q", message)
		}
		t.Logf("提爾佛頓沿路記到 %d 句話，落回原文 %d 句", len(messages), len(fallbacks))
	})
}

// walkTilvertonSegment 進段之後從落點做廣度優先，回傳走到的格子數。
// 沿路把玩家看得到的字收進 messages，並把 (block, 地形碼) 記進
// campaignVisitedCells 讓可達性盤點看得到。
func walkTilvertonSegment(t *testing.T, blocks map[uint8][]byte, catalog geo.Catalog,
	seg segment.Segment, messages map[string]bool) int {
	t.Helper()
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, seg.Block)
	state.SetGeoCatalog(catalog)
	// ⚠ `ECL2/0x01` 是**開場**，走的是 `BeginAdventure`，它要的是**角色名冊**
	// 不只是戰鬥員——少了名冊會回「adventure requires a created or loaded party」，
	// 那一段就整段被跳過（第一版就是這樣漏掉開場那一張圖的）。
	state.partyRoster = party.Roster{{
		ID: "walker", Name: "走訪者", Race: party.RaceHuman,
		Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 18, Intelligence: 10, Wisdom: 10,
			Dexterity: 16, Constitution: 16, Charisma: 10},
	}}
	// ⚠ 隊伍撐起來只為了讓入口伏擊不會把走訪擋在門口。**只給盤點用**，
	// 它不代表正常隊伍的強度，所以這一條不宣稱「一般玩家打得過」。
	if err := state.SetParty([]combat.Fighter{{
		ID: "walker", Name: "走訪者", Side: combat.SideParty,
		HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
		AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1, DamageBonus: 100,
		AttacksPerTurn: 8, InitiativeBonus: 100,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := state.EnterSegment(seg); err != nil {
		t.Skipf("%s 進不去：%v", seg.ID, err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: seg.Member, BlockID: state.Area.Current3DMapBlockID})
	if !ok {
		t.Skipf("%s 讀不到 GEO", seg.ID)
	}
	collect := func() {
		if state.Message != "" {
			messages[state.Message] = true
		}
		if state.Prompt != "" {
			messages[state.Prompt] = true
		}
		if state.Mode == ModeDungeon && state.session != nil {
			campaignVisitedCells[campaignCellKey{
				block: state.session.CurrentBlockID(), terrain: state.DungeonWallRoof,
			}] = true
		}
	}
	settle := func() bool {
		for step := 0; step < 60 && state.Mode != ModeDungeon; step++ {
			collect()
			if state.CombatActive() {
				for turn := 0; turn < 400 && state.CombatActive(); turn++ {
					if err := state.CombatAct(); err != nil {
						return false
					}
				}
				continue
			}
			choice := 0
			if state.Mode == ModePlace && len(state.Choices) > 0 {
				choice = len(state.Choices) - 1
			}
			if err := state.Continue(); err != nil {
				if selectErr := state.Select(choice); selectErr != nil {
					return false
				}
			}
		}
		collect()
		return state.Mode == ModeDungeon
	}
	if !settle() {
		t.Skipf("%s 入口推不回地城（停在 %v）", seg.ID, state.Mode)
	}

	type point struct{ x, y int }
	start := point{state.DungeonX, state.DungeonY}
	seen := map[point]bool{start: true}
	queue := []point{start}
	visited := 1
	for len(queue) > 0 && visited < 260 {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range []int{0, 2, 4, 6} {
			deltaX, deltaY := normalDungeonDelta(direction)
			next := point{current.x + deltaX, current.y + deltaY}
			if next.x < 0 || next.x >= geo.Width || next.y < 0 || next.y >= geo.Height || seen[next] {
				continue
			}
			state.SetDungeonGeometryView(current.x, current.y, uint8(direction))
			state.DungeonWallRoof = grid.CellWrapped(current.x, current.y).Terrain
			if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
				continue
			}
			if err := state.MoveDungeon(grid, deltaX, deltaY, direction); err != nil {
				continue
			}
			seen[next] = true
			visited++
			if !settle() {
				// 推不回地城就不從這一格繼續往外走；已經到過的算數。
				continue
			}
			queue = append(queue, next)
		}
	}
	return visited
}
