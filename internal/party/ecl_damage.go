package party

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// DamageOutcome records one ECL DAMAGE target transaction. The resolver is
// deliberately injected with dice so an event replay can be deterministic.
type DamageOutcome struct {
	TargetIndex int
	DamageRoll  int
	Applied     int
	SaveRoll    int
	Saved       bool
	Hit         bool
	Health      HealthStatus
}

// DamageHitResolver supplies the original CanHitTarget branch. It receives
// the raw command bonus and the same injected d20 source used by damage and
// target selection, so natural 1/20 handling remains visible to the adapter.
type DamageHitResolver func(target Character, bonus int, rollDie func(int) int) (bool, error)

// ECLHitContext carries the combat state inspected by Type_16 effects.
// Callers that do not have action state can use its zero value.
type ECLHitContext struct {
	ActionDelay int
	CombatRound int
}

// ApplyDamage clamps HP at zero and reports the amount actually removed. The
// status-aware variant below preserves the reference DOS down/death rules.
func (c *Character) ApplyDamage(amount int) int {
	applied, _ := c.ApplyDamageWithHealthStatus(amount)
	return applied
}

// ApplyDamageWithHealthStatus projects reference damage_player. A 1..9 point
// overkill leaves a character dying, a 10+ overkill kills, and exact zero is
// unconscious unless the previous state was animated. Downed characters keep
// HP at zero and cannot remain in combat.
func (c *Character) ApplyDamageWithHealthStatus(amount int) (int, HealthStatus) {
	if c == nil || amount <= 0 || c.HitPoints <= 0 {
		if c == nil {
			return 0, HealthStatusOK
		}
		return 0, c.HealthStatus
	}
	previous := c.HealthStatus
	applied := amount
	if applied > c.HitPoints {
		applied = c.HitPoints
	}
	newHP, overkill := c.HitPoints, 0
	if amount <= c.HitPoints {
		newHP = c.HitPoints - amount
	} else {
		overkill = amount - c.HitPoints
	}
	status := previous
	if overkill > 9 || (newHP == 0 && previous == HealthStatusAnimated) {
		status = HealthStatusDead
	} else if overkill > 0 {
		status = HealthStatusDying
	} else if newHP == 0 {
		status = HealthStatusUnconscious
	}
	c.HealthStatus = status
	if status == HealthStatusDying {
		c.Bleeding = overkill
	} else {
		c.Bleeding = 0
	}
	if status != HealthStatusOK && status != HealthStatusAnimated {
		c.HitPoints = 0
	} else {
		c.HitPoints = newHP
	}
	return applied, status
}

// CanHitECLDamageTarget implements the verified CanHitTarget arithmetic after
// the caller supplies the target's projected AC. It uses the default hit
// context; callers with combat action state should use the Context variant.
func CanHitECLDamageTarget(target Character, armorClass, bonus int, rollDie func(int) int) (bool, error) {
	return CanHitECLDamageTargetWithContext(target, armorClass, bonus, ECLHitContext{}, rollDie)
}

// CanHitECLDamageTargetWithContext projects the verified Type_16 hit effects.
// Natural 20 is first changed to 100 in the reference engine, then effects
// are evaluated; blink can therefore still force a miss when action delay is
// zero. The first FX effect-data byte stores displace's persistent 0x10
// consumed bit, matching the reference affect_data byte.
func CanHitECLDamageTargetWithContext(target Character, armorClass, bonus int, context ECLHitContext, rollDie func(int) int) (bool, error) {
	if rollDie == nil {
		return false, fmt.Errorf("CanHitTarget requires an injected d20 roller")
	}
	roll := rollDie(20)
	if roll < 1 || roll > 20 {
		return false, fmt.Errorf("CanHitTarget roll %d is outside 1..20", roll)
	}
	if roll == 1 {
		return false, nil
	}
	if roll == 20 {
		roll = 100
	}
	for index, effect := range target.Effects {
		if !effect.Active {
			continue
		}
		switch effect.Kind {
		case 0x19, 0x47: // invisibility / invisible
			roll -= 4
		case 0x25: // blink: AffectBlink when action.delay == 0
			if context.ActionDelay == 0 {
				roll = -1
			}
		case 0x59: // displace: AffectDisplace
			if context.CombatRound == 0 && roll == 0 {
				target.Effects[index].Data[0] &= 0x0F
			} else if target.Effects[index].Data[0]&0x10 == 0 {
				roll = -1
				target.Effects[index].Data[0] |= 0x10
			}
		}
	}
	return roll >= 0 && roll+bonus > armorClass, nil
}

// ApplyECLDamage resolves the verified selected-target and whole-party forms
// of CoAB DAMAGE. The original flags 0x80/0x40/0x20/0x10 and saveFlags 0x80
// are preserved here; random-target and CanHitTarget branches remain explicit
// errors until their party selection context is available.
func (r Roster) ApplyECLDamage(request ecl.DamageRequest, selectedIndex int, rollDie, rollSave func(int) int) ([]DamageOutcome, error) {
	return r.ApplyECLDamageWithHitResolver(request, selectedIndex, rollDie, rollSave, nil)
}

// ApplyECLDamageWithHitResolver extends ApplyECLDamage with the verified
// random-target/CanHitTarget branch. A nil hit resolver keeps that branch an
// explicit boundary while preserving the older selected-target API.
func (r Roster) ApplyECLDamageWithHitResolver(request ecl.DamageRequest, selectedIndex int, rollDie, rollSave func(int) int, hitTarget DamageHitResolver) ([]DamageOutcome, error) {
	if len(r) == 0 {
		return nil, fmt.Errorf("ECL DAMAGE requires a non-empty party")
	}
	if request.DiceCount == 0 || request.DiceSize == 0 {
		return nil, fmt.Errorf("ECL DAMAGE dice must be positive: count=%d size=%d", request.DiceCount, request.DiceSize)
	}
	if rollDie == nil || rollSave == nil {
		return nil, fmt.Errorf("ECL DAMAGE requires injected dice and saving-throw rollers")
	}
	rollDamage := func() (int, error) {
		total := int(request.Bonus)
		for i := 0; i < int(request.DiceCount); i++ {
			roll := rollDie(int(request.DiceSize))
			if roll < 1 || roll > int(request.DiceSize) {
				return 0, fmt.Errorf("damage die roll %d is outside 1..%d", roll, request.DiceSize)
			}
			total += roll
		}
		return total, nil
	}
	damage, err := rollDamage()
	if err != nil {
		return nil, err
	}
	if request.Flags&0x80 == 0 {
		if hitTarget == nil {
			return nil, fmt.Errorf("ECL DAMAGE random-target branch requires CanHitTarget resolver")
		}
		count := int(uint8(request.Flags))
		outcomes := make([]DamageOutcome, 0, count)
		for index := 0; index < count; index++ {
			targetRoll := rollDie(len(r))
			if targetRoll < 1 || targetRoll > len(r) {
				return nil, fmt.Errorf("ECL DAMAGE target roll %d is outside 1..%d", targetRoll, len(r))
			}
			targetIndex := targetRoll - 1
			hit, hitErr := hitTarget(r[targetIndex], int(uint8(request.SaveFlags)), rollDie)
			if hitErr != nil {
				return nil, hitErr
			}
			outcome := DamageOutcome{TargetIndex: targetIndex, DamageRoll: damage, Hit: hit, Health: r[targetIndex].HealthStatus}
			if hit {
				outcome.Applied, outcome.Health = r[targetIndex].ApplyDamageWithHealthStatus(damage)
			}
			outcomes = append(outcomes, outcome)
			// The reference rerolls the next damage packet after every random
			// target attempt; retaining that order keeps replay streams stable.
			damage, err = rollDamage()
			if err != nil {
				return nil, err
			}
		}
		return outcomes, nil
	}
	saveBonus := int(request.Flags & 0x1F)
	saveType := uint8(request.SaveFlags & 7)
	resolveTarget := func(index int, saveType uint8, autoDamage, noSave bool) (DamageOutcome, error) {
		if index < 0 || index >= len(r) {
			return DamageOutcome{}, fmt.Errorf("ECL DAMAGE target index %d is outside party", index)
		}
		outcome := DamageOutcome{TargetIndex: index, DamageRoll: damage, Health: r[index].HealthStatus}
		if autoDamage {
			outcome.Applied, outcome.Health = r[index].ApplyDamageWithHealthStatus(damage)
			return outcome, nil
		}
		if noSave {
			outcome.Applied, outcome.Health = r[index].ApplyDamageWithHealthStatus(damage)
			return outcome, nil
		}
		if saveType >= 5 {
			return DamageOutcome{}, fmt.Errorf("ECL DAMAGE save type %d is outside 0..4", saveType)
		}
		roll := rollSave(20)
		if roll < 1 || roll > 20 {
			return DamageOutcome{}, fmt.Errorf("saving throw roll %d is outside 1..20", roll)
		}
		outcome.SaveRoll = roll
		if roll == 1 {
			outcome.Saved = false
		} else if roll == 20 {
			outcome.Saved = true
		} else if len(r[index].SavingThrows) != 5 {
			return DamageOutcome{}, fmt.Errorf("character %q has no five-byte saving throws", r[index].ID)
		} else {
			outcome.Saved = roll+saveBonus+int(r[index].SavingThrowBonus) >= int(r[index].SavingThrows[saveType])
		}
		if !outcome.Saved || request.Flags&0x10 != 0 {
			outcome.Applied, outcome.Health = r[index].ApplyDamageWithHealthStatus(damage)
		}
		return outcome, nil
	}

	if request.Flags&0x40 != 0 {
		outcomes := make([]DamageOutcome, 0, len(r))
		for index := range r {
			outcome, outcomeErr := resolveTarget(index, saveType, request.Flags&0x20 != 0, false)
			if outcomeErr != nil {
				return nil, outcomeErr
			}
			outcomes = append(outcomes, outcome)
		}
		return outcomes, nil
	}
	if request.SaveFlags&0x80 == 0 {
		return nil, fmt.Errorf("ECL DAMAGE selected target requires saveFlags 0x80")
	}
	// The reference uses save type - 1 for the selected-character branch;
	// type zero is the explicit no-save case.
	selectedNoSave := saveType == 0
	if saveType > 0 {
		saveType--
	}
	outcome, err := resolveTarget(selectedIndex, saveType, false, selectedNoSave)
	if err != nil {
		return nil, err
	}
	return []DamageOutcome{outcome}, nil
}
