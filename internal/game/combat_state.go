package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// StartCombat creates the first playable battle adapter. Party and encounter
// records are supplied by the data layer, so this state machine does not
// invent MON*CHA statistics or DOS party layout.
func (s *State) StartCombat(party, enemies []combat.Fighter, seed int64) error {
	if len(party) == 0 || len(enemies) == 0 {
		return fmt.Errorf("combat needs at least one party member and enemy")
	}
	fighters := make([]combat.Fighter, 0, len(party)+len(enemies))
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
		fighters = append(fighters, fighter)
		enemyIndex++
	}
	battle, err := combat.NewBattle(fighters, seed)
	if err != nil {
		return err
	}
	if err := s.applyDataPackCombatModifiers(battle); err != nil {
		return err
	}
	turns, err := battle.StartRound()
	if err != nil {
		return err
	}
	s.battle = battle
	if err := s.SetParty(party); err != nil {
		return err
	}
	s.combatTurns = turns
	s.combatTurnIndex = 0
	s.combatTargetIndex = 0
	s.combatVisual = nil
	s.combatMessage = s.catalog.Text("combat_started", "戰鬥開始！")
	s.Mode = ModeCombat
	return s.advanceCombatToParty()
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
	applyCombatTeamWrites(result.MonsterSpawns, result.CombatTeamWrites, len(party))
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
		enemies[index].Name = localizeMonsterName(s.catalog, enemies[index].Name)
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

func localizeMonsterName(catalog locale.Catalog, name string) string {
	switch strings.TrimSpace(strings.ToUpper(name)) {
	case "HIPPOGRIFF":
		return catalog.Text("monster_hippogriff", "鷹馬")
	case "FIGHTER":
		return catalog.Text("monster_fighter", "戰士")
	case "BLACK DRAGON":
		return catalog.Text("monster_black_dragon", "黑龍")
	case "DK ELF FIGHTER":
		return catalog.Text("monster_dark_elf_fighter", "黑暗精靈戰士")
	case "DARK ELF MAGE":
		return catalog.Text("monster_dark_elf_mage", "黑暗精靈法師")
	case "DARK ELF CLERIC":
		return catalog.Text("monster_dark_elf_cleric", "黑暗精靈牧師")
	case "EFREETI":
		return catalog.Text("monster_efreeti", "伊弗利特")
	case "SALAMANDER":
		return catalog.Text("monster_salamander", "火蜥蜴")
	case "DRACOLICH":
		return catalog.Text("monster_dracolich", "龍巫妖")
	case "HOODED MEDUSA":
		return catalog.Text("monster_hooded_medusa", "兜帽梅杜莎")
	case "BEHOLDER":
		return catalog.Text("monster_beholder", "眼魔")
	case "MINOTAUR":
		return catalog.Text("monster_minotaur", "牛頭人")
	case "ZHENTIL FIGHTER":
		return catalog.Text("monster_zhentil_fighter", "散提爾堡戰士")
	case "ZHENTIL MAGE":
		return catalog.Text("monster_zhentil_mage", "散提爾堡法師")
	case "ZHENTIL CLERIC":
		return catalog.Text("monster_zhentil_cleric", "散提爾堡牧師")
	case "HIGH PRIEST":
		return catalog.Text("monster_high_priest", "大祭司")
	case "ZHENTRIM CLERIC":
		return catalog.Text("monster_zhentrim_cleric", "散塔林牧師")
	case "ZHENTRIM FGHTR":
		return catalog.Text("monster_zhentrim_fighter", "散塔林戰士")
	case "ZHENTRIM MAGE":
		return catalog.Text("monster_zhentrim_mage", "散塔林法師")
	case "CULTIST":
		return catalog.Text("monster_moander_cultist", "摩安德教徒")
	case "MOGION":
		return catalog.Text("monster_mogion", "摩貢")
	case "SHAMBLING MOUND":
		return catalog.Text("monster_shambling_mound", "蔓生怪")
	case "BIT O' MOANDER":
		return catalog.Text("monster_bit_of_moander", "摩安德殘軀")
	default:
		return name
	}
}

func applyCombatTeamWrites(spawns []ecl.MonsterSpawn, writes []ecl.CombatTeamWrite, partyCount int) {
	for _, write := range writes {
		monsterIndex := write.TeamListIndex - partyCount
		for spawnIndex := range spawns {
			count := int(spawns[spawnIndex].Count)
			if count == 0 {
				count = 1
			}
			if monsterIndex >= 0 && monsterIndex < count {
				mask := uint64(1) << monsterIndex
				if write.Value == 0 || write.Value == 0x80 {
					spawns[spawnIndex].PartyMask |= mask
				} else if write.Value == 0x81 {
					spawns[spawnIndex].PartyMask &^= mask
				}
				break
			}
			monsterIndex -= count
		}
	}
}

func (s *State) CombatActive() bool { return s.battle != nil && s.Mode == ModeCombat }

// EnableCombatVisualTimeline makes actor handoff wait for renderer playback.
// It is opt-in so headless rules tests and non-visual adapters retain their
// deterministic synchronous behavior.
func (s *State) EnableCombatVisualTimeline(enabled bool) {
	s.combatVisualEnabled = enabled
	if !enabled {
		s.combatVisual = nil
	}
}

func (s *State) CombatVisualEvent() (combat.VisualEvent, bool) {
	if s.combatVisual == nil {
		return combat.VisualEvent{}, false
	}
	return *s.combatVisual, true
}

func (s *State) CombatVisualPending() bool { return s.combatVisual != nil }

func (s *State) queueCombatVisual(event combat.VisualEvent) bool {
	if !s.combatVisualEnabled {
		return false
	}
	s.combatVisualSerial++
	event.Serial = s.combatVisualSerial
	s.combatVisual = &event
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
	event := *s.combatVisual
	frame := event.FrameAt(elapsed)
	if !s.combatVisualTravelSent && frame.Phase >= combat.VisualTravel {
		switch {
		case event.Kind == combat.VisualMissile:
			s.requestSound(SoundArrow)
		case event.Kind == combat.VisualLineSpell:
			s.requestSound(SoundLightning)
		case event.Kind == combat.VisualAreaSpell && event.Effect == "fireball":
			s.requestSound(SoundFireball)
		case event.Kind == combat.VisualMagicMissile || event.Kind == combat.VisualAreaSpell:
			s.requestSound(SoundCast)
		}
		s.combatVisualTravelSent = true
	}
	if impact, ok := event.Impact(frame); ok {
		if frame.Phase >= combat.VisualImpact && frame.ImpactIndex > s.combatVisualImpactSent {
			switch event.Kind {
			case combat.VisualMagicMissile, combat.VisualLineSpell:
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
		fmt.Sprintf(s.catalog.Text("combat_view_name", "角色：%s"), fighter.Name),
		fmt.Sprintf(s.catalog.Text("combat_view_hp", "生命：%d/%d"), fighter.HitPoints, fighter.MaxHitPoints),
		fmt.Sprintf(s.catalog.Text("combat_view_ac", "護甲等級：%d"), fighter.ArmorClass),
		fmt.Sprintf(s.catalog.Text("combat_view_attack", "攻擊加值：%d"), fighter.AttackBonus),
	}
}

func (s *State) CombatMessage() string { return s.combatMessage }

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
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_moved", "%s 移動到 (%d,%d)。"), moveResult.Fighter.Name, moveResult.Fighter.CombatX, moveResult.Fighter.CombatY)
		if len(moveResult.FreeAttacks) > 0 {
			last := moveResult.FreeAttacks[len(moveResult.FreeAttacks)-1]
			s.requestAttackSounds(moveResult.FreeAttacks)
			s.combatMessage += " " + fmt.Sprintf(s.catalog.Text("combat_free_attack", "移動時遭受免費反擊，受到 %d 點傷害。"), last.Damage)
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
		if s.combatSpellIsProtectionFromGood() {
			caster, ok := s.combatPartyTurn()
			if !ok {
				return nil
			}
			return s.protectionFromGoodTargets(caster)
		}
		return s.livingBySide(combat.SideEnemy)
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
		if status == party.HealthStatusDead {
			continue
		}
		targets = append(targets, fighter)
	}
	return targets
}

// CombatSpellTargetsEnemy distinguishes the class-specific spell ID 7:
// Magic Missile is a magic-user enemy target, while Protection from Good is a
// cleric party target.
func (s *State) CombatSpellTargetsEnemy() bool {
	switch s.combatCastingSpell {
	case MagicMissileSpellID:
		return !s.combatSpellIsProtectionFromGood()
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
	if spellID == MagicMissileSpellID && !s.CombatCanCastMagicMissile() && !s.CombatCanCastProtectionFromGood() {
		return fmt.Errorf("spell 0x%02X is unavailable for this caster", spellID)
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
	if spellID != BlessSpellID && spellID != CurseSpellID && spellID != CauseLightWoundsSpellID && spellID != ProtectionFromEvilSpellID && spellID != MagicMissileSpellID && spellID != StinkingCloudSpellID && spellID != CloudkillSpellID && spellID != FireballSpellID && spellID != LightningBoltSpellID && spellID != CureLightWoundsSpellID {
		return fmt.Errorf("spell 0x%02X is not implemented in combat", spellID)
	}
	s.combatCastingSpell = spellID
	s.combatSpellTargetsPoint = false
	if spellID == ProtectionFromGoodSpellID {
		caster, ok := s.combatPartyTurn()
		if !ok {
			return fmt.Errorf("it is not a living party turn")
		}
		if _, ok := s.combatCasterClass(caster.ID); !ok {
			return fmt.Errorf("caster %q has no class", caster.ID)
		}
		cleric := s.combatCasterHasClass(caster.ID, party.ClassCleric)
		if cleric {
			s.combatCastingClass, s.combatCastingClassSet = party.ClassCleric, true
			s.combatSpellTargetIndex = 0
			return nil
		}
		class, _ := s.combatCasterClass(caster.ID)
		s.combatCastingClass, s.combatCastingClassSet = class, true
		targets := s.livingBySide(combat.SideEnemy)
		if s.combatTargetIndex >= len(targets) {
			s.combatTargetIndex = 0
		}
		s.combatSpellTargetIndex = s.combatTargetIndex
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
	if spellID == StinkingCloudSpellID || spellID == CloudkillSpellID || spellID == FireballSpellID || spellID == LightningBoltSpellID {
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
	if (s.combatCastingSpell != StinkingCloudSpellID && s.combatCastingSpell != CloudkillSpellID && s.combatCastingSpell != FireballSpellID && s.combatCastingSpell != LightningBoltSpellID) || !s.combatSpellTargetsPoint {
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_magic_missile", "%s 施放魔法飛彈攻擊 %s，%d 枚造成 %d 點傷害。"), caster.Name, target.Name, result.Missiles, result.Damage)
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
	for _, impact := range result.Impacts {
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID, To: impact.Point, Hit: true, Killed: impact.TargetHP <= 0,
			Damage: impact.Damage, Saved: impact.Saved,
		})
		totalDamage += impact.Damage
	}
	segments := make([]combat.VisualPathSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		segments = append(segments, combat.VisualPathSegment{
			From: segment.From, To: segment.To,
			HasImpact: segment.HasImpact, ImpactIndex: segment.ImpactIndex,
		})
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_lightning_bolt", "%s 施放閃電束，命中 %d 次，共造成 %d 點傷害。"),
		caster.Name, len(result.Impacts), totalDamage,
	)
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
		s.catalog.Text("combat_stinking_cloud", "%s 製造一片惡臭雲霧，籠罩 %d 名目標。"),
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
		s.catalog.Text("combat_cloudkill", "%s 製造一片致命毒雲，影響 %d 名目標。"),
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
	for _, impact := range result.Impacts {
		impacts = append(impacts, combat.VisualImpactTarget{
			TargetID: impact.TargetID,
			To:       positions[impact.TargetID],
			Hit:      true,
			Killed:   impact.TargetHP <= 0,
		})
		totalDamage += impact.Damage
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_fireball", "%s 施放火球術，波及 %d 名目標，共造成 %d 點傷害。"),
		caster.Name, len(result.Impacts), totalDamage,
	)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_protection_from_good", "%s 對 %s 施放防護善良，效果持續 %d 回合。"), caster.Name, target.Name, duration)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_protection_from_evil", "%s 對 %s 施放防護邪惡，效果持續 %d 回合。"), caster.Name, target.Name, duration)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cause_light_wounds", "%s 對 %s 施放造成輕傷，造成 %d 點傷害。"), caster.Name, target.Name, result.Damage)
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
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_curse_immune", "%s 對 %s 施放詛咒，但目標與我方相鄰，法術未生效。"), caster.Name, target.Name)
	} else {
		s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_curse", "%s 對 %s 施放詛咒，敵方攻擊加值降低 1。"), caster.Name, target.Name)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_bless", "%s 施放祝福，隊伍攻擊加值提高 1。"), caster.Name)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cure_light_wounds", "%s 對 %s 施放治療輕傷，恢復 %d HP。"), caster.Name, target.Name, result.Healing)
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
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_done", "%s 結束回合。"), attacker.Name)
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
	for s.battle != nil && s.battle.Status() == combat.StatusActive {
		if s.combatTurnIndex >= len(s.combatTurns) {
			return s.advanceCombatRound()
		}
		turn := s.combatTurns[s.combatTurnIndex]
		fighter, ok := s.fighter(turn.FighterID)
		if !ok || fighter.HitPoints <= 0 {
			s.combatTurnIndex++
			continue
		}
		if fighter.MonsterIsHeld() {
			s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_monster_held", "%s 無法行動。"), fighter.Name)
			s.combatTurnIndex++
			continue
		}
		if fighter.CloudIncapacitated() {
			helpless := fighter.HelplessTurns > 0
			if _, err := s.battle.ConsumeCloudIncapacitation(fighter.ID); err != nil {
				return err
			}
			if helpless {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cloud_helpless", "%s 因噁心而動彈不得。"), fighter.Name)
			} else {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_cloud_coughing", "%s 不斷咳嗽，無法行動。"), fighter.Name)
			}
			s.combatTurnIndex++
			continue
		}
		if fighter.Side == combat.SideParty && !fighter.QuickFight {
			return nil
		}
		targetSide := combat.SideParty
		if fighter.Side == combat.SideParty {
			targetSide = combat.SideEnemy
		}
		targets := s.livingBySide(targetSide)
		if len(targets) == 0 {
			return s.finishCombat()
		}
		target, err := s.battle.SelectCombatTarget(fighter.ID, targetSide)
		if err != nil {
			return err
		}
		if hasMonsterMagicMissile(fighter) {
			result, spellErr := s.battle.CastMonsterMagicMissile(fighter.ID, target.ID)
			if spellErr == nil {
				s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_monster_magic_missile", "%s 施放魔法飛彈攻擊 %s，造成 %d 點傷害。"), fighter.Name, target.Name, result.Damage)
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
	turns, err := s.battle.StartRound()
	if err != nil {
		return err
	}
	s.combatTurns = turns
	s.combatTurnIndex = 0
	return s.advanceCombatToParty()
}

func (s *State) finishCombat() error {
	if s.battle == nil {
		return fmt.Errorf("combat is not initialized")
	}
	s.combatVisual = nil
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
				s.Message = "財寶等待素材載入：" + err.Error()
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
			s.Message = "事件畫面"
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
		s.Message = s.catalog.Text("combat_started", "戰鬥開始（戰鬥資料尚未完成）")
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
			s.Prompt = localizePrompt(s.catalog, menu.Prompt)
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
		return fmt.Sprintf(catalog.Text("combat_miss", "%s 攻擊 %s 未命中。"), attacker.Name, target.Name)
	}
	return fmt.Sprintf(catalog.Text("combat_hit", "%s 攻擊 %s，造成 %d 點傷害。"), attacker.Name, target.Name, result.Damage)
}

func formatMultiAttackMessage(catalog interface{ Text(string, string) string }, attacker combat.Fighter, results []combat.AttackResult) string {
	hits, damage := 0, 0
	for _, result := range results {
		if result.Hit {
			hits++
			damage += result.Damage
		}
	}
	return fmt.Sprintf(catalog.Text("combat_multi_attack", "%s 連續攻擊 %d 次，命中 %d 次，造成 %d 點傷害。"), attacker.Name, len(results), hits, damage)
}

func combatResultMessage(catalog interface{ Text(string, string) string }, status combat.Status) string {
	switch status {
	case combat.StatusPartyWon:
		return catalog.Text("combat_victory", "戰鬥勝利！")
	case combat.StatusEnemyWon:
		return catalog.Text("combat_defeat", "戰鬥失敗。")
	default:
		return catalog.Text("combat_draw", "戰鬥結束。")
	}
}
