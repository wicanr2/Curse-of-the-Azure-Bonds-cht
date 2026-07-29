package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestCombatVisualMissileDefersVictoryAndOrdersSounds(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	heroes := []combat.Fighter{{
		ID: "archer", Name: "弓手", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0,
		AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: 30, MissileWeapon: true,
		HasCombatPosition: true, CombatX: 1, CombatY: 3,
	}}
	enemies := []combat.Fighter{{
		ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 1, MaxHitPoints: 1, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 5, CombatY: 3,
	}}
	if err := state.StartCombat(heroes, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualMissile || event.ActorID != "archer" ||
		event.TargetID != "goblin" || !event.Killed || state.Mode != ModeCombat {
		t.Fatalf("missile visual=%+v ok=%v mode=%v", event, ok, state.Mode)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("sounds emitted before visual phase: %v", sounds)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundMissile {
		t.Fatalf("travel sounds=%v", sounds)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration + combat.VisualTravelDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundHit {
		t.Fatalf("impact sounds=%v", sounds)
	}
	deathAt := combat.VisualWindupDuration + combat.VisualTravelDuration +
		combat.VisualImpactDuration + combat.VisualCommitDuration
	if err := state.AdvanceCombatVisual(deathAt); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundDeath {
		t.Fatalf("death sounds=%v", sounds)
	}
	if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("deferred victory mode=%v status=%v", state.Mode, state.CombatStatus())
	}
}

func TestCombatVisualMagicMissileCarriesProjectileCount(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{MagicMissileSpellID},
	}}
	heroes := []combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2,
	}}
	enemies := []combat.Fighter{{
		ID: "ogre", Name: "食人魔", Side: combat.SideEnemy,
		HitPoints: 50, MaxHitPoints: 50, ArmorClass: 0,
		HasCombatPosition: true, CombatX: 5, CombatY: 4,
	}}
	if err := state.StartCombat(heroes, enemies, 9); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(MagicMissileSpellID); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualMagicMissile || event.Projectiles != 3 ||
		event.From != (combat.TilePoint{X: 1, Y: 2}) ||
		event.To != (combat.TilePoint{X: 5, Y: 4}) {
		t.Fatalf("magic visual=%+v ok=%v", event, ok)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration + combat.VisualTravelDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundMagicHit {
		t.Fatalf("magic impact sounds=%v", sounds)
	}
}

func TestCombatFireballPlayerPathQueuesOrderedAreaImpacts(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	state.partyRoster = party.Roster{
		{ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
			SpellSlots: []uint8{FireballSpellID}},
		{ID: "ally", Name: "戰士", Class: party.ClassFighter, Level: 5},
	}
	saves := []uint8{1, 1, 1, 1, 1}
	heroes := []combat.Fighter{
		{ID: "mage", Name: "法師", Side: combat.SideParty,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
			HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves},
		{ID: "ally", Name: "戰士", Side: combat.SideParty,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: -10,
			HasCombatPosition: true, CombatX: 3, CombatY: 2, SavingThrows: saves},
	}
	enemies := []combat.Fighter{
		{ID: "orc-near", Name: "半獸人", Side: combat.SideEnemy,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 2, SavingThrows: saves},
		{ID: "orc-far", Name: "半獸人", Side: combat.SideEnemy,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 8, CombatY: 2, SavingThrows: saves},
	}
	if err := state.StartCombat(heroes, enemies, 23); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastFireball() {
		t.Fatal("memorized Fireball was not exposed on the normal combat turn")
	}
	if err := state.CombatSelectTarget(1); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(FireballSpellID); err != nil {
		t.Fatal(err)
	}
	if point, ok := state.CombatSpellTargetPoint(); !ok || point != (combat.TilePoint{X: 5, Y: 2}) {
		t.Fatalf("initial Fireball point=(%+v,%v)", point, ok)
	}
	if err := state.CombatMoveSpellTarget(-1, 0); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(FireballSpellID); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualAreaSpell || event.Effect != "fireball" ||
		event.To != (combat.TilePoint{X: 4, Y: 2}) || len(event.Impacts) != 2 {
		t.Fatalf("Fireball visual=%+v ok=%v", event, ok)
	}
	gotTargets := map[string]bool{}
	for _, impact := range event.Impacts {
		gotTargets[impact.TargetID] = true
	}
	if !gotTargets["ally"] || !gotTargets["orc-near"] || gotTargets["mage"] || gotTargets["orc-far"] {
		t.Fatalf("Fireball targets=%v", gotTargets)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Fireball slot not consumed: %v", state.partyRoster[0].SpellSlots)
	}
}

func TestCombatVisualEnemyTurnStopsAtOneActionUntilHandoff(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	heroes := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: -20,
		InitiativeBonus: -20, HasCombatPosition: true, CombatX: 1, CombatY: 3,
	}}
	enemies := []combat.Fighter{{
		ID: "orc", Name: "半獸人", Side: combat.SideEnemy,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
		InitiativeBonus: 30, HasCombatPosition: true, CombatX: 4, CombatY: 3,
	}}
	if err := state.StartCombat(heroes, enemies, 3); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.ActorID != "orc" {
		t.Fatalf("enemy visual=%+v ok=%v", event, ok)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "orc" {
		t.Fatalf("turn advanced before playback: active=%+v ok=%v", active, ok)
	}
	if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
		t.Fatal(err)
	}
	if state.CombatVisualPending() {
		t.Fatalf("unexpected second visual before party input: %+v", state.combatVisual)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "hero" {
		t.Fatalf("handoff active=%+v ok=%v", active, ok)
	}
}
