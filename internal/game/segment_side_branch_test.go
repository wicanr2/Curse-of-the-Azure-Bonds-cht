package game

import (
	"archive/zip"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// 段內支線：主線走到某一段的當下，那張地圖上除了主線路線以外的格子演出什麼。
//
// ★ 這跟 `cmd/cell-sweep` 問的不是同一件事。那一支**每一格都重新進段**、用一支
// 撐起來的盤點隊伍，答的是「這一格演不演得出來、是不是中文」。這裡站上去的是
// **帶著主線進度、once-only 旗標已經被主線設過**的隊伍，所以它是另一次取樣：
// 同一格在不同旗標狀態下演的可能是別的東西，也可能什麼都不演。
//
// ⚠ **「沒演出來」在這裡不是缺口訊號。** 段界那個時點，那張圖上多數事件已經被
// 主線自己觸發掉了（once-only），演不出來是正常的。這一支要擋的是**演出來的字
// 落回原文**，以及走訪整個塌掉（非空閘門）。

// sideBranchCell 是一個分派索引在某一份段界快照的狀態下站上去的結果。
type sideBranchCell struct {
	index  int
	x, y   int
	roof   uint8
	text   string
	facing uint8
	search bool
}

// played 為真代表這一格真的演出了字。
func (c sideBranchCell) played() bool { return c.text != "" }

// sideBranchSweep 是一份段界快照掃完的結果。
type sideBranchSweep struct {
	name     string
	block    uint8
	geoSet   uint8
	geoBlock uint8
	mask     int
	skipped  string
	cells    []sideBranchCell
}

// sideBranchSeeds 是演不出來時要換幾顆 ECL 亂數種子再試。墓園的盜墓者是 100 抽
// 32，一次不一定演得出來；配上兩種搜尋狀態與四個朝向，一格最多試 16 次。
// ⚠ 上限要克制：這一支掛在整條主線測試底下，演不出來的格子是**每一格都試滿**，
// 種子數直接乘進整包測試的時間。
const sideBranchSeeds = 2

// sweepSideBranches 把一張地圖上每一個地形分派索引的第一格逐格站上去。
//
// restore 每次都給一份**乾淨的**快照狀態：once-only 旗標與戰鬥結果會互相污染，
// 接續著跑會讓後面的格子看起來沒內容。
func sweepSideBranches(name string, load func() (State, error),
	blocks map[uint8][]byte, catalogs map[uint8]geo.Catalog) sideBranchSweep {
	out := sideBranchSweep{name: name}
	// ⚠ 段界快照不一定停在地城的行走畫面上——有幾份停在事件或選單上。停在那裡
	// 就掃不了，但那不是「這一段沒有段內支線」，只是還沒把畫面推完。
	// `settleIntoDungeon` 按繼續／選第一項把它推回地城，推不回去才跳過。
	restore := func() (State, error) {
		state, err := load()
		if err != nil {
			return state, err
		}
		return state, settleIntoDungeon(&state)
	}
	probe, err := restore()
	if err != nil {
		out.skipped = "推不回地城：" + err.Error()
		return out
	}
	// ★ 分派器與地圖**一定要取自同一個來源**——讀回來、推穩之後的那個狀態。
	// 段界記的 block 是**那一段結束時**的 block，常常已經走進下一段而地圖還沒換；
	// 拿它去查分派器會用 A 段的地形碼表配 B 段的地圖，得到一堆看起來像內容缺口
	// 的假紅。
	out.block = probe.session.CurrentBlockID()
	out.geoSet, out.geoBlock = probe.GeoMapSet, probe.GeoMapBlock
	dispatch := eclcells.Analyze(blocks[out.block])
	if !dispatch.Found || dispatch.TableForm {
		out.skipped = "這一段沒有地形碼分派器"
		return out
	}
	out.mask = dispatch.Mask
	// 分派器自己宣告的地圖區塊要跟狀態上的一致，否則兩邊講的不是同一張圖。
	// ⚠ `GeoBlock` 為 0 是「這個 block 裡找不到載圖指令」，不是「配的是第 0 張」；
	// 那種情況以狀態上的地圖為準，不要當成不一致把整段丟掉。
	if dispatch.GeoBlock != 0 && dispatch.GeoBlock != out.geoBlock {
		out.skipped = fmt.Sprintf("分派器配的是 GEO 區塊 %#02x，狀態在 %#02x",
			dispatch.GeoBlock, out.geoBlock)
		return out
	}
	catalog, ok := catalogs[out.geoSet]
	if !ok {
		out.skipped = fmt.Sprintf("讀不到 GEO%d", out.geoSet)
		return out
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: out.geoSet, BlockID: out.geoBlock})
	if !ok {
		out.skipped = fmt.Sprintf("GEO%d 裡沒有區塊 %#02x", out.geoSet, out.geoBlock)
		return out
	}

	firstCell := map[int][2]int{}
	roofs := map[int]uint8{}
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			roof := grid.CellWrapped(x, y).Terrain
			index := int(roof) & dispatch.Mask
			if _, seen := firstCell[index]; !seen {
				firstCell[index], roofs[index] = [2]int{x, y}, roof
			}
		}
	}
	for _, index := range dispatch.Indexes {
		// 索引 0 是「沒有事件的地面」，全圖大半都是它。
		if index == 0 {
			continue
		}
		cell, ok := firstCell[index]
		if !ok {
			continue
		}
		result := sideBranchCell{index: index, x: cell[0], y: cell[1], roof: roofs[index]}
		out.cells = append(out.cells, standOnSideBranch(restore, result))
	}
	return out
}

// standOnSideBranch 把隊伍放到一格上跑生命週期，演不出來就換條件再試。
//
// ⚠ 朝向是 0..7 而 `C04D ＝ 朝向 ÷ 2`：邊界那一圈的處理常式是
// `COMPARE C04D <方向> / IF <> / EXIT`，只掃 0..3 會讓南、西兩面整批假零。
func standOnSideBranch(restore func() (State, error), result sideBranchCell) sideBranchCell {
	for seed := 1; seed <= sideBranchSeeds; seed++ {
		for _, search := range []bool{false, true} {
			for _, facing := range []uint8{0, 2, 4, 6} {
				state, err := restore()
				if err != nil {
					continue
				}
				// ★ 讀回來的狀態身上還掛著**上一句話**（存檔存的就是當時畫面上
				// 的字）。不清掉就會把它當成「這一格演出來的」——每一格都「演得
				// 出來」而且演的是同一句，143 格全是假的。
				state.Message, state.Prompt = "", ""
				state.SetECLSeed(int64(seed))
				state.SetDungeonGeometryView(result.x, result.y, facing)
				state.DungeonWallRoof = result.roof
				state.DungeonSearchEnabled = search
				if err := state.RunDungeonLifecycle(); err != nil {
					continue
				}
				text := strings.TrimSpace(state.Message)
				if text == "" {
					text = strings.TrimSpace(state.Prompt)
				}
				if text == "" {
					continue
				}
				result.text, result.facing, result.search = text, facing, search
				return result
			}
		}
	}
	return result
}

// summarize 把一份掃描結果排成一行給 `t.Log`。
func (s sideBranchSweep) summarize() string {
	if s.skipped != "" {
		return fmt.Sprintf("%s：跳過（%s）", s.name, s.skipped)
	}
	played := make([]string, 0, len(s.cells))
	missing := make([]string, 0, len(s.cells))
	for _, cell := range s.cells {
		label := fmt.Sprintf("%d@(%d,%d)", cell.index, cell.x, cell.y)
		if cell.played() {
			played = append(played, label)
			continue
		}
		missing = append(missing, label)
	}
	sort.Strings(played)
	sort.Strings(missing)
	return fmt.Sprintf("%s（ECL block %#02x、GEO%d/%#02x，遮罩 %#02x）：%d／%d 格演得出來；沒演出來 %s",
		s.name, s.block, s.geoSet, s.geoBlock, s.mask, len(played), len(s.cells),
		strings.Join(missing, " "))
}

// optionalZipData 讀 image 裡的一個成員，不存在就回 nil。⚠ 第一章沒有
// `GEO1.DAX`——世界地圖不走地城地圖，缺它是正常的，不是壞掉。
func optionalZipData(archive *zip.ReadCloser, name string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
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

// settleIntoDungeonSteps 是把畫面推回地城最多按幾下。**要有上限**：世界地圖上的
// 段界一直按下去會走進旅途，那不是這一支要量的東西。
const settleIntoDungeonSteps = 8

// settleIntoDungeon 把一份讀回來的快照推到地城的行走畫面上。推不到就回錯誤，
// 由呼叫端決定跳過。
func settleIntoDungeon(state *State) error {
	for step := 0; step < settleIntoDungeonSteps; step++ {
		if state.Mode == ModeDungeon && !state.CombatActive() {
			return nil
		}
		if err := stepRestoredCampaignState(state); err != nil {
			return err
		}
	}
	if state.Mode != ModeDungeon {
		return fmt.Errorf("按了 %d 下還沒回到地城", settleIntoDungeonSteps)
	}
	return nil
}

// firstRunes 截前 n 個字，超過就加省略號。
func firstRunes(text string, n int) string {
	glyphs := []rune(text)
	if len(glyphs) <= n {
		return text
	}
	return string(glyphs[:n]) + "…"
}
