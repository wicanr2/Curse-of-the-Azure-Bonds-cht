package game

// SoundEvent is a renderer-neutral gameplay sound intent. Platform adapters
// map the semantic event to DOS WAV, PC-98 SOUNDFX, or another audio backend;
// the game rules must not assume that different ports share selector numbers.
type SoundEvent string

const (
	SoundStop      SoundEvent = "stop"
	SoundCast      SoundEvent = "cast"
	SoundMiss      SoundEvent = "miss"
	SoundSpellHit  SoundEvent = "spell_hit"
	SoundDead      SoundEvent = "dead"
	SoundWhistle   SoundEvent = "whistle"
	SoundHit       SoundEvent = "hit"
	SoundLightning SoundEvent = "lightning"
	SoundSwish     SoundEvent = "swish"
	SoundStep      SoundEvent = "step"
	SoundFireball  SoundEvent = "fireball"
	SoundArrow     SoundEvent = "arrow"
	SoundOverture  SoundEvent = "overture"
	SoundCombat    SoundEvent = "combat"
	SoundCrash     SoundEvent = "crash"
)

func (s *State) requestSound(event SoundEvent) {
	s.pendingSoundEvents = append(s.pendingSoundEvents, event)
}

// ConsumeSoundEvents transfers one-shot audio intents to a platform adapter.
// Events are consumed exactly once; an adapter may ignore unknown/no-op IDs.
func (s *State) ConsumeSoundEvents() []SoundEvent {
	events := append([]SoundEvent(nil), s.pendingSoundEvents...)
	s.pendingSoundEvents = nil
	return events
}
