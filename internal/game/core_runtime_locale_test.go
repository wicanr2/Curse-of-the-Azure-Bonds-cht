package game

import "testing"

func TestCoreRuntimeKeysHaveFormalTranslations(t *testing.T) {
	catalog := combatVisualCatalog(t)
	for _, key := range []string{
		"combat_missile_adjacent_error", "combat_action_error", "selected_character",
		"select_character", "event_picture", "combat_started", "character_hp_choice",
		"inn_restored", "spell_unknown", "ecl_string_input_prompt", "game_time",
	} {
		if got := catalog.Text(key, key); got == "" || got == key {
			t.Errorf("formal catalog does not translate %q", key)
		}
	}
}
