package game

import "testing"

func TestParlayRuntimeKeysHaveFormalTranslations(t *testing.T) {
	catalog := combatVisualCatalog(t)
	for _, key := range []string{
		"parlay_menu_prompt", "parlay_haughty", "parlay_sly", "parlay_meek",
		"parlay_nice", "parlay_abusive", "encounter_parlay_done",
	} {
		if got := catalog.Text(key, key); got == "" || got == key {
			t.Errorf("formal catalog does not translate %q", key)
		}
	}
}
