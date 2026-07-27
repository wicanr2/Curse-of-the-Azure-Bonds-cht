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
}

// DamageHitResolver supplies the original CanHitTarget branch. It receives
// the raw command bonus and the same injected d20 source used by damage and
// target selection, so natural 1/20 handling remains visible to the adapter.
type DamageHitResolver func(target Character, bonus int, rollDie func(int) int) (bool, error)

// ApplyDamage clamps HP at zero and reports the amount actually removed.
func (c *Character) ApplyDamage(amount int) int {
	if c == nil || amount <= 0 || c.HitPoints <= 0 {
		return 0
	}
	if amount > c.HitPoints {
		amount = c.HitPoints
	}
	c.HitPoints -= amount
	return amount
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
			outcome := DamageOutcome{TargetIndex: targetIndex, DamageRoll: damage, Hit: hit}
			if hit {
				outcome.Applied = r[targetIndex].ApplyDamage(damage)
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
		outcome := DamageOutcome{TargetIndex: index, DamageRoll: damage}
		if autoDamage {
			outcome.Applied = r[index].ApplyDamage(damage)
			return outcome, nil
		}
		if noSave {
			outcome.Applied = r[index].ApplyDamage(damage)
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
			outcome.Saved = roll+saveBonus >= int(r[index].SavingThrows[saveType])
		}
		if !outcome.Saved || request.Flags&0x10 != 0 {
			outcome.Applied = r[index].ApplyDamage(damage)
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
