package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	enginedamage "github.com/wicanr2/golden-box-remake-engine/combat/damage"
	enginemodifier "github.com/wicanr2/golden-box-remake-engine/combat/modifier"
	engineposthit "github.com/wicanr2/golden-box-remake-engine/combat/posthit"
	enginequickspell "github.com/wicanr2/golden-box-remake-engine/combat/quickspell"
	engineresistance "github.com/wicanr2/golden-box-remake-engine/combat/resistance"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
)

// StartCombat creates the first playable battle adapter. Party and encounter
// records are supplied by the data layer, so this state machine does not
// invent MON*CHA statistics or DOS party layout.
func (s *State) StartCombat(party, enemies []combat.Fighter, seed int64) error {
	if len(party) == 0 || len(enemies) == 0 {
		return fmt.Errorf("combat needs at least one party member and enemy")
	}
	fighters := make([]combat.Fighter, 0, len(party)+len(enemies))
	if len(party)+len(enemies) > 72 {
		return fmt.Errorf("combat has %d fighters, original OBJECTLIST capacity is 72", len(party)+len(enemies))
	}
	s.combatReferenceCoords = false
	for _, fighter := range append(append([]combat.Fighter(nil), party...), enemies...) {
		if fighter.HasCombatPosition && (fighter.CombatX >= 16 || fighter.CombatY >= 10) {
			s.combatReferenceCoords = true
			break
		}
	}
	partyIndex := 0
	for _, fighter := range party {
		if fighter.Side != combat.SideParty {
			return fmt.Errorf("fighter %q is not marked as party", fighter.ID)
		}
		if !fighter.HasCombatPosition {
			tile := combat.FormationTile(fighter.Side, partyIndex)
			fighter.HasCombatPosition, fighter.CombatX, fighter.CombatY = true, tile.X, tile.Y
		}
		if fighter.CombatSize == 0 {
			fighter.CombatSize = 1
		}
		iconDirection, _ := combat.IconDirectionForTeam(s.combatMapDirection, fighter.Side)
		fighter.IconDirection = iconDirection
		fighter.LegacyObjectID = uint8(len(fighters) + 1)
		fighters = append(fighters, fighter)
		partyIndex++
	}
	enemyIndex := 0
	for _, fighter := range enemies {
		if fighter.Side != combat.SideEnemy {
			return fmt.Errorf("fighter %q is not marked as enemy", fighter.ID)
		}
		if !fighter.HasCombatPosition {
			tile := combat.FormationTile(fighter.Side, enemyIndex)
			fighter.HasCombatPosition, fighter.CombatX, fighter.CombatY = true, tile.X, tile.Y
		}
		if fighter.CombatSize == 0 {
			fighter.CombatSize = 1
		}
		iconDirection, _ := combat.IconDirectionForTeam(s.combatMapDirection, fighter.Side)
		fighter.IconDirection = iconDirection
		fighter.LegacyObjectID = uint8(len(fighters) + 1)
		fighters = append(fighters, fighter)
		enemyIndex++
	}
	battle, err := combat.NewBattle(fighters, seed)
	if err != nil {
		return err
	}
	if s.dataPack != nil {
		rules, err := s.dataPack.ResolveCombatAffectRules()
		if err != nil {
			return fmt.Errorf("resolve combat affect rules: %w", err)
		}
		battle.SetDamageRules(rules)
		conditionalRules, err := s.dataPack.ResolveCombatConditionalModifiers()
		if err != nil {
			return fmt.Errorf("resolve combat conditional modifiers: %w", err)
		}
		battle.SetConditionalModifierRules(conditionalRules)
		magicResistanceRules, err := s.dataPack.ResolveCombatMagicResistanceRules()
		if err != nil {
			return fmt.Errorf("resolve combat magic resistance rules: %w", err)
		}
		battle.SetMagicResistanceRules(magicResistanceRules)
		postHitRules, err := s.dataPack.ResolveCombatPostHitRules()
		if err != nil {
			return fmt.Errorf("resolve combat post-hit rules: %w", err)
		}
		battle.SetPostHitRules(postHitRules)
	} else {
		battle.SetDamageRules([]enginedamage.Rule(nil))
		battle.SetConditionalModifierRules([]enginemodifier.Rule(nil))
		battle.SetMagicResistanceRules([]engineresistance.Rule(nil))
		battle.SetPostHitRules([]engineposthit.Rule(nil))
	}
	if err := s.applyDataPackCombatModifiers(battle); err != nil {
		return err
	}
	if err := battle.BeginScheduledRound(); err != nil {
		return err
	}
	s.battle = battle
	if err := s.SetParty(party); err != nil {
		return err
	}
	s.combatTurns = nil
	s.combatTurnIndex = 0
	s.combatDelayedTurns = make(map[int]bool)
	s.combatTargetIndex = 0
	s.combatQuickMagic = false
	s.combatVisual = nil
	s.combatVisualElapsed = 0
	s.combatMessage = s.catalog.Text("combat_started", "combat_started")
	s.Mode = ModeCombat
	return s.advanceCombatToParty()
}

// SetCombatLineTerrain installs the renderer-owned combat map projection used
// by autonomous reflecting line effects. The callback is title-neutral and
// deliberately not serialized; frontends restore it when constructing State.
func (s *State) SetCombatLineTerrain(terrain combat.LineTerrain) {
	s.combatLineTerrain = terrain
}

// SetCombatScanMapProvider installs the current title-owned TACTICALMAP
// projection. It is intentionally not serialized: the frontend reconstructs
// floor buffers and restores this read-only adapter when State is created.
func (s *State) SetCombatScanMapProvider(provider func() (enginescan.TacticalMap, error)) {
	s.combatScanMapProvider = provider
}

func (s *State) applyDataPackCombatModifiers(battle *combat.Battle) error {
	if s.dataPack == nil || s.session == nil {
		return nil
	}
	for _, definition := range s.dataPack.CombatModifiers {
		if definition.SourceNamespace != "ecl_work" {
			return fmt.Errorf("combat modifier %q uses unsupported source namespace %q",
				definition.ID, definition.SourceNamespace)
		}
		raw, found := s.session.MemoryValue(definition.SourceAddress)
		if !found {
			continue
		}
		modifier, err := definition.Decode(raw)
		if err != nil {
			return err
		}
		var side combat.Side
		switch definition.Side {
		case "party":
			side = combat.SideParty
		case "enemy":
			side = combat.SideEnemy
		default:
			return fmt.Errorf("combat modifier %q uses unsupported side %q", definition.ID, definition.Side)
		}
		switch definition.Target {
		case "attack_roll":
			if err := battle.SetSideAttackRollModifier(side, modifier); err != nil {
				return err
			}
		default:
			return fmt.Errorf("combat modifier %q uses unsupported target %q", definition.ID, definition.Target)
		}
	}
	return nil
}

// StartEncounter is the data bridge from the bounded ECL runner to the
// playable battle state. ECL supplies only spawn IDs/counts; MON*CHA supplies
// the combat statistics, and the caller supplies the decoded party roster.
func (s *State) StartEncounter(result ecl.RunResult, records map[uint8]monster.Record, party []combat.Fighter, seed int64) error {
	return s.StartEncounterWithAffects(result, records, nil, party, seed)
}

// StartEncounterWithAffects is the data bridge including the chapter-local
// MON*SPC table. Raw effects are preserved on enemy fighters but are not yet
// interpreted as combat rules.
func (s *State) StartEncounterWithAffects(result ecl.RunResult, records map[uint8]monster.Record, affects map[uint8][]monster.AffectRecord, party []combat.Fighter, seed int64) error {
	if !result.CombatRequested {
		return fmt.Errorf("ECL result does not request combat")
	}
	s.combatReturnMode = ModeWilderness
	if s.eclMenuReturnMode == ModeDungeon || s.eventReturnMode == ModeDungeon {
		s.combatReturnMode = ModeDungeon
	}
	// Monster IDs use the same global chapter ranges as ECL block IDs, but a
	// world dispatcher may summon a previous area's monster after NEWECL.
	// Resolve every spawn from its ID namespace instead of assuming that the
	// current ECL block owns the MON table (ECL1 block 0x50 summons MON5 0x3C).
	resolvedRecords := make(map[uint8]monster.Record, len(records))
	for id, record := range records {
		resolvedRecords[id] = record
	}
	resolvedAffects := make(map[uint8][]monster.AffectRecord, len(affects))
	for id, list := range affects {
		resolvedAffects[id] = list
	}
	for _, spawn := range result.MonsterSpawns {
		chapter := monsterChapterForBlock(spawn.MonsterID)
		if record, ok := s.monsterRecordsByECL[chapter][spawn.MonsterID]; ok {
			resolvedRecords[spawn.MonsterID] = record
		}
		if list, ok := s.monsterAffectsByECL[chapter][spawn.MonsterID]; ok {
			resolvedAffects[spawn.MonsterID] = list
		}
	}
	ecl.ApplyCombatTeamWrites(result.MonsterSpawns, result.CombatTeamWrites)
	enemies, err := monster.BuildEnemiesWithAffects(result.MonsterSpawns, resolvedRecords, resolvedAffects)
	if err != nil {
		return err
	}
	enemyIndex := 0
	for _, spawn := range result.MonsterSpawns {
		for count := uint8(0); count < spawn.Count && enemyIndex < len(enemies); count++ {
			enemies[enemyIndex].SpriteSet = monsterChapterForBlock(spawn.MonsterID)
			enemyIndex++
		}
	}
	for index := range enemies {
		if s.dataPack != nil {
			if localized, found := s.dataPack.LocalizeCombatantName(enemies[index].Name, s.catalog.Language); found {
				enemies[index].Name = localized
			}
		}
		if enemies[index].SpriteSet == 0 {
			enemies[index].SpriteSet = s.Area.GameArea
		}
		if result.MonsterSetup != nil {
			enemies[index].AnimationBlock = result.MonsterSetup.SpriteID
			enemies[index].HasAnimation = true
		}
	}
	for index := 0; index < len(enemies); {
		if enemies[index].Side == combat.SideParty {
			party = append(party, enemies[index])
			enemies = append(enemies[:index], enemies[index+1:]...)
			continue
		}
		index++
	}
	return s.StartCombat(party, enemies, seed)
}

func (s *State) CombatActive() bool { return s.battle != nil && s.Mode == ModeCombat }

// EnableCombatVisualTimeline makes actor handoff wait for renderer playback.
// It is opt-in so headless rules tests and non-visual adapters retain their
// deterministic synchronous behavior.
func (s *State) EnableCombatVisualTimeline(enabled bool) {
	s.combatVisualEnabled = enabled
	if !enabled {
		s.combatVisual = nil
		s.combatVisualElapsed = 0
	}
}

func (s *State) CombatVisualEvent() (combat.VisualEvent, bool) {
	if s.combatVisual == nil {
		return combat.VisualEvent{}, false
	}
	return *s.combatVisual, true
}

func (s *State) CombatVisualPending() bool { return s.combatVisual != nil }

// CombatVisualElapsed is the authoritative timeline position. The frontend
// supplies monotonic clock deltas; keeping the committed position in State
// allows an active visual transaction to survive save/load.
func (s *State) CombatVisualElapsed() time.Duration { return s.combatVisualElapsed }

func (s *State) queueCombatVisual(event combat.VisualEvent) bool {
	if !s.combatVisualEnabled {
		return false
	}
	s.combatVisualSerial++
	event.Serial = s.combatVisualSerial
	s.combatVisual = &event
	s.combatVisualElapsed = 0
	s.combatVisualTravelSent = false
	s.combatVisualImpactSent = -1
	s.combatVisualDeathSent = -1
	s.combatVisualAdvanceTurn = true
	return true
}

// AdvanceCombatVisual commits phase-associated sound and releases the next
// actor only after the renderer has presented the complete action.
func (s *State) AdvanceCombatVisual(elapsed time.Duration) error {
	if s.combatVisual == nil {
		return nil
	}
	if elapsed < 0 {
		return fmt.Errorf("negative combat visual elapsed time %s", elapsed)
	}
	if elapsed < s.combatVisualElapsed {
		return fmt.Errorf("combat visual elapsed time regressed from %s to %s", s.combatVisualElapsed, elapsed)
	}
	event := *s.combatVisual
	if elapsed > event.Duration() {
		elapsed = event.Duration()
	}
	s.combatVisualElapsed = elapsed
	frame := event.FrameAt(elapsed)
	if !s.combatVisualTravelSent && frame.Phase >= combat.VisualTravel {
		switch {
		case event.Kind == combat.VisualMissile:
			s.requestSound(SoundArrow)
		case event.Kind == combat.VisualLineSpell:
			s.requestSound(SoundLightning)
		case event.Kind == combat.VisualAreaSpell && event.Effect == "fireball":
			s.requestSound(SoundFireball)
		case event.Kind == combat.VisualTwinkle:
			s.requestSound(SoundCast)
		case event.Kind == combat.VisualMagicMissile || event.Kind == combat.VisualAreaSpell:
			s.requestSound(SoundCast)
		}
		s.combatVisualTravelSent = true
	}
	if impact, ok := event.Impact(frame); ok {
		if frame.Phase >= combat.VisualImpact && frame.ImpactIndex > s.combatVisualImpactSent {
			switch event.Kind {
			case combat.VisualMagicMissile, combat.VisualLineSpell, combat.VisualTwinkle:
				s.requestSound(SoundSpellHit)
			case combat.VisualAreaSpell:
				if event.Effect != "stinking_cloud" && event.Effect != "cloudkill" {
					s.requestSound(SoundSpellHit)
				}
			default:
				if impact.Hit {
					s.requestSound(SoundHit)
				} else {
					s.requestSound(SoundMiss)
				}
			}
			s.combatVisualImpactSent = frame.ImpactIndex
		}
		if impact.Killed && frame.Phase >= combat.VisualDeath && frame.ImpactIndex > s.combatVisualDeathSent {
			s.requestSound(SoundDead)
			s.combatVisualDeathSent = frame.ImpactIndex
		}
	} else if frame.Phase == combat.VisualHandoff {
		// A renderer may advance directly to the terminal frame. Keep the
		// timeline release deterministic; phase sounds are presentation cues
		// and are intentionally not replayed in a burst after being skipped.
		s.combatVisualImpactSent = event.ImpactCount() - 1
		for index := 0; index < event.ImpactCount(); index++ {
			impact, _ := event.ImpactAt(index)
			if impact.Killed {
				s.combatVisualDeathSent = index
			}
		}
	}
	if !frame.Done {
		return nil
	}
	s.combatVisual = nil
	s.combatVisualElapsed = 0
	if s.combatVisualAdvanceTurn {
		s.combatTurnIndex++
		s.combatVisualAdvanceTurn = false
	}
	if s.battle == nil || s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) CombatStatus() combat.Status {
	if s.battle == nil {
		return combat.StatusDraw
	}
	return s.battle.Status()
}

func (s *State) CombatFighters() []combat.Fighter {
	if s.battle == nil {
		return nil
	}
	return s.battle.Fighters()
}

func (s *State) CombatPersistentAreas() []combat.PersistentArea {
	if s.battle == nil {
		return nil
	}
	return s.battle.PersistentAreas()
}

func (s *State) CombatTurns() []combat.Turn {
	return append([]combat.Turn(nil), s.combatTurns...)
}

// CombatActiveFighter exposes the current turn's fighter to the renderer.
// The camera follows this position but does not own combat turn mutation.
func (s *State) CombatActiveFighter() (combat.Fighter, bool) {
	if s.battle == nil || s.combatTurnIndex >= len(s.combatTurns) {
		return combat.Fighter{}, false
	}
	return s.fighter(s.combatTurns[s.combatTurnIndex].FighterID)
}

// BeginCombatView opens the RuleBook's read-only character screen without
// consuming the active fighter's turn.
func (s *State) BeginCombatView() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	fighter, ok := s.CombatActiveFighter()
	if !ok || fighter.Side != combat.SideParty {
		return fmt.Errorf("it is not a party character turn")
	}
	s.combatView = true
	s.combatViewFighterID = fighter.ID
	return nil
}

func (s *State) EndCombatView() {
	s.combatView = false
	s.combatViewFighterID = ""
}

func (s *State) CombatViewActive() bool { return s.combatView }

func (s *State) CombatViewFighter() (combat.Fighter, bool) {
	if !s.combatView || s.battle == nil {
		return combat.Fighter{}, false
	}
	return s.fighter(s.combatViewFighterID)
}

// CombatViewLines is a renderer-neutral Traditional Chinese summary. The
// same state boundary can feed a text UI or a later Gold Box front end.
func (s *State) CombatViewLines() []string {
	fighter, ok := s.CombatViewFighter()
	if !ok {
		return nil
	}
	return []string{
		fmt.Sprintf(s.catalog.Text("combat_view_name", "combat_view_name"), fighter.Name),
		fmt.Sprintf(s.catalog.Text("combat_view_hp", "combat_view_hp"), fighter.HitPoints, fighter.MaxHitPoints),
		fmt.Sprintf(s.catalog.Text("combat_view_ac", "combat_view_ac"), fighter.ArmorClass),
		fmt.Sprintf(s.catalog.Text("combat_view_attack", "combat_view_attack"), fighter.AttackBonus),
	}
}

func (s *State) CombatMessage() string { return s.combatMessage }

// CombatVisualMessage projects typed visual impacts into localized text while
// preserving the renderer-neutral phase ordering carried by VisualFrame.
func (s *State) CombatVisualMessage(event combat.VisualEvent, frame combat.VisualFrame, fallback string) string {
	impact, ok := event.Impact(frame)
	if !ok {
		return fallback
	}
	name := impact.TargetID
	if fighter, found := s.fighter(impact.TargetID); found {
		name = fighter.Name
	}
	if event.Effect == "cloudkill" {
		if frame.Phase != combat.VisualImpact && frame.Phase != combat.VisualCommit && frame.Phase != combat.VisualDeath {
			return fallback
		}
		key := "combat_visual_cloudkill_killed"
		if impact.Saved {
			key = "combat_visual_cloudkill_saved"
		}
		return fmt.Sprintf(s.catalog.Text(key, key), name)
	}
	if event.Effect == "stinking_cloud" {
		if frame.Phase != combat.VisualImpact && frame.Phase != combat.VisualCommit {
			return fallback
		}
		if impact.Saved {
			return fmt.Sprintf(s.catalog.Text("combat_visual_stinking_cloud_saved", "combat_visual_stinking_cloud_saved"), name)
		}
		return fmt.Sprintf(s.catalog.Text("combat_visual_stinking_cloud_failed", "combat_visual_stinking_cloud_failed"), name, impact.Damage)
	}
	if impact.Resisted {
		if frame.Phase != combat.VisualImpact {
			return fallback
		}
		switch event.Kind {
		case combat.VisualAreaSpell:
			if event.Effect == "fireball" {
				return fmt.Sprintf(s.catalog.Text("combat_visual_fireball_resisted", "combat_visual_fireball_resisted"), name)
			}
		case combat.VisualLineSpell:
			return fmt.Sprintf(s.catalog.Text("combat_visual_line_resisted", "combat_visual_line_resisted"), name)
		}
	}
	if event.Kind != combat.VisualLineSpell || impact.Protected {
		return fallback
	}
	switch frame.Phase {
	case combat.VisualImpact:
		return fmt.Sprintf(s.catalog.Text("combat_visual_line_damage", "combat_visual_line_damage"), name, impact.Damage)
	case combat.VisualCommit:
		key := "combat_visual_line_save_failed"
		if impact.Saved {
			key = "combat_visual_line_save_succeeded"
		}
		return fmt.Sprintf(s.catalog.Text(key, key), name)
	default:
		return fallback
	}
}

func (s *State) CombatMoveMode() bool { return s.combatMoveMode }

func (s *State) CombatUsesReferenceCoordinates() bool { return s.combatReferenceCoords }

func (s *State) BeginCombatMove() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	if _, ok := s.combatPartyTurn(); !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	fighter, _ := s.combatPartyTurn()
	s.combatMoveRemaining = fighter.MovementAllowance
	if s.combatMoveRemaining < 1 {
		s.combatMoveRemaining = 1
	}
	s.combatMoveMode = true
	return nil
}

func (s *State) CancelCombatMove() {
	s.combatMoveMode = false
	s.combatMoveRemaining = 0
}

func (s *State) CombatMoveRemaining() int { return s.combatMoveRemaining }

func (s *State) CombatMove(dx, dy int) error {
	return s.CombatMoveWithTerrain(dx, dy, nil)
}

func (s *State) CombatMoveWithTerrain(dx, dy int, terrain combat.MovementTerrain) error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	moveResult, err := s.battle.MoveWithTerrainAndFreeAttacks(caster.ID, dx, dy, s.combatMoveRemaining, terrain)
	if err != nil {
		return err
	}
	s.combatMoveRemaining -= moveResult.MovementCost
	endTurn := moveResult.Attack != nil || s.combatMoveRemaining <= 0
	s.combatMoveMode = !endTurn
	if moveResult.Attack != nil {
		target, ok := s.fighter(moveResult.Attack.TargetID)
		if !ok {
			return fmt.Errorf("move attack target %q disappeared", moveResult.Attack.TargetID)
		}
		s.combatMessage = formatAttackMessage(s.catalog, moveResult.Fighter, target, *moveResult.Attack)
		s.requestAttackSounds([]combat.AttackResult{*moveResult.Attack})
	} else {
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_moved", "combat_moved"), moveResult.Fighter.Name, moveResult.Fighter.CombatX, moveResult.Fighter.CombatY)
		if len(moveResult.GuardAttacks) > 0 {
			last := moveResult.GuardAttacks[len(moveResult.GuardAttacks)-1]
			s.requestAttackSounds(moveResult.GuardAttacks)
			s.combatMessage += " " + fmt.Sprintf(s.catalog.Text("combat_guard_reaction", "combat_guard_reaction"), last.Damage)
		}
		if len(moveResult.FreeAttacks) > 0 {
			last := moveResult.FreeAttacks[len(moveResult.FreeAttacks)-1]
			s.requestAttackSounds(moveResult.FreeAttacks)
			s.combatMessage += " " + fmt.Sprintf(s.catalog.Text("combat_free_attack", "combat_free_attack"), last.Damage)
		}
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	if s.combatMoveMode {
		return nil
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) CombatTargetIndex() int { return s.combatTargetIndex }

func (s *State) CombatTargets() []combat.Fighter {
	return s.livingBySide(combat.SideEnemy)
}

// CombatSelectTarget changes the selected living enemy. The renderer maps
// left/right input to this method; dead enemies are skipped.
func (s *State) CombatSelectTarget(delta int) error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return fmt.Errorf("combat has no living enemies")
	}
	s.combatTargetIndex = (s.combatTargetIndex + delta) % len(enemies)
	if s.combatTargetIndex < 0 {
		s.combatTargetIndex += len(enemies)
	}
	return nil
}

func (s *State) combatPartyTurn() (combat.Fighter, bool) {
	if s.combatTurnIndex >= len(s.combatTurns) {
		return combat.Fighter{}, false
	}
	turn := s.combatTurns[s.combatTurnIndex]
	fighter, ok := s.fighter(turn.FighterID)
	if !ok || fighter.Side != combat.SideParty || fighter.HitPoints <= 0 {
		return combat.Fighter{}, false
	}
	return fighter, true
}

// CombatCanCastMagicMissile exposes the UI gate for the one verified combat
// spell. Slot mutation remains inside CombatCast so a disabled input cannot
// consume a spell.
func (s *State) CombatCanCastMagicMissile() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == MagicMissileSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastSleep() bool {
	if !s.CombatActive() || s.combatScanMapProvider == nil {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == SleepSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastFireball() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == FireballSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastLightningBolt() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == LightningBoltSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastStinkingCloud() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == StinkingCloudSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastCloudkill() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			for _, spellID := range character.SpellSlots {
				if spellID == CloudkillSpellID {
					return true
				}
			}
		}
	}
	return false
}

func (s *State) CombatCanCastCureLightWounds() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	wounded := false
	for _, fighter := range s.combatHealingTargets() {
		if fighter.HitPoints < fighter.MaxHitPoints {
			wounded = true
			break
		}
	}
	if !wounded {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == CureLightWoundsSpellID {
				return true
			}
		}
	}
	return false
}

// CombatCanCastBless exposes the cleric's verified first-level Bless slot.
// Bless is a no-target confirmation spell in the current combat UI.
func (s *State) CombatCanCastBless() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == BlessSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastCurse() bool {
	if !s.CombatActive() || len(s.livingBySide(combat.SideEnemy)) == 0 {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == CurseSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastCauseLightWounds() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(s.causeLightWoundsTargets(caster)) == 0 {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == CauseLightWoundsSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastProtectionFromEvil() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(s.protectionFromEvilTargets(caster)) == 0 {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == ProtectionFromEvilSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCanCastProtectionFromGood() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(s.protectionFromGoodTargets(caster)) == 0 {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassCleric) {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == ProtectionFromGoodSpellID {
				return true
			}
		}
	}
	return false
}

func (s *State) CombatCastingSpell() uint8 { return s.combatCastingSpell }

func (s *State) CombatSpellTargetIndex() int { return s.combatSpellTargetIndex }

func (s *State) CombatSpellTargetPoint() (combat.TilePoint, bool) {
	return s.combatSpellTargetPoint, s.combatSpellTargetsPoint
}

func (s *State) CombatSpellTargets() []combat.Fighter {
	switch s.combatCastingSpell {
	case CureLightWoundsSpellID:
		return s.combatHealingTargets()
	case MagicMissileSpellID:
		return s.livingBySide(combat.SideEnemy)
	case ProtectionFromGoodSpellID:
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.protectionFromGoodTargets(caster)
	case CurseSpellID:
		return s.livingBySide(combat.SideEnemy)
	case CauseLightWoundsSpellID:
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.causeLightWoundsTargets(caster)
	case ProtectionFromEvilSpellID:
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.protectionFromEvilTargets(caster)
	default:
		return nil
	}
}

// combatHealingTargets mirrors the reference heal_player target boundary:
// wounded living party members and downed unconscious/dying/animated party
// members are legal; dead characters are not. A downed target stays out of
// combat placement until an explicit combat-heal/placement operation.
func (s *State) combatHealingTargets() []combat.Fighter {
	if s.battle == nil {
		return nil
	}
	targets := make([]combat.Fighter, 0)
	for _, fighter := range s.battle.Fighters() {
		if fighter.Side != combat.SideParty || fighter.HitPoints >= fighter.MaxHitPoints {
			continue
		}
		status := party.HealthStatusOK
		for _, character := range s.partyRoster {
			if character.ID == fighter.ID {
				status = character.HealthStatus
				break
			}
		}
		if status == party.HealthStatusDead || status == party.HealthStatusStoned {
			continue
		}
		targets = append(targets, fighter)
	}
	return targets
}

// quickCureTarget projects PC-98 COMPTARGCURE into stable fighter identities.
// DXDIR/DYDIR scan N, NE, E, SE, S, SW, W, NW and self. A strictly lower
// current HP replaces the active candidate; the caster, scanned last, also
// replaces it when below half HP. A legal down-player marker wins only when
// the active candidate has at least eight HP.
func (s *State) quickCureTarget(caster combat.Fighter) (combat.Fighter, bool) {
	directions := [...]combat.TilePoint{
		{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0},
		{X: 1, Y: 1}, {X: 0, Y: 1}, {X: -1, Y: 1},
		{X: -1, Y: 0}, {X: -1, Y: -1}, {X: 0, Y: 0},
	}
	targets := s.combatHealingTargets()
	var activeTarget, downedTarget combat.Fighter
	activeFound, downedFound := false, false
	for _, direction := range directions {
		x, y := caster.CombatX+direction.X, caster.CombatY+direction.Y
		for _, target := range targets {
			if target.CombatX != x || target.CombatY != y {
				continue
			}
			if target.DownedCorpse {
				downedTarget, downedFound = target, true
				continue
			}
			if !activeFound || target.HitPoints < activeTarget.HitPoints ||
				(target.ID == caster.ID && target.HitPoints < target.MaxHitPoints/2) {
				activeTarget, activeFound = target, true
			}
		}
	}
	if downedFound && (!activeFound || activeTarget.HitPoints >= 8) {
		return downedTarget, true
	}
	if activeFound {
		return activeTarget, true
	}
	return combat.Fighter{}, false
}

// quickAreaSpellTarget is the bounded CoAB adapter for the PC-98 Quick
// MinRange path. The reference suitability routine builds a SCAN candidate
// list around each possible combat target before it accepts the spell. The
// exact linked-list tie/random policy is still outside this adapter; the
// stable fighter order is retained until that consumer is closed.
func (s *State) quickAreaSpellTarget(caster combat.Fighter, spellID, minRange uint8) (combat.TilePoint, bool, error) {
	if minRange == 0 {
		if spellID != CloudkillSpellID {
			return combat.TilePoint{}, false, fmt.Errorf("Quick spell 0x%02X requires a nonzero scan range", spellID)
		}
		if s.combatLineTerrain == nil {
			return combat.TilePoint{}, false, fmt.Errorf("Quick spell 0x%02X combat terrain projection is unavailable", spellID)
		}
		for _, target := range s.livingBySide(combat.SideEnemy) {
			if !target.HasCombatPosition || !s.combatLineTerrain(target.CombatX, target.CombatY).Valid {
				continue
			}
			// The PC-98 helper carries a candidate object pointer into
			// CASTCOMBATSPELL. The complete pointer-to-grid projection and
			// candidate tie/random policy are not closed yet; stable living
			// fighter order is the bounded adapter policy for zero MinRange.
			return combat.TilePoint{X: target.CombatX, Y: target.CombatY}, true, nil
		}
		return combat.TilePoint{}, false, nil
	}
	if s.combatScanMapProvider == nil {
		return combat.TilePoint{}, false, fmt.Errorf("Quick spell 0x%02X TACTICALMAP projection is unavailable", spellID)
	}
	tacticalMap, err := s.combatScanMapProvider()
	if err != nil {
		return combat.TilePoint{}, false, fmt.Errorf("build Quick spell 0x%02X TACTICALMAP: %w", spellID, err)
	}
	for _, target := range s.livingBySide(combat.SideEnemy) {
		if !target.HasCombatPosition {
			continue
		}
		center := combat.TilePoint{X: target.CombatX, Y: target.CombatY}
		ordered, err := s.battle.BuildLegacyAreaScanTargetIDs(
			tacticalMap, caster.ID,
			enginescan.Point{X: center.X, Y: center.Y}, combat.SideEnemy,
			int(minRange), 0xff,
		)
		if err != nil {
			return combat.TilePoint{}, false, fmt.Errorf("build Quick spell 0x%02X SCAN targets: %w", spellID, err)
		}
		if len(ordered) != 0 {
			return center, true, nil
		}
	}
	return combat.TilePoint{}, false, nil
}

// quickLineSpellTarget is the bounded CoAB adapter for a Quick line spell.
// The PC-98 target helper carries a candidate object pointer, while the
// complete pointer-to-grid projection and candidate tie/random policy remain
// unresolved. Until that consumer is closed, retain the stable living-enemy
// order and require the line-terrain projection before handing a point to the
// existing reflecting-line runtime.
func (s *State) quickLineSpellTarget(caster combat.Fighter) (combat.TilePoint, bool, error) {
	if s.combatLineTerrain == nil {
		return combat.TilePoint{}, false, fmt.Errorf("Quick Lightning Bolt combat terrain projection is unavailable")
	}
	for _, target := range s.livingBySide(combat.SideEnemy) {
		if !target.HasCombatPosition {
			continue
		}
		point := combat.TilePoint{X: target.CombatX, Y: target.CombatY}
		if !s.combatLineTerrain(point.X, point.Y).Valid {
			continue
		}
		return point, true, nil
	}
	return combat.TilePoint{}, false, nil
}

// quickTargetedSpellTarget projects the existing targeted cleric spell
// contracts into Quick. The original PC-98 object-pointer candidate order is
// not closed; the current adapter keeps the target list already used by the
// manual spell path and selects its first legal stable entry.
func (s *State) quickTargetedSpellTarget(caster combat.Fighter, spellID uint8) (combat.Fighter, bool, error) {
	var targets []combat.Fighter
	switch spellID {
	case CurseSpellID:
		targets = s.livingBySide(combat.SideEnemy)
	case CauseLightWoundsSpellID:
		targets = s.causeLightWoundsTargets(caster)
	case ProtectionFromEvilSpellID:
		targets = s.protectionFromEvilTargets(caster)
	case ProtectionFromGoodSpellID:
		targets = s.protectionFromGoodTargets(caster)
	default:
		return combat.Fighter{}, false, fmt.Errorf("Quick spell 0x%02X is not a targeted cleric spell", spellID)
	}
	if len(targets) == 0 {
		return combat.Fighter{}, false, nil
	}
	return targets[0], true, nil
}

func (s *State) quickSleepAreaTarget(caster combat.Fighter, minRange uint8) (combat.TilePoint, bool, error) {
	return s.quickAreaSpellTarget(caster, SleepSpellID, minRange)
}

// CombatSpellTargetsEnemy follows the global spell-table identity stored in
// player records. Magic Missile is 0x0F; cleric Protection from Good is 0x07.
func (s *State) CombatSpellTargetsEnemy() bool {
	switch s.combatCastingSpell {
	case MagicMissileSpellID:
		return true
	case CurseSpellID, CauseLightWoundsSpellID:
		return true
	default:
		return false
	}
}

// BeginCombatCast enters the RuleBook CAST target step without consuming a
// spell. Enter confirms it; Escape can cancel it in the renderer.
func (s *State) BeginCombatCast(spellID uint8) error {
	if spellID == BlessSpellID && !s.CombatCanCastBless() {
		return fmt.Errorf("Bless is unavailable")
	}
	if spellID == CurseSpellID && !s.CombatCanCastCurse() {
		return fmt.Errorf("Curse is unavailable")
	}
	if spellID == CauseLightWoundsSpellID && !s.CombatCanCastCauseLightWounds() {
		return fmt.Errorf("Cause Light Wounds is unavailable")
	}
	if spellID == ProtectionFromEvilSpellID && !s.CombatCanCastProtectionFromEvil() {
		return fmt.Errorf("Protection from Evil is unavailable")
	}
	if spellID == MagicMissileSpellID && !s.CombatCanCastMagicMissile() {
		return fmt.Errorf("Magic Missile is unavailable")
	}
	if spellID == SleepSpellID && !s.CombatCanCastSleep() {
		return fmt.Errorf("Sleep is unavailable")
	}
	if spellID == ProtectionFromGoodSpellID && !s.CombatCanCastProtectionFromGood() {
		return fmt.Errorf("Protection from Good is unavailable")
	}
	if spellID == FireballSpellID && !s.CombatCanCastFireball() {
		return fmt.Errorf("Fireball is unavailable")
	}
	if spellID == LightningBoltSpellID && !s.CombatCanCastLightningBolt() {
		return fmt.Errorf("Lightning Bolt is unavailable")
	}
	if spellID == StinkingCloudSpellID && !s.CombatCanCastStinkingCloud() {
		return fmt.Errorf("Stinking Cloud is unavailable")
	}
	if spellID == CloudkillSpellID && !s.CombatCanCastCloudkill() {
		return fmt.Errorf("Cloudkill is unavailable")
	}
	if spellID == CureLightWoundsSpellID && !s.CombatCanCastCureLightWounds() {
		return fmt.Errorf("Cure Light Wounds is unavailable")
	}
	if spellID != BlessSpellID && spellID != CurseSpellID && spellID != CauseLightWoundsSpellID && spellID != ProtectionFromEvilSpellID && spellID != ProtectionFromGoodSpellID && spellID != MagicMissileSpellID && spellID != SleepSpellID && spellID != StinkingCloudSpellID && spellID != CloudkillSpellID && spellID != FireballSpellID && spellID != LightningBoltSpellID && spellID != CureLightWoundsSpellID {
		return fmt.Errorf("spell 0x%02X is not implemented in combat", spellID)
	}
	s.combatCastingSpell = spellID
	s.combatSpellTargetsPoint = false
	if spellID == ProtectionFromGoodSpellID {
		_, ok := s.combatPartyTurn()
		if !ok {
			return fmt.Errorf("it is not a living party turn")
		}
		s.combatCastingClass, s.combatCastingClassSet = party.ClassCleric, true
		s.combatSpellTargetIndex = 0
		return nil
	}
	if spellID == BlessSpellID {
		s.combatSpellTargetIndex = 0
		return nil
	}
	if spellID == CurseSpellID {
		targets := s.livingBySide(combat.SideEnemy)
		if s.combatTargetIndex >= len(targets) {
			s.combatTargetIndex = 0
		}
		s.combatSpellTargetIndex = s.combatTargetIndex
		return nil
	}
	if spellID == CauseLightWoundsSpellID {
		s.combatSpellTargetIndex = 0
		return nil
	}
	if spellID == ProtectionFromEvilSpellID {
		s.combatSpellTargetIndex = 0
		return nil
	}
	if spellID == MagicMissileSpellID {
		targets := s.livingBySide(combat.SideEnemy)
		if s.combatTargetIndex >= len(targets) {
			s.combatTargetIndex = 0
		}
		s.combatSpellTargetIndex = s.combatTargetIndex
		return nil
	}
	if spellID == SleepSpellID || spellID == StinkingCloudSpellID || spellID == CloudkillSpellID || spellID == FireballSpellID || spellID == LightningBoltSpellID {
		targets := s.livingBySide(combat.SideEnemy)
		if len(targets) == 0 {
			return fmt.Errorf("combat has no living enemies")
		}
		if s.combatTargetIndex >= len(targets) {
			s.combatTargetIndex = 0
		}
		target := targets[s.combatTargetIndex]
		s.combatSpellTargetPoint = combat.TilePoint{X: target.CombatX, Y: target.CombatY}
		s.combatSpellTargetsPoint = true
		return nil
	}
	targets := s.combatHealingTargets()
	s.combatSpellTargetIndex = 0
	for index, target := range targets {
		if target.HitPoints < target.MaxHitPoints {
			s.combatSpellTargetIndex = index
			break
		}
	}
	return nil
}

func (s *State) CancelCombatCast() {
	s.combatCastingSpell = 0
	s.combatCastingClassSet = false
	s.combatSpellTargetIndex = 0
	s.combatSpellTargetsPoint = false
}

// clearCombatActionFor mirrors the reference per-player Action.Clear call
// for the UI-owned portion of the active turn. Battle owns the fighter's
// renderer-neutral CombatAction fields; State owns the pending target/cast
// selection and must clear it when that same fighter is downed.
func (s *State) clearCombatActionFor(fighterID string) {
	if s.combatTurnIndex >= len(s.combatTurns) || s.combatTurns[s.combatTurnIndex].FighterID != fighterID {
		return
	}
	s.CancelCombatCast()
	s.CancelCombatMove()
	s.EndCombatView()
}

func (s *State) CombatSelectSpellTarget(delta int) error {
	if s.combatCastingSpell == 0 {
		return fmt.Errorf("no spell target is being selected")
	}
	targets := s.CombatSpellTargets()
	if len(targets) == 0 {
		return fmt.Errorf("spell has no living targets")
	}
	s.combatSpellTargetIndex = (s.combatSpellTargetIndex + delta) % len(targets)
	if s.combatSpellTargetIndex < 0 {
		s.combatSpellTargetIndex += len(targets)
	}
	if s.CombatSpellTargetsEnemy() {
		s.combatTargetIndex = s.combatSpellTargetIndex
	}
	return nil
}

// CombatMoveSpellTarget moves an area-spell center on the original 32x16
// combat map. It is separate from fighter target cycling because Fireball may
// intentionally be centered on an empty tile and can harm either side.
func (s *State) CombatMoveSpellTarget(dx, dy int) error {
	if (s.combatCastingSpell != SleepSpellID && s.combatCastingSpell != StinkingCloudSpellID && s.combatCastingSpell != CloudkillSpellID && s.combatCastingSpell != FireballSpellID && s.combatCastingSpell != LightningBoltSpellID) || !s.combatSpellTargetsPoint {
		return fmt.Errorf("no area spell target is being selected")
	}
	next := combat.TilePoint{X: s.combatSpellTargetPoint.X + dx, Y: s.combatSpellTargetPoint.Y + dy}
	if next.X < 0 || next.X >= 32 || next.Y < 0 || next.Y >= 16 {
		return fmt.Errorf("spell target (%d,%d) is outside the combat map", next.X, next.Y)
	}
	s.combatSpellTargetPoint = next
	return nil
}

// CombatCast applies the verified first-level Magic Missile path. It consumes
// exactly one memorized slot, targets the current enemy selection, and then
// advances the same deterministic enemy-turn boundary as a weapon attack.
func (s *State) CombatCast(spellID uint8) error {
	return s.CombatCastWithTerrain(spellID, nil)
}

// ConfirmCombatCast commits the current manual CAST selection through the
// same CASTCOMBATSPELL delay handoff used by Quick combat. Point-targeted
// spells still require their own coordinate transaction and therefore remain
// on the immediate path until that original boundary is proven.
func (s *State) ConfirmCombatCast(terrain combat.LineTerrain) error {
	spellID := s.combatCastingSpell
	if spellID == 0 {
		return fmt.Errorf("no combat spell is being selected")
	}
	if s.dataPack == nil {
		return fmt.Errorf("combat spell casting metadata is unavailable")
	}
	definition, found := s.dataPack.FindCombatAISpell(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X has no casting metadata", spellID)
	}
	delay := enginequickspell.Spell{CastingTime: definition.CastingTime}.CastingDelayUnits()
	if delay == 0 {
		return s.CombatCastWithTerrain(spellID, terrain)
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	targetID := ""
	var err error
	if s.combatSpellTargetsPoint {
		err = s.battle.BeginPendingPointSpellAction(
			caster.ID, spellID, delay, s.combatSpellTargetPoint.X, s.combatSpellTargetPoint.Y,
		)
	} else if targets := s.CombatSpellTargets(); len(targets) > 0 {
		if s.combatSpellTargetIndex >= len(targets) {
			return fmt.Errorf("spell target index %d is unavailable", s.combatSpellTargetIndex)
		}
		targetID = targets[s.combatSpellTargetIndex].ID
		err = s.battle.BeginPendingTargetedSpellAction(caster.ID, spellID, delay, targetID)
	} else {
		err = s.battle.BeginPendingTargetedSpellAction(caster.ID, spellID, delay, "")
	}
	if err != nil {
		return err
	}
	s.CancelCombatCast()
	if s.combatDelayedTurns == nil {
		s.combatDelayedTurns = make(map[int]bool)
	}
	s.combatDelayedTurns[s.combatTurnIndex] = true
	s.combatTurnIndex++
	s.combatMessage = fmt.Sprintf(s.catalog.Text(
		"combat_quick_magic_casting", "combat_quick_magic_casting",
	), caster.Name, campSpellLabel(s.catalog, party.ClassCleric, spellID))
	return s.advanceCombatToParty()
}

// CombatCastWithTerrain keeps ordinary target spells terrain-neutral while
// allowing reflecting line effects to consume a title adapter's combat map.
func (s *State) CombatCastWithTerrain(spellID uint8, terrain combat.LineTerrain) error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	if spellID == BlessSpellID {
		return s.combatCastBless()
	}
	if spellID == CurseSpellID {
		return s.combatCastCurse()
	}
	if spellID == CauseLightWoundsSpellID {
		return s.combatCastCauseLightWounds()
	}
	if spellID == ProtectionFromEvilSpellID {
		return s.combatCastProtectionFromEvil()
	}
	if spellID == ProtectionFromGoodSpellID && s.combatSpellIsProtectionFromGood() {
		return s.combatCastProtectionFromGood()
	}
	if spellID == CureLightWoundsSpellID {
		return s.combatCastCureLightWounds()
	}
	if spellID == FireballSpellID {
		return s.combatCastFireball()
	}
	if spellID == SleepSpellID {
		return s.combatCastSleep()
	}
	if spellID == LightningBoltSpellID {
		return s.combatCastLightningBolt(terrain)
	}
	if spellID == StinkingCloudSpellID {
		return s.combatCastStinkingCloud(terrain)
	}
	if spellID == CloudkillSpellID {
		return s.combatCastCloudkill(terrain)
	}
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	if spellID != MagicMissileSpellID {
		return fmt.Errorf("spell 0x%02X is not implemented in combat", spellID)
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a magic-user in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == spellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Magic Missile", caster.ID)
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return s.finishCombat()
	}
	if s.combatTargetIndex >= len(enemies) {
		s.combatTargetIndex = 0
	}
	target := enemies[s.combatTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastMagicMissile(caster.ID, target.ID, casterLevel(s.partyRoster[characterIndex]))
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	s.CancelCombatCast()
	if result.Resisted {
		format := s.catalog.Text("combat_magic_resisted", "")
		if format != "" {
			s.combatMessage = fmt.Sprintf(format, caster.Name, target.Name, result.Missiles)
		} else {
			s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_magic_missile", ""), caster.Name, target.Name, result.Missiles, result.Damage)
		}
	} else {
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_magic_missile", ""), caster.Name, target.Name, result.Missiles, result.Damage)
	}
	if s.queueMagicMissileVisual(caster, target, result.Missiles, result.TargetHP <= 0) {
		return nil
	}
	s.combatTurnIndex++
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) combatCastLightningBolt(terrain combat.LineTerrain) error {
	if s.combatCastingSpell != LightningBoltSpellID || !s.combatSpellTargetsPoint {
		return fmt.Errorf("Lightning Bolt target is not being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a magic-user in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == LightningBoltSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Lightning Bolt", caster.ID)
	}
	target := s.combatSpellTargetPoint
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...,
	)
	result, err := s.battle.CastReflectingLineSpell(
		caster.ID, LightningBoltSpellID, target, casterLevel(s.partyRoster[characterIndex]),
		combat.ReflectingLineOptions{
			WeightedBudget: 14, FirstReflectionOriginThreshold: 8, FirstReflectionPenalty: 8,
			DamageFlags: combat.DamageFlagElectricity | combat.DamageFlagMagic,
		},
		terrain,
	)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, LightningBoltSpellID,
		)
		return err
	}
	impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	totalDamage := 0
	protectedCount, resistedCount := 0, 0
	for _, impact := range result.Impacts {
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID, To: impact.Point, Hit: true, Killed: impact.TargetHP <= 0,
			Damage: impact.Damage, Saved: impact.Saved, Resisted: impact.Resisted, Protected: impact.Protected,
		})
		totalDamage += impact.Damage
		if impact.Protected {
			protectedCount++
		}
		if impact.Resisted {
			resistedCount++
		}
	}
	segments := make([]combat.VisualPathSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		segments = append(segments, combat.VisualPathSegment{
			From: segment.From, To: segment.To,
			HasImpact: segment.HasImpact, ImpactIndex: segment.ImpactIndex,
		})
	}
	s.CancelCombatCast()
	messageID := "combat_lightning_bolt"
	fallback := messageID
	arguments := []any{caster.Name, len(result.Impacts), totalDamage}
	if protectedCount > 0 && resistedCount > 0 {
		messageID = "combat_lightning_bolt_protected_resisted"
		fallback = messageID
		arguments = append(arguments, protectedCount, resistedCount)
	} else if resistedCount > 0 {
		messageID = "combat_lightning_bolt_resisted"
		fallback = messageID
		arguments = append(arguments, resistedCount)
	} else if protectedCount > 0 {
		messageID = "combat_lightning_bolt_protected"
		fallback = messageID
		arguments = append(arguments, protectedCount)
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text(messageID, fallback), arguments...)
	if s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualLineSpell, Effect: "lightning_bolt", ActorID: caster.ID,
		From: combat.TilePoint{X: caster.CombatX, Y: caster.CombatY}, To: target,
		Hit: len(impacts) != 0, Impacts: impacts, TravelImpacts: result.TravelImpacts,
		Segments: segments,
	}) {
		return nil
	}
	s.combatTurnIndex++
	s.requestSound(SoundLightning)
	for range impacts {
		s.requestSound(SoundSpellHit)
	}
	for _, impact := range impacts {
		if impact.Killed {
			s.requestSound(SoundDead)
		}
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) combatCastStinkingCloud(terrain combat.LineTerrain) error {
	if s.combatCastingSpell != StinkingCloudSpellID || !s.combatSpellTargetsPoint {
		return fmt.Errorf("Stinking Cloud target is not being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a magic-user in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == StinkingCloudSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Stinking Cloud", caster.ID)
	}
	center := s.combatSpellTargetPoint
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...,
	)
	var areaTerrain combat.AreaTerrain
	if terrain != nil {
		areaTerrain = func(x, y int) bool {
			cell := terrain(x, y)
			return cell.Valid && !cell.Reflect
		}
	}
	result, err := s.battle.CastStinkingCloud(
		caster.ID, center, casterLevel(s.partyRoster[characterIndex]), areaTerrain,
	)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, StinkingCloudSpellID,
		)
		return err
	}
	impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	for _, impact := range result.Impacts {
		target, found := s.fighter(impact.TargetID)
		if !found {
			continue
		}
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID,
			To:       combat.TilePoint{X: target.CombatX, Y: target.CombatY},
			Hit:      true,
			Saved:    impact.Saved,
			Damage:   impact.HelplessTurns,
		})
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_stinking_cloud", "combat_stinking_cloud"),
		caster.Name, len(result.Impacts),
	)
	if s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualAreaSpell, Effect: "stinking_cloud", ActorID: caster.ID,
		From: combat.TilePoint{X: caster.CombatX, Y: caster.CombatY}, To: center,
		Hit: len(impacts) != 0, Impacts: impacts, PersistentAreaID: result.Area.ID,
	}) {
		return nil
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) combatCastCloudkill(terrain combat.LineTerrain) error {
	if s.combatCastingSpell != CloudkillSpellID || !s.combatSpellTargetsPoint {
		return fmt.Errorf("Cloudkill target is not being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex, spellIndex := -1, -1
	for index, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(party.ClassMagicUser) {
			continue
		}
		characterIndex = index
		for slot, memorized := range character.SpellSlots {
			if memorized == CloudkillSpellID {
				spellIndex = slot
				break
			}
		}
	}
	if characterIndex < 0 || spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Cloudkill", caster.ID)
	}
	center := s.combatSpellTargetPoint
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...,
	)
	var areaTerrain combat.AreaTerrain
	if terrain != nil {
		areaTerrain = func(x, y int) bool {
			cell := terrain(x, y)
			return cell.Valid && !cell.Reflect
		}
	}
	result, err := s.battle.CastCloudkill(
		caster.ID, center, casterLevel(s.partyRoster[characterIndex]), areaTerrain,
	)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, CloudkillSpellID)
		return err
	}
	impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	for _, impact := range result.Impacts {
		target, found := s.fighter(impact.TargetID)
		if !found && !impact.Killed {
			continue
		}
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID,
			To:       combat.TilePoint{X: center.X, Y: center.Y},
			Hit:      true, Saved: impact.Saved, Killed: impact.Killed,
		})
		if found {
			impacts[len(impacts)-1].To = combat.TilePoint{X: target.CombatX, Y: target.CombatY}
		}
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_cloudkill", "combat_cloudkill"),
		caster.Name, len(result.Impacts),
	)
	if s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualAreaSpell, Effect: "cloudkill", ActorID: caster.ID,
		From: combat.TilePoint{X: caster.CombatX, Y: caster.CombatY}, To: center,
		Hit: len(impacts) != 0, Impacts: impacts, PersistentAreaID: result.Area.ID,
	}) {
		return nil
	}
	s.combatTurnIndex++
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) combatCastSleep() error {
	if s.combatCastingSpell != SleepSpellID || !s.combatSpellTargetsPoint {
		return fmt.Errorf("Sleep target is not being selected")
	}
	if s.combatScanMapProvider == nil {
		return fmt.Errorf("Sleep TACTICALMAP projection is unavailable")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a magic-user in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == SleepSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Sleep", caster.ID)
	}
	tacticalMap, err := s.combatScanMapProvider()
	if err != nil {
		return fmt.Errorf("build Sleep TACTICALMAP: %w", err)
	}
	center := enginescan.Point{X: s.combatSpellTargetPoint.X, Y: s.combatSpellTargetPoint.Y}
	ordered, err := s.battle.BuildLegacyAreaScanTargetIDs(
		tacticalMap, caster.ID, center, combat.SideEnemy, 1, 0xff,
	)
	if err != nil {
		return fmt.Errorf("build Sleep SCAN targets: %w", err)
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...,
	)
	result, err := s.battle.CastSleepOrdered(
		caster.ID, ordered, casterLevel(s.partyRoster[characterIndex]),
	)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, SleepSpellID,
		)
		return err
	}
	resisted := 0
	visualImpacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	for _, impact := range result.Impacts {
		if impact.Resisted {
			resisted++
			continue
		}
		if target, found := s.battle.Fighter(impact.TargetID); found {
			visualImpacts = append(visualImpacts, combat.VisualImpactTarget{
				TargetID: target.ID,
				To:       combat.TilePoint{X: target.CombatX, Y: target.CombatY},
				Hit:      true,
			})
		}
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_sleep", "combat_sleep"),
		caster.Name, len(result.Impacts)-resisted, resisted,
	)
	if len(visualImpacts) != 0 && s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualTwinkle, Effect: "sleep", ActorID: caster.ID,
		From:    combat.TilePoint{X: caster.CombatX, Y: caster.CombatY},
		Impacts: visualImpacts,
	}) {
		return nil
	}
	s.combatTurnIndex++
	s.requestSound(SoundCast)
	for range visualImpacts {
		s.requestSound(SoundSpellHit)
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) combatCastFireball() error {
	if s.combatCastingSpell != FireballSpellID || !s.combatSpellTargetsPoint {
		return fmt.Errorf("Fireball target is not being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassMagicUser) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a magic-user in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == FireballSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Fireball", caster.ID)
	}
	positions := make(map[string]combat.TilePoint)
	for _, fighter := range s.battle.Fighters() {
		if fighter.HasCombatPosition {
			positions[fighter.ID] = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
		}
	}
	center := s.combatSpellTargetPoint
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...,
	)
	result, err := s.battle.CastFireball(caster.ID, center, casterLevel(s.partyRoster[characterIndex]))
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, FireballSpellID)
		return err
	}
	impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	totalDamage := 0
	protectedCount, resistedCount := 0, 0
	for _, impact := range result.Impacts {
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID:  impact.TargetID,
			To:        positions[impact.TargetID],
			Hit:       true,
			Killed:    impact.TargetHP <= 0,
			Damage:    impact.Damage,
			Saved:     impact.Saved,
			Resisted:  impact.Resisted,
			Protected: impact.Protected,
		})
		totalDamage += impact.Damage
		if impact.Protected {
			protectedCount++
		}
		if impact.Resisted {
			resistedCount++
		}
	}
	s.CancelCombatCast()
	messageID := "combat_fireball"
	fallback := messageID
	arguments := []any{caster.Name, len(result.Impacts), totalDamage}
	if protectedCount > 0 && resistedCount > 0 {
		messageID = "combat_fireball_protected_resisted"
		fallback = messageID
		arguments = append(arguments, protectedCount, resistedCount)
	} else if resistedCount > 0 {
		messageID = "combat_fireball_resisted"
		fallback = messageID
		arguments = append(arguments, resistedCount)
	} else if protectedCount > 0 {
		messageID = "combat_fireball_protected"
		fallback = messageID
		arguments = append(arguments, protectedCount)
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text(messageID, fallback), arguments...)
	if s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualAreaSpell, Effect: "fireball", ActorID: caster.ID,
		From: combat.TilePoint{X: caster.CombatX, Y: caster.CombatY},
		To:   center, Hit: len(impacts) != 0, Impacts: impacts,
	}) {
		return nil
	}
	s.combatTurnIndex++
	s.requestSound(SoundFireball)
	for range impacts {
		s.requestSound(SoundSpellHit)
	}
	for _, impact := range impacts {
		if impact.Killed {
			s.requestSound(SoundDead)
		}
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func (s *State) protectionFromEvilTargets(caster combat.Fighter) []combat.Fighter {
	targets := s.livingBySide(combat.SideParty)
	if !caster.HasCombatPosition {
		return targets
	}
	filtered := make([]combat.Fighter, 0, len(targets))
	for _, target := range targets {
		if target.ID == caster.ID {
			filtered = append(filtered, target)
			continue
		}
		if !target.HasCombatPosition {
			filtered = append(filtered, target)
			continue
		}
		dx := caster.CombatX - target.CombatX
		if dx < 0 {
			dx = -dx
		}
		dy := caster.CombatY - target.CombatY
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 && (dx != 0 || dy != 0) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func (s *State) protectionFromGoodTargets(caster combat.Fighter) []combat.Fighter {
	targets := s.livingBySide(combat.SideParty)
	if !caster.HasCombatPosition {
		return targets
	}
	filtered := make([]combat.Fighter, 0, len(targets))
	for _, target := range targets {
		if target.ID == caster.ID || !target.HasCombatPosition {
			filtered = append(filtered, target)
			continue
		}
		dx := caster.CombatX - target.CombatX
		if dx < 0 {
			dx = -dx
		}
		dy := caster.CombatY - target.CombatY
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 && (dx != 0 || dy != 0) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func (s *State) combatCasterClass(casterID string) (party.Class, bool) {
	for _, character := range s.partyRoster {
		if character.ID == casterID {
			return character.Class, true
		}
	}
	return 0, false
}

func (s *State) combatCasterHasClass(casterID string, class party.Class) bool {
	for _, character := range s.partyRoster {
		if character.ID == casterID {
			return character.HasClass(class)
		}
	}
	return false
}

func (s *State) combatSpellIsProtectionFromGood() bool {
	if s.combatCastingClassSet {
		return s.combatCastingClass == party.ClassCleric
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	_, ok = s.combatCasterClass(caster.ID)
	return ok && s.combatCasterHasClass(caster.ID, party.ClassCleric)
}

func (s *State) combatCastProtectionFromGood() error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != ProtectionFromGoodSpellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == ProtectionFromGoodSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Protection from Good", caster.ID)
	}
	targets := s.protectionFromGoodTargets(caster)
	if s.combatSpellTargetIndex < 0 || s.combatSpellTargetIndex >= len(targets) {
		return fmt.Errorf("no adjacent party member can receive Protection from Good")
	}
	target := targets[s.combatSpellTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	duration := 3 * casterLevel(s.partyRoster[characterIndex])
	_, err := s.battle.CastProtectionFromGood(caster.ID, target.ID, casterLevel(s.partyRoster[characterIndex]))
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, ProtectionFromGoodSpellID)
		return err
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_protection_from_good", "combat_protection_from_good"), caster.Name, target.Name, duration)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) combatCastProtectionFromEvil() error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != ProtectionFromEvilSpellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == ProtectionFromEvilSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Protection from Evil", caster.ID)
	}
	targets := s.protectionFromEvilTargets(caster)
	if s.combatSpellTargetIndex < 0 || s.combatSpellTargetIndex >= len(targets) {
		return fmt.Errorf("no adjacent party member can receive Protection from Evil")
	}
	target := targets[s.combatSpellTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	duration := 3 * casterLevel(s.partyRoster[characterIndex])
	_, err := s.battle.CastProtectionFromEvil(caster.ID, target.ID, casterLevel(s.partyRoster[characterIndex]))
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, ProtectionFromEvilSpellID)
		return err
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_protection_from_evil", "combat_protection_from_evil"), caster.Name, target.Name, duration)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) causeLightWoundsTargets(caster combat.Fighter) []combat.Fighter {
	targets := s.livingBySide(combat.SideEnemy)
	if !caster.HasCombatPosition {
		return targets
	}
	filtered := make([]combat.Fighter, 0, len(targets))
	for _, target := range targets {
		if !target.HasCombatPosition {
			filtered = append(filtered, target)
			continue
		}
		dx := caster.CombatX - target.CombatX
		if dx < 0 {
			dx = -dx
		}
		dy := caster.CombatY - target.CombatY
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 && (dx != 0 || dy != 0) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func (s *State) combatCastCauseLightWounds() error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != CauseLightWoundsSpellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == CauseLightWoundsSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Cause Light Wounds", caster.ID)
	}
	targets := s.causeLightWoundsTargets(caster)
	if s.combatSpellTargetIndex < 0 || s.combatSpellTargetIndex >= len(targets) {
		return fmt.Errorf("no adjacent enemy can receive Cause Light Wounds")
	}
	target := targets[s.combatSpellTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastCauseLightWounds(caster.ID, target.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, CauseLightWoundsSpellID)
		return err
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cause_light_wounds", "combat_cause_light_wounds"), caster.Name, target.Name, result.Damage)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) combatCastCurse() error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != CurseSpellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == CurseSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Curse", caster.ID)
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return s.finishCombat()
	}
	if s.combatSpellTargetIndex >= len(enemies) {
		s.combatSpellTargetIndex = 0
	}
	target := enemies[s.combatSpellTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastCurse(caster.ID, target.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, CurseSpellID)
		return err
	}
	s.CancelCombatCast()
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if result.Targets == 0 {
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_curse_immune", "combat_curse_immune"), caster.Name, target.Name)
	} else {
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_curse", "combat_curse"), caster.Name, target.Name)
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) combatCastBless() error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != BlessSpellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == BlessSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Bless", caster.ID)
	}
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastBless(caster.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, BlessSpellID)
		return err
	}
	s.CancelCombatCast()
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_bless", "combat_bless"), caster.Name)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	_ = result
	return s.advanceCombatToParty()
}

func (s *State) combatCastCureLightWounds() error {
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(party.ClassCleric) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q is not a cleric in the party roster", caster.ID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == CureLightWoundsSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized Cure Light Wounds", caster.ID)
	}
	targets := s.combatHealingTargets()
	targetIndex := s.combatSpellTargetIndex
	if s.combatCastingSpell == 0 {
		targetIndex = -1
		for index, target := range targets {
			if target.HitPoints < target.MaxHitPoints {
				targetIndex = index
				break
			}
		}
	}
	if targetIndex < 0 || targetIndex >= len(targets) || targets[targetIndex].HitPoints >= targets[targetIndex].MaxHitPoints {
		return fmt.Errorf("no wounded party member can receive Cure Light Wounds")
	}
	target := targets[targetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots[:spellIndex], s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastCureLightWounds(caster.ID, target.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, CureLightWoundsSpellID)
		return err
	}
	for index := range s.partyRoster {
		if s.partyRoster[index].ID == target.ID {
			s.partyRoster[index].HitPoints = result.TargetHP
		}
	}
	for index := range s.party {
		if s.party[index].ID == target.ID {
			s.party[index].HitPoints = result.TargetHP
		}
	}
	s.CancelCombatCast()
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cure_light_wounds", "combat_cure_light_wounds"), caster.Name, target.Name, result.Healing)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func casterLevel(character party.Character) int {
	if character.Level < 1 {
		return 1
	}
	return character.Level
}

// CombatAct executes the current party member's attack and advances through
// enemy turns until the next party member is ready. One input therefore maps
// to one player action while enemy behavior remains deterministic by seed.
func (s *State) CombatAct() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	if s.combatTurnIndex >= len(s.combatTurns) {
		return s.advanceCombatRound()
	}
	turn := s.combatTurns[s.combatTurnIndex]
	attacker, ok := s.fighter(turn.FighterID)
	if !ok || attacker.Side != combat.SideParty {
		return s.advanceCombatToParty()
	}
	if attacker.QuickFight || attacker.CombatAction.SpellID != 0 {
		return s.advanceCombatToParty()
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return s.finishCombat()
	}
	results, err := s.combatAttackSequence(attacker)
	if err != nil {
		return err
	}
	if len(results) == 1 {
		target, ok := s.fighter(results[0].TargetID)
		if !ok {
			return fmt.Errorf("attack target %q disappeared", results[0].TargetID)
		}
		s.combatMessage = formatAttackMessage(s.catalog, attacker, target, results[0])
	} else {
		s.combatMessage = formatMultiAttackMessage(s.catalog, attacker, results)
	}
	target, _ := s.fighter(results[0].TargetID)
	if s.queueAttackVisual(attacker, target, results) {
		return nil
	}
	s.combatTurnIndex++
	s.requestAttackSounds(results)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

// CombatDone ends the active party fighter's turn without attacking. It is
// the RuleBook Combat Menu DONE action and follows the same enemy-turn path
// as other completed party actions.
func (s *State) CombatDone() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	if s.combatMoveMode || s.combatCastingSpell != 0 || s.combatView {
		return fmt.Errorf("combat action is still being selected")
	}
	attacker, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	if err := s.battle.ClearAction(attacker.ID); err != nil {
		return err
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_done", "combat_done"), attacker.Name)
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) CombatMainMenuText() string {
	return s.catalog.Text("combat_menu_main", "combat_menu_main")
}

func (s *State) CombatHitPointsLabel() string {
	return s.catalog.Text("combat_hud_hit_points", "combat_hud_hit_points")
}

func (s *State) CombatArmorClassLabel() string {
	return s.catalog.Text("combat_hud_armor_class", "combat_hud_armor_class")
}

// CombatSelectionPrompt owns the localized selection transaction while the
// renderer remains responsible only for placement and color.
func (s *State) CombatSelectionPrompt() (string, bool) {
	if s.combatCastingSpell != 0 {
		if s.combatCastingSpell == BlessSpellID {
			return s.catalog.Text("combat_prompt_confirm_spell", "combat_prompt_confirm_spell"), true
		}
		key := "combat_prompt_select_fighter_target"
		switch s.combatCastingSpell {
		case FireballSpellID:
			key = "combat_prompt_select_fireball_center"
		case SleepSpellID:
			key = "combat_prompt_select_sleep_center"
		case LightningBoltSpellID:
			key = "combat_prompt_select_lightning_direction"
		case StinkingCloudSpellID:
			key = "combat_prompt_select_stinking_cloud_corner"
		case CloudkillSpellID:
			key = "combat_prompt_select_cloudkill_center"
		}
		return s.catalog.Text(key, key), true
	}
	if s.combatMoveMode {
		return fmt.Sprintf(s.catalog.Text("combat_prompt_move", "combat_prompt_move"), s.combatMoveRemaining), true
	}
	return "", false
}

func (s *State) CombatQuickSpellHints() []string {
	hints := make([]string, 0, 12)
	appendHint := func(available bool, key string) {
		if available {
			hints = append(hints, s.catalog.Text(key, key))
		}
	}
	appendHint(s.CombatCanCastMagicMissile(), "combat_hint_magic_missile")
	appendHint(s.CombatCanCastSleep(), "combat_hint_sleep")
	appendHint(s.CombatCanCastFireball(), "combat_hint_fireball")
	appendHint(s.CombatCanCastLightningBolt(), "combat_hint_lightning_bolt")
	appendHint(s.CombatCanCastStinkingCloud(), "combat_hint_stinking_cloud")
	appendHint(s.CombatCanCastCloudkill(), "combat_hint_cloudkill")
	appendHint(s.CombatCanCastCureLightWounds(), "combat_hint_cure_light_wounds")
	appendHint(s.CombatCanCastBless(), "combat_hint_bless")
	appendHint(s.CombatCanCastCurse(), "combat_hint_curse")
	appendHint(s.CombatCanCastCauseLightWounds(), "combat_hint_cause_light_wounds")
	appendHint(s.CombatCanCastProtectionFromEvil(), "combat_hint_protection_from_evil")
	appendHint(s.CombatCanCastProtectionFromGood(), "combat_hint_protection_from_good")
	return hints
}

func (s *State) CombatTargetStatus(name string) string {
	return fmt.Sprintf(s.catalog.Text("combat_target_status", "combat_target_status"), name)
}

func (s *State) CombatQuickStatus(hints []string) string {
	return fmt.Sprintf(s.catalog.Text("combat_quick_status", "combat_quick_status"), strings.Join(hints, "　"))
}

func (s *State) CombatDoneMenuText() string {
	options := make([]string, 0, 6)
	if s.CombatCanGuard() {
		options = append(options, s.catalog.Text("combat_menu_guard", "combat_menu_guard"))
	}
	options = append(options,
		s.catalog.Text("combat_menu_delay", "combat_menu_delay"),
		s.catalog.Text("combat_menu_quit", "combat_menu_quit"),
	)
	if s.CombatCanBandage() {
		options = append(options, s.catalog.Text("combat_menu_bandage", "combat_menu_bandage"))
	}
	options = append(options,
		s.catalog.Text("combat_menu_speed", "combat_menu_speed"),
		s.catalog.Text("combat_menu_exit", "combat_menu_exit"),
	)
	return strings.Join(options, "　")
}

func (s *State) CombatCanGuard() bool {
	fighter, ok := s.combatPartyTurn()
	return ok && s.battle.CanGuard(fighter.ID)
}

// CombatGuard arms the original one-shot adjacent-entry reaction and ends
// the selected fighter's current action.
func (s *State) CombatGuard() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	attacker, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	if err := s.battle.GuardAction(attacker.ID); err != nil {
		return err
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_guard", "combat_guard"), attacker.Name)
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) CombatCanBandage() bool {
	if !s.CombatActive() {
		return false
	}
	for _, character := range s.partyRoster {
		if character.HealthStatus != party.HealthStatusDying {
			continue
		}
		if fighter, ok := s.fighter(character.ID); ok && fighter.Side == combat.SideParty {
			return true
		}
	}
	return false
}

// CombatBandage changes only the first TeamList-order dying party member to
// unconscious, clears bleeding, then consumes the acting character's turn.
func (s *State) CombatBandage() error {
	if !s.CombatCanBandage() {
		return fmt.Errorf("no dying party member can be bandaged")
	}
	attacker, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	bandaged := ""
	for index := range s.partyRoster {
		character := &s.partyRoster[index]
		fighter, exists := s.fighter(character.ID)
		if !exists || fighter.Side != combat.SideParty || character.HealthStatus != party.HealthStatusDying {
			continue
		}
		character.HealthStatus = party.HealthStatusUnconscious
		character.Bleeding = 0
		bandaged = character.Name
		break
	}
	if err := s.battle.ClearAction(attacker.ID); err != nil {
		return err
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_bandage", "combat_bandage"), attacker.Name, bandaged)
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// CombatQuick delegates the currently selected PC to the existing combat AI
// without consuming a separate action.
func (s *State) CombatQuick() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	fighter, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	if err := s.battle.SetQuickFight(fighter.ID); err != nil {
		return err
	}
	return s.advanceCombatToParty()
}

// CombatQuickAll performs the original ALT+Q handoff. The currently selected
// fighter receives delay 20 and the complete TeamList becomes AI-controlled;
// Space may restore manually controllable PCs even while an action timeline
// is playing.
func (s *State) CombatQuickAll() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	fighter, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	if err := s.battle.SetAllQuickFight(fighter.ID); err != nil {
		return err
	}
	s.combatMessage = s.catalog.Text("combat_quick_all", "combat_quick_all")
	return s.advanceCombatToParty()
}

// CombatToggleQuickMagic mirrors the ALT+M combat flag. It permits manually
// controllable Quick fighters to enter the original spell-priority selector;
// NPCs bypass this flag in the reference runtime.
func (s *State) CombatToggleQuickMagic() (bool, error) {
	if !s.CombatActive() {
		return false, fmt.Errorf("combat is not active")
	}
	s.combatQuickMagic = !s.combatQuickMagic
	messageID, fallback := "combat_quick_magic_off", "combat_quick_magic_off"
	if s.combatQuickMagic {
		messageID, fallback = "combat_quick_magic_on", "combat_quick_magic_on"
	}
	s.combatMessage = s.catalog.Text(messageID, fallback)
	return s.combatQuickMagic, nil
}

func (s *State) CombatQuickMagicEnabled() bool { return s.combatQuickMagic }

func (s *State) CombatManualControl() int {
	if s.battle == nil {
		return 0
	}
	changed := s.battle.SetPlayerCharactersManual()
	if changed > 0 {
		// QuickFight is stored on the Player record, not only the transient
		// Battle view. Keep the persistent party projection aligned so a combat
		// continuation cannot silently re-enable Quick in the next encounter.
		s.syncPartyFromBattle()
		s.combatMessage = s.catalog.Text("combat_manual_control", "combat_manual_control")
	}
	return changed
}

func (s *State) CombatSpeed() uint8 { return uint8(s.combatSpeed) }

func (s *State) CombatSpeedSlower() bool {
	return s.combatSpeed.Slower()
}

func (s *State) CombatSpeedFaster() bool {
	return s.combatSpeed.Faster()
}

func (s *State) CombatSpeedMenuText() string {
	options := make([]string, 0, 3)
	if s.combatSpeed < 9 {
		options = append(options, s.catalog.Text("combat_speed_slower", "combat_speed_slower"))
	}
	if s.combatSpeed > 0 {
		options = append(options, s.catalog.Text("combat_speed_faster", "combat_speed_faster"))
	}
	options = append(options, s.catalog.Text("combat_menu_exit", "combat_menu_exit"))
	return fmt.Sprintf(s.catalog.Text("combat_speed_value", "combat_speed_value"), s.combatSpeed, strings.Join(options, "　"))
}

// CombatDelay defers the active party fighter to delay tier one while keeping
// the action in the current round. This is the D choice inside the original
// combat DONE submenu, not the top-level DONE action itself.
func (s *State) CombatDelay() error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	if s.combatMoveMode || s.combatCastingSpell != 0 || s.combatView {
		return fmt.Errorf("combat action is still being selected")
	}
	attacker, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	if err := s.battle.DelayAction(attacker.ID); err != nil {
		return err
	}
	if s.combatDelayedTurns == nil {
		s.combatDelayedTurns = make(map[int]bool)
	}
	s.combatDelayedTurns[s.combatTurnIndex] = true
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_delay", "combat_delay"), attacker.Name)
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// combatAttackSequence keeps the game adapter responsible for target cursor
// policy. If a target falls, remaining weapon attacks use the next living
// enemy at the same cursor position, matching the RuleBook's Aim behavior.
func (s *State) combatAttackSequence(attacker combat.Fighter) ([]combat.AttackResult, error) {
	attacks := attacker.AttacksPerTurn
	if attacks < 1 {
		attacks = 1
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return nil, nil
	}
	if s.combatTargetIndex >= len(enemies) {
		s.combatTargetIndex = 0
	}
	// Validate before the ammunition transaction so a rejected adjacent
	// missile attack cannot consume arrows or bolts.
	if err := s.battle.ValidateAttack(attacker.ID, enemies[s.combatTargetIndex].ID); err != nil {
		return nil, err
	}
	if err := s.consumeCombatAmmunition(attacker, attacks); err != nil {
		return nil, err
	}
	results := make([]combat.AttackResult, 0, attacks)
	for len(results) < attacks && s.battle.Status() == combat.StatusActive {
		enemies = s.livingBySide(combat.SideEnemy)
		if len(enemies) == 0 {
			break
		}
		if s.combatTargetIndex >= len(enemies) {
			s.combatTargetIndex = 0
		}
		sequence, err := s.battle.AttackSequence(attacker.ID, enemies[s.combatTargetIndex].ID)
		if err != nil {
			return nil, err
		}
		results = append(results, sequence...)
		if len(sequence) == 0 || sequence[len(sequence)-1].TargetHP > 0 {
			break
		}
	}
	return results, nil
}

func (s *State) consumeCombatAmmunition(attacker combat.Fighter, shots int) error {
	if attacker.AmmunitionType == 0 || len(s.ammunitionItemTypes) == 0 {
		return nil
	}
	for index := range s.partyRoster {
		if s.partyRoster[index].ID != attacker.ID {
			continue
		}
		return s.partyRoster[index].ConsumeAmmunition(attacker.AmmunitionType, shots, s.ammunitionItemTypes)
	}
	return fmt.Errorf("ammunition owner %q is not in party roster", attacker.ID)
}

func (s *State) advanceCombatToParty() error {
	// Every path that consumes a turn advances combatTurnIndex before coming
	// back here. Synchronize all completed Action.delay values in one place so
	// visibility effects do not depend on which UI/AI action ended the turn.
	if interrupted := s.consumeCombatSpellInterruptions(); interrupted != "" {
		if s.combatMessage == "" {
			s.combatMessage = interrupted
		} else {
			s.combatMessage += "\n" + interrupted
		}
	}
	for s.battle != nil && s.battle.Status() == combat.StatusActive {
		for index := 0; index < s.combatTurnIndex && index < len(s.combatTurns); index++ {
			if s.combatDelayedTurns[index] {
				continue
			}
			if err := s.battle.CompleteAction(s.combatTurns[index].FighterID); err != nil {
				return err
			}
		}
		if s.combatTurnIndex >= len(s.combatTurns) {
			if s.battle.DynamicInitiativeActive() {
				turn, ok, err := s.battle.NextScheduledTurn()
				if err != nil {
					return err
				}
				if ok {
					s.combatTurns = append(s.combatTurns, turn)
					continue
				}
			}
			return s.advanceCombatRound()
		}
		turn := s.combatTurns[s.combatTurnIndex]
		fighter, ok := s.fighter(turn.FighterID)
		if !ok || fighter.HitPoints <= 0 {
			s.combatTurnIndex++
			continue
		}
		if fighter.MonsterIsHeld() {
			// PC-98 effects 1Fh/33h/34h/35h share the effect-table
			// handler that calls CLEARACTION. Unlike PUTDAMAGE and Cloudkill
			// effect 44h, this path does not call the memorized-slot consumer
			// and therefore must not enqueue a SpellInterruption.
			if err := s.battle.ClearAction(fighter.ID); err != nil {
				return err
			}
			s.clearCombatActionFor(fighter.ID)
			s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_monster_held", "combat_monster_held"), fighter.Name)
			s.combatTurnIndex++
			continue
		}
		if fighter.CloudIncapacitated() {
			helpless := fighter.HelplessTurns > 0
			if _, err := s.battle.ConsumeCloudIncapacitation(fighter.ID); err != nil {
				return err
			}
			if helpless {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cloud_helpless", "combat_cloud_helpless"), fighter.Name)
			} else {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cloud_coughing", "combat_cloud_coughing"), fighter.Name)
			}
			s.combatTurnIndex++
			continue
		}
		if fighter.CombatAction.SpellID != 0 {
			return s.resolvePendingSpell(fighter)
		}
		if fighter.Side == combat.SideParty && !fighter.QuickFight {
			return nil
		}
		if fighter.QuickFight {
			if err := s.battle.BeginQuickFightAction(fighter.ID); err != nil {
				return err
			}
			fighter, _ = s.fighter(fighter.ID)
			handled, err := s.tryQuickSpell(fighter)
			if err != nil {
				return err
			}
			if handled {
				return nil
			}
		}
		targetSide := combat.SideParty
		if fighter.Side == combat.SideParty {
			targetSide = combat.SideEnemy
		}
		targets := s.livingBySide(targetSide)
		if len(targets) == 0 {
			return s.finishCombat()
		}
		if fighter.MonsterThrowsLightning() && s.battle.Round() < 4 && s.combatLineTerrain != nil {
			point := combat.TilePoint{}
			target, found, err := s.battle.SelectRangedCombatTarget(fighter.ID, targetSide, combat.TargetSelectionOptions{
				MaxRange: 10,
				Terrain:  s.combatLineTerrain,
				VisibleTo: func(observer, target combat.Fighter) bool {
					return target.VisibleTo(observer)
				},
			})
			if err != nil {
				return err
			}
			if found {
				point = combat.TilePoint{X: target.CombatX, Y: target.CombatY}
			}
			if err := s.castMonsterLightning(fighter, point); err != nil {
				return err
			}
			return nil
		}
		target, err := s.battle.SelectCombatTarget(fighter.ID, targetSide)
		if err != nil {
			return err
		}
		if hasMonsterMagicMissile(fighter) {
			result, spellErr := s.battle.CastMonsterMagicMissile(fighter.ID, target.ID)
			if spellErr == nil {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_monster_magic_missile", "combat_monster_magic_missile"), fighter.Name, target.Name, result.Damage)
				if s.queueMagicMissileVisual(fighter, target, result.Missiles, result.TargetHP <= 0) {
					return nil
				}
				s.combatTurnIndex++
				s.requestSound(SoundCast)
				s.requestSound(SoundSpellHit)
				if s.battle.Status() != combat.StatusActive {
					return s.finishCombat()
				}
				return s.advanceCombatToParty()
			}
		}
		var resolvedResults []combat.AttackResult
		if fighter.AttacksPerTurn > 1 {
			results, err := s.battle.AttackSequence(fighter.ID, target.ID)
			if err != nil {
				return err
			}
			if len(results) == 1 {
				target, ok := s.fighter(results[0].TargetID)
				if !ok {
					return fmt.Errorf("enemy attack target %q disappeared", results[0].TargetID)
				}
				s.combatMessage = formatAttackMessage(s.catalog, fighter, target, results[0])
			} else {
				s.combatMessage = formatMultiAttackMessage(s.catalog, fighter, results)
			}
			resolvedResults = results
		} else {
			result, err := s.battle.Attack(fighter.ID, target.ID)
			if err != nil {
				return err
			}
			resolvedTarget, ok := s.fighter(result.TargetID)
			if !ok {
				return fmt.Errorf("enemy attack target %q disappeared", result.TargetID)
			}
			s.combatMessage = formatAttackMessage(s.catalog, fighter, resolvedTarget, result)
			resolvedResults = []combat.AttackResult{result}
		}
		if s.queueAttackVisual(fighter, target, resolvedResults) {
			return nil
		}
		s.combatTurnIndex++
		s.requestAttackSounds(resolvedResults)
	}
	return s.finishCombat()
}

func (s *State) tryQuickSpell(fighter combat.Fighter) (bool, error) {
	if fighter.Side != combat.SideParty || (fighter.ControlMorale < 0x80 && !s.combatQuickMagic) {
		return false, nil
	}
	characterIndex := -1
	for index := range s.partyRoster {
		if s.partyRoster[index].ID == fighter.ID {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 || len(s.partyRoster[characterIndex].SpellSlots) == 0 {
		return false, nil
	}
	if s.dataPack == nil {
		return false, fmt.Errorf("quick spell game-pack metadata is unavailable")
	}
	spellID, found, err := s.battle.SelectQuickSpell(
		s.partyRoster[characterIndex].SpellSlots,
		func(id uint8) (enginequickspell.Spell, bool) {
			definition, ok := s.dataPack.FindCombatAISpell(id)
			return enginequickspell.Spell{
				ID: definition.SpellID, Priority: definition.Priority,
				CastOn: definition.CastOn, MinRange: definition.MinRange,
				CastingTime: definition.CastingTime,
			}, ok
		},
		func(spell enginequickspell.Spell, minimumPriority uint8) (bool, error) {
			if spell.ID == LightningBoltSpellID {
				_, ok, err := s.quickLineSpellTarget(fighter)
				return ok, err
			}
			if spell.MinRange != 0 || spell.ID == StinkingCloudSpellID || spell.ID == CloudkillSpellID {
				switch spell.ID {
				case SleepSpellID, FireballSpellID, StinkingCloudSpellID, CloudkillSpellID:
					_, ok, err := s.quickAreaSpellTarget(fighter, spell.ID, spell.MinRange)
					return ok, err
				}
				return false, fmt.Errorf(
					"quick spell 0x%02X requires unresolved area-safety predicate %d",
					spell.ID, spell.MinRange,
				)
			}
			if spell.ID == BlessSpellID {
				return s.CombatCanCastBless(), nil
			}
			if spell.ID == CureLightWoundsSpellID {
				_, ok := s.quickCureTarget(fighter)
				return ok && s.CombatCanCastCureLightWounds(), nil
			}
			switch spell.ID {
			case CurseSpellID:
				_, ok, err := s.quickTargetedSpellTarget(fighter, spell.ID)
				return ok && s.CombatCanCastCurse(), err
			case CauseLightWoundsSpellID:
				_, ok, err := s.quickTargetedSpellTarget(fighter, spell.ID)
				return ok && s.CombatCanCastCauseLightWounds(), err
			case ProtectionFromEvilSpellID:
				_, ok, err := s.quickTargetedSpellTarget(fighter, spell.ID)
				return ok && s.CombatCanCastProtectionFromEvil(), err
			case ProtectionFromGoodSpellID:
				_, ok, err := s.quickTargetedSpellTarget(fighter, spell.ID)
				return ok && s.CombatCanCastProtectionFromGood(), err
			}
			if spell.ID != MagicMissileSpellID {
				// Preserve the original choice probability. Unsupported cast
				// handoff is handled after selection instead of silently making
				// the spell unsuitable and changing later PRNG traffic.
				return len(s.livingBySide(combat.SideEnemy)) > 0, nil
			}
			return s.CombatCanCastMagicMissile(), nil
		},
	)
	if err != nil {
		if fighter.ControlMorale < 0x80 {
			s.battle.SetPlayerCharactersManual()
			s.syncPartyFromBattle()
			s.combatMessage = fmt.Sprintf(s.catalog.Text(
				"combat_quick_magic_metadata_missing",
				"combat_quick_magic_metadata_missing",
			), err.Error())
			return true, nil
		}
		return false, err
	}
	if !found {
		return false, nil
	}
	definition, _ := s.dataPack.FindCombatAISpell(spellID)
	selected := enginequickspell.Spell{CastingTime: definition.CastingTime}
	if spellID == BlessSpellID || spellID == CureLightWoundsSpellID {
		if err := s.BeginCombatCast(spellID); err != nil {
			return false, err
		}
		targetID := ""
		if spellID == CureLightWoundsSpellID {
			target, ok := s.quickCureTarget(fighter)
			if !ok {
				s.CancelCombatCast()
				return false, fmt.Errorf("quick Cure Light Wounds has no eligible target")
			}
			targetID = target.ID
		}
		if err := s.battle.BeginPendingTargetedSpellAction(fighter.ID, spellID, selected.CastingDelayUnits(), targetID); err != nil {
			return false, err
		}
		s.CancelCombatCast()
		if s.combatDelayedTurns == nil {
			s.combatDelayedTurns = make(map[int]bool)
		}
		s.combatDelayedTurns[s.combatTurnIndex] = true
		s.combatTurnIndex++
		s.combatMessage = fmt.Sprintf(s.catalog.Text(
			"combat_quick_magic_casting", "combat_quick_magic_casting",
		), fighter.Name, campSpellLabel(s.catalog, party.ClassCleric, spellID))
		return true, nil
	}
	if spellID == SleepSpellID || spellID == FireballSpellID ||
		spellID == StinkingCloudSpellID || spellID == CloudkillSpellID {
		center, ok, err := s.quickAreaSpellTarget(fighter, spellID, definition.MinRange)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("Quick spell 0x%02X has no legal area target", spellID)
		}
		if err := s.BeginCombatCast(spellID); err != nil {
			return false, err
		}
		s.combatSpellTargetPoint = center
		s.combatSpellTargetsPoint = true
		if delay := selected.CastingDelayUnits(); delay > 0 {
			if err := s.battle.BeginPendingPointSpellAction(
				fighter.ID, spellID, delay, center.X, center.Y,
			); err != nil {
				return false, err
			}
			s.CancelCombatCast()
			if s.combatDelayedTurns == nil {
				s.combatDelayedTurns = make(map[int]bool)
			}
			s.combatDelayedTurns[s.combatTurnIndex] = true
			s.combatTurnIndex++
			s.combatMessage = fmt.Sprintf(s.catalog.Text(
				"combat_quick_magic_casting", "combat_quick_magic_casting",
			), fighter.Name, campSpellLabel(s.catalog, party.ClassMagicUser, spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if spellID == CurseSpellID || spellID == CauseLightWoundsSpellID ||
		spellID == ProtectionFromEvilSpellID || spellID == ProtectionFromGoodSpellID {
		target, ok, err := s.quickTargetedSpellTarget(fighter, spellID)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("Quick spell 0x%02X has no legal targeted recipient", spellID)
		}
		if err := s.BeginCombatCast(spellID); err != nil {
			return false, err
		}
		targets := s.CombatSpellTargets()
		targetIndex := -1
		for index := range targets {
			if targets[index].ID == target.ID {
				targetIndex = index
				break
			}
		}
		if targetIndex < 0 {
			s.CancelCombatCast()
			return false, fmt.Errorf("Quick spell 0x%02X target %q is unavailable", spellID, target.ID)
		}
		s.combatSpellTargetIndex = targetIndex
		if s.CombatSpellTargetsEnemy() {
			s.combatTargetIndex = targetIndex
		}
		if delay := selected.CastingDelayUnits(); delay > 0 {
			if err := s.battle.BeginPendingTargetedSpellAction(
				fighter.ID, spellID, delay, target.ID,
			); err != nil {
				return false, err
			}
			s.CancelCombatCast()
			if s.combatDelayedTurns == nil {
				s.combatDelayedTurns = make(map[int]bool)
			}
			s.combatDelayedTurns[s.combatTurnIndex] = true
			s.combatTurnIndex++
			s.combatMessage = fmt.Sprintf(s.catalog.Text(
				"combat_quick_magic_casting", "combat_quick_magic_casting",
			), fighter.Name, campSpellLabel(s.catalog, party.ClassCleric, spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if spellID == LightningBoltSpellID {
		point, ok, err := s.quickLineSpellTarget(fighter)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("Quick Lightning Bolt has no legal line target")
		}
		if err := s.BeginCombatCast(spellID); err != nil {
			return false, err
		}
		s.combatSpellTargetPoint = point
		s.combatSpellTargetsPoint = true
		if delay := selected.CastingDelayUnits(); delay > 0 {
			if err := s.battle.BeginPendingPointSpellAction(
				fighter.ID, spellID, delay, point.X, point.Y,
			); err != nil {
				return false, err
			}
			s.CancelCombatCast()
			if s.combatDelayedTurns == nil {
				s.combatDelayedTurns = make(map[int]bool)
			}
			s.combatDelayedTurns[s.combatTurnIndex] = true
			s.combatTurnIndex++
			s.combatMessage = fmt.Sprintf(s.catalog.Text(
				"combat_quick_magic_casting", "combat_quick_magic_casting",
			), fighter.Name, campSpellLabel(s.catalog, party.ClassMagicUser, spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if spellID != MagicMissileSpellID {
		if fighter.ControlMorale < 0x80 {
			s.battle.SetPlayerCharactersManual()
			s.syncPartyFromBattle()
			s.combatMessage = fmt.Sprintf(s.catalog.Text(
				"combat_quick_magic_unsupported",
				"combat_quick_magic_unsupported",
			), spellID)
			return true, nil
		}
		return false, fmt.Errorf("NPC quick spell 0x%02X is not implemented", spellID)
	}
	target, err := s.battle.SelectCombatTarget(fighter.ID, combat.SideEnemy)
	if err != nil {
		return false, err
	}
	enemies := s.livingBySide(combat.SideEnemy)
	for index := range enemies {
		if enemies[index].ID == target.ID {
			s.combatTargetIndex = index
			break
		}
	}
	if err := s.BeginCombatCast(spellID); err != nil {
		return false, err
	}
	return true, s.CombatCast(spellID)
}

func (s *State) resolvePendingSpell(fighter combat.Fighter) error {
	spellID := fighter.CombatAction.SpellID
	switch spellID {
	case BlessSpellID, CurseSpellID, CureLightWoundsSpellID, CauseLightWoundsSpellID,
		ProtectionFromEvilSpellID, ProtectionFromGoodSpellID, FireballSpellID,
		LightningBoltSpellID, CloudkillSpellID, SleepSpellID:
	default:
		return fmt.Errorf("pending spell 0x%02X is not implemented", spellID)
	}
	if err := s.BeginCombatCast(spellID); err != nil {
		return err
	}
	resolved, targetID := uint8(0), ""
	if fighter.CombatAction.HasTargetPoint {
		var x, y int
		var hasPoint bool
		var err error
		resolved, x, y, hasPoint, err = s.battle.TakePendingPointSpellAction(fighter.ID)
		if err != nil {
			return err
		}
		if !hasPoint {
			return fmt.Errorf("pending point spell 0x%02X lost its target", spellID)
		}
		s.combatSpellTargetPoint = combat.TilePoint{X: x, Y: y}
		s.combatSpellTargetsPoint = true
	} else {
		var err error
		resolved, targetID, err = s.battle.TakePendingTargetedSpellAction(fighter.ID)
		if err != nil {
			return err
		}
	}
	if resolved != spellID {
		return fmt.Errorf("pending quick spell changed from 0x%02X to 0x%02X", spellID, resolved)
	}
	if targetID != "" {
		targets := s.CombatSpellTargets()
		targetIndex := -1
		for index := range targets {
			if targets[index].ID == targetID {
				targetIndex = index
				break
			}
		}
		if targetIndex < 0 {
			return fmt.Errorf("pending spell 0x%02X target %q is unavailable", spellID, targetID)
		}
		s.combatSpellTargetIndex = targetIndex
		if s.CombatSpellTargetsEnemy() {
			s.combatTargetIndex = targetIndex
		}
	}
	return s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
}

func (s *State) castMonsterLightning(caster combat.Fighter, point combat.TilePoint) error {
	origin := combat.TilePoint{X: caster.CombatX, Y: caster.CombatY}
	if point == origin || !s.combatLineTerrain(point.X, point.Y).Valid {
		s.combatMessage = fmt.Sprintf(s.catalog.Text(
			"combat_monster_lightning_bolt_no_target",
			"combat_monster_lightning_bolt_no_target",
		), caster.Name)
		s.combatTurnIndex++
		s.requestSound(SoundLightning)
		return s.advanceCombatToParty()
	}
	result, err := s.battle.CastReflectingLineSpell(
		caster.ID, LightningBoltSpellID, point, 1,
		combat.ReflectingLineOptions{
			WeightedBudget: 10, FirstReflectionOriginThreshold: 8, FirstReflectionPenalty: 8,
			DamageFlags:       combat.DamageFlagElectricity | combat.DamageFlagMagic,
			InitialDamageDice: 16, PathDamageDice: 16, DamageDiceSides: 6,
		},
		s.combatLineTerrain,
	)
	if err != nil {
		return err
	}
	impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
	totalDamage, protectedCount, resistedCount := 0, 0, 0
	for _, impact := range result.Impacts {
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID, To: impact.Point, Hit: true, Killed: impact.TargetHP <= 0,
			Damage: impact.Damage, Saved: impact.Saved, Resisted: impact.Resisted, Protected: impact.Protected,
		})
		totalDamage += impact.Damage
		if impact.Protected {
			protectedCount++
		}
		if impact.Resisted {
			resistedCount++
		}
	}
	segments := make([]combat.VisualPathSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		segments = append(segments, combat.VisualPathSegment{
			From: segment.From, To: segment.To,
			HasImpact: segment.HasImpact, ImpactIndex: segment.ImpactIndex,
		})
	}
	messageID := "combat_monster_lightning_bolt"
	fallback := messageID
	arguments := []any{caster.Name, len(result.Impacts), totalDamage}
	if protectedCount > 0 && resistedCount > 0 {
		messageID = "combat_monster_lightning_bolt_protected_resisted"
		fallback = messageID
		arguments = append(arguments, protectedCount, resistedCount)
	} else if resistedCount > 0 {
		messageID = "combat_monster_lightning_bolt_resisted"
		fallback = messageID
		arguments = append(arguments, resistedCount)
	} else if protectedCount > 0 {
		messageID = "combat_monster_lightning_bolt_protected"
		fallback = messageID
		arguments = append(arguments, protectedCount)
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text(messageID, fallback), arguments...)
	if s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualLineSpell, Effect: "lightning_bolt", ActorID: caster.ID,
		From: origin, To: point,
		Hit: len(impacts) != 0, Impacts: impacts, TravelImpacts: result.TravelImpacts,
		Segments: segments,
	}) {
		return nil
	}
	s.combatTurnIndex++
	s.requestSound(SoundLightning)
	for range impacts {
		s.requestSound(SoundSpellHit)
	}
	for _, impact := range impacts {
		if impact.Killed {
			s.requestSound(SoundDead)
		}
	}
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	return s.advanceCombatToParty()
}

func hasMonsterMagicMissile(fighter combat.Fighter) bool {
	if fighter.MonsterSpellUses[0] == 0 {
		return false
	}
	for _, spellID := range fighter.MonsterSpellIDs {
		if spellID == combat.MonsterMagicMissileSpellID {
			return true
		}
	}
	return false
}

func (s *State) requestAttackSounds(results []combat.AttackResult) {
	for _, result := range results {
		if result.Hit {
			s.requestSound(SoundHit)
		} else {
			s.requestSound(SoundMiss)
		}
		if result.TargetHP <= 0 {
			s.requestSound(SoundDead)
		}
	}
}

func (s *State) queueAttackVisual(attacker, target combat.Fighter, results []combat.AttackResult) bool {
	if len(results) == 0 {
		return false
	}
	kind := combat.VisualMelee
	if attacker.MissileWeapon {
		kind = combat.VisualMissile
	}
	hit, killed := false, false
	for _, result := range results {
		hit = hit || result.Hit
		killed = killed || result.TargetHP <= 0
	}
	return s.queueCombatVisual(combat.VisualEvent{
		Kind: kind, ActorID: attacker.ID, TargetID: target.ID,
		From: combat.TilePoint{X: attacker.CombatX, Y: attacker.CombatY},
		To:   combat.TilePoint{X: target.CombatX, Y: target.CombatY},
		Hit:  hit, Killed: killed, Projectiles: len(results),
	})
}

func (s *State) queueMagicMissileVisual(caster, target combat.Fighter, missiles int, killed bool) bool {
	return s.queueCombatVisual(combat.VisualEvent{
		Kind: combat.VisualMagicMissile, ActorID: caster.ID, TargetID: target.ID,
		From: combat.TilePoint{X: caster.CombatX, Y: caster.CombatY},
		To:   combat.TilePoint{X: target.CombatX, Y: target.CombatY},
		Hit:  true, Killed: killed, Projectiles: missiles,
	})
}

func (s *State) advanceCombatRound() error {
	if err := s.battle.BeginScheduledRound(); err != nil {
		return err
	}
	s.combatTurns = nil
	s.combatTurnIndex = 0
	s.combatDelayedTurns = make(map[int]bool)
	return s.advanceCombatToParty()
}

func (s *State) finishCombat() error {
	if s.battle == nil {
		return fmt.Errorf("combat is not initialized")
	}
	interrupted := s.consumeCombatSpellInterruptions()
	s.combatVisual = nil
	s.combatVisualElapsed = 0
	s.combatVisualAdvanceTurn = false
	s.CancelCombatCast()
	s.CancelCombatMove()
	s.EndCombatView()
	s.Mode = ModeEvent
	s.syncPartyFromBattle()
	s.removeTemporaryCombatAllies()
	s.eventReturnMode = s.combatReturnMode
	if s.eventReturnMode != ModeDungeon {
		s.eventReturnMode = ModeWilderness
	}
	s.OriginalEvent = "COMBAT"
	s.combatMessage = combatResultMessage(s.catalog, s.battle.Status())
	if interrupted != "" {
		s.combatMessage = interrupted + "\n" + s.combatMessage
	}
	s.Message = s.combatMessage
	if s.battle.Status() == combat.StatusPartyWon {
		if len(s.pendingTreasure) > 0 {
			if err := s.ResolveTreasureRequests(); err != nil {
				return err
			}
		}
		if len(s.pendingTreasureItems) > 0 {
			s.treasureResumeECL = s.session != nil && len(s.eclBlock) > 0
			s.enterTreasureMenu()
			return nil
		}
		if continued, err := s.continueECLAfterEngineBoundary(); err != nil {
			return err
		} else if continued {
			return nil
		}
	}
	return nil
}

// consumeCombatSpellInterruptions projects the original positive-damage
// PUTDAMAGE transaction into the title-owned memorized spell roster. Battle
// has already atomically cleared the pending action and target.
func (s *State) consumeCombatSpellInterruptions() string {
	if s.battle == nil {
		return ""
	}
	events := s.battle.TakeSpellInterruptions()
	if len(events) == 0 {
		return ""
	}
	messages := make([]string, 0, len(events))
	for _, event := range events {
		name := event.FighterID
		for index := range s.partyRoster {
			if s.partyRoster[index].ID != event.FighterID {
				continue
			}
			name = s.partyRoster[index].Name
			for slot, spellID := range s.partyRoster[index].SpellSlots {
				if spellID != event.SpellID {
					continue
				}
				s.partyRoster[index].SpellSlots = append(
					s.partyRoster[index].SpellSlots[:slot],
					s.partyRoster[index].SpellSlots[slot+1:]...,
				)
				break
			}
			break
		}
		messages = append(messages, fmt.Sprintf(s.catalog.Text(
			"combat_spell_interrupted", "combat_spell_interrupted",
		), name))
	}
	return strings.Join(messages, "\n")
}

func (s *State) removeTemporaryCombatAllies() {
	persistent := s.party[:0]
	for _, fighter := range s.party {
		if !fighter.TemporaryAlly {
			persistent = append(persistent, fighter)
		}
	}
	s.party = persistent
}

// continueECLAfterEngineBoundary resumes the runtime state saved after
// CMD_Combat dispatched combat, CityShop or Temple. Combat calls it after
// victory; ECL-backed services call it when their UI closes.
func (s *State) continueECLAfterEngineBoundary() (bool, error) {
	return s.continueECLAfterEngineBoundaryDepth(0)
}

func (s *State) continueECLAfterEngineBoundaryDepth(depth int) (bool, error) {
	if s.session == nil || len(s.eclBlock) == 0 {
		return false, nil
	}
	if depth >= 8 {
		return false, fmt.Errorf("ECL continuation exceeded 8 engine-only boundaries")
	}
	// Long post-combat scripts can build the next encounter before reaching
	// their next player-visible boundary. Mogion's victory continuation is
	// one such sequence and exceeds the short menu-resume budget.
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.RunInteractiveSeed(500, nil, s.eclSeed)
	if err != nil {
		return false, err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	s.eclBlock = s.session.CurrentData()
	if start, startErr := s.session.InitialEntry(); startErr == nil {
		s.eclStart = start
	}
	s.applyGeoMapLoad(result)
	s.applyLoadPieces(result)
	s.applyECLCallSignals(result)
	s.applySpellSignals(result)
	s.applyECLDamageSignals(result)
	if err := s.applyECLNPCSignals(result); err != nil {
		return false, err
	}
	if err := s.applyECLClockSignals(result); err != nil {
		return false, err
	}
	s.applyECLInventorySignals(result)
	s.applyECLTreasureSignals(result)
	s.applyECLRobSignals(result)
	if result.ShopRequested {
		return true, s.enterECLShop(result)
	}
	if result.TempleRequested {
		return true, s.enterECLTemple()
	}
	treasureReady := false
	if len(result.TreasureRequests) > 0 {
		deferUntilVictory := result.CombatRequested && len(result.MonsterSpawns) > 0
		if deferUntilVictory {
			treasureReady = false
		} else {
			beforeMoney := s.moneyPool
			beforeGems, beforeJewelry := s.treasureGems, s.treasureJewelry
			beforeItems := len(s.pendingTreasureItems)
			if err := s.ResolveTreasureRequests(); err != nil {
				s.Message = fmt.Sprintf(s.catalog.Text("treasure_assets_pending", "treasure_assets_pending"), err)
			} else {
				// Pure money/gem/jewelry TREASURE still owns a visible service
				// boundary, while an all-zero request does not add an empty UI.
				treasureReady = result.CombatRequested &&
					(s.moneyPool != beforeMoney ||
						s.treasureGems != beforeGems ||
						s.treasureJewelry != beforeJewelry ||
						len(s.pendingTreasureItems) != beforeItems)
			}
		}
	}
	s.applyCitySelection()
	if len(result.Text) > 0 {
		s.unlockJournalEntries(result.Text)
		s.Message = s.localizeECLText(result.Text)
	}

	if result.WaitingForString && len(result.StringInputRequests) > 0 {
		s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
		return true, nil
	}
	if result.PictureRequested {
		s.Mode = ModeEvent
		if !s.picturesEnabled {
			s.PictureRequested = false
			s.PictureBlock = result.PictureBlock
			s.OriginalEvent = "PICTURE"
			if handled, handleErr := s.continueAfterSuppressedPicture(result); handled || handleErr != nil {
				return handled, handleErr
			}
			return true, nil
		}
		s.PictureRequested = true
		s.PictureBlock = result.PictureBlock
		s.BigPictureRequested = result.BigPictureRequested
		if result.PictureHeadBlockSet {
			s.SceneHeadBlock = uint8(result.PictureHeadBlock)
		}
		s.SceneCharacterRequested = !result.BigPictureRequested && s.SceneHeadBlock != 0xFF
		if s.SceneCharacterRequested {
			s.SceneBodyBlock = uint8(result.PictureBlock)
		}
		s.OriginalEvent = "PICTURE"
		if result.CombatRequested || result.ShopRequested || result.TempleRequested ||
			result.WaitingForMenu || result.WaitingForString {
			pending := result
			pending.PictureRequested = false
			s.pendingPictureResult = &pending
		}
		if s.Message == "" {
			s.Message = s.catalog.Text("event_picture", "event_picture")
		}
		return true, nil
	}
	// TREASURE followed by COMBAT without monster spawns is the reference
	// treasure-service boundary, not a battle.
	if treasureReady {
		s.treasureResumeECL = true
		if s.eclMenuReturnMode == ModeDungeon {
			s.enterTreasureMenuFor(ModeDungeon)
		} else {
			s.enterTreasureMenu()
		}
		if hasMeaningfulECLText(result.Text) {
			s.Message = s.localizeECLText(result.Text)
		}
		return true, nil
	}
	if result.CombatRequested {
		records := s.monsterRecordsForCurrentECL()
		if len(result.MonsterSpawns) > 0 && len(s.party) > 0 && len(records) > 0 {
			if err := s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed); err != nil {
				return false, err
			}
			return true, nil
		}
		s.Mode = ModeEvent
		s.OriginalEvent = "COMBAT"
		s.Message = s.catalog.Text("combat_started", "combat_started")
		return true, nil
	}
	if handled, err := s.applyECLProgram(result); handled || err != nil {
		return handled, err
	}
	if result.WaitingForMenu && len(result.Menus) > 0 {
		menu := result.Menus[len(result.Menus)-1]
		s.Choices = make([]string, 0, len(menu.Options))
		s.currentOriginalChoices = append([]string(nil), menu.Options...)
		for _, option := range menu.Options {
			s.Choices = append(s.Choices, s.localizeOption(option))
		}
		if menu.Prompt != "" {
			s.Prompt = s.localizeECLPrompt(menu.Prompt)
		}
		s.Mode = ModeWilderness
		return true, nil
	}
	if result.Exited && !hasMeaningfulECLText(result.Text) {
		return false, nil
	}
	if !hasMeaningfulECLText(result.Text) && !result.PictureRequested && !result.CombatRequested &&
		!result.ShopRequested && !result.TempleRequested && !result.WaitingForMenu && !result.WaitingForString &&
		len(result.CallAddresses) > 0 {
		// CALL is an immediate engine adapter boundary. A combat continuation
		// may reach redraw/cleanup helpers before the next script-visible
		// event; consume those calls and keep the same ECL PC moving.
		return s.continueECLAfterEngineBoundaryDepth(depth + 1)
	}
	if !hasMeaningfulECLText(result.Text) && !result.PictureRequested && !result.CombatRequested &&
		!result.ShopRequested && !result.TempleRequested && !result.WaitingForMenu && !result.WaitingForString {
		// A quiet continuation is not a new world event. Restore the mode that
		// owned combat so dungeon encounters remain inside their map.
		s.Mode = s.combatReturnMode
		s.eventReturnMode = s.combatReturnMode
		return true, nil
	}
	s.Mode = ModeEvent
	s.eventReturnMode = s.combatReturnMode
	if len(result.Text) > 0 {
		s.OriginalEvent = result.Text[len(result.Text)-1]
	}
	return true, nil
}

func (s *State) syncPartyFromBattle() {
	if s.battle == nil || len(s.party) == 0 {
		return
	}
	for index := range s.party {
		if fighter, ok := s.fighter(s.party[index].ID); ok {
			s.party[index] = fighter
			for rosterIndex := range s.partyRoster {
				if s.partyRoster[rosterIndex].ID == fighter.ID {
					s.partyRoster[rosterIndex].HitPoints = fighter.HitPoints
					if fighter.MaxHitPoints > 0 {
						s.partyRoster[rosterIndex].MaxHitPoints = fighter.MaxHitPoints
					}
					break
				}
			}
		}
	}
}

func (s *State) fighter(id string) (combat.Fighter, bool) {
	for _, fighter := range s.battle.Fighters() {
		if fighter.ID == id {
			return fighter, true
		}
	}
	return combat.Fighter{}, false
}

func (s *State) livingBySide(side combat.Side) []combat.Fighter {
	if s.battle == nil {
		return nil
	}
	all := s.battle.Fighters()
	living := make([]combat.Fighter, 0, len(all))
	for _, fighter := range all {
		if fighter.Side == side && fighter.HitPoints > 0 {
			living = append(living, fighter)
		}
	}
	return living
}

func formatAttackMessage(catalog interface{ Text(string, string) string }, attacker, target combat.Fighter, result combat.AttackResult) string {
	if !result.Hit {
		return fmt.Sprintf(catalog.Text("combat_miss", "combat_miss"), attacker.Name, target.Name)
	}
	fireDamage, fireProtected := attackFireEffectSummary(result)
	if fireProtected {
		return fmt.Sprintf(catalog.Text("combat_hit_fire_protected", ""), attacker.Name, target.Name, result.Damage)
	}
	if fireDamage > 0 {
		return fmt.Sprintf(catalog.Text("combat_hit_with_fire", ""), attacker.Name, target.Name, result.Damage, fireDamage)
	}
	return fmt.Sprintf(catalog.Text("combat_hit", "combat_hit"), attacker.Name, target.Name, result.Damage)
}

func formatMultiAttackMessage(catalog interface{ Text(string, string) string }, attacker combat.Fighter, results []combat.AttackResult) string {
	hits, damage, fireDamage, fireProtected := 0, 0, 0, false
	for _, result := range results {
		if result.Hit {
			hits++
			damage += result.Damage
		}
		effectDamage, effectProtected := attackFireEffectSummary(result)
		fireDamage += effectDamage
		fireProtected = fireProtected || effectProtected
	}
	if fireDamage > 0 || fireProtected {
		return fmt.Sprintf(catalog.Text("combat_multi_attack_with_fire", ""), attacker.Name, len(results), hits, damage, fireDamage)
	}
	return fmt.Sprintf(catalog.Text("combat_multi_attack", "combat_multi_attack"), attacker.Name, len(results), hits, damage)
}

func attackFireEffectSummary(result combat.AttackResult) (damage int, protected bool) {
	for _, effect := range result.Effects {
		if effect.DamageFlags&combat.DamageFlagFire == 0 {
			continue
		}
		damage += effect.Damage
		protected = protected || effect.Protected
	}
	return damage, protected
}

func combatResultMessage(catalog interface{ Text(string, string) string }, status combat.Status) string {
	switch status {
	case combat.StatusPartyWon:
		return catalog.Text("combat_victory", "combat_victory")
	case combat.StatusEnemyWon:
		return catalog.Text("combat_defeat", "combat_defeat")
	default:
		return catalog.Text("combat_draw", "combat_draw")
	}
}
