package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
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
		fighters = append(fighters, fighter)
		enemyIndex++
	}
	battle, err := combat.NewBattle(fighters, seed)
	if err != nil {
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
	s.combatMessage = s.catalog.Text("combat_started", "戰鬥開始！")
	s.Mode = ModeCombat
	return s.advanceCombatToParty()
}

// StartEncounter is the data bridge from the bounded ECL runner to the
// playable battle state. ECL supplies only spawn IDs/counts; MON*CHA supplies
// the combat statistics, and the caller supplies the decoded party roster.
func (s *State) StartEncounter(result ecl.RunResult, records map[uint8]monster.Record, party []combat.Fighter, seed int64) error {
	if !result.CombatRequested {
		return fmt.Errorf("ECL result does not request combat")
	}
	enemies, err := monster.BuildEnemies(result.MonsterSpawns, records)
	if err != nil {
		return err
	}
	for index := range enemies {
		enemies[index].SpriteSet = s.Area.GameArea
		if result.MonsterSetup != nil {
			enemies[index].AnimationBlock = result.MonsterSetup.SpriteID
			enemies[index].HasAnimation = true
		}
	}
	return s.StartCombat(party, enemies, seed)
}

func (s *State) CombatActive() bool { return s.battle != nil && s.Mode == ModeCombat }

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

func (s *State) CombatTurns() []combat.Turn {
	return append([]combat.Turn(nil), s.combatTurns...)
}

func (s *State) CombatMessage() string { return s.combatMessage }

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
		if character.ID != caster.ID || character.Class != party.ClassMagicUser {
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

func (s *State) CombatCanCastCureLightWounds() bool {
	if !s.CombatActive() {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	wounded := false
	for _, fighter := range s.livingBySide(combat.SideParty) {
		if fighter.HitPoints < fighter.MaxHitPoints {
			wounded = true
			break
		}
	}
	if !wounded {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || character.Class != party.ClassCleric {
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
		if character.ID != caster.ID || character.Class != party.ClassCleric {
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
		if character.ID != caster.ID || character.Class != party.ClassCleric {
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
		if character.ID != caster.ID || character.Class != party.ClassCleric {
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
		if character.ID != caster.ID || character.Class != party.ClassCleric {
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

func (s *State) CombatCastingSpell() uint8 { return s.combatCastingSpell }

func (s *State) CombatSpellTargetIndex() int { return s.combatSpellTargetIndex }

func (s *State) CombatSpellTargets() []combat.Fighter {
	switch s.combatCastingSpell {
	case CureLightWoundsSpellID:
		return s.livingBySide(combat.SideParty)
	case MagicMissileSpellID:
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
	if spellID == CureLightWoundsSpellID && !s.CombatCanCastCureLightWounds() {
		return fmt.Errorf("Cure Light Wounds is unavailable")
	}
	if spellID != BlessSpellID && spellID != CurseSpellID && spellID != CauseLightWoundsSpellID && spellID != ProtectionFromEvilSpellID && spellID != MagicMissileSpellID && spellID != CureLightWoundsSpellID {
		return fmt.Errorf("spell 0x%02X is not implemented in combat", spellID)
	}
	s.combatCastingSpell = spellID
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
	targets := s.livingBySide(combat.SideParty)
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
	s.combatSpellTargetIndex = 0
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
	if s.combatCastingSpell == MagicMissileSpellID || s.combatCastingSpell == CurseSpellID || s.combatCastingSpell == CauseLightWoundsSpellID {
		s.combatTargetIndex = s.combatSpellTargetIndex
	}
	return nil
}

// CombatCast applies the verified first-level Magic Missile path. It consumes
// exactly one memorized slot, targets the current enemy selection, and then
// advances the same deterministic enemy-turn boundary as a weapon attack.
func (s *State) CombatCast(spellID uint8) error {
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
	if spellID == CureLightWoundsSpellID {
		return s.combatCastCureLightWounds()
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
		if character.ID == caster.ID && character.Class == party.ClassMagicUser {
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
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
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
		if character.ID == caster.ID && character.Class == party.ClassCleric {
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
		if character.ID == caster.ID && character.Class == party.ClassCleric {
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
		if character.ID == caster.ID && character.Class == party.ClassCleric {
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
		if character.ID == caster.ID && character.Class == party.ClassCleric {
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
		if character.ID == caster.ID && character.Class == party.ClassCleric {
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
	targets := s.livingBySide(combat.SideParty)
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
	s.CancelCombatCast()
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
	if s.combatTargetIndex >= len(enemies) {
		s.combatTargetIndex = 0
	}
	target := enemies[s.combatTargetIndex]
	result, err := s.battle.Attack(attacker.ID, target.ID)
	if err != nil {
		return err
	}
	s.combatMessage = formatAttackMessage(s.catalog, attacker, target, result)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
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
		if fighter.Side == combat.SideParty {
			return nil
		}
		party := s.livingBySide(combat.SideParty)
		if len(party) == 0 {
			return s.finishCombat()
		}
		result, err := s.battle.Attack(fighter.ID, party[0].ID)
		if err != nil {
			return err
		}
		s.combatMessage = formatAttackMessage(s.catalog, fighter, party[0], result)
		s.combatTurnIndex++
	}
	return s.finishCombat()
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
	s.CancelCombatCast()
	s.Mode = ModeEvent
	s.syncPartyFromBattle()
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = "COMBAT"
	s.combatMessage = combatResultMessage(s.catalog, s.battle.Status())
	s.Message = s.combatMessage
	return nil
}

func (s *State) syncPartyFromBattle() {
	if s.battle == nil || len(s.party) == 0 {
		return
	}
	for index := range s.party {
		if fighter, ok := s.fighter(s.party[index].ID); ok {
			s.party[index] = fighter
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
