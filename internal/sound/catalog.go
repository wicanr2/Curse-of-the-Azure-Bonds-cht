// Package sound contains the reference PC sound IDs and the extracted WAV
// asset mapping. Playback stays in the Ebiten adapter so game state remains
// deterministic and testable without an audio device.
package sound

// ID is the byte-sized sound selector used by the reference seg044.PlaySound.
type ID int8

const (
	Stop ID = 0
	NoOp ID = 1

	Missile  ID = 2
	MagicHit ID = 3
	Death    ID = 5
	Sound5   ID = 6
	Hit      ID = 7
	Miss     ID = 9
	Step     ID = 10
	Sound10  ID = 11
	Start    ID = 13
)

// AssetName returns the extracted WAV corresponding to the reference sound
// selector. Missing reference selectors intentionally return false: the
// original engine also leaves several sound IDs without a sample.
func AssetName(id ID) (string, bool) {
	assets := map[ID]string{
		Missile:  "missle.wav",
		MagicHit: "magic_hit.wav",
		Death:    "death.wav",
		Sound5:   "sound_5.wav",
		Hit:      "hit.wav",
		Miss:     "miss.wav",
		Step:     "step.wav",
		Sound10:  "sound_10.wav",
		Start:    "start_sound.wav",
	}
	name, ok := assets[id]
	return name, ok
}
