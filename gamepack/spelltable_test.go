package gamepack

import (
	"fmt"
	"sort"
	"testing"
)

// 這一組測試把**規格上寫過的每一個數字**釘在committed 的 JSON 上。
//
// ★ 為什麼要釘：法術表是從資料段 dump 算出來的，重跑 `cmd/spell-table-export`
// 只要基底或筆數寫錯，輸出仍然是一份「看起來合理」的 JSON——欄位齊全、值域正常、
// 沒有任何錯誤訊息。唯一擋得住的方式是拿原作已經被逐條驗過的事實來對。
//
// ⚠ 這些數字**不是快照**。每一條都能回指一份規格與一個原作行為
// （某支法術的環數、AD&D 的施法節數、哪幾支只能紮營施放）。
// 「有 100 筆」那種純計數只留一條，其餘全部是可回推的語意。

func loadSpells(t *testing.T) *SpellTable {
	t.Helper()
	table, err := Spells()
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func spellByID(t *testing.T, spellID int) SpellEntry {
	t.Helper()
	entry, ok := SpellByID(spellID)
	if !ok {
		t.Fatalf("spell %d is missing from the table", spellID)
	}
	return entry
}

// 表的長度是**量出來的**，不是估的（spec 815）。索引 101 起是別的資料，
// 而那些位元組解出來一樣像法術——spec 1009 就是這樣把索引 113 當成一支法術，
// 五種目標模式的統計因此整組偏高。
func TestSpellTableCoversTheMeasuredHundredEntries(t *testing.T) {
	table := loadSpells(t)
	if len(table.Spells) != 100 {
		t.Fatalf("表有 %d 筆，量測值是 100（spec 815）", len(table.Spells))
	}
	for index, spell := range table.Spells {
		if spell.SpellID != index+1 {
			t.Fatalf("第 %d 筆的編號是 %d；編號就是索引，不能有洞", index, spell.SpellID)
		}
		if len(spell.Raw) != 32 {
			t.Fatalf("法術 %d 的 raw 是 %q，16 bytes 應該是 32 個十六進位字元",
				spell.SpellID, spell.Raw)
		}
	}
}

// spec 1016 逐條讀出 `+0` ＝ 施法職業、`+1` ＝ 環數，並與 AD&D 1e 對過。
// 兩支 Hold Person 分屬牧師 2 環與法師 3 環是最有力的一條：同名不同表項。
func TestSpellClassAndLevelMatchTheAdndCrossCheck(t *testing.T) {
	samples := []struct {
		id    int
		name  string
		class string
		level int
	}{
		{1, "Bless", "cleric", 1},
		{3, "Cure Light Wounds", "cleric", 1},
		{5, "Detect Magic", "cleric", 1},
		{11, "Detect Magic", "magic-user", 1},
		{21, "Sleep", "magic-user", 1},
		{23, "Hold Person", "cleric", 2},
		{25, "Silence, 15' Radius", "cleric", 2},
		{47, "Fireball", "magic-user", 3},
		{49, "Hold Person", "magic-user", 3},
		{51, "Lightning Bolt", "magic-user", 3},
		{79, "Faerie Fire", "druid", 1},
	}
	for _, sample := range samples {
		spell := spellByID(t, sample.id)
		if spell.Name != sample.name {
			t.Fatalf("法術 %d 的名字是 %q，want %q（名稱表基底或筆距錯了）",
				sample.id, spell.Name, sample.name)
		}
		if spell.CasterClass != sample.class || spell.Level != sample.level {
			t.Fatalf("法術 %d（%s）是 %s %d 環，want %s %d 環",
				sample.id, spell.Name, spell.CasterClass, spell.Level, sample.class, sample.level)
		}
	}
}

// 四個職業碼的分佈與 13 筆占位（spec 815）。占位那幾筆**不能從表裡刪掉**：
// 編號就是索引，刪一筆後面全錯。
func TestPlaceholderSpellsAreTheThirteenUnnamedEntries(t *testing.T) {
	table := loadSpells(t)
	classes := map[int]int{}
	var placeholders []int
	for _, spell := range table.Spells {
		classes[spell.CasterClassID]++
		if spell.Placeholder {
			placeholders = append(placeholders, spell.SpellID)
		}
	}
	want := map[int]int{0: 36, 1: 4, 2: 45, 3: 15}
	for class, count := range want {
		if classes[class] != count {
			t.Fatalf("職業碼 %d 有 %d 支，spec 815 量到 %d 支", class, classes[class], count)
		}
	}
	wantPlaceholders := []int{57, 59, 60, 61, 62, 63, 64, 65, 95, 96, 97, 98, 99}
	if fmt.Sprint(placeholders) != fmt.Sprint(wantPlaceholders) {
		t.Fatalf("占位的編號是 %v，spec 815 量到 %v", placeholders, wantPlaceholders)
	}
	// 15 支職業碼 3 裡有 13 支是占位，另兩支是 Animate Dead。
	for _, id := range []int{36, 90} {
		spell := spellByID(t, id)
		if spell.Name != "Animate Dead" || spell.CasterClassID != 3 {
			t.Fatalf("法術 %d 是 %q（職業碼 %d），want Animate Dead／3",
				id, spell.Name, spell.CasterClassID)
		}
	}
}

// `+0Bh = 0` 是「只能紮營施放」，而那 8 支**正好是 AD&D 裡的非戰鬥法術**
// （spec 827）。這條同時擋住兩種錯：欄位位移錯、以及把 0 誤讀成「不能施放」。
func TestCampOnlySpellsAreTheEightNonCombatSpells(t *testing.T) {
	table := loadSpells(t)
	var campOnly []string
	for _, spell := range table.Spells {
		if spell.CampOnly {
			campOnly = append(campOnly, spell.Name)
		}
	}
	want := []string{
		"Friends", "Read Magic", "Find Traps", "Knock",
		"Strength", "Cure Disease", "Neutralize Poison", "Raise Dead",
	}
	if fmt.Sprint(campOnly) != fmt.Sprint(want) {
		t.Fatalf("只能紮營施放的是 %v，spec 827 列的是 %v", campOnly, want)
	}
	// 其餘 92 支的值只有 1 或 2 兩種；兩者的差別原作沒讀出來，
	// **不要**在 remake 這側替它取語意。
	counts := map[int]int{}
	for _, spell := range table.Spells {
		counts[spell.CombatUse]++
	}
	if counts[0] != 8 || counts[1] != 44 || counts[2] != 48 {
		t.Fatalf("`+0Bh` 的分佈是 %v，spec 827 量到 0:8 1:44 2:48", counts)
	}
}

// 施法時間的單位換算與 AD&D 完全對得上（1 回合 ＝ 10 節、1 大回合 ＝ 100 節），
// 而原作實際用的是 `div 3`（spec 827）。兩件事分開驗：資料對不對、換算對不對。
func TestCastingTimeMatchesAdndSegmentsAndTheDivideByThreeDelay(t *testing.T) {
	samples := []struct {
		id       int
		segments int
		delay    int
	}{
		{15, 1, 0},   // Magic Missile：1 節 ⇒ 當回合結算
		{21, 1, 0},   // Sleep
		{47, 3, 1},   // Fireball：3 節
		{51, 3, 1},   // Lightning Bolt
		{1, 10, 3},   // Bless：1 回合 ＝ 10 節
		{39, 100, 33}, // Cure Disease：1 大回合 ＝ 100 節
	}
	for _, sample := range samples {
		spell := spellByID(t, sample.id)
		if spell.CastingTimeSegments != sample.segments {
			t.Fatalf("%s 的施法時間是 %d 節，want %d",
				spell.Name, spell.CastingTimeSegments, sample.segments)
		}
		if spell.CastingDelay != sample.delay {
			t.Fatalf("%s 的延遲是 %d，`有號(%d) div 3` 應該是 %d",
				spell.Name, spell.CastingDelay, sample.segments, sample.delay)
		}
	}
	counts := map[int]int{}
	for _, spell := range loadSpells(t).Spells {
		counts[spell.CastingTimeSegments]++
	}
	want := map[int]int{0: 12, 1: 17, 2: 5, 3: 13, 4: 13, 5: 13, 6: 4, 7: 6, 8: 4, 10: 11, 100: 2}
	if fmt.Sprint(sortedPairs(counts)) != fmt.Sprint(sortedPairs(want)) {
		t.Fatalf("施法時間分佈是 %v，spec 827 量到 %v", sortedPairs(counts), sortedPairs(want))
	}
}

// 目標模式（`+6` 低 nibble，spec 1009）。固定目標數與半徑階序都與 1e 對得上：
// 牧師版 Hold Person 影響 3 人、法師版 4 人；Fireball 半徑 3 > Bless 2 >
// Sleep 1 > Lightning Bolt 0（線狀法術的形狀由別處決定）。
func TestTargetModesReproduceTheAdndShapes(t *testing.T) {
	fixed := []struct {
		id    int
		count int
	}{
		{3, 1},  // Cure Light Wounds
		{10, 1}, // Charm Person
		{23, 3}, // Hold Person（牧師）
		{49, 4}, // Hold Person（法師）
		{94, 4}, // Hold Monsters
	}
	for _, sample := range fixed {
		spell := spellByID(t, sample.id)
		if spell.TargetModeKind != "fixed" || spell.TargetCount != sample.count {
			t.Fatalf("%s 的目標是 %s／%d 個，want fixed／%d 個",
				spell.Name, spell.TargetModeKind, spell.TargetCount, sample.count)
		}
	}
	area := []struct {
		id     int
		radius int
	}{
		{47, 3}, // Fireball
		{1, 2},  // Bless
		{87, 2}, // Ice Storm
		{21, 1}, // Sleep
		{91, 1}, // Cloud Kill
		{51, 0}, // Lightning Bolt
		{92, 0}, // Cone of Cold
	}
	for _, sample := range area {
		spell := spellByID(t, sample.id)
		if spell.TargetModeKind != "area" || spell.AreaRadius != sample.radius {
			t.Fatalf("%s 的目標是 %s／半徑 %d，want area／半徑 %d",
				spell.Name, spell.TargetModeKind, spell.AreaRadius, sample.radius)
		}
	}
	// 模式 5（逐一點選、累計等級權重到 2d4 上限）只有兩支。
	var weighted []string
	for _, spell := range loadSpells(t).Spells {
		if spell.TargetModeKind == "weighted-picks" {
			weighted = append(weighted, spell.Name)
		}
	}
	if fmt.Sprint(weighted) != fmt.Sprint([]string{"Faerie Fire", "Charm Monsters"}) {
		t.Fatalf("模式 5 的法術是 %v，want [Faerie Fire Charm Monsters]", weighted)
	}
	// 模式 0Fh 在 100 筆裡**只有一支**。spec 1009 原本寫兩支，第二支是索引 113
	// ——那已經在表外。
	var locked []int
	for _, spell := range loadSpells(t).Spells {
		if spell.TargetModeKind == "locked-or-area" {
			locked = append(locked, spell.SpellID)
		}
	}
	if fmt.Sprint(locked) != fmt.Sprint([]int{25}) {
		t.Fatalf("模式 0Fh 的法術是 %v，表裡只有 25（Silence, 15' Radius）", locked)
	}
	// 那一支在資料裡帶了半徑 1，原作卻因為先 `and 0Fh` 再 `shr 4` 永遠拿到 0。
	// remake 要不要修是設計決定；這條只確保**資料還在**，修不修都看得到。
	if silence := spellByID(t, 25); silence.DeclaredRadius != 1 {
		t.Fatalf("Silence, 15' Radius 的宣告半徑是 %d，資料裡是 1", silence.DeclaredRadius)
	}
}

// `+9` 是豁免類別。表裡只出現 0 與 4，而 0 那幾筆全是毒性法術——
// 與 AD&D 1e 的第 0 類（麻痺／毒／死亡魔法）吻合，也就對上 remake 角色記錄的
// `SavingThrows[0..4]` 五個位元組。
func TestPoisonSpellsUseTheFirstSavingThrowCategory(t *testing.T) {
	var category0 []string
	for _, spell := range loadSpells(t).Spells {
		if spell.SaveCategory == 0 {
			label := spell.Name
			if spell.Placeholder {
				label = fmt.Sprintf("<%d>", spell.SpellID)
			}
			category0 = append(category0, label)
		}
	}
	want := []string{"Stinking Cloud", "<61>", "Poison"}
	if fmt.Sprint(category0) != fmt.Sprint(want) {
		t.Fatalf("豁免類別 0 的是 %v，want %v", category0, want)
	}
}

// `+8 = 2` 與噴吐武器共用同一個傷害參數值，而這六筆在 1e 裡正好都是
// 「豁免成功傷害減半」。`+8 = 0` 則是**連擲都不擲**（spec 731）。
func TestSaveHalvingSpellsAreTheDamageSpells(t *testing.T) {
	var halving []string
	for _, spell := range loadSpells(t).Spells {
		if !spell.SaveHalvesDamage() {
			continue
		}
		label := spell.Name
		if spell.Placeholder {
			label = fmt.Sprintf("<%d>", spell.SpellID)
		}
		halving = append(halving, label)
	}
	want := []string{"Fireball", "Lightning Bolt", "<60>", "<64>", "Flame Strike", "Cone of Cold"}
	if fmt.Sprint(halving) != fmt.Sprint(want) {
		t.Fatalf("豁免減半的是 %v，want %v", halving, want)
	}
	// Cloudkill 在 1e 沒有豁免（低 HD 直接死），表上 `+8` 就是 0。
	if cloudkill := spellByID(t, 91); cloudkill.RequiresSave() {
		t.Fatalf("Cloud Kill 的 `+8` 是 %d，原作是 0（不擲豁免）", cloudkill.SaveKind)
	}
}

// 持續時間的兩組係數與兩個「補成 1」的修正（spec 705）。
func TestDurationFormulaReproducesTheOriginalCorrections(t *testing.T) {
	bless := spellByID(t, 1)
	// Bless：`0 × 等級 + 6`，等級不影響。
	for _, level := range []int{1, 6, 12} {
		if got := bless.PrimaryDuration(level, false); got != 6 {
			t.Fatalf("Bless 在 %d 級的持續時間是 %d，want 6", level, got)
		}
	}
	// 效果來自物品時等級一律當 6（`DS:7563h` 非 0，spec 705／1016）。
	fireball := spellByID(t, 47)
	if got, want := fireball.PrimaryDuration(3, true), 1*6+10; got != want {
		t.Fatalf("Fireball 由物品發動時是 %d，want %d（等級寫死 6）", got, want)
	}
	// `+2 = 0FFh` 是哨兵：算出來等於 0FFh 就補成 1。七支「Cause／Shocking」
	// 系的法術都是這個值。
	for _, id := range []int{4, 20, 38, 40, 44, 72, 80} {
		spell := spellByID(t, id)
		if spell.DurationPrimary.Base != 0xFF {
			t.Fatalf("法術 %d（%s）的 `+2` 是 %d，這一組應該是 0FFh",
				id, spell.Name, spell.DurationPrimary.Base)
		}
		if got := spell.PrimaryDuration(5, false); got != 1 {
			t.Fatalf("法術 %d 的持續時間是 %d，哨兵值應該補成 1", id, got)
		}
	}
}

// game pack 手寫的 12 支法術宣告，必須與原作表**逐欄相符**。
//
// ★ 這條是整組測試的重點。手寫宣告的每一個數字原作都已經有值了，
// 兩邊各存一份就會漂移，而漂移的方向永遠是「手寫的那份看起來還好，
// 只是跟原作不一樣」——玩家看不出來，測試也不會紅。
func TestCombatPlayerSpellsAgreeWithTheOriginalTable(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	classNames := map[string]string{"cleric": "cleric", "magic_user": "magic-user", "druid": "druid"}
	if len(pack.CombatPlayerSpells) == 0 {
		t.Fatal("pack 沒有宣告任何 combat_player_spells")
	}
	for _, declared := range pack.CombatPlayerSpells {
		spell, ok := SpellByID(int(declared.SpellID))
		if !ok {
			t.Fatalf("%s 宣告的 spell_id %d 不在原作表裡", declared.ID, declared.SpellID)
		}
		if spell.Placeholder {
			t.Fatalf("%s 指到占位項 %d，玩家取不到那一格", declared.ID, declared.SpellID)
		}
		if want := classNames[declared.CasterClass]; want != spell.CasterClass {
			t.Fatalf("%s 宣告 caster_class=%q，原作表是 %q",
				declared.ID, declared.CasterClass, spell.CasterClass)
		}
		if int(declared.CastingTime) != spell.CastingTimeSegments {
			t.Fatalf("%s 宣告 casting_time=%d，原作表是 %d 節",
				declared.ID, declared.CastingTime, spell.CastingTimeSegments)
		}
		if spell.CampOnly {
			t.Fatalf("%s 是只能紮營施放的法術，不該出現在戰鬥法術清單", declared.ID)
		}
	}
}

func sortedPairs(counts map[int]int) [][2]int {
	pairs := make([][2]int, 0, len(counts))
	for key, value := range counts {
		pairs = append(pairs, [2]int{key, value})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })
	return pairs
}
