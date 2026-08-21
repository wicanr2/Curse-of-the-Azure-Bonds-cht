package monster

// 隨機戰利品的造物品常式：原作 `CREATERNDTREASURE`
// （DOS `overlay-21:0FEDh`，PC-98 `overlay-21:010FCh`），逐條讀出見 spec 1036。
//
// `27h TREASURE` 的 `ItemBlock > 80h` 那一路只擲出**物品類別**（spec 1151），
// 名稱三段、加值、重量、價值與卷軸法術全部由這一支填。所以少了它，隨機開出來
// 的東西只有一個類別名，沒有 `+N`、沒有名稱修飾、價值是 0。
//
// ★ 名稱不是一句字串，是三段索引：`NameNumbers[0..2]` ＝ 原作的
// `namenum(1..3)`（`+2Fh`／`+30h`／`+31h`）。第 `0A1h`(161) 筆起是「＋N」後綴，
// 所以中文化那一段必須是組字。

// RollFunc 擲一顆骰子，回傳 1..sides。呼叫端擁有亂數流。
type RollFunc func(sides int) int

// BuildRandomTreasureItem 依物品類別造出一筆完整的物品記錄。
func BuildRandomTreasureItem(itemType uint8, roll RollFunc) ItemRecord {
	// 骨架：`FillChar(節點, 3Fh, 0)` 之後只有這幾格不是 0。
	// `+35h`（名稱隱藏旗標）預設 6——⚠ 不是布林。
	item := ItemRecord{Type: itemType, HiddenNameFlags: 6}

	templateIndex := 0
	switch {
	case isGeneralRandomItem(itemType):
		plus := rollItemPlus(roll)
		switch {
		case itemType == 0x15:
			// 另擲 1d5，＝5 時改走特殊範本 49（`overlay-21:10A8h`）。
			if roll(5) == 5 {
				templateIndex = 49
			}
			item.NameNumbers = [3]uint8{0, uint8(plus + 0xA1), 0x15}
		case itemType == 0x1C:
			item.NameNumbers = [3]uint8{0, uint8(plus + 0xA1), 0x1C}
		case itemType == 0x32 || itemType == 0x33:
			item.NameNumbers = [3]uint8{uint8(plus + 0xA1), '1', itemType}
			item.HiddenNameFlags = 4
		case itemType == 0x34:
			item.NameNumbers = [3]uint8{uint8(plus + 0xA1), '2', itemType}
			item.HiddenNameFlags = 4
		case itemType >= 0x35 && itemType <= 0x3A:
			item.NameNumbers = [3]uint8{uint8(plus + 0xA1), '0', itemType}
			item.HiddenNameFlags = 4
		case itemType == 0x49:
			item.NameNumbers = [3]uint8{0, uint8(plus + 0xA1), 0x3D}
		case itemType == 0x4D:
			// ★ `plus := plus × 2 ＋ 2` ⇒ 4 或 6，再依值挑 namenum(1)。
			plus = plus*2 + 2
			first := uint8(0)
			if plus == 4 {
				first = 0xDD
			} else if plus == 6 {
				first = 0xDE
			}
			item.NameNumbers = [3]uint8{first, 0xA7, 0x4F}
		case itemType == 0x5D:
			item.NameNumbers = [3]uint8{uint8(plus + 0xA1), 0xE0, 0x42}
		default:
			item.NameNumbers = [3]uint8{0, uint8(plus + 0xA1), itemType}
		}
		item.Plus = plus
		weight, count := randomItemWeight(itemType)
		item.Weight = weight
		item.Count = count
		item.Value = int16(plus) * randomItemValueCoefficient(itemType)
	case itemType == 0x3D || itemType == 0x3E:
		buildScroll(&item, itemType, roll)
	}

	if index := randomItemTemplateIndex(itemType, templateIndex, roll); index != 0 {
		applyRandomItemTemplate(&item, index)
	}
	return item
}

// rollItemPlus 是 `overlay-21:0FAFh`：擲 1d20，1..14 給 ＋1、15..20 給 ＋2。
func rollItemPlus(roll RollFunc) int {
	switch value := roll(20); {
	case value >= 1 && value <= 14:
		return 1
	case value >= 15 && value <= 20:
		return 2
	}
	return 0
}

func isGeneralRandomItem(itemType uint8) bool {
	if itemType >= 1 && itemType <= 0x3B {
		return true
	}
	return itemType == 0x49 || itemType == 0x4D || itemType == 0x5D
}

// randomItemWeight 回傳 `+37h` 重量與 `+39h` 數量。
// ★ 粗體那六個（150／200／250／300／350／450）正是 AD&D 1e 的護甲重量，
// 一個不差——所以 `34h`..`3Ah` 是護甲那一段，這一欄是重量不是價格。
func randomItemWeight(itemType uint8) (weight int16, count uint8) {
	switch itemType {
	case 0x08, 0x4D:
		return 10, 0
	case 0x06:
		return 15, 0
	case 0x15:
		return 20, 0
	case 0x09:
		return 25, 5
	case 0x07:
		return 30, 0
	case 0x25:
		return 35, 0
	case 0x16, 0x1E:
		return 40, 0
	case 0x02, 0x14, 0x1D, 0x1F, 0x20, 0x21, 0x27, 0x2A, 0x2C, 0x2E, 0x3B:
		return 50, 0
	case 0x0A, 0x1A, 0x24:
		return 60, 0
	case 0x01, 0x0D, 0x0E, 0x23:
		return 75, 0
	case 0x0B, 0x10, 0x19, 0x1B, 0x29, 0x2F:
		return 80, 0
	case 0x04, 0x0F, 0x17, 0x22, 0x2B, 0x2D, 0x33:
		return 100, 0
	case 0x03, 0x18, 0x28:
		return 125, 0
	case 0x05, 0x0C, 0x11, 0x13, 0x32:
		return 150, 0
	case 0x12:
		return 175, 0
	case 0x34:
		return 200, 0
	case 0x26, 0x35:
		return 250, 0
	case 0x37:
		return 300, 0
	case 0x39:
		return 350, 0
	case 0x36, 0x38:
		return 400, 0
	case 0x3A:
		return 450, 0
	case 0x5D:
		// ⚠ 原作 `149Fh` 這一格寫的是 `2Fh` 或 `5Dh`，但 `2Fh` 在 `13DBh`（80）
		// 就被攔下了，走不到這裡。照原作行為只留 `5Dh`。
		return 1, 0
	}
	return 40, 0x0A
}

// randomItemValueCoefficient 是 `+3Ah` ＝ `plus × 係數`。
// ⚠ `plus` 可以是 0 ⇒ 價值 0（spec 1013 的「value = 0 會被靜靜改成 1」要一起看）。
func randomItemValueCoefficient(itemType uint8) int16 {
	switch itemType {
	case 0x3B:
		return 2500
	case 0x1C, 0x49:
		return 150
	case 0x35, 0x36, 0x4D:
		return 3000
	case 0x37, 0x38:
		return 3500
	case 0x39:
		return 4000
	case 0x3A:
		return 5000
	}
	return 2000
}

// buildScroll 是類別 `3Dh`／`3Eh`：擲 1..3 個法術，逐個擲環數再擲該環的法術編號，
// 法術直接存進 `special(1..3)`（`+3Ch`..`+3Eh`），價值是各環 × 300 的總和。
func buildScroll(item *ItemRecord, itemType uint8, roll RollFunc) {
	spellCount := roll(3)
	third := uint8(0xD0)
	if itemType == 0x3D {
		third = 0xD1
	}
	item.NameNumbers = [3]uint8{0, uint8(spellCount + 0xD1), third}
	item.Plus = 1
	item.Weight = 25
	item.Value = 0
	for slot := 0; slot < spellCount && slot < len(item.Affects); slot++ {
		ring := roll(5)
		item.Affects[slot] = rollScrollSpell(itemType, ring, roll)
		item.Value += int16(ring) * 300
	}
}

// rollScrollSpell：兩個類別的法術編號區間互不重疊，形狀上是兩種施法職業的卷軸。
func rollScrollSpell(itemType uint8, ring int, roll RollFunc) uint8 {
	if itemType == 0x3D {
		switch ring {
		case 1:
			return uint8(roll(13) + 8)
		case 2:
			return uint8(roll(7) + 28)
		case 3:
			return uint8(roll(11) + 44)
		case 4:
			return uint8(roll(9) + 80)
		case 5:
			return uint8(roll(4) + 90)
		}
		return 0
	}
	switch ring {
	case 1:
		return uint8(roll(8))
	case 2:
		return uint8(roll(7) + 21)
	case 3:
		return uint8(roll(8) + 36)
	case 4:
		return uint8(roll(5) + 65)
	case 5:
		return uint8(roll(6) + 70)
	}
	return 0
}

// randomItemTemplateIndex 挑特殊物品範本表 `DS:82Ch` 的索引（`overlay-21:1795h`）。
// `existing` 是一般物品那一段先決定好的（目前只有類別 `15h` 會用到）。
func randomItemTemplateIndex(itemType uint8, existing int, roll RollFunc) int {
	if existing != 0 {
		return existing
	}
	switch itemType {
	case 0x3F, 0x43:
		return 41
	case 0x4E, 0x4F:
		return 33
	case 0x54, 0x5C:
		return 9
	case 0x47:
		switch value := roll(8); {
		case value >= 1 && value <= 5:
			return 17
		case value >= 6 && value <= 8:
			return 1
		}
	}
	return 0
}

// randomItemTemplate 是 `DS:82Ch` 那張 word 陣列，一筆 8 個 word、索引從 1 起。
// 七筆逐 word 兩平台相同。⚠ 索引 25 那一筆在造物品這一支沒有被用到。
var randomItemTemplate = map[int]struct {
	NameNumbers [3]uint8
	Weight      int16
	Value       int16
	Affects     [3]uint8
}{
	1:  {[3]uint8{185, 187, 64}, 1, 800, [3]uint8{3, 99, 0}},
	9:  {[3]uint8{239, 167, 64}, 1, 1100, [3]uint8{1, 59, 0}},
	17: {[3]uint8{185, 167, 64}, 1, 400, [3]uint8{1, 3, 0}},
	25: {[3]uint8{173, 167, 64}, 1, 450, [3]uint8{1, 48, 0}},
	33: {[3]uint8{206, 167, 69}, 1, 11000, [3]uint8{30, 15, 0}},
	41: {[3]uint8{226, 167, 100}, 10, 15000, [3]uint8{0, 38, 131}},
	49: {[3]uint8{157, 167, 21}, 20, 3000, [3]uint8{1, 51, 0}},
}

// applyRandomItemTemplate 覆蓋名稱三段、重量、價值與三個效果槽，
// 並把加值與豁免加值都設成 1（`overlay-21:1804h` 起）。
func applyRandomItemTemplate(item *ItemRecord, index int) {
	entry, ok := randomItemTemplate[index]
	if !ok {
		return
	}
	item.NameNumbers = entry.NameNumbers
	item.Plus = 1
	item.PlusSave = 1
	item.Weight = entry.Weight
	item.Count = 0
	item.Value = entry.Value
	item.Affects = entry.Affects
}
