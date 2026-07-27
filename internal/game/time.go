package game

import (
	"fmt"
	"math"
)

// ReferenceTimeScales are the observed DOS ovr021 timeScales table. Slots
// are kept as raw units so callers can preserve the original rest/map clock.
var referenceTimeScales = [...]uint16{10, 10, 6, 24, 30, 12, 0x100}

// GameTimeSlots returns a copy of the raw seven-slot clock.
func (s *State) GameTimeSlots() [7]uint16 { return s.gameClock }

// GameAgeCycles reports slot-6 overflows observed by the shared adapter.
// Character-age field writeback remains a player-record task.
func (s *State) GameAgeCycles() uint32 { return s.gameAgeCycles }

// AdvanceGameTime mirrors step_game_time: it advances the raw clock and then
// expires finite party/battle effects using the corresponding elapsed minutes.
func (s *State) AdvanceGameTime(timeSlot int, amount uint16) error {
	if timeSlot < 0 || timeSlot >= len(referenceTimeScales) {
		return fmt.Errorf("game time slot %d is outside 0..6", timeSlot)
	}
	if amount == 0 {
		return nil
	}
	s.addGameClock(timeSlot, amount)
	minutes := uint64(amount)
	for slot := timeSlot; slot > 1; slot-- {
		minutes *= uint64(referenceTimeScales[slot-1])
	}
	for minutes > 0 {
		chunk := minutes
		if chunk > math.MaxUint16 {
			chunk = math.MaxUint16
		}
		s.advanceEffects(uint16(chunk))
		minutes -= chunk
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
	}
}

func (s *State) advanceEffects(minutes uint16) {
	for index := range s.partyRoster {
		s.partyRoster[index].AdvanceEffects(minutes)
	}
	if s.battle != nil {
		s.battle.AdvanceMonsterAffects(minutes)
	}
}

// AdvanceCombatEffects is a small adapter for tests/frontends that already
// have an elapsed minute count and do not want to mutate the world clock.
func (s *State) AdvanceCombatEffects(minutes uint16) int {
	if minutes == 0 || s.battle == nil {
		return 0
	}
	return s.battle.AdvanceMonsterAffects(minutes)
}
