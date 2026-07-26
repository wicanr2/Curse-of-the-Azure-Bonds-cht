package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// StartCombat creates the first playable battle adapter. Party and encounter
// records are supplied by the data layer, so this state machine does not
// invent MON*CHA statistics or DOS party layout.
func (s *State) StartCombat(party, enemies []combat.Fighter, seed int64) error {
	if len(party) == 0 || len(enemies) == 0 {
		return fmt.Errorf("combat needs at least one party member and enemy")
	}
	fighters := make([]combat.Fighter, 0, len(party)+len(enemies))
	for _, fighter := range party {
		if fighter.Side != combat.SideParty {
			return fmt.Errorf("fighter %q is not marked as party", fighter.ID)
		}
		fighters = append(fighters, fighter)
	}
	for _, fighter := range enemies {
		if fighter.Side != combat.SideEnemy {
			return fmt.Errorf("fighter %q is not marked as enemy", fighter.ID)
		}
		fighters = append(fighters, fighter)
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
