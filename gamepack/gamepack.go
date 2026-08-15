package gamepack

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

// Files is kept title-local: the reusable engine never imports CoAB data.
//
//go:embed events/*.json
var Files embed.FS

var (
	defaultOnce sync.Once
	defaultPack *goldenbox.Pack
	defaultErr  error
)

func Default() (*goldenbox.Pack, error) {
	defaultOnce.Do(func() {
		data, err := Files.ReadFile("events/pit-of-moander.json")
		if err != nil {
			defaultErr = fmt.Errorf("read embedded CoAB game pack: %w", err)
			return
		}
		defaultPack, defaultErr = goldenbox.LoadPackBytes(data)
	})
	return defaultPack, defaultErr
}

// CharacterTables 是原版常駐資料段裡的建角規則表（spec 1099），
// 由 cmd/dseg-export 從 DOS START.EXE 的資料段匯出。
//
// 種族索引是**原作的 `+74h` 編號 1..7**（1 矮人…6 半獸人 7 人類），
// 不是 internal/party 的 Race 常數（那個從 0 起算）。
type CharacterTables struct {
	Source string      `json:"source"`
	Spec   string      `json:"spec"`
	Races  []RaceRules `json:"races"`
	// ClassRequirements 依職業組合編號（角色^[75h]，0..0Ch）索引，
	// 每筆六個屬性最低要求，順序為力、智、睿、敏、體、魅。
	ClassRequirements [][]int            `json:"class_requirements"`
	ClassCombinations []ClassCombination `json:"class_combinations"`
	// ClassSlots 依職業槽索引 0..7，帶生命骰與起始金錢參數（spec 850／1101）。
	ClassSlots []ClassSlotRules `json:"class_slots"`
	// ConstitutionHPBonus 的索引 0 對應體質 ConstitutionHPBonusFrom（spec 869）。
	ConstitutionHPBonusFrom int   `json:"constitution_hp_bonus_from"`
	ConstitutionHPBonus     []int `json:"constitution_hp_bonus"`
	// 戰士系的額外體質加值（spec 869）：判斷用職業組合編號 +75h，
	// 所以只有單職戰士／聖騎士／遊俠拿得到，多職組合拿不到。
	FighterConstitutionCombos    []int `json:"fighter_constitution_combos"`
	FighterConstitutionBonusFrom int   `json:"fighter_constitution_bonus_from"`
	FighterConstitutionBonus     []int `json:"fighter_constitution_bonus"`
}

// ClassSlotRules 是單一職業槽的生命骰參數（spec 850）。
type ClassSlotRules struct {
	Slot            int    `json:"slot"`
	Name            string `json:"name"`
	HitDiceCount    int    `json:"hit_dice_count"`
	HitDiceSize     int    `json:"hit_dice_size"`
	HitDiceLevelCap int    `json:"hit_dice_level_cap"`
	// 起始金錢在原版被擲出來就丟掉（spec 1101 §四），保留只為亂數序列。
	StartingMoneyCount int `json:"starting_money_count"`
	StartingMoneySize  int `json:"starting_money_size"`
}

type RaceRules struct {
	RaceID int    `json:"race_id"`
	Name   string `json:"name"`
	// Selectable 是「建角種族選單列不列這一項」（spec 1102 §一）。
	// 半獸人資料齊全但選單沒有分支收它。
	Selectable bool `json:"selectable"`
	// Size 是原作的 +144h：1 小體型、2 中體型（spec 1093 §七）。
	Size           int            `json:"size"`
	StrengthMale   StrengthLimits `json:"strength_male"`
	StrengthFemale StrengthLimits `json:"strength_female"`
	Intelligence   AbilityRange   `json:"intelligence"`
	Wisdom         AbilityRange   `json:"wisdom"`
	Dexterity      AbilityRange   `json:"dexterity"`
	Constitution   AbilityRange   `json:"constitution"`
	Charisma       AbilityRange   `json:"charisma"`
	// ClassChoices 是「建角選單第 n 項 → 職業組合編號」的對照（spec 1093）。
	ClassChoices []int         `json:"class_choices"`
	StartingAges []StartingAge `json:"starting_ages"`
}

type AbilityRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type StrengthLimits struct {
	Min           int `json:"min"`
	Max           int `json:"max"`
	PercentileMax int `json:"percentile_max"`
}

// StartingAge 的 ClassSlot 用原作的職業槽名（spec 1084 的順序：
// 牧師／德魯伊／戰士／聖騎士／遊俠／法師／盜賊）。
type StartingAge struct {
	ClassSlot string `json:"class_slot"`
	BaseAge   int    `json:"base_age"`
	DiceCount int    `json:"dice_count"`
	DiceSize  int    `json:"dice_size"`
}

//go:embed rules/*.json
var ruleFiles embed.FS

var (
	tablesOnce sync.Once
	tables     *CharacterTables
	tablesErr  error
)

// Tables 回傳嵌入的建角規則表。
func Tables() (*CharacterTables, error) {
	tablesOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/character-tables.json")
		if err != nil {
			tablesErr = fmt.Errorf("read embedded character tables: %w", err)
			return
		}
		parsed := &CharacterTables{}
		if err := json.Unmarshal(data, parsed); err != nil {
			tablesErr = fmt.Errorf("parse embedded character tables: %w", err)
			return
		}
		tables = parsed
	})
	return tables, tablesErr
}

// RaceByID 依原作的 +74h 種族編號取一列。
func (t *CharacterTables) RaceByID(raceID int) (RaceRules, bool) {
	for _, race := range t.Races {
		if race.RaceID == raceID {
			return race, true
		}
	}
	return RaceRules{}, false
}

// StartingAgeFor 依種族編號與原作職業槽名取起始年齡。
// 找不到代表原作沒有為這個組合準備年齡（不是年齡 0）。
func (t *CharacterTables) StartingAgeFor(raceID int, classSlot string) (StartingAge, bool) {
	race, ok := t.RaceByID(raceID)
	if !ok {
		return StartingAge{}, false
	}
	for _, entry := range race.StartingAges {
		if entry.ClassSlot == classSlot {
			return entry, true
		}
	}
	return StartingAge{}, false
}

// StartingAgeLookup 讓 internal/party 不必認識本套件的型別就能查表
// （party.StartingAgeLookup 的實作）。
type StartingAgeLookup struct{ tables *CharacterTables }

// AgeLookup 回傳可注入 internal/party 的查詢器。
func AgeLookup() (StartingAgeLookup, error) {
	loaded, err := Tables()
	if err != nil {
		return StartingAgeLookup{}, err
	}
	return StartingAgeLookup{tables: loaded}, nil
}

func (l StartingAgeLookup) StartingAgeFor(raceID int, classSlot string) (int, int, int, bool) {
	if l.tables == nil {
		return 0, 0, 0, false
	}
	entry, ok := l.tables.StartingAgeFor(raceID, classSlot)
	if !ok {
		return 0, 0, 0, false
	}
	return entry.BaseAge, entry.DiceCount, entry.DiceSize, true
}

// AbilityLimitsFor 依原作的種族編號、性別（0 男 1 女）與職業組合編號，
// 回傳六個屬性各自的下限與上限。
//
// 規則來自 spec 1086／1099：
//   - 種族表給每個屬性一組（下限, 上限）；力量的下限／上限／百分比上限
//     三格再依性別分。
//   - 職業組合表給六個屬性的最低要求，`0` 代表無要求；兩者取較嚴的下限。
func (t *CharacterTables) AbilityLimitsFor(raceID, gender, classCombo int) ([6][2]int, bool) {
	race, ok := t.RaceByID(raceID)
	if !ok {
		return [6][2]int{}, false
	}
	strength := race.StrengthMale
	if gender == 1 {
		strength = race.StrengthFemale
	}
	limits := [6][2]int{
		{strength.Min, strength.Max},
		{race.Intelligence.Min, race.Intelligence.Max},
		{race.Wisdom.Min, race.Wisdom.Max},
		{race.Dexterity.Min, race.Dexterity.Max},
		{race.Constitution.Min, race.Constitution.Max},
		{race.Charisma.Min, race.Charisma.Max},
	}
	if classCombo >= 0 && classCombo < len(t.ClassRequirements) {
		for index, minimum := range t.ClassRequirements[classCombo] {
			if index >= len(limits) {
				break
			}
			if minimum > limits[index][0] {
				limits[index][0] = minimum
			}
		}
	}
	return limits, true
}

// StrengthPercentileMax 回傳該種族／性別的力量百分比上限（0 代表不能有 18/xx）。
func (t *CharacterTables) StrengthPercentileMax(raceID, gender int) (int, bool) {
	race, ok := t.RaceByID(raceID)
	if !ok {
		return 0, false
	}
	if gender == 1 {
		return race.StrengthFemale.PercentileMax, true
	}
	return race.StrengthMale.PercentileMax, true
}

// AbilityLimitLookup 是可注入 internal/party 的屬性上下限查詢器。
type AbilityLimitLookup struct{ tables *CharacterTables }

func LimitLookup() (AbilityLimitLookup, error) {
	loaded, err := Tables()
	if err != nil {
		return AbilityLimitLookup{}, err
	}
	return AbilityLimitLookup{tables: loaded}, nil
}

func (l AbilityLimitLookup) AbilityLimits(raceID, gender, classCombo int) ([6][2]int, bool) {
	if l.tables == nil {
		return [6][2]int{}, false
	}
	return l.tables.AbilityLimitsFor(raceID, gender, classCombo)
}

// ClassCombination 是職業組合編號的語意（spec 1093 §二）。
// 這張表不是資料段的表，來源是建角流程的 switch。
type ClassCombination struct {
	Combo int `json:"combo"`
	// ClassSlots 是原作職業槽索引：0 牧師 1 德魯伊 2 戰士 3 聖騎士
	// 4 遊俠 5 法師 6 盜賊 7 武僧。
	ClassSlots         []int `json:"class_slots"`
	StartingExperience int   `json:"starting_experience"`
	// Alignments 是這個組合能選的陣營編號（spec 1102 §二）。原作的陣營選單
	// 只列這幾項，例如聖騎士只有守序善一項。
	Alignments []int  `json:"alignments"`
	Note       string `json:"note,omitempty"`
}

// CombinationByID 依職業組合編號取語意。
func (t *CharacterTables) CombinationByID(combo int) (ClassCombination, bool) {
	for _, entry := range t.ClassCombinations {
		if entry.Combo == combo {
			return entry, true
		}
	}
	return ClassCombination{}, false
}

// SlotRules 依職業槽索引取生命骰參數。
func (t *CharacterTables) SlotRules(slot int) (ClassSlotRules, bool) {
	for _, entry := range t.ClassSlots {
		if entry.Slot == slot {
			return entry, true
		}
	}
	return ClassSlotRules{}, false
}

// ConstitutionHPBonusFor 依體質目前值（角色^[19h]）取 HP 加值（spec 869）。
// 超出表的有效範圍就夾到兩端——原作直接用體質當索引，取到的是相鄰資料，
// 這裡不重現那個越界行為。
func (t *CharacterTables) ConstitutionHPBonusFor(constitution int) int {
	if len(t.ConstitutionHPBonus) == 0 {
		return 0
	}
	index := constitution - t.ConstitutionHPBonusFrom
	if index < 0 {
		index = 0
	}
	if index >= len(t.ConstitutionHPBonus) {
		index = len(t.ConstitutionHPBonus) - 1
	}
	return t.ConstitutionHPBonus[index]
}

// FighterConstitutionBonusFor 是戰士系的額外體質加值（spec 869）。
// classCombo 是原作的 +75h；不在名單裡就沒有加值。
func (t *CharacterTables) FighterConstitutionBonusFor(classCombo, constitution int) int {
	eligible := false
	for _, combo := range t.FighterConstitutionCombos {
		if combo == classCombo {
			eligible = true
			break
		}
	}
	if !eligible || len(t.FighterConstitutionBonus) == 0 {
		return 0
	}
	index := constitution - t.FighterConstitutionBonusFrom
	if index < 0 {
		return 0
	}
	if index >= len(t.FighterConstitutionBonus) {
		index = len(t.FighterConstitutionBonus) - 1
	}
	return t.FighterConstitutionBonus[index]
}

// HitDiceLookup 把生命骰與體質加值兩組資料包給 internal/party，
// 讓規則資料留在 game pack、機制留在 Go。
type HitDiceLookup struct{ tables *CharacterTables }

// HitDice 取得 game pack 的生命骰查詢。
func HitDice() (HitDiceLookup, error) {
	parsed, err := Tables()
	if err != nil {
		return HitDiceLookup{}, err
	}
	return HitDiceLookup{tables: parsed}, nil
}

// HitDiceFor 回傳職業槽一級的骰數、面數與體質加值的等級上限。
func (l HitDiceLookup) HitDiceFor(slot int) (count, size, levelCap int, ok bool) {
	if l.tables == nil {
		return 0, 0, 0, false
	}
	entry, found := l.tables.SlotRules(slot)
	if !found {
		return 0, 0, 0, false
	}
	return entry.HitDiceCount, entry.HitDiceSize, entry.HitDiceLevelCap, true
}

// ConstitutionHPBonus 是一般職業的體質加值。
func (l HitDiceLookup) ConstitutionHPBonus(constitution int) int {
	if l.tables == nil {
		return 0
	}
	return l.tables.ConstitutionHPBonusFor(constitution)
}

// FighterConstitutionHPBonus 是戰士系職業組合的額外體質加值。
func (l HitDiceLookup) FighterConstitutionHPBonus(classCombo, constitution int) int {
	if l.tables == nil {
		return 0
	}
	return l.tables.FighterConstitutionBonusFor(classCombo, constitution)
}
