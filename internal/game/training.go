package game

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
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
	s.Prompt = s.catalog.Text("training_select_character", "training_select_character")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("training_character_summary", "training_character_summary"),
			character.Name, s.localizedCharacterClassName(character.Class), character.Level, character.Experience))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TRAIN_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("training_exit", "training_exit"))
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
		return s.finishTrainingResult(s.trainingResult + fmt.Sprintf(
			s.catalog.Text("training_learned_spell_suffix", "training_learned_spell_suffix"),
			s.trainingSpellName(uint8(spellID))))
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
		return s.trainingFailure(s.catalog.Text("training_requires_healthy", "training_requires_healthy"))
	}
	if characterCoinGoldWorth(character) < trainingCost {
		return s.trainingFailure(s.catalog.Text("training_insufficient_gold", "training_insufficient_gold"))
	}
	if !eligible {
		return s.trainingFailure(s.catalog.Text("training_insufficient_experience", "training_insufficient_experience"))
	}
	s.trainingCharacterIndex = index
	s.trainingConfirmMenu = true
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("training_confirm_prompt", "training_confirm_prompt"),
		character.Name, level+1, s.localizedCharacterClassName(info.Class))
	s.Choices = []string{
		s.catalog.Text("training_confirm", "training_confirm"),
		s.catalog.Text("training_cancel", "training_cancel"),
	}
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

func (s *State) localizedCharacterClassName(class party.Class) string {
	key := "class_unknown"
	switch class {
	case party.ClassCleric:
		key = "class_cleric"
	case party.ClassFighter:
		key = "class_fighter"
	case party.ClassRanger:
		key = "class_ranger"
	case party.ClassPaladin:
		key = "class_paladin"
	case party.ClassMagicUser:
		key = "class_magic_user"
	case party.ClassThief:
		key = "class_thief"
	}
	return s.catalog.Text(key, key)
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
	result := fmt.Sprintf(s.catalog.Text("training_success", "training_success"),
		character.Name, oldLevel+1, s.localizedCharacterClassName(info.Class))
	if increase > 0 {
		result += fmt.Sprintf(s.catalog.Text("training_hp_increase", "training_hp_increase"), increase)
	} else {
		result += s.catalog.Text("training_dual_class_no_hp", "training_dual_class_no_hp")
	}
	if info.Class == party.ClassMagicUser || info.Class == party.ClassRanger && oldLevel+1 > 8 {
		if candidates := trainingSpellCandidates(*character); len(candidates) > 0 {
			s.trainingSpellMenu = true
			s.trainingSpellChoices = candidates
			s.trainingResult = result
			s.Mode = ModeWilderness
			s.Prompt = s.catalog.Text("training_select_spell", "training_select_spell")
			s.Choices = make([]string, 0, len(candidates))
			s.currentOriginalChoices = make([]string, 0, len(candidates))
			for _, spellID := range candidates {
				s.Choices = append(s.Choices, s.trainingSpellName(spellID))
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
	Key   string
}

// trainingSpells 是訓練所可以選的法術，直接由**原作法術表**過濾出來：
// 施法職業是德魯伊（1）或法師（2）的每一支。牧師系（0）不在清單裡——
// AD&D 的牧師不挑法術；職業碼 3 那 15 筆玩家取不到。
//
// ★ 這裡曾經是一份手打的 49 列表格（編號、職業、環數、locale 鍵各寫一次）。
// 四欄原作全部都有值，手打的那份只會漂移：它的 locale 鍵用的是「法師第幾支」
// 那套編號，而同一個前綴的其他鍵用全域編號，兩套在 1..13 這段會互相蓋掉。
func trainingSpellCatalog() []trainingSpell {
	table, err := gamepack.Spells()
	if err != nil || table == nil {
		return nil
	}
	spells := make([]trainingSpell, 0, len(table.Spells))
	for _, spell := range table.Spells {
		if spell.Placeholder || spell.Level < 1 || spell.Level > 5 {
			continue
		}
		if spell.CasterClassID != 1 && spell.CasterClassID != 2 {
			continue
		}
		key, ok := spellMessageID(uint8(spell.SpellID))
		if !ok {
			continue
		}
		spells = append(spells, trainingSpell{
			ID:    uint8(spell.SpellID),
			Class: spell.CasterClassID,
			Level: spell.Level,
			Key:   key,
		})
	}
	return spells
}

var trainingSpells = trainingSpellCatalog()

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

func (s *State) trainingSpellName(spellID uint8) string {
	for _, spell := range trainingSpells {
		if spell.ID == spellID {
			return s.catalog.Text(spell.Key, spell.Key)
		}
	}
	return fmt.Sprintf(s.catalog.Text("spell_unknown", "spell_unknown 0x%02X"), spellID)
}

func recalculateTrainingSpellCounts(character *party.Character) {
	character.SpellCastCount[0] = [5]uint8{}
	character.SpellCastCount[1] = [5]uint8{}
	character.SpellCastCount[2] = [5]uint8{}
	clericLevel := character.ClassLevel(party.ClassCleric)
	if clericLevel > 0 {
		// DOS DS:42BAh／PC-98 DS:734Fh 的累計結果（spec 809/810 exact）。
		cleric := [...][5]uint8{
			{}, {1, 0, 0, 0, 0}, {2, 0, 0, 0, 0}, {2, 1, 0, 0, 0},
			{3, 2, 0, 0, 0}, {3, 3, 1, 0, 0}, {3, 3, 2, 0, 0},
			{3, 3, 2, 1, 0}, {3, 3, 3, 2, 0}, {4, 4, 3, 2, 1},
			{4, 4, 3, 3, 2}, {5, 4, 4, 3, 2}, {6, 5, 5, 3, 2},
		}
		level := clericLevel
		if level >= len(cleric) {
			level = len(cleric) - 1
		}
		character.SpellCastCount[0] = cleric[level]
		wisdom := character.Abilities.Wisdom
		if wisdom >= 13 && character.SpellCastCount[0][0] > 0 {
			character.SpellCastCount[0][0]++
		}
		if wisdom >= 14 && character.SpellCastCount[0][0] > 0 {
			character.SpellCastCount[0][0]++
		}
		if wisdom >= 15 && character.SpellCastCount[0][1] > 0 {
			character.SpellCastCount[0][1]++
		}
		if wisdom >= 16 && character.SpellCastCount[0][1] > 0 {
			character.SpellCastCount[0][1]++
		}
		if wisdom >= 17 && character.SpellCastCount[0][2] > 0 {
			character.SpellCastCount[0][2]++
		}
		if wisdom >= 18 && character.SpellCastCount[0][3] > 0 {
			character.SpellCastCount[0][3]++
		}
	}
	paladinLevel := character.ClassLevel(party.ClassPaladin)
	if paladinLevel >= 9 {
		paladin := [...][5]uint8{{}, {}, {}, {}, {}, {}, {}, {}, {}, {1}, {2}, {2, 1}, {2, 2}}
		level := paladinLevel
		if level >= len(paladin) {
			level = len(paladin) - 1
		}
		for spellLevel := range character.SpellCastCount[0] {
			character.SpellCastCount[0][spellLevel] += paladin[level][spellLevel]
		}
	}
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
