package game

import "testing"

func TestTreasureRuntimeKeysHaveFormalTranslations(t *testing.T) {
	catalog := combatVisualCatalog(t)
	for _, key := range []string{
		"treasure_prompt", "treasure_exit", "treasure_ready", "treasure_take_prompt",
		"treasure_cancel", "treasure_taken", "treasure_skipped", "treasure_assets_pending",
	} {
		if got := catalog.Text(key, key); got == "" || got == key {
			t.Errorf("formal catalog does not translate %q", key)
		}
	}
}
