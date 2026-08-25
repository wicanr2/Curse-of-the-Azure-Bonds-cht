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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"path/filepath"
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
func (c cellResult) played() bool {
	return c.language == tooltext.Text("h.72726d8818f6") || c.language == tooltext.Text("h.354b28c85333")
}

type blockSweep struct {
	id       string
	eclBlock uint8
	geoSet   uint8
	geoBlock uint8
	mask     int
	// note ＝「這一段掃不了」的原因。⚠ 渲染時看到它就整段跳過，
	// **不要拿它放一般註記**——那會讓一段掃得好好的段從報表裡消失，
	// 而數字只是安靜地少掉。
	note string
	// fromSnapshot ＝ 這一段是從主線的段內快照進的（冷進不去）。
	fromSnapshot bool
	cells        []cellResult
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
	out := flag.String("out", "docs/audit/cell-sweep.md", tooltext.Text("h.aff4479ab1b9"))
	// ★ 機器可讀的分母。 走訪可達性要拿「實測試過哪些索引」當分母，
	// 兩邊各自重算一定會漂——第一版就是這樣得到 299 而實測是 250，
	// **兩個數都自洽，放在一起就是假的覆蓋率**。所以分母由這一支輸出，
	// `cmd/cell-reachability` 只消費、不重算。
	indexJSON := flag.String("index-json", "", tooltext.Text("h.5a44f43df721"))
	seeds := flag.Int("seeds", 8, tooltext.Text("h.1686177b156b"))
	only := flag.String("only", "", tooltext.Text("h.3151e083a021"))
	snapshots := flag.String("snapshots", "workplace/campaign-frames/snapshots",
		tooltext.Text("h.72010d5e943b"))
	flag.Parse()
	snapshotDir = *snapshots

	data, err := loadCorpus(*image, *localePath, *seeds)
	if err != nil {
		log.Fatal(err)
	}
	segments := segment.All()
	if *only != "" {
		picked, ok := segment.Lookup(*only)
		if !ok {
			log.Fatal(tooltext.Format("h.6728cf22b7af", *only))
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
	fmt.Print(tooltext.Format("h.1db1f23a7ba1", counts["block"], counts[tooltext.Text("h.9c9a4e85c872")], counts[tooltext.Text("h.72726d8818f6")], counts[tooltext.Text("h.354b28c85333")], counts[tooltext.Text("h.0ff689f870c6")], *out))
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
		sweep.note = tooltext.Text("h.488ecde7b0b1")
		return sweep
	}
	dispatch := eclcells.Analyze(payload)
	if !dispatch.Found {
		sweep.note = tooltext.Text("h.c58d07204221")
		if dispatch.TableForm {
			sweep.note += tooltext.Text("h.2a24418bbe10") +
				tooltext.Text("h.dc050ed3e9c0")
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
		sweep.note = tooltext.Format("h.39cb0957d471", sweep.geoSet)
		return sweep
	}
	// 先確認這一段進得去。進不去就整段記一次原因——每一格重複同一句只是雜訊，
	// 而且會讓「沒演出來」的數字被一個進不去的段灌爆。
	if _, err := enterDungeon(data, seg); err != nil {
		sweep.note = tooltext.Text("h.1e5d9d1f7560") + err.Error()
		return sweep
	}
	// ⚠ **不要寫進 `note`**：`note` 的意思是「這一段掃不了」，渲染時看到它就
	// 整段跳過。把來源註記寫進去會讓一段**掃得好好的**段從報表裡消失，
	// 而數字只是安靜地少掉——分母從 259 掉回 250 而且沒有任何錯誤訊息。
	sweep.fromSnapshot = snapshotEntered[seg.ID]
	for _, index := range dispatch.Indexes {
		// 索引 0 是「沒有事件的地面」，全圖大半都是它。
		if index == 0 {
			continue
		}
		cell, ok := firstCell[index]
		if !ok {
			sweep.cells = append(sweep.cells, cellResult{
				index: index, note: tooltext.Text("h.5e63f0393483"),
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
	result.note = tooltext.Text("h.0ff689f870c6")
	if lastErr != nil {
		result.note = tooltext.Text("h.883dcdb58171") + lastErr.Error()
	}
	return result
}

// silentShapes 是 `silentShape` 回得出來的四種形狀，照報表的順序。
var silentShapes = []string{
	tooltext.Text("h.dda6ee401ff2"),
	tooltext.Text("h.fa920a07b092"),
	tooltext.Text("h.59a613ab8858"),
	tooltext.Text("h.cace6089a166"),
}

// silentShape 把「沒演出來」歸到 spec 的四種形狀之一。
//
// ★ 為什麼要歸類。 只留一個「沒演出來 10」的話，下一輪會把它當待辦重查一次
// ——而這四種形狀早就判過了，它們是**盤點的限制**不是 remake 的缺口。判過的
// 結論要留在**產生報表的程式碼裡**，敘述會被讀成歷史，程式碼才會跟著報表一起
// 每次重生。
//
// ⚠ 歸不了類的要留在「還沒歸類」那一格：出現第五種形狀時它才會非零，
// 那時候才是真的有東西要看。
func silentShape(guard string) string {
	if guard == "" {
		return ""
	}
	if strings.Contains(guard, "4BF0") || strings.Contains(guard, "4BF1") {
		return tooltext.Text("h.dda6ee401ff2")
	}
	compare := strings.Index(guard, "COMPARE")
	if compare < 0 {
		return tooltext.Text("h.cace6089a166")
	}
	if exit := strings.Index(guard, "EXIT"); exit >= 0 && exit < compare {
		return tooltext.Text("h.59a613ab8858")
	}
	return tooltext.Text("h.fa920a07b092")
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
// snapshotDir 是主線段內快照的目錄（`captureInsideSegment` 產生）。設了之後，
// **冷進不去的段**會改從快照進——巫師塔與魔法商店那兩段的入口是一整段過場，
// 冷開一支隊伍走完會停在世界地圖，於是它們整段不在分母裡（spec 1193）。
var snapshotDir string

// snapshotEntered 記著哪些段是**從快照進**的。報表要標明——那些段的隊伍帶著
// 主線的旗標，和冷進的隊伍不是同一個起點，混在一起讀會高估「冷開就走得到」。
var snapshotEntered = map[string]bool{}

func enterDungeon(data corpus, seg segment.Segment) (*game.State, error) {
	state, err := data.NewParty()
	if err != nil {
		return nil, err
	}
	if err := state.EnterSegment(seg); err != nil {
		// 冷進不去就試段內快照。
		if fromSnapshot, snapErr := enterFromSnapshot(data, seg); snapErr == nil {
			return fromSnapshot, nil
		}
		return nil, tooltext.Errorf("h.8e8a76bdd692", seg.ID, err)
	}
	// ⚠ 盤點用的隊伍一律撐起來：有的段一進去就開打（古熔岩洞的伏擊），臨時建的
	// 一名角色會死在入口，後面一格都盤點不到。**只給盤點用**。
	if err := gamecorpus.BoostParty(&state); err != nil {
		return nil, err
	}
	trail, err := settleToDungeon(&state, 16)
	if err != nil {
		return nil, tooltext.Errorf("h.be4cbbc76969", seg.ID, err)
	}
	if state.Mode != game.ModeDungeon {
		if fromSnapshot, snapErr := enterFromSnapshot(data, seg); snapErr == nil {
			return fromSnapshot, nil
		}
		return nil, tooltext.Errorf("h.5e90960fec28", seg.ID, modeName(state.Mode), strings.Join(trail, " → "))
	}
	return &state, nil
}

// enterFromSnapshot 從主線的段內快照進這一段。
//
// ★ 為什麼要這條路：有些段的入口是**一整段過場**（巫師塔是德拉坎德羅斯那一串、
// 魔法商店那一段會回到世界地圖），冷開一支隊伍走完會停在世界地圖而不是地城，
// 於是那些段整段落在分母外——而它們的格子**玩家真的走得到**。
//
// ⚠ 快照是主線跑出來的，所以隊伍**帶著那一刻的旗標**。這和冷進的隊伍不同，
// 報表要標明那一段是從快照進的，不要混在一起讀。
func enterFromSnapshot(data corpus, seg segment.Segment) (*game.State, error) {
	if snapshotDir == "" {
		return nil, tooltext.Errorf("h.168af957deb3")
	}
	path := filepath.Join(snapshotDir, fmt.Sprintf("inside-block-%02X.json", seg.Block))
	state, err := data.NewParty()
	if err != nil {
		return nil, err
	}
	if err := state.LoadPartyFile(path); err != nil {
		return nil, err
	}
	snapshotEntered[seg.ID] = true
	if err := gamecorpus.BoostParty(&state); err != nil {
		return nil, err
	}
	if _, err := settleToDungeon(&state, 16); err != nil {
		return nil, err
	}
	if state.Mode != game.ModeDungeon {
		return nil, tooltext.Errorf("h.c44dec34df4b", modeName(state.Mode))
	}
	if block, ok := state.CurrentECLBlockID(); !ok || block != seg.Block {
		return nil, tooltext.Errorf("h.25eaa92efb12", block, seg.Block)
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
		return tooltext.Text("h.72726d8818f6")
	case hasLatin:
		return tooltext.Text("h.354b28c85333")
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
			if cell.note == tooltext.Text("h.5e63f0393483") {
				counts[tooltext.Text("h.cfe926ec77bf")]++
				continue
			}
			counts[tooltext.Text("h.9c9a4e85c872")]++
			switch {
			case cell.language == tooltext.Text("h.72726d8818f6"):
				counts[tooltext.Text("h.72726d8818f6")]++
			case cell.language == tooltext.Text("h.354b28c85333"):
				counts[tooltext.Text("h.354b28c85333")]++
			default:
				counts[tooltext.Text("h.0ff689f870c6")]++
				if shape := silentShape(cell.guard); shape != "" {
					counts[tooltext.Text("h.6b82ea5a970c")+shape]++
				} else {
					counts[tooltext.Text("h.3dbc2b54edb2")]++
				}
			}
			if cell.search {
				counts[tooltext.Text("h.2028ba65dbb5")]++
			}
			if cell.facing != 0 {
				counts[tooltext.Text("h.8047ab192f95")]++
			}
			if cell.seed > 1 {
				counts[tooltext.Text("h.f5e3d0b240ef")]++
			}
			if cell.precondition != "" {
				counts[tooltext.Text("h.aea4a28d595c")]++
			}
			// 重訪只在「有演出來」的格子上才有意義。
			if cell.played() && cell.outcomes > 1 {
				counts[tooltext.Text("h.1db93ca47c3e")]++
			}
			if cell.played() {
				switch cell.revisitKind {
				case tooltext.Text("h.82619569592c"):
					counts[tooltext.Text("h.dccfd35e55fd")]++
				case tooltext.Text("h.4a7e9a926adb"):
					counts[tooltext.Text("h.0bea72afb1f4")]++
				case tooltext.Text("h.f2632a784b0d"):
					counts[tooltext.Text("h.27d7edd5bc39")]++
				case tooltext.Text("h.a93b5a585c5f"):
					counts[tooltext.Text("h.4d4546c37c7c")]++
				default:
					counts[tooltext.Text("h.4523332d1adc")]++
				}
			}
		}
	}
	return counts
}

func render(sweeps []blockSweep) string {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.733e35bac38c") +
		tooltext.Text("h.1c1026e5ffe8") +
		tooltext.Text("h.37470b1b1a1b") +
		tooltext.Text("h.c6c0cb21fd20") +
		tooltext.Text("h.88dc906ad62e") +
		tooltext.Text("h.51cc7aab3588") +
		tooltext.Text("h.7163a81533a2") +
		tooltext.Text("h.6968e4d36c11") +
		tooltext.Text("h.7764430013eb") +
		tooltext.Text("h.5efd81768d99") +
		tooltext.Text("h.f9f834c38cea") +
		tooltext.Text("h.0e110e5eab83") +
		tooltext.Text("h.99e223cfc96a") +
		tooltext.Text("h.a6a3345c3d0b") +
		tooltext.Text("h.82b39604fd54") +
		tooltext.Text("h.d2d470c863d5") +
		tooltext.Text("h.8453f541cb15") +
		tooltext.Text("h.8cd4cbfa2e43") +
		tooltext.Text("h.7123ad6aa2bb"))
	for _, sweep := range sweeps {
		out.WriteString(fmt.Sprintf("## `%s`\n\n", sweep.id))
		if sweep.note != "" {
			out.WriteString(sweep.note + "\n\n")
			continue
		}
		if sweep.fromSnapshot {
			out.WriteString(tooltext.Text("h.bbdc6360e91c") +
				tooltext.Text("h.33894862d14d") +
				tooltext.Text("h.ee5268c4e060"))
		}
		out.WriteString(tooltext.Format("h.3e15dad6efdd", sweep.geoSet, sweep.geoBlock, sweep.mask))
		out.WriteString(tooltext.Text("h.af0e128b1082"))
		out.WriteString("|---:|---|---|---|---|---|---|\n")
		for _, cell := range sweep.cells {
			if cell.note == tooltext.Text("h.5e63f0393483") {
				out.WriteString(fmt.Sprintf("| %d | — | — | — | — | %s |\n",
					cell.index, cell.note))
				continue
			}
			extra := make([]string, 0, 3)
			if cell.search {
				extra = append(extra, tooltext.Text("h.2028ba65dbb5"))
			}
			if cell.facing != 0 {
				extra = append(extra, tooltext.Text("h.2019f9645edc")+facingName(cell.facing))
			}
			if cell.seed > 1 {
				extra = append(extra, tooltext.Format("h.3668e46d5caa", cell.seed))
			}
			if cell.precondition != "" {
				extra = append(extra, tooltext.Text("h.5c7e2bf7e64f")+cell.precondition+"`")
			}
			condition := tooltext.Text("h.47b0a29e00eb")
			if len(extra) > 0 {
				condition = strings.Join(extra, "＋")
			}
			text := cell.note
			if !cell.played() && cell.guard != "" {
				text += tooltext.Text("h.53559cef3d6d") + cell.guard + "`）"
			}
			if cell.played() {
				text = "「" + firstLine(cell.text) + "」"
				if cell.note != "" {
					text += "（" + cell.note + "）"
				}
				if cell.precondition != "" {
					text += tooltext.Text("h.53559cef3d6d") + cell.guard + "`）"
				}
			}
			if !cell.played() {
				condition = "—"
			}
			revisit := cell.revisitKind
			if !cell.played() || revisit == "" {
				revisit = "—"
			}
			if revisit == tooltext.Text("h.f2632a784b0d") {
				revisit = tooltext.Text("h.590f60558380") + firstLine(cell.revisit)
			}
			if cell.played() && cell.outcomes > 1 {
				// ⚠ 兩種結果的**第一句可能一樣**（差在後面幾句：擲到的怪、
				// 金額、成敗）。照樣印第一句會看起來像「兩種一模一樣」，
				// 讓人以為判定壞掉。
				other := firstLine(cell.otherOutcome)
				if other == firstLine(cell.text) {
					other = tooltext.Text("h.ccc8ad329481")
				}
				revisit += tooltext.Format("h.b7e74bc63b91", cell.outcomes, other)
			}
			out.WriteString(fmt.Sprintf("| %d | `%s` | `%02X` | %s | %s | %s | %s |\n",
				cell.index, cell.cell, cell.roof, condition, cell.language, revisit, text))
		}
		out.WriteString("\n")
	}
	counts := summarise(sweeps)
	out.WriteString(tooltext.Text("h.6bbcc6239f03"))
	out.WriteString(tooltext.Format("h.2b84e540005b", counts["block"]))
	out.WriteString(tooltext.Format("h.413125d75db1", counts[tooltext.Text("h.9c9a4e85c872")]))
	out.WriteString(tooltext.Format("h.7b4bfb3dbd00", counts[tooltext.Text("h.72726d8818f6")]))
	out.WriteString(tooltext.Format("h.afbfaaf5cb13", counts[tooltext.Text("h.354b28c85333")]))
	out.WriteString(tooltext.Format("h.efd4a0b96e7b", counts[tooltext.Text("h.0ff689f870c6")]))
	for _, shape := range silentShapes {
		out.WriteString(tooltext.Format("h.16fbadb6337f", shape, counts[tooltext.Text("h.6b82ea5a970c")+shape]))
	}
	out.WriteString(tooltext.Format("h.b3502130a4b3", counts[tooltext.Text("h.3dbc2b54edb2")]))
	out.WriteString(tooltext.Format("h.f3f415ce1b46", counts[tooltext.Text("h.2028ba65dbb5")]))
	out.WriteString(tooltext.Format("h.92c00a5d3bbb", counts[tooltext.Text("h.8047ab192f95")]))
	out.WriteString(tooltext.Format("h.e63caf827124", counts[tooltext.Text("h.f5e3d0b240ef")]))
	out.WriteString(tooltext.Format("h.747efa696546", counts[tooltext.Text("h.aea4a28d595c")]))
	out.WriteString(tooltext.Format("h.7017ae52195b", counts[tooltext.Text("h.cfe926ec77bf")]))
	out.WriteString(tooltext.Text("h.a82214d20ded"))
	out.WriteString(tooltext.Text("h.f9168c45c622") +
		tooltext.Text("h.b02ab6211076") +
		tooltext.Text("h.487481824449") +
		tooltext.Text("h.3f429bf0a03b"))
	out.WriteString(tooltext.Text("h.d282507df6f3") +
		tooltext.Text("h.cc399f9c1e70"))
	out.WriteString(tooltext.Text("h.078719e7005f") +
		tooltext.Text("h.ffad29f86a64") +
		tooltext.Text("h.6d61a1fc665b"))
	out.WriteString(tooltext.Text("h.a50220beba0e") +
		tooltext.Text("h.2a48f9abadd0") +
		tooltext.Text("h.f1f6e4543c0f") +
		tooltext.Text("h.3c5c59d5e5c6") +
		tooltext.Text("h.f8639528c609"))
	out.WriteString(tooltext.Text("h.25ce02601826"))
	out.WriteString(tooltext.Format("h.3257b93d5caa", counts[tooltext.Text("h.dccfd35e55fd")]))
	out.WriteString(tooltext.Format("h.24881ef9d582", counts[tooltext.Text("h.0bea72afb1f4")]))
	out.WriteString(tooltext.Format("h.8f6d0eb96243", counts[tooltext.Text("h.27d7edd5bc39")]))
	out.WriteString(tooltext.Format("h.e00e7042b328", counts[tooltext.Text("h.4d4546c37c7c")]))
	out.WriteString(tooltext.Format("h.dc5fed61be76", counts[tooltext.Text("h.4523332d1adc")]))
	out.WriteString(tooltext.Text("h.29dae100ad4c"))
	out.WriteString(tooltext.Text("h.9fae3631b172") +
		tooltext.Text("h.94dab993acc1") +
		tooltext.Text("h.91e84460d026") +
		tooltext.Text("h.c651cd0dc32b"))
	out.WriteString(tooltext.Text("h.4eaa9664a68a") +
		tooltext.Text("h.f4a5b679e388") +
		tooltext.Text("h.5c498f734707"))
	out.WriteString(tooltext.Format("h.1cd10c830cef", counts[tooltext.Text("h.72726d8818f6")]+counts[tooltext.Text("h.354b28c85333")]))
	out.WriteString(tooltext.Format("h.cf4e5e0fd6c9", counts[tooltext.Text("h.1db93ca47c3e")]))
	return out.String()
}

// facingName 把朝向碼翻成方位。朝向是 0..7 順時針，原作的 `C04D` ＝ 朝向 ÷ 2。
func facingName(facing uint8) string {
	switch facing / 2 {
	case 0:
		return tooltext.Text("h.01bc01b32975")
	case 1:
		return tooltext.Text("h.469e009b15a0")
	case 2:
		return tooltext.Text("h.b2f94b103450")
	case 3:
		return tooltext.Text("h.b9f0c815459c")
	}
	return "?"
}

func modeName(mode game.Mode) string {
	switch mode {
	case game.ModeTitle:
		return tooltext.Text("h.6fe38ed1ee10")
	case game.ModeWilderness:
		return tooltext.Text("h.58f78bc6a875")
	case game.ModeEvent:
		return tooltext.Text("h.c560201b331c")
	case game.ModeMap:
		return tooltext.Text("h.90e1b1b8a537")
	case game.ModePlace:
		return tooltext.Text("h.af2cfdb525ff")
	case game.ModeCombat:
		return tooltext.Text("h.625dd417c2c3")
	case game.ModeJournal:
		return tooltext.Text("h.53ce36b5da9e")
	case game.ModeCharacterCreation:
		return tooltext.Text("h.894d0b190ab5")
	case game.ModeDungeon:
		return tooltext.Text("h.c014266366bb")
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
		return "", tooltext.Text("h.a93b5a585c5f")
	}
	if state.Mode != game.ModeDungeon {
		return "", tooltext.Text("h.a93b5a585c5f")
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		return "", tooltext.Text("h.91e450980f03")
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
		return "", tooltext.Text("h.82619569592c")
	case second == first:
		return second, tooltext.Text("h.4a7e9a926adb")
	default:
		return second, tooltext.Text("h.f2632a784b0d")
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
		trail = append(trail, tooltext.Format("h.e0fe457c0fcc", modeName(state.Mode), len(state.Choices), firstLine(playerText(state))))
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return trail, tooltext.Errorf("h.effbedd6d7bb", err)
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
				return trail, tooltext.Errorf("h.503ee730e545", err, selectErr)
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
			if cell.note == tooltext.Text("h.5e63f0393483") {
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
