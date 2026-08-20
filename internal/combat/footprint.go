package combat

import enginefootprint "github.com/wicanr2/golden-box-remake-engine/combat/footprint"

// Footprint 是體型碼展開後佔的格數。矩形的重疊與相鄰在共用 engine 的
// `combat/footprint`；**體型碼對應到哪個形狀是作品資料**，留在這裡。
//
// CoAB 觀察到的 `field_DE & 7`：1 一般、2 直立、3 橫躺、4 大型（2×2）。
type Footprint = enginefootprint.Shape

func FootprintForSize(size uint8) Footprint {
	switch size & 7 {
	case 2:
		return Footprint{Width: 1, Height: 2}
	case 3:
		return Footprint{Width: 2, Height: 1}
	case 4:
		return Footprint{Width: 2, Height: 2}
	default:
		return Footprint{Width: 1, Height: 1}
	}
}

// footprintBoxAt 把戰鬥員放到指定格，回傳它佔的矩形。
func footprintBoxAt(fighter Fighter, x, y int) enginefootprint.Box {
	return enginefootprint.NewBox(x, y, FootprintForSize(fighter.CombatSize))
}

func FootprintsOverlapAt(first Fighter, firstX, firstY int, second Fighter) bool {
	return enginefootprint.Overlaps(footprintBoxAt(first, firstX, firstY),
		footprintBoxAt(second, second.CombatX, second.CombatY))
}

func footprintAdjacent(first, second Fighter) bool {
	return enginefootprint.Adjacent(footprintBoxAt(first, first.CombatX, first.CombatY),
		footprintBoxAt(second, second.CombatX, second.CombatY))
}
