package game

import (
	"fmt"
	"strings"
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
			text := catalog.Text(test.key, test.key)
			// ⚠ **失敗訊息不代入細節。** `detail` 是 Go 的錯誤字串
			// （`open : no such file or directory`），端到畫面上就是一句英文。
			// 譯文裡沒有 `%s` 就整句照用；照樣 `Sprintf` 會補上
			// `%!(EXTRA string=…)`，那反而更難看。
			want := text
			if strings.Contains(text, "%s") {
				want = fmt.Sprintf(text, test.detail)
			}
			if got != want {
				t.Fatalf("message=%q want=%q", got, want)
			}
			if test.result == FileOperationFailed && strings.Contains(got, test.detail) {
				t.Fatalf("失敗訊息不該帶著細節 %q：%q", test.detail, got)
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

func TestPreviewDiagnosticContractUsesFormalCatalog(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	labels := []struct {
		label PreviewLabel
		key   string
	}{
		{PreviewLabelAreaGeoMissing, "preview_area_geo_missing"},
		{PreviewLabelAreaSymbolsMissing, "preview_area_symbols_missing"},
		{PreviewLabelAreaTitle, "preview_area_title"},
		{PreviewLabelAreaOriginalViewport, "preview_area_original_viewport"},
		{PreviewLabelAreaPartyMarker, "preview_area_party_marker"},
		{PreviewLabelAreaDescription, "preview_area_description"},
		{PreviewLabelAreaReturn, "preview_area_return"},
		{PreviewLabelTileTitle, "preview_tile_title"},
		{PreviewLabelGeoMissing, "preview_geo_missing"},
		{PreviewLabelGeoCursorHelp, "preview_geo_cursor_help"},
		{PreviewLabelDungeonTitle, "preview_dungeon_title"},
		{PreviewLabelDungeonFloorMissing, "preview_dungeon_floor_missing"},
		{PreviewLabelDungeonControls, "preview_dungeon_controls"},
	}
	for _, test := range labels {
		if got, want := state.PreviewLabelText(test.label), catalog.Text(test.key, ""); got != want {
			t.Errorf("preview label %d=%q want=%q", test.label, got, want)
		}
	}

	err := fmt.Errorf("MISSING ASSET")
	formats := []struct {
		name string
		got  string
		key  string
		args []any
	}{
		{"pieces failed", state.PreviewPiecesFailedText(err), "preview_pieces_failed", []any{err}},
		{"pieces loaded", state.PreviewPiecesLoadedText(1, 2, 3), "preview_pieces_loaded", []any{uint16(1), uint16(2), uint16(3)}},
		{"GEO missing", state.PreviewGeoMapMissingText(4, 0x2A), "preview_geo_map_missing", []any{uint8(4), uint8(0x2A)}},
		{"AREA source", state.PreviewAreaSourceText(4, 0x2A), "preview_area_source", []any{uint8(4), uint8(0x2A)}},
		{"AREA position", state.PreviewAreaPositionText(7, 9), "preview_area_position", []any{7, 9}},
		{"AREA direction", state.PreviewAreaDirectionText("NORTH"), "preview_area_direction", []any{"NORTH"}},
		{"GEO title", state.PreviewGeoTitleText("GEO4"), "preview_geo_title", []any{"GEO4"}},
		{"dungeon status", state.PreviewDungeonStatusText(7, 9, "NORTH", 4, 0x2A), "preview_dungeon_status", []any{7, 9, "NORTH", uint8(4), uint8(0x2A)}},
		{"dungeon wall", state.PreviewDungeonWallText(0x11, 0x82), "preview_dungeon_wall", []any{uint8(0x11), uint8(0x82)}},
		{"door status", state.PreviewDungeonDoorStatusText(3), "preview_dungeon_door_status", []any{uint8(3)}},
		{"dungeon pieces", state.PreviewDungeonPiecesText(1, 2, 3), "preview_dungeon_pieces", []any{uint16(1), uint16(2), uint16(3)}},
	}
	for _, test := range formats {
		if want := fmt.Sprintf(catalog.Text(test.key, ""), test.args...); test.got != want {
			t.Errorf("%s=%q want=%q", test.name, test.got, want)
		}
	}

	doorCases := []struct {
		pick, knock bool
		key         string
	}{
		{false, false, "preview_dungeon_door_bash"},
		{true, false, "preview_dungeon_door_pick"},
		{false, true, "preview_dungeon_door_knock"},
		{true, true, "preview_dungeon_door_pick_knock"},
	}
	for _, test := range doorCases {
		if got, want := state.PreviewDungeonDoorHelpText(test.pick, test.knock), catalog.Text(test.key, ""); got != want {
			t.Errorf("door pick=%v knock=%v got=%q want=%q", test.pick, test.knock, got, want)
		}
	}

	state.gameClock[4], state.gameClock[5], state.gameClock[6] = 12, 3, 1362
	if got, want := state.OverlandDateText(), fmt.Sprintf(catalog.Text("overland_date", ""), 12, 3, 1362); got != want {
		t.Fatalf("overland date=%q want=%q", got, want)
	}
}

func TestDemoNamesAndWorldMapPreviewUseFormalCatalog(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	names := []struct {
		name DemoFighterName
		key  string
	}{
		{DemoNameArcherErin, "demo_name_archer_erin"},
		{DemoNameOrc, "demo_name_orc"},
		{DemoNameFighterErin, "demo_name_fighter_erin"},
		{DemoNameZhentMage, "demo_name_zhent_mage"},
		{DemoNameMageErin, "demo_name_mage_erin"},
		{DemoNameFighterBran, "demo_name_fighter_bran"},
		{DemoNameOrcCaptain, "demo_name_orc_captain"},
		{DemoNamePartyFighter, "demo_name_party_fighter"},
		{DemoNamePartyRanger, "demo_name_party_ranger"},
		{DemoNamePartyCleric, "demo_name_party_cleric"},
		{DemoNamePartyWizard, "demo_name_party_wizard"},
	}
	for _, test := range names {
		if got, want := state.DemoFighterName(test.name), catalog.Text(test.key, ""); got != want {
			t.Errorf("demo name %d=%q want=%q", test.name, got, want)
		}
	}

	state.PrepareWorldMapPreview()
	if state.Mode != ModeWilderness || state.Area.CurrentCity != 4 || state.Location != LocationStandingStone ||
		state.LocationName != catalog.Text("standing_stone", "") ||
		len(state.Choices) != 3 || state.Choices[0] != catalog.Text("enter_city", "") ||
		state.Choices[1] != catalog.Text("journey_on", "") || state.Choices[2] != catalog.Text("camp", "") ||
		state.Prompt != catalog.Text("press_button", "") {
		t.Fatalf("world preview state=%+v", state)
	}
}
