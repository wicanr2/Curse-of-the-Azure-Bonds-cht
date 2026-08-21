package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// SpellTable 是原作 100 筆法術主表（`rules/spell-table.json`，由
// `cmd/spell-table-export` 從常駐資料段產生）。
//
// ★ 為什麼要整張表，而不是只放 remake 有 handler 的那幾支。 施法時間、能不能
// 在戰鬥中施放、豁免類別、持續時間係數這些**全部是資料**，原作沒有一支法術是
// 靠程式碼寫死這些值的。只搬有 handler 的那幾支，等於把資料重打一次——而重打
// 的版本會漂移。整張表放進來之後，「還沒實作的是哪幾支」才問得出來。
type SpellTable struct {
	SchemaVersion      int          `json:"schema_version"`
	Source             string       `json:"source"`
	Spec               string       `json:"spec"`
	NameTableBase      string       `json:"name_table_base"`
	AttributeTableBase string       `json:"attribute_table_base"`
	Spells             []SpellEntry `json:"spells"`
}

// SpellEntry 的欄位語意見 `cmd/spell-table-export` 與 spec 1111。
// 這裡刻意不重寫一次註解——兩份會漂移。
type SpellEntry struct {
	SpellID       int    `json:"spell_id"`
	Name          string `json:"name"`
	Placeholder   bool   `json:"placeholder"`
	CasterClass   string `json:"caster_class"`
	CasterClassID int    `json:"caster_class_id"`
	Level         int    `json:"level"`

	DurationPrimary   SpellDuration `json:"duration_primary"`
	DurationSecondary SpellDuration `json:"duration_secondary"`

	TargetMode     int    `json:"target_mode"`
	TargetModeKind string `json:"target_mode_kind"`
	TargetCount    int    `json:"target_count,omitempty"`
	AreaRadius     int    `json:"area_radius,omitempty"`
	DeclaredRadius int    `json:"declared_radius,omitempty"`

	SelectMode int `json:"select_mode"`

	SaveKind         int    `json:"save_kind"`
	SaveCategory     int    `json:"save_category"`
	SaveCategoryName string `json:"save_category_name,omitempty"`

	EffectID int `json:"effect_id"`

	CombatUse int  `json:"combat_use"`
	CampOnly  bool `json:"camp_only"`

	CastingTimeSegments int `json:"casting_time_segments"`
	CastingDelay        int `json:"casting_delay"`

	AIPriority          int  `json:"ai_priority"`
	RequiresTargetCheck bool `json:"requires_target_check"`
	ScanRadius          int  `json:"scan_radius"`

	Raw string `json:"raw"`
}

// SpellDuration 是「每級 × 施法者等級 ＋ 基礎」。原作有兩組係數，
// 由呼叫端決定用哪一組（spec 705／712）。
type SpellDuration struct {
	Base     int `json:"base"`
	PerLevel int `json:"per_level"`
}

// 原作的兩個「補成 1」修正與 byte 回捲（spec 705 `overlay-22:0D67h`）。
const (
	// spellDurationSentinel：算出來是 0FFh 時原作補成 1。
	spellDurationSentinel = 0xFF
	// spellItemCasterLevel：`DS:7563h` 非 0（效果來自物品，spec 733／1016）時，
	// 等級一律當 6，不查施法者。
	spellItemCasterLevel = 6
)

// PrimaryDuration 重現 spec 705 的算式，**含 byte 回捲**。
//
// ⚠ 回捲不是 bug 修一修就好：`每級 × 等級` 超過 255 會繞回去，而原版的存檔與
// 攻略都是在有回捲的版本上累積的。要不要夾限是 remake 的設計決定，這支只忠實
// 重現原作，夾限請在呼叫端做並記錄決定。
func (s SpellEntry) PrimaryDuration(casterLevel int, fromItem bool) int {
	level := casterLevel
	if fromItem {
		level = spellItemCasterLevel
	}
	value := byte(s.DurationPrimary.PerLevel*level + s.DurationPrimary.Base)
	if value == 0 && s.TargetModeByte() != 0 {
		return 1
	}
	if value == spellDurationSentinel {
		return 1
	}
	return int(value)
}

// SecondaryDuration 是 spec 712 的第二組係數，形狀相同：
// `+5 × 等級 + 4`。原作在這一組沒有那兩個「補成 1」的修正。
func (s SpellEntry) SecondaryDuration(casterLevel int) int {
	return int(byte(s.DurationSecondary.PerLevel*casterLevel + s.DurationSecondary.Base))
}

// TargetModeByte 還原 `+6` 的完整位元組（低 nibble 是目標模式、高 nibble 是
// 宣告半徑）。spec 705 的「至少撐 1」判的是**整個位元組**非零，不是模式。
func (s SpellEntry) TargetModeByte() int { return s.DeclaredRadius<<4 | s.TargetMode }

// RequiresSave 依 spec 731：`+8 = 0` 時原作連擲都不擲。
func (s SpellEntry) RequiresSave() bool { return s.SaveKind != 0 }

// SaveHalvesDamage 是 `+8 = 2`。這個值同時出現在噴吐武器的傷害呼叫
// （spec 720／723／725／735 的第三個參數都是 2），而 `+8 = 2` 的六支法術是
// Fireball／Lightning Bolt／Flame Strike／Cone of Cold 與兩筆占位——AD&D 1e
// 裡正好都是「豁免成功傷害減半」。
//
// `overlay-23 entry#20`（DOS `overlay-23:1FD6h`）已經讀完（spec 1061／581）：
// 有做存活檢定時依這個值分兩支——**1 把傷害歸 0、2 除以 2**，其餘值什麼都不做。
// 所以 `= 1` 有語意（豁免就沒事，`internal/combat` 的 saveNegatesKind），
// 而全表兩筆的 `= 3` **沒有任何分支**——不要為了對稱補上去。
func (s SpellEntry) SaveHalvesDamage() bool { return s.SaveKind == 2 }

// 檔案由 gamepack.go 的 `//go:embed rules/*.json` 一起收進 ruleFiles。

var (
	spellTableOnce sync.Once
	spellTable     *SpellTable
	spellTableErr  error
)

// Spells 回傳嵌入的法術主表。
func Spells() (*SpellTable, error) {
	spellTableOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/spell-table.json")
		if err != nil {
			spellTableErr = fmt.Errorf("read embedded spell table: %w", err)
			return
		}
		parsed := &SpellTable{}
		if err := json.Unmarshal(data, parsed); err != nil {
			spellTableErr = fmt.Errorf("parse embedded spell table: %w", err)
			return
		}
		spellTable = parsed
	})
	return spellTable, spellTableErr
}

// SpellByID 依原作編號取一筆。編號是**索引**，占位那 13 筆一樣查得到——
// 玩家取不到它們，但存檔裡的位元組可能出現，靜靜回 false 會把匯入問題藏起來。
func SpellByID(spellID int) (SpellEntry, bool) {
	table, err := Spells()
	if err != nil || table == nil {
		return SpellEntry{}, false
	}
	if spellID < 1 || spellID > len(table.Spells) {
		return SpellEntry{}, false
	}
	entry := table.Spells[spellID-1]
	if entry.SpellID != spellID {
		return SpellEntry{}, false
	}
	return entry, true
}
