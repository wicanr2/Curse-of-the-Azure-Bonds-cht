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
	ClassRequirements [][]int `json:"class_requirements"`
}

type RaceRules struct {
	RaceID         int            `json:"race_id"`
	Name           string         `json:"name"`
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
