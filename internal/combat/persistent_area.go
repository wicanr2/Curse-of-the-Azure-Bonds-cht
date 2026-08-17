package combat

import (
	"fmt"
	"sort"
)

type PersistentAreaKind uint8

const (
	PersistentAreaStinkingCloud PersistentAreaKind = iota + 1
	PersistentAreaCloudkill
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
	SaveRequired  bool
	Killed        bool
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

var cloudkillOffsets = [...]TilePoint{
	{X: 0, Y: 0},
	{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0}, {X: 1, Y: 1},
	{X: 0, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 0}, {X: -1, Y: -1},
}

// CastCloudkill creates the reference target-centred 3x3 poisonous cloud.
// Hit Dice 0-4 die automatically, 5 save versus Poison at -4, 6 save
// unmodified, and 7+ are unaffected.
func (b *Battle) CastCloudkill(casterID string, center TilePoint, level int, terrain AreaTerrain) (PersistentAreaResult, error) {
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
	cells := make([]PersistentAreaCell, 0, len(cloudkillOffsets))
	for _, offset := range cloudkillOffsets {
		cell := PersistentAreaCell{X: center.X + offset.X, Y: center.Y + offset.Y}
		if terrain == nil || terrain(cell.X, cell.Y) {
			cells = append(cells, cell)
		}
	}
	if len(cells) == 0 {
		return PersistentAreaResult{}, fmt.Errorf("cloudkill has no passable cells")
	}
	b.nextArea++
	area := PersistentArea{
		ID: b.nextArea, Kind: PersistentAreaCloudkill, CasterID: casterID,
		Center: center, CreatedRound: b.round, ExpiresRound: b.round + level,
		Cells: append([]PersistentAreaCell(nil), cells...),
	}
	b.areas = append(b.areas, area)

	targets := make([]Fighter, 0)
	for _, fighter := range b.fighters {
		if fighter.HitPoints > 0 && fighter.HasCombatPosition &&
			fighter.HitDice < 7 && areaIntersectsFighter(cells, fighter) {
			if fighter.HitDice >= 5 && len(fighter.SavingThrows) == 0 {
				b.areas = b.areas[:len(b.areas)-1]
				return PersistentAreaResult{}, fmt.Errorf("fighter %q has no poison saving throw", fighter.ID)
			}
			targets = append(targets, fighter)
		}
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].ID < targets[j].ID })
	result := PersistentAreaResult{Area: area, Impacts: make([]PersistentAreaImpact, 0, len(targets))}
	for _, target := range targets {
		impact := PersistentAreaImpact{TargetID: target.ID}
		switch target.HitDice {
		case 0, 1, 2, 3, 4:
			impact.Killed = true
		case 5, 6:
			modifier := 0
			if target.HitDice == 5 {
				modifier = -4
			}
			impact.SaveRequired = true
			roll := b.rng.Intn(20) + 1
			conditional := target.MonsterConditionalModifierAgainst(caster)
			impact.Saved = roll == 20 || roll != 1 &&
				roll+modifier+target.SavingThrowBonus+conditional.SavingThrowDelta >= int(target.SavingThrows[0])
			impact.Killed = !impact.Saved
		}
		if impact.Killed {
			// PC-98 Cloudkill applies raw effect 44h before the direct-death
			// handoff. Its combat consumer independently cancels a pending
			// spell and consumes the matching memorized slot; it does not use
			// PUTDAMAGE's positive-damage branch.
			b.interruptPendingSpell(&target)
			b.fighters[target.ID] = target
			if err := b.SetHitPoints(target.ID, 0); err != nil {
				return PersistentAreaResult{}, err
			}
		}
		result.Impacts = append(result.Impacts, impact)
	}
	return result, nil
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
		conditional := target.MonsterConditionalModifierAgainst(caster)
		saved := roll == 20 || roll != 1 &&
			roll+target.SavingThrowBonus+conditional.SavingThrowDelta >= int(target.SavingThrows[0])
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

func (b *Battle) cloudkillIntersectsAt(fighter Fighter, x, y int) bool {
	fighter.CombatX, fighter.CombatY = x, y
	for _, area := range b.areas {
		if area.Kind == PersistentAreaCloudkill && areaIntersectsFighter(area.Cells, fighter) {
			return true
		}
	}
	return false
}

// 兩種雲寫上去的地形碼（spec 1128）。
//
// ★★ 這兩個碼就是 spec 1119 那兩種「障礙格」的來源。原作的雲霧 handler 在
// 鋪完格子之後把地形碼寫進地圖格的 `+7`：
//
//	惡臭之雲（`ov22@2B21h`）  地形碼 `1Eh` ⇒ 豁免過得了就照走
//	致命毒雲（`ov22@53FDh`／`5464h`）地形碼 `1Ch` ⇒ 等級夠高（`≥ 7`）就照走
//
// 先前 `ObstacleTerrainBlocks` 有規則卻**沒有任何生產者**——地形碼查詢從來沒被
// 掛上。兩種雲就是那個缺掉的生產者。
func (kind PersistentAreaKind) ObstacleTerrainCode() (uint8, bool) {
	switch kind {
	case PersistentAreaStinkingCloud:
		return ObstacleTerrainSaveable, true
	case PersistentAreaCloudkill:
		return ObstacleTerrainVeteran, true
	default:
		return 0, false
	}
}

// PersistentAreaTerrainCode 回答「這一格現在是不是障礙格」。
//
// 兩種雲疊在同一格時取**先鋪的那一個**：原作把地形碼直接寫進地圖格，後寫的會
// 蓋掉先寫的，但兩份物件清單是分開的（`ds:755Bh`／`ds:755Fh`），本函式照
// 建立順序回答，不猜覆蓋規則。
func (b *Battle) PersistentAreaTerrainCode(x, y int) (uint8, bool) {
	if b == nil {
		return 0, false
	}
	for _, area := range b.areas {
		code, ok := area.Kind.ObstacleTerrainCode()
		if !ok {
			continue
		}
		for _, cell := range area.Cells {
			if cell.X == x && cell.Y == y {
				return code, true
			}
		}
	}
	return 0, false
}
