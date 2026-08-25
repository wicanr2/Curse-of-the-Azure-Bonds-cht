// Command segment-handoff 回答「分段驗收時，直入**漏掉**了多少上一段留下來的狀態」。
//
// ★ 存在的理由：`cmd/segment-seams` 已經比過接縫的 `LastECL`（18／18 對得上），
// 但那只是**一個欄位**。`AGENTS.md` §3 那條 ⚠ 講的是整包交接：
//
//	debug 旗標注入的是合成起始狀態，未必等於上一段真的跑出來的結束狀態；
//	兩端都綠不等於接縫通過。
//
// 這一支把「未必等於」變成數字：拿**主線真的走到那一段時的快照**，對上
// **直入那一段之後的狀態**，逐格比 ECL 記憶體。
//
// ⚠ 差異本身**不是缺陷**——直入本來就沒有前面的劇情。有意義的是**其中哪幾格
// 那一段的程式碼真的會讀**：那些才是「直入驗過了，但驗的不是主線會遇到的情況」。
// 讀取點由靜態可達指令的運算元推出來（`code 01h`／`03h` 是記憶體讀，`02h` 是
// 立即值，見 `internal/ecl` 的 `operandValue`）。
//
// ⚠ 靜態可達是**下界**：跟不到的碼不在分母裡，所以「會讀的格子」只會少不會多。
//
// 用法：
//
//	go run ./cmd/segment-handoff -snapshots workplace/campaign-frames/snapshots \
//	    -output docs/audit/segment-handoff.md -json docs/audit/segment-handoff.json
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

var members = []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"}

type handoffRow struct {
	ID string `json:"id"`
	// ReadCells 是這一段靜態可達的碼會**讀**的記憶體格數。
	ReadCells int `json:"read_cells"`
	// CampaignCells 是主線走到這一段時，記憶體裡有值的格數。
	CampaignCells int `json:"campaign_cells"`
	// Dropped 是主線有、直入沒有或值不同的格數 ＝ 直入漏掉的交接。
	Dropped int `json:"dropped"`
	// Risky 是 Dropped 之中**這一段的碼真的會讀**的那些。
	Risky []uint16 `json:"risky,omitempty"`
	// SeededDropped／SeededRisky 是**把交接狀態鋪回直入那側之後**還剩多少。
	// ★ 這是修法的效果：`State.SeedHandoffMemory` 讓分段驗收從主線真實的交接
	// 狀態進場，而不是合成的乾淨狀態。
	SeededDropped int      `json:"seeded_dropped"`
	SeededRisky   []uint16 `json:"seeded_risky,omitempty"`
	// ResidueOther 是這一段**歸不了因**的那幾格。★ 逐格印出來才追得下去：
	// 只有一個總數的話，下一步只能重跑整支工具。
	ResidueOther []uint16 `json:"residue_other,omitempty"`
	// Source 是主線那一份的取樣點：`arrival`（剛換到那一段）或 `inside`
	// （第一次站上地城，晚一些）。⚠ 兩者不可混為一談，見 `snapshotMemory`。
	Source string `json:"source,omitempty"`
	Note   string `json:"note,omitempty"`
}

type report struct {
	Schema string `json:"schema"`
	// Segments 是有快照可比的段數。
	Segments int `json:"segments"`
	// TotalDropped／TotalRisky 是各段加總。
	TotalDropped int `json:"total_dropped"`
	TotalRisky   int `json:"total_risky"`
	// RiskySegments 是「漏掉的格子裡有這一段會讀的」段數。
	RiskySegments int `json:"risky_segments"`
	// 鋪上交接狀態之後的同一組數字。
	SeededDropped       int `json:"seeded_dropped"`
	SeededRisky         int `json:"seeded_risky"`
	SeededRiskySegments int `json:"seeded_risky_segments"`
	// 剩下的差異按成因分：引擎進段時自己設的、由位置推出來的視圖暫存器、其餘。
	ResidueEngineSet int `json:"residue_engine_set"`
	ResidueViewCell  int `json:"residue_view_cell"`
	// ResidueLifecycleOwned 是「鋪了也沒用」的格子：直入那一側鋪與不鋪算出來的
	// 值**一模一樣** ⇒ 進段的 initial lifecycle 自己把它算掉了。
	//
	// ★ 這種格子**不是交接缺口**：它是那一段自己決定的東西，交接給不給都一樣。
	// 混進「歸不了因」會讓剩下的數字看起來比實際大。
	ResidueLifecycleOwned int `json:"residue_lifecycle_owned"`
	// ResidueRouteLookup／ResidueScratch：世界路線查表三格與 ECL 通用工作暫存器。
	// ⚠ 兩者都**不是交接缺口**：前者是「這一次要去哪」的當下運算值（直入沒有
	// 「正要去哪」這件事），後者是上一次運算剩下的東西（拿兩段不同執行歷史的
	// 工作暫存器互比沒有意義）。
	ResidueRouteLookup int `json:"residue_route_lookup"`
	ResidueScratch     int `json:"residue_scratch"`
	ResidueOther       int `json:"residue_other"`
	// FromArrival／FromInside 是取樣點的分佈。⚠ 一定要分開數：`inside` 那幾段
	// 量到的是「交接 ＋ 主線又走了一段」，混進總數會讓整份報表看起來比實際緊。
	FromArrival int `json:"from_arrival"`
	FromInside  int `json:"from_inside"`
	// Failed 是**比不成**的段數。⚠ 一定要和 `Segments` 分開：比不成會讓
	// 「漏掉 0 格」看起來和「完全對得上」一樣，而那兩件事天差地遠。
	Failed int          `json:"failed"`
	Rows   []handoffRow `json:"rows"`
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.33c48eb91ff0"))
	localePath := flag.String("locale", "assets/locale/zh-TW.json", tooltext.Text("h.9c3b4db6568f"))
	snapshots := flag.String("snapshots", "workplace/campaign-frames/snapshots", tooltext.Text("h.c2472b3d6b25"))
	partySize := flag.Int("party", 6, tooltext.Text("h.02936fb74120"))
	arrivals := flag.String("arrivals", "workplace/campaign-frames/arrivals",
		tooltext.Text("h.3fb0d0c8faa5"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	outputJSON := flag.String("json", "", tooltext.Text("h.7299c956bbb9"))
	flag.Parse()

	readBy := readAddressesByBlock(*image)

	data, err := os.ReadFile(*localePath)
	if err != nil {
		log.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		log.Fatal(err)
	}
	eclBlocks, initialECL, err := loadECLBlocks(*image)
	if err != nil {
		log.Fatal(err)
	}

	doc := report{Schema: "coab-segment-handoff/1"}
	for _, item := range segment.All() {
		campaign, source, ok := snapshotMemory(*arrivals, *snapshots, item.Block)
		if !ok {
			continue
		}
		row := handoffRow{ID: item.ID, ReadCells: len(readBy[item.Block]),
			CampaignCells: len(campaign), Source: source}

		// 跑**兩趟**：不鋪交接狀態（現況）與鋪上（修法）。兩個數字擺在一起，
		// 「這個修法有沒有用」才不是靠敘述，是靠同一份報表裡的兩欄。
		plain, err := directMemory(catalog, eclBlocks, initialECL, item, *partySize, nil)
		if err != nil {
			row.Note = err.Error()
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		seeded, err := directMemory(catalog, eclBlocks, initialECL, item, *partySize, campaign)
		if err != nil {
			row.Note = err.Error()
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		for address, value := range campaign {
			if plain[address] != value {
				row.Dropped++
				if readBy[item.Block][address] {
					row.Risky = append(row.Risky, address)
				}
			}
			if seeded[address] != value {
				row.SeededDropped++
				if readBy[item.Block][address] {
					row.SeededRisky = append(row.SeededRisky, address)
					switch {
					case residueKind(address) == "engine-set":
						doc.ResidueEngineSet++
					case residueKind(address) == "view-cell":
						doc.ResidueViewCell++
					case residueKind(address) == "route-lookup":
						doc.ResidueRouteLookup++
					case residueKind(address) == "scratch":
						doc.ResidueScratch++
					case plain[address] == seeded[address]:
						// 鋪與不鋪算出來一樣 ⇒ lifecycle 自己算掉了。
						doc.ResidueLifecycleOwned++
					default:
						doc.ResidueOther++
						row.ResidueOther = append(row.ResidueOther, address)
					}
				}
			}
		}
		sort.Slice(row.Risky, func(l, r int) bool { return row.Risky[l] < row.Risky[r] })
		sort.Slice(row.SeededRisky, func(l, r int) bool { return row.SeededRisky[l] < row.SeededRisky[r] })
		doc.SeededDropped += row.SeededDropped
		doc.SeededRisky += len(row.SeededRisky)
		if len(row.SeededRisky) > 0 {
			doc.SeededRiskySegments++
		}
		doc.Segments++
		if row.Source == "arrival" {
			doc.FromArrival++
		} else {
			doc.FromInside++
		}
		doc.TotalDropped += row.Dropped
		doc.TotalRisky += len(row.Risky)
		if len(row.Risky) > 0 {
			doc.RiskySegments++
		}
		doc.Rows = append(doc.Rows, row)
	}

	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(doc, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	text := render(doc)
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr,
		"segments=%d dropped=%d risky=%d | seeded: dropped=%d risky=%d (engine=%d view=%d lifecycle=%d route=%d scratch=%d other=%d) failed=%d\n",
		doc.Segments, doc.TotalDropped, doc.TotalRisky,
		doc.SeededDropped, doc.SeededRisky,
		doc.ResidueEngineSet, doc.ResidueViewCell, doc.ResidueLifecycleOwned,
		doc.ResidueRouteLookup, doc.ResidueScratch, doc.ResidueOther, doc.Failed)
}

// readAddressesByBlock 收集每個 block 的碼**會讀**的記憶體位址。
//
// ⚠ 只認 `code 01h`／`03h`：那兩個是 `operandValue` 真的去查 `memory[Word]` 的。
// `02h` 帶的也是 word，但那是**立即值**——把它算進來會讓每個 block 都「讀」一堆
// 其實只是常數的數字，而那個膨脹看起來和真的讀取一模一樣。
func readAddressesByBlock(image string) map[uint8]map[uint16]bool {
	reader, err := zip.OpenReader(image)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	out := map[uint8]map[uint16]bool{}
	for _, member := range members {
		payload, err := zipMember(&reader.Reader, member)
		if err != nil {
			continue
		}
		parsed, err := dax.Parse(payload)
		if err != nil {
			continue
		}
		for _, raw := range parsed {
			points, _, err := ecl.EntryPoints(raw.Data, 5)
			if err != nil {
				continue
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-ecl.CodeAddressBase)
			}
			graph, err := ecl.TraceGraph(raw.Data, starts, len(raw.Data)*8)
			if err != nil {
				continue
			}
			if out[raw.Entry.ID] == nil {
				out[raw.Entry.ID] = map[uint16]bool{}
			}
			for _, instruction := range graph.Instructions {
				for _, operand := range instruction.Operands {
					if !operand.WordSet {
						continue
					}
					if operand.Code == 0x01 || operand.Code == 0x03 {
						out[raw.Entry.ID][operand.Word] = true
					}
				}
			}
		}
	}
	return out
}

// snapshotMemory 讀主線走到那一段時的 ECL 記憶體。
//
// ★ **剛換到那一段的快照優先**（`arrival-block-XX.json`）：那才是交接的取樣點。
// 段內快照（`inside-block-XX.json`）是隊伍第一次站上地城時存的，那時 initial
// lifecycle 已經跑完、隊伍也走了幾步 ⇒ 用它量到的是「交接 ＋ 又走了一段」。
// 兩者都印出來，才看得出取樣點差多少。
func snapshotMemory(arrivals, inside string, block uint8) (map[uint16]uint16, string, bool) {
	if memory, ok := arrivalSample(
		filepath.Join(arrivals, fmt.Sprintf("arrival-block-%02X.json", block))); ok {
		return memory, "arrival", true
	}
	if memory, ok := memoryFromFile(
		filepath.Join(inside, fmt.Sprintf("inside-block-%02X.json", block))); ok {
		return memory, "inside", true
	}
	return nil, "", false
}

// arrivalSample 讀換段那一刻的記憶體取樣。
//
// ⚠ 這個檔**不是存檔**（見 `internal/game.ArrivalSample`）：取樣點在指令執行到
// 一半的地方，載回來當存檔用會解不出指令。它只有記憶體，而記憶體正是這一支要的。
func arrivalSample(path string) (map[uint16]uint16, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var file struct {
		Schema string            `json:"schema"`
		Memory map[string]uint16 `json:"memory"`
	}
	if json.Unmarshal(raw, &file) != nil || file.Schema != "coab-arrival-sample/1" {
		return nil, false
	}
	return parseMemory(file.Memory), true
}

func parseMemory(raw map[string]uint16) map[uint16]uint16 {
	out := make(map[uint16]uint16, len(raw))
	for key, value := range raw {
		address, err := strconv.ParseUint(key, 10, 16)
		if err != nil {
			continue
		}
		out[uint16(address)] = value
	}
	return normalise(out)
}

// normalise 把兩側的記憶體收斂到**同一個定義域**，否則比出來的數字沒有意義。
//
// ⚠ 兩個來源本來就不一樣：到達取樣走 `MemorySnapshot()`（**整份**，含程式碼視窗
// 與零），存檔那一側走存檔編碼器（零不收、程式碼另存一塊）。不對齊就直接比的話，
// 差異會從 1,148 暴增到 103,896——**而那個數字看起來一樣像個結果**。
//
// 兩條規則：
//
//   - 去掉程式碼視窗 `8000h..9DFFh`：那是位元組碼不是狀態。
//   - 去掉值為 0 的格：原作分不出「寫過 0」與「沒寫過」，存檔那一側也不收。
func normalise(memory map[uint16]uint16) map[uint16]uint16 {
	out := make(map[uint16]uint16, len(memory))
	for address, value := range memory {
		if value == 0 || (address >= 0x8000 && address <= 0x9DFF) {
			continue
		}
		out[address] = value
	}
	return out
}

func memoryFromFile(path string) (map[uint16]uint16, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var file struct {
		Session struct {
			Memory map[string]uint16 `json:"memory"`
		} `json:"ecl_session"`
	}
	if json.Unmarshal(raw, &file) != nil {
		return nil, false
	}
	return parseMemory(file.Session.Memory), true
}

func loadECLBlocks(image string) (map[uint8][]byte, uint8, error) {
	reader, err := zip.OpenReader(image)
	if err != nil {
		return nil, 0, err
	}
	defer reader.Close()
	blocks := map[uint8][]byte{}
	first := uint8(0)
	found := false
	for _, member := range members {
		payload, err := zipMember(&reader.Reader, member)
		if err != nil {
			continue
		}
		parsed, err := dax.Parse(payload)
		if err != nil {
			continue
		}
		for _, raw := range parsed {
			blocks[raw.Entry.ID] = raw.Data
			if !found {
				first, found = raw.Entry.ID, true
			}
		}
	}
	if !found {
		return nil, 0, tooltext.Errorf("h.e201a34780b4", image)
	}
	return blocks, first, nil
}

func zipMember(reader *zip.Reader, name string) ([]byte, error) {
	for _, file := range reader.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer handle.Close()
		return io.ReadAll(handle)
	}
	return nil, tooltext.Errorf("h.844fa35500c5", name)
}

// directMemory 造一個直入那一段的狀態，回傳正規化過的 ECL 記憶體。
// `seed` 非 nil 就先把交接狀態鋪上去（`State.SeedHandoffMemory`）。
//
// ⚠ 每一趟都要**全新的 State**：鋪過交接狀態的那一個不能重用，否則第二趟量到的
// 是第一趟的殘留。
func directMemory(catalog locale.Catalog, blocks map[uint8][]byte, initial uint8,
	item segment.Segment, partySize int, seed map[uint16]uint16) (map[uint16]uint16, error) {
	state := game.NewStateFromECLBlocks(catalog, blocks, initial)
	// ⚠ 直入要先有隊伍，`cmd/azure-bonds-game -segment` 也是這樣做的。
	// 少了這一步 `EnterSegment` 會失敗，而失敗路徑會讓「漏掉幾格」印成 **0**
	// ——那個零看起來和「完全對得上」一模一樣。
	if err := createParty(&state, partySize); err != nil {
		return nil, tooltext.Errorf("h.c28594bfaffd", err)
	}
	if seed != nil {
		state.SeedHandoffMemory(seed)
	}
	if err := state.EnterSegment(item); err != nil {
		return nil, tooltext.Errorf("h.3cd21f238059", err)
	}
	// ⚠ 兩側走**同一個存取器**：到達取樣是 `MemorySnapshot()`，這裡也是。
	// 一邊走存檔編碼器會讓定義域不同（不收 0、程式碼另存），差異暴增兩個
	// 數量級——而那個數字看起來一樣像個結果。
	memory := normalise(state.ECLMemorySnapshot())
	if len(memory) == 0 {
		return nil, tooltext.Errorf("h.a085e513f65b")
	}
	return memory, nil
}

// residueKind 把「鋪了交接狀態還是不一樣」的格子按成因分類。
//
// ★ 分類的意義：引擎進段時**自己就要設**的格子（`4BF2h` LastECL、`7ED5h`、
// `7EC9h`）與由隊伍位置推出來的視圖暫存器（`C04Bh`..`C05Fh`，如 `C04Fh`
// ＝ 牆頂）**本來就不該被交接狀態蓋過去**——它們不是劇情旗標。把它們和真正
// 剩下來的差異混在一起數，會讓修法看起來比實際差。
func residueKind(address uint16) string {
	switch address {
	case 0x4BF2, 0x7ED5, 0x7EC9:
		return "engine-set"
	// ★ 世界路線的查表三格：`ECL1/0x50:110Ch GETTABLE 9D13 4C9D 4CA1`——
	// `4C9Bh` 是目的地選擇槽（remake 這一側是 `pendingWorldDestination`），
	// `4C9Dh` 是查表的索引，`4CA1h` 是查出來的城市（`CurrentCity`，spec 1163）。
	//
	// ⚠ 直入**沒有「正要去哪裡」**這件事，所以這三格必然和主線不同。
	// 它們不是劇情旗標，是「這一次世界移動要去哪」的當下運算值。
	case 0x4C9B, 0x4C9D, 0x4CA1:
		return "route-lookup"
	// ★ ECL 的通用工作暫存器。`7F79h`／`7F7Ah`／`7F7Bh` 是全作最熱的三格
	// （167／42／20 次），落在「ECL 私有、引擎側查無存取」那一側，而 spec 1096 §四
	// 說它們**形狀上就是通用工作暫存器**（被算術指令反覆使用）。
	//
	// ⚠ 工作暫存器的值是「上一次運算剩下的東西」。拿兩段**不同執行歷史**的
	// 工作暫存器互比沒有意義——不同不代表少了什麼。
	case 0x7F79, 0x7F7A, 0x7F7B:
		return "scratch"
	}
	if address >= 0xC04B && address <= 0xC05F {
		return "view-cell"
	}
	return "other"
}

// createParty 造一支隊伍。
//
// ⚠ **人數要和主線一樣（六人）**：`cmd/azure-bonds-game -segment` 只造一名，
// 而隊伍大小會讓一整批隊伍相關的 ECL 格子必然不同——那些差異和「交接漏掉了
// 什麼」無關，只是把上界撐大。
func createParty(state *game.State, size int) error {
	if len(state.PartyFighters()) > 0 {
		return nil
	}
	if err := state.OpenCharacterCreation(); err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		if err := state.AddCreationCharacter(0); err != nil {
			return tooltext.Errorf("h.4d919dc48068", i+1, err)
		}
	}
	return state.FinishCharacterCreation()
}

func render(doc report) string {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.59f18c557cc1"))
	out.WriteString(tooltext.Text("h.95349d4856f7"))
	out.WriteString(tooltext.Text("h.4f17098667a5") +
		tooltext.Text("h.99ebe84d85ac") +
		tooltext.Text("h.20b4362ffd39"))
	out.WriteString(tooltext.Text("h.2868e9aed2fd") +
		tooltext.Text("h.2ec3dc8ccedf") +
		tooltext.Text("h.41700d4e55a0"))
	out.WriteString(tooltext.Text("h.2628b6dcd71f") +
		tooltext.Text("h.25c3a8630e1a") +
		tooltext.Text("h.2c9975f89333"))
	out.WriteString(tooltext.Text("h.3bceddd7c6b5") +
		tooltext.Text("h.50ed660b50e3") +
		tooltext.Text("h.57113ce6fc71") +
		tooltext.Text("h.37eba446b1c3"))
	out.WriteString(tooltext.Text("h.a10baf9369e2") +
		tooltext.Text("h.4585f60cfaa2"))
	out.WriteString(tooltext.Text("h.bd479e441fc9") +
		tooltext.Text("h.2854ed0a6ab5") +
		tooltext.Text("h.7e2c91d6ed97"))
	out.WriteString(tooltext.Text("h.f05ea5517a68") +
		"|---|---|---:|---:|---:|---|---:|---|\n")
	for _, row := range doc.Rows {
		risky := "—"
		if len(row.Risky) > 0 {
			parts := make([]string, 0, len(row.Risky))
			for _, address := range row.Risky {
				parts = append(parts, fmt.Sprintf("`%04Xh`", address))
			}
			risky = "**" + strconv.Itoa(len(row.Risky)) + "**：" + strings.Join(parts, "、")
		}
		if row.Note != "" {
			risky = row.Note
		}
		source := row.Source
		if source == "arrival" {
			source = tooltext.Text("h.3e6ac92fe2dc")
		} else if source == "inside" {
			source = tooltext.Text("h.4c19bf484f18")
		}
		seededRisky := "—"
		if len(row.SeededRisky) > 0 {
			parts := make([]string, 0, len(row.SeededRisky))
			for _, address := range row.SeededRisky {
				parts = append(parts, fmt.Sprintf("`%04Xh`", address))
			}
			seededRisky = strconv.Itoa(len(row.SeededRisky)) + "：" + strings.Join(parts, "、")
		}
		fmt.Fprintf(&out, "| `%s` | %s | %d | %d | %d | %s | **%d** | %s |\n",
			row.ID, source, row.ReadCells, row.CampaignCells, row.Dropped, risky,
			row.SeededDropped, seededRisky)
	}
	fmt.Fprintf(&out, tooltext.Text("h.bdbf082c9d34")+
		tooltext.Text("h.c48c79220480")+
		tooltext.Text("h.dcd46adce27b"),
		doc.Segments, doc.FromArrival, doc.FromInside,
		doc.TotalDropped, doc.TotalRisky, doc.RiskySegments)
	fmt.Fprintf(&out, tooltext.Text("h.78c328b7540d")+
		tooltext.Text("h.39db215a70fe"),
		doc.SeededDropped, doc.TotalDropped, doc.SeededRisky, doc.TotalRisky,
		doc.SeededRiskySegments, doc.RiskySegments)
	fmt.Fprintf(&out, tooltext.Text("h.b6e52d920240")+
		tooltext.Text("h.70b76f50f66b")+
		tooltext.Text("h.d85de8b2d654")+
		tooltext.Text("h.f6534dc2960d")+
		tooltext.Text("h.bc94c4430a33")+
		tooltext.Text("h.a4c17a90f36c")+
		tooltext.Text("h.045f3be5334b")+
		tooltext.Text("h.72489c47680c")+
		tooltext.Text("h.b1d3291a561e")+
		tooltext.Text("h.c7a0914181ef")+
		tooltext.Text("h.15baca7c7821"),
		doc.SeededRisky, doc.ResidueEngineSet, doc.ResidueViewCell,
		doc.ResidueLifecycleOwned, doc.ResidueRouteLookup, doc.ResidueScratch,
		doc.ResidueOther)
	if doc.FromInside > 0 {
		out.WriteString(tooltext.Text("h.e8e7d88ec22a") +
			tooltext.Text("h.2f8e97034b7d") +
			tooltext.Text("h.c69cb87f20a7"))
	}
	if doc.Failed > 0 {
		fmt.Fprintf(&out, tooltext.Text("h.670e220eb8a9")+
			tooltext.Text("h.fa98f99f9b5d")+
			tooltext.Text("h.1a30b82661f9"), doc.Failed)
	}
	return out.String()
}
