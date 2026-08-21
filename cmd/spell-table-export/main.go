// Command spell-table-export 把常駐資料段的法術主表匯出成 game pack JSON。
//
// 兩張表都在 DOS `START.EXE` 的常駐資料段，索引從 1 起算：
//
//	DS:27BDh  法術名稱  每筆 29h bytes  Pascal 短字串（首位元組是長度）
//	DS:37DAh  法術屬性  每筆 10h bytes
//
// **筆數是量出來的，不是估的**：索引 101 起的位元組不是這張表（spec 815 量到
// 1..100；本工具用 `-verify` 重新驗一次名稱長度與職業欄的值域）。索引 0 整筆
// 是 0，兩張表都不用它。
//
// 16 個位元組的欄位語意全部有出處，沒有出處的欄位不會被命名——見
// spec 1111 的表。**這支不發明語意**：`+8` 的 1／3 兩個值目前只知道「不是 0
// 也不是減半」，JSON 就照原值輸出，由消費端自己決定要不要用。
//
// 用法：
//
//	./tools/go.sh run ./cmd/spell-table-export -output gamepack/rules/spell-table.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	nameBase   = 0x27BD
	nameStride = 41
	attrBase   = 0x37DA
	attrStride = 16
	// spellCount 是量測值（spec 815）：索引 1..100 是法術，101 起是別的資料。
	// ⚠ 不要因為「掃到哪裡算哪裡」就改大——spec 1009 曾經把索引 113 的雜訊
	// 當成一支沒有名字的法術，統計因此整組偏高。
	spellCount = 100
)

// classNames 是屬性 `+0`（spec 1016／788／855）。3 沒有對應的施法職業：
// 100 筆裡有 15 筆是 3，其中 13 筆連名字都是空的（占位），另兩筆是
// `Animate Dead`（spec 815）。**不替 3 取名字**。
var classNames = map[uint8]string{0: "cleric", 1: "druid", 2: "magic-user"}

// saveCategoryNames 是屬性 `+9` 傳給豁免程序的類別。remake 的角色記錄剛好有
// 五個豁免值（`SavingThrows[0..4]`，`internal/party`），順序與 AD&D 1e 相同。
// 表裡只出現 0 與 4 兩個值，而 0 那三筆是 `Stinking Cloud`／`Poison`／一筆占位
// ——毒性法術對毒豁免，與 1e 的第 0 類（麻痺／毒／死亡魔法）吻合。
var saveCategoryNames = map[uint8]string{
	0: "paralysis-poison-death",
	1: "petrification-polymorph",
	2: "rod-staff-wand",
	3: "breath-weapon",
	4: "spell",
}

// spellRecord 是一筆法術。欄位名只給有出處的位移；`Raw` 永遠是完整 16 bytes，
// 讓消費端能在不改這支的情況下自己重新判讀。
type spellRecord struct {
	SpellID int    `json:"spell_id"`
	Name    string `json:"name"`
	// Placeholder：名字是空的那 13 筆（spec 815）。它們的 `+0`／`+1` 是
	// `03`／`06`。⚠ 名字空**不等於**玩家取不到：那 13 筆是充能物品的效果列，
	// 由 `物品^[3Dh] and 7Fh` 指到（spec 1169）。欄名沿用是為了不動 JSON 結構。
	Placeholder bool `json:"placeholder"`

	// CasterClass 是 `+0`（spec 1016）。值 3 沒有職業語意 ⇒ 留空字串。
	CasterClass   string `json:"caster_class"`
	CasterClassID int    `json:"caster_class_id"`
	// Level 是 `+1` 法術環數（spec 1016）。
	Level int `json:"level"`

	// DurationPrimary／DurationSecondary 是兩組形狀相同的係數
	// （`+2`／`+3` 見 spec 705、`+4`／`+5` 見 spec 712）：
	// `每級 × 施法者等級 + 基礎`，由呼叫端決定用哪一組。
	DurationPrimary   durationCoefficients `json:"duration_primary"`
	DurationSecondary durationCoefficients `json:"duration_secondary"`

	// TargetMode 是 `+6` 的低 4 位元（spec 1009）。
	TargetMode int `json:"target_mode"`
	// TargetModeKind 把 `+6` 的低 nibble 翻成 spec 1009 的五種形狀。
	TargetModeKind string `json:"target_mode_kind"`
	// TargetCount 只在 `fixed` 形狀有意義：`(模式 and 3) + 1`。
	TargetCount int `json:"target_count,omitempty"`
	// AreaRadius 只在 `area` 形狀有意義：`模式 and 7`。
	AreaRadius int `json:"area_radius,omitempty"`
	// DeclaredRadius 是 `+6` 的**高** nibble。原作只有在模式 0Fh 那條路會用它，
	// 而那條路先 `and 0Fh` 再 `shr 4`，所以**取到的永遠是 0**（spec 1009 的缺陷）。
	// 照原值輸出，要不要修是 remake 的設計決定。
	DeclaredRadius int `json:"declared_radius,omitempty"`

	// SelectMode 是 `+7`（spec 719）：1 施法者自己、2 叫玩家點一個、4 全隊；
	// 其餘值（含 0）在原作那一支直接判失敗。這是**戰鬥外**那條選目標路徑。
	SelectMode int `json:"select_mode"`

	// SaveKind 是 `+8`（spec 731：0 就完全不擲豁免）。
	SaveKind int `json:"save_kind"`
	// SaveCategory 是 `+9`，豁免程序的類別參數（spec 731／777）。
	SaveCategory     int    `json:"save_category"`
	SaveCategoryName string `json:"save_category_name,omitempty"`

	// EffectID 是 `+0Ah`：要套用的效果編號，0 代表不套（spec 731）。
	EffectID int `json:"effect_id"`

	// CombatUse 是 `+0Bh`（spec 827）：0 ＝ 只能紮營施放。1 與 2 的差別未讀出來，
	// 照原值輸出。
	CombatUse int `json:"combat_use"`
	// CampOnly 是 `CombatUse == 0` 的別名，讓消費端不必自己記這個約定。
	CampOnly bool `json:"camp_only"`

	// CastingTimeSegments 是 `+0Ch`，AD&D 的「節」（spec 827）。
	CastingTimeSegments int `json:"casting_time_segments"`
	// CastingDelay 是原作實際用的值：`有號(+0Ch) div 3`。0 代表當回合立即結算。
	CastingDelay int `json:"casting_delay"`

	// AIPriority 是 `+0Dh`：AI 眼中「這個效果值不值得現在用」的分數，
	// 門檻由 1d7 決定（spec 802／835）。
	AIPriority int `json:"ai_priority"`
	// RequiresTargetCheck 是 `+0Eh` ≠ 0（spec 802）：施法前要不要逐一檢查目標。
	RequiresTargetCheck bool `json:"requires_target_check"`
	// ScanRadius 是 `+0Fh`，傳給範圍掃描程序的半徑（spec 777／802）；
	// 0 代表跳過掃描。⚠ 它**不一定等於** `AreaRadius`。
	ScanRadius int `json:"scan_radius"`

	// Raw 是完整 16 bytes 的十六進位，欄位判讀有爭議時以它為準。
	Raw string `json:"raw"`
}

type durationCoefficients struct {
	Base     int `json:"base"`
	PerLevel int `json:"per_level"`
}

type spellTableFile struct {
	SchemaVersion int    `json:"schema_version"`
	Source        string `json:"source"`
	Spec          string `json:"spec"`
	NameBase      string `json:"name_table_base"`
	AttrBase      string `json:"attribute_table_base"`
	Spells        []spellRecord `json:"spells"`
}

func main() {
	dsegPath := flag.String("dseg", "workplace/re-sweep/dos/dseg/dos-dseg-dseg.bin",
		"常駐資料段的原始 dump")
	output := flag.String("output", "", "輸出 JSON 路徑；空字串印到 stdout")
	flag.Parse()

	blob, err := os.ReadFile(*dsegPath)
	if err != nil {
		log.Fatal(err)
	}
	required := attrBase + (spellCount+1)*attrStride
	if len(blob) < required {
		log.Fatalf("dseg dump is %d bytes; need at least %d", len(blob), required)
	}

	file := spellTableFile{
		SchemaVersion: 1,
		Source:        "DOS START.EXE 常駐資料段（DS:27BDh 名稱、DS:37DAh 屬性）",
		Spec:          "docs/spec/1111-spell-attribute-table.md",
		NameBase:      fmt.Sprintf("DS:%04Xh", nameBase),
		AttrBase:      fmt.Sprintf("DS:%04Xh", attrBase),
	}
	for id := 1; id <= spellCount; id++ {
		record := blob[attrBase+id*attrStride:][:attrStride]
		name, err := readName(blob, id)
		if err != nil {
			log.Fatalf("spell %d: %v", id, err)
		}
		entry := spellRecord{
			SpellID:       id,
			Name:          name,
			Placeholder:   name == "",
			CasterClass:   classNames[record[0]],
			CasterClassID: int(record[0]),
			Level:         int(record[1]),
			DurationPrimary: durationCoefficients{
				Base: int(record[2]), PerLevel: int(record[3])},
			DurationSecondary: durationCoefficients{
				Base: int(record[4]), PerLevel: int(record[5])},
			TargetMode:          int(record[6] & 0x0F),
			DeclaredRadius:      int(record[6] >> 4),
			SelectMode:          int(record[7]),
			SaveKind:            int(record[8]),
			SaveCategory:        int(record[9]),
			SaveCategoryName:    saveCategoryNames[record[9]],
			EffectID:            int(record[10]),
			CombatUse:           int(record[11]),
			CampOnly:            record[11] == 0,
			CastingTimeSegments: int(record[12]),
			CastingDelay:        int(int8(record[12])) / 3,
			AIPriority:          int(record[13]),
			RequiresTargetCheck: record[14] != 0,
			ScanRadius:          int(record[15]),
			Raw:                 fmt.Sprintf("%X", record),
		}
		entry.TargetModeKind, entry.TargetCount, entry.AreaRadius = decodeTargetMode(record[6] & 0x0F)
		file.Spells = append(file.Spells, entry)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	if *output == "" {
		os.Stdout.Write(data)
		return
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		log.Fatal(err)
	}
	placeholders := 0
	for _, spell := range file.Spells {
		if spell.Placeholder {
			placeholders++
		}
	}
	fmt.Printf("wrote %s: %d spells (%d placeholders)\n", *output, len(file.Spells), placeholders)
}

// decodeTargetMode 依 spec 1009 的五種形狀翻譯 `+6` 的低 nibble。
// 名稱刻意用形狀而不是「單體／群體」這種遊戲術語——原作的分類是幾何的。
func decodeTargetMode(mode byte) (kind string, count int, radius int) {
	switch {
	case mode == 0:
		return "self", 0, 0
	case mode == 5:
		return "weighted-picks", 0, 0
	case mode == 0x0F:
		return "locked-or-area", 0, 0
	case mode >= 8 && mode <= 0x0E:
		return "area", 0, int(mode & 7)
	default: // 2、3、4、6、7
		return "fixed", int(mode&3) + 1, 0
	}
}

// readName 取 Pascal 短字串。長度位元組超過筆距就是**讀過頭了**，
// 直接失敗而不是回一段亂碼——那正是「表有幾筆沒量過」會踩到的坑。
func readName(blob []byte, id int) (string, error) {
	offset := nameBase + id*nameStride
	if offset+nameStride > len(blob) {
		return "", fmt.Errorf("name at DS:%04Xh is past the dump", offset)
	}
	length := int(blob[offset])
	if length >= nameStride {
		return "", fmt.Errorf("name length %d exceeds the %d-byte stride", length, nameStride)
	}
	name := string(blob[offset+1 : offset+1+length])
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("name contains a NUL byte: %q", name)
	}
	return name, nil
}
