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
	ID              string
	Name            string
	Side            Side
	HitPoints       int
	MaxHitPoints    int
	ArmorClass      int
	AttackBonus     int
	DamageDiceCount int
	DamageDiceSides int
	DamageBonus     int
	InitiativeBonus int
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

// StartRound rolls the reference engine's d20-style initiative input for all
// living fighters. Ties are deterministic by fighter ID for reproducibility.
func (b *Battle) StartRound() ([]Turn, error) {
	if b.status != StatusActive {
		return nil, fmt.Errorf("battle is already over")
	}
	b.round++
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
	critical := attackRoll == 20
	hit := critical || (attackRoll != 1 && attackRoll+attacker.AttackBonus >= target.ArmorClass)
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
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return AttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
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
