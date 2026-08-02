package game

import "testing"

func TestCombatRuntimeKeysHaveFormalTranslations(t *testing.T) {
	catalog := combatVisualCatalog(t)
	keys := []string{
		"combat_started", "combat_view_name", "combat_view_hp", "combat_view_ac",
		"combat_view_attack", "combat_moved", "combat_guard_reaction", "combat_free_attack",
		"combat_quick_magic_casting", "combat_lightning_bolt", "combat_lightning_bolt_protected",
		"combat_stinking_cloud", "combat_cloudkill", "combat_sleep", "combat_fireball",
		"combat_fireball_protected", "combat_protection_from_good", "combat_protection_from_evil",
		"combat_cause_light_wounds", "combat_curse_immune", "combat_curse", "combat_bless",
		"combat_cure_light_wounds", "combat_done", "combat_menu_main", "combat_menu_guard",
		"combat_menu_delay", "combat_menu_quit", "combat_menu_bandage", "combat_menu_speed",
		"combat_menu_exit", "combat_guard", "combat_bandage", "combat_quick_all",
		"combat_quick_magic_off", "combat_quick_magic_on", "combat_manual_control",
		"combat_speed_slower", "combat_speed_faster", "combat_speed_value", "combat_delay",
		"combat_monster_held", "combat_cloud_helpless", "combat_cloud_coughing",
		"combat_monster_magic_missile", "combat_quick_magic_metadata_missing",
		"combat_quick_magic_unsupported", "combat_monster_lightning_bolt_no_target",
		"combat_monster_lightning_bolt", "combat_monster_lightning_bolt_protected",
		"combat_spell_interrupted", "event_picture", "combat_miss", "combat_hit",
		"combat_multi_attack", "combat_victory", "combat_defeat", "combat_draw",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == "" || got == key {
			t.Errorf("formal catalog does not translate %q", key)
		}
	}
}
