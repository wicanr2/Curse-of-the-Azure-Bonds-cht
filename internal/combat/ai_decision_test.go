package combat

import "testing"

// `+0F7h` 的編碼：bit 7 是「有效」旗標，低 7 位乘 2 才是士氣值（spec 758）。
//
// ⚠ 把整個位元組當士氣值是最容易犯的錯：`0B3h` 看起來像 179，
// 實際是「旗標已設、原值 102」——那是資料裡最常見的一個值。
func TestMoraleValueDecodesTheHalvedByteWithItsValidBit(t *testing.T) {
	for _, item := range []struct {
		raw   uint8
		value int
		valid bool
	}{
		{0x00, 0, false}, // 旗標沒設 ⇒ 無效
		{0x7F, 0, false},
		{0xB3, 102, true}, // 80h or 33h ⇒ 51 × 2 ＝ 102
		{0x80, 0, true},   // 旗標設了但原值 0
		{0xFF, 0, true},   // 7Fh × 2 ＝ 254 > 66h ⇒ 當 0
		{0x90, 32, true},
	} {
		value, valid := MoraleValue(item.raw)
		if value != item.value || valid != item.valid {
			t.Fatalf("%02Xh → (%d, %v)，want (%d, %v)",
				item.raw, value, valid, item.value, item.valid)
		}
	}
}

// 門檻掃描的骨架：**1d7 一定要擲**，即使根本沒有候選——原作在檢查清單之前
// 就擲了，省掉它會讓後續的亂數序列整條偏掉。
func TestThresholdScanAlwaysRollsTheRoundCount(t *testing.T) {
	rolls := 0
	roll := func(sides int) int {
		rolls++
		if sides != aiThresholdRoundsDie {
			t.Fatalf("第 %d 次擲的是 %dd，第一擲應該是 1d7", rolls, sides)
		}
		return 3
	}
	seen := []int{}
	if _, ok := AIThresholdScan(roll, func(threshold int) (uint8, bool) {
		seen = append(seen, threshold)
		return 0, false
	}); ok {
		t.Fatal("沒有候選卻回報選到了")
	}
	if rolls != 1 {
		t.Fatalf("擲了 %d 次骰，輪數只該擲一次", rolls)
	}
	// 門檻從 7 開始每輪降 1，掃 3 輪 ⇒ 7、6、5。
	if len(seen) != 3 || seen[0] != 7 || seen[1] != 6 || seen[2] != 5 {
		t.Fatalf("門檻序列=%v，want [7 6 5]", seen)
	}
}

func aiSpellBattle(t *testing.T, spells []uint8, morale uint8, seed int64) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 3,
			MonsterSpellIDs: spells, ControlMorale: morale},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10},
	}, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 原作的 off-by-one（AI 看不到 `+1Eh` 那一格）**不在這一層重現**：
// `MonsterSpellIDs` 是壓縮過的清單，索引 0 不等於 `+1Eh`，跳過它等於隨機丟掉
// 一支真的法術。這條測試釘住「不跳過」，免得日後有人照著 spec 836 的字面補上去。
func TestSpellCandidatesDoNotSkipTheCompactedFirstEntry(t *testing.T) {
	battle := aiSpellBattle(t, []uint8{47, 0, 0}, 0xB3, 1)
	priority := func(id uint8) (int, bool) {
		if id == 47 {
			return 7, true
		}
		return 0, false
	}
	found := false
	for attempt := 0; attempt < 50 && !found; attempt++ {
		choice, ok, err := battle.AIChooseSpell("caster", priority)
		if err != nil {
			t.Fatal(err)
		}
		if ok && choice == 47 {
			found = true
		}
	}
	if !found {
		t.Fatal("壓縮清單的第一支法術一次都沒被選到——off-by-one 被錯套在這一層了")
	}
}

// 分數夠高就選得到；分數不足時要靠門檻降下來才有機會。
func TestAiPicksAHighPriorityCandidate(t *testing.T) {
	battle := aiSpellBattle(t, []uint8{0, 47}, 0xB3, 7)
	priority := func(id uint8) (int, bool) {
		if id == 47 {
			return 7, true
		}
		return 0, false
	}
	choice, found, err := battle.AIChooseSpell("caster", priority)
	if err != nil {
		t.Fatal(err)
	}
	if !found || choice != 47 {
		t.Fatalf("choice=%d found=%v，分數 7 在第一輪就該過門檻", choice, found)
	}
}

// 查不到分數的候選一律略過——**不能當成分數 0**，那會讓沒有資料的法術在門檻
// 降到 0 之後突然變成合法選擇。
func TestAiSkipsCandidatesWithoutAPriorityScore(t *testing.T) {
	battle := aiSpellBattle(t, []uint8{0, 200}, 0xB3, 3)
	unknown := func(uint8) (int, bool) { return 0, false }
	for attempt := 0; attempt < 20; attempt++ {
		if _, found, err := battle.AIChooseSpell("caster", unknown); err != nil {
			t.Fatal(err)
		} else if found {
			t.Fatal("查不到分數的候選被選走了")
		}
	}
}

// 士氣崩了（`+0F7h` 有效但值為 0）就不施法。
func TestBrokenMoraleStopsSpellcasting(t *testing.T) {
	battle := aiSpellBattle(t, []uint8{0, 47}, 0x80, 5)
	choice, found, err := battle.AIChooseSpell("caster",
		func(uint8) (int, bool) { return 7, true })
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatalf("士氣崩了還選了法術 %d", choice)
	}
}

// 用道具那一支每輪**全掃**，第一個過門檻的就用；效果碼要套原作的重映射。
func TestAiItemScanUsesTheFirstEligibleEffectAndRemapsIt(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "user", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 2,
			ReadiedItemEffects: []uint8{0x40, 0x12}},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10},
	}, 9)
	if err != nil {
		t.Fatal(err)
	}
	// 0x40 > 0x38 ⇒ 減 0x17 ＝ 0x29；0x12 不變。只有 0x29 有分數。
	choice, found, err := battle.AIChooseItemEffect("user", func(id uint8) (int, bool) {
		if id == 0x29 {
			return 7, true
		}
		return 0, false
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found || choice != 0x29 {
		t.Fatalf("choice=%02Xh found=%v，want 29h（40h 減 17h）", choice, found)
	}
}

func TestItemEffectRemapOnlyAppliesAboveTheThreshold(t *testing.T) {
	if got := AIRemapItemEffect(0x38); got != 0x38 {
		t.Fatalf("38h → %02Xh，界線是 `> 38h` 才減", got)
	}
	if got := AIRemapItemEffect(0x39); got != 0x39-0x17 {
		t.Fatalf("39h → %02Xh，want %02Xh", got, 0x39-0x17)
	}
}
