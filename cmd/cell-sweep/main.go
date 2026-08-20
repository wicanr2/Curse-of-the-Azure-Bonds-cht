// cell-sweep 把「哪一格演哪一場」的對照表**實際站上去跑一遍**：對每個有地形
// 分派的 block、每一個分派索引，直接進段、把隊伍放到那個地形碼的格子上、跑一次
// 地城生命週期，記下玩家真的看到什麼字。
//
// ★ 存在的理由：`cmd/ecl-cell-events` 只讀 bytecode，回答「站上那一格會跳到哪支
// 處理常式」；它答不了「跳過去之後演不演得出來、演出來是不是中文」。那一段落差
// 正是中文化會漏的地方——處理常式接上了、文字沒接，對照表照樣是滿的。
//
// ⚠ 每一格都**重新進段**：once-only 旗標與戰鬥結果會互相污染，接續著跑會讓後面
// 的格子看起來沒內容。
//
// ⚠ 這支量的是**內容與語系**，不是可達性。它把隊伍撐起來（`boostSweepParty`
// 的同一套）好讓入口戰鬥不會擋住盤點，所以「這一格演得出來」不等於「正常隊伍走
// 得到這一格」。走得到走不到由主線分段測試負責。
//
// ⚠ 演不出來有四種成因，這支只分得出前兩種：地圖上沒有那個地形碼、要搜尋才演、
// `RANDOM` 擋著（多換幾顆種子）、前置劇情旗標沒有。最後一種會落在「沒演出來」，
// 要人去讀處理常式的守衛。
//
// 用法：
//
//	go run ./cmd/cell-sweep
//	go run ./cmd/cell-sweep -out docs/audit/cell-sweep.md -seeds 8
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// cellResult 是一個分派索引站上去的結果。
type cellResult struct {
	index    int
	cell     string
	roof     uint8
	text     string
	language string
	search   bool
	facing   uint8
	seed     int64
	note     string
	guard    string
}

// played 為真代表這一格真的演出了字。
func (c cellResult) played() bool { return c.language == "中文" || c.language == "原文" }

type blockSweep struct {
	id       string
	geoSet   uint8
	geoBlock uint8
	mask     int
	note     string
	cells    []cellResult
}

type corpus struct {
	blocks  map[uint8][]byte
	records map[uint8]map[uint8]monster.Record
	items   map[uint16][]monster.ItemRecord
	geo     map[uint8]geo.Catalog
	catalog locale.Catalog
	seeds   int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	out := flag.String("out", "docs/audit/cell-sweep.md", "輸出的 markdown")
	seeds := flag.Int("seeds", 8, "演不出來時要換幾顆 ECL 亂數種子再試")
	only := flag.String("only", "", "只掃這一段（`ECL4/0x25`）；留白就掃全部")
	flag.Parse()

	data, err := loadCorpus(*image, *localePath, *seeds)
	if err != nil {
		log.Fatal(err)
	}
	segments := segment.All()
	if *only != "" {
		picked, ok := segment.Lookup(*only)
		if !ok {
			log.Fatalf("註冊表沒有 %q", *only)
		}
		segments = []segment.Segment{picked}
	}
	sweeps := make([]blockSweep, 0, len(segments))
	for _, seg := range segments {
		sweeps = append(sweeps, sweepBlock(data, seg))
	}
	if err := os.WriteFile(*out, []byte(render(sweeps)), 0o644); err != nil {
		log.Fatal(err)
	}
	counts := summarise(sweeps)
	fmt.Printf("block=%d 逐格試過=%d 中文=%d 落回原文=%d 沒演出來=%d → %s\n",
		counts["block"], counts["試過"], counts["中文"], counts["原文"],
		counts["沒演出來"], *out)
}

func loadCorpus(imagePath, localePath string, seeds int) (corpus, error) {
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		return corpus{}, err
	}
	defer archive.Close()

	data := corpus{
		blocks:  map[uint8][]byte{},
		records: map[uint8]map[uint8]monster.Record{},
		geo:     map[uint8]geo.Catalog{},
		seeds:   seeds,
	}
	// 寶物的物品表也要載：段的入口有 `TREASURE`（散提爾堡的魔法商店就是），
	// 沒載會在入口就跑不動，那一整段的每一格都變成「跑不動」。
	itemPayloads := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		payload := memberPayload(archive, fmt.Sprintf("ITEM%d.DAX", chapter))
		if payload == nil {
			continue
		}
		itemPayloads[uint8(chapter)] = payload
	}
	items, err := game.ParseTreasureItemBlocks(itemPayloads)
	if err != nil {
		return corpus{}, err
	}
	data.items = items
	for chapter := 1; chapter <= 6; chapter++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", chapter))
		if payload == nil {
			return corpus{}, fmt.Errorf("image 裡沒有 ECL%d.DAX", chapter)
		}
		parsed, err := dax.Parse(payload)
		if err != nil {
			return corpus{}, err
		}
		for _, block := range parsed {
			data.blocks[block.Entry.ID] = block.Data
		}
		// 六章的怪物表都要載：段的入口有可能直接開戰，少載一章那一段就會
		// 安靜地跳過開戰，看起來像「入口沒事」。
		monsters := memberPayload(archive, fmt.Sprintf("MON%dCHA.DAX", chapter))
		if monsters == nil {
			continue
		}
		parsedMonsters, err := dax.Parse(monsters)
		if err != nil {
			return corpus{}, err
		}
		chapterRecords := map[uint8]monster.Record{}
		for _, block := range parsedMonsters {
			record, err := monster.Parse(block.Data)
			if err != nil {
				return corpus{}, fmt.Errorf("MON%dCHA block %#02x: %w", chapter, block.Entry.ID, err)
			}
			chapterRecords[block.Entry.ID] = record
		}
		data.records[uint8(chapter)] = chapterRecords

		payload = memberPayload(archive, fmt.Sprintf("GEO%d.DAX", chapter))
		if payload == nil {
			continue
		}
		catalog := geo.NewCatalog()
		if err := catalog.AddDAX(uint8(chapter), payload); err != nil {
			return corpus{}, fmt.Errorf("GEO%d: %w", chapter, err)
		}
		data.geo[uint8(chapter)] = catalog
	}
	localeData, err := os.ReadFile(localePath)
	if err != nil {
		return corpus{}, err
	}
	data.catalog, err = locale.Load(localeData)
	if err != nil {
		return corpus{}, err
	}
	return data, nil
}

func sweepBlock(data corpus, seg segment.Segment) blockSweep {
	sweep := blockSweep{id: seg.ID}
	payload, ok := data.blocks[seg.Block]
	if !ok {
		sweep.note = "image 裡沒有這個 block"
		return sweep
	}
	dispatch := eclcells.Analyze(payload)
	if !dispatch.Found {
		sweep.note = "沒有以地形碼分派的每格事件"
		if dispatch.TableForm {
			sweep.note += "；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）"
		}
		return sweep
	}
	sweep.geoSet, sweep.geoBlock, sweep.mask = seg.Member, dispatch.GeoBlock, dispatch.Mask

	firstCell := map[int][2]int{}
	roofs := map[int]uint8{}
	catalog, hasCatalog := data.geo[sweep.geoSet]
	if hasCatalog {
		if grid, hasGrid := catalog.Lookup(geo.MapRef{Set: sweep.geoSet, BlockID: sweep.geoBlock}); hasGrid {
			for y := 0; y < geo.Height; y++ {
				for x := 0; x < geo.Width; x++ {
					roof := grid.CellWrapped(x, y).Terrain
					index := int(roof) & dispatch.Mask
					if _, seen := firstCell[index]; !seen {
						firstCell[index], roofs[index] = [2]int{x, y}, roof
					}
				}
			}
		}
	}
	if !hasCatalog {
		sweep.note = fmt.Sprintf("讀不到 GEO%d", sweep.geoSet)
		return sweep
	}
	// 先確認這一段進得去。進不去就整段記一次原因——每一格重複同一句只是雜訊，
	// 而且會讓「沒演出來」的數字被一個進不去的段灌爆。
	if _, err := enterDungeon(data, seg); err != nil {
		sweep.note = "進不去：" + err.Error()
		return sweep
	}
	for _, index := range dispatch.Indexes {
		// 索引 0 是「沒有事件的地面」，全圖大半都是它。
		if index == 0 {
			continue
		}
		cell, ok := firstCell[index]
		if !ok {
			sweep.cells = append(sweep.cells, cellResult{
				index: index, note: "地圖上沒有這個地形碼",
			})
			continue
		}
		result := standOnCell(data, seg, index, cell[0], cell[1], roofs[index])
		if !result.played() {
			// 沒演出來就把處理常式開頭的判斷帶出來——「為什麼沒反應」的答案
			// 寫在那幾條指令裡，一格一格去反組譯太貴。
			result.guard = dispatch.Guards[index]
		}
		sweep.cells = append(sweep.cells, result)
	}
	return sweep
}

// standOnCell 把隊伍放到一格上跑生命週期。演不出來就換條件再試：先加搜尋，
// 再換亂數種子——`RANDOM` 擋著的場景（墓園的盜墓者是 100 抽 32）一次不一定演。
func standOnCell(data corpus, seg segment.Segment, index, x, y int, roof uint8) cellResult {
	result := cellResult{index: index, cell: fmt.Sprintf("(%d,%d)", x, y), roof: roof}
	var lastErr error
	for seed := 1; seed <= data.seeds; seed++ {
		for _, search := range []bool{false, true} {
			// ⚠ 邊界那一圈的處理常式是 `COMPARE C04D <方向> / IF <> / EXIT`：
			// **要面對那個方向才演**。只用固定朝向掃，猶拉什的出城口整批會
			// 看起來沒內容——那是假零。
			//
			// ⚠ 朝向是 0..7，`C04D` ＝ 朝向 ÷ 2。掃 0..3 只蓋得到 `C04D`
			// 的 0 與 1，南、西兩面照樣落空。
			for _, facing := range []uint8{0, 2, 4, 6} {
				state, err := enterDungeon(data, seg)
				if err != nil {
					lastErr = err
					continue
				}
				state.SetECLSeed(int64(seed))
				state.SetDungeonGeometryView(x, y, facing)
				state.DungeonWallRoof = roof
				state.DungeonSearchEnabled = search
				if err := state.RunDungeonLifecycle(); err != nil {
					lastErr = err
					continue
				}
				text := playerText(state)
				if text == "" {
					continue
				}
				result.text, result.language = text, languageOf(text)
				result.search, result.seed, result.facing = search, int64(seed), facing
				return result
			}
		}
	}
	result.language = "—"
	result.note = "沒演出來"
	if lastErr != nil {
		result.note = "跑不動：" + lastErr.Error()
	}
	return result
}

// playerText 取這一格演出來的字。敘述有兩條路：一般事件走 Message，遭遇選單的
// 旁白走 Prompt。只看其中一條會讓另一條整批看起來是空的。
func playerText(state *game.State) string {
	if message := strings.TrimSpace(state.Message); message != "" {
		return message
	}
	return strings.TrimSpace(state.Prompt)
}

// enterDungeon 建一支盤點用隊伍、直接進段，並把入口的事件／選單／戰鬥推完，
// 直到停在地城模式上。
func enterDungeon(data corpus, seg segment.Segment) (*game.State, error) {
	state := game.NewStateFromECLBlocks(data.catalog, data.blocks, 0x50)
	for chapter, chapterRecords := range data.records {
		state.SetMonsterRecordsForECL(chapter, chapterRecords)
	}
	state.SetTreasureItemBlocks(data.items)
	if err := state.OpenCharacterCreation(); err != nil {
		return nil, err
	}
	if err := state.AddCreationCharacter(0); err != nil {
		return nil, err
	}
	if err := state.FinishCharacterCreation(); err != nil {
		return nil, err
	}
	if err := state.EnterSegment(seg); err != nil {
		return nil, fmt.Errorf("進 %s：%w", seg.ID, err)
	}
	if err := boostSweepParty(&state); err != nil {
		return nil, err
	}
	trail := make([]string, 0, 16)
	for step := 0; step < 16 && state.Mode != game.ModeDungeon; step++ {
		trail = append(trail, fmt.Sprintf("%s／%d 選項／%s",
			modeName(state.Mode), len(state.Choices), firstLine(playerText(&state))))
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return nil, fmt.Errorf("%s 的入口戰鬥：%w", seg.ID, err)
				}
			}
			continue
		}
		// ⚠ 地點模式（商店、神殿、旅店）的第一個選項是「買」，選下去會在
		// 商店選單裡繞不出來——散提爾堡的魔法商店就是這樣把整段擋住的。
		// 那裡要選最後一項（離開）。事件模式的第一項才是「繼續」。
		choice := 0
		if state.Mode == game.ModePlace && len(state.Choices) > 0 {
			choice = len(state.Choices) - 1
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(choice); selectErr != nil {
				return nil, fmt.Errorf("%s 的入口推不動：continue=%v select=%v",
					seg.ID, err, selectErr)
			}
		}
	}
	if state.Mode != game.ModeDungeon {
		return nil, fmt.Errorf("進 %s 之後停在%s，不是地城；推進軌跡：%s",
			seg.ID, modeName(state.Mode), strings.Join(trail, " → "))
	}
	return &state, nil
}

// boostSweepParty 把隊伍撐到足以走完內容盤點的程度。**只給盤點用**：段測試不准
// 這樣做，那會把「這一段打得贏嗎」偷偷換成「這一段演得出來嗎」。
func boostSweepParty(state *game.State) error {
	party := state.PartyFighters()
	if len(party) == 0 {
		return fmt.Errorf("盤點用的隊伍是空的")
	}
	for index := range party {
		party[index].HitPoints, party[index].MaxHitPoints = 999, 999
		party[index].ArmorClass = -10
		party[index].AttackBonus = 100
		party[index].DamageDiceCount, party[index].DamageDiceSides = 1, 1
		party[index].DamageBonus = 100
		party[index].AttacksPerTurn = 8
		party[index].InitiativeBonus = 100
	}
	return state.SetParty(party)
}

// languageOf 判定一段玩家看得到的字是中文、原文還是沒有字。原作文字是英文，
// 所以「沒有漢字但有英文字母」就是落回原文。
func languageOf(text string) string {
	hasHan, hasLatin := false, false
	for _, glyph := range text {
		switch {
		case unicode.Is(unicode.Han, glyph):
			hasHan = true
		case glyph >= 'A' && glyph <= 'Z', glyph >= 'a' && glyph <= 'z':
			hasLatin = true
		}
	}
	switch {
	case hasHan:
		return "中文"
	case hasLatin:
		return "原文"
	}
	return "—"
}

func summarise(sweeps []blockSweep) map[string]int {
	counts := map[string]int{}
	for _, sweep := range sweeps {
		if sweep.note != "" {
			continue
		}
		counts["block"]++
		for _, cell := range sweep.cells {
			if cell.note == "地圖上沒有這個地形碼" {
				counts["沒有地形碼"]++
				continue
			}
			counts["試過"]++
			switch {
			case cell.language == "中文":
				counts["中文"]++
			case cell.language == "原文":
				counts["原文"]++
			default:
				counts["沒演出來"]++
			}
			if cell.search {
				counts["要搜尋"]++
			}
			if cell.facing != 0 {
				counts["要轉向"]++
			}
			if cell.seed > 1 {
				counts["要換種子"]++
			}
		}
	}
	return counts
}

func render(sweeps []blockSweep) string {
	var out strings.Builder
	out.WriteString("# 逐格實測（站上去真的演了什麼）\n\n" +
		"由 `cmd/cell-sweep` 產生，不要手改。\n\n" +
		"對每個有地形分派的 block、每一個分派索引，直接進段、把隊伍放到那個地形碼\n" +
		"的第一格上、跑一次地城生命週期，記下玩家真的看到的字。對照表\n" +
		"（`docs/audit/ecl-cell-events.md`）只讀 bytecode，回答「會跳到哪支處理\n" +
		"常式」；這一份回答「跳過去之後演不演得出來、是不是中文」。\n\n" +
		"⚠ 每一格都重新進段，once-only 旗標不互相污染。\n" +
		"⚠ 盤點用的隊伍被撐起來過，好讓入口戰鬥不擋住盤點。**「演得出來」不等於\n" +
		"「正常隊伍走得到」**——可達性由主線分段測試負責。\n" +
		"⚠ 「沒演出來」有兩種成因這支分不出來：`RANDOM` 沒抽中（已經換過種子），\n" +
		"以及處理常式的前置劇情旗標沒有。要人去讀那支處理常式的守衛。\n" +
		"⚠ 索引 0 是「沒有事件的地面」，不列。\n\n")
	for _, sweep := range sweeps {
		out.WriteString(fmt.Sprintf("## `%s`\n\n", sweep.id))
		if sweep.note != "" {
			out.WriteString(sweep.note + "\n\n")
			continue
		}
		out.WriteString(fmt.Sprintf("地圖：`GEO%d/0x%02X`；索引 ＝ 地形碼 `& 0x%02X`\n\n",
			sweep.geoSet, sweep.geoBlock, sweep.mask))
		out.WriteString("| 索引 | 格子 | 地形碼 | 條件 | 語言 | 演出來的第一句／沒演的守衛 |\n")
		out.WriteString("|---:|---|---|---|---|---|\n")
		for _, cell := range sweep.cells {
			if cell.note == "地圖上沒有這個地形碼" {
				out.WriteString(fmt.Sprintf("| %d | — | — | — | — | %s |\n",
					cell.index, cell.note))
				continue
			}
			extra := make([]string, 0, 3)
			if cell.search {
				extra = append(extra, "要搜尋")
			}
			if cell.facing != 0 {
				extra = append(extra, "面向"+facingName(cell.facing))
			}
			if cell.seed > 1 {
				extra = append(extra, fmt.Sprintf("第 %d 顆種子", cell.seed))
			}
			condition := "站上去"
			if len(extra) > 0 {
				condition = strings.Join(extra, "＋")
			}
			text := cell.note
			if !cell.played() && cell.guard != "" {
				text += "（守衛：`" + cell.guard + "`）"
			}
			if cell.played() {
				text = "「" + firstLine(cell.text) + "」"
				if cell.note != "" {
					text += "（" + cell.note + "）"
				}
			}
			if !cell.played() {
				condition = "—"
			}
			out.WriteString(fmt.Sprintf("| %d | `%s` | `%02X` | %s | %s | %s |\n",
				cell.index, cell.cell, cell.roof, condition, cell.language, text))
		}
		out.WriteString("\n")
	}
	counts := summarise(sweeps)
	out.WriteString("## 摘要\n\n| 項目 | 數 |\n|---|---:|\n")
	out.WriteString(fmt.Sprintf("| 有地形分派的 block | %d |\n", counts["block"]))
	out.WriteString(fmt.Sprintf("| 逐格試過的索引 | %d |\n", counts["試過"]))
	out.WriteString(fmt.Sprintf("| 演出來是中文 | %d |\n", counts["中文"]))
	out.WriteString(fmt.Sprintf("| 演出來落回原文 | %d |\n", counts["原文"]))
	out.WriteString(fmt.Sprintf("| 沒演出來 | %d |\n", counts["沒演出來"]))
	out.WriteString(fmt.Sprintf("| 其中要搜尋才演 | %d |\n", counts["要搜尋"]))
	out.WriteString(fmt.Sprintf("| 其中要面對特定方向才演 | %d |\n", counts["要轉向"]))
	out.WriteString(fmt.Sprintf("| 其中要換亂數種子才演 | %d |\n", counts["要換種子"]))
	out.WriteString(fmt.Sprintf("| 分派表有、地圖上沒有那個地形碼 | %d |\n",
		counts["沒有地形碼"]))
	return out.String()
}

// facingName 把朝向碼翻成方位。朝向是 0..7 順時針，原作的 `C04D` ＝ 朝向 ÷ 2。
func facingName(facing uint8) string {
	switch facing / 2 {
	case 0:
		return "北"
	case 1:
		return "東"
	case 2:
		return "南"
	case 3:
		return "西"
	}
	return "?"
}

func modeName(mode game.Mode) string {
	switch mode {
	case game.ModeTitle:
		return "標題"
	case game.ModeWilderness:
		return "世界地圖"
	case game.ModeEvent:
		return "事件"
	case game.ModeMap:
		return "地圖"
	case game.ModePlace:
		return "地點"
	case game.ModeCombat:
		return "戰鬥"
	case game.ModeJournal:
		return "手札"
	case game.ModeCharacterCreation:
		return "建角"
	case game.ModeDungeon:
		return "地城"
	}
	return "?"
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "—"
	}
	runes := []rune(strings.ReplaceAll(text, "\n", " "))
	if len(runes) > 32 {
		return string(runes[:32]) + "…"
	}
	return string(runes)
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		return payload
	}
	return nil
}
