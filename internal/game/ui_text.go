package game

import (
	"fmt"
	"strings"
)

// PlayerUILabel identifies static player-facing chrome without exposing
// locale keys to Ebiten.
type PlayerUILabel uint8

const (
	PlayerUILabelSelectHelp PlayerUILabel = iota
	PlayerUILabelSaveLoadHelp
	PlayerUILabelContinueHelp
	PlayerUILabelShadowdaleMapTitle
	PlayerUILabelMapControls
	PlayerUILabelOverlandTitle
	PlayerUILabelOverlandControls
	PlayerUILabelSceneCharacterMissing
	PlayerUILabelBigPictureContinue
	PlayerUILabelBigPictureMissing
	PlayerUILabelPictureMissing
	PlayerUILabelCharacterNameHeader
	PlayerUILabelCombatViewTitle
	PlayerUILabelCombatViewReturn
	PlayerUILabelFallen
	PlayerUILabelDungeonDoorHelp
	PlayerUILabelDungeonExploreHelp
	PlayerUILabelJournalImageHint
	PlayerUILabelJournalImageControls
	PlayerUILabelJournalImageMissing
)

func (s *State) PlayerUILabel(label PlayerUILabel) string {
	key := "player_ui_unknown"
	switch label {
	case PlayerUILabelSelectHelp:
		key = "player_select_help"
	case PlayerUILabelSaveLoadHelp:
		key = "player_save_load_help"
	case PlayerUILabelContinueHelp:
		key = "player_continue_help"
	case PlayerUILabelShadowdaleMapTitle:
		key = "shadowdale_map_title"
	case PlayerUILabelMapControls:
		key = "area_navigation_help"
	case PlayerUILabelOverlandTitle:
		key = "overland_map_title"
	case PlayerUILabelOverlandControls:
		key = "overland_map_help"
	case PlayerUILabelSceneCharacterMissing:
		key = "scene_character_missing"
	case PlayerUILabelBigPictureContinue:
		key = "big_picture_continue"
	case PlayerUILabelBigPictureMissing:
		key = "big_picture_missing"
	case PlayerUILabelPictureMissing:
		key = "event_picture_missing"
	case PlayerUILabelCharacterNameHeader:
		key = "character_name_header"
	case PlayerUILabelCombatViewTitle:
		key = "combat_view_title"
	case PlayerUILabelCombatViewReturn:
		key = "combat_view_return"
	case PlayerUILabelFallen:
		key = "combat_fallen"
	case PlayerUILabelDungeonDoorHelp:
		key = "dungeon_door_help"
	case PlayerUILabelDungeonExploreHelp:
		key = "dungeon_explore_help"
	case PlayerUILabelJournalImageHint:
		key = "journal_image_hint"
	case PlayerUILabelJournalImageControls:
		key = "journal_image_controls"
	case PlayerUILabelJournalImageMissing:
		key = "journal_image_missing"
	}
	return s.catalog.Text(key, key)
}

func (s *State) AreaMapPositionText() string {
	return fmt.Sprintf(s.catalog.Text("area_map_position", "(%d,%d)"), s.MapX, s.MapY)
}

func (s *State) OverlandCurrentLocationText(name string) string {
	return fmt.Sprintf(s.catalog.Text("overland_current_location", "%s"), name)
}

// LocaleLanguage lets game-pack overlays follow the State-owned catalog
// instead of hard-coding a locale in the renderer.
func (s *State) LocaleLanguage() string { return s.catalog.Language }

// DungeonMessage identifies the player-visible outcome of a typed dungeon
// operation. Door flags, spell consumption, and unlock mutation remain in the
// existing rules; this enum controls presentation only.
type DungeonMessage uint8

const (
	DungeonMessageLockedPrompt DungeonMessage = iota
	DungeonMessagePickUnavailable
	DungeonMessagePickSucceeded
	DungeonMessagePickFailed
	DungeonMessageKnockSurfaceUnavailable
	DungeonMessageKnockSucceeded
	DungeonMessageBashUnavailable
	DungeonMessageBashSucceeded
	DungeonMessageBashFailed
)

func (s *State) DungeonMessageText(message DungeonMessage) string {
	key := "dungeon_message_unknown"
	switch message {
	case DungeonMessageLockedPrompt:
		key = "dungeon_door_locked_prompt"
	case DungeonMessagePickUnavailable:
		key = "dungeon_pick_unavailable"
	case DungeonMessagePickSucceeded:
		key = "dungeon_pick_succeeded"
	case DungeonMessagePickFailed:
		key = "dungeon_pick_failed"
	case DungeonMessageKnockSurfaceUnavailable:
		key = "dungeon_knock_surface_unavailable"
	case DungeonMessageKnockSucceeded:
		key = "dungeon_knock_succeeded"
	case DungeonMessageBashUnavailable:
		key = "dungeon_bash_unavailable"
	case DungeonMessageBashSucceeded:
		key = "dungeon_bash_succeeded"
	case DungeonMessageBashFailed:
		key = "dungeon_bash_failed"
	}
	return s.catalog.Text(key, key)
}

func (s *State) DungeonLifecycleErrorText(err error) string {
	return fmt.Sprintf(s.catalog.Text("dungeon_lifecycle_failed", "dungeon_lifecycle_failed: %s"), err)
}

func (s *State) DungeonKnockUnavailableText(spellID uint8) string {
	return fmt.Sprintf(s.catalog.Text("dungeon_knock_spell_unavailable", "dungeon_knock_spell_unavailable: 0x%02X"), spellID)
}

// PreviewLabel identifies research and asset-diagnostic chrome. These labels
// remain localized even though they are not part of the original game UI.
type PreviewLabel uint8

const (
	PreviewLabelAreaGeoMissing PreviewLabel = iota
	PreviewLabelAreaSymbolsMissing
	PreviewLabelAreaTitle
	PreviewLabelAreaOriginalViewport
	PreviewLabelAreaPartyMarker
	PreviewLabelAreaDescription
	PreviewLabelAreaReturn
	PreviewLabelTileTitle
	PreviewLabelGeoMissing
	PreviewLabelGeoCursorHelp
	PreviewLabelDungeonTitle
	PreviewLabelDungeonFloorMissing
	PreviewLabelDungeonControls
)

func (s *State) PreviewLabelText(label PreviewLabel) string {
	key := "preview_unknown"
	switch label {
	case PreviewLabelAreaGeoMissing:
		key = "preview_area_geo_missing"
	case PreviewLabelAreaSymbolsMissing:
		key = "preview_area_symbols_missing"
	case PreviewLabelAreaTitle:
		key = "preview_area_title"
	case PreviewLabelAreaOriginalViewport:
		key = "preview_area_original_viewport"
	case PreviewLabelAreaPartyMarker:
		key = "preview_area_party_marker"
	case PreviewLabelAreaDescription:
		key = "preview_area_description"
	case PreviewLabelAreaReturn:
		key = "preview_area_return"
	case PreviewLabelTileTitle:
		key = "preview_tile_title"
	case PreviewLabelGeoMissing:
		key = "preview_geo_missing"
	case PreviewLabelGeoCursorHelp:
		key = "preview_geo_cursor_help"
	case PreviewLabelDungeonTitle:
		key = "preview_dungeon_title"
	case PreviewLabelDungeonFloorMissing:
		key = "preview_dungeon_floor_missing"
	case PreviewLabelDungeonControls:
		key = "preview_dungeon_controls"
	}
	return s.catalog.Text(key, key)
}

func (s *State) PreviewPiecesFailedText(err error) string {
	return fmt.Sprintf(s.catalog.Text("preview_pieces_failed", "preview_pieces_failed: %s"), err)
}

func (s *State) PreviewPiecesLoadedText(first, second, third uint16) string {
	return fmt.Sprintf(s.catalog.Text("preview_pieces_loaded", "preview_pieces_loaded: %d/%d/%d"), first, second, third)
}

func (s *State) PreviewGeoMapMissingText(set, block uint8) string {
	return fmt.Sprintf(s.catalog.Text("preview_geo_map_missing", "preview_geo_map_missing: %d/%02X"), set, block)
}

func (s *State) PreviewAreaSourceText(set, block uint8) string {
	return fmt.Sprintf(s.catalog.Text("preview_area_source", "GEO%d/%02X"), set, block)
}

func (s *State) PreviewAreaPositionText(x, y int) string {
	return fmt.Sprintf(s.catalog.Text("preview_area_position", "(%d,%d)"), x, y)
}

func (s *State) PreviewAreaDirectionText(direction string) string {
	return fmt.Sprintf(s.catalog.Text("preview_area_direction", "%s"), direction)
}

func (s *State) PreviewGeoTitleText(label string) string {
	return fmt.Sprintf(s.catalog.Text("preview_geo_title", "%s"), label)
}

func (s *State) PreviewDungeonStatusText(x, y int, direction string, set, block uint8) string {
	return fmt.Sprintf(s.catalog.Text("preview_dungeon_status", "(%d,%d) %s GEO%d/%02X"), x, y, direction, set, block)
}

func (s *State) PreviewDungeonWallText(wall, roof uint8) string {
	return fmt.Sprintf(s.catalog.Text("preview_dungeon_wall", "%02X/%02X"), wall, roof)
}

func (s *State) PreviewDungeonDoorStatusText(flags uint8) string {
	return fmt.Sprintf(s.catalog.Text("preview_dungeon_door_status", "%d"), flags)
}

func (s *State) PreviewDungeonDoorHelpText(pick, knock bool) string {
	key := "preview_dungeon_door_bash"
	if pick && knock {
		key = "preview_dungeon_door_pick_knock"
	} else if pick {
		key = "preview_dungeon_door_pick"
	} else if knock {
		key = "preview_dungeon_door_knock"
	}
	return s.catalog.Text(key, key)
}

func (s *State) PreviewDungeonPiecesText(first, second, third uint16) string {
	return fmt.Sprintf(s.catalog.Text("preview_dungeon_pieces", "%d/%d/%d"), first, second, third)
}

func (s *State) OverlandDateText() string {
	clock := s.GameTimeDisplay()
	return fmt.Sprintf(s.catalog.Text("overland_date", "%d/%d/%d"), clock.Day, clock.Month, clock.Year)
}

type DemoFighterName uint8

const (
	DemoNameArcherErin DemoFighterName = iota
	DemoNameOrc
	DemoNameFighterErin
	DemoNameZhentMage
	DemoNameMageErin
	DemoNameFighterBran
	DemoNameOrcCaptain
	DemoNamePartyFighter
	DemoNamePartyRanger
	DemoNamePartyCleric
	DemoNamePartyWizard
)

func (s *State) DemoFighterName(name DemoFighterName) string {
	key := "demo_name_unknown"
	switch name {
	case DemoNameArcherErin:
		key = "demo_name_archer_erin"
	case DemoNameOrc:
		key = "demo_name_orc"
	case DemoNameFighterErin:
		key = "demo_name_fighter_erin"
	case DemoNameZhentMage:
		key = "demo_name_zhent_mage"
	case DemoNameMageErin:
		key = "demo_name_mage_erin"
	case DemoNameFighterBran:
		key = "demo_name_fighter_bran"
	case DemoNameOrcCaptain:
		key = "demo_name_orc_captain"
	case DemoNamePartyFighter:
		key = "demo_name_party_fighter"
	case DemoNamePartyRanger:
		key = "demo_name_party_ranger"
	case DemoNamePartyCleric:
		key = "demo_name_party_cleric"
	case DemoNamePartyWizard:
		key = "demo_name_party_wizard"
	}
	return s.catalog.Text(key, key)
}

func (s *State) PrepareWorldMapPreview() {
	s.Mode = ModeWilderness
	s.Area.CurrentCity = 4
	s.Location = LocationStandingStone
	s.LocationName = s.catalog.Text("standing_stone", "standing_stone")
	s.Choices = []string{
		s.localizeOption("ENTER CITY"),
		s.localizeOption("JOURNEY ON"),
		s.localizeOption("CAMP"),
	}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	s.Prompt = s.catalog.Text("press_button", "press_button")
}

// FileOperation identifies a player-facing persistence operation without
// exposing locale keys to the platform adapter.
type FileOperation uint8

const (
	FileOperationSave FileOperation = iota
	FileOperationLoad
	FileOperationAudioRestore
)

// FileOperationResult distinguishes successful persistence from an error.
type FileOperationResult uint8

const (
	FileOperationSucceeded FileOperationResult = iota
	FileOperationFailed
)

// FileOperationMessage resolves persistence feedback through the State-owned
// catalog. detail is a path for success and an error description for failure.
func (s *State) FileOperationMessage(operation FileOperation, result FileOperationResult, detail string) string {
	key := "file_operation_unknown"
	switch operation {
	case FileOperationSave:
		if result == FileOperationSucceeded {
			key = "save_succeeded"
		} else {
			key = "save_failed"
		}
	case FileOperationLoad:
		if result == FileOperationSucceeded {
			key = "load_succeeded"
		} else {
			key = "load_failed"
		}
	case FileOperationAudioRestore:
		key = "audio_restore_failed"
	}
	text := s.catalog.Text(key, key+": %s")
	// ⚠ **失敗訊息不要把 Go 的錯誤字串端到玩家面前。** `open : no such file or
	// directory` 是給開發者看的英文，而它會整句出現在畫面上。譯文裡沒有 `%s`
	// 就代表這一條不吃細節——照樣 `Sprintf` 會補上 `%!(EXTRA string=…)`。
	if !strings.Contains(text, "%s") {
		return text
	}
	return fmt.Sprintf(text, detail)
}

// RenameInputText and RenameInputHelp form the renderer-neutral ALTER rename
// editor contract. The frontend supplies geometry, font, and color only.
func (s *State) RenameInputText() string {
	return fmt.Sprintf(s.catalog.Text("rename_input_value", "%s_"), s.renameName)
}

func (s *State) RenameInputHelp() string {
	return s.catalog.Text("rename_input_help", "rename_input_help")
}

// ECLStringInputText and ECLStringInputHelp expose the active VM input
// boundary without embedding its Traditional Chinese labels in Ebiten.
func (s *State) ECLStringInputText() string {
	return fmt.Sprintf(s.catalog.Text("ecl_string_input_value", "%s_"), s.ECLStringValue())
}

func (s *State) ECLStringInputHelp() string {
	return fmt.Sprintf(s.catalog.Text("ecl_string_input_help", "ecl_string_input_help"), s.ECLStringMaxLength())
}
