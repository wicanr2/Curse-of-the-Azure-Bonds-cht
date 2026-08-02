package game

import (
	"fmt"
	"testing"
)

func TestPlayerFileAndTextInputContractUsesFormalCatalog(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)

	fileCases := []struct {
		name      string
		operation FileOperation
		result    FileOperationResult
		key       string
		detail    string
	}{
		{"save success", FileOperationSave, FileOperationSucceeded, "save_succeeded", "SAVE.PTY"},
		{"save failure", FileOperationSave, FileOperationFailed, "save_failed", "DISK FULL"},
		{"load success", FileOperationLoad, FileOperationSucceeded, "load_succeeded", "SAVE.PTY"},
		{"load failure", FileOperationLoad, FileOperationFailed, "load_failed", "BAD DATA"},
		{"audio restore failure", FileOperationAudioRestore, FileOperationFailed, "audio_restore_failed", "BAD SNAPSHOT"},
	}
	for _, test := range fileCases {
		t.Run(test.name, func(t *testing.T) {
			got := state.FileOperationMessage(test.operation, test.result, test.detail)
			want := fmt.Sprintf(catalog.Text(test.key, test.key), test.detail)
			if got != want {
				t.Fatalf("message=%q want=%q", got, want)
			}
		})
	}

	state.renameName = "RENAME"
	if got, want := state.RenameInputText(), fmt.Sprintf(catalog.Text("rename_input_value", ""), "RENAME"); got != want {
		t.Fatalf("rename input=%q want=%q", got, want)
	}
	if got, want := state.RenameInputHelp(), catalog.Text("rename_input_help", ""); got != want {
		t.Fatalf("rename help=%q want=%q", got, want)
	}

	state.eclStringEditing = true
	state.eclStringMaxLength = 8
	state.eclStringValue = "ANSWER"
	if got, want := state.ECLStringInputText(), fmt.Sprintf(catalog.Text("ecl_string_input_value", ""), "ANSWER"); got != want {
		t.Fatalf("ECL input=%q want=%q", got, want)
	}
	if got, want := state.ECLStringInputHelp(), fmt.Sprintf(catalog.Text("ecl_string_input_help", ""), 8); got != want {
		t.Fatalf("ECL help=%q want=%q", got, want)
	}
}
