package game

import (
	"fmt"
	"testing"
)

func TestCreationAndWorldRuntimeKeysHaveFormalTranslations(t *testing.T) {
	catalog := combatVisualCatalog(t)
	keys := []string{
		"party_ready", "enter_city", "journey_on", "camp", "journal_title", "journal_close",
		"press_button", "encounter_flee_done", "pics_monsters_off_message", "hap_dungeon_prompt",
		"npc_akabar", "npc_alias", "npc_dragonbait", "world_travel_map_prompt",
		"shadowdale_map_prompt", "tilverton", "dungeon_prompt", "inn", "leave",
		"place_prompt", "standing_stone", "essembra", "hap",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == key || got == "" {
			t.Errorf("formal catalog does not translate %q", key)
		}
	}

	state := NewState(catalog)
	state.LocationName = "LOCATION"
	if got, want := state.placePrompt(), fmt.Sprintf(catalog.Text("place_prompt", ""), "LOCATION"); got != want {
		t.Fatalf("place prompt=%q want=%q", got, want)
	}

	worldLocations := []struct {
		value uint16
		key   string
	}{
		{0, "tilverton"}, {4, "standing_stone"}, {8, "essembra"}, {9, "hap"},
	}
	for _, test := range worldLocations {
		state.setWorldLocation(test.value)
		if got, want := state.LocationName, catalog.Text(test.key, ""); got != want {
			t.Errorf("world location %d=%q want=%q", test.value, got, want)
		}
	}

	state.restoreWildernessMenu()
	if state.Prompt != catalog.Text("press_button", "") || len(state.Choices) != 3 ||
		state.Choices[0] != catalog.Text("enter_city", "") ||
		state.Choices[1] != catalog.Text("journey_on", "") ||
		state.Choices[2] != catalog.Text("camp", "") {
		t.Fatalf("wilderness menu prompt=%q choices=%#v", state.Prompt, state.Choices)
	}
}
