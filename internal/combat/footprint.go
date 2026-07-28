package combat

// Footprint maps the reference CombatMap.size shape code to occupied cells.
// field_DE&7 values observed in CoAB are: 1 normal, 2 vertical, 3 horizontal,
// and 4 large (2x2).
type Footprint struct {
	Width  int
	Height int
}

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

func FootprintsOverlapAt(first Fighter, firstX, firstY int, second Fighter) bool {
	a := FootprintForSize(first.CombatSize)
	b := FootprintForSize(second.CombatSize)
	return firstX < second.CombatX+b.Width &&
		firstX+a.Width > second.CombatX &&
		firstY < second.CombatY+b.Height &&
		firstY+a.Height > second.CombatY
}

func footprintAdjacent(first, second Fighter) bool {
	a := FootprintForSize(first.CombatSize)
	b := FootprintForSize(second.CombatSize)
	left := max(first.CombatX, second.CombatX)
	right := min(first.CombatX+a.Width, second.CombatX+b.Width)
	top := max(first.CombatY, second.CombatY)
	bottom := min(first.CombatY+a.Height, second.CombatY+b.Height)
	if left < right && top < bottom {
		return false
	}
	dx := max(0, max(first.CombatX, second.CombatX)-min(first.CombatX+a.Width, second.CombatX+b.Width))
	dy := max(0, max(first.CombatY, second.CombatY)-min(first.CombatY+a.Height, second.CombatY+b.Height))
	return dx == 0 && dy == 0
}
