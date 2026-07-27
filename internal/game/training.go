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
const trainingSpellCommandPrefix = "TRAIN_SPELL_"

func (s *State) enterTrainingMenu() {
	s.trainingMenu = true
	s.trainingConfirmMenu = false
	s.trainingSpellMenu = false
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
	if len(originalChoice) > len(trainingSpellCommandPrefix) &&
		originalChoice[:len(trainingSpellCommandPrefix)] == trainingSpellCommandPrefix {
		if !s.trainingSpellMenu {
			return fmt.Errorf("training spell selected outside spell menu")
		}
		spellID, err := strconv.Atoi(originalChoice[len(trainingSpellCommandPrefix):])
		if err != nil || spellID < 1 || spellID > 100 {
			return fmt.Errorf("invalid training spell command %q", originalChoice)
		}
		allowed := false
		for _, candidate := range s.trainingSpellChoices {
			if candidate == uint8(spellID) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("spell %d is not available for training", spellID)
		}
		character := &s.partyRoster[s.trainingCharacterIndex]
		character.KnownSpells = append(character.KnownSpells, uint8(spellID))
		if err := s.syncTempleCharacter(s.trainingCharacterIndex); err != nil {
			return err
		}
		s.trainingSpellMenu = false
		return s.finishTrainingResult(s.trainingResult + fmt.Sprintf(" 並學會%s。", trainingSpellName(uint8(spellID))))
	}
	switch originalChoice {
	case "TRAIN_EXIT":
		s.trainingMenu = false
		s.trainingConfirmMenu = false
		s.trainingSpellMenu = false
		s.Mode = ModeDungeon
		s.Message = ""
		s.Prompt = ""
		s.Choices = nil
		s.currentOriginalChoices = nil
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
	recalculateTrainingSpellCounts(character)
	increase := 0
	if character.HitDice == 0 || character.HitDice > character.MulticlassLevel {
		increase = (hitRoll + trainingConstitutionBonus(*character)) / classCount
		if increase < 1 {
			increase = 1
		}
		character.MaxHitPoints += increase
		character.HitPoints += increase
	}
	if err := s.syncTempleCharacter(index); err != nil {
		return err
	}
	result := fmt.Sprintf("恭喜！%s成為 %d 級%s", character.Name, oldLevel+1, characterClassName(info.Class))
	if increase > 0 {
		result += fmt.Sprintf("，最大 HP 增加 %d。", increase)
	} else {
		result += "；超過原職業等級前不增加 HP。"
	}
	if info.Class == party.ClassMagicUser || info.Class == party.ClassRanger && oldLevel+1 > 8 {
		if candidates := trainingSpellCandidates(*character); len(candidates) > 0 {
			s.trainingSpellMenu = true
			s.trainingSpellChoices = candidates
			s.trainingResult = result
			s.Mode = ModeWilderness
			s.Prompt = "選擇一個新法術"
			s.Choices = make([]string, 0, len(candidates))
			s.currentOriginalChoices = make([]string, 0, len(candidates))
			for _, spellID := range candidates {
				s.Choices = append(s.Choices, trainingSpellName(spellID))
				s.currentOriginalChoices = append(s.currentOriginalChoices, trainingSpellCommandPrefix+strconv.Itoa(int(spellID)))
			}
			return nil
		}
	}
	return s.finishTrainingResult(result)
}

func (s *State) finishTrainingResult(message string) error {
	s.trainingConfirmMenu = false
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.Message = message
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

type trainingSpell struct {
	ID    uint8
	Class int
	Level int
	Name  string
}

var trainingSpells = []trainingSpell{
	{9, 2, 1, "燃燒之手"}, {10, 2, 1, "魅惑人類"}, {11, 2, 1, "偵測魔法"},
	{12, 2, 1, "變巨術"}, {13, 2, 1, "縮小術"}, {14, 2, 1, "交友術"},
	{15, 2, 1, "魔法飛彈"}, {16, 2, 1, "防護邪惡"}, {17, 2, 1, "防護善良"},
	{18, 2, 1, "閱讀魔法"}, {19, 2, 1, "護盾術"}, {20, 2, 1, "電爪"},
	{21, 2, 1, "睡眠術"},
	{29, 2, 2, "偵測隱形"}, {30, 2, 2, "隱形術"}, {31, 2, 2, "敲擊術"},
	{32, 2, 2, "鏡影術"}, {33, 2, 2, "衰弱射線"}, {34, 2, 2, "惡臭之雲"},
	{35, 2, 2, "力量術"},
	{45, 2, 3, "閃現術"}, {46, 2, 3, "解除魔法"}, {47, 2, 3, "火球術"},
	{48, 2, 3, "加速術"}, {49, 2, 3, "人類定身術"}, {50, 2, 3, "十呎隱形術"},
	{51, 2, 3, "閃電束"}, {52, 2, 3, "十呎防護邪惡"}, {53, 2, 3, "十呎防護善良"},
	{54, 2, 3, "防護普通飛彈"}, {55, 2, 3, "緩慢術"},
	{81, 2, 4, "魅惑怪物"}, {82, 2, 4, "困惑術"}, {83, 2, 4, "次元門"},
	{84, 2, 4, "恐懼術"}, {85, 2, 4, "火焰護盾"}, {86, 2, 4, "笨拙術"},
	{87, 2, 4, "冰風暴"}, {88, 2, 4, "次級法術無效結界"}, {89, 2, 4, "移除詛咒"},
	{100, 2, 4, "降咒術"},
	{91, 2, 5, "死雲術"}, {92, 2, 5, "冰錐術"}, {93, 2, 5, "弱智術"},
	{94, 2, 5, "怪物定身術"},
	{77, 1, 1, "德魯伊偵測魔法"}, {78, 1, 1, "糾纏術"},
	{79, 1, 1, "妖火術"}, {80, 1, 1, "動物隱形術"},
}

func trainingSpellCandidates(character party.Character) []uint8 {
	known := make(map[uint8]bool, len(character.KnownSpells))
	for _, spellID := range character.KnownSpells {
		known[spellID] = true
	}
	result := make([]uint8, 0)
	for _, spell := range trainingSpells {
		if !known[spell.ID] && spell.Level >= 1 && spell.Level <= 5 &&
			character.SpellCastCount[spell.Class][spell.Level-1] > 0 {
			result = append(result, spell.ID)
		}
	}
	return result
}

func trainingSpellName(spellID uint8) string {
	for _, spell := range trainingSpells {
		if spell.ID == spellID {
			return spell.Name
		}
	}
	return fmt.Sprintf("未知法術 0x%02X", spellID)
}

func recalculateTrainingSpellCounts(character *party.Character) {
	character.SpellCastCount[1] = [5]uint8{}
	character.SpellCastCount[2] = [5]uint8{}
	magicUserLevel := character.ClassLevel(party.ClassMagicUser)
	if magicUserLevel > 0 {
		character.SpellCastCount[2][0] = 1
		table := [...][5]uint8{
			{1, 0, 0, 0, 0}, {0, 1, 0, 0, 0}, {1, 1, 0, 0, 0},
			{1, 0, 1, 0, 0}, {0, 0, 1, 0, 0}, {0, 1, 0, 1, 0},
			{0, 0, 1, 1, 0}, {0, 0, 0, 0, 1}, {0, 1, 0, 0, 1},
			{0, 0, 1, 1, 1}, {0, 0, 0, 1, 1},
		}
		for level := 0; level <= magicUserLevel-2 && level < len(table); level++ {
			for spellLevel := 0; spellLevel < 5; spellLevel++ {
				character.SpellCastCount[2][spellLevel] += table[level][spellLevel]
			}
		}
	}
	rangerLevel := character.ClassLevel(party.ClassRanger)
	if rangerLevel > 7 {
		table := [...][5]uint8{
			{}, {}, {}, {}, {}, {}, {}, {},
			{1, 0, 0, 0, 0}, {0, 0, 0, 1, 0}, {1, 0, 0, 0, 0},
			{0, 0, 0, 1, 0}, {0, 1, 0, 0, 0},
		}
		for level := 8; level <= rangerLevel && level < len(table); level++ {
			for spellLevel := 0; spellLevel < 3; spellLevel++ {
				character.SpellCastCount[1][spellLevel] += table[level][spellLevel]
			}
			for spellLevel := 3; spellLevel < 5; spellLevel++ {
				character.SpellCastCount[2][spellLevel-3] += table[level][spellLevel]
			}
		}
	}
}
