package game

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 服務段的選單走訪：酒吧（`ECL4/0x23`）沒有地形分派，BFS 走訪
//（`TestTilvertonRoute`）碰不到它——實跑指令覆蓋因此停在 17.8%。
// 這裡用「每種選單策略各走一遍」把飲料與傳聞選單逐項輪到；
// 魔法商店（`ECL4/0x25`）有自己的地城圖，走 BFS 那一份清單。
//
// ⚠ 這是**內容走訪**不是玩法宣稱：隊伍帶著錢與物品只為了讓買賣分支開得起來。
// 走到的選單數是下界——要旗標才出現的分支不在裡面。
func TestRealSideServiceMenusAreWalkable(t *testing.T) {
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
	for _, id := range []string{"ECL4/0x23"} {
		seg, ok := segment.Lookup(id)
		if !ok {
			t.Fatalf("段註冊表裡沒有 %s", id)
		}
		t.Run(id, func(t *testing.T) {
			menus := map[string]bool{}
			steps := 0
			for _, pick := range []int{0, 1, 2, 3, -1} {
				steps += walkServiceSegment(t, pick, blocks, catalog, seg, menus)
			}
			// 正對照：一個選單都沒輪到代表進段就失敗了，不是「服務很小」。
			if len(menus) < 2 || steps < 8 {
				t.Fatalf("%s 只輪到 %d 種選單、%d 步，服務段不可能這麼小", id, len(menus), steps)
			}
			t.Logf("%s 輪到 %d 種選單、共 %d 步", id, len(menus), steps)
		})
	}
}

// walkServiceSegment 進段後照 `pick` 策略連選，直到沒有選單或步數用盡。
// 回傳走的步數；相異選單（以選項串接為鍵）收進 menus。
func walkServiceSegment(t *testing.T, pick int, blocks map[uint8][]byte,
	catalog geo.Catalog, seg segment.Segment, menus map[string]bool) int {
	t.Helper()
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, seg.Block)
	state.SetGeoCatalog(catalog)
	state.partyRoster = party.Roster{{
		ID: "walker", Name: "走訪者", Race: party.RaceHuman,
		Class: party.ClassFighter, Level: 5,
		Abilities: party.Abilities{Strength: 18, Intelligence: 10, Wisdom: 10,
			Dexterity: 16, Constitution: 16, Charisma: 10},
		// 錢與一件可賣的物品：讓買賣、鑑定那幾條分支開得起來。
		Gold:      5000,
		Equipment: []monster.ItemRecord{{Name: "LONG SWORD", Type: 1, Count: 1}},
	}}
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
	steps := 0
	for ; steps < 60; steps++ {
		if testing.Verbose() {
			t.Logf("step %d: mode=%d choices=%v msg=%.40q prompt=%.30q",
				steps, state.Mode, state.Choices, state.Message, state.Prompt)
		}
		choices := state.Choices
		if state.Mode == ModeEvent && len(choices) == 0 {
			if err := state.Continue(); err != nil {
				break
			}
			continue
		}
		if len(choices) == 0 {
			break
		}
		key := ""
		for _, choice := range choices {
			key += choice + "｜"
		}
		menus[key] = true
		index := pick
		if index < 0 || index >= len(choices) {
			index = len(choices) - 1
		}
		if err := state.Select(index); err != nil {
			break
		}
		// 走出服務段（block 換掉）才結束這一輪——酒吧的飲料選單本來就開在
		// 荒野模式，不能拿模式當離開判斷。
		if state.session != nil && state.session.CurrentBlockID() != seg.Block {
			break
		}
	}
	return steps
}
