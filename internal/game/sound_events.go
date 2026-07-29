package game

// SoundEvent is the reference seg044 sound selector exposed as a renderer-
// neutral intent. Keeping this type in game avoids making deterministic rules
// depend on Ebiten or an audio device.
type SoundEvent int8

const (
	SoundStop     SoundEvent = 0
	SoundNoOp     SoundEvent = 1
	SoundMissile  SoundEvent = 2
	SoundMagicHit SoundEvent = 3
	SoundDeath    SoundEvent = 5
	SoundGeneric5 SoundEvent = 6
	SoundHit      SoundEvent = 7
	// SoundLightning preserves the reference sound_8 selector even though the
	// recovered PC resource table has no WAV sample for it.
	SoundLightning SoundEvent = 8
	SoundMiss      SoundEvent = 9
	SoundStep      SoundEvent = 10
	SoundGeneric10 SoundEvent = 11
	SoundStart     SoundEvent = 13
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
