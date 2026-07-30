// Package pc98sfx imports the PC-9801 speaker-effect program from the user's
// exact local GAME.EXE. The executable bytes remain local and are never
// embedded in the remake.
package pc98sfx

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	GameSHA256     = "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
	tableOffset    = 0xE66C
	selectorCount  = 16
	wordsPerEffect = 20
)

// StepKind identifies one action performed by GAME.EXE SOUNDFX.
type StepKind string

const (
	StepPulse StepKind = "pulse"
	StepDelay StepKind = "delay"
)

// Step is a renderer-neutral reconstruction of one SOUNDFX action.
// FrequencyOrPeriod preserves the original SOUND argument without assuming
// that it is an IBM-PC PIT frequency.
type Step struct {
	Kind              StepKind `json:"kind"`
	FrequencyOrPeriod uint16   `json:"frequency_or_period,omitempty"`
	PulseCount        uint16   `json:"pulse_count,omitempty"`
	DelayMilliseconds uint16   `json:"delay_milliseconds,omitempty"`
}

// Effect is one selector accepted by GAME.EXE SOUNDFX.
type Effect struct {
	Selector int    `json:"selector"`
	Symbol   string `json:"symbol"`
	Event    string `json:"event,omitempty"`
	NoOp     bool   `json:"no_op,omitempty"`
	Source   string `json:"source"`
	Steps    []Step `json:"steps,omitempty"`
}

var selectorMetadata = [...]struct {
	symbol string
	event  string
}{
	{"SOUNDOFF", ""},
	{"SOUNDON", ""},
	{"CASTFX", "cast"},
	{"MISSFX", "miss"},
	{"SPELLHITFX", "spell_hit"},
	{"DEADFX", "dead"},
	{"WHISTLEFX", "whistle"},
	{"HITFX", "hit"},
	{"LIGHTNINGFX", "lightning"},
	{"SWISHFX", "swish"},
	{"PADFX", "step"},
	{"FIREBALLFX", "fireball"},
	{"ARROWFX", "arrow"},
	{"OVERTUREFX", "overture"},
	{"COMBATFX", "combat"},
	{"CRASHFX", "crash"},
}

// SelectorForEvent maps a renderer-neutral gameplay event to the exact PC-98
// SOUNDFX selector named by the Borland symbol table.
func SelectorForEvent(event string) (int, bool) {
	if event == "stop" {
		return 255, true
	}
	for selector, metadata := range selectorMetadata {
		if metadata.event == event {
			return selector, true
		}
	}
	return 0, false
}

// Import verifies GAME.EXE and reconstructs all selectors without retaining
// references to the commercial input.
func Import(game []byte) ([]Effect, error) {
	sum := sha256.Sum256(game)
	if digest := hex.EncodeToString(sum[:]); digest != GameSHA256 {
		return nil, fmt.Errorf("GAME.EXE SHA-256 %s, want %s", digest, GameSHA256)
	}
	tableSize := selectorCount * wordsPerEffect * 2
	if tableOffset+tableSize > len(game) {
		return nil, fmt.Errorf("GAME.EXE SOUNDFX table exceeds file bounds")
	}

	effects := make([]Effect, selectorCount)
	for selector := range effects {
		effects[selector] = decodeEffect(game, selector)
	}
	return effects, nil
}

func decodeEffect(game []byte, selector int) Effect {
	metadata := selectorMetadata[selector]
	effect := Effect{
		Selector: selector,
		Symbol:   metadata.symbol,
		Event:    metadata.event,
	}
	switch selector {
	case 0, 1, 13, 14, 15:
		effect.NoOp = true
		effect.Source = "immediate-return"
		return effect
	case 2, 4, 6, 9:
		effect.Source = "formula"
		effect.Steps = []Step{{
			Kind:              StepPulse,
			FrequencyOrPeriod: uint16(selector * 10),
			PulseCount:        uint16(250 / selector),
		}}
		return effect
	default:
		effect.Source = "table"
	}

	base := tableOffset + selector*wordsPerEffect*2
	// The original loop starts at one, not zero, and stops after word 20.
	for index := 1; index <= wordsPerEffect; index++ {
		value := binary.LittleEndian.Uint16(game[base+index*2:])
		if value == 0 {
			break
		}
		if value > 2500 {
			effect.Steps = append(effect.Steps, Step{
				Kind:              StepDelay,
				DelayMilliseconds: 5,
			})
			continue
		}
		effect.Steps = append(effect.Steps, Step{
			Kind:              StepPulse,
			FrequencyOrPeriod: value,
			PulseCount:        uint16(2000 / value),
		})
	}
	return effect
}
