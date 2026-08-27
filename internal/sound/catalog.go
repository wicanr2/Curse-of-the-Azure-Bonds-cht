// Package sound contains the reference PC sound IDs and the converted OGG
// asset mapping. Playback stays in the Ebiten adapter so game state remains
// deterministic and testable without an audio device.
package sound

// ID is the byte-sized selector used by the DOS WAV reference adapter.
type ID uint8

const (
	Stop ID = 0
	NoOp ID = 1

	Missile   ID = 2
	MagicHit  ID = 3
	Death     ID = 5
	Sound5    ID = 6
	Hit       ID = 7
	Lightning ID = 8
	Miss      ID = 9
	Step      ID = 10
	Sound10   ID = 11
	Start     ID = 13
	Combat    ID = 14
	Crash     ID = 15
)

// Event is the platform-neutral sound vocabulary accepted by adapters.
// It mirrors game.SoundEvent without importing game into the renderer package.
type Event string

// DOSID maps a semantic event to the recovered DOS WAV selector. PC-98 uses a
// different mapping for several events and must not call this function.
func DOSID(event Event) (ID, bool) {
	ids := map[Event]ID{
		"stop":      Stop,
		"cast":      Missile,
		"arrow":     Missile,
		"miss":      Miss,
		"spell_hit": MagicHit,
		"dead":      Death,
		"whistle":   Sound5,
		"hit":       Hit,
		"lightning": Lightning,
		"swish":     Miss,
		"step":      Step,
		"fireball":  Sound10,
		"combat":    Combat,
		"crash":     Crash,
		"overture":  Start,
	}
	id, ok := ids[event]
	return id, ok
}

// AssetName returns the OGG corresponding to the reference sound
// selector. Missing reference selectors intentionally return false: the
// original engine also leaves several sound IDs without a sample.
func AssetName(id ID) (string, bool) {
	assets := map[ID]string{
		Missile:  "missle.ogg",
		MagicHit: "magic_hit.ogg",
		Death:    "death.ogg",
		Sound5:   "sound_5.ogg",
		Hit:      "hit.ogg",
		Miss:     "miss.ogg",
		Step:     "step.ogg",
		Sound10:  "sound_10.ogg",
		Start:    "start_sound.ogg",
	}
	name, ok := assets[id]
	return name, ok
}
