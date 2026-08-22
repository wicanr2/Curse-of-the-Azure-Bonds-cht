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
// ⚠ 演不出來有四種成因：地圖上沒有那個地形碼、要搜尋才演、`RANDOM` 擋著
// （多換幾顆種子）、前置劇情旗標沒有。**第四種現在也分得出來**：守衛裡的
// `COMPARE <格子> <值>` 拆得出來，把那幾格設成比對的值再站一次——演得出來就是
// 「需要前置狀態」，還是演不出來才是真的沒接（spec 1177）。
//
// 用法：
//
//	go run ./cmd/cell-sweep
//	go run ./cmd/cell-sweep -out docs/audit/cell-sweep.md -seeds 8
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
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
	// precondition 非空代表這一格是**把守衛的格子設好之後**才演出來的，
	// 內容就是那組前置狀態。
	precondition string
	// revisit 是**同一次進段裡再跑一次生命週期**演出來的字，用來分出
	// 「只演一次」與「每次都演」。空字串代表第二次什麼都沒演。
	revisit string
	// revisitKind 是重訪的判定：`同` ／ `只演一次` ／ `不同`。
	revisitKind string
	// outcomes 是**同一格、同一組條件、換不同亂數種子**演出來的相異敘述數。
	// 大於 1 代表這一格有靠骰子分歧的分支（豁免、技能檢定、遭遇表）。
	outcomes int
	// otherOutcome 是第二種結果的第一句，讓分歧看得到而不只是個數字。
	otherOutcome string
}

// played 為真代表這一格真的演出了字。
func (c cellResult) played() bool { return c.language == "中文" || c.language == "原文" }

type blockSweep struct {
	id       string
	eclBlock uint8
	geoSet   uint8
	geoBlock uint8
	mask     int
	note     string
	cells    []cellResult
}

// corpus 是原版資料（共用的載入流程在 internal/gamecorpus）加上這一支自己的
// 掃描參數。
type corpus struct {
	gamecorpus.Corpus
	seeds int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	out := flag.String("out", "docs/audit/cell-sweep.md", "輸出的 markdown")
	// ★ 機器可讀的分母。 走訪可達性要拿「實測試過哪些索引」當分母，
	// 兩邊各自重算一定會漂——第一版就是這樣得到 299 而實測是 250，
	// **兩個數都自洽，放在一起就是假的覆蓋率**。所以分母由這一支輸出，
	// `cmd/cell-reachability` 只消費、不重算。
	indexJSON := flag.String("index-json", "", "把「逐格試過哪些索引」寫成 JSON（可達性盤點用）")
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
	if *indexJSON != "" {
		if err := writeIndexJSON(*indexJSON, sweeps); err != nil {
			log.Fatal(err)
		}
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
	data, err := gamecorpus.Load(imagePath, localePath)
	if err != nil {
		return corpus{}, err
	}
	return corpus{Corpus: data, seeds: seeds}, nil
}

func sweepBlock(data corpus, seg segment.Segment) blockSweep {
	sweep := blockSweep{id: seg.ID, eclBlock: seg.Block}
	payload, ok := data.Blocks[seg.Block]
	if !ok {
		sweep.note = "image 裡沒有這個 block"
		return sweep
	}
	dispatch := eclcells.Analyze(payload)
	if !dispatch.Found {
		sweep.note = "沒有以地形碼分派的每格事件"
		if dispatch.TableForm {
			sweep.note += "；改用 `GETTABLE` ＋ `ON GOTO` 查表分派——那一種解得出來，" +
				"但不在本工具的範圍內，見 `cmd/ecl-cell-events`"
		}
		return sweep
	}
	sweep.geoSet, sweep.geoBlock, sweep.mask = seg.Member, dispatch.GeoBlock, dispatch.Mask

	firstCell := map[int][2]int{}
	roofs := map[int]uint8{}
	catalog, hasCatalog := data.Geo[sweep.geoSet]
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
		result := standOnCell(data, seg, index, cell[0], cell[1], roofs[index], nil)
		if !result.played() {
			// 沒演出來就把處理常式開頭的判斷帶出來——「為什麼沒反應」的答案
			// 寫在那幾條指令裡，一格一格去反組譯太貴。
			result.guard = dispatch.Guards[index]
			// 再試一次：把守衛比對的格子設成它要的值。演得出來就代表這一格
			// 有接、只是缺前置狀態；那是**盤點的限制**不是 remake 的缺口。
			result = retryWithGuardCells(data, seg, index, cell[0], cell[1],
				roofs[index], dispatch.GuardCells[index], result)
		}
		sweep.cells = append(sweep.cells, result)
	}
	return sweep
}

// standOnCell 把隊伍放到一格上跑生命週期。演不出來就換條件再試：先加搜尋，
// 再換亂數種子——`RANDOM` 擋著的場景（墓園的盜墓者是 100 抽 32）一次不一定演。
func standOnCell(data corpus, seg segment.Segment, index, x, y int, roof uint8,
	preset map[uint16]uint16) cellResult {
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
				for address, value := range preset {
					state.SetECLMemoryValue(address, value)
				}
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
				result.revisit, result.revisitKind = revisitOnce(state, text)
				// ⚠ 傳的是**這一次的 `Message`**，不是 `text`。`text` 走
				// `playerText`（Message 空就退回 Prompt），把它丟進一組
				// `Message` 集合裡會**讓每一格都多算一種**——集合裡混了兩種
				// 語意的字串。第一版就是這樣把 2 種印成 3 種的。
				result.outcomes, result.otherOutcome = seedOutcomes(
					data, seg, x, y, roof, facing, search, preset, strings.TrimSpace(state.Message))
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

// maxGuardPresetCells 是一次最多幫幾個守衛格子擺前置狀態。守衛裡比對的格子
// 通常一到兩個；放寬只會讓組合數爆掉，而且「要擺五個旗標才演得出來」本身就
// 不是一個有用的結論。
const maxGuardPresetCells = 2

// retryWithGuardCells 把守衛比對的格子設成它要的值再站一次。
//
// ★ 為什麼要這一步。 先前「沒演出來」是一個混在一起的桶子：可能是 remake 沒接，
// 也可能只是這一格要前置劇情。兩者的處置完全不同，而分辨的成本是**一格一格去
// 反組譯守衛**。守衛裡的 `COMPARE` 已經拆得出來，直接擺上去再跑一次就分得開。
//
// ⚠ 這是**滿足守衛**不是重現劇情：擺出來的狀態不保證是正常玩下來會有的。
// 所以演得出來只結論到「有接、缺前置」，報表也照實寫出擺了什麼。
func retryWithGuardCells(data corpus, seg segment.Segment, index, x, y int, roof uint8,
	cells []eclcells.GuardCompare, silent cellResult) cellResult {
	if len(cells) == 0 {
		return silent
	}
	unique := make([]eclcells.GuardCompare, 0, maxGuardPresetCells)
	seen := map[uint16]bool{}
	for _, cell := range cells {
		// 只避「條件成立就 `EXIT`」那幾條——別的 `COMPARE` 動了只會擾亂分派。
		if !cell.ExitsOnMatch {
			continue
		}
		if seen[cell.Address] || len(unique) >= maxGuardPresetCells {
			continue
		}
		seen[cell.Address] = true
		unique = append(unique, cell)
	}
	preset := make(map[uint16]uint16, len(unique))
	parts := make([]string, 0, len(unique))
	for _, cell := range unique {
		// ⚠ 不是設成比對的值就好：`COMPARE 4C01 01 / IF >= / EXIT` 要的是
		// **小於 1**，設成 1 反而保證離開（spec 1177）。
		value, ok := cell.AvoidValue()
		if !ok {
			continue
		}
		preset[cell.Address] = value
		parts = append(parts, fmt.Sprintf("%04X=%02X", cell.Address, value))
	}
	if len(preset) == 0 {
		return silent
	}
	result := standOnCell(data, seg, index, x, y, roof, preset)
	if !result.played() {
		return silent
	}
	result.guard = silent.guard
	result.precondition = strings.Join(parts, " ")
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
	state, err := data.NewParty()
	if err != nil {
		return nil, err
	}
	if err := state.EnterSegment(seg); err != nil {
		return nil, fmt.Errorf("進 %s：%w", seg.ID, err)
	}
	// ⚠ 盤點用的隊伍一律撐起來：有的段一進去就開打（古熔岩洞的伏擊），臨時建的
	// 一名角色會死在入口，後面一格都盤點不到。**只給盤點用**。
	if err := gamecorpus.BoostParty(&state); err != nil {
		return nil, err
	}
	trail, err := settleToDungeon(&state, 16)
	if err != nil {
		return nil, fmt.Errorf("%s 的入口：%w", seg.ID, err)
	}
	if state.Mode != game.ModeDungeon {
		return nil, fmt.Errorf("進 %s 之後停在%s，不是地城；推進軌跡：%s",
			seg.ID, modeName(state.Mode), strings.Join(trail, " → "))
	}
	return &state, nil
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
			if cell.precondition != "" {
				counts["要前置"]++
			}
			// 重訪只在「有演出來」的格子上才有意義。
			if cell.played() && cell.outcomes > 1 {
				counts["換種子演出不同"]++
			}
			if cell.played() {
				switch cell.revisitKind {
				case "只演一次":
					counts["重訪：只演一次"]++
				case "同":
					counts["重訪：每次都演"]++
				case "不同":
					counts["重訪：演出別的"]++
				case "推不回地城":
					counts["重訪：推不回地城"]++
				default:
					counts["重訪：跑不動"]++
				}
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
		"⚠ 「前置」欄是**第二次站上去**的結果：守衛裡 `COMPARE <格子> <值> / IF <op> /\n" +
		"EXIT` 那幾條拆得出來，把格子設成避開 `EXIT` 的值再站一次。演得出來就代表\n" +
		"這一格**有接、只是缺前置劇情**（spec 1177）。⚠ 那是**滿足守衛**不是重現劇情，\n" +
		"擺出來的狀態不保證正常玩下來會有。\n" +
		"⚠ 剩下的「沒演出來」四種形狀，都是**盤點的限制**不是 remake 的缺口：\n" +
		"（1）守衛比的是移動前快照（`4BF0`／`4BF1`）——這支是把隊伍放上去不是走過去；\n" +
		"（2）擺好的旗標被該段自己的前導覆蓋（`4C01` 這一類）；\n" +
		"（3）擷取到的守衛跨過了真正的處理常式（開頭就有 `EXIT` 的那幾格）；\n" +
		"（4）處理常式本來就不講話（`PICTURE FF` 只是把圖關掉）。\n" +
		"⚠ 索引 0 是「沒有事件的地面」，不列。\n\n")
	for _, sweep := range sweeps {
		out.WriteString(fmt.Sprintf("## `%s`\n\n", sweep.id))
		if sweep.note != "" {
			out.WriteString(sweep.note + "\n\n")
			continue
		}
		out.WriteString(fmt.Sprintf("地圖：`GEO%d/0x%02X`；索引 ＝ 地形碼 `& 0x%02X`\n\n",
			sweep.geoSet, sweep.geoBlock, sweep.mask))
		out.WriteString("| 索引 | 格子 | 地形碼 | 條件 | 語言 | 重訪 | 演出來的第一句／沒演的守衛 |\n")
		out.WriteString("|---:|---|---|---|---|---|---|\n")
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
			if cell.precondition != "" {
				extra = append(extra, "前置 `"+cell.precondition+"`")
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
				if cell.precondition != "" {
					text += "（守衛：`" + cell.guard + "`）"
				}
			}
			if !cell.played() {
				condition = "—"
			}
			revisit := cell.revisitKind
			if !cell.played() || revisit == "" {
				revisit = "—"
			}
			if revisit == "不同" {
				revisit = "**不同**：" + firstLine(cell.revisit)
			}
			if cell.played() && cell.outcomes > 1 {
				// ⚠ 兩種結果的**第一句可能一樣**（差在後面幾句：擲到的怪、
				// 金額、成敗）。照樣印第一句會看起來像「兩種一模一樣」，
				// 讓人以為判定壞掉。
				other := firstLine(cell.otherOutcome)
				if other == firstLine(cell.text) {
					other = "（差異在第一句之後）"
				}
				revisit += fmt.Sprintf("<br>骰子分歧 %d 種：%s", cell.outcomes, other)
			}
			out.WriteString(fmt.Sprintf("| %d | `%s` | `%02X` | %s | %s | %s | %s |\n",
				cell.index, cell.cell, cell.roof, condition, cell.language, revisit, text))
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
	out.WriteString(fmt.Sprintf("| 其中要先擺好守衛的旗標才演 | %d |\n", counts["要前置"]))
	out.WriteString(fmt.Sprintf("| 分派表有、地圖上沒有那個地形碼 | %d |\n",
		counts["沒有地形碼"]))
	out.WriteString("\n### 重訪：同一格再跑一次生命週期\n\n")
	out.WriteString("★ 這一支本來每一格都**重新進段**（once-only 旗標會互相污染）。" +
		"那個設計是對的，副作用是**第二次踏上同一格從來沒被觀察過**——" +
		"而「全城市／全房間走訪」缺的分母之一正是重訪分支：原作大量用 " +
		"`SAVE <旗標>` ＋ `IF <旗標>` 讓事件只演一次。\n\n")
	out.WriteString("⚠ 「只演一次」**不是缺陷**：原作本來就有大量一次性事件。" +
		"這裡給的是分母，不是待辦清單。\n\n")
	out.WriteString("⚠ 量的是「同一格再跑一次生命週期」，**不是「走開再走回來」**。" +
		"once-only 旗標的機制與移動無關，所以對那一類是準的；" +
		"靠移動事件觸發的處理常式會被低估。**這是重訪的代理指標，不是重訪本身。**\n\n")
	out.WriteString("⚠ 第二次**只看 `Message`**（新的敘述），不退回 `Prompt`。" +
		"第一版退回 Prompt，於是 74 格被判成「演出別的字」——拆開來 53 格是" +
		"「請按任意鍵繼續」、12 格是撿寶物的提示、5 格是地城 HUD，**真的有新敘述的只有 4 格**。" +
		"退回 Prompt 會把 UI 文字算成劇情，分母灌水近二十倍。" +
		"代價是把「旁白放在 Prompt 上」的那一類算成沒有新敘述 ⇒ **這一欄是下界**。\n\n")
	out.WriteString("| 第二次的行為 | 數 |\n|---|---:|\n")
	out.WriteString(fmt.Sprintf("| 只演一次（第二次沒有新敘述）| %d |\n", counts["重訪：只演一次"]))
	out.WriteString(fmt.Sprintf("| 每次都演（第二次一樣）| %d |\n", counts["重訪：每次都演"]))
	out.WriteString(fmt.Sprintf("| 第二次有**不同的敘述** | %d |\n", counts["重訪：演出別的"]))
	out.WriteString(fmt.Sprintf("| 第一次之後推不回地城（量不到）| %d |\n",
		counts["重訪：推不回地城"]))
	out.WriteString(fmt.Sprintf("| 第二次跑不動 | %d |\n", counts["重訪：跑不動"]))
	out.WriteString("\n### 失敗分支：換亂數種子會不會演出不同的敘述\n\n")
	out.WriteString("★ 走訪缺的另一個分母。主掃描**找到第一顆演得出來的種子就回來**，" +
		"所以豁免沒過／技能檢定失敗／遭遇表擲到別側那一半**結構上不會被看到**" +
		"——報表是滿的，而那一半從來沒跑過。這裡把同一格、同一組條件換每一顆種子" +
		"各站一次，數相異的敘述。\n\n")
	out.WriteString("⚠ 這是「**換種子會不會演出不同的敘述**」，不是「有幾條失敗分支」。" +
		"一格可能有三條分支而種子只打到兩條；也可能敘述相同而只有數值不同（傷害、金額）。" +
		"⇒ **下界**，而且是「有沒有骰子分歧」的證據，不是分支數。\n\n")
	out.WriteString(fmt.Sprintf("| 有內容的格子 | %d |\n|---|---:|\n",
		counts["中文"]+counts["原文"]))
	out.WriteString(fmt.Sprintf("| 其中換種子會演出**不同敘述**的 | %d |\n",
		counts["換種子演出不同"]))
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

// revisitOnce 在**同一次進段**裡再跑一次地城生命週期，回傳第二次演出來的字
// 與判定。
//
// ★ 存在的理由：這一支本來每一格都重新進段（註解寫著「once-only 旗標會互相
// 污染」）——那個設計正確，但副作用是**第二次踏上同一格的行為從來沒被觀察過**。
// 「全城市／全房間走訪」缺的分母之一正是重訪分支：原作大量使用
// `SAVE <旗標>` ＋ `IF <旗標>` 讓事件只演一次，而 remake 接不接得住這一半，
// 在只走一次的盤點裡是看不見的。
//
// ⚠ 這量的是「**同一格再跑一次生命週期**」，不是「走開再走回來」。once-only
// 旗標的機制與移動無關，所以對那一類是準的；但如果某一格的處理常式是靠
// 移動事件觸發，這個做法會低估。**它是重訪的代理指標，不是重訪本身。**
//
// ⚠ 第二次不再演出字**不代表 remake 漏接**——原作本來就有大量只演一次的事件。
// 這一欄提供的是分母（有多少格有重訪分支），不是缺陷清單。
func revisitOnce(state *game.State, first string) (string, string) {
	// ⚠ 第一次跑完畫面停在事件上，**要先推回地城**再跑第二次。少了這一步
	// `RunDungeonLifecycle` 會正確地拒絕，而那個拒絕看起來像「remake 跑不動」。
	if _, err := settleToDungeon(state, 24); err != nil {
		return "", "推不回地城"
	}
	if state.Mode != game.ModeDungeon {
		return "", "推不回地城"
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		return "", "跑不動"
	}
	// ⚠ 第二次**只看 `Message`**，不用 `playerText`。
	//
	// 第一版用了 `playerText`（Message 沒有就退回 Prompt），結果 74 格被判成
	// 「第二次演出別的字」——拆開來看，53 格是「請按任意鍵或 Enter 繼續」、
	// 12 格是撿寶物的選擇提示、5 格是地城 HUD 那一行，**真正有新敘述的只有 4 格**。
	// 退回 Prompt 會把 UI 文字算成劇情內容，把分母灌水近二十倍。
	//
	// ⚠ 代價是**低估**：原作有些遭遇把旁白放在 Prompt 上（`playerText` 的註解
	// 就是為此而寫），那一類重訪會被算成「沒有新敘述」。方向是保守的——
	// 這一欄寧可少報，不要報出一份假的待辦清單。
	second := strings.TrimSpace(state.Message)
	switch {
	case second == "":
		return "", "只演一次"
	case second == first:
		return second, "同"
	default:
		return second, "不同"
	}
}

// settleToDungeon 把事件／選單／戰鬥推完，直到停在地城模式。回傳推進軌跡，
// 停不下來時連軌跡一起報出去——**「推不動」要看得到卡在哪一格**。
//
// ★ 抽出來是因為重訪也要用：第一次生命週期跑完之後畫面停在事件上，
// 不先推回地城就直接再跑一次，`RunDungeonLifecycle` 會正確地拒絕，
// 而那個拒絕會被記成「跑不動」——**看起來像 remake 有問題，其實是量法錯了**。
func settleToDungeon(state *game.State, limit int) ([]string, error) {
	trail := make([]string, 0, limit)
	for step := 0; step < limit && state.Mode != game.ModeDungeon; step++ {
		trail = append(trail, fmt.Sprintf("%s／%d 選項／%s",
			modeName(state.Mode), len(state.Choices), firstLine(playerText(state))))
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return trail, fmt.Errorf("戰鬥：%w", err)
				}
			}
			continue
		}
		// ⚠ 地點模式（商店、神殿、旅店）的第一個選項是「買」，選下去會在商店選單
		// 裡繞不出來——散提爾堡的魔法商店就是這樣把整段擋住的。那裡要選最後一項
		// （離開）。事件模式的第一項才是「繼續」。
		choice := 0
		if state.Mode == game.ModePlace && len(state.Choices) > 0 {
			choice = len(state.Choices) - 1
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(choice); selectErr != nil {
				return trail, fmt.Errorf("推不動：continue=%v select=%v", err, selectErr)
			}
		}
	}
	return trail, nil
}

// seedOutcomes 把**同一格、同一組條件**換不同亂數種子各站一次，數出相異的敘述。
//
// ★ 存在的理由：「全城市／全房間走訪」缺的另一個分母是**失敗分支**——豁免沒過、
// 技能檢定失敗、遭遇表擲到別的那一側。主掃描找到第一顆演得出來的種子就回來了，
// 所以另一側**結構上不會被看到**：報表滿的，而那一半從來沒跑過。
//
// ⚠ 這是「**換種子會不會演出不同的敘述**」，不是「有幾條失敗分支」。一格可能
// 有三條分支而八顆種子只打到兩條；也可能敘述相同而數值不同（傷害、金額）。
// ⇒ 這一欄是**下界**，而且是「有沒有骰子分歧」的證據，不是分支數。
//
// ⚠ 同樣只看 `Message`：退回 `Prompt` 會把 UI 文字算成劇情（重訪那一欄踩過，
// 分母灌水近二十倍）。
func seedOutcomes(data corpus, seg segment.Segment, x, y int, roof, facing uint8,
	search bool, preset map[uint16]uint16, first string) (int, string) {
	seen := map[string]bool{}
	if first != "" {
		seen[first] = true
	}
	other := ""
	for seed := 1; seed <= data.seeds; seed++ {
		state, err := enterDungeon(data, seg)
		if err != nil {
			continue
		}
		state.SetECLSeed(int64(seed))
		for address, value := range preset {
			state.SetECLMemoryValue(address, value)
		}
		state.SetDungeonGeometryView(x, y, facing)
		state.DungeonWallRoof = roof
		state.DungeonSearchEnabled = search
		if err := state.RunDungeonLifecycle(); err != nil {
			continue
		}
		message := strings.TrimSpace(state.Message)
		if message == "" || seen[message] {
			continue
		}
		seen[message] = true
		if other == "" {
			other = message
		}
	}
	return len(seen), other
}

// writeIndexJSON 輸出「這一次逐格實測**真的試過**哪些 (段, ECL block, 分派索引)」。
//
// ⚠ 只收真的試過的：進不去的段、地圖上沒有那個地形碼的索引都不算。
// 可達性的分母必須與這一份逐字相同，否則覆蓋率是拿兩把不同的尺在比。
func writeIndexJSON(path string, sweeps []blockSweep) error {
	// ⚠ `Mask` 一定要一起輸出：可達性那一側拿到的是**地形碼**，
	// 要用 `地形碼 and Mask` 才對得回索引。少了它下游只能自己重算 mask，
	// 而「自己重算」正是分母會漂掉的原因。
	type record struct {
		Segment string `json:"segment"`
		Block   uint8  `json:"block"`
		Mask    int    `json:"mask"`
		Index   int    `json:"index"`
		Played  bool   `json:"played"`
	}
	records := make([]record, 0, 256)
	for _, sweep := range sweeps {
		if sweep.note != "" {
			continue
		}
		for _, cell := range sweep.cells {
			if cell.note == "地圖上沒有這個地形碼" {
				continue
			}
			records = append(records, record{
				Segment: sweep.id, Block: sweep.eclBlock, Mask: sweep.mask,
				Index: cell.index, Played: cell.played(),
			})
		}
	}
	payload, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}
