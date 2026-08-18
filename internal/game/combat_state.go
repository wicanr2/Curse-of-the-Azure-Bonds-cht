package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	enginedamage "github.com/wicanr2/golden-box-remake-engine/combat/damage"
	enginemodifier "github.com/wicanr2/golden-box-remake-engine/combat/modifier"
	enginespell "github.com/wicanr2/golden-box-remake-engine/combat/monsterspell"
	engineposthit "github.com/wicanr2/golden-box-remake-engine/combat/posthit"
	enginequickspell "github.com/wicanr2/golden-box-remake-engine/combat/quickspell"
	enginequicktarget "github.com/wicanr2/golden-box-remake-engine/combat/quicktarget"
	engineresistance "github.com/wicanr2/golden-box-remake-engine/combat/resistance"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
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
	// 佈陣要避開站不上去的格：原作 `try_place_combatant` 是在候選格通過
	// occupancy 與 ground 檢查之後才寫座標（spec 88）。先前的 fallback 只算
	// 格號不看地面，開場就會有人站在牆裡或樹叢上。
	taken := map[combat.TilePoint]bool{}
	for _, fighter := range append(append([]combat.Fighter(nil), party...), enemies...) {
		if fighter.HasCombatPosition {
			taken[combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}] = true
		}
	}
	partyIndex := 0
	for _, fighter := range party {
		if fighter.Side != combat.SideParty {
			return fmt.Errorf("fighter %q is not marked as party", fighter.ID)
		}
		if !fighter.HasCombatPosition {
			tile := s.combatFormationTile(fighter.Side, partyIndex, taken)
			taken[tile] = true
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
			tile := s.combatFormationTile(fighter.Side, enemyIndex, taken)
			taken[tile] = true
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
		monsterSpellRules, err := s.dataPack.ResolveCombatMonsterSpellRules()
		if err != nil {
			return fmt.Errorf("resolve combat monster spell rules: %w", err)
		}
		battle.SetMonsterSpellRules(monsterSpellRules)
	} else {
		battle.SetDamageRules([]enginedamage.Rule(nil))
		battle.SetConditionalModifierRules([]enginemodifier.Rule(nil))
		battle.SetMagicResistanceRules([]engineresistance.Rule(nil))
		battle.SetPostHitRules([]engineposthit.Rule(nil))
		battle.SetMonsterSpellRules([]enginespell.Rule(nil))
	}
	if err := s.applyDataPackCombatModifiers(battle); err != nil {
		return err
	}
	if err := applyHouseRules(battle); err != nil {
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


// combatPlacementProbeLimit 是佈陣時最多往後試幾個編制格。3 欄 × 16 列是
// DUNGCOM fallback 戰鬥地圖的高度，足以讓整隊在一條縱帶裡找到落腳點。
const combatPlacementProbeLimit = 3 * 16

// combatFormationTile 回傳「真的站得上去」的編制格。順序是從自己的編制號
// 往後試，第一個**沒被佔用且地面可通行**的就用；一路試不到才退回原本那格。
//
// ⚠ 只補「候選格要能站」這一條。沒有 SETUP MONSTER 的距離與 occupancy 表
// 之前，掃描順序仍是本作自訂的 fallback，不宣稱與原作逐格相同。
// 沒裝地形投影時（例如單元測試自己組 Battle）行為與先前完全一樣。
func (s *State) combatFormationTile(side combat.Side, ordinal int, taken map[combat.TilePoint]bool) combat.TilePoint {
	tile := combat.FormationTile(side, ordinal)
	if s.combatMovementTerrain == nil {
		return tile
	}
	for probe := ordinal; probe < ordinal+combatPlacementProbeLimit; probe++ {
		candidate := combat.FormationTile(side, probe)
		if taken[candidate] {
			continue
		}
		if cost, ok := s.combatMovementTerrain(candidate.X, candidate.Y); ok && cost > 0 {
			return candidate
		}
	}
	return tile
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
	if s.dungeonLifecycleActive || s.Area.InDungeon || s.Mode == ModeDungeon || s.eclMenuReturnMode == ModeDungeon || s.eventReturnMode == ModeDungeon {
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
	pendingSoundStart := len(s.pendingSoundEvents)
	if err := s.StartCombat(party, enemies, seed); err != nil {
		return err
	}
	// PC-98 GAME.EXE names this transition COMBATFX (selector 14). The DOS
	// resource set has no corresponding extracted WAV, so the platform adapter
	// may safely ignore it while the semantic event remains available to exact
	// PC-98 SOUNDFX playback.
	s.pendingSoundEvents = append(s.pendingSoundEvents, SoundEvent(""))
	copy(s.pendingSoundEvents[pendingSoundStart+1:], s.pendingSoundEvents[pendingSoundStart:])
	s.pendingSoundEvents[pendingSoundStart] = SoundCombat
	return nil
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
	// 走出戰場邊界就是原作的逃跑（spec 799／1112）：Gold Box 沒有 FLEE 指令，
	// 邊界那一步本身就是嘗試脫離。
	if leavesCombatMap(caster, dx, dy) {
		return s.attemptCombatEscape(caster)
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

// SetCombatMovementTerrain 裝上繪製端擁有的戰鬥地圖投影，供 **AI 移動**使用。
//
// ★ 玩家移動的地形是每一步由前端傳進來的（`CombatMoveWithTerrain`），
// 但 AI 的一步不是玩家按出來的，沒有那個傳入點。兩者必須是**同一份地形**，
// 否則怪物會走玩家走不了的格子。
func (s *State) SetCombatMovementTerrain(terrain combat.MovementTerrain) {
	s.combatMovementTerrain = terrain
}

// approachMonsterTarget 跑 AI 的移動階段。回傳「走完之後打不打得到」與
// 「有沒有真的移動」。
//
// ⚠ 地形沒裝上時走空曠地形。那會讓怪物穿過牆——所以前端一定要叫
// `SetCombatMovementTerrain`；這裡不假裝有地形，也不因為沒有地形就不動，
// 後者會讓怪物永遠站著不動而看起來像「AI 還沒接」。
func (s *State) approachMonsterTarget(fighter, target combat.Fighter) (bool, bool, error) {
	if !fighter.HasCombatPosition || !target.HasCombatPosition {
		// 沒有座標的遭遇（合成 fixture、純劇情戰鬥）不走移動階段。
		return true, false, nil
	}
	if fighter.MovementAllowance <= 0 {
		// **這場遭遇沒有提供移動資料**，不是「這隻怪不會動」——原作的怪物記錄
		// 一律有移動率（`internal/monster` 從 MON 記錄帶進來）。合成 fixture
		// 沒填，所以在這裡直接讓它打。⚠ 真的遇到移動率 0 的正式怪物，
		// 這一行會讓它隔空攻擊；那時要修的是資料來源，不是這個條件。
		return true, false, nil
	}
	mode, err := s.battle.BeginMonsterTurn(fighter.ID)
	if err != nil {
		return false, false, err
	}
	approach, err := s.battle.MonsterApproach(fighter.ID, target.ID, mode, s.combatMovementTerrain)
	if err != nil {
		return false, false, err
	}
	return approach.InWeaponRange, len(approach.Steps) > 0, nil
}

// combatMapWidth／combatMapHeight 是原作戰鬥地圖的格數。走出這個範圍就是離場。
const (
	combatMapWidth  = 32
	combatMapHeight = 16
)

// leavesCombatMap 判斷這一步會不會踏出戰場。腳印大於一格的戰鬥員以左上角
// 計算——與 `MoveWithTerrainAndFreeAttacks` 的地形檢查同一個約定。
func leavesCombatMap(fighter combat.Fighter, dx, dy int) bool {
	if !fighter.HasCombatPosition {
		return false
	}
	x, y := fighter.CombatX+dx, fighter.CombatY+dy
	return x < 0 || x >= combatMapWidth || y < 0 || y >= combatMapHeight
}

// attemptCombatEscape 跑原作的逃跑判定並把結果講給玩家。
//
// ⚠ 只處理**隊伍**這一側。怪物逃跑是 `RE-07`／`ENG-08`：它們的決策在 AI 那一層，
// 不是玩家按方向鍵走出去。
func (s *State) attemptCombatEscape(fighter combat.Fighter) error {
	attempt, err := s.battle.AttemptEscape(fighter.ID)
	if err != nil {
		return err
	}
	if !attempt.Escaped {
		// 原作的 'Escape is blocked' 不接名字，訊息也不結束回合——
		// 擋住之後角色還站在原地，移動力沒有扣。
		s.combatMessage = s.catalog.Text("combat_escape_blocked", "combat_escape_blocked")
		return nil
	}
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_got_away", "combat_got_away"), fighter.Name)
	s.CancelCombatMove()
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
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

// combatCanCastProtectionSpell 是「這個編號的保護法術現在施得出來嗎」。
// 職業與編號都從 game pack 的宣告讀——寫死就會讓法師版（16／17）永遠是灰的。
func (s *State) combatCanCastProtectionSpell(spellID uint8,
	targets func(combat.Fighter) []combat.Fighter) bool {
	if !s.CombatActive() {
		return false
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return false
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(targets(caster)) == 0 {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(required) {
			continue
		}
		for _, memorized := range character.SpellSlots {
			if memorized == spellID {
				return true
			}
		}
	}
	return false
}

// CombatCanCastProtectionFromEvil 保留牧師版（法術 6）的既有介面。
func (s *State) CombatCanCastProtectionFromEvil() bool {
	return s.combatCanCastProtectionSpell(ProtectionFromEvilSpellID, s.protectionFromEvilTargets)
}

// CombatCanCastProtectionFromGood 保留牧師版（法術 7）的既有介面。
func (s *State) CombatCanCastProtectionFromGood() bool {
	return s.combatCanCastProtectionSpell(ProtectionFromGoodSpellID, s.protectionFromGoodTargets)
}

func (s *State) CombatCastingSpell() uint8 { return s.combatCastingSpell }

func (s *State) CombatSpellTargetIndex() int { return s.combatSpellTargetIndex }

func (s *State) CombatSpellTargetPoint() (combat.TilePoint, bool) {
	return s.combatSpellTargetPoint, s.combatSpellTargetsPoint
}

func (s *State) CombatSpellTargets() []combat.Fighter {
	definition, found := s.combatPlayerSpellDefinition(s.combatCastingSpell)
	if !found {
		return nil
	}
	switch definition.Behavior {
	case "cure_light_wounds":
		return s.combatHealingTargets()
	case "protection_from_good":
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.protectionFromGoodTargets(caster)
	case "cause_light_wounds":
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.causeLightWoundsTargets(caster)
	case "protection_from_evil":
		caster, ok := s.combatPartyTurn()
		if !ok {
			return nil
		}
		return s.protectionFromEvilTargets(caster)
	default:
		if definition.TargetMode == "enemy" {
			return s.livingBySide(combat.SideEnemy)
		}
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

const quickTargetRuleID = "coab.pc98.quick-target-candidate-chain"

const enemyPhysicalTargetRuleID = "coab.pc98.enemy-physical-target"

const combatActionRuleID = "coab.pc98.combat-action"

func (s *State) clearSameTeamActionTargetOnQuick() (bool, error) {
	if s.dataPack == nil {
		// Synthetic states from before the game-pack rule retain the verified
		// CoAB default; production packs must declare the policy explicitly.
		return true, nil
	}
	definition, found := s.dataPack.FindCombatActionRule(combatActionRuleID)
	if !found {
		return false, fmt.Errorf("combat action game-pack rule %q is unavailable", combatActionRuleID)
	}
	if err := definition.Validate(); err != nil {
		return false, err
	}
	return definition.ClearSameTeamOnQuick, nil
}

func (s *State) selectEnemyPhysicalTarget(attackerID string, targetSide combat.Side) (combat.Fighter, bool, error) {
	if s.combatScanMapProvider == nil || s.dataPack == nil {
		// Synthetic tests and packs from before the target producer rule retain
		// the bounded stable-ID fallback; the production CoAB pack declares the
		// terrain-aware rule below.
		if s.battle == nil {
			return combat.Fighter{}, false, fmt.Errorf("combat battle is unavailable")
		}
		target, err := s.battle.SelectCombatTarget(attackerID, targetSide)
		return target, err == nil, err
	}
	definition, found := s.dataPack.FindCombatTargetRule(enemyPhysicalTargetRuleID)
	if !found {
		return combat.Fighter{}, false, fmt.Errorf("enemy target game-pack rule %q is unavailable", enemyPhysicalTargetRuleID)
	}
	rule, err := definition.ToRule()
	if err != nil {
		return combat.Fighter{}, false, err
	}
	tacticalMap, err := s.combatScanMapProvider()
	if err != nil {
		return combat.Fighter{}, false, fmt.Errorf("build enemy target TACTICALMAP: %w", err)
	}
	target, selected, err := s.battle.SelectLegacyScanCombatTarget(
		attackerID, targetSide, combat.LegacyTargetSelectionOptions{
			TacticalMap:   tacticalMap,
			MaxRange:      definition.MaxRange,
			Arc:           definition.Arc,
			Rule:          rule,
			RetryWithXRay: definition.RetryWithXRay,
			VisibleTo: func(observer, target combat.Fighter) bool {
				return target.VisibleTo(observer)
			},
		},
	)
	if err != nil {
		return combat.Fighter{}, false, fmt.Errorf("select enemy SCAN target: %w", err)
	}
	return target, selected, nil
}

func (s *State) quickTargetRule() (enginequicktarget.Rule, error) {
	if s.dataPack == nil {
		return enginequicktarget.Rule{}, fmt.Errorf("Quick target game-pack metadata is unavailable")
	}
	definition, found := s.dataPack.FindCombatAITargetRule(quickTargetRuleID)
	if !found {
		return enginequicktarget.Rule{}, fmt.Errorf("Quick target game-pack rule %q is unavailable", quickTargetRuleID)
	}
	rule, err := definition.ToRule()
	if err != nil {
		return enginequicktarget.Rule{}, err
	}
	return rule, nil
}

// orderQuickTargetCandidates applies the data-pack-declared portion of the
// recovered PC-98 target boundary. The original pointer-chain tie/retry/
// random policy is represented by the separate selector below; this function
// only orders legal candidates by their preserved one-based legacy combat
// object identity.
func (s *State) orderQuickTargetCandidates(targets []combat.Fighter) ([]combat.Fighter, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	rule, err := s.quickTargetRule()
	if err != nil {
		return nil, err
	}
	candidates := make([]enginequicktarget.Candidate, len(targets))
	byID := make(map[string]combat.Fighter, len(targets))
	for index, target := range targets {
		candidates[index] = enginequicktarget.Candidate{ID: target.ID, LegacyObjectID: target.LegacyObjectID}
		byID[target.ID] = target
	}
	ordered, err := enginequicktarget.Order(candidates, rule)
	if err != nil {
		return nil, err
	}
	result := make([]combat.Fighter, len(ordered))
	for index, candidate := range ordered {
		result[index] = byID[candidate.ID]
	}
	return result, nil
}

func (s *State) quickTargetCandidates(side combat.Side) ([]combat.Fighter, error) {
	if s.battle == nil {
		return nil, nil
	}
	candidates := make([]combat.Fighter, 0)
	for _, fighter := range s.battle.FightersInCombatOrder() {
		if fighter.Side == side && fighter.HitPoints > 0 {
			candidates = append(candidates, fighter)
		}
	}
	return s.orderQuickTargetCandidates(candidates)
}

// selectQuickTargetCandidate applies the recovered bounded retry policy once,
// after spell suitability has already established that at least one legal
// candidate exists. Keeping this out of the suitability callback prevents a
// rejected spell slot from consuming the target helper's random roll twice.
func (s *State) selectQuickTargetCandidate(
	targets []combat.Fighter,
	suitable func(combat.Fighter, uint8) (bool, error),
) (combat.Fighter, bool, error) {
	if len(targets) == 0 {
		return combat.Fighter{}, false, nil
	}
	rule, err := s.quickTargetRule()
	if err != nil {
		return combat.Fighter{}, false, err
	}
	candidates := make([]enginequicktarget.Candidate, len(targets))
	byID := make(map[string]combat.Fighter, len(targets))
	for index, target := range targets {
		candidates[index] = enginequicktarget.Candidate{ID: target.ID, LegacyObjectID: target.LegacyObjectID}
		byID[target.ID] = target
	}
	selected, found, err := s.battle.SelectQuickTarget(candidates, rule, func(candidate enginequicktarget.Candidate, minimumPriority uint8) (bool, error) {
		target, ok := byID[candidate.ID]
		if !ok {
			return false, fmt.Errorf("Quick target candidate %q disappeared", candidate.ID)
		}
		return suitable(target, minimumPriority)
	})
	if err != nil || !found {
		return combat.Fighter{}, found, err
	}
	target, ok := byID[selected.ID]
	if !ok {
		return combat.Fighter{}, false, fmt.Errorf("Quick target candidate %q disappeared after selection", selected.ID)
	}
	return target, true, nil
}

// selectQuickTargetOne applies the projected legacy candidate order with one
// random draw. It is deliberately separate from the priority-retry selector:
// PC-98 Magic Missile uses the general target consumer, while the candidate
// producer and full pointer-chain tie policy remain outside this bounded
// adapter.
func (s *State) selectQuickTargetOne(targets []combat.Fighter) (combat.Fighter, bool, error) {
	if len(targets) == 0 {
		return combat.Fighter{}, false, nil
	}
	rule, err := s.quickTargetRule()
	if err != nil {
		return combat.Fighter{}, false, err
	}
	candidates := make([]enginequicktarget.Candidate, len(targets))
	byID := make(map[string]combat.Fighter, len(targets))
	for index, target := range targets {
		candidates[index] = enginequicktarget.Candidate{ID: target.ID, LegacyObjectID: target.LegacyObjectID}
		byID[target.ID] = target
	}
	selected, found, err := s.battle.SelectQuickTargetOne(candidates, rule)
	if err != nil || !found {
		return combat.Fighter{}, found, err
	}
	target, ok := byID[selected.ID]
	if !ok {
		return combat.Fighter{}, false, fmt.Errorf("Quick target candidate %q disappeared after single draw", selected.ID)
	}
	return target, true, nil
}

func (s *State) combatPlayerSpellDefinition(spellID uint8) (goldenbox.CombatPlayerSpellDefinition, bool) {
	if s.dataPack == nil {
		return goldenbox.CombatPlayerSpellDefinition{}, false
	}
	return s.dataPack.FindCombatPlayerSpell(spellID, "")
}

// combatPlayerSpellLabel keeps combat messages on the same stable message ID
// as the game-pack spell contract. The first-level camp catalog remains the
// compatibility fallback for imported packs created before this field existed.
func (s *State) combatPlayerSpellLabel(spellID uint8) string {
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if found && definition.MessageID != "" {
		if localized := s.catalog.Text(definition.MessageID, ""); localized != "" {
			return localized
		}
	}
	return campSpellLabel(s.catalog, spellID)
}

func (s *State) quickSpellPriority(spellID uint8) (uint8, error) {
	if s.dataPack == nil {
		return 0, fmt.Errorf("Quick spell game-pack metadata is unavailable")
	}
	definition, found := s.dataPack.FindCombatAISpell(spellID)
	if !found {
		return 0, fmt.Errorf("Quick spell metadata for 0x%02X is unavailable", spellID)
	}
	return definition.Priority, nil
}

func (s *State) quickAreaTargetLegal(caster combat.Fighter, spellID, minRange uint8, target combat.Fighter) (bool, error) {
	if !target.HasCombatPosition {
		return false, nil
	}
	if minRange == 0 {
		if spellID != CloudkillSpellID {
			return false, fmt.Errorf("Quick spell 0x%02X requires a nonzero scan range", spellID)
		}
		if s.combatLineTerrain == nil {
			return false, fmt.Errorf("Quick spell 0x%02X combat terrain projection is unavailable", spellID)
		}
		return s.combatLineTerrain(target.CombatX, target.CombatY).Valid, nil
	}
	if s.combatScanMapProvider == nil {
		return false, fmt.Errorf("Quick spell 0x%02X TACTICALMAP projection is unavailable", spellID)
	}
	tacticalMap, err := s.combatScanMapProvider()
	if err != nil {
		return false, fmt.Errorf("build Quick spell 0x%02X TACTICALMAP: %w", spellID, err)
	}
	center := combat.TilePoint{X: target.CombatX, Y: target.CombatY}
	ordered, err := s.battle.BuildLegacyAreaScanTargetIDs(
		tacticalMap, caster.ID,
		enginescan.Point{X: center.X, Y: center.Y}, combat.SideEnemy,
		int(minRange), 0xff,
	)
	if err != nil {
		return false, fmt.Errorf("build Quick spell 0x%02X SCAN targets: %w", spellID, err)
	}
	return len(ordered) != 0, nil
}

func (s *State) quickAreaSpellHasTarget(caster combat.Fighter, spellID, minRange uint8) (bool, error) {
	targets, err := s.quickTargetCandidates(combat.SideEnemy)
	if err != nil {
		return false, err
	}
	for _, target := range targets {
		ok, err := s.quickAreaTargetLegal(caster, spellID, minRange, target)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// quickAreaSpellTarget is the bounded CoAB adapter for the PC-98 Quick
// MinRange path. The reference suitability routine builds a SCAN candidate
// list around each possible combat target before it accepts the spell. The
// candidate order and bounded retry policy are supplied by the engine rule;
// the spell-specific geometry remains a CoAB adapter.
func (s *State) quickAreaSpellTarget(caster combat.Fighter, spellID, minRange uint8) (combat.TilePoint, bool, error) {
	targets, err := s.quickTargetCandidates(combat.SideEnemy)
	if err != nil {
		return combat.TilePoint{}, false, err
	}
	priority, err := s.quickSpellPriority(spellID)
	if err != nil {
		return combat.TilePoint{}, false, err
	}
	target, found, err := s.selectQuickTargetCandidate(targets, func(target combat.Fighter, minimumPriority uint8) (bool, error) {
		if priority < minimumPriority {
			return false, nil
		}
		return s.quickAreaTargetLegal(caster, spellID, minRange, target)
	})
	if err != nil || !found {
		return combat.TilePoint{}, found, err
	}
	return combat.TilePoint{X: target.CombatX, Y: target.CombatY}, true, nil
}

func (s *State) quickLineTargetLegal(target combat.Fighter) (bool, error) {
	if !target.HasCombatPosition {
		return false, nil
	}
	if s.combatLineTerrain == nil {
		return false, fmt.Errorf("Quick Lightning Bolt combat terrain projection is unavailable")
	}
	return s.combatLineTerrain(target.CombatX, target.CombatY).Valid, nil
}

func (s *State) quickLineSpellHasTarget() (bool, error) {
	targets, err := s.quickTargetCandidates(combat.SideEnemy)
	if err != nil {
		return false, err
	}
	for _, target := range targets {
		ok, err := s.quickLineTargetLegal(target)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// quickLineSpellTarget is the bounded CoAB adapter for a Quick line spell.
// The PC-98 target helper carries a candidate object pointer, while the
// candidate pointer-to-grid projection remains title-specific. The recovered
// priority retry is applied before handing a point to the existing
// reflecting-line runtime.
func (s *State) quickLineSpellTarget(caster combat.Fighter) (combat.TilePoint, bool, error) {
	targets, err := s.quickTargetCandidates(combat.SideEnemy)
	if err != nil {
		return combat.TilePoint{}, false, err
	}
	priority, err := s.quickSpellPriority(LightningBoltSpellID)
	if err != nil {
		return combat.TilePoint{}, false, err
	}
	target, found, err := s.selectQuickTargetCandidate(targets, func(target combat.Fighter, minimumPriority uint8) (bool, error) {
		if priority < minimumPriority {
			return false, nil
		}
		return s.quickLineTargetLegal(target)
	})
	if err != nil || !found {
		return combat.TilePoint{}, found, err
	}
	return combat.TilePoint{X: target.CombatX, Y: target.CombatY}, true, nil
}

func (s *State) quickTargetedSpellCandidates(caster combat.Fighter, spellID uint8) ([]combat.Fighter, error) {
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
		return nil, fmt.Errorf("Quick spell 0x%02X is not a targeted cleric spell", spellID)
	}
	return s.quickTargetCandidatesForFighters(targets)
}

func (s *State) quickTargetCandidatesForFighters(targets []combat.Fighter) ([]combat.Fighter, error) {
	return s.orderQuickTargetCandidates(targets)
}

func (s *State) quickTargetedSpellHasTarget(caster combat.Fighter, spellID uint8) (bool, error) {
	targets, err := s.quickTargetedSpellCandidates(caster, spellID)
	if err != nil {
		return false, err
	}
	return len(targets) != 0, nil
}

// quickTargetedSpellTarget projects the existing targeted cleric spell
// contracts into Quick. The target list remains the manual spell legality
// adapter; engine quicktarget supplies the recovered priority retry boundary.
func (s *State) quickTargetedSpellTarget(caster combat.Fighter, spellID uint8) (combat.Fighter, bool, error) {
	targets, err := s.quickTargetedSpellCandidates(caster, spellID)
	if err != nil {
		return combat.Fighter{}, false, err
	}
	priority, err := s.quickSpellPriority(spellID)
	if err != nil {
		return combat.Fighter{}, false, err
	}
	return s.selectQuickTargetCandidate(targets, func(target combat.Fighter, minimumPriority uint8) (bool, error) {
		return priority >= minimumPriority, nil
	})
}

func (s *State) quickSleepAreaTarget(caster combat.Fighter, minRange uint8) (combat.TilePoint, bool, error) {
	return s.quickAreaSpellTarget(caster, SleepSpellID, minRange)
}

// CombatSpellTargetsEnemy follows the target mode declared by the game pack,
// while the original numeric SpellID remains available for effect rules.
func (s *State) CombatSpellTargetsEnemy() bool {
	definition, found := s.combatPlayerSpellDefinition(s.combatCastingSpell)
	return found && definition.TargetMode == "enemy"
}

// BeginCombatCast enters the RuleBook CAST target step without consuming a
// spell. Enter confirms it; Escape can cancel it in the renderer.
func (s *State) BeginCombatCast(spellID uint8) error {
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	switch definition.Behavior {
	case "bless":
		if !s.CombatCanCastBless() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "curse":
		if !s.CombatCanCastCurse() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "cause_light_wounds":
		if !s.CombatCanCastCauseLightWounds() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "protection_from_evil":
		if !s.combatCanCastProtectionSpell(spellID, s.protectionFromEvilTargets) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "protection_from_good":
		if !s.combatCanCastProtectionSpell(spellID, s.protectionFromGoodTargets) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "magic_missile":
		if !s.CombatCanCastMagicMissile() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "sleep":
		if !s.CombatCanCastSleep() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "fireball":
		if !s.CombatCanCastFireball() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "lightning_bolt":
		if !s.CombatCanCastLightningBolt() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "stinking_cloud":
		if !s.CombatCanCastStinkingCloud() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "cloudkill":
		if !s.CombatCanCastCloudkill() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "cure_light_wounds":
		if !s.CombatCanCastCureLightWounds() {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "damage_dice_area":
		if !s.combatCanCastAreaDamageDice(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "morale_break_area":
		if !s.combatCanCastAreaMoraleBreak(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "slay_living":
		if !s.combatCanCastSlayLiving(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "cure_blindness":
		if !s.combatCanCastCureBlindness(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "apply_effect_codes":
		if !s.combatCanCastApplyEffectCodes(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "message_only":
		if !s.combatCanCastMessageOnly(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "dispel_magic":
		if !s.combatCanCastDispelMagic(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "remove_curse":
		if !s.combatCanCastRemoveCurse(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "dimension_door":
		if !s.combatCanCastDimensionDoor(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "fire_shield":
		if !s.combatCanCastFireShield(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "cone_of_cold":
		if !s.combatCanCastConeOfCold(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "restoration":
		if !s.combatCanCastRestoration(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "heal_dice", "damage_dice":
		if !s.combatCanCastSpellDice(spellID, definition.Behavior == "heal_dice") {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "effect":
		// 資料驅動那一批也要能走選目標這條路，否則三十幾支效果類法術
		// 只能從 `CombatCast` 直接施放，玩家挑不了目標。
		if !s.combatCanCastEffectSpell(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	case "protection_from_evil_10ft", "protection_from_good_10ft":
		if !s.combatCanCastProtectionRadius(spellID) {
			return fmt.Errorf("%s is unavailable", s.combatPlayerSpellLabel(spellID))
		}
	default:
		return fmt.Errorf("spell 0x%02X uses unsupported combat behavior %q", spellID, definition.Behavior)
	}
	s.combatCastingSpell = spellID
	s.combatSpellTargetsPoint = false
	if definition.Behavior == "protection_from_good" {
		_, ok := s.combatPartyTurn()
		if !ok {
			return fmt.Errorf("it is not a living party turn")
		}
		// 職業取自宣告：法術 7 是牧師、17 是法師。寫死牧師會讓法師版在
		// `CombatCast` 那一關被自己剛設的值擋下來。
		required, ok := combatSpellCasterClasses[definition.CasterClass]
		if !ok {
			return fmt.Errorf("spell 0x%02X declares unknown caster class %q",
				spellID, definition.CasterClass)
		}
		s.combatCastingClass, s.combatCastingClassSet = required, true
		s.combatSpellTargetIndex = 0
		return nil
	}
	if definition.TargetMode == "none" {
		s.combatSpellTargetIndex = 0
		return nil
	}
	if definition.TargetMode == "enemy" {
		targets := s.CombatSpellTargets()
		if s.combatTargetIndex >= len(targets) {
			s.combatTargetIndex = 0
		}
		s.combatSpellTargetIndex = s.combatTargetIndex
		return nil
	}
	if definition.TargetMode == "area_point" || definition.TargetMode == "line" {
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
	if definition.TargetMode != "party_member" {
		return fmt.Errorf("spell 0x%02X uses unsupported target mode %q", spellID, definition.TargetMode)
	}
	targets := s.CombatSpellTargets()
	s.combatSpellTargetIndex = 0
	if definition.Behavior == "cure_light_wounds" {
		for index, target := range targets {
			if target.HitPoints < target.MaxHitPoints {
				s.combatSpellTargetIndex = index
				break
			}
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
	definition, found := s.combatPlayerSpellDefinition(s.combatCastingSpell)
	if !found || (definition.TargetMode != "area_point" && definition.TargetMode != "line") || !s.combatSpellTargetsPoint {
		return fmt.Errorf("no area spell target is being selected")
	}
	next := combat.TilePoint{X: s.combatSpellTargetPoint.X + dx, Y: s.combatSpellTargetPoint.Y + dy}
	if next.X < 0 || next.X >= combatMapWidth || next.Y < 0 || next.Y >= combatMapHeight {
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
	definition, found := s.combatPlayerSpellDefinition(spellID)
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
	), caster.Name, s.combatPlayerSpellLabel(spellID))
	return s.advanceCombatToParty()
}

// CombatCastWithTerrain keeps ordinary target spells terrain-neutral while
// allowing reflecting line effects to consume a title adapter's combat map.
func (s *State) CombatCastWithTerrain(spellID uint8, terrain combat.LineTerrain) error {
	if !s.CombatActive() {
		return fmt.Errorf("combat is not active")
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	switch definition.Behavior {
	case "bless":
		return s.combatCastBless()
	case "curse":
		return s.combatCastCurse()
	case "cause_light_wounds":
		return s.combatCastCauseLightWounds()
	case "protection_from_evil_10ft", "protection_from_good_10ft":
		required, ok := combatSpellCasterClasses[definition.CasterClass]
		if !ok {
			return fmt.Errorf("spell 0x%02X declares unknown caster class %q",
				spellID, definition.CasterClass)
		}
		if !s.combatSpellCasterClassMatches(spellID) {
			return fmt.Errorf("%s requires a %s caster",
				s.combatPlayerSpellLabel(spellID), definition.CasterClass)
		}
		return s.combatCastProtectionRadius(spellID, required,
			definition.Behavior == "protection_from_evil_10ft")
	case "protection_from_evil", "protection_from_good":
		required, ok := combatSpellCasterClasses[definition.CasterClass]
		if !ok {
			return fmt.Errorf("spell 0x%02X declares unknown caster class %q",
				spellID, definition.CasterClass)
		}
		if !s.combatSpellCasterClassMatches(spellID) {
			return fmt.Errorf("%s requires a %s caster",
				s.combatPlayerSpellLabel(spellID), definition.CasterClass)
		}
		if definition.Behavior == "protection_from_evil" {
			return s.combatCastProtectionFromEvil(spellID, required)
		}
		return s.combatCastProtectionFromGood(spellID, required)
	case "cure_light_wounds":
		return s.combatCastCureLightWounds()
	case "fireball":
		return s.combatCastFireball()
	case "sleep":
		return s.combatCastSleep()
	case "lightning_bolt":
		return s.combatCastLightningBolt(terrain)
	case "stinking_cloud":
		return s.combatCastStinkingCloud(terrain)
	case "cloudkill":
		return s.combatCastCloudkill(terrain)
	case "effect":
		return s.combatCastEffectSpell(spellID)
	case "heal_dice":
		return s.combatCastSpellDice(spellID, true)
	case "damage_dice":
		return s.combatCastSpellDice(spellID, false)
	case "damage_dice_area":
		return s.combatCastAreaDamageDice(spellID)
	case "morale_break_area":
		return s.combatCastAreaMoraleBreak(spellID)
	case "slay_living":
		return s.combatCastSlayLiving(spellID)
	case "cure_blindness":
		return s.combatCastCureBlindness(spellID)
	case "apply_effect_codes":
		return s.combatCastApplyEffectCodes(spellID)
	case "message_only":
		return s.combatCastMessageOnly(spellID)
	case "dispel_magic":
		return s.combatCastDispelMagic(spellID)
	case "remove_curse":
		return s.combatCastRemoveCurse(spellID)
	case "dimension_door":
		return s.combatCastDimensionDoor(spellID)
	case "fire_shield":
		return s.combatCastFireShield(spellID)
	case "cone_of_cold":
		return s.combatCastConeOfCold(spellID)
	case "restoration":
		return s.combatCastRestoration(spellID)
	case "magic_missile":
		// The effect helper below still consumes the original spell-table ID;
		// the game-pack behavior token only selects this verified path.
	default:
		return fmt.Errorf("spell 0x%02X uses unsupported combat behavior %q", spellID, definition.Behavior)
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
		return fmt.Errorf("caster %q has no memorized spell 0x%02X", caster.ID, spellID)
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

// combatSpellCasterClassMatches 檢查施法者是不是這支法術宣告的職業。
//
// ★ 這一條原本寫死「必須是牧師」，因為當時只宣告了法術 7（牧師版）。
// 法術主表（spec 1111）裡防護善良有兩支：7 是牧師、17 是法師，效果碼同樣是
// `09h`。職業改成從宣告讀，7 的限制原封不動，17 也能宣告。
func (s *State) combatSpellCasterClassMatches(spellID uint8) bool {
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return false
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		// 宣告了看不懂的職業就擋下來——放行等於讓任何人都能施。
		return false
	}
	if s.combatCastingClassSet {
		return s.combatCastingClass == required
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	_, ok = s.combatCasterClass(caster.ID)
	return ok && s.combatCasterHasClass(caster.ID, required)
}

// combatSpellCasterClasses 是 game pack 的 `caster_class` 字串對到職業。
var combatSpellCasterClasses = map[string]party.Class{
	"cleric":     party.ClassCleric,
	"magic_user": party.ClassMagicUser,
}

// 防護邪惡／防護善良共用一條路。
//
// ★ 原本是兩份幾乎一樣的程式碼，各自寫死法術編號（6／7）與職業（牧師）。
// 法術主表（spec 1111）裡這兩個效果各有兩支：6／7 是牧師版，16／17 是法師版，
// 效果碼同樣是 `08h`／`09h`。寫死編號等於宣告了法師版也施不出來——
// game pack 上看起來已經接好，實際一施就報錯。
type protectionSpellPlan struct {
	spellID      uint8
	requiredClass party.Class
	targets      func(combat.Fighter) []combat.Fighter
	cast         func(casterID, targetID string, casterLevel int) error
	messageKey   string
	missingLabel string
}

func (s *State) combatCastProtectionSpell(plan protectionSpellPlan) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != plan.spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(plan.requiredClass) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q cannot cast %s", caster.ID, plan.missingLabel)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == plan.spellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized %s", caster.ID, plan.missingLabel)
	}
	targets := plan.targets(caster)
	if s.combatSpellTargetIndex < 0 || s.combatSpellTargetIndex >= len(targets) {
		return fmt.Errorf("no adjacent party member can receive %s", plan.missingLabel)
	}
	target := targets[s.combatSpellTargetIndex]
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	level := casterLevel(s.partyRoster[characterIndex])
	duration := 3 * level
	if err := plan.cast(caster.ID, target.ID, level); err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, plan.spellID)
		return err
	}
	// 共用的施法投射物（spec 1126）：保護法術沒有走 `combatFinishSpell`，
	// 所以這一段要自己排，否則這四支會是全表唯一沒有演出的法術。
	if casterID, from, to, ok := s.sharedSpellCastGeometry(); ok {
		s.queueCombatVisual(combat.VisualEvent{
			Kind: combat.VisualMagicMissile, Effect: "spell_cast_shared",
			ActorID: casterID, From: from, To: to,
		})
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text(plan.messageKey, plan.messageKey),
		caster.Name, target.Name, duration)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

func (s *State) combatCastProtectionFromGood(spellID uint8, requiredClass party.Class) error {
	return s.combatCastProtectionSpell(protectionSpellPlan{
		spellID: spellID, requiredClass: requiredClass,
		targets: s.protectionFromGoodTargets,
		cast: func(casterID, targetID string, level int) error {
			_, err := s.battle.CastProtectionFromGood(casterID, targetID, level)
			return err
		},
		messageKey:   "combat_protection_from_good",
		missingLabel: "Protection from Good",
	})
}

func (s *State) combatCastProtectionFromEvil(spellID uint8, requiredClass party.Class) error {
	return s.combatCastProtectionSpell(protectionSpellPlan{
		spellID: spellID, requiredClass: requiredClass,
		targets: s.protectionFromEvilTargets,
		cast: func(casterID, targetID string, level int) error {
			_, err := s.battle.CastProtectionFromEvil(casterID, targetID, level)
			return err
		},
		messageKey:   "combat_protection_from_evil",
		missingLabel: "Protection from Evil",
	})
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

// combatCastEffectSpell 是**資料驅動**的施法路：效果碼、持續時間、豁免類別
// 全部來自法術主表（spec 1111），一支法術不需要一段程式碼（spec 1117）。
//
// ⚠ 它只負責「把效果記上去」。效果碼本身的語意由 `internal/combat` 的
// `MonsterAffect*` 判讀——記得上去不等於解讀得了，覆蓋報告要分開講。
func (s *State) combatCastEffectSpell(spellID uint8) error {
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	spell, ok := gamepack.SpellByID(int(spellID))
	if !ok || spell.EffectID == 0 {
		return fmt.Errorf("spell 0x%02X has no effect in the original table", spellID)
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex, spellIndex := s.memorizedSpellSlot(caster.ID, spellID)
	if characterIndex < 0 {
		return fmt.Errorf("caster %q has no memorized spell 0x%02X", caster.ID, spellID)
	}
	targets := s.effectSpellTargets(caster, spellID, definition.TargetMode)
	if len(targets) == 0 {
		return fmt.Errorf("spell 0x%02X has no living target", spellID)
	}
	level := casterLevel(s.partyRoster[characterIndex])
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastEffectSpell(caster.ID, targets, combat.EffectSpellRequest{
		SpellID:      spellID,
		EffectKind:   uint8(spell.EffectID),
		Duration:     spell.PrimaryDuration(level, false),
		SaveKind:     spell.SaveKind,
		SaveCategory: spell.SaveCategory,
		CasterLevel:  level,
	})
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	applied := 0
	for _, impact := range result.Impacts {
		if impact.Applied {
			applied++
		}
	}
	s.CancelCombatCast()
	s.requestSound(SoundCast)
	if applied > 0 {
		s.requestSound(SoundSpellHit)
	}
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_effect_spell", "combat_effect_spell"),
		caster.Name, s.combatPlayerSpellLabel(spellID), applied)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// memorizedSpellSlot 找出誰記了這支法術、記在第幾格。
func (s *State) memorizedSpellSlot(casterID string, spellID uint8) (int, int) {
	for index, character := range s.partyRoster {
		if character.ID != casterID {
			continue
		}
		for slot, memorized := range character.SpellSlots {
			if memorized == spellID {
				return index, slot
			}
		}
		return -1, -1
	}
	return -1, -1
}

// effectSpellTargets 依 game pack 宣告的目標模式取目標。
//
// ⚠ 目標模式由 pack **手寫宣告**，不是從屬性表的 `+0Bh`（1／2）推出來的：
// 那兩個值的差別原作沒讀出來（spec 827），拿沒讀出來的欄位推「敵方／友方」
// 會得到一個自洽但沒有依據的規則。
func (s *State) effectSpellTargets(caster combat.Fighter, spellID uint8, mode string) []string {
	switch mode {
	case "none":
		return []string{caster.ID}
	case "party_member":
		living := s.livingBySide(combat.SideParty)
		targets := make([]string, 0, len(living))
		for _, fighter := range living {
			targets = append(targets, fighter.ID)
		}
		return targets
	case "enemy":
		enemies := s.livingBySide(combat.SideEnemy)
		if len(enemies) == 0 {
			return nil
		}
		index := s.combatTargetIndex
		if index >= len(enemies) {
			index = 0
		}
		return []string{enemies[index].ID}
	case "area_point":
		// 半徑取自法術主表；表裡沒寫就是 0，也就是**只有那一格**。
		// 恐懼術（84）與沉默術（25）就是這一類。
		radius := 0
		if entry, ok := gamepack.SpellByID(int(spellID)); ok {
			radius = entry.AreaRadius
		}
		targets := make([]string, 0)
		for _, fighter := range s.battle.Fighters() {
			if fighter.HitPoints <= 0 || fighter.Escaped || !fighter.HasCombatPosition {
				continue
			}
			if !combat.FighterWithinRadius(fighter, s.combatSpellTargetPoint, radius) {
				continue
			}
			targets = append(targets, fighter.ID)
		}
		return targets
	default:
		return nil
	}
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
	clearSameTeamTarget, err := s.clearSameTeamActionTargetOnQuick()
	if err != nil {
		return err
	}
	if err := s.battle.SetQuickFightWithPolicy(fighter.ID, clearSameTeamTarget); err != nil {
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
	clearSameTeamTarget, err := s.clearSameTeamActionTargetOnQuick()
	if err != nil {
		return err
	}
	if err := s.battle.SetAllQuickFightWithPolicy(fighter.ID, clearSameTeamTarget); err != nil {
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
		// 回合開頭的效果記錄寫入（`CHECKFX(07h)`，spec 1123）：纏繞術把移動率
		// 設成 0、妖火讓護甲變差。這些不是暫存修正，是直接改記錄。
		// ★ 原作在**分派到 AI 或玩家選單之前**就呼叫它（`overlay-08 entry#4`＝
		// COMBAT 單元的回合開始重設，spec 804：`CHECKFX(7)` 在 `p^[198h]` 那個
		// 二選一之前）。擺在「交還 UI」之後等於玩家操作的隊員永遠不會被套。
		// ⚠ 這一段在玩家還在選的期間會被重新走到（移動、選目標都會回到這裡），
		// 所以這個時機的修正必須是冪等的；`TestCanActTimingModifiersAreIdempotent`
		// 擋住之後有人在這個時機加上加減型修正。
		if _, err := s.battle.ApplyEffectRecordWrites(fighter.ID, combat.CheckFXCanAct); err != nil {
			return err
		}
		fighter, _ = s.fighter(fighter.ID)
		if fighter.Side == combat.SideParty && !fighter.QuickFight {
			return nil
		}
		// 自動換裝在 AI 回合開頭（spec 1120）。原作的換裝屬於 AI 模組，
		// 所以管的是「這一回合由電腦操作的人」——包含開了快速戰鬥的隊員。
		if _, err := s.autoEquipBeforeAITurn(fighter.ID); err != nil {
			return err
		}
		fighter, _ = s.fighter(fighter.ID)
		if fighter.Side == combat.SideEnemy {
			// 士氣檢定在 AI 回合開頭（spec 1122）：門檻跟著受傷程度走，
			// 過不了而且跑得掉就撤退，並印「驚慌逃竄」。
			if _, err := s.battle.CheckMorale(fighter.ID); err != nil {
				return err
			}
			fighter, _ = s.fighter(fighter.ID)
			if messageID, show := fighter.PanicMessageID(); show {
				s.combatMessage = fmt.Sprintf(s.catalog.Text(messageID, messageID), fighter.Name)
				s.combatTurnIndex++
				continue
			}
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
		monsterSpellRules := fighter.MonsterSpecialSpellRules(s.battle.Round())
		if len(monsterSpellRules) > 0 {
			monsterSpellRule := monsterSpellRules[0]
			point := combat.TilePoint{}
			if s.combatLineTerrain != nil {
				target, found, err := s.battle.SelectRangedCombatTarget(fighter.ID, targetSide, combat.TargetSelectionOptions{
					MaxRange: monsterSpellRule.TargetRange,
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
			}
			if err := s.castMonsterLightning(fighter, point, monsterSpellRule); err != nil {
				return err
			}
			return nil
		}
		// 原作先看射程內有沒有人（spec 838 §五）：有就從候選裡均勻隨機挑一個，
		// 不必移動；沒有才輪到既有的 SCAN 規則挑一個遠目標再走過去。
		target, found, err := s.battle.SelectInRangeTarget(fighter.ID, targetSide)
		if err != nil {
			return err
		}
		inRange := found
		if !found {
			target, found, err = s.selectEnemyPhysicalTarget(fighter.ID, targetSide)
			if err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("enemy %q has no reachable target", fighter.ID)
		}
		if fighter.Side == combat.SideEnemy && !inRange {
			// 原作的 AI 回合是「先走到打得到，再打」（spec 830／838）。
			// 走不到就這一回合只移動——不是站在原地隔空攻擊。
			reached, moved, err := s.approachMonsterTarget(fighter, target)
			if err != nil {
				return err
			}
			if !reached {
				if moved {
					fighter, _ = s.fighter(fighter.ID)
					s.combatMessage = fmt.Sprintf(
						s.catalog.Text("combat_moved", "combat_moved"),
						fighter.Name, fighter.CombatX, fighter.CombatY)
				}
				s.combatTurnIndex++
				continue
			}
			fighter, _ = s.fighter(fighter.ID)
		}
		if hasMonsterMagicMissile(fighter) && s.monsterWantsToCast(fighter, combat.MonsterMagicMissileSpellID) {
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
		if drained := s.applyLevelDrain(fighter, resolvedResults); drained != "" {
			s.combatMessage = drained
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
			playerSpell, playerSpellFound := s.combatPlayerSpellDefinition(spell.ID)
			if !playerSpellFound {
				return false, fmt.Errorf("quick spell 0x%02X is not declared in combat_player_spells", spell.ID)
			}
			if playerSpell.TargetMode == "line" {
				ok, err := s.quickLineSpellHasTarget()
				return ok, err
			}
			if playerSpell.TargetMode == "area_point" {
				ok, err := s.quickAreaSpellHasTarget(fighter, spell.ID, spell.MinRange)
				return ok, err
			}
			if spell.MinRange != 0 {
				return false, fmt.Errorf(
					"quick spell 0x%02X requires unresolved area-safety predicate %d",
					spell.ID, spell.MinRange,
				)
			}
			if playerSpell.Behavior == "bless" {
				return s.CombatCanCastBless(), nil
			}
			if playerSpell.Behavior == "cure_light_wounds" {
				_, ok := s.quickCureTarget(fighter)
				return ok && s.CombatCanCastCureLightWounds(), nil
			}
			switch playerSpell.Behavior {
			case "curse":
				ok, err := s.quickTargetedSpellHasTarget(fighter, spell.ID)
				return ok && s.CombatCanCastCurse(), err
			case "cause_light_wounds":
				ok, err := s.quickTargetedSpellHasTarget(fighter, spell.ID)
				return ok && s.CombatCanCastCauseLightWounds(), err
			case "protection_from_evil":
				ok, err := s.quickTargetedSpellHasTarget(fighter, spell.ID)
				return ok && s.CombatCanCastProtectionFromEvil(), err
			case "protection_from_good":
				ok, err := s.quickTargetedSpellHasTarget(fighter, spell.ID)
				return ok && s.CombatCanCastProtectionFromGood(), err
			}
			if playerSpell.Behavior != "magic_missile" {
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
	aiDefinition, _ := s.dataPack.FindCombatAISpell(spellID)
	playerDefinition, playerFound := s.combatPlayerSpellDefinition(spellID)
	if !playerFound {
		return false, fmt.Errorf("quick spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	selected := enginequickspell.Spell{CastingTime: playerDefinition.CastingTime}
	if playerDefinition.Behavior == "bless" || playerDefinition.Behavior == "cure_light_wounds" {
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
		), fighter.Name, s.combatPlayerSpellLabel(spellID))
		return true, nil
	}
	if playerDefinition.TargetMode == "area_point" {
		center, ok, err := s.quickAreaSpellTarget(fighter, spellID, aiDefinition.MinRange)
		if err != nil {
			return false, err
		}
		if !ok {
			// The recovered target helper may exhaust its bounded priority
			// passes after the spell selector has already chosen a slot. Keep
			// the slot untouched and let the normal combat action continue;
			// this is distinct from an adapter/data error above.
			return false, nil
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
			), fighter.Name, s.combatPlayerSpellLabel(spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if (playerDefinition.TargetMode == "enemy" || playerDefinition.TargetMode == "party_member") &&
		playerDefinition.Behavior != "magic_missile" {
		target, ok, err := s.quickTargetedSpellTarget(fighter, spellID)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
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
			), fighter.Name, s.combatPlayerSpellLabel(spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if playerDefinition.TargetMode == "line" {
		point, ok, err := s.quickLineSpellTarget(fighter)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
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
			), fighter.Name, s.combatPlayerSpellLabel(spellID))
			return true, nil
		}
		return true, s.CombatCastWithTerrain(spellID, s.combatLineTerrain)
	}
	if playerDefinition.Behavior != "magic_missile" {
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
	candidates, err := s.quickTargetCandidates(combat.SideEnemy)
	if err != nil {
		return false, err
	}
	target, found, err := s.selectQuickTargetOne(candidates)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
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
	if _, found := s.combatPlayerSpellDefinition(spellID); !found {
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

func (s *State) castMonsterLightning(caster combat.Fighter, point combat.TilePoint, rule enginespell.Rule) error {
	origin := combat.TilePoint{X: caster.CombatX, Y: caster.CombatY}
	if s.combatLineTerrain == nil || point == origin || !s.combatLineTerrain(point.X, point.Y).Valid {
		s.combatMessage = fmt.Sprintf(s.catalog.Text(
			"combat_monster_lightning_bolt_no_target",
			"combat_monster_lightning_bolt_no_target",
		), caster.Name)
		s.combatTurnIndex++
		s.requestSound(SoundLightning)
		return s.advanceCombatToParty()
	}
	result, err := s.battle.CastReflectingLineSpell(
		caster.ID, rule.SpellID, point, rule.CasterLevel,
		combat.ReflectingLineOptions{
			WeightedBudget: rule.LineBudget, FirstReflectionOriginThreshold: rule.FirstReflectionOriginThreshold,
			FirstReflectionPenalty: rule.FirstReflectionPenalty,
			DamageFlags:            rule.DamageMask,
			InitialDamageDice:      rule.InitialDamageDice, PathDamageDice: rule.PathDamageDice,
			DamageDiceSides: rule.DamageDiceSides,
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

// monsterWantsToCast 跑原作的施法決策（spec 836）並回答「這一回合它想不想放
// 這一支法術」。
//
// ★ 在這之前只要身上有魔法飛彈就一定放。原作不是這樣：AI 每回合擲 1d7 決定
// 「肯將就到什麼程度」，門檻從 7 往下降，法術要靠屬性表 `+0Dh` 的分數過門檻
// （spec 802／1111）。分數低的法術在運氣好的回合根本不會被考慮。
//
// ⚠ 決策選到 remake 還沒實作的法術時，這裡回 false，回合就退回物理攻擊——
// **不是把回合丟掉**。缺的是那支法術的效果（`ENG-09`），不是決策。
func (s *State) monsterWantsToCast(fighter combat.Fighter, spellID uint8) bool {
	if s.battle == nil {
		return false
	}
	choice, found, err := s.battle.AIChooseSpell(fighter.ID, func(id uint8) (int, bool) {
		spell, ok := gamepack.SpellByID(int(id))
		if !ok || spell.Placeholder {
			return 0, false
		}
		return spell.AIPriority, true
	})
	if err != nil || !found {
		return false
	}
	return choice == spellID
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
	case combat.StatusPartyFled:
		return catalog.Text("combat_party_fled", "combat_party_fled")
	default:
		return catalog.Text("combat_draw", "combat_draw")
	}
}
