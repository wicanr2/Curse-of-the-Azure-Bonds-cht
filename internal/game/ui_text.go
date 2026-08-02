package game

import "fmt"

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
	return fmt.Sprintf(s.catalog.Text(key, key+": %s"), detail)
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
