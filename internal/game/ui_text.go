package game

import "fmt"

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
