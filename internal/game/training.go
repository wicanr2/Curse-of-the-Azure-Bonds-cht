package game

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

type trainingClass struct {
	Class     party.Class
	Slot      int
	Mask      uint8
	HitDie    int
	MaxHitDie int
	FixedHP   int
	Threshold []uint32
}

var trainingClasses = []trainingClass{
	{party.ClassCleric, 0, 0x02, 8, 10, 2, []uint32{0, 1501, 3001, 6001, 13001, 27501, 55001, 110001, 225001, 450001}},
	{party.ClassFighter, 2, 0x08, 10, 10, 3, []uint32{0, 2001, 4001, 8001, 18001, 35001, 70001, 125001, 250001, 500001, 750001, 1000001}},
	{party.ClassPaladin, 3, 0x10, 10, 10, 3, []uint32{0, 2751, 5501, 12001, 24001, 45001, 95001, 175001, 350001, 700001, 1050001}},
	{party.ClassRanger, 4, 0x20, 8, 11, 2, []uint32{0, 2251, 4501, 10001, 20001, 40001, 90001, 150001, 225001, 325001, 650001}},
	{party.ClassMagicUser, 5, 0x01, 4, 12, 1, []uint32{0, 2501, 5001, 10001, 22501, 40001, 60001, 90001, 135001, 250001, 375001}},
	{party.ClassThief, 6, 0x04, 6, 11, 2, []uint32{0, 1251, 2501, 5001, 10001, 20001, 42501, 70001, 110001, 160001, 220001, 440001}},
}

const trainingCost = uint32(1000)

func (s *State) enterTrainingMenu() {
	s.trainingMenu = true
	s.trainingConfirmMenu = false
	s.Mode = ModeWilderness
	s.eclMenuReturnMode = ModeDungeon
	s.eventReturnMode = ModeDungeon
	s.Prompt = "訓練哪一位角色？"
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%s %d 級／%d XP）",
			character.Name, characterClassName(character.Class), character.Level, character.Experience))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TRAIN_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, "離開訓練場")
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TRAIN_EXIT")
	s.Message = ""
}

func (s *State) selectTraining(originalChoice string) error {
	switch originalChoice {
	case "TRAIN_EXIT":
		s.trainingMenu = false
		s.trainingConfirmMenu = false
		s.Mode = ModeDungeon
		s.Message = ""
		return nil
	case "TRAIN_CONFIRM":
		return s.applyTraining(s.trainingCharacterIndex)
	case "TRAIN_CANCEL":
		s.enterTrainingMenu()
		return nil
	}
	const prefix = "TRAIN_CHARACTER_"
	if len(originalChoice) <= len(prefix) || originalChoice[:len(prefix)] != prefix {
		return fmt.Errorf("unknown training command %q", originalChoice)
	}
	index, err := strconv.Atoi(originalChoice[len(prefix):])
	if err != nil || index < 0 || index >= len(s.partyRoster) {
		return fmt.Errorf("invalid training character command %q", originalChoice)
	}
	character := s.partyRoster[index]
	info, level, eligible := trainableClass(character, 0xFF)
	if character.HealthStatus != party.HealthStatusOK {
		return s.trainingFailure("訓練場只接受神智清醒的人。")
	}
	if characterCoinGoldWorth(character) < trainingCost {
		return s.trainingFailure("訓練費用是 1000 GP。")
	}
	if !eligible {
		return s.trainingFailure("經驗值不足，現在還不能升級。")
	}
	s.trainingCharacterIndex = index
	s.trainingConfirmMenu = true
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf("%s將成為 %d 級%s。要支付 1000 GP 訓練嗎？",
		character.Name, level+1, characterClassName(info.Class))
	s.Choices = []string{"確定", "取消"}
	s.currentOriginalChoices = []string{"TRAIN_CONFIRM", "TRAIN_CANCEL"}
	s.Message = ""
	return nil
}

func (s *State) trainingFailure(message string) error {
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.Message = message
	return nil
}

func trainableClass(character party.Character, trainerMask uint8) (trainingClass, int, bool) {
	selected := trainingClasses[0]
	selectedLevel := 0
	selectedThreshold := uint32(0)
	for _, info := range trainingClasses {
		level := character.ClassLevel(info.Class)
		if level <= 0 || level >= len(info.Threshold) || info.Mask&trainerMask == 0 {
			continue
		}
		if trainingRaceClassLimited(character, info.Class, level) {
			continue
		}
		threshold := info.Threshold[level]
		if threshold > 0 && threshold <= character.Experience && threshold >= selectedThreshold {
			selected, selectedLevel, selectedThreshold = info, level, threshold
		}
	}
	return selected, selectedLevel, selectedThreshold > 0
}

func (s *State) applyTraining(index int) error {
	if index < 0 || index >= len(s.partyRoster) {
		return fmt.Errorf("training character %d is outside roster", index)
	}
	character := &s.partyRoster[index]
	info, oldLevel, eligible := trainableClass(*character, 0xFF)
	if !eligible || character.HealthStatus != party.HealthStatusOK ||
		characterCoinGoldWorth(*character) < trainingCost {
		return fmt.Errorf("training eligibility changed before confirmation")
	}
	subtractCharacterGoldWorth(character, trainingCost)
	classCount := 0
	for _, candidate := range trainingClasses {
		if character.ClassLevel(candidate.Class) > 0 {
			classCount++
		}
	}
	if classCount == 0 {
		classCount = 1
	}
	hitRoll := info.FixedHP
	if oldLevel < info.MaxHitDie {
		rng := rand.New(rand.NewSource(s.fixSeed))
		s.fixSeed++
		diceCount := 1
		if info.Class == party.ClassRanger && oldLevel == 1 {
			diceCount = 2
		}
		roll := func() int {
			total := 0
			for count := 0; count < diceCount; count++ {
				total += rng.Intn(info.HitDie) + 1
			}
			return total
		}
		hitRoll = roll()
		if second := roll(); second > hitRoll {
			hitRoll = second
		}
	}
	increase := (hitRoll + trainingConstitutionBonus(*character)) / classCount
	if increase < 1 {
		increase = 1
	}
	if character.ClassLevels == [8]uint8{} {
		character.Level++
		character.HitDice = uint8(character.Level)
	} else {
		character.ClassLevels[info.Slot]++
		character.Level = 0
		for _, level := range character.ClassLevels {
			if int(level) > character.Level {
				character.Level = int(level)
			}
		}
		character.HitDice = uint8(character.Level)
	}
	if character.HitDice > 0 && character.HitDice <= character.MulticlassLevel {
		if err := s.syncTempleCharacter(index); err != nil {
			return err
		}
		s.trainingConfirmMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModeWilderness
		s.Message = fmt.Sprintf("恭喜！%s成為 %d 級%s；超過原職業等級前不增加 HP。",
			character.Name, oldLevel+1, characterClassName(info.Class))
		return nil
	}
	character.MaxHitPoints += increase
	character.HitPoints += increase
	if err := s.syncTempleCharacter(index); err != nil {
		return err
	}
	s.trainingConfirmMenu = false
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.Message = fmt.Sprintf("恭喜！%s成為 %d 級%s，最大 HP 增加 %d。",
		character.Name, oldLevel+1, characterClassName(info.Class), increase)
	return nil
}

func trainingConstitutionBonus(character party.Character) int {
	con := character.Abilities.Constitution
	base := 0
	switch {
	case con == 3:
		base = -2
	case con >= 4 && con <= 6:
		base = -1
	case con == 15:
		base = 1
	case con >= 16:
		base = 2
	}
	total := 0
	for _, info := range trainingClasses {
		level := character.ClassLevel(info.Class)
		if level <= 0 || level >= info.MaxHitDie {
			continue
		}
		classBonus := base
		if character.Class == party.ClassFighter || character.Class == party.ClassPaladin || character.Class == party.ClassRanger {
			switch {
			case con == 17:
				classBonus++
			case con == 18:
				classBonus += 2
			case con >= 19 && con <= 20:
				classBonus += 3
			case con >= 21 && con <= 23:
				classBonus += 4
			case con >= 24:
				classBonus += 5
			}
		}
		total += classBonus
		if info.Class == party.ClassRanger && level == 1 {
			total *= 2
		}
	}
	return total
}

func trainingRaceClassLimited(character party.Character, class party.Class, level int) bool {
	strength := character.Abilities.StrengthFull
	if strength == 0 {
		strength = character.Abilities.Strength
	}
	switch character.Race {
	case party.RaceDwarf:
		return class == party.ClassFighter &&
			(level == 9 || level == 8 && strength == 17 || level == 7 && strength < 17)
	case party.RaceElf:
		if class == party.ClassFighter {
			return level == 7 || level == 6 && strength == 17 || level == 5 && strength < 17
		}
		if class == party.ClassMagicUser {
			intelligence := character.Abilities.Intelligence
			return level == 11 || level == 10 && intelligence == 17 || level == 9 && intelligence < 17
		}
	case party.RaceGnome:
		return class == party.ClassFighter && (level == 6 || level == 5 && strength < 18)
	case party.RaceHalfElf:
		if class == party.ClassCleric {
			return level == 5
		}
		if class == party.ClassFighter || class == party.ClassRanger || class == party.ClassMagicUser {
			return level == 8 || level == 7 && strength == 17 || level == 6 && strength < 17
		}
	case party.RaceHalfling:
		return class == party.ClassFighter &&
			(level == 6 || level == 5 && strength == 17 || level == 4 && strength < 17)
	}
	return false
}
