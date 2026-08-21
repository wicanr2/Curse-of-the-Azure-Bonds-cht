package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// scrollCatalog 是原版 `ITEMS` 那張表的最小版本：只有槽這一格會被卷軸判別讀。
// `3Ch` 的槽是 `0Ah`——**它不是卷軸**（spec 1171）。
func scrollCatalog() monster.BaseItemCatalog {
	items := make([]monster.BaseItem, 128)
	for index := range items {
		items[index] = monster.BaseItem{Type: uint8(index), Slot: 0x01}
	}
	items[0x3C].Slot = 0x0A
	items[0x3D].Slot = 0x0B
	items[0x3E].Slot = monster.ClericalScrollSlot
	return monster.BaseItemCatalog{Items: items}
}

// scrollItem 造一張卷軸：名字編號 `+30h` 從 `0D4h` 起算，三個 Affects 是三支法術。
func scrollItem(name string, itemType uint8, hidden uint8, spells [3]uint8) monster.ItemRecord {
	return monster.ItemRecord{
		Name: name, Type: itemType, HiddenNameFlags: hidden,
		NameNumbers: [3]uint8{0, 0xD4, 0}, Affects: spells,
	}
}

func scrollState(t *testing.T, class party.Class, levels [8]uint8, items []monster.ItemRecord) *State {
	t.Helper()
	state := useItemState(t, items, 0x11)
	state.SetItemCatalog(scrollCatalog())
	state.partyRoster[0].Class = class
	state.partyRoster[0].ClassLevels = levels
	return state
}

// ★ `3Ch` 名字叫 `Scroll`，但原作按槽判 ⇒ 它是**充能物品**，不是卷軸。
func TestScrollPredicateUsesTheSlotNotTheItemType(t *testing.T) {
	catalog := scrollCatalog()
	if catalog.IsScroll(0x3C) {
		t.Fatal("`3Ch` 的槽是 `0Ah`，不該被判成卷軸")
	}
	for _, itemType := range []uint8{0x3D, 0x3E} {
		if !catalog.IsScroll(itemType) {
			t.Fatalf("`%02Xh` 應該是卷軸", itemType)
		}
	}
}

// 隱藏名稱旗標非 0 ⇒ 沒有法術辨識就讀不出來（原作連選單都不會出現）。
func TestUnidentifiedScrollListsNoSpellUntilReadMagic(t *testing.T) {
	levels := [8]uint8{}
	levels[5] = 6 // 法師
	state := scrollState(t, party.ClassMagicUser, levels, []monster.ItemRecord{
		scrollItem("Magic User Scroll", 0x3D, 0x06, [3]uint8{0x0F, 0x53, 0x5D})})
	if got := state.CombatScrollSpells(0); len(got) != 0 {
		t.Fatalf("沒辨識過的卷軸不該列得出法術：%+v", got)
	}
	if _, err := state.battle.CastEffectSpell("p1", []string{"p1"}, combat.EffectSpellRequest{
		SpellID: 0x1D, EffectKind: readMagicAffectKind, Duration: 20, CasterLevel: 6,
	}); err != nil {
		t.Fatal(err)
	}
	spells := state.CombatScrollSpells(0)
	if len(spells) != 3 {
		t.Fatalf("法術辨識之後應該列出三支，實際 %d", len(spells))
	}
	if spells[0].SpellID != 0x0F {
		t.Fatalf("第一支應該是 `0Fh`，實際 %02Xh", spells[0].SpellID)
	}
}

// 牧師讀牧師卷軸（槽 `0Ch`）不需要法術辨識。
func TestClericReadsClericalScrollWithoutReadMagic(t *testing.T) {
	levels := [8]uint8{}
	levels[0] = 5 // 牧師
	state := scrollState(t, party.ClassCleric, levels, []monster.ItemRecord{
		scrollItem("Clerical Scroll", 0x3E, 0x06, [3]uint8{0x03, 0, 0})})
	if got := state.CombatScrollSpells(0); len(got) != 1 {
		t.Fatalf("牧師應該讀得出牧師卷軸：%+v", got)
	}
	// 同一張卷軸換成法師卷軸的槽就讀不出來了。
	state.partyRoster[0].Equipment[0].Type = 0x3D
	if got := state.CombatScrollSpells(0); len(got) != 0 {
		t.Fatalf("法師卷軸不該對牧師開放：%+v", got)
	}
}

// ★★ 消耗：清掉那個槽、名字編號減一；**唸完第三支**才把卷軸丟掉。
func TestScrollIsDiscardedOnlyAfterTheThirdSpell(t *testing.T) {
	levels := [8]uint8{}
	levels[0] = 8
	state := scrollState(t, party.ClassCleric, levels, []monster.ItemRecord{
		scrollItem("Clerical Scroll", 0x3E, 0, [3]uint8{0x03, 0x03, 0x03})})
	for round := 1; round <= 3; round++ {
		if _, ok := state.combatPartyTurn(); !ok {
			t.Skipf("第 %d 次時已經不是隊員的回合", round)
		}
		if err := state.CombatUseItem(); err != nil {
			t.Fatalf("第 %d 次唸卷軸：%v", round, err)
		}
		if round == 3 {
			break
		}
		if len(state.partyRoster[0].Equipment) != 1 {
			t.Fatalf("第 %d 次之後卷軸不該消失", round)
		}
		item := state.partyRoster[0].Equipment[0]
		if want := scrollNameNumberBase - uint8(round); item.NameNumbers[1] != want {
			t.Fatalf("第 %d 次之後名字編號 ＝ %02X，want %02X", round, item.NameNumbers[1], want)
		}
	}
	if len(state.partyRoster[0].Equipment) != 0 {
		t.Fatalf("唸完第三支之後卷軸應該消失：%+v", state.partyRoster[0].Equipment)
	}
}

// 非施法職業唸不動：印 `oops!`，而且**卷軸不會被消耗**（spec 921）。
func TestFighterFailsAndKeepsTheScroll(t *testing.T) {
	levels := [8]uint8{}
	levels[2] = 9 // 戰士
	state := scrollState(t, party.ClassFighter, levels, []monster.ItemRecord{
		scrollItem("Magic User Scroll", 0x3D, 0, [3]uint8{0x0F, 0, 0})})
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用卷軸：%v", err)
	}
	if len(state.partyRoster[0].Equipment) != 1 {
		t.Fatal("失敗的那一次不該消耗卷軸")
	}
	if got := state.partyRoster[0].Equipment[0]; got.Affects[0] != 0x0F || got.NameNumbers[1] != 0xD4 {
		t.Fatalf("失敗的那一次不該動到卷軸：%+v", got)
	}
}

// ★★ 十級以上盜賊有 75% 機率——1e 的規則就是這一條。九級以下必定失敗。
func TestThiefNeedsLevelTenToTryAScroll(t *testing.T) {
	for _, item := range []struct {
		level uint8
		tries bool
	}{{level: 9, tries: false}, {level: 10, tries: true}} {
		levels := [8]uint8{}
		levels[6] = item.level
		state := scrollState(t, party.ClassThief, levels, []monster.ItemRecord{
			scrollItem("Magic User Scroll", 0x3D, 0, [3]uint8{0x0F, 0, 0})})
		succeeded := 0
		for attempt := 0; attempt < 40; attempt++ {
			if state.scrollCastSucceeds(0) {
				succeeded++
			}
		}
		if item.tries && succeeded == 0 {
			t.Fatalf("十級盜賊應該有機會成功，40 次一次都沒有")
		}
		if !item.tries && succeeded != 0 {
			t.Fatalf("九級盜賊應該必定失敗，卻成功了 %d 次", succeeded)
		}
	}
}
