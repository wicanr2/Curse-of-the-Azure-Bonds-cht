// Package combat contains the platform-neutral AD&D combat core. ECL and
// Ebiten adapters can provide party/enemy data without embedding rendering or
// DOS memory assumptions here.
package combat

import (
	"fmt"
	"math/rand"
	"sort"
)

type Side uint8

const (
	SideParty Side = iota
	SideEnemy
)

type Status uint8

const (
	StatusActive Status = iota
	StatusPartyWon
	StatusEnemyWon
	StatusDraw
)

type Fighter struct {
	ID   string
	Name string
	Side Side
	Evil bool
	Good bool
	// SpriteSet/SpriteBlock identify the original CPIC asset when the fighter
	// came from a LOAD MONSTER descriptor. A zero block means the renderer may
	// choose a deterministic party fallback sprite.
	SpriteSet   uint8
	SpriteBlock uint8
	// AnimationBlock is the SPRIT block from SETUP MONSTER. It is separate
	// from SpriteBlock because the original engine loads CPIC and SPRIT from
	// different DAX families.
	AnimationBlock uint8
	HasAnimation   bool
	// Party icon fields mirror the original Player combat record. They are
	// renderer-neutral so save/import decoding can fill the actual values.
	HasPartyIcon   bool
	PartyHeadBlock uint8
	PartyBodyBlock uint8
	PartyIconID    uint8
	PartyIconSize  uint8
	IconDirection  uint8
	IconAttack     bool
	// CombatMap position/size. A future Area/ECL placement decoder can set
	// these directly; StartCombat supplies a deterministic fallback otherwise.
	HasCombatPosition    bool
	CombatX              int
	CombatY              int
	CombatSize           uint8
	HitPoints            int
	MaxHitPoints         int
	ArmorClass           int
	AttackBonus          int
	Blessed              bool
	BlessRounds          int
	Cursed               bool
	CurseRounds          int
	ProtectedFromEvil    bool
	ProtectionEvilRounds int
	ProtectedFromGood    bool
	ProtectionGoodRounds int
	DamageDiceCount      int
	DamageDiceSides      int
	DamageBonus          int
	AttacksPerTurn       int
	AmmunitionType       uint8
	MovementAllowance    int
	WeaponRange          int
	MissileWeapon        bool
	ThrownWeapon         bool
	InitiativeBonus      int
}

type Turn struct {
	FighterID  string
	Initiative int
}

type AttackResult struct {
	AttackerID string
	TargetID   string
	AttackRoll int
	Total      int
	Hit        bool
	Critical   bool
	Damage     int
	TargetHP   int
}

type MoveResult struct {
	Fighter     Fighter
	Attack      *AttackResult
	FreeAttacks []AttackResult
}

type SpellResult struct {
	CasterID string
	TargetID string
	SpellID  uint8
	Missiles int
	Damage   int
	Healing  int
	TargetHP int
	Targets  int
}

type Battle struct {
	fighters map[string]Fighter
	rng      *rand.Rand
	round    int
	status   Status
}

func NewBattle(fighters []Fighter, seed int64) (*Battle, error) {
	if len(fighters) == 0 {
		return nil, fmt.Errorf("battle needs at least one fighter")
	}
	b := &Battle{fighters: make(map[string]Fighter, len(fighters)), rng: rand.New(rand.NewSource(seed)), status: StatusActive}
	for _, fighter := range fighters {
		if fighter.ID == "" {
			return nil, fmt.Errorf("fighter has empty ID")
		}
		if _, exists := b.fighters[fighter.ID]; exists {
			return nil, fmt.Errorf("duplicate fighter ID %q", fighter.ID)
		}
		if fighter.MaxHitPoints <= 0 || fighter.HitPoints < 0 || fighter.HitPoints > fighter.MaxHitPoints {
			return nil, fmt.Errorf("fighter %q has invalid hit points", fighter.ID)
		}
		if fighter.ArmorClass < -20 || fighter.ArmorClass > 30 {
			return nil, fmt.Errorf("fighter %q has invalid armor class", fighter.ID)
		}
		if fighter.DamageDiceCount < 0 || fighter.DamageDiceSides < 0 {
			return nil, fmt.Errorf("fighter %q has invalid damage dice", fighter.ID)
		}
		b.fighters[fighter.ID] = fighter
	}
	return b, nil
}

func (b *Battle) Round() int { return b.round }

func (b *Battle) Status() Status { return b.status }

func (b *Battle) Fighters() []Fighter {
	output := make([]Fighter, 0, len(b.fighters))
	for _, fighter := range b.fighters {
		output = append(output, fighter)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].ID < output[j].ID })
	return output
}

// ValidateAttack checks non-random attack preconditions. Game adapters can
// call it before committing resources such as ammunition, while Attack calls
// it before consuming the deterministic RNG stream.
func (b *Battle) ValidateAttack(attackerID, targetID string) error {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 || target.HitPoints <= 0 {
		return fmt.Errorf("dead fighter cannot attack")
	}
	if attacker.HasCombatPosition && target.HasCombatPosition && attacker.MissileWeapon && adjacent(attacker, target) && !attacker.ThrownWeapon {
		return fmt.Errorf("missile weapon cannot attack an adjacent target")
	}
	return nil
}

// StartRound rolls the reference engine's d20-style initiative input for all
// living fighters. Ties are deterministic by fighter ID for reproducibility.
func (b *Battle) StartRound() ([]Turn, error) {
	if b.status != StatusActive {
		return nil, fmt.Errorf("battle is already over")
	}
	b.round++
	b.advanceBlessDurations()
	b.advanceCurseDurations()
	b.advanceProtectionDurations()
	turns := make([]Turn, 0, len(b.fighters))
	ids := make([]string, 0, len(b.fighters))
	for id := range b.fighters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fighter := b.fighters[id]
		if fighter.HitPoints > 0 {
			turns = append(turns, Turn{FighterID: fighter.ID, Initiative: b.rng.Intn(20) + 1 + fighter.InitiativeBonus})
		}
	}
	sort.Slice(turns, func(i, j int) bool {
		if turns[i].Initiative != turns[j].Initiative {
			return turns[i].Initiative > turns[j].Initiative
		}
		return turns[i].FighterID < turns[j].FighterID
	})
	return turns, nil
}

// ResolveAttack applies the recovered attack rule with injected dice. A
// natural 1 misses, a natural 20 always hits, otherwise d20+AttackBonus must
// meet the target AC. damageRoll is the already rolled weapon-dice total.
func (b *Battle) ResolveAttack(attackerID, targetID string, attackRoll, damageRoll int) (AttackResult, error) {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return AttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return AttackResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return AttackResult{}, fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 || target.HitPoints <= 0 {
		return AttackResult{}, fmt.Errorf("dead fighter cannot attack")
	}
	if attackRoll < 1 || attackRoll > 20 {
		return AttackResult{}, fmt.Errorf("attack roll %d is outside d20", attackRoll)
	}
	if damageRoll < 0 {
		return AttackResult{}, fmt.Errorf("negative damage roll")
	}
	if err := b.ValidateAttack(attackerID, targetID); err != nil {
		return AttackResult{}, err
	}
	critical := attackRoll == 20
	targetArmorClass := target.ArmorClass
	if attacker.Evil && target.ProtectedFromEvil {
		targetArmorClass += 2
	}
	if attacker.Good && target.ProtectedFromGood {
		targetArmorClass += 2
	}
	hit := critical || (attackRoll != 1 && attackRoll+attacker.AttackBonus >= targetArmorClass)
	damage := 0
	if hit {
		damage = damageRoll + attacker.DamageBonus
		if damage < 0 {
			damage = 0
		}
		if damage > target.HitPoints {
			damage = target.HitPoints
		}
		target.HitPoints -= damage
		b.fighters[targetID] = target
		b.updateStatus()
	}
	return AttackResult{AttackerID: attackerID, TargetID: targetID, AttackRoll: attackRoll, Total: attackRoll + attacker.AttackBonus, Hit: hit, Critical: critical, Damage: damage, TargetHP: target.HitPoints}, nil
}

// Attack rolls a normal attack using the battle's deterministic RNG. Keeping
// the dice source inside Battle makes the game adapter reproducible by seed,
// while ResolveAttack remains available for exact rule regression tests.
func (b *Battle) Attack(attackerID, targetID string) (AttackResult, error) {
	if err := b.ValidateAttack(attackerID, targetID); err != nil {
		return AttackResult{}, err
	}
	attacker := b.fighters[attackerID]
	if attacker.DamageDiceCount < 1 || attacker.DamageDiceSides < 1 {
		return b.ResolveAttack(attackerID, targetID, b.rng.Intn(20)+1, 0)
	}
	attackRoll := b.rng.Intn(20) + 1
	damageRoll := 0
	for i := 0; i < attacker.DamageDiceCount; i++ {
		damageRoll += b.rng.Intn(attacker.DamageDiceSides) + 1
	}
	return b.ResolveAttack(attackerID, targetID, attackRoll, damageRoll)
}

// AttackSequence resolves the number of attacks granted by the readied
// weapon's RateOfFire projection. A zero value keeps old callers at one
// attack. Target selection after a target falls belongs to the game adapter.
func (b *Battle) AttackSequence(attackerID, targetID string) ([]AttackResult, error) {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return nil, fmt.Errorf("unknown attacker %q", attackerID)
	}
	attacks := attacker.AttacksPerTurn
	if attacks < 1 {
		attacks = 1
	}
	results := make([]AttackResult, 0, attacks)
	for index := 0; index < attacks && b.status == StatusActive; index++ {
		result, err := b.Attack(attackerID, targetID)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if result.TargetHP <= 0 {
			break
		}
	}
	return results, nil
}

// Move translates a living fighter by one grid step, or attacks an enemy
// occupying the destination square. Battlefield bounds, terrain and the
// RuleBook's rear-facing details remain owned by a future map/rules adapter.
func (b *Battle) Move(fighterID string, dx, dy int) (Fighter, error) {
	result, err := b.MoveWithFreeAttacks(fighterID, dx, dy)
	return result.Fighter, err
}

// MoveWithFreeAttacks also applies the RuleBook's bounded free-attack trigger
// when a party fighter leaves an enemy's adjacent square. A party fighter
// entering an enemy square resolves an attack without changing squares;
// facing/rear AC is still left to the future combat-map rules layer.
func (b *Battle) MoveWithFreeAttacks(fighterID string, dx, dy int) (MoveResult, error) {
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return MoveResult{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	if b.status != StatusActive {
		return MoveResult{}, fmt.Errorf("battle is already over")
	}
	if fighter.HitPoints <= 0 {
		return MoveResult{}, fmt.Errorf("dead fighter cannot move")
	}
	if dx < -1 || dx > 1 || dy < -1 || dy > 1 || (dx == 0 && dy == 0) {
		return MoveResult{}, fmt.Errorf("move delta (%d,%d) is not one grid step", dx, dy)
	}
	if !fighter.HasCombatPosition {
		return MoveResult{}, fmt.Errorf("fighter %q has no combat position", fighterID)
	}
	old := fighter
	nextX, nextY := fighter.CombatX+dx, fighter.CombatY+dy
	for _, other := range b.Fighters() {
		if other.ID == fighterID || other.HitPoints <= 0 || !other.HasCombatPosition || other.CombatX != nextX || other.CombatY != nextY {
			continue
		}
		if fighter.Side == SideParty && other.Side == SideEnemy {
			attack, err := b.Attack(fighterID, other.ID)
			if err != nil {
				return MoveResult{}, err
			}
			return MoveResult{Fighter: fighter, Attack: &attack}, nil
		}
		return MoveResult{}, fmt.Errorf("destination (%d,%d) is occupied", nextX, nextY)
	}
	fighter.CombatX, fighter.CombatY = nextX, nextY
	b.fighters[fighterID] = fighter
	result := MoveResult{Fighter: fighter}
	if old.Side == SideParty {
		for _, enemy := range b.fighters {
			if enemy.Side != SideEnemy || enemy.HitPoints <= 0 || !enemy.HasCombatPosition {
				continue
			}
			if adjacent(old, enemy) && !adjacent(fighter, enemy) {
				attack, err := b.Attack(enemy.ID, fighter.ID)
				if err != nil {
					return MoveResult{}, err
				}
				result.FreeAttacks = append(result.FreeAttacks, attack)
				if b.status != StatusActive {
					break
				}
			}
		}
	}
	return result, nil
}

// CastMagicMissile applies the verified RuleBook first-level effect: every
// missile deals 2-5 damage and has no saving throw. The seed-owned RNG keeps
// the result replayable while the game adapter owns spell-slot consumption.
func (b *Battle) CastMagicMissile(casterID, targetID string, level int) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if level < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	missiles := (level + 1) / 2
	if missiles > 6 {
		missiles = 6
	}
	damage := 0
	for index := 0; index < missiles; index++ {
		damage += b.rng.Intn(4) + 2
	}
	if damage > target.HitPoints {
		damage = target.HitPoints
	}
	target.HitPoints -= damage
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 7, Missiles: missiles, Damage: damage, TargetHP: target.HitPoints}, nil
}

// CastCureLightWounds applies the verified 1-8 HP touch-heal effect. The
// caller decides whether the caster has the clerical slot and which friendly
// target is legal; Battle only owns deterministic dice and HP mutation.
func (b *Battle) CastCureLightWounds(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot use Cure Light Wounds")
	}
	healing := b.rng.Intn(8) + 1
	if healing > target.MaxHitPoints-target.HitPoints {
		healing = target.MaxHitPoints - target.HitPoints
	}
	if healing < 0 {
		healing = 0
	}
	target.HitPoints += healing
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 3, Healing: healing, TargetHP: target.HitPoints}, nil
}

// CastCauseLightWounds applies the verified 1-8 HP touch damage. The target
// must be an adjacent living enemy when both fighters carry CombatMap
// positions; position-less direct callers retain the bounded fallback.
func (b *Battle) CastCauseLightWounds(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideEnemy {
		return SpellResult{}, fmt.Errorf("Cause Light Wounds target %q is not an enemy", targetID)
	}
	if caster.HasCombatPosition && target.HasCombatPosition && !adjacent(caster, target) {
		return SpellResult{}, fmt.Errorf("Cause Light Wounds target %q is out of touch range", targetID)
	}
	damage := b.rng.Intn(8) + 1
	if damage > target.HitPoints {
		damage = target.HitPoints
	}
	target.HitPoints -= damage
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 4, Damage: damage, TargetHP: target.HitPoints}, nil
}

// CastProtectionFromEvil applies the verified conditional AC protection. The
// caller supplies a living party target; alignment-aware attack resolution is
// intentionally kept in Battle rather than mutating base ArmorClass.
func (b *Battle) CastProtectionFromEvil(casterID, targetID string, casterLevel int) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideParty {
		return SpellResult{}, fmt.Errorf("Protection from Evil target %q is not party", targetID)
	}
	if casterLevel < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	if target.ProtectedFromEvil {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 6}, nil
	}
	target.ProtectedFromEvil = true
	target.ProtectionEvilRounds = 3 * casterLevel
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 6, Targets: 1}, nil
}

func (b *Battle) CastProtectionFromGood(casterID, targetID string, casterLevel int) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideParty {
		return SpellResult{}, fmt.Errorf("Protection from Good target %q is not party", targetID)
	}
	if casterLevel < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	if target.ProtectedFromGood {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 7}, nil
	}
	target.ProtectedFromGood = true
	target.ProtectionGoodRounds = 3 * casterLevel
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 7, Targets: 1}, nil
}

func adjacent(first, second Fighter) bool {
	dx := first.CombatX - second.CombatX
	if dx < 0 {
		dx = -dx
	}
	dy := first.CombatY - second.CombatY
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1 && (dx != 0 || dy != 0)
}

// CastBless applies the verified first-level party-wide attack bonus, using
// CombatMap adjacency and the six-round duration contract.
func (b *Battle) CastBless(casterID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	affected := 0
	for id, fighter := range b.fighters {
		if fighter.Side != SideParty || fighter.HitPoints <= 0 || fighter.Blessed || b.adjacentToLivingEnemy(fighter) {
			continue
		}
		fighter.AttackBonus++
		fighter.Blessed = true
		fighter.BlessRounds = 6
		b.fighters[id] = fighter
		affected++
	}
	return SpellResult{CasterID: casterID, SpellID: 1, Targets: affected}, nil
}

func (b *Battle) adjacentToLivingEnemy(party Fighter) bool {
	if !party.HasCombatPosition {
		return false
	}
	for _, enemy := range b.fighters {
		if enemy.Side != SideEnemy || enemy.HitPoints <= 0 || !enemy.HasCombatPosition {
			continue
		}
		if adjacent(party, enemy) {
			return true
		}
	}
	return false
}

func (b *Battle) advanceBlessDurations() {
	for id, fighter := range b.fighters {
		if !fighter.Blessed || fighter.BlessRounds <= 0 {
			continue
		}
		fighter.BlessRounds--
		if fighter.BlessRounds == 0 {
			fighter.AttackBonus--
			fighter.Blessed = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) CastCurse(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideEnemy {
		return SpellResult{}, fmt.Errorf("Curse target %q is not an enemy", targetID)
	}
	if target.Cursed || b.adjacentToLivingParty(target) {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 2}, nil
	}
	target.AttackBonus--
	target.Cursed = true
	target.CurseRounds = 6
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 2, Targets: 1}, nil
}

func (b *Battle) adjacentToLivingParty(enemy Fighter) bool {
	if !enemy.HasCombatPosition {
		return false
	}
	for _, party := range b.fighters {
		if party.Side != SideParty || party.HitPoints <= 0 || !party.HasCombatPosition {
			continue
		}
		if adjacent(enemy, party) {
			return true
		}
	}
	return false
}

func (b *Battle) advanceCurseDurations() {
	for id, fighter := range b.fighters {
		if !fighter.Cursed || fighter.CurseRounds <= 0 {
			continue
		}
		fighter.CurseRounds--
		if fighter.CurseRounds == 0 {
			fighter.AttackBonus++
			fighter.Cursed = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) advanceProtectionDurations() {
	for id, fighter := range b.fighters {
		if !fighter.ProtectedFromEvil || fighter.ProtectionEvilRounds <= 0 {
			continue
		}
		fighter.ProtectionEvilRounds--
		if fighter.ProtectionEvilRounds == 0 {
			fighter.ProtectedFromEvil = false
		}
		b.fighters[id] = fighter
	}
	for id, fighter := range b.fighters {
		if !fighter.ProtectedFromGood || fighter.ProtectionGoodRounds <= 0 {
			continue
		}
		fighter.ProtectionGoodRounds--
		if fighter.ProtectionGoodRounds == 0 {
			fighter.ProtectedFromGood = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) updateStatus() {
	partyAlive, enemyAlive := false, false
	for _, fighter := range b.fighters {
		if fighter.HitPoints <= 0 {
			continue
		}
		if fighter.Side == SideParty {
			partyAlive = true
		} else {
			enemyAlive = true
		}
	}
	switch {
	case partyAlive && enemyAlive:
		b.status = StatusActive
	case partyAlive:
		b.status = StatusPartyWon
	case enemyAlive:
		b.status = StatusEnemyWon
	default:
		b.status = StatusDraw
	}
}
