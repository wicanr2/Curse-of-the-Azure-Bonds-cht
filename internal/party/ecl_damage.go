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
}

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
	if request.Flags&0x80 == 0 {
		return nil, fmt.Errorf("ECL DAMAGE random-target branch is not resolved")
	}

	damage, err := rollDamage()
	if err != nil {
		return nil, err
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
