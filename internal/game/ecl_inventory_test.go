package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestECLDestroyItemsMutatesPersistentRoster(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "dark-elf", Equipment: []monster.ItemRecord{
		{Type: 0x5E, Readied: true},
		{Type: 0x23},
	}}}
	state.applyECLInventorySignals(ecl.RunResult{DestroyItemIDs: []uint16{0x5E}})
	if len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 0x23 {
		t.Fatalf("roster=%#v", state.partyRoster[0].Equipment)
	}
}

func TestEncounterTreasureResolvesOnlyAfterVictory(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 10, MaxHitPoints: 10}}
	state.applyECLTreasureSignals(ecl.RunResult{TreasureRequests: []ecl.TreasureRequest{{
		Coins: [7]uint16{0, 0, 0, 2000, 3000, 8, 4}, ItemBlock: 0x82,
	}}})
	if state.MoneyPool() != 0 || len(state.PendingTreasureItems()) != 0 {
		t.Fatal("encounter reward was granted before combat")
	}
	if err := state.StartCombat(
		[]combat.Fighter{{
			ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
			ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100,
		}},
		[]combat.Fighter{{ID: "guard", Name: "守衛", Side: combat.SideEnemy, HitPoints: 1, MaxHitPoints: 1}},
		7,
	); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.MoneyPool() != 17000 {
		t.Fatalf("victory money=%d, want 17000 GP", state.MoneyPool())
	}
	if gems, jewelry := state.TreasurePool(); gems != 8 || jewelry != 4 {
		t.Fatalf("victory treasure=(%d,%d), want (8,4)", gems, jewelry)
	}
	if !state.treasureMenu || len(state.PendingTreasureItems()) != 2 {
		t.Fatalf("victory treasure menu=%v items=%#v", state.treasureMenu, state.PendingTreasureItems())
	}
}

func TestECLTreasureResolvesMoneyAndItemBlock(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "fighter", Name: "戰士"}}
	state.SetTreasureItemBlocks(map[uint16][]monster.ItemRecord{
		0x02: {{Type: 36, Name: "長劍", Value: 100}},
	})
	state.applyECLTreasureSignals(ecl.RunResult{TreasureRequests: []ecl.TreasureRequest{{
		Coins:     [7]uint16{1, 2, 0, 2, 0, 3, 4},
		ItemBlock: 0x02,
	}}})
	if err := state.ResolveTreasureRequests(); err != nil {
		t.Fatal(err)
	}
	if state.MoneyPool() != 2 {
		t.Fatalf("money pool=%d, want 2 GP", state.MoneyPool())
	}
	if state.MoneyPoolCopperRemainder() != 21 {
		t.Fatalf("money remainder=%d copper, want 21", state.MoneyPoolCopperRemainder())
	}
	if gems, jewelry := state.TreasurePool(); gems != 3 || jewelry != 4 {
		t.Fatalf("treasure pool=(%d,%d), want (3,4)", gems, jewelry)
	}
	if len(state.PendingTreasureItems()) != 1 {
		t.Fatalf("pending items=%#v", state.PendingTreasureItems())
	}
	if err := state.TakeTreasureItem(0, 0); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 36 {
		t.Fatalf("equipment=%#v", state.partyRoster[0].Equipment)
	}
}

func TestECLTreasureCarriesSubGoldTypedCoinValue(t *testing.T) {
	state := NewState(testCatalog())
	for count := 0; count < 2; count++ {
		state.applyECLTreasureSignals(ecl.RunResult{
			TreasureRequests: []ecl.TreasureRequest{{
				Coins:     [7]uint16{0, 0, 1, 0, 0, 0, 0},
				ItemBlock: 0xFF,
			}},
		})
		if err := state.ResolveTreasureRequests(); err != nil {
			t.Fatal(err)
		}
	}
	if state.MoneyPool() != 1 || state.MoneyPoolCopperRemainder() != 0 {
		t.Fatalf("two electrum pieces pool=%d GP remainder=%d copper, want 1/0",
			state.MoneyPool(), state.MoneyPoolCopperRemainder())
	}
}

func TestECLTreasureResolvesReferenceRandomCount(t *testing.T) {
	state := NewState(testCatalog())
	state.SetECLSeed(7)
	state.applyECLTreasureSignals(ecl.RunResult{TreasureRequests: []ecl.TreasureRequest{{ItemBlock: 0x82}}})
	if err := state.ResolveTreasureRequests(); err != nil {
		t.Fatal(err)
	}
	items := state.PendingTreasureItems()
	if len(items) != 2 {
		t.Fatalf("random items=%#v, want two records", items)
	}
	// ★ 隨機那一路要跑完整的 `CREATERNDTREASURE`（spec 1036），不是只放一個類別。
	// 種子 7 這兩件：`+1` 的類別 9（重量 25、`+39h` ＝ 5、價值 ＝ 1 × 2000），
	// 與類別 `3Eh` 的一張卷軸（一個法術 ⇒ namenum(2) ＝ `0D1h+1`，
	// namenum(3) ＝ `0D0h`，第四環 1d5+65 擲到 69，價值 ＝ 4 × 300）。
	arrow := monster.ItemRecord{
		Type: 0x09, NameNumbers: [3]uint8{0, 0xA2, 0x09}, Plus: 1,
		HiddenNameFlags: 6, Weight: 25, Count: 5, Value: 2000,
	}
	scroll := monster.ItemRecord{
		Type: 0x3E, NameNumbers: [3]uint8{0, 0xD2, 0xD0}, Plus: 1,
		HiddenNameFlags: 6, Weight: 25, Value: 1200, Affects: [3]uint8{69, 0, 0},
	}
	if items[0] != arrow || items[1] != scroll {
		t.Fatalf("random items=%#v, want %#v and %#v", items, arrow, scroll)
	}
}

func TestTreasureMenuAssignsSelectedItemToCharacter(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{ID: "fighter", Name: "戰士"}}
	state.pendingTreasureItems = []monster.ItemRecord{{Type: 36, Count: 1}}
	state.enterTreasureMenu()
	if state.Prompt != catalog.Text("treasure_prompt", "") ||
		state.Message != catalog.Text("treasure_ready", "") ||
		state.Choices[len(state.Choices)-1] != catalog.Text("treasure_exit", "") {
		t.Fatalf("treasure menu prompt=%q message=%q choices=%#v", state.Prompt, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil || !state.treasureTakeMenu {
		t.Fatalf("item selection err=%v takeMenu=%v", err, state.treasureTakeMenu)
	}
	if state.Prompt != catalog.Text("treasure_take_prompt", "") ||
		state.Choices[len(state.Choices)-1] != catalog.Text("treasure_cancel", "") {
		t.Fatalf("treasure take prompt=%q choices=%#v", state.Prompt, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != catalog.Text("treasure_taken", "") ||
		len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 36 {
		t.Fatalf("mode=%d message=%q equipment=%#v", state.Mode, state.Message, state.partyRoster[0].Equipment)
	}
}

func TestTreasureMenuCancelAndSkipUseStableLocaleContract(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{ID: "fighter", Name: "DISPLAY"}}
	state.pendingTreasureItems = []monster.ItemRecord{{Type: 36, Count: 1}}
	state.enterTreasureMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil || state.treasureTakeMenu {
		t.Fatalf("cancel err=%v takeMenu=%v", err, state.treasureTakeMenu)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != catalog.Text("treasure_skipped", "") || len(state.pendingTreasureItems) != 0 {
		t.Fatalf("skip mode=%v message=%q items=%#v", state.Mode, state.Message, state.pendingTreasureItems)
	}
}
