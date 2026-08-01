package combat

import "fmt"

// LineCell is the title-neutral terrain projection used by reflecting line
// effects. Invalid cells terminate a path; Reflect cells consume their step
// and reverse the ray.
type LineCell struct {
	Valid   bool
	Reflect bool
}

type LineTerrain func(x, y int) LineCell

type ReflectingLineOptions struct {
	WeightedBudget                 int
	FirstReflectionOriginThreshold int
	FirstReflectionPenalty         int
	DamageFlags                    uint8
	// InitialDamageDice and PathDamageDice opt into two independently rolled
	// damage pools. Zero preserves the ordinary shared level-d6 spell rule.
	InitialDamageDice int
	PathDamageDice    int
	DamageDiceSides   int
}

type LineSpellSegment struct {
	From        TilePoint
	To          TilePoint
	HasImpact   bool
	ImpactIndex int
	Reflected   bool
}

type LineSpellResult struct {
	CasterID      string
	SpellID       uint8
	Target        TilePoint
	BaseDamage    int
	InitialDamage int
	PathDamage    int
	TravelImpacts int
	Impacts       []LineSpellImpact
	Segments      []LineSpellSegment
}

type LineSpellImpact struct {
	AreaSpellImpact
	Point TilePoint
}

// CastReflectingLineSpell resolves a shared level-d6 damage roll along a
// weighted grid ray. Cardinal steps cost two and diagonal steps cost three.
// A fighter is hit once while the ray remains inside its contiguous footprint,
// but may be hit again after the ray leaves and later re-enters it.
func (b *Battle) CastReflectingLineSpell(casterID string, spellID uint8, target TilePoint, level int, options ReflectingLineOptions, terrain LineTerrain) (LineSpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return LineSpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return LineSpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || !caster.HasCombatPosition {
		return LineSpellResult{}, fmt.Errorf("dead or unplaced fighter cannot cast")
	}
	if level < 1 {
		return LineSpellResult{}, fmt.Errorf("caster level must be positive")
	}
	if options.WeightedBudget < 1 {
		return LineSpellResult{}, fmt.Errorf("line spell budget must be positive")
	}
	origin := TilePoint{X: caster.CombatX, Y: caster.CombatY}
	if target == origin {
		return LineSpellResult{}, fmt.Errorf("line spell target must differ from caster position")
	}
	if terrain == nil {
		return LineSpellResult{}, fmt.Errorf("line spell terrain is required")
	}
	if !terrain(target.X, target.Y).Valid {
		return LineSpellResult{}, fmt.Errorf("line spell target (%d,%d) is invalid", target.X, target.Y)
	}

	initialDamage, pathDamage := 0, 0
	separateDamage := options.InitialDamageDice > 0 || options.PathDamageDice > 0
	if separateDamage {
		if options.InitialDamageDice < 1 || options.PathDamageDice < 1 || options.DamageDiceSides < 1 {
			return LineSpellResult{}, fmt.Errorf("separate line damage requires positive initial, path, and side dice")
		}
		initialDamage = b.rollDamage(options.InitialDamageDice, options.DamageDiceSides)
	} else {
		initialDamage = b.rollDamage(level, 6)
		pathDamage = initialDamage
	}
	result := LineSpellResult{
		CasterID: casterID, SpellID: spellID, Target: target, BaseDamage: initialDamage,
		InitialDamage: initialDamage,
		Impacts:       make([]LineSpellImpact, 0), Segments: make([]LineSpellSegment, 0),
	}

	lastFighterID := ""
	if fighter, found := b.livingFighterAt(target); found {
		impact, err := b.applyLineSpellDamage(fighter, initialDamage, options.DamageFlags)
		if err != nil {
			return LineSpellResult{}, err
		}
		result.Impacts = append(result.Impacts, LineSpellImpact{AreaSpellImpact: impact, Point: target})
		result.TravelImpacts = 1
		lastFighterID = fighter.ID
	}
	if separateDamage {
		pathDamage = b.rollDamage(options.PathDamageDice, options.DamageDiceSides)
	}
	result.PathDamage = pathDamage

	dx, dy := target.X-origin.X, target.Y-origin.Y
	direction := 1
	remaining := options.WeightedBudget
	segmentStart := target
	current := target
	lastPoint := target
	stepper := newRayStepper(target, dx, dy, direction, options.WeightedBudget)
	firstReflection := true
	for remaining > 0 {
		next, stepCost, stepped := stepper.step()
		if !stepped || stepCost > remaining {
			break
		}
		cell := terrain(next.X, next.Y)
		if !cell.Valid {
			break
		}
		remaining -= stepCost
		current, lastPoint = next, next

		if cell.Reflect {
			if current != segmentStart {
				result.Segments = append(result.Segments, LineSpellSegment{
					From: segmentStart, To: current, Reflected: true,
				})
			}
			segmentStart = current
			lastFighterID = ""
			if firstReflection && options.FirstReflectionOriginThreshold > 0 &&
				weightedTileDistance(current, origin) <= options.FirstReflectionOriginThreshold {
				remaining -= options.FirstReflectionPenalty
				if remaining < 0 {
					remaining = 0
				}
			}
			firstReflection = false
			direction = -direction
			stepper = newRayStepper(current, dx, dy, direction, remaining)
			continue
		}

		fighter, found := b.livingFighterAt(current)
		if !found {
			lastFighterID = ""
			continue
		}
		if fighter.ID == lastFighterID {
			continue
		}
		impact, err := b.applyLineSpellDamage(fighter, pathDamage, options.DamageFlags)
		if err != nil {
			return LineSpellResult{}, err
		}
		impactIndex := len(result.Impacts)
		result.Impacts = append(result.Impacts, LineSpellImpact{AreaSpellImpact: impact, Point: current})
		result.Segments = append(result.Segments, LineSpellSegment{
			From: segmentStart, To: current, HasImpact: true, ImpactIndex: impactIndex,
		})
		segmentStart = current
		lastFighterID = fighter.ID
	}
	if lastPoint != segmentStart {
		result.Segments = append(result.Segments, LineSpellSegment{From: segmentStart, To: lastPoint})
	}
	b.updateStatus()
	return result, nil
}

func (b *Battle) rollDamage(count, sides int) int {
	damage := 0
	for roll := 0; roll < count; roll++ {
		damage += b.rng.Intn(sides) + 1
	}
	return damage
}

func weightedTileDistance(left, right TilePoint) int {
	dx, dy := abs(left.X-right.X), abs(left.Y-right.Y)
	return 2*max(dx, dy) + min(dx, dy)
}

func (b *Battle) livingFighterAt(point TilePoint) (Fighter, bool) {
	for _, fighter := range b.Fighters() {
		if fighter.HitPoints <= 0 || !fighter.HasCombatPosition {
			continue
		}
		footprint := FootprintForSize(fighter.CombatSize)
		if point.X >= fighter.CombatX && point.X < fighter.CombatX+footprint.Width &&
			point.Y >= fighter.CombatY && point.Y < fighter.CombatY+footprint.Height {
			return fighter, true
		}
	}
	return Fighter{}, false
}

func (b *Battle) applyLineSpellDamage(target Fighter, damage int, damageFlags uint8) (AreaSpellImpact, error) {
	if len(target.SavingThrows) <= 4 {
		return AreaSpellImpact{}, fmt.Errorf("fighter %q has no spell saving throw", target.ID)
	}
	saveRoll := b.rng.Intn(20) + 1
	saved := saveRoll == 20 ||
		saveRoll != 1 && saveRoll+target.SavingThrowBonus >= int(target.SavingThrows[4])
	applied := damage
	if saved {
		applied /= 2
	}
	protected := target.MonsterProtectedFromDamage(damageFlags)
	if protected {
		applied = 0
	}
	if applied > target.HitPoints {
		applied = target.HitPoints
	}
	target.HitPoints -= applied
	b.fighters[target.ID] = target
	return AreaSpellImpact{
		TargetID: target.ID, Damage: applied, TargetHP: target.HitPoints, Saved: saved,
		Protected: protected,
	}, nil
}

type rayStepper struct {
	current             TilePoint
	target              TilePoint
	diffX, diffY        int
	signX, signY, delta int
}

func newRayStepper(origin TilePoint, dx, dy, direction, budget int) rayStepper {
	if budget < 1 {
		budget = 1
	}
	target := TilePoint{
		X: origin.X + dx*direction*budget,
		Y: origin.Y + dy*direction*budget,
	}
	diffX, diffY := abs(target.X-origin.X), abs(target.Y-origin.Y)
	return rayStepper{
		current: origin, target: target, diffX: diffX, diffY: diffY,
		signX: sign(target.X - origin.X), signY: sign(target.Y - origin.Y),
	}
}

func (stepper *rayStepper) step() (TilePoint, int, bool) {
	if stepper.diffX >= stepper.diffY {
		if stepper.current.X == stepper.target.X {
			return stepper.current, 0, false
		}
		stepper.current.X += stepper.signX
		stepper.delta += stepper.diffY * 2
		cost := 2
		if stepper.delta >= stepper.diffX {
			stepper.current.Y += stepper.signY
			stepper.delta -= stepper.diffX * 2
			cost++
		}
		return stepper.current, cost, true
	}
	if stepper.current.Y == stepper.target.Y {
		return stepper.current, 0, false
	}
	stepper.current.Y += stepper.signY
	stepper.delta += stepper.diffX * 2
	cost := 2
	if stepper.delta >= stepper.diffY {
		stepper.current.X += stepper.signX
		stepper.delta -= stepper.diffY * 2
		cost++
	}
	return stepper.current, cost, true
}

func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}
