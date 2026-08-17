package save

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
)

// CurrentGameVersion is the version of the remake's resumable game save.
const CurrentGameVersion = 12

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
	Version            int                  `json:"version"`
	Characters         party.Roster         `json:"characters"`
	Area               area.State           `json:"area"`
	Mode               uint8                `json:"mode"`
	Location           uint8                `json:"location"`
	MapX               int                  `json:"map_x"`
	MapY               int                  `json:"map_y"`
	DungeonX           int                  `json:"dungeon_x"`
	DungeonY           int                  `json:"dungeon_y"`
	DungeonDir         uint8                `json:"dungeon_direction"`
	DungeonWallType    uint8                `json:"dungeon_wall_type"`
	DungeonWallRoof    uint8                `json:"dungeon_wall_roof"`
	GameTime           [7]uint16            `json:"game_time"`
	GameAgeCycles      uint32               `json:"game_age_cycles"`
	ECLSession         *ecl.SessionSnapshot `json:"ecl_session,omitempty"`
	Combat             *CombatSnapshot      `json:"combat,omitempty"`
	Music              *MusicSnapshot       `json:"music,omitempty"`
	OneShotAudio       *audiostate.Snapshot `json:"one_shot_audio,omitempty"`
	JournalMessageIDs  []string             `json:"journal_message_ids,omitempty"`
	DungeonSearch      bool                 `json:"dungeon_search,omitempty"`
	DungeonSearchEdges []string             `json:"dungeon_search_edges,omitempty"`
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
	return EncodeGameWithAudioState(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, session, activeCombat, music, nil)
}

// EncodeGameWithAudioState adds the bounded active one-shot continuation used
// by remake save v9. The older signature remains available to data-neutral
// callers and intentionally writes no platform playback snapshot.
func EncodeGameWithAudioState(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot, activeCombat *CombatSnapshot, music *MusicSnapshot, oneShots *audiostate.Snapshot) ([]byte, error) {
	return EncodeGameWithJournalState(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, session, activeCombat, music, oneShots, nil)
}

// EncodeGameWithJournalState stores stable game-pack message IDs rather than
// localized page text, allowing translation updates and locale changes after
// a save was written.
func EncodeGameWithJournalState(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot, activeCombat *CombatSnapshot, music *MusicSnapshot, oneShots *audiostate.Snapshot, journalMessageIDs []string) ([]byte, error) {
	return EncodeGameWithAdventureState(roster, areaState, mode, location, mapX, mapY, dungeonX, dungeonY, dungeonDirection, dungeonWallType, dungeonWallRoof, gameTime, gameAgeCycles, session, activeCombat, music, oneShots, journalMessageIDs, false, nil)
}

// EncodeGameWithAdventureState adds remake-owned dungeon interaction state to
// the stable-ID save. The edge IDs come from the loaded game pack; this save
// layer does not know whether an edge is a secret passage, a door, or another
// title-owned map mutation.
// EncodeGameFile 直接序列化一份組好的存檔。
//
// ★ 它存在的理由是**測試要能組出一份「每個欄位都有值」的存檔**：
// `EncodeGameWithAdventureState` 收 20 個位置參數，測試照著填一次等於把同一個
// 容易出錯的地方再抄一遍。正式流程仍走那一支（它帶著驗證），這一支只補
// `Version` 並做同樣的驗證。
func EncodeGameFile(file GameFile) ([]byte, error) {
	file.Version = CurrentGameVersion
	if err := file.Characters.Validate(); err != nil {
		return nil, err
	}
	if err := validateJournalMessageIDs(file.JournalMessageIDs); err != nil {
		return nil, err
	}
	if err := validateDungeonSearchEdges(file.DungeonSearchEdges); err != nil {
		return nil, err
	}
	return json.MarshalIndent(file, "", "  ")
}

func EncodeGameWithAdventureState(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY, dungeonX, dungeonY int, dungeonDirection, dungeonWallType, dungeonWallRoof uint8, gameTime [7]uint16, gameAgeCycles uint32, session *ecl.SessionSnapshot, activeCombat *CombatSnapshot, music *MusicSnapshot, oneShots *audiostate.Snapshot, journalMessageIDs []string, dungeonSearch bool, dungeonSearchEdges []string) ([]byte, error) {
	if err := roster.Validate(); err != nil {
		return nil, err
	}
	if err := validateJournalMessageIDs(journalMessageIDs); err != nil {
		return nil, err
	}
	if err := validateDungeonSearchEdges(dungeonSearchEdges); err != nil {
		return nil, err
	}
	if oneShots != nil {
		if err := oneShots.Validate(); err != nil {
			return nil, fmt.Errorf("one-shot audio snapshot: %w", err)
		}
		copy := audiostate.Clone(*oneShots)
		oneShots = &copy
	}
	return json.MarshalIndent(GameFile{
		Version: CurrentGameVersion, Characters: roster, Area: areaState,
		Mode: mode, Location: location, MapX: mapX, MapY: mapY,
		DungeonX: dungeonX, DungeonY: dungeonY, DungeonDir: dungeonDirection,
		DungeonWallType: dungeonWallType, DungeonWallRoof: dungeonWallRoof,
		GameTime: gameTime, GameAgeCycles: gameAgeCycles, ECLSession: session,
		Combat: activeCombat, Music: music, OneShotAudio: oneShots,
		JournalMessageIDs:  append([]string(nil), journalMessageIDs...),
		DungeonSearch:      dungeonSearch,
		DungeonSearchEdges: append([]string(nil), dungeonSearchEdges...),
	}, "", "  ")
}

func validateJournalMessageIDs(messageIDs []string) error {
	seen := make(map[string]bool, len(messageIDs))
	for index, messageID := range messageIDs {
		if messageID == "" {
			return fmt.Errorf("journal message ID %d is empty", index)
		}
		if seen[messageID] {
			return fmt.Errorf("duplicate journal message ID %q", messageID)
		}
		seen[messageID] = true
	}
	return nil
}

func validateDungeonSearchEdges(edgeIDs []string) error {
	seen := make(map[string]bool, len(edgeIDs))
	for index, edgeID := range edgeIDs {
		if strings.TrimSpace(edgeID) == "" {
			return fmt.Errorf("dungeon search edge ID %d is empty", index)
		}
		if seen[edgeID] {
			return fmt.Errorf("duplicate dungeon search edge ID %q", edgeID)
		}
		seen[edgeID] = true
	}
	return nil
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
	if file.OneShotAudio != nil {
		if file.Version < 9 {
			return GameFile{}, fmt.Errorf("game save version %d cannot contain one-shot audio continuation", file.Version)
		}
		if err := file.OneShotAudio.Validate(); err != nil {
			return GameFile{}, fmt.Errorf("game save one-shot audio continuation: %w", err)
		}
	}
	if len(file.JournalMessageIDs) > 0 {
		if file.Version < 10 {
			return GameFile{}, fmt.Errorf("game save version %d cannot contain journal message IDs", file.Version)
		}
		if err := validateJournalMessageIDs(file.JournalMessageIDs); err != nil {
			return GameFile{}, fmt.Errorf("game save journal: %w", err)
		}
	}
	if len(file.DungeonSearchEdges) > 0 {
		if file.Version < 12 {
			return GameFile{}, fmt.Errorf("game save version %d cannot contain dungeon search edges", file.Version)
		}
		if err := validateDungeonSearchEdges(file.DungeonSearchEdges); err != nil {
			return GameFile{}, fmt.Errorf("game save dungeon search: %w", err)
		}
	}
	if file.DungeonSearch && file.Version < 12 {
		return GameFile{}, fmt.Errorf("game save version %d cannot contain dungeon search state", file.Version)
	}
	if err := file.Characters.Validate(); err != nil {
		return GameFile{}, err
	}
	return file, nil
}
