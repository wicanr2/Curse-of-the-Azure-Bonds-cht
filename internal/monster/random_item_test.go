package monster

import "testing"

// scriptedRolls 依序餵回預先排好的骰值，並檢查骰的面數符合預期，
// 讓「哪一顆骰先擲」也一起被釘住——原作的擲骰順序決定了同一條亂數流
// 會產生什麼，順序錯了結果就跟原版分岔。
type scriptedRolls struct {
	t     *testing.T
	sides []int
	value []int
	index int
}

func (s *scriptedRolls) roll(sides int) int {
	s.t.Helper()
	if s.index >= len(s.value) {
		s.t.Fatalf("第 %d 次擲骰（1d%d）超出腳本", s.index+1, sides)
	}
	if want := s.sides[s.index]; sides != want {
		s.t.Fatalf("第 %d 次擲骰是 1d%d，預期 1d%d", s.index+1, sides, want)
	}
	value := s.value[s.index]
	s.index++
	return value
}

func TestBuildRandomTreasureItem(t *testing.T) {
	tests := []struct {
		name     string
		itemType uint8
		sides    []int
		values   []int
		want     ItemRecord
	}{{
		name: "一般武器：namenum(2) 是 0A1h＋plus", itemType: 0x24,
		sides: []int{20}, values: []int{10},
		want: ItemRecord{Type: 0x24, NameNumbers: [3]uint8{0, 0xA2, 0x24},
			Plus: 1, HiddenNameFlags: 6, Weight: 60, Value: 2000},
	}, {
		name: "plus 分界：1d20 ＝ 14 仍是 ＋1", itemType: 0x24,
		sides: []int{20}, values: []int{14},
		want: ItemRecord{Type: 0x24, NameNumbers: [3]uint8{0, 0xA2, 0x24},
			Plus: 1, HiddenNameFlags: 6, Weight: 60, Value: 2000},
	}, {
		name: "plus 分界：1d20 ＝ 15 起是 ＋2", itemType: 0x24,
		sides: []int{20}, values: []int{15},
		want: ItemRecord{Type: 0x24, NameNumbers: [3]uint8{0, 0xA3, 0x24},
			Plus: 2, HiddenNameFlags: 6, Weight: 60, Value: 4000},
	}, {
		// ★ 這一格是 spec 1036 重量表的死分支：`2Fh` 在 80 那一列就被攔下，
		// 後面「`2Fh`／`5Dh` ⇒ 1」的 `2Fh` 永遠走不到。
		name: "2Fh 的重量是 80，不是 1", itemType: 0x2F,
		sides: []int{20}, values: []int{1},
		want: ItemRecord{Type: 0x2F, NameNumbers: [3]uint8{0, 0xA2, 0x2F},
			Plus: 1, HiddenNameFlags: 6, Weight: 80, Value: 2000},
	}, {
		name: "5Dh 的重量才是 1", itemType: 0x5D,
		sides: []int{20}, values: []int{1},
		want: ItemRecord{Type: 0x5D, NameNumbers: [3]uint8{0xA2, 0xE0, 0x42},
			Plus: 1, HiddenNameFlags: 6, Weight: 1, Value: 2000},
	}, {
		name: "4Dh：plus ＝ plus × 2 ＋ 2，namenum(1) 依 4／6 分", itemType: 0x4D,
		sides: []int{20}, values: []int{1},
		want: ItemRecord{Type: 0x4D, NameNumbers: [3]uint8{0xDD, 0xA7, 0x4F},
			Plus: 4, HiddenNameFlags: 6, Weight: 10, Value: 12000},
	}, {
		name: "4Dh：擲到 ＋2 時 namenum(1) 換成 0DEh", itemType: 0x4D,
		sides: []int{20}, values: []int{20},
		want: ItemRecord{Type: 0x4D, NameNumbers: [3]uint8{0xDE, 0xA7, 0x4F},
			Plus: 6, HiddenNameFlags: 6, Weight: 10, Value: 18000},
	}, {
		name: "板甲 3Ah：identified ＝ 4，重量 450", itemType: 0x3A,
		sides: []int{20}, values: []int{3},
		want: ItemRecord{Type: 0x3A, NameNumbers: [3]uint8{0xA2, '0', 0x3A},
			Plus: 1, HiddenNameFlags: 4, Weight: 450, Value: 5000},
	}, {
		name: "盾 32h：namenum(2) 是字元 '1'", itemType: 0x32,
		sides: []int{20}, values: []int{3},
		want: ItemRecord{Type: 0x32, NameNumbers: [3]uint8{0xA2, '1', 0x32},
			Plus: 1, HiddenNameFlags: 4, Weight: 150, Value: 2000},
	}, {
		name: "34h：namenum(2) 是字元 '2'", itemType: 0x34,
		sides: []int{20}, values: []int{3},
		want: ItemRecord{Type: 0x34, NameNumbers: [3]uint8{0xA2, '2', 0x34},
			Plus: 1, HiddenNameFlags: 4, Weight: 200, Value: 2000},
	}, {
		name: "49h：namenum(3) 借用 3Dh，重量吃預設那一列", itemType: 0x49,
		sides: []int{20}, values: []int{3},
		want: ItemRecord{Type: 0x49, NameNumbers: [3]uint8{0, 0xA2, 0x3D},
			Plus: 1, HiddenNameFlags: 6, Weight: 40, Count: 0x0A, Value: 150},
	}, {
		name: "類別 9 另外寫 +39h ＝ 5", itemType: 0x09,
		sides: []int{20}, values: []int{3},
		want: ItemRecord{Type: 0x09, NameNumbers: [3]uint8{0, 0xA2, 0x09},
			Plus: 1, HiddenNameFlags: 6, Weight: 25, Count: 5, Value: 2000},
	}, {
		name: "15h 的 1d5 不是 5 就只走一般那一段", itemType: 0x15,
		sides: []int{20, 5}, values: []int{3, 4},
		want: ItemRecord{Type: 0x15, NameNumbers: [3]uint8{0, 0xA2, 0x15},
			Plus: 1, HiddenNameFlags: 6, Weight: 20, Value: 2000},
	}, {
		name: "15h 的 1d5 ＝ 5 改走範本 49，整組覆蓋", itemType: 0x15,
		sides: []int{20, 5}, values: []int{3, 5},
		want: ItemRecord{Type: 0x15, NameNumbers: [3]uint8{157, 167, 21},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 20, Value: 3000,
			Affects: [3]uint8{1, 51, 0}},
	}, {
		name: "3Dh 卷軸：法術直接存進 special(1..3)", itemType: 0x3D,
		sides: []int{3, 5, 11, 5, 13}, values: []int{2, 3, 5, 1, 2},
		want: ItemRecord{Type: 0x3D, NameNumbers: [3]uint8{0, 0xD3, 0xD1},
			Plus: 1, HiddenNameFlags: 6, Weight: 25, Value: 1200,
			Affects: [3]uint8{49, 10, 0}},
	}, {
		name: "3Eh 卷軸：法術編號區間與 3Dh 不重疊", itemType: 0x3E,
		sides: []int{3, 5, 6}, values: []int{1, 5, 6},
		want: ItemRecord{Type: 0x3E, NameNumbers: [3]uint8{0, 0xD2, 0xD0},
			Plus: 1, HiddenNameFlags: 6, Weight: 25, Value: 1500,
			Affects: [3]uint8{76, 0, 0}},
	}, {
		name: "47h 擲 1d8 ＝ 1..5 用範本 17", itemType: 0x47,
		sides: []int{8}, values: []int{5},
		want: ItemRecord{Type: 0x47, NameNumbers: [3]uint8{185, 167, 64},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 1, Value: 400,
			Affects: [3]uint8{1, 3, 0}},
	}, {
		name: "47h 擲 1d8 ＝ 6..8 用範本 1", itemType: 0x47,
		sides: []int{8}, values: []int{6},
		want: ItemRecord{Type: 0x47, NameNumbers: [3]uint8{185, 187, 64},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 1, Value: 800,
			Affects: [3]uint8{3, 99, 0}},
	}, {
		name: "3Fh 用範本 41，這一筆的 special(3) 不是 0", itemType: 0x3F,
		want: ItemRecord{Type: 0x3F, NameNumbers: [3]uint8{226, 167, 100},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 10, Value: 15000,
			Affects: [3]uint8{0, 38, 131}},
	}, {
		name: "54h 用範本 9", itemType: 0x54,
		want: ItemRecord{Type: 0x54, NameNumbers: [3]uint8{239, 167, 64},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 1, Value: 1100,
			Affects: [3]uint8{1, 59, 0}},
	}, {
		name: "4Eh 用範本 33", itemType: 0x4E,
		want: ItemRecord{Type: 0x4E, NameNumbers: [3]uint8{206, 167, 69},
			Plus: 1, PlusSave: 1, HiddenNameFlags: 6, Weight: 1, Value: 11000,
			Affects: [3]uint8{30, 15, 0}},
	}, {
		// 三條路都沒接住的類別：原作只留骨架，一顆骰都不擲。
		name: "沒有對應分支的類別只留骨架", itemType: 0x60,
		want: ItemRecord{Type: 0x60, HiddenNameFlags: 6},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rolls := &scriptedRolls{t: t, sides: test.sides, value: test.values}
			got := BuildRandomTreasureItem(test.itemType, rolls.roll)
			if got != test.want {
				t.Fatalf("造出來的物品 ＝ %#v，預期 %#v", got, test.want)
			}
			if rolls.index != len(test.values) {
				t.Fatalf("擲了 %d 顆骰，腳本排了 %d 顆", rolls.index, len(test.values))
			}
		})
	}
}

// TestBuildRandomTreasureItemScrollRings 逐格釘住兩種卷軸的環→法術編號表。
// 每一格都取該環骰的最小值與最大值，所以區間兩端與偏移量都被夾住。
func TestBuildRandomTreasureItemScrollRings(t *testing.T) {
	tests := []struct {
		itemType   uint8
		ring       int
		spellSides int
		low, high  uint8
	}{
		{0x3D, 1, 13, 9, 21},
		{0x3D, 2, 7, 29, 35},
		{0x3D, 3, 11, 45, 55},
		{0x3D, 4, 9, 81, 89},
		{0x3D, 5, 4, 91, 94},
		{0x3E, 1, 8, 1, 8},
		{0x3E, 2, 7, 22, 28},
		{0x3E, 3, 8, 37, 44},
		{0x3E, 4, 5, 66, 70},
		{0x3E, 5, 6, 71, 76},
	}
	for _, test := range tests {
		for _, bound := range []struct {
			spell uint8
			face  int
		}{{test.low, 1}, {test.high, test.spellSides}} {
			rolls := &scriptedRolls{t: t,
				sides: []int{3, 5, test.spellSides},
				value: []int{1, test.ring, bound.face}}
			item := BuildRandomTreasureItem(test.itemType, rolls.roll)
			if item.Affects[0] != bound.spell {
				t.Fatalf("類別 %02Xh 第 %d 環擲 %d ⇒ 法術 %d，預期 %d",
					test.itemType, test.ring, bound.face, item.Affects[0], bound.spell)
			}
			if want := int16(test.ring) * 300; item.Value != want {
				t.Fatalf("類別 %02Xh 第 %d 環的價值 ＝ %d，預期 %d",
					test.itemType, test.ring, item.Value, want)
			}
		}
	}
}

// TestRandomItemWeightAndValueTables 逐列釘住重量表與價值係數表。
// 兩張表都是從 `overlay-21:1282h`／`14C5h` 那兩條 if-else 鏈直接抄下來的，
// 抄漏一格不會有任何症狀，只會讓某個類別安靜地掉進預設列。
func TestRandomItemWeightAndValueTables(t *testing.T) {
	weights := map[int16][]uint8{
		10:  {0x08, 0x4D},
		15:  {0x06},
		20:  {0x15},
		25:  {0x09},
		30:  {0x07},
		35:  {0x25},
		40:  {0x16, 0x1E},
		50:  {0x02, 0x14, 0x1D, 0x1F, 0x20, 0x21, 0x27, 0x2A, 0x2C, 0x2E, 0x3B},
		60:  {0x0A, 0x1A, 0x24},
		75:  {0x01, 0x0D, 0x0E, 0x23},
		80:  {0x0B, 0x10, 0x19, 0x1B, 0x29, 0x2F},
		100: {0x04, 0x0F, 0x17, 0x22, 0x2B, 0x2D, 0x33},
		125: {0x03, 0x18, 0x28},
		150: {0x05, 0x0C, 0x11, 0x13, 0x32},
		175: {0x12},
		200: {0x34},
		250: {0x26, 0x35},
		300: {0x37},
		350: {0x39},
		400: {0x36, 0x38},
		450: {0x3A},
		1:   {0x5D},
	}
	for want, types := range weights {
		for _, itemType := range types {
			if got, _ := randomItemWeight(itemType); got != want {
				t.Fatalf("類別 %02Xh 的重量 ＝ %d，預期 %d", itemType, got, want)
			}
		}
	}
	// 預設列：沒被上面接住的類別重量 40，並另寫 `+39h` ＝ 0Ah。
	if got, count := randomItemWeight(0x1C); got != 40 || count != 0x0A {
		t.Fatalf("預設列的重量／數量 ＝ %d／%d，預期 40／10", got, count)
	}
	if _, count := randomItemWeight(0x09); count != 5 {
		t.Fatalf("類別 9 的 +39h ＝ %d，預期 5", count)
	}

	values := map[int16][]uint8{
		2500: {0x3B},
		150:  {0x1C, 0x49},
		3000: {0x35, 0x36, 0x4D},
		3500: {0x37, 0x38},
		4000: {0x39},
		5000: {0x3A},
		2000: {0x01, 0x24, 0x33},
	}
	for want, types := range values {
		for _, itemType := range types {
			if got := randomItemValueCoefficient(itemType); got != want {
				t.Fatalf("類別 %02Xh 的價值係數 ＝ %d，預期 %d", itemType, got, want)
			}
		}
	}
}

// TestBuildRandomTreasureItemNamesShowPlus 釘住隨機寶物的加值怎麼進名字。
//
// ⚠ 加值走的是**名稱編號** `A2h..A6h`（`+1`..`+5`），不是 `+32h` 那個欄位
// （spec 1178）。而 `+35h` 預設 6 ＝ 藏住前兩個成分，所以剛撿到的時候玩家
// 只看得到類別名——**加值要鑑定過才會出現**，這是原作行為不是缺譯。
func TestBuildRandomTreasureItemNamesShowPlus(t *testing.T) {
	rolls := &scriptedRolls{t: t, sides: []int{20}, value: []int{15}}
	item := BuildRandomTreasureItem(0x24, rolls.roll)
	if item.NameNumbers != [3]uint8{0, 0xA3, 0x24} {
		t.Fatalf("名稱編號 ＝ %v，預期 {0, A3h, 24h}", item.NameNumbers)
	}
	text := stubTextResolver{"item_type_24": "長劍", "item_name_24": "長劍", "item_name_A3": "＋2"}
	if got := LocalizedItemName(item, text); got != "長劍" {
		t.Fatalf("未鑑定時 ＝ %q，預期 %q", got, "長劍")
	}
	item.HiddenNameFlags = 0
	if got := LocalizedItemName(item, text); got != "長劍＋2" {
		t.Fatalf("鑑定後 ＝ %q，預期 %q", got, "長劍＋2")
	}
}

type stubTextResolver map[string]string

func (s stubTextResolver) Text(key, fallback string) string {
	if value, ok := s[key]; ok {
		return value
	}
	return fallback
}
