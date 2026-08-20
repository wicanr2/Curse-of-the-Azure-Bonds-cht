package game

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 原版建角是四段選單接一次擲屬性（spec 1093 §一）：
//
//	選單('Pick Race')      → 角色^[74h]  := 種族
//	選單('Pick Gender')    → 角色^[119h] := 性別
//	選單('Pick Class')     → 角色^[75h]  := byte[DS:3FF8h + 種族×0Eh + 選單索引]
//	選單('Pick Alignment') → 角色^[11Bh] := 陣營
//	repeat …擲屬性…決定 HP… 問('Reroll stats? ') until 答 = 'N'
//	問('Character name: ')
//	基準值 := 目前值
//	if 問('Save <名字>? ') <> 'N' then 存檔
//
// ⚠ 兩個順序容易擺錯（`overlay-17:00782h` 的 15CAh 迴圈邊界）：
// **HP 在重擲迴圈裡面**（每擲一次屬性就重算一次），
// **基準值複製在名字輸入之後**（不是擲完屬性就抄）。
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
	CreationStepName
	CreationStepSave
	CreationStepDone
)

// guidedRaces 依原作的 +74h 種族編號 1..7 排列（spec 1084／884）。
// 原作的種族字串表還有索引 0 = Monster（spec 1100），那是非玩家種族，
// 建角選單不列。
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

// guidedAlignments 是九個陣營編碼，順序由原作字串表確認（spec 1100）：
// 陣營 = 守序軸×3 + 善惡軸（善惡軸 0 善／1 中立／2 惡），
// 索引 4 的原作用字是 True Neutral。{2, 5, 8} 正好是三個 Evil，
// 與 spec 1066 的「(陣營+1) mod 3 = 0 ⇒ 邪惡」吻合。
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

// raceByDOSID 把原作的 +74h 種族編號換回本套件的 Race。
func raceByDOSID(dosRaceID int) (party.Race, bool) {
	for _, race := range guidedRaces {
		if race.DOSRaceID == dosRaceID {
			return race.Race, true
		}
	}
	return 0, false
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
	s.GuidedActive = true
	s.GuidedCursor = 0
	s.CreationMessage = s.LocaleText("creation_pick_race")
	return nil
}

// GuidedCreationOptions 回傳目前這一段的選項。
func (s *State) GuidedCreationOptions() ([]GuidedCreationOption, error) {
	switch s.GuidedStep {
	case CreationStepRace:
		// 選單只列 selectable 的種族：原作的顯示迴圈沒有分支收半獸人
		// （spec 1102 §一），所以建角取不到它。
		tables, err := gamepack.Tables()
		if err != nil {
			return nil, err
		}
		options := make([]GuidedCreationOption, 0, len(guidedRaces))
		for _, race := range guidedRaces {
			entry, ok := tables.RaceByID(race.DOSRaceID)
			if !ok || !entry.Selectable {
				continue
			}
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
		// 可選陣營由職業組合決定（spec 1102 §二）：聖騎士只有守序善、
		// 遊俠只有三個善、盜賊系少掉兩個善。
		tables, err := gamepack.Tables()
		if err != nil {
			return nil, err
		}
		combo, ok := tables.CombinationByID(int(s.GuidedDraft.RawClassID))
		if !ok {
			return nil, fmt.Errorf("class combination %d is missing from the character tables",
				s.GuidedDraft.RawClassID)
		}
		options := make([]GuidedCreationOption, 0, len(combo.Alignments))
		for _, code := range combo.Alignments {
			if code < 0 || code >= len(guidedAlignments) {
				return nil, fmt.Errorf("class combination %d allows unknown alignment %d",
					combo.Combo, code)
			}
			options = append(options, GuidedCreationOption{
				Label: s.LocaleText(guidedAlignments[code].LocaleKey), Value: code,
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
	s.GuidedCursor = 0
	switch s.GuidedStep {
	case CreationStepRace:
		// 選單索引與種族編號不再一一對應（半獸人被跳過），一律用 Value。
		race, ok := raceByDOSID(value)
		if !ok {
			return fmt.Errorf("creation option gave unknown DOS race id %d", value)
		}
		s.GuidedDraft.Race = race
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
	rolls, err := gamepack.AbilityRoll()
	if err != nil {
		return err
	}
	// 擲點是 `3d6 + 1`、每個屬性六次取最大、種族調整加在每一次上（spec 1103）。
	values, err := party.RollCreationAbilities(rolls, raceID, seed)
	if err != nil {
		return err
	}
	// 擲出來的值夾進種族與職業允許的範圍。
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
	// HP 也在重擲迴圈裡（`overlay-17:00782h` 的 15CAh 迴圈，spec 1101）：
	// 屬性重擲一次，體質變了，HP 就跟著重算。
	if err := s.rollGuidedHitPoints(seed); err != nil {
		return err
	}
	s.CreationMessage = s.LocaleText("creation_reroll_prompt")
	return nil
}

// rollGuidedHitPoints 依 spec 1101 決定 +78h／+12Ch／+1A4h 三格。
func (s *State) rollGuidedHitPoints(seed int64) error {
	dice, err := gamepack.HitDice()
	if err != nil {
		return err
	}
	// 體質用的是**目前值**（角色^[19h]，spec 869）。
	constitution, err := s.GuidedDraft.Abilities.CurrentValue(4)
	if err != nil {
		return err
	}
	points, err := party.RollCreationHitPoints(dice, s.GuidedDraft.ClassLevels,
		int(s.GuidedDraft.RawClassID), constitution, seed)
	if err != nil {
		return err
	}
	s.GuidedDraft.MaxHitPoints = points.MaxHitPoints
	s.GuidedDraft.BaseMaxHitPoints = points.BaseMaxHitPoints
	// 原作直接補滿：角色^[1A4h] := 角色^[78h]。
	s.GuidedDraft.HitPoints = points.MaxHitPoints
	// 命中能力 `+73h` 與護甲起點 `+124h`（spec 1140）。原作建角先寫兩個
	// 預設值，訓練場再依職業與等級重算 `+73h`；一級角色兩條路的差別只在
	// 法師與盜賊那兩張表，所以這裡直接查表。
	combatBase, err := gamepack.CombatBase()
	if err != nil {
		return err
	}
	attack, armor, err := party.CreationCombatBase(combatBase, s.GuidedDraft.ClassLevels)
	if err != nil {
		return err
	}
	s.GuidedDraft.AttackAbility = attack
	s.GuidedDraft.BaseArmorClass = armor
	// `+125h`：建角寫 1，力量的命中／傷害調整才會生效（spec 694／697）。
	s.GuidedDraft.AbilityAdjustments = 1
	return nil
}

// AcceptGuidedAbilities 是 `Reroll stats? ` 答 'N' 那一步：離開重擲迴圈，
// 進入名字輸入。年齡在這裡決定——原作是在陣營選完之後（`1230h`）就算好的，
// 但它只看種族與職業，兩者在進入重擲迴圈前就已固定，所以擺在這裡結果相同
// （spec 1093 §五）。
func (s *State) AcceptGuidedAbilities(seed int64) error {
	if s.GuidedStep != CreationStepAbilities {
		return fmt.Errorf("cannot accept abilities at creation step %d", s.GuidedStep)
	}
	if s.GuidedDraft.Abilities == (party.Abilities{}) {
		return fmt.Errorf("abilities have not been rolled yet")
	}
	ages, err := gamepack.AgeLookup()
	if err != nil {
		return err
	}
	age, err := party.RollStartingAgeFrom(ages, s.GuidedDraft.Race, s.GuidedDraft.Class, seed)
	if err != nil {
		return err
	}
	s.GuidedDraft.Age = age

	s.GuidedName = ""
	s.GuidedStep = CreationStepName
	s.CreationMessage = s.LocaleText("creation_name_prompt")
	return nil
}

// AppendGuidedName 把輸入的字加進名字。上限是 remake 自己的 20 字元
// ——原版的欄位只有 15 個，但 remake 有自己的存檔格式（只需要能讀原版存檔）。
func (s *State) AppendGuidedName(chars []rune) error {
	if s.GuidedStep != CreationStepName {
		return fmt.Errorf("name input is unavailable at creation step %d", s.GuidedStep)
	}
	if utf8.RuneCountInString(s.GuidedName)+len(chars) > guidedNameLimit {
		return fmt.Errorf("character name is limited to %d characters", guidedNameLimit)
	}
	s.GuidedName += string(chars)
	return nil
}

// BackspaceGuidedName 刪掉名字的最後一個字（依字元而不是位元組，中文才不會斷）。
func (s *State) BackspaceGuidedName() error {
	if s.GuidedStep != CreationStepName {
		return fmt.Errorf("name input is unavailable at creation step %d", s.GuidedStep)
	}
	name := []rune(s.GuidedName)
	if len(name) > 0 {
		s.GuidedName = string(name[:len(name)-1])
	}
	return nil
}

// CommitGuidedName 收下名字並把目前值抄成基準值（spec 1093 §六之二）。
// ⚠ 這個複製在**名字之後**才做，不是擲完屬性就做——原作的順序是
// 名字輸入 → 基準值複製 → `Save <名字>?`。
func (s *State) CommitGuidedName() error {
	if s.GuidedStep != CreationStepName {
		return fmt.Errorf("name input is unavailable at creation step %d", s.GuidedStep)
	}
	// 原作在名字為空時直接重問（`2441h` 的 `jz loc_240D`）。
	if strings.TrimSpace(s.GuidedName) == "" {
		return fmt.Errorf("character name cannot be empty")
	}
	s.GuidedDraft.Name = s.GuidedName
	s.GuidedDraft.Abilities.SyncBaseFromCurrent()
	s.GuidedStep = CreationStepSave
	s.CreationMessage = fmt.Sprintf(s.LocaleText("creation_save_prompt"), s.GuidedName)
	return nil
}

// ConfirmGuidedSave 是原作最後那句 `Save <名字>? `：回答 'N' 就把角色丟掉，
// 其餘任何鍵都存檔（`24BBh` 只比對 'N'）。
func (s *State) ConfirmGuidedSave(save bool) error {
	if s.GuidedStep != CreationStepSave {
		return fmt.Errorf("save prompt is unavailable at creation step %d", s.GuidedStep)
	}
	if !save {
		// 原作直接 FreeMem 離開，角色不留。
		message := s.LocaleText("creation_discarded")
		if err := s.CancelGuidedCreation(); err != nil {
			return err
		}
		s.CreationMessage = message
		return nil
	}
	if len(s.CreationRoster) >= 6 {
		return fmt.Errorf("party already has six characters")
	}
	character := s.GuidedDraft
	character.ID = fmt.Sprintf("guided-%d", len(s.CreationRoster)+1)
	s.CreationRoster = append(s.CreationRoster, character)
	// 原作一次呼叫只建一個角色，存完就回上一層。
	message := fmt.Sprintf(s.LocaleText("creation_added"),
		character.Name, len(s.CreationRoster))
	if err := s.CancelGuidedCreation(); err != nil {
		return err
	}
	s.GuidedStep = CreationStepDone
	s.CreationMessage = message
	return nil
}

// guidedNameLimit 是 remake 的名字上限。原版是 15（`0Fh`）。
const guidedNameLimit = 20

// MoveGuidedCursor 在目前這一段的選項之間移動，兩端環繞——原作四個選單
// 都是同一組上下鍵操作（spec 1093 §一）。
func (s *State) MoveGuidedCursor(delta int) error {
	options, err := s.GuidedCreationOptions()
	if err != nil {
		return err
	}
	if len(options) == 0 {
		return fmt.Errorf("creation step %d has no options", s.GuidedStep)
	}
	s.GuidedCursor = (s.GuidedCursor + delta + len(options)) % len(options)
	return nil
}

// CancelGuidedCreation 離開四段流程，回到進入前的模式。
func (s *State) CancelGuidedCreation() error {
	s.GuidedActive = false
	s.GuidedStep = CreationStepRace
	s.Mode = s.creationReturnMode
	if s.Mode == ModeCharacterCreation || s.Mode == ModeTitle {
		s.Mode = ModeWilderness
	}
	return nil
}
