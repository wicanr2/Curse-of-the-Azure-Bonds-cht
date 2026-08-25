package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// 怪物特殊攻擊（spec 1202）。
//
// ★ 為什麼在這裡不在 engine：這五支是 CoAB overlay-22 尾端 handler 的轉錄
//（效果碼、命中率、次數、豁免類別都是本作的 asm 立即數），不是跨作品機制。
// 型別轉換與控制流在 `internal/combat/special_attack.go`。

// SpecialAttackDefinition 是 rules/special-attacks.json 的一筆。
type SpecialAttackDefinition struct {
	ID         string `json:"id"`
	EffectKind uint8  `json:"effect_kind"`
	Form       string `json:"form"`
	MissChance int    `json:"miss_chance,omitempty"`
	// MaxDistance 是 `entry#33`（半格 div 2）的放棄門檻。
	MaxDistance    int   `json:"max_distance,omitempty"`
	Uses           int   `json:"uses,omitempty"`
	DamageMask     uint8 `json:"damage_mask,omitempty"`
	ConstantDamage int   `json:"constant_damage,omitempty"`
	SaveCategory   int   `json:"save_category"`
	// TargetRange／AreaRadius 是近似值（選目標模式碼與 sub_175Bh 未讀）。
	TargetRange       int   `json:"target_range"`
	AreaRadius        int   `json:"area_radius,omitempty"`
	ParalysisEffect   uint8 `json:"paralysis_effect,omitempty"`
	ParalysisDuration int   `json:"paralysis_duration,omitempty"`
	// Message 家族是 locale 的 stable ID。
	Message          string `json:"message"`
	MissMessage      string `json:"miss_message,omitempty"`
	ParalyzedMessage string `json:"paralyzed_message,omitempty"`
}

// SpecialAttackTable 是整份表。
type SpecialAttackTable struct {
	SchemaVersion int                       `json:"schema_version"`
	Source        string                    `json:"source"`
	Spec          string                    `json:"spec"`
	Notes         []string                  `json:"notes"`
	Attacks       []SpecialAttackDefinition `json:"attacks"`
}

var (
	specialAttacksOnce sync.Once
	specialAttacks     *SpecialAttackTable
	specialAttacksErr  error
)

// SpecialAttacks 回傳嵌入的特殊攻擊表。
func SpecialAttacks() (*SpecialAttackTable, error) {
	specialAttacksOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/special-attacks.json")
		if err != nil {
			specialAttacksErr = fmt.Errorf("read embedded special attacks: %w", err)
			return
		}
		parsed := &SpecialAttackTable{}
		if err := json.Unmarshal(data, parsed); err != nil {
			specialAttacksErr = fmt.Errorf("parse embedded special attacks: %w", err)
			return
		}
		if len(parsed.Attacks) == 0 {
			specialAttacksErr = fmt.Errorf("special attack table is empty")
			return
		}
		for _, attack := range parsed.Attacks {
			if attack.ID == "" || attack.EffectKind == 0 || attack.Form == "" ||
				attack.Message == "" {
				specialAttacksErr = fmt.Errorf("special attack %q is missing required fields", attack.ID)
				return
			}
		}
		specialAttacks = parsed
	})
	return specialAttacks, specialAttacksErr
}
