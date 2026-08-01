package game

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func combatVisualCatalog(t *testing.T) locale.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

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
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundArrow {
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
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundDead {
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
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 2 ||
		sounds[0] != SoundCast || sounds[1] != SoundSpellHit {
		t.Fatalf("magic impact sounds=%v", sounds)
	}
}

func TestMonsterEffect84QueuesLightningWithTwoDamagePools(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.EnableCombatVisualTimeline(true)
	state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 12 && y >= 0 && y < 6}
	})
	fighters := []combat.Fighter{
		{ID: "tyranthraxus", Name: "提朗瑟克斯", Side: combat.SideEnemy,
			HitPoints: 200, MaxHitPoints: 200, ArmorClass: 0,
			HasCombatPosition: true, CombatX: 1, CombatY: 2,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x84, Innate: true}}},
		{ID: "hero", Name: "英雄", Side: combat.SideParty,
			HitPoints: 200, MaxHitPoints: 200, ArmorClass: 0,
			HasCombatPosition: true, CombatX: 3, CombatY: 2,
			SavingThrows: []uint8{30, 30, 30, 30, 30}},
	}
	battle, err := combat.NewBattle(fighters, 415)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := battle.StartRound(); err != nil {
		t.Fatal(err)
	}
	state.battle = battle
	state.Mode = ModeCombat
	state.combatTurns = []combat.Turn{{FighterID: "tyranthraxus"}, {FighterID: "hero"}}
	if err := state.advanceCombatToParty(); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualLineSpell || event.Effect != "lightning_bolt" ||
		event.ActorID != "tyranthraxus" || event.To != (combat.TilePoint{X: 3, Y: 2}) ||
		len(event.Impacts) == 0 || event.Impacts[0].TargetID != "hero" {
		t.Fatalf("monster lightning visual=%+v ok=%v", event, ok)
	}
	if got, _ := battle.Fighter("hero"); got.HitPoints >= 200 {
		t.Fatalf("monster lightning did not damage target: %+v", got)
	}
	wantMessage := state.catalog.Text("combat_monster_lightning_bolt", "")
	if wantMessage == "" || state.combatMessage == "" {
		t.Fatalf("monster lightning localization message=%q format=%q", state.combatMessage, wantMessage)
	}
}

func TestMonsterEffect84ConsumesTurnWhenNoRangedTargetIsReachable(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.EnableCombatVisualTimeline(true)
	state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
		if x == 3 && y == 1 {
			return combat.LineCell{Valid: true, Reflect: true}
		}
		return combat.LineCell{Valid: x >= 1 && x < 8 && y >= 1 && y < 5}
	})
	battle, err := combat.NewBattle([]combat.Fighter{
		{ID: "monster", Name: "放電怪物", Side: combat.SideEnemy,
			HitPoints: 50, MaxHitPoints: 50, ArmorClass: 0,
			AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x84, Innate: true}}},
		{ID: "hero", Name: "英雄", Side: combat.SideParty,
			HitPoints: 50, MaxHitPoints: 50, ArmorClass: 0,
			HasCombatPosition: true, CombatX: 3, CombatY: 1,
			SavingThrows: []uint8{30, 30, 30, 30, 30}},
	}, 416)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := battle.StartRound(); err != nil {
		t.Fatal(err)
	}
	state.battle = battle
	state.Mode = ModeCombat
	state.combatTurns = []combat.Turn{{FighterID: "monster"}, {FighterID: "hero"}}
	if err := state.advanceCombatToParty(); err != nil {
		t.Fatal(err)
	}
	if state.CombatVisualPending() {
		t.Fatalf("unreachable fallback unexpectedly queued visual: %+v", state.combatVisual)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "hero" {
		t.Fatalf("effect 84 did not consume monster turn: active=%+v ok=%v", active, ok)
	}
	if got, _ := battle.Fighter("hero"); got.HitPoints != 50 {
		t.Fatalf("unreachable target received fallback attack: %+v", got)
	}
	want := state.catalog.Text("combat_monster_lightning_bolt_no_target", "")
	if want == "" || state.combatMessage == "" {
		t.Fatalf("no-target localization message=%q format=%q", state.combatMessage, want)
	}
}

func TestMonsterEffect84StopsBeforeOriginalRoundFour(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.EnableCombatVisualTimeline(true)
	state.SetCombatLineTerrain(func(x, y int) combat.LineCell { return combat.LineCell{Valid: true} })
	battle, err := combat.NewBattle([]combat.Fighter{
		{ID: "monster", Name: "怪物", Side: combat.SideEnemy, HitPoints: 50, MaxHitPoints: 50,
			AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 1, CombatY: 1,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x84, Innate: true}}},
		{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 50, MaxHitPoints: 50,
			ArmorClass: 0, HasCombatPosition: true, CombatX: 3, CombatY: 1,
			SavingThrows: []uint8{30, 30, 30, 30, 30}},
	}, 415)
	if err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 4; round++ {
		if _, err := battle.StartRound(); err != nil {
			t.Fatal(err)
		}
	}
	state.battle = battle
	state.Mode = ModeCombat
	state.combatTurns = []combat.Turn{{FighterID: "monster"}, {FighterID: "hero"}}
	if err := state.advanceCombatToParty(); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualMelee {
		t.Fatalf("round-four action=%+v ok=%v", event, ok)
	}
}

func TestCombatFireballPlayerPathQueuesOrderedAreaImpacts(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
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
			HasCombatPosition: true, CombatX: 5, CombatY: 2, SavingThrows: saves,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x70, Innate: true}}},
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
		if impact.TargetID == "orc-near" && (!impact.Protected || impact.Damage != 0) {
			t.Fatalf("Fireball protected impact=%+v", impact)
		}
	}
	if !gotTargets["ally"] || !gotTargets["orc-near"] || gotTargets["mage"] || gotTargets["orc-far"] {
		t.Fatalf("Fireball targets=%v", gotTargets)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Fireball slot not consumed: %v", state.partyRoster[0].SpellSlots)
	}
	wantMessage := fmt.Sprintf(
		state.catalog.Text("combat_fireball_protected", ""), "法師", 2, event.Impacts[0].Damage+event.Impacts[1].Damage, 1,
	)
	if state.CombatMessage() != wantMessage {
		t.Fatalf("Fireball protected message=%q want=%q", state.CombatMessage(), wantMessage)
	}
}

func TestCombatLightningBoltPlayerPathConsumesSlotAndQueuesSegments(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.EnableCombatVisualTimeline(true)
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{LightningBoltSpellID},
	}}
	saves := []uint8{30, 30, 30, 30, 30}
	heroes := []combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves,
	}}
	enemies := []combat.Fighter{
		{ID: "near", Name: "近敵", Side: combat.SideEnemy,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 3, CombatY: 2, SavingThrows: saves,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x87, Innate: true}}},
		{ID: "far", Name: "遠敵", Side: combat.SideEnemy,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 2, SavingThrows: saves},
	}
	if err := state.StartCombat(heroes, enemies, 29); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastLightningBolt() {
		t.Fatal("memorized Lightning Bolt was not exposed on the normal combat turn")
	}
	if err := state.CombatSelectTarget(1); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(LightningBoltSpellID); err != nil {
		t.Fatal(err)
	}
	if point, ok := state.CombatSpellTargetPoint(); !ok || point != (combat.TilePoint{X: 3, Y: 2}) {
		t.Fatalf("initial Lightning Bolt point=(%+v,%v)", point, ok)
	}
	terrain := func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 10 && y >= 0 && y < 6}
	}
	if err := state.CombatCastWithTerrain(LightningBoltSpellID, terrain); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualLineSpell || event.Effect != "lightning_bolt" ||
		event.To != (combat.TilePoint{X: 3, Y: 2}) || event.TravelImpacts != 1 ||
		len(event.Impacts) != 2 || len(event.Segments) < 2 {
		t.Fatalf("Lightning Bolt visual=%+v ok=%v", event, ok)
	}
	if event.Impacts[0].TargetID != "near" || event.Impacts[1].TargetID != "far" {
		t.Fatalf("Lightning Bolt impacts=%+v", event.Impacts)
	}
	if !event.Impacts[0].Protected || event.Impacts[0].Damage != 0 || event.Impacts[1].Protected {
		t.Fatalf("Lightning Bolt elemental protection impacts=%+v", event.Impacts)
	}
	wantMessage := fmt.Sprintf(
		state.catalog.Text("combat_lightning_bolt_protected", ""), "法師", 2,
		event.Impacts[0].Damage+event.Impacts[1].Damage, 1,
	)
	if state.CombatMessage() != wantMessage {
		t.Fatalf("Lightning Bolt protected message=%q want=%q", state.CombatMessage(), wantMessage)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Lightning Bolt slot not consumed: %v", state.partyRoster[0].SpellSlots)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundLightning {
		t.Fatalf("Lightning Bolt travel sounds=%v", sounds)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration + combat.VisualTravelDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); len(sounds) != 1 || sounds[0] != SoundSpellHit {
		t.Fatalf("Lightning Bolt impact sounds=%v", sounds)
	}
}

func TestCombatLightningBoltRestoresSlotWhenTerrainIsUnavailable(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{LightningBoltSpellID},
	}}
	saves := []uint8{30, 30, 30, 30, 30}
	if err := state.StartCombat([]combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves,
	}}, []combat.Fighter{{
		ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 3, CombatY: 2, SavingThrows: saves,
	}}, 31); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(LightningBoltSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCastWithTerrain(LightningBoltSpellID, nil); err == nil {
		t.Fatal("Lightning Bolt unexpectedly cast without terrain")
	}
	if slots := state.partyRoster[0].SpellSlots; len(slots) != 1 || slots[0] != LightningBoltSpellID {
		t.Fatalf("Lightning Bolt slot rollback=%v", slots)
	}
}

func TestCombatStinkingCloudPlayerPathConsumesSlotAndCreatesPersistentArea(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{StinkingCloudSpellID},
	}}
	saves := []uint8{10, 10, 10, 10, 10}
	if err := state.StartCombat([]combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves,
	}}, []combat.Fighter{{
		ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, InitiativeBonus: -20,
		HasCombatPosition: true, CombatX: 4, CombatY: 2, SavingThrows: saves,
	}}, 41); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastStinkingCloud() {
		t.Fatal("memorized Stinking Cloud was not exposed on the normal combat turn")
	}
	if err := state.BeginCombatCast(StinkingCloudSpellID); err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 7}
	}
	if err := state.CombatCastWithTerrain(StinkingCloudSpellID, terrain); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualAreaSpell || event.Effect != "stinking_cloud" ||
		event.PersistentAreaID == 0 || len(event.Impacts) != 1 {
		t.Fatalf("Stinking Cloud visual=%+v ok=%v", event, ok)
	}
	areas := state.CombatPersistentAreas()
	if len(areas) != 1 || len(areas[0].Cells) != 4 || areas[0].ID != event.PersistentAreaID {
		t.Fatalf("persistent areas=%+v", areas)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Stinking Cloud slot not consumed: %v", state.partyRoster[0].SpellSlots)
	}
	if err := state.AdvanceCombatVisual(combat.VisualWindupDuration + combat.VisualTravelDuration); err != nil {
		t.Fatal(err)
	}
	if sounds := state.ConsumeSoundEvents(); !reflect.DeepEqual(sounds, []SoundEvent{SoundCast}) {
		t.Fatalf("Stinking Cloud cast sound=%v, want [%s]", sounds, SoundCast)
	}
	if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
		t.Fatal(err)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "mage" {
		t.Fatalf("cloud-affected enemy turn was not skipped: active=%+v ok=%v", active, ok)
	}
}

func TestCombatStinkingCloudRestoresSlotWhenEveryCellIsBlocked(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 5,
		SpellSlots: []uint8{StinkingCloudSpellID},
	}}
	saves := []uint8{10, 10, 10, 10, 10}
	if err := state.StartCombat([]combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves,
	}}, []combat.Fighter{{
		ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, InitiativeBonus: -20,
		HasCombatPosition: true, CombatX: 4, CombatY: 2, SavingThrows: saves,
	}}, 43); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(StinkingCloudSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCastWithTerrain(StinkingCloudSpellID, func(int, int) combat.LineCell {
		return combat.LineCell{Valid: true, Reflect: true}
	}); err == nil {
		t.Fatal("Stinking Cloud unexpectedly cast into four walls")
	}
	if slots := state.partyRoster[0].SpellSlots; len(slots) != 1 || slots[0] != StinkingCloudSpellID {
		t.Fatalf("Stinking Cloud slot rollback=%v", slots)
	}
	if len(state.CombatPersistentAreas()) != 0 {
		t.Fatal("failed cast left a persistent area")
	}
}

func TestCombatCloudkillPlayerPathConsumesSlotAndKillsLowHitDiceTarget(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 7,
		SpellSlots: []uint8{CloudkillSpellID},
	}}
	saves := []uint8{10, 10, 10, 10, 10}
	if err := state.StartCombat([]combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty, HitDice: 7,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 0, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 1, CombatY: 2, SavingThrows: saves,
	}}, []combat.Fighter{{
		ID: "enemy", Name: "半獸人", Side: combat.SideEnemy, HitDice: 4,
		HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, InitiativeBonus: -20,
		HasCombatPosition: true, CombatX: 4, CombatY: 2, SavingThrows: saves,
	}}, 47); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastCloudkill() {
		t.Fatal("memorized Cloudkill was not exposed on the normal combat turn")
	}
	if err := state.BeginCombatCast(CloudkillSpellID); err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 7}
	}
	if err := state.CombatCastWithTerrain(CloudkillSpellID, terrain); err != nil {
		t.Fatal(err)
	}
	event, ok := state.CombatVisualEvent()
	if !ok || event.Effect != "cloudkill" || event.PersistentAreaID == 0 ||
		len(event.Impacts) != 1 || !event.Impacts[0].Killed {
		t.Fatalf("Cloudkill visual=%+v ok=%v", event, ok)
	}
	areas := state.CombatPersistentAreas()
	if len(areas) != 1 || len(areas[0].Cells) != 9 {
		t.Fatalf("Cloudkill persistent areas=%+v", areas)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Cloudkill slot not consumed: %v", state.partyRoster[0].SpellSlots)
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
