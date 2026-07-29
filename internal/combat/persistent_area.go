package combat

import (
	"fmt"
	"sort"
)

type PersistentAreaKind uint8

const (
	PersistentAreaStinkingCloud PersistentAreaKind = iota + 1
)

// AreaTerrain keeps title-specific background tables outside the combat core.
// true means an effect may occupy the cell.
type AreaTerrain func(x, y int) bool

type PersistentAreaCell struct {
	X int
	Y int
}

type PersistentArea struct {
	ID           uint64
	Kind         PersistentAreaKind
	CasterID     string
	Center       TilePoint
	CreatedRound int
	ExpiresRound int
	Cells        []PersistentAreaCell
}

type PersistentAreaImpact struct {
	TargetID      string
	Saved         bool
	CoughingTurns int
	HelplessTurns int
}

type PersistentAreaResult struct {
	Area    PersistentArea
	Impacts []PersistentAreaImpact
}

var stinkingCloudOffsets = [...]TilePoint{
	{X: 0, Y: 0},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
	{X: 0, Y: 1},
}

// CastStinkingCloud creates the reference target-anchored 2x2 noxious cloud.
// Every intersecting combatant saves versus Poison once, even if a large
// footprint touches more than one cloud cell.
func (b *Battle) CastStinkingCloud(casterID string, center TilePoint, level int, terrain AreaTerrain) (PersistentAreaResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return PersistentAreaResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return PersistentAreaResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return PersistentAreaResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if level < 1 {
		return PersistentAreaResult{}, fmt.Errorf("caster level must be positive")
	}
	cells := make([]PersistentAreaCell, 0, len(stinkingCloudOffsets))
	for _, offset := range stinkingCloudOffsets {
		cell := PersistentAreaCell{X: center.X + offset.X, Y: center.Y + offset.Y}
		if terrain != nil && !terrain(cell.X, cell.Y) {
			continue
		}
		cells = append(cells, cell)
	}
	if len(cells) == 0 {
		return PersistentAreaResult{}, fmt.Errorf("stinking cloud has no passable cells")
	}
	b.nextArea++
	area := PersistentArea{
		ID: b.nextArea, Kind: PersistentAreaStinkingCloud, CasterID: casterID,
		Center: center, CreatedRound: b.round, ExpiresRound: b.round + level,
		Cells: append([]PersistentAreaCell(nil), cells...),
	}
	b.areas = append(b.areas, area)

	targets := make([]Fighter, 0)
	for _, fighter := range b.fighters {
		if fighter.HitPoints <= 0 || !fighter.HasCombatPosition || !areaIntersectsFighter(cells, fighter) {
			continue
		}
		if len(fighter.SavingThrows) == 0 {
			b.areas = b.areas[:len(b.areas)-1]
			return PersistentAreaResult{}, fmt.Errorf("fighter %q has no poison saving throw", fighter.ID)
		}
		targets = append(targets, fighter)
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].ID < targets[j].ID })
	result := PersistentAreaResult{Area: area, Impacts: make([]PersistentAreaImpact, 0, len(targets))}
	for _, target := range targets {
		roll := b.rng.Intn(20) + 1
		saved := roll == 20 || roll != 1 &&
			roll+target.SavingThrowBonus >= int(target.SavingThrows[0])
		impact := PersistentAreaImpact{TargetID: target.ID, Saved: saved}
		if saved {
			target.CoughingTurns = max(target.CoughingTurns, 1)
			impact.CoughingTurns = 1
		} else {
			duration := b.rng.Intn(4) + 2
			target.HelplessTurns = max(target.HelplessTurns, duration)
			impact.HelplessTurns = duration
		}
		b.fighters[target.ID] = target
		result.Impacts = append(result.Impacts, impact)
	}
	return result, nil
}

func (b *Battle) PersistentAreas() []PersistentArea {
	output := make([]PersistentArea, len(b.areas))
	for index, area := range b.areas {
		output[index] = area
		output[index].Cells = append([]PersistentAreaCell(nil), area.Cells...)
	}
	return output
}

func (b *Battle) PersistentAreaAt(x, y int) (PersistentAreaKind, bool) {
	for index := len(b.areas) - 1; index >= 0; index-- {
		for _, cell := range b.areas[index].Cells {
			if cell.X == x && cell.Y == y {
				return b.areas[index].Kind, true
			}
		}
	}
	return 0, false
}

func (f Fighter) CloudIncapacitated() bool {
	return f.CoughingTurns > 0 || f.HelplessTurns > 0
}

// ConsumeCloudIncapacitation advances action-counted nausea only when the
// fighter's initiative is actually skipped.
func (b *Battle) ConsumeCloudIncapacitation(fighterID string) (Fighter, error) {
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return Fighter{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	if fighter.CoughingTurns > 0 {
		fighter.CoughingTurns--
	}
	if fighter.HelplessTurns > 0 {
		fighter.HelplessTurns--
	}
	b.fighters[fighterID] = fighter
	return fighter, nil
}

func (b *Battle) expirePersistentAreas() {
	if len(b.areas) == 0 {
		return
	}
	kept := b.areas[:0]
	for _, area := range b.areas {
		if area.ExpiresRound > b.round {
			kept = append(kept, area)
		}
	}
	b.areas = kept
}

func areaIntersectsFighter(cells []PersistentAreaCell, fighter Fighter) bool {
	footprint := FootprintForSize(fighter.CombatSize)
	for _, cell := range cells {
		if cell.X >= fighter.CombatX && cell.X < fighter.CombatX+footprint.Width &&
			cell.Y >= fighter.CombatY && cell.Y < fighter.CombatY+footprint.Height {
			return true
		}
	}
	return false
}
