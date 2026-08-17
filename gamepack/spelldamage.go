package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// 傷害／治療類法術的骰子（spec 1124）。
//
// ★ 屬性表的 `+0Ah` 為 0 的那一批，骰子不在表裡而在各自的 handler。
// 這份 JSON 由 `scripts/spell_damage_table.py` 從 handler 本體抽出來。
//
// ⚠ `shape` 記的是**呼叫了哪一支擲骰入口**，不是用途：`entry9` 既被治療輕傷用，
// 也被解除魔法拿去擲 `1d100`。用途要看該支法術自己。

// SpellDamageEntry 是一支法術的骰子摘要。
type SpellDamageEntry struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	CasterClass string `json:"caster_class"`
	CampOnly    bool   `json:"camp_only"`
	Overlay     int    `json:"overlay"`
	Entry       int    `json:"entry"`
	Offset      int    `json:"offset"`
	// Shape 是 `entry9`／`entry10`／`ambiguous`（handler 裡擲了不只一次）／
	// `unread`（沒有可辨識的擲骰）。
	Shape      string `json:"shape"`
	DiceCount  int    `json:"dice_count,omitempty"`
	DiceSides  int    `json:"dice_sides,omitempty"`
	// ScalesWithCasterLevel 是「骰數 0」＝ 用施法者等級當骰數。
	ScalesWithCasterLevel bool `json:"scales_with_caster_level,omitempty"`
}

// SpellDamageTable 是整份表。
type SpellDamageTable struct {
	SchemaVersion int                         `json:"schema_version"`
	Source        string                      `json:"source"`
	Spec          string                      `json:"spec"`
	Spells        map[string]SpellDamageEntry `json:"spells"`
}

var (
	spellDamageOnce sync.Once
	spellDamage     *SpellDamageTable
	spellDamageErr  error
)

// SpellDamage 回傳嵌入的傷害骰表。
func SpellDamage() (*SpellDamageTable, error) {
	spellDamageOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/spell-damage.json")
		if err != nil {
			spellDamageErr = fmt.Errorf("read embedded spell damage table: %w", err)
			return
		}
		parsed := &SpellDamageTable{}
		if err := json.Unmarshal(data, parsed); err != nil {
			spellDamageErr = fmt.Errorf("parse embedded spell damage table: %w", err)
			return
		}
		if len(parsed.Spells) == 0 {
			spellDamageErr = fmt.Errorf("spell damage table is empty")
			return
		}
		spellDamage = parsed
	})
	return spellDamage, spellDamageErr
}

// Dice 回傳一支法術的骰數與面數。`shape` 不是 `entry9`／`entry10` 時回 false
// ——那代表 handler 裡沒有唯一一次擲骰，數字不能拿來用。
func (t *SpellDamageTable) Dice(spellID uint8, casterLevel int) (count, sides int, ok bool) {
	if t == nil {
		return 0, 0, false
	}
	entry, found := t.Spells[fmt.Sprintf("%d", spellID)]
	if !found || (entry.Shape != "entry9" && entry.Shape != "entry10") {
		return 0, 0, false
	}
	count = entry.DiceCount
	if entry.ScalesWithCasterLevel {
		count = casterLevel
	}
	if count <= 0 || entry.DiceSides <= 0 {
		return 0, 0, false
	}
	return count, entry.DiceSides, true
}
