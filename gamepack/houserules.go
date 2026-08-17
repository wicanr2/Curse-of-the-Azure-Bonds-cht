package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// CoAB 專屬的**偏離規則**（spec 1129）。
//
// ★★ 這一份刻意**不**進共用 engine 的 pack schema。裡面的東西不是原作行為，
// 把它塞進共用 schema 等於邀請別的作品也長出同一個欄位；house rule 應該待在
// 它屬於的那一款遊戲裡。
//
// ⚠ 這裡的每一條都要寫清楚「為什麼偏離」與「誰決定的」，否則過幾輪之後
// 沒有人分得出它是還原還是自己加的。

// LevelDrainRule 宣告哪一隻怪的哪一個天生效果碼會吸等級、一次吸幾級。
type LevelDrainRule struct {
	ID string `json:"id"`
	// Monster 只是給人看的標記，比對用的是 EffectKind。
	Monster    string `json:"monster"`
	EffectKind uint8  `json:"effect_kind"`
	Levels     int    `json:"levels"`
	Reason     string `json:"reason"`
}

// HouseRules 是整份偏離規則。
type HouseRules struct {
	SchemaVersion int              `json:"schema_version"`
	Source        string           `json:"source"`
	Spec          string           `json:"spec"`
	LevelDrain    []LevelDrainRule `json:"level_drain"`
}

var (
	houseRulesOnce sync.Once
	houseRules     *HouseRules
	houseRulesErr  error
)

// LoadHouseRules 回傳嵌入的偏離規則。
func LoadHouseRules() (*HouseRules, error) {
	houseRulesOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/house-rules.json")
		if err != nil {
			houseRulesErr = fmt.Errorf("read embedded house rules: %w", err)
			return
		}
		parsed := &HouseRules{}
		if err := json.Unmarshal(data, parsed); err != nil {
			houseRulesErr = fmt.Errorf("parse embedded house rules: %w", err)
			return
		}
		for _, rule := range parsed.LevelDrain {
			if rule.EffectKind == 0 || rule.Levels <= 0 {
				houseRulesErr = fmt.Errorf("house rule %q needs a nonzero effect_kind and positive levels", rule.ID)
				return
			}
		}
		houseRules = parsed
	})
	return houseRules, houseRulesErr
}
