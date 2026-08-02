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

func TestAdventureChromeContractUsesFormalCatalog(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	labels := []struct {
		label PlayerUILabel
		key   string
	}{
		{PlayerUILabelSelectHelp, "player_select_help"},
		{PlayerUILabelSaveLoadHelp, "player_save_load_help"},
		{PlayerUILabelContinueHelp, "player_continue_help"},
		{PlayerUILabelShadowdaleMapTitle, "shadowdale_map_title"},
		{PlayerUILabelMapControls, "area_navigation_help"},
		{PlayerUILabelOverlandTitle, "overland_map_title"},
		{PlayerUILabelOverlandControls, "overland_map_help"},
		{PlayerUILabelSceneCharacterMissing, "scene_character_missing"},
		{PlayerUILabelBigPictureContinue, "big_picture_continue"},
		{PlayerUILabelBigPictureMissing, "big_picture_missing"},
		{PlayerUILabelPictureMissing, "event_picture_missing"},
		{PlayerUILabelCharacterNameHeader, "character_name_header"},
		{PlayerUILabelCombatViewTitle, "combat_view_title"},
		{PlayerUILabelCombatViewReturn, "combat_view_return"},
		{PlayerUILabelFallen, "combat_fallen"},
		{PlayerUILabelDungeonDoorHelp, "dungeon_door_help"},
		{PlayerUILabelDungeonExploreHelp, "dungeon_explore_help"},
	}
	for _, test := range labels {
		if got, want := state.PlayerUILabel(test.label), catalog.Text(test.key, ""); got != want {
			t.Errorf("label %d=%q want=%q", test.label, got, want)
		}
	}

	state.MapX, state.MapY = 3, 11
	if got, want := state.AreaMapPositionText(), fmt.Sprintf(catalog.Text("area_map_position", ""), 3, 11); got != want {
		t.Fatalf("map position=%q want=%q", got, want)
	}
	if got, want := state.OverlandCurrentLocationText("LOCATION"), fmt.Sprintf(catalog.Text("overland_current_location", ""), "LOCATION"); got != want {
		t.Fatalf("current location=%q want=%q", got, want)
	}
	if got := state.LocaleLanguage(); got != catalog.Language {
		t.Fatalf("locale language=%q want=%q", got, catalog.Language)
	}
}

func TestDungeonOperationMessagesUseTypedFormalCatalog(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	messages := []struct {
		message DungeonMessage
		key     string
	}{
		{DungeonMessageLockedPrompt, "dungeon_door_locked_prompt"},
		{DungeonMessagePickUnavailable, "dungeon_pick_unavailable"},
		{DungeonMessagePickSucceeded, "dungeon_pick_succeeded"},
		{DungeonMessagePickFailed, "dungeon_pick_failed"},
		{DungeonMessageKnockSurfaceUnavailable, "dungeon_knock_surface_unavailable"},
		{DungeonMessageKnockSucceeded, "dungeon_knock_succeeded"},
		{DungeonMessageBashUnavailable, "dungeon_bash_unavailable"},
		{DungeonMessageBashSucceeded, "dungeon_bash_succeeded"},
		{DungeonMessageBashFailed, "dungeon_bash_failed"},
	}
	for _, test := range messages {
		if got, want := state.DungeonMessageText(test.message), catalog.Text(test.key, ""); got != want {
			t.Errorf("message %d=%q want=%q", test.message, got, want)
		}
	}

	err := fmt.Errorf("BROKEN EVENT")
	if got, want := state.DungeonLifecycleErrorText(err), fmt.Sprintf(catalog.Text("dungeon_lifecycle_failed", ""), err); got != want {
		t.Fatalf("lifecycle error=%q want=%q", got, want)
	}
	if got, want := state.DungeonKnockUnavailableText(0x29), fmt.Sprintf(catalog.Text("dungeon_knock_spell_unavailable", ""), 0x29); got != want {
		t.Fatalf("Knock unavailable=%q want=%q", got, want)
	}
}
