package game

import (
	"fmt"
	"math"

	engineeffecttime "github.com/wicanr2/golden-box-remake-engine/combat/effecttime"
)

// ReferenceTimeScales are the observed DOS ovr021 timeScales table. Slots
// are kept as raw units so callers can preserve the original rest/map clock.
var referenceTimeScales = [...]uint16{10, 10, 6, 24, 30, 12, 0x100}

// GameTimeDisplay is the renderer-neutral view of the reference Area1 clock.
// The raw order is deliberately kept in GameTimeSlots; this view only applies
// the reference field mapping used by the map-position display.
type GameTimeDisplay struct {
	Minute uint16
	Hour   uint16
	Day    uint16
	Month  uint16
	Year   uint16
}

// GameTimeDisplay returns the reference clock fields used by the remake HUD.
// Area1 stores decimal minutes as separate tens and ones slots.
func (s *State) GameTimeDisplay() GameTimeDisplay {
	clock := s.gameClock
	return GameTimeDisplay{
		Minute: clock[2]*10 + clock[1],
		Hour:   clock[3],
		Day:    clock[4],
		Month:  clock[5],
		Year:   clock[6],
	}
}

// GameTimeText is the compact Traditional Chinese clock shown by frontends.
func (s *State) GameTimeText() string {
	clock := s.GameTimeDisplay()
	return fmt.Sprintf(s.catalog.Text("game_time", "game_time"), clock.Hour, clock.Minute, clock.Day, clock.Month, clock.Year)
}

// GameTimeSlots returns a copy of the raw seven-slot clock.
func (s *State) GameTimeSlots() [7]uint16 { return s.gameClock }

// GameAgeCycles reports slot-6 overflows observed by the shared adapter.
// Character-age field writeback remains a player-record task.
func (s *State) GameAgeCycles() uint32 { return s.gameAgeCycles }

// AdvanceGameTime mirrors step_game_time: it advances the raw clock and then
// expires finite party/battle effects using the corresponding EFFECTREC ticks.
func (s *State) AdvanceGameTime(timeSlot int, amount uint16) error {
	if timeSlot < 0 || timeSlot >= len(referenceTimeScales) {
		return fmt.Errorf("game time slot %d is outside 0..6", timeSlot)
	}
	if amount == 0 {
		return nil
	}
	s.addGameClock(timeSlot, amount)
	s.Area.GameTime = s.gameClock
	ticks, err := engineeffecttime.DurationTicks(engineeffecttime.Unit(timeSlot), amount)
	if err != nil {
		return err
	}
	for ticks > 0 {
		chunk := ticks
		if chunk > math.MaxUint16 {
			chunk = math.MaxUint16
		}
		s.advanceEffects(uint16(chunk))
		ticks -= chunk
	}
	return nil
}

// AdvanceGameTimeHours is the REST-facing adapter. Reference rest_time uses
// slot 1 in minute-sized steps; chunking keeps the public uint16 API safe for
// unusually long deterministic tests.
func (s *State) AdvanceGameTimeHours(hours int) error {
	if hours < 0 {
		return fmt.Errorf("rest hours cannot be negative: %d", hours)
	}
	for hours > 0 {
		chunk := hours
		if chunk > math.MaxUint16/60 {
			chunk = math.MaxUint16 / 60
		}
		if err := s.AdvanceGameTime(1, uint16(chunk*60)); err != nil {
			return err
		}
		hours -= chunk
	}
	return nil
}

func (s *State) addGameClock(timeSlot int, amount uint16) {
	carry := uint32(amount)
	for slot := timeSlot; slot < len(s.gameClock) && carry > 0; slot++ {
		total := uint32(s.gameClock[slot]) + carry
		base := uint32(referenceTimeScales[slot])
		s.gameClock[slot] = uint16(total % base)
		carry = total / base
	}
	if carry > 0 {
		s.gameAgeCycles += carry
		for index := range s.partyRoster {
			age := int64(s.partyRoster[index].Age) + int64(carry)
			if age > math.MaxInt16 {
				age = math.MaxInt16
			}
			if age < math.MinInt16 {
				age = math.MinInt16
			}
			s.partyRoster[index].Age = int16(age)
		}
	}
}

func (s *State) advanceEffects(ticks uint16) {
	for index := range s.partyRoster {
		s.partyRoster[index].AdvanceEffects(ticks)
	}
	if s.battle != nil {
		s.battle.AdvanceMonsterAffects(ticks)
	}
}

// AdvanceCombatEffects is a small adapter for tests/frontends that already
// have an elapsed EFFECTREC tick count and do not want to mutate the clock.
func (s *State) AdvanceCombatEffects(ticks uint16) int {
	if ticks == 0 || s.battle == nil {
		return 0
	}
	return s.battle.AdvanceMonsterAffects(ticks)
}
