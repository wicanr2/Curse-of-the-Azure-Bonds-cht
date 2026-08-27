package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestCombatPlayerSpellContractDrivesTargetAndMessageLookup(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "mage", Class: party.ClassMagicUser, Level: 1,
		SpellSlots: []uint8{MagicMissileSpellID},
	}}
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "mage", Name: "mage", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20}},
		[]combat.Fighter{{ID: "target", Name: "target", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20}},
		17,
	); err != nil {
		t.Fatal(err)
	}
	definition, found := state.combatPlayerSpellDefinition(MagicMissileSpellID)
	if !found {
		t.Fatal("formal player spell contract was not loaded")
	}
	if got, want := state.combatPlayerSpellLabel(MagicMissileSpellID), state.catalog.Text(definition.MessageID, ""); got != want {
		t.Fatalf("player spell label=%q want catalog message %q", got, want)
	}
	if err := state.BeginCombatCast(MagicMissileSpellID); err != nil {
		t.Fatal(err)
	}
	if !state.CombatSpellTargetsEnemy() {
		t.Fatal("enemy target mode was not applied from the player spell contract")
	}
	state.CancelCombatCast()

	for index := range state.dataPack.CombatPlayerSpells {
		if state.dataPack.CombatPlayerSpells[index].SpellID == MagicMissileSpellID {
			state.dataPack.CombatPlayerSpells[index].TargetMode = "none"
			break
		}
	}
	if err := state.BeginCombatCast(MagicMissileSpellID); err != nil {
		t.Fatal(err)
	}
	if state.CombatSpellTargetsEnemy() {
		t.Fatal("mutated player target mode was ignored by the combat state")
	}
}

func TestCombatSpellChoicesFollowActiveCasterSlotOrderAndFoldDuplicates(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "mage", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{BurningHandsSpellID, MagicMissileSpellID, BurningHandsSpellID, 0xff},
	}}
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "mage", Name: "mage", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20}},
		[]combat.Fighter{{ID: "target", Name: "target", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20}},
		17,
	); err != nil {
		t.Fatal(err)
	}
	choices := state.CombatSpellChoices()
	if len(choices) != 2 {
		t.Fatalf("CAST choices=%v, want two declared unique memorized spells", choices)
	}
	if choices[0].SpellID != BurningHandsSpellID || choices[1].SpellID != MagicMissileSpellID {
		t.Fatalf("CAST choice order=%v, want [%d %d]", choices, BurningHandsSpellID, MagicMissileSpellID)
	}
	if choices[0].Label == "" || choices[1].Label == "" {
		t.Fatalf("CAST choices must carry localized labels: %v", choices)
	}
}
