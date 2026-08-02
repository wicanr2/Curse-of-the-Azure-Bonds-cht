package save

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
)

// CurrentGameVersion is the version of the remake's resumable game save.
const CurrentGameVersion = 8

type MusicSnapshot struct {
	TrackID string                            `json:"track_id"`
	Stream  *pc98music.TrackPCMStreamSnapshot `json:"stream,omitempty"`
}

// CombatSnapshot preserves the game-adapter side of an active Battle. Map
// callbacks and decoded source assets are intentionally excluded and are
// reconstructed by the frontend before the next action is resolved.
type CombatSnapshot struct {
	Battle             combat.BattleSnapshot `json:"battle"`
	Turns              []combat.Turn         `json:"turns,omitempty"`
	TurnIndex          int                   `json:"turn_index"`
	DelayedTurns       map[int]bool          `json:"delayed_turns,omitempty"`
	TargetIndex        int                   `json:"target_index"`
	CastingSpell       uint8                 `json:"casting_spell"`
	CastingClass       uint8                 `json:"casting_class"`
	CastingClassSet    bool                  `json:"casting_class_set"`
	SpellTargetIndex   int                   `json:"spell_target_index"`
	SpellTargetPoint   combat.TilePoint      `json:"spell_target_point"`
	SpellTargetsPoint  bool                  `json:"spell_targets_point"`
	MoveMode           bool                  `json:"move_mode"`
	MoveRemaining      int                   `json:"move_remaining"`
	Speed              uint8                 `json:"speed"`
	QuickMagic         bool                  `json:"quick_magic"`
	ReferenceCoords    bool                  `json:"reference_coordinates"`
	View               bool                  `json:"view"`
	ViewFighterID      string                `json:"view_fighter_id,omitempty"`
	Message            string                `json:"message,omitempty"`
	ReturnMode         uint8                 `json:"return_mode"`
	VisualSerial       uint64                `json:"visual_serial"`
	VisualEnabled      bool                  `json:"visual_enabled"`
	Visual             *combat.VisualEvent   `json:"visual,omitempty"`
	VisualElapsedNanos int64                 `json:"visual_elapsed_nanos"`
	VisualTravelSent   bool                  `json:"visual_travel_sent"`
	VisualImpactSent   int                   `json:"visual_impact_sent"`
	VisualDeathSent    int                   `json:"visual_death_sent"`
	VisualAdvanceTurn  bool                  `json:"visual_advance_turn"`
}

// GameFile contains the party plus the platform-neutral adventure state that
// the remake can currently restore. Numeric mode/location values are kept
// here to avoid coupling the save package to the game UI package.
type GameFile struct {
	Version         int                  `json:"version"`
	Characters      party.Roster         `json:"characters"`
	Area            area.State           `json:"area"`
	Mode            uint8                `json:"mode"`
	Location        uint8                `json:"location"`
	MapX            int                  `json:"map_x"`
	MapY            int                  `json:"map_y"`
	DungeonX        int                  `json:"dungeon_x"`
	DungeonY        int                  `json:"dungeon_y"`
	DungeonDir      uint8                `json:"dungeon_direction"`
	DungeonWallType uint8                `json:"dungeon_wall_type"`
	DungeonWallRoof uint8                `json:"dungeon_wall_roof"`
	GameTime        [7]uint16            `json:"game_time"`
	GameAgeCycles   uint32               `json:"game_age_cycles"`
	ECLSession      *ecl.SessionSnapshot `json:"ecl_session,omitempty"`
	Combat          *CombatSnapshot      `json:"combat,omitempty"`
	Music           *MusicSnapshot       `json:"music,omitempty"`
}

func EncodeGame(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY int) ([]byte, error) {
	// seg001.Init initializes the original dungeon camera at (7, 0x0D).
	return EncodeGameWithDungeonState(roster, areaState, mode, location, mapX, mapY, 7, 13, 0, 0, 0)
}

// EncodeGameWithDungeon writes the current remake adventure state including
// the dungeon 3D position/direction. EncodeGame remains a legacy-signature
// convenience for callers that do not own dungeon state.
func EncodeGameWithDungeon(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection uint8) ([]byte, error) {
	return EncodeGameWithDungeonState(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, 0, 0)
}

// EncodeGameWithDungeonState also preserves the reference mapWallType and
// mapWallRoof cache values from the original five-byte map segment.
func EncodeGameWithDungeonState(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8) ([]byte, error) {
	return EncodeGameWithTime(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, [7]uint16{}, 0)
}

// EncodeGameWithTime serializes the resumable adventure state including the
// reference seven-slot clock. The older helpers remain source-compatible and
// encode a zero clock for callers that do not own time state.
func EncodeGameWithTime(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32) ([]byte, error) {
	return EncodeGameWithSession(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, nil)
}

// EncodeGameWithSession adds the mutable ECL continuation. The snapshot owns
// only runtime state and code-memory differences; original ECL bytes remain in
// the player-supplied game image.
func EncodeGameWithSession(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot) ([]byte, error) {
	return EncodeGameWithCombat(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, session, nil)
}

// EncodeGameWithCombat adds a bounded active-battle continuation to save v7.
func EncodeGameWithCombat(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot, activeCombat *CombatSnapshot) ([]byte, error) {
	return EncodeGameWithAudio(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, session, activeCombat, nil)
}

// EncodeGameWithAudio adds the stable track identity and optional exact PC-98
// synthesis continuation used by remake save v8.
func EncodeGameWithAudio(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot, activeCombat *CombatSnapshot, music *MusicSnapshot) ([]byte, error) {
	if err := roster.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(GameFile{
		Version: CurrentGameVersion, Characters: roster, Area: areaState,
		Mode: mode, Location: location, MapX: mapX, MapY: mapY,
		DungeonX: dungeonX, DungeonY: dungeonY, DungeonDir: dungeonDirection,
		DungeonWallType: dungeonWallType, DungeonWallRoof: dungeonWallRoof,
		GameTime: gameTime, GameAgeCycles: gameAgeCycles, ECLSession: session,
		Combat: activeCombat, Music: music,
	}, "", "  ")
}

func DecodeGame(data []byte) (GameFile, error) {
	var file GameFile
	if err := json.Unmarshal(data, &file); err != nil {
		return GameFile{}, fmt.Errorf("game save JSON: %w", err)
	}
	// Version 1 was a party-only save. Accept it and use safe defaults for
	// fields introduced by the resumable game format.
	if file.Version < 1 || file.Version > CurrentGameVersion {
		return GameFile{}, fmt.Errorf("unsupported game save version %d", file.Version)
	}
	if file.Combat != nil && file.Version < 7 {
		return GameFile{}, fmt.Errorf("game save version %d cannot contain active combat", file.Version)
	}
	if file.Music != nil {
		if file.Version < 8 {
			return GameFile{}, fmt.Errorf("game save version %d cannot contain music continuation", file.Version)
		}
		if file.Music.TrackID == "" {
			return GameFile{}, fmt.Errorf("game save music track ID is empty")
		}
		if file.Music.Stream != nil {
			if err := file.Music.Stream.ValidatePersistent(); err != nil {
				return GameFile{}, fmt.Errorf("game save music continuation: %w", err)
			}
		}
	}
	if err := file.Characters.Validate(); err != nil {
		return GameFile{}, err
	}
	return file, nil
}
