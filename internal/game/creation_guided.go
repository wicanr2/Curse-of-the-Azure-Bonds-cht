package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 原版建角是四段選單接一次擲屬性（spec 1093 §一）：
//
//	選單('Pick Race')      → 角色^[74h]  := 種族
//	選單('Pick Gender')    → 角色^[119h] := 性別
//	選單('Pick Class')     → 角色^[75h]  := byte[DS:3FF8h + 種族×0Eh + 選單索引]
//	選單('Pick Alignment') → 角色^[11Bh] := 陣營
//	…擲屬性…repeat 問('Reroll stats? ') until 不再重擲
//
// 職業那一段的選項不是固定清單：要先用種族去查可選職業表，把「第 n 個選項」
// 翻成真正的職業組合編號（spec 1099 §五，資料在 game pack）。
type CreationStep uint8

const (
	CreationStepRace CreationStep = iota
	CreationStepGender
	CreationStepClass
	CreationStepAlignment
	CreationStepAbilities
	CreationStepDone
)

// guidedRaces 依原作的 +74h 種族編號 1..7 排列（spec 1084／884）。
var guidedRaces = []struct {
	DOSRaceID int
	Race      party.Race
	LocaleKey string
}{
	{1, party.RaceDwarf, "race_dwarf"},
	{2, party.RaceElf, "race_elf"},
	{3, party.RaceGnome, "race_gnome"},
	{4, party.RaceHalfElf, "race_half_elf"},
	{5, party.RaceHalfling, "race_halfling"},
	{6, party.RaceHalfOrc, "race_half_orc"},
	{7, party.RaceHuman, "race_human"},
}

var guidedGenders = []struct {
	Gender    party.Gender
	LocaleKey string
}{
	{party.GenderMale, "gender_male"},
	{party.GenderFemale, "gender_female"},
}

// guidedAlignments 是九個陣營編碼。spec 1066 只確認 (陣營+1) mod 3 = 0
// ——也就是 {2, 5, 8}——是邪惡；本表把它推成 AD&D 的九宮格
// 「陣營 = 守序軸×3 + 善惡軸」，善惡軸 0 善／1 中立／2 惡。
// ⚠ 這是 strong inference，不是 exact：原作的選單文字尚未取出佐證。
var guidedAlignments = []struct {
	Code      uint8
	LocaleKey string
}{
	{0, "alignment_lawful_good"},
	{1, "alignment_lawful_neutral"},
	{2, "alignment_lawful_evil"},
	{3, "alignment_neutral_good"},
	{4, "alignment_true_neutral"},
	{5, "alignment_neutral_evil"},
	{6, "alignment_chaotic_good"},
	{7, "alignment_chaotic_neutral"},
	{8, "alignment_chaotic_evil"},
}

// GuidedCreationOption 是目前這一段的一個可選項。Value 的意義依段別而定：
// 種族段是 DOS 種族編號、性別段是 0/1、職業段是職業組合編號、陣營段是陣營碼。
type GuidedCreationOption struct {
	Label string
	Value int
}

// BeginGuidedCreation 開始原版的四段建角流程。
func (s *State) BeginGuidedCreation() error {
	if s.Mode == ModeCombat {
		return fmt.Errorf("character creation is unavailable during combat")
	}
	s.creationReturnMode = s.Mode
	s.Mode = ModeCharacterCreation
	s.GuidedStep = CreationStepRace
	s.GuidedDraft = party.Character{Level: 1}
	s.GuidedClassCombos = nil
	s.CreationMessage = s.LocaleText("creation_pick_race")
	return nil
}

// GuidedCreationOptions 回傳目前這一段的選項。
func (s *State) GuidedCreationOptions() ([]GuidedCreationOption, error) {
	switch s.GuidedStep {
	case CreationStepRace:
		options := make([]GuidedCreationOption, 0, len(guidedRaces))
		for _, race := range guidedRaces {
			options = append(options, GuidedCreationOption{
				Label: s.LocaleText(race.LocaleKey), Value: race.DOSRaceID,
			})
		}
		return options, nil
	case CreationStepGender:
		options := make([]GuidedCreationOption, 0, len(guidedGenders))
		for _, gender := range guidedGenders {
			options = append(options, GuidedCreationOption{
				Label: s.LocaleText(gender.LocaleKey), Value: int(gender.Gender),
			})
		}
		return options, nil
	case CreationStepClass:
		// 可選職業由種族決定（spec 1099 §五）。
		tables, err := gamepack.Tables()
		if err != nil {
			return nil, err
		}
		raceID, ok := party.DOSRaceID(s.GuidedDraft.Race)
		if !ok {
			return nil, fmt.Errorf("draft race %d has no DOS race id", s.GuidedDraft.Race)
		}
		race, ok := tables.RaceByID(raceID)
		if !ok {
			return nil, fmt.Errorf("race %d is missing from the character tables", raceID)
		}
		s.GuidedClassCombos = append([]int(nil), race.ClassChoices...)
		options := make([]GuidedCreationOption, 0, len(race.ClassChoices))
		for _, combo := range race.ClassChoices {
			options = append(options, GuidedCreationOption{
				Label: s.classComboName(combo), Value: combo,
			})
		}
		return options, nil
	case CreationStepAlignment:
		options := make([]GuidedCreationOption, 0, len(guidedAlignments))
		for _, alignment := range guidedAlignments {
			options = append(options, GuidedCreationOption{
				Label: s.LocaleText(alignment.LocaleKey), Value: int(alignment.Code),
			})
		}
		return options, nil
	}
	return nil, fmt.Errorf("creation step %d has no menu", s.GuidedStep)
}

// SelectGuidedOption 選定目前這一段的第 index 個選項並推進到下一段。
func (s *State) SelectGuidedOption(index int) error {
	options, err := s.GuidedCreationOptions()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(options) {
		return fmt.Errorf("creation option %d is out of range", index)
	}
	value := options[index].Value
	switch s.GuidedStep {
	case CreationStepRace:
		s.GuidedDraft.Race = guidedRaces[index].Race
		s.GuidedStep = CreationStepGender
		s.CreationMessage = s.LocaleText("creation_pick_gender")
	case CreationStepGender:
		s.GuidedDraft.Gender = party.Gender(value)
		s.GuidedStep = CreationStepClass
		s.CreationMessage = s.LocaleText("creation_pick_class")
	case CreationStepClass:
		// 原作寫進 +75h 的是職業組合編號，不是選單索引。
		if err := applyClassCombination(&s.GuidedDraft, value); err != nil {
			return err
		}
		s.GuidedStep = CreationStepAlignment
		s.CreationMessage = s.LocaleText("creation_pick_alignment")
	case CreationStepAlignment:
		s.GuidedDraft.Alignment = uint8(value)
		s.GuidedDraft.AlignmentKnown = true
		s.GuidedStep = CreationStepAbilities
		s.CreationMessage = s.LocaleText("creation_ability_prompt")
	default:
		return fmt.Errorf("creation step %d does not take a menu selection", s.GuidedStep)
	}
	return nil
}

// classComboKeys 把職業組合編號對到 game pack 既有的顯示名 key
// （spec 1093 §二 的組合表）。德魯伊（1）與武僧（7）原作有槽位但
// 本作沒有可選職業，所以沒有對應的 key。
var classComboKeys = map[int]string{
	0: "creation.class.cleric", 2: "creation.class.fighter",
	3: "creation.class.paladin", 4: "creation.class.ranger",
	5: "creation.class.magic-user", 6: "creation.class.thief",
	8: "creation.class.cleric-fighter", 9: "creation.class.cleric-fighter-magic-user",
	10: "creation.class.cleric-ranger", 11: "creation.class.cleric-magic-user",
	12: "creation.class.cleric-thief", 13: "creation.class.fighter-magic-user",
	14: "creation.class.fighter-thief", 15: "creation.class.fighter-magic-user-thief",
	16: "creation.class.magic-user-thief",
}

// classComboName 取職業組合的顯示名。文字在 game pack 而不是 UI locale，
// 因為那是作品內容（AGENTS.md §2）。
func (s *State) classComboName(combo int) string {
	key, ok := classComboKeys[combo]
	if !ok {
		return fmt.Sprintf("class-combo-%d", combo)
	}
	if s.dataPack != nil {
		if text, ok := s.dataPack.Text(key, s.catalog.Language); ok && text != "" {
			return text
		}
	}
	return key
}

// classSlotToClass 把原作的職業槽索引換成本套件的 Class。
// 槽順序見 spec 1084：0 牧師 1 德魯伊 2 戰士 3 聖騎士 4 遊俠 5 法師 6 盜賊 7 武僧。
// ⚠ 德魯伊（1）與武僧（7）在本套件沒有對應的 Class。
func classSlotToClass(slot int) (party.Class, bool) {
	switch slot {
	case 0:
		return party.ClassCleric, true
	case 2:
		return party.ClassFighter, true
	case 3:
		return party.ClassPaladin, true
	case 4:
		return party.ClassRanger, true
	case 5:
		return party.ClassMagicUser, true
	case 6:
		return party.ClassThief, true
	}
	return 0, false
}

// applyClassCombination 依職業組合編號設定職業槽等級、主職業與起始經驗值。
// 組合語意來自 game pack（spec 1093 §二），不在 Go 裡重寫。
func applyClassCombination(draft *party.Character, combo int) error {
	tables, err := gamepack.Tables()
	if err != nil {
		return err
	}
	entry, ok := tables.CombinationByID(combo)
	if !ok {
		return fmt.Errorf("class combination %d is missing from the character tables", combo)
	}
	draft.RawClassID = uint8(combo)
	draft.ClassLevels = [8]uint8{}
	primarySet := false
	for _, slot := range entry.ClassSlots {
		if slot < 0 || slot >= len(draft.ClassLevels) {
			return fmt.Errorf("class combination %d has slot %d out of range", combo, slot)
		}
		draft.ClassLevels[slot] = 1
		if class, ok := classSlotToClass(slot); ok && !primarySet {
			draft.Class = class
			primarySet = true
		}
	}
	if !primarySet {
		return fmt.Errorf("class combination %d has no representable primary class", combo)
	}
	draft.Level = 1
	// 起始經驗值就是 25,000 平分（spec 1093 §二）。
	draft.Experience = uint32(entry.StartingExperience)
	return nil
}

// RollGuidedAbilities 擲六個屬性並依種族／性別／職業夾值。
// 原作是「擲完問 Reroll stats?，不滿意就重擲」（spec 1093 §一）。
func (s *State) RollGuidedAbilities(seed int64) error {
	if s.GuidedStep != CreationStepAbilities {
		return fmt.Errorf("ability roll is unavailable at creation step %d", s.GuidedStep)
	}
	limits, err := gamepack.LimitLookup()
	if err != nil {
		return err
	}
	raceID, ok := party.DOSRaceID(s.GuidedDraft.Race)
	if !ok {
		return fmt.Errorf("draft race %d has no DOS race id", s.GuidedDraft.Race)
	}
	bounds, ok := limits.AbilityLimits(raceID, int(s.GuidedDraft.Gender), int(s.GuidedDraft.RawClassID))
	if !ok {
		return fmt.Errorf("race %d has no ability limits", raceID)
	}
	rolled := party.RollAbilities(seed)
	// 擲出來的值夾進種族與職業允許的範圍。
	values := [6]int{rolled.Strength, rolled.Intelligence, rolled.Wisdom,
		rolled.Dexterity, rolled.Constitution, rolled.Charisma}
	for index := range values {
		if values[index] < bounds[index][0] {
			values[index] = bounds[index][0]
		}
		if values[index] > bounds[index][1] {
			values[index] = bounds[index][1]
		}
	}
	s.GuidedDraft.Abilities = party.Abilities{
		Strength: values[0], Intelligence: values[1], Wisdom: values[2],
		Dexterity: values[3], Constitution: values[4], Charisma: values[5],
	}
	s.CreationMessage = s.LocaleText("creation_reroll_prompt")
	return nil
}
