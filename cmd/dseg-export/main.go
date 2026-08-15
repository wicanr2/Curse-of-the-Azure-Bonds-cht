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
	RaceID          int            `json:"race_id"`
	Name            string         `json:"name"`
	StrengthMale    strengthLimits `json:"strength_male"`
	StrengthFemale  strengthLimits `json:"strength_female"`
	Intelligence    abilityRange   `json:"intelligence"`
	Wisdom          abilityRange   `json:"wisdom"`
	Dexterity       abilityRange   `json:"dexterity"`
	Constitution    abilityRange   `json:"constitution"`
	Charisma        abilityRange   `json:"charisma"`
	// ClassChoices 是「選單第 n 項 → 職業組合編號」的對照（spec 1093）。
	// 用 []int 而不是 []uint8：encoding/json 會把 []uint8 編成 base64。
	ClassChoices []int         `json:"class_choices"`
	StartingAges []startingAge `json:"starting_ages"`
}

type characterRules struct {
	Source            string     `json:"source"`
	Spec              string     `json:"spec"`
	Races             []raceRules `json:"races"`
	ClassRequirements [][]int     `json:"class_requirements"`
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
	required := classRequireBase + classCombinations*classRequireSize
	if len(blob) < required {
		log.Fatalf("dseg dump 只有 %d bytes，至少需要 %d", len(blob), required)
	}

	rules := characterRules{
		Source: "DOS START.EXE 常駐資料段",
		Spec:   "docs/spec/1099-character-creation-data-tables.md",
	}
	for race := firstRace; race <= lastRace; race++ {
		limits := blob[abilityLimitsBase+race*abilityLimitsSize:]
		choices := blob[classChoicesBase+race*classChoicesSize:][:classChoicesSize]
		entry := raceRules{
			RaceID: race,
			Name:   raceNames[race],
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
			log.Fatalf("種族 %d 的可選職業數 %d 超出一列容量", race, count)
		}
		for _, choice := range choices[1 : 1+count] {
			entry.ClassChoices = append(entry.ClassChoices, int(choice))
		}
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
	fmt.Printf("已寫入 %s：%d 個種族、%d 個職業組合\n",
		*output, len(rules.Races), len(rules.ClassRequirements))
}
