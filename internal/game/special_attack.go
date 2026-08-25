package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// 怪物特殊攻擊的 AI 分派（spec 1202）。
//
// 位置比照 `84h` 丟電光：AI 回合的特殊行動階段，在一般物理攻擊之前。
// `Refrained`（亂數、距離、次數或同類檢查的靜默放棄）不消耗回合——
// 原作 handler 靜默返回後 AI 照走正常行動；`Missed`（吐酸擲失）有訊息、
// 回合照樣消耗。

// monsterSpecialAttack 嘗試發動一個特殊攻擊。回傳：
//   - handled：這個回合已由特殊攻擊處理完（訊息已設、回合已推進）。
//   - queued：排了視覺事件——呼叫端要直接返回，讓視覺時間軸接手；
//     沒排的話呼叫端 continue 回合迴圈（**不要遞迴**回 advanceCombatToParty：
//     凝視不造成傷害，全隊麻痺的僵局會讓遞迴無限深）。
func (s *State) monsterSpecialAttack(fighter combat.Fighter, targetSide combat.Side) (bool, bool, error) {
	for _, rule := range fighter.MonsterSpecialAttackRules() {
		target, found := s.battle.NearestSpecialAttackTarget(fighter.ID, targetSide, rule.TargetRange)
		if !found {
			continue
		}
		if rule.Form == combat.SpecialAttackGaze && target.MonsterIsHeld() {
			// 已經麻痺的目標不再凝視（原作沒有這道檢查；remake 加上以免
			// 「無傷害的攻擊對永遠動不了的目標」變成打不完的僵局），
			// 落回一般攻擊。
			continue
		}
		var result combat.SpecialAttackResult
		var err error
		switch rule.Form {
		case combat.SpecialAttackSpit, combat.SpecialAttackBreathTouch:
			result, err = s.battle.SpecialAttackSingle(fighter.ID, target.ID, rule)
		case combat.SpecialAttackGaze:
			result, err = s.battle.SpecialAttackGazeAt(fighter.ID, target.ID, rule)
		case combat.SpecialAttackBreathArea, combat.SpecialAttackBreathAreaSameSide:
			result, err = s.battle.SpecialAttackAreaBreath(fighter.ID,
				combat.TilePoint{X: target.CombatX, Y: target.CombatY}, rule)
		default:
			return false, false, fmt.Errorf("special attack %q has unhandled form %q", rule.ID, rule.Form)
		}
		if err != nil {
			return false, false, err
		}
		if result.Refrained {
			continue
		}
		return true, s.finishSpecialAttack(fighter, target, rule, result), nil
	}
	return false, false, nil
}

// finishSpecialAttack 顯示訊息、排視覺並推進回合。回傳「排了視覺沒」。
func (s *State) finishSpecialAttack(fighter, target combat.Fighter,
	rule combat.SpecialAttackRule, result combat.SpecialAttackResult) bool {
	visualQueued := false
	switch {
	case result.Missed:
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text(rule.MissMessage, rule.MissMessage), fighter.Name)
	case rule.Form == combat.SpecialAttackGaze:
		impact := result.Impacts[0]
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text(rule.Message, rule.Message), fighter.Name, target.Name)
		if impact.Paralyzed {
			s.combatMessage += "\n" + fmt.Sprintf(
				s.catalog.Text(rule.ParalyzedMessage, rule.ParalyzedMessage), target.Name)
		} else {
			s.combatMessage += "\n" + fmt.Sprintf(
				s.catalog.Text("combat_gaze_resisted", "combat_gaze_resisted"), target.Name)
		}
		visualQueued = s.queueMagicMissileVisual(fighter, target, 1, false)
	case rule.Form == combat.SpecialAttackSpit || rule.Form == combat.SpecialAttackBreathTouch:
		impact := result.Impacts[0]
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text(rule.Message, rule.Message), fighter.Name, target.Name, impact.Damage)
		visualQueued = s.queueMagicMissileVisual(fighter, target, 1, impact.TargetHP <= 0)
	default:
		total := 0
		impacts := make([]combat.VisualImpactTarget, 0, len(result.Impacts))
		for _, impact := range result.Impacts {
			total += impact.Damage
			if hit, ok := s.battle.Fighter(impact.TargetID); ok {
				impacts = append(impacts, combat.VisualImpactTarget{
					TargetID: impact.TargetID,
					To:       combat.TilePoint{X: hit.CombatX, Y: hit.CombatY},
					Hit:      true, Killed: impact.TargetHP <= 0,
					Damage: impact.Damage, Saved: impact.Saved,
				})
			}
		}
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text(rule.Message, rule.Message), fighter.Name, len(result.Impacts), total)
		// ⚠ 視覺近似：原作的區域吐息播的是共用投射物（槽 18＝COMSPR 區塊 5）
		// 加逐目標命中——與火球的資產是同一族，這裡借火球的區域事件。
		visualQueued = s.queueCombatVisual(combat.VisualEvent{
			Kind: combat.VisualAreaSpell, Effect: "fireball", ActorID: fighter.ID,
			From: combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY},
			To:   combat.TilePoint{X: target.CombatX, Y: target.CombatY},
			Hit:  len(impacts) != 0, Impacts: impacts,
		})
	}
	if visualQueued {
		return true
	}
	s.combatTurnIndex++
	for _, impact := range result.Impacts {
		if impact.Damage > 0 {
			s.requestSound(SoundSpellHit)
		}
		if impact.TargetHP <= 0 && impact.Damage > 0 {
			s.requestSound(SoundDead)
		}
	}
	return false
}
