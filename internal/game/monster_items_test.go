package game

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 怪物的物品鏈（`MON*ITM`）先前只在 NPC 入隊時用得到；戰鬥裡的怪物身上什麼都
// 沒有，所以原作的 AI 換裝（spec 1120）在怪物側**連資料都不存在**。
//
// 這一條驗兩件事：原始資料裡真的有帶物品的怪，以及鏡射過去的欄位對得上。
// ⚠ 先擋「原始資料非空」再擋鏡射——資料整批讀不到的話，鏡射永遠正確地產出空的。
func TestMonsterItemChainsExistAndMirror(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	armed, total, richest := 0, 0, 0
	var sample monster.ItemRecord
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "MON"+strconv.Itoa(chapter)+"ITM.DAX"))
		if parseErr != nil {
			t.Fatalf("MON%dITM: %v", chapter, parseErr)
		}
		for _, block := range blocks {
			items, itemErr := monster.ParseItems(block.Data)
			if itemErr != nil {
				t.Fatalf("MON%dITM block %#02x: %v", chapter, block.Entry.ID, itemErr)
			}
			total++
			if len(items) == 0 {
				continue
			}
			armed++
			if len(items) > richest {
				richest = len(items)
				sample = items[0]
			}
		}
	}
	if total == 0 {
		t.Fatal("六章的 MON*ITM 一個區塊都沒有")
	}
	if armed == 0 {
		t.Fatal("六章的 MON*ITM 裡沒有一隻怪帶物品，接線等於沒接")
	}
	t.Logf("MON*ITM：%d 個區塊、%d 隻帶物品、最多的一隻 %d 件", total, armed, richest)

	mirrored := combatMonsterItems([]monster.ItemRecord{sample})
	if len(mirrored) != 1 {
		t.Fatalf("鏡射之後有 %d 件，應該是 1 件", len(mirrored))
	}
	got := mirrored[0]
	if got.Name != sample.Name || got.Type != sample.Type || got.Plus != sample.Plus ||
		got.Readied != sample.Readied || got.Cursed != sample.Cursed ||
		got.Count != sample.Count || got.Weight != sample.Weight ||
		got.Value != sample.Value || got.Affects != sample.Affects {
		t.Errorf("鏡射之後是 %+v，原始是 %+v", got, sample)
	}
	if combatMonsterItems(nil) != nil {
		t.Error("空的物品鏈應該鏡射成 nil，不是空 slice")
	}
}

// 開戰時，怪物身上的物品要跟著進到戰鬥員身上——原作的 AI 換裝讀的就是這條鏈。
//
// ⚠ 同一個 spawn 生出來的多隻怪各自要有一份：共用同一個 slice 的話，一隻換裝
// 會影響到全部，而且症狀只在「同一種怪出現兩隻以上」時才看得到。
func TestEncounterCarriesMonsterItemsPerFighter(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "亞勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
			Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	// ⚠ 章節取自**怪物 ID 的命名空間**，不是目前的 block：ID 7 落在第二章。
	state.SetMonsterItemsForECL(2, map[uint8][]monster.ItemRecord{
		7: {{Name: "LONG SWORD", Type: 4, Count: 1, Readied: true},
			{Name: "DAGGER", Type: 1, Count: 1}},
	})
	result := ecl.RunResult{CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 2, IconBlock: 81}}}
	records := map[uint8]monster.Record{
		7: {Name: "FIRE KNIFE", HitPoints: 8, MaxHitPoints: 8, AttacksPerTurn: 1},
	}
	partyFighters := []combat.Fighter{{ID: "p1", Name: "亞勇",
		Side: combat.SideParty, HitPoints: 12, MaxHitPoints: 12}}
	if err := state.StartEncounter(result, records, partyFighters, 11); err != nil {
		t.Fatal(err)
	}
	armed := 0
	seen := map[*combat.MonsterItem]bool{}
	for _, fighter := range state.battle.Fighters() {
		if fighter.Side == combat.SideParty {
			continue
		}
		if len(fighter.MonsterItems) != 2 {
			t.Errorf("%s 身上有 %d 件物品，應該是 2 件", fighter.Name, len(fighter.MonsterItems))
			continue
		}
		if fighter.MonsterItems[0].Name != "LONG SWORD" || !fighter.MonsterItems[0].Readied {
			t.Errorf("%s 的第一件是 %+v", fighter.Name, fighter.MonsterItems[0])
		}
		armed++
		if seen[&fighter.MonsterItems[0]] {
			t.Errorf("%s 與別隻共用同一份物品鏈", fighter.Name)
		}
		seen[&fighter.MonsterItems[0]] = true
	}
	if armed != 2 {
		t.Errorf("帶著物品的怪有 %d 隻，應該是 2 隻", armed)
	}
}
