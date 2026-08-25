// Command dseg-export 把常駐資料段裡的建角規則表匯出成 game pack JSON 片段。
//
// 原始位元組來自 tools/ida/dump_data_segments.py 倒出的 dseg dump，
// DS:xxxx 就是該檔的位移。四張表的基底與筆距見 spec 1099：
//
//	DS:3F86h  種族屬性上下限   每列 10h bytes   索引 = 種族 1..7
//	DS:3FF8h  種族可選職業     每列 0Eh bytes   索引 = 種族 1..7
//	DS:404Ch  起始年齡         每列 1Ch bytes   索引 = 種族 1..7，欄 = 職業槽 0..6
//	DS:4172h  職業組合最低要求 每筆 6 bytes     索引 = 職業組合 0..0Ch
//
// 種族編號是 1..7（spec 1084／884），0 不是合法值；三張表在資料段連續相鄰，
// 所以不能用「索引 0」去取任何一張表。
//
// 用法：
//
//	./tools/go.sh run ./cmd/dseg-export -dseg <path> -output <path>
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
)

const (
	abilityLimitsBase = 0x3F86
	abilityLimitsSize = 0x10
	classChoicesBase  = 0x3FF8
	classChoicesSize  = 0x0E
	startingAgeBase   = 0x404C
	startingAgeSize   = 0x1C
	classRequireBase  = 0x4172
	classRequireSize  = 6
	// DS:41D8h 職業組合可選陣營，每列 0Ah bytes，[0] 是筆數（spec 884／1102）。
	alignmentBase = 0x41D8
	alignmentSize = 0x0A
	// 生命骰四張平行表，每張 8 bytes，索引 = 職業槽（spec 850）。
	hitDiceCountBase = 0x0818
	hitDiceSizeBase  = 0x0820
	hitDiceCapBase   = 0x3EB9
	// DS:45B1h／45B2h 起始金錢的骰子參數，交錯排列，筆距 2（spec 1101 §四）。
	startingMoneyBase = 0x45B1
	// DS:427Fh 體質 → HP 加值，直接用體質值當索引（spec 869）。有效範圍 3..25。
	constitutionBonusBase = 0x427F
	constitutionMin       = 3
	constitutionMax       = 25

	firstRace = 1 // spec 1084／884：+74h 的種族編號 1..7，0 不是合法值
	lastRace  = 7
	// 職業槽順序（spec 1084 第五次確認）：牧師／德魯伊／戰士／聖騎士／
	// 遊俠／法師／盜賊／武僧。年齡表只排到盜賊，共 7 欄。
	ageColumns = 7
	// 職業組合編號 0..10h ＝ 17 筆（spec 1093）。第 18 筆（41D8h）是
	// `9 0 1 2 3 4`，遞增序列，明顯不屬於本表 ⇒ 表到 10h 為止。
	classCombinations = 17
)

// raceNames 只用於產出時的可讀註記，索引即原作的 +74h 種族編號。
var raceNames = map[int]string{
	1: "dwarf", 2: "elf", 3: "gnome", 4: "half-elf",
	5: "halfling", 6: "half-orc", 7: "human",
}

// classSlotNames 是年齡表的欄名，順序見 spec 1084。
var classSlotNames = [ageColumns]string{
	"cleric", "druid", "fighter", "paladin", "ranger", "magic-user", "thief",
}

// allClassSlotNames 是完整的八個職業槽（spec 1084）。年齡表只排到盜賊，
// 生命骰與體質加值則八個都有。
var allClassSlotNames = [8]string{
	"cleric", "druid", "fighter", "paladin", "ranger", "magic-user", "thief", "monk",
}

// selectableRaces 是建角種族選單實際列出的六項（spec 1102 §一）。
// 6（半獸人）資料齊全但選單沒有分支收它 ⇒ 建角取不到。
var selectableRaces = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 7: true}

// raceSizes 是 +144h，來自建角流程的 case 分支而非資料表（spec 1093 §七）。
// 1 ＝ 小體型、2 ＝ 中體型；6／7 走同一個 else。
var raceSizes = map[int]int{1: 1, 2: 2, 3: 1, 4: 2, 5: 1, 6: 2, 7: 2}

// raceAbilityAdjustments 是建角擲點時套的種族調整（spec 1103 §三）。
// 這也是程式碼裡的 case 分支，不是資料段的表。
// 六個位置依序是力、智、睿、敏、體、魅；地精／半精靈／人類走 else，全 0。
var raceAbilityAdjustments = map[int][6]int{
	1: {0, 0, 0, 0, 1, -1}, // 矮人
	2: {0, 0, 0, 1, -1, 0}, // 精靈
	3: {},                  // 地精
	4: {},                  // 半精靈
	5: {-1, 0, 0, 1, 0, 0}, // 半身人
	6: {1, 0, 0, 0, 1, -2}, // 半獸人
	7: {},                  // 人類
}

type abilityRange struct {
	Min uint8 `json:"min"`
	Max uint8 `json:"max"`
}

type strengthLimits struct {
	Min           uint8 `json:"min"`
	Max           uint8 `json:"max"`
	PercentileMax uint8 `json:"percentile_max"`
}

type startingAge struct {
	ClassSlot string `json:"class_slot"`
	BaseAge   uint16 `json:"base_age"`
	DiceCount uint8  `json:"dice_count"`
	DiceSize  uint8  `json:"dice_size"`
}

type raceRules struct {
	RaceID int    `json:"race_id"`
	Name   string `json:"name"`
	// Selectable 是「建角種族選單列不列這一項」（spec 1102 §一）。
	Selectable bool `json:"selectable"`
	// Size 是 +144h：1 小體型、2 中體型（spec 1093 §七）。
	Size int `json:"size"`
	// AbilityAdjustments 依力、智、睿、敏、體、魅六個位置（spec 1103 §三）。
	AbilityAdjustments []int          `json:"ability_adjustments"`
	StrengthMale       strengthLimits `json:"strength_male"`
	StrengthFemale     strengthLimits `json:"strength_female"`
	Intelligence       abilityRange   `json:"intelligence"`
	Wisdom             abilityRange   `json:"wisdom"`
	Dexterity          abilityRange   `json:"dexterity"`
	Constitution       abilityRange   `json:"constitution"`
	Charisma           abilityRange   `json:"charisma"`
	// ClassChoices 是「選單第 n 項 → 職業組合編號」的對照（spec 1093）。
	// 用 []int 而不是 []uint8：encoding/json 會把 []uint8 編成 base64。
	ClassChoices []int         `json:"class_choices"`
	StartingAges []startingAge `json:"starting_ages"`
}

// classCombination 是職業組合編號的語意。這張表**不是資料段的表**，
// 來源是 spec 1093 §二逐條讀出的建角 switch；放進同一份 JSON 是為了讓
// remake 只讀資料、不在 Go 裡重寫規則。
type classCombination struct {
	Combo int `json:"combo"`
	// ClassSlots 是原作職業槽索引（0 牧師 1 德魯伊 2 戰士 3 聖騎士
	// 4 遊俠 5 法師 6 盜賊 7 武僧）。
	ClassSlots []int `json:"class_slots"`
	// StartingExperience 是 25,000 平分後的值（spec 1093）。
	StartingExperience int `json:"starting_experience"`
	// Alignments 是這個組合能選的陣營編號，取自 DS:41D8h（spec 1102 §二）。
	Alignments []int `json:"alignments"`
	// Note 記錄原作的例外。
	Note string `json:"note,omitempty"`
}

// classSlotRules 是每個職業槽的生命骰與起始金錢參數。
type classSlotRules struct {
	Slot int    `json:"slot"`
	Name string `json:"name"`
	// HitDiceCount／HitDiceSize：一級擲幾顆幾面（spec 850）。
	HitDiceCount uint8 `json:"hit_dice_count"`
	HitDiceSize  uint8 `json:"hit_dice_size"`
	// HitDiceLevelCap 是 DS:3EB9h：等級 >= 這個值就不再擲骰、也不再給體質加值。
	HitDiceLevelCap uint8 `json:"hit_dice_level_cap"`
	// StartingMoneyCount／StartingMoneySize 是 AD&D 起始金錢的骰子參數。
	// ⚠ 原作擲了但沒有使用（spec 1101 §四），保留是為了亂數序列一致。
	StartingMoneyCount uint8 `json:"starting_money_count"`
	StartingMoneySize  uint8 `json:"starting_money_size"`
}

// referenceClassCombinations 依 spec 1093 §二的表。
var referenceClassCombinations = []classCombination{
	{Combo: 0, ClassSlots: []int{0}, StartingExperience: 25000},
	{Combo: 1, ClassSlots: []int{1}, StartingExperience: 25000},
	{Combo: 2, ClassSlots: []int{2}, StartingExperience: 25000},
	{Combo: 3, ClassSlots: []int{3}, StartingExperience: 25000,
		Note: "paladin: also sets +191h := 1 and adds effect 08h"},
	{Combo: 4, ClassSlots: []int{4}, StartingExperience: 25000,
		Note: "ranger: also adds effect 86h"},
	{Combo: 5, ClassSlots: []int{5}, StartingExperience: 25000},
	{Combo: 6, ClassSlots: []int{6}, StartingExperience: 25000},
	{Combo: 7, ClassSlots: []int{7}, StartingExperience: 25000},
	{Combo: 8, ClassSlots: []int{0, 2}, StartingExperience: 12500},
	{Combo: 9, ClassSlots: []int{0, 2, 5}, StartingExperience: 8333},
	{Combo: 10, ClassSlots: []int{0, 4}, StartingExperience: 12500,
		Note: "cleric/ranger: also adds effect 86h"},
	{Combo: 11, ClassSlots: []int{0, 5}, StartingExperience: 12500},
	{Combo: 12, ClassSlots: []int{0, 6}, StartingExperience: 12500},
	{Combo: 13, ClassSlots: []int{2, 5}, StartingExperience: 12500},
	{Combo: 14, ClassSlots: []int{2, 6}, StartingExperience: 12500},
	{Combo: 15, ClassSlots: []int{2, 5, 6}, StartingExperience: 8333},
	{Combo: 16, ClassSlots: []int{5, 6}, StartingExperience: 8333,
		Note: "two classes but receives the three-class 8,333; reference behaviour, see spec 1093"},
}

type characterRules struct {
	Source            string             `json:"source"`
	Spec              string             `json:"spec"`
	Races             []raceRules        `json:"races"`
	ClassRequirements [][]int            `json:"class_requirements"`
	ClassCombinations []classCombination `json:"class_combinations"`
	ClassSlots        []classSlotRules   `json:"class_slots"`
	// ConstitutionHPBonus 的索引 0 對應體質 constitutionMin。
	ConstitutionHPBonusFrom int   `json:"constitution_hp_bonus_from"`
	ConstitutionHPBonus     []int `json:"constitution_hp_bonus"`
	// AbilityRoll 是建角擲點的參數（spec 1103 §二）。
	AbilityRoll abilityRollRules `json:"ability_roll"`
	// 戰士系的額外體質加值（spec 869）。判斷用的是**職業組合編號**
	// `+75h`，不是職業槽 ⇒ 只有單職戰士／聖騎士／遊俠拿得到，多職拿不到。
	FighterConstitutionCombos    []int `json:"fighter_constitution_combos"`
	FighterConstitutionBonusFrom int   `json:"fighter_constitution_bonus_from"`
	FighterConstitutionBonus     []int `json:"fighter_constitution_bonus"`
}

// abilityRollRules 是 `ROLLDICE(3, 6) + 1`，每個屬性擲 Attempts 次取最大。
// ⚠ Bonus 與 Attempts 都不是 AD&D 桌上規則，是這個引擎自己加的（spec 1103）。
type abilityRollRules struct {
	DiceCount int `json:"dice_count"`
	DiceSize  int `json:"dice_size"`
	Bonus     int `json:"bonus"`
	Attempts  int `json:"attempts"`
}

// referenceFighterConstitutionBonus 是 spec 869 的 case 分支，索引 0 對應體質 17。
// 這不是資料段的表——原作把它寫死在程式碼裡，抄進 JSON 是為了讓 remake
// 只讀資料、不在 Go 裡重寫規則。
var (
	referenceFighterCombos    = []int{2, 3, 4}
	referenceFighterBonusFrom = 17
	referenceFighterBonus     = []int{1, 2, 3, 3, 4, 4, 4, 5, 5}
)

func main() {
	dsegPath := flag.String("dseg", "workplace/re-sweep/dos/dseg/dos-dseg-dseg.bin",
		tooltext.Text("h.721c8d5ef615"))
	output := flag.String("output", "", tooltext.Text("h.e114c26aade1"))
	flag.Parse()

	blob, err := os.ReadFile(*dsegPath)
	if err != nil {
		log.Fatal(err)
	}
	required := classRequireBase + classCombinations*classRequireSize
	if len(blob) < required {
		log.Fatalf("dseg dump is %d bytes; need at least %d", len(blob), required)
	}

	rules := characterRules{
		Source: tooltext.Text("h.1040debcb867"),
		Spec:   "docs/spec/1099-character-creation-data-tables.md, 1101, 1102",
	}
	for race := firstRace; race <= lastRace; race++ {
		limits := blob[abilityLimitsBase+race*abilityLimitsSize:]
		choices := blob[classChoicesBase+race*classChoicesSize:][:classChoicesSize]
		entry := raceRules{
			RaceID:     race,
			Name:       raceNames[race],
			Selectable: selectableRaces[race],
			Size:       raceSizes[race],
			// 每組的第二個位元組是女性欄；spec 1086 的索引算式是
			// 「基底 + 種族×10h + 角色^[119h]（性別）」。
			StrengthMale:   strengthLimits{Min: limits[0], Max: limits[2], PercentileMax: limits[4]},
			StrengthFemale: strengthLimits{Min: limits[1], Max: limits[3], PercentileMax: limits[5]},
			Intelligence:   abilityRange{Min: limits[6], Max: limits[7]},
			Wisdom:         abilityRange{Min: limits[8], Max: limits[9]},
			Dexterity:      abilityRange{Min: limits[10], Max: limits[11]},
			Constitution:   abilityRange{Min: limits[12], Max: limits[13]},
			Charisma:       abilityRange{Min: limits[14], Max: limits[15]},
		}
		// 每列的第一個位元組是「這個種族有幾個可選職業」，其後才是職業組合
		// 編號。spec 1093 的算式 byte[3FF8h + 種族×0Eh + 選單索引] 因此是
		// 從索引 1 開始取。驗證：人類 6 個 ＝ 牧師／戰士／法師／盜賊／聖騎士／
		// 遊俠，精靈 7 個、矮人 3 個（戰士／盜賊／戰士-盜賊），全部符合 AD&D 1e。
		count := int(choices[0])
		if count > classChoicesSize-1 {
			log.Fatalf("race %d declares %d class choices, exceeding the row capacity", race, count)
		}
		for _, choice := range choices[1 : 1+count] {
			entry.ClassChoices = append(entry.ClassChoices, int(choice))
		}
		adjustments := raceAbilityAdjustments[race]
		entry.AbilityAdjustments = append(entry.AbilityAdjustments, adjustments[:]...)
		for column := 0; column < ageColumns; column++ {
			row := blob[startingAgeBase+race*startingAgeSize+column*4:]
			count, size := row[2], row[3]
			if count == 0 || size == 0 {
				// 該種族不能走這個職業槽——原作留 0，不是年齡 0。
				continue
			}
			entry.StartingAges = append(entry.StartingAges, startingAge{
				ClassSlot: classSlotNames[column],
				BaseAge:   binary.LittleEndian.Uint16(row[:2]),
				DiceCount: count,
				DiceSize:  size,
			})
		}
		rules.Races = append(rules.Races, entry)
	}
	for combo := 0; combo < classCombinations; combo++ {
		row := blob[classRequireBase+combo*classRequireSize:][:classRequireSize]
		requirement := make([]int, 0, classRequireSize)
		for _, value := range row {
			requirement = append(requirement, int(value))
		}
		rules.ClassRequirements = append(rules.ClassRequirements, requirement)
	}

	rules.ClassCombinations = referenceClassCombinations
	for i := range rules.ClassCombinations {
		row := blob[alignmentBase+i*alignmentSize:][:alignmentSize]
		count := int(row[0])
		if count > alignmentSize-1 {
			log.Fatalf("class combination %d declares %d alignments, exceeding the row capacity", i, count)
		}
		for _, value := range row[1 : 1+count] {
			rules.ClassCombinations[i].Alignments =
				append(rules.ClassCombinations[i].Alignments, int(value))
		}
	}

	for slot, name := range allClassSlotNames {
		rules.ClassSlots = append(rules.ClassSlots, classSlotRules{
			Slot:               slot,
			Name:               name,
			HitDiceCount:       blob[hitDiceCountBase+slot],
			HitDiceSize:        blob[hitDiceSizeBase+slot],
			HitDiceLevelCap:    blob[hitDiceCapBase+slot],
			StartingMoneyCount: blob[startingMoneyBase+slot*2],
			StartingMoneySize:  blob[startingMoneyBase+slot*2+1],
		})
	}

	rules.AbilityRoll = abilityRollRules{DiceCount: 3, DiceSize: 6, Bonus: 1, Attempts: 6}
	rules.FighterConstitutionCombos = referenceFighterCombos
	rules.FighterConstitutionBonusFrom = referenceFighterBonusFrom
	rules.FighterConstitutionBonus = referenceFighterBonus

	rules.ConstitutionHPBonusFrom = constitutionMin
	for value := constitutionMin; value <= constitutionMax; value++ {
		// 表存的是有號位元組（0FEh ＝ −2、0FFh ＝ −1）。
		rules.ConstitutionHPBonus =
			append(rules.ConstitutionHPBonus, int(int8(blob[constitutionBonusBase+value])))
	}

	data, err := json.MarshalIndent(rules, "", "  ")
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
	fmt.Printf("wrote %s: %d races, %d class combinations\n",
		*output, len(rules.Races), len(rules.ClassRequirements))
}
