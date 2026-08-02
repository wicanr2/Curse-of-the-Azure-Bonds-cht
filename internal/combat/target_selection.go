package combat

import (
	"fmt"
	"sort"

	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
	enginescanorder "github.com/wicanr2/golden-box-remake-engine/combat/scanorder"
)

// TargetVisibility supplies title-neutral status visibility. Terrain
// visibility is handled separately by TargetSelectionOptions.Terrain.
type TargetVisibility func(attacker, target Fighter) bool

// TargetSelectionOptions describes the shared Gold Box ranged-target
// boundary. MaxRange uses the original displayed range unit; the internal
// grid uses two weighted units for a cardinal step and three for a diagonal.
type TargetSelectionOptions struct {
	MaxRange  int
	Terrain   LineTerrain
	VisibleTo TargetVisibility
}

type rangedTargetCandidate struct {
	fighter  Fighter
	distance int
}

// ScanTargetRecord binds one exact legacy SCAN record to the remake fighter
// identity. ObjectID is the original combat-object index, not a derived map or
// slice position. Candidate discovery remains the caller's responsibility.
type ScanTargetRecord struct {
	ObjectID  uint8
	TargetID  string
	Distance  uint8
	Direction uint8
}

// OrderScanTargetIDs validates the title projection and applies the recovered
// PC-98 three-byte record order. It does not discover candidates or infer LOS,
// footprint, terrain, side or spell suitability.
func OrderScanTargetIDs(records []ScanTargetRecord) ([]string, error) {
	entries := make([]enginescanorder.Entry, len(records))
	targetByObject := make(map[uint8]string, len(records))
	seenTarget := make(map[string]struct{}, len(records))
	for index, record := range records {
		if record.ObjectID == 0 {
			return nil, fmt.Errorf("SCAN object ID zero is outside the original one-based list")
		}
		if record.TargetID == "" {
			return nil, fmt.Errorf("SCAN object %d has an empty target ID", record.ObjectID)
		}
		if _, duplicate := targetByObject[record.ObjectID]; duplicate {
			return nil, fmt.Errorf("duplicate SCAN object ID %d", record.ObjectID)
		}
		if _, duplicate := seenTarget[record.TargetID]; duplicate {
			return nil, fmt.Errorf("duplicate SCAN target ID %q", record.TargetID)
		}
		targetByObject[record.ObjectID] = record.TargetID
		seenTarget[record.TargetID] = struct{}{}
		entries[index] = enginescanorder.Entry{
			ObjectID: record.ObjectID, Distance: record.Distance, Direction: record.Direction,
		}
	}
	enginescanorder.Sort(entries)
	ordered := make([]string, len(entries))
	for index, entry := range entries {
		ordered[index] = targetByObject[entry.ObjectID]
	}
	return ordered, nil
}

// ScanTargetCandidate binds one title-owned legacy object identity and exact
// footprint expansion to the remake fighter identity. Callers must derive the
// object ID from the original COMPOBJ builder; slice position is not evidence.
type ScanTargetCandidate struct {
	ObjectID uint8
	TargetID string
	Cells    []enginescan.Point
}

// LegacyScanObjects projects the original OBJECTLIST/IDLIST identity table
// from Battle's preserved CHARACTERLIST order. It excludes inactive records
// exactly at the title boundary represented by dead or unplaced fighters.
func (b *Battle) LegacyScanObjects(sourceID string, targetSide Side) (
	uint8, []enginescan.Point, []ScanTargetCandidate, error,
) {
	if b == nil {
		return 0, nil, nil, fmt.Errorf("battle is nil")
	}
	source, ok := b.fighters[sourceID]
	if !ok {
		return 0, nil, nil, fmt.Errorf("unknown SCAN source %q", sourceID)
	}
	if source.LegacyObjectID == 0 || !source.HasCombatPosition || source.HitPoints <= 0 {
		return 0, nil, nil, fmt.Errorf("SCAN source %q has no active legacy object", sourceID)
	}
	sourceCells := scanFootprintCells(source)
	candidates := make([]ScanTargetCandidate, 0, len(b.fighterOrder))
	seen := make(map[uint8]struct{}, len(b.fighterOrder))
	for _, fighterID := range b.fighterOrder {
		fighter := b.fighters[fighterID]
		if fighter.LegacyObjectID == 0 {
			return 0, nil, nil, fmt.Errorf("fighter %q has no legacy combat object ID", fighter.ID)
		}
		if fighter.LegacyObjectID > 72 {
			return 0, nil, nil, fmt.Errorf("fighter %q object ID %d exceeds OBJECTLIST capacity", fighter.ID, fighter.LegacyObjectID)
		}
		if _, duplicate := seen[fighter.LegacyObjectID]; duplicate {
			return 0, nil, nil, fmt.Errorf("duplicate legacy combat object ID %d", fighter.LegacyObjectID)
		}
		seen[fighter.LegacyObjectID] = struct{}{}
		if fighter.Side != targetSide || fighter.HitPoints <= 0 || !fighter.HasCombatPosition {
			continue
		}
		candidates = append(candidates, ScanTargetCandidate{
			ObjectID: fighter.LegacyObjectID,
			TargetID: fighter.ID,
			Cells:    scanFootprintCells(fighter),
		})
	}
	return source.LegacyObjectID, sourceCells, candidates, nil
}

func scanFootprintCells(fighter Fighter) []enginescan.Point {
	footprint := FootprintForSize(fighter.CombatSize)
	cells := make([]enginescan.Point, 0, footprint.Width*footprint.Height)
	for y := fighter.CombatY; y < fighter.CombatY+footprint.Height; y++ {
		for x := fighter.CombatX; x < fighter.CombatX+footprint.Width; x++ {
			cells = append(cells, enginescan.Point{X: x, Y: y})
		}
	}
	return cells
}

// BuildScanTargetIDs runs the recovered terrain-aware producer and returns
// stable fighter IDs in original SCAN order. Arc is the original INARC sector:
// 0..7 select one direction and 8/FFh accept every direction.
func BuildScanTargetIDs(
	tacticalMap enginescan.TacticalMap,
	sourceID uint8,
	sourceCells []enginescan.Point,
	candidates []ScanTargetCandidate,
	maxRange int,
	arc uint8,
) ([]string, error) {
	objects := make([]enginescan.Object, len(candidates))
	targetByObject := make(map[uint8]string, len(candidates))
	seenTarget := make(map[string]struct{}, len(candidates))
	for index, candidate := range candidates {
		if candidate.TargetID == "" {
			return nil, fmt.Errorf("SCAN object %d has an empty target ID", candidate.ObjectID)
		}
		if _, duplicate := targetByObject[candidate.ObjectID]; duplicate {
			return nil, fmt.Errorf("duplicate SCAN object ID %d", candidate.ObjectID)
		}
		if _, duplicate := seenTarget[candidate.TargetID]; duplicate {
			return nil, fmt.Errorf("duplicate SCAN target ID %q", candidate.TargetID)
		}
		targetByObject[candidate.ObjectID] = candidate.TargetID
		seenTarget[candidate.TargetID] = struct{}{}
		objects[index] = enginescan.Object{ID: candidate.ObjectID, Cells: candidate.Cells}
	}
	entries, err := tacticalMap.Build(
		enginescan.Object{ID: sourceID, Cells: sourceCells}, objects, maxRange, arc,
	)
	if err != nil {
		return nil, err
	}
	ordered := make([]string, len(entries))
	for index, entry := range entries {
		ordered[index] = targetByObject[entry.ObjectID]
	}
	return ordered, nil
}

// BuildLegacyScanTargetIDs closes the title adapter transaction from the
// original CHARACTERLIST/IDLIST identity projection through SCAN ordering to
// stable remake fighter IDs.
func (b *Battle) BuildLegacyScanTargetIDs(
	tacticalMap enginescan.TacticalMap,
	sourceID string,
	targetSide Side,
	maxRange int,
	arc uint8,
) ([]string, error) {
	objectID, sourceCells, candidates, err := b.LegacyScanObjects(sourceID, targetSide)
	if err != nil {
		return nil, err
	}
	return BuildScanTargetIDs(tacticalMap, objectID, sourceCells, candidates, maxRange, arc)
}

// BuildLegacyAreaScanTargetIDs preserves the combat-object identity table but
// uses an explicitly selected map cell as SCAN's source. PC-98 combat spell
// targeting writes the selected X/Y separately from the caster pointer before
// calling SCAN; using the caster footprint here would silently change Sleep's
// area of effect.
func (b *Battle) BuildLegacyAreaScanTargetIDs(
	tacticalMap enginescan.TacticalMap,
	identitySourceID string,
	center enginescan.Point,
	targetSide Side,
	maxRange int,
	arc uint8,
) ([]string, error) {
	objectID, _, candidates, err := b.LegacyScanObjects(identitySourceID, targetSide)
	if err != nil {
		return nil, err
	}
	return BuildScanTargetIDs(
		tacticalMap, objectID, []enginescan.Point{center}, candidates, maxRange, arc,
	)
}

// SelectRangedCombatTarget builds the reachable opposing list before using
// the battle PRNG. It preserves the reference 20-attempt, remove-invisible-
// candidate loop. found=false is a normal result: some original callers use
// their own no-target continuation rather than converting it into an error.
func (b *Battle) SelectRangedCombatTarget(attackerID string, targetSide Side, options TargetSelectionOptions) (target Fighter, found bool, err error) {
	if b == nil {
		return Fighter{}, false, fmt.Errorf("battle is nil")
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return Fighter{}, false, fmt.Errorf("unknown attacker %q", attackerID)
	}
	if b.status != StatusActive {
		return Fighter{}, false, fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 || !attacker.HasCombatPosition {
		return Fighter{}, false, fmt.Errorf("dead or unplaced fighter cannot select a ranged target")
	}
	if options.MaxRange < 0 {
		return Fighter{}, false, fmt.Errorf("target range cannot be negative")
	}
	if options.Terrain == nil {
		return Fighter{}, false, fmt.Errorf("target terrain is required")
	}

	candidates := make([]rangedTargetCandidate, 0, len(b.fighters))
	for _, fighter := range b.fighters {
		if fighter.Side != targetSide || fighter.HitPoints <= 0 || !fighter.HasCombatPosition {
			continue
		}
		distance, reachable := rangedTargetDistance(attacker, fighter, options.MaxRange, options.Terrain)
		if reachable {
			candidates = append(candidates, rangedTargetCandidate{fighter: fighter, distance: distance})
		}
	}
	// The reference list is distance-sorted before random selection. ID is the
	// stable remake tie-break until the original combatant-array tie order is
	// represented explicitly by the battle adapter.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].fighter.ID < candidates[j].fighter.ID
	})

	for attempts := 0; attempts < 20 && len(candidates) > 0; attempts++ {
		index := b.rng.Intn(len(candidates))
		candidate := candidates[index]
		if options.VisibleTo == nil || options.VisibleTo(attacker, candidate.fighter) {
			return candidate.fighter, true, nil
		}
		candidates = append(candidates[:index], candidates[index+1:]...)
	}
	return Fighter{}, false, nil
}

func rangedTargetDistance(attacker, target Fighter, maxRange int, terrain LineTerrain) (int, bool) {
	attackerFootprint := FootprintForSize(attacker.CombatSize)
	targetFootprint := FootprintForSize(target.CombatSize)
	maximumWeightedDistance := maxRange*2 + 1
	best := maximumWeightedDistance + 1
	for attackerY := attacker.CombatY; attackerY < attacker.CombatY+attackerFootprint.Height; attackerY++ {
		for attackerX := attacker.CombatX; attackerX < attacker.CombatX+attackerFootprint.Width; attackerX++ {
			from := TilePoint{X: attackerX, Y: attackerY}
			for targetY := target.CombatY; targetY < target.CombatY+targetFootprint.Height; targetY++ {
				for targetX := target.CombatX; targetX < target.CombatX+targetFootprint.Width; targetX++ {
					to := TilePoint{X: targetX, Y: targetY}
					distance := weightedTileDistance(from, to)
					if distance >= best || distance > maximumWeightedDistance || !unobstructedTargetRay(from, to, terrain) {
						continue
					}
					best = distance
				}
			}
		}
	}
	return best, best <= maximumWeightedDistance
}

func unobstructedTargetRay(from, to TilePoint, terrain LineTerrain) bool {
	if from == to {
		return true
	}
	stepper := newRayStepper(from, to.X-from.X, to.Y-from.Y, 1, 1)
	for {
		next, _, stepped := stepper.step()
		if !stepped {
			return false
		}
		cell := terrain(next.X, next.Y)
		if !cell.Valid || cell.Reflect {
			return false
		}
		if next == to {
			return true
		}
	}
}
