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
	Note  string   `json:"note,omitempty"`
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
	// Failed 是**比不成**的段數。⚠ 一定要和 `Segments` 分開：比不成會讓
	// 「漏掉 0 格」看起來和「完全對得上」一樣，而那兩件事天差地遠。
	Failed int `json:"failed"`
	Rows          []handoffRow `json:"rows"`
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "語系檔")
	snapshots := flag.String("snapshots", "workplace/campaign-frames/snapshots", "主線快照目錄")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	outputJSON := flag.String("json", "", "JSON 輸出路徑")
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

	temp, err := os.MkdirTemp("", "handoff")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(temp)

	doc := report{Schema: "coab-segment-handoff/1"}
	for _, item := range segment.All() {
		campaign, ok := snapshotMemory(*snapshots, item.Block)
		if !ok {
			continue
		}
		row := handoffRow{ID: item.ID, ReadCells: len(readBy[item.Block]), CampaignCells: len(campaign)}

		state := game.NewStateFromECLBlocks(catalog, eclBlocks, initialECL)
		// ⚠ 直入要先有隊伍，`cmd/azure-bonds-game -segment` 也是這樣做的。
		// 少了這一步 `EnterSegment`／`SavePartyFile` 會失敗，而失敗路徑會讓
		// 「漏掉幾格」印成 **0**——那個零看起來和「完全對得上」一模一樣。
		if err := createOnePartyMember(&state); err != nil {
			row.Note = fmt.Sprintf("建不出隊伍：%v", err)
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		if err := state.EnterSegment(item); err != nil {
			row.Note = fmt.Sprintf("直入進不去：%v", err)
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		path := filepath.Join(temp, fmt.Sprintf("%02X.json", item.Block))
		if err := state.SavePartyFile(path); err != nil {
			row.Note = fmt.Sprintf("直入存不出來：%v", err)
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		direct, ok := memoryFromFile(path)
		if !ok {
			row.Note = "直入的存檔讀不回來"
			doc.Failed++
			doc.Rows = append(doc.Rows, row)
			continue
		}
		for address, value := range campaign {
			if direct[address] == value {
				continue
			}
			row.Dropped++
			if readBy[item.Block][address] {
				row.Risky = append(row.Risky, address)
			}
		}
		sort.Slice(row.Risky, func(l, r int) bool { return row.Risky[l] < row.Risky[r] })
		doc.Segments++
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
	fmt.Fprintf(os.Stderr, "segments=%d dropped=%d risky=%d riskySegments=%d failed=%d\n",
		doc.Segments, doc.TotalDropped, doc.TotalRisky, doc.RiskySegments, doc.Failed)
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

// snapshotMemory 讀主線走到那一段時的 ECL 記憶體。段內快照優先，沒有就算了。
func snapshotMemory(dir string, block uint8) (map[uint16]uint16, bool) {
	for _, name := range []string{
		fmt.Sprintf("inside-block-%02X.json", block),
	} {
		if memory, ok := memoryFromFile(filepath.Join(dir, name)); ok {
			return memory, true
		}
	}
	return nil, false
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
	out := make(map[uint16]uint16, len(file.Session.Memory))
	for key, value := range file.Session.Memory {
		address, err := strconv.ParseUint(key, 10, 16)
		if err != nil {
			continue
		}
		out[uint16(address)] = value
	}
	return out, true
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
		return nil, 0, fmt.Errorf("%s 裡沒有 ECL 區塊", image)
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
	return nil, fmt.Errorf("%s 不在 image 裡", name)
}

// createOnePartyMember 造一名隊員，和 `cmd/azure-bonds-game -segment` 同一條路。
func createOnePartyMember(state *game.State) error {
	if len(state.PartyFighters()) > 0 {
		return nil
	}
	if err := state.OpenCharacterCreation(); err != nil {
		return err
	}
	if err := state.AddCreationCharacter(0); err != nil {
		return err
	}
	return state.FinishCharacterCreation()
}

func render(doc report) string {
	var out strings.Builder
	out.WriteString("# 分段驗收漏掉多少交接：直入的狀態 vs 主線走到時的狀態\n\n")
	out.WriteString("由 `cmd/segment-handoff` 產生，不要手改。\n\n")
	out.WriteString("★ `cmd/segment-seams` 比的是接縫的 `LastECL` 一個欄位（18／18 對得上）。" +
		"這一份比的是**整包 ECL 記憶體**：主線真的走到那一段時的快照，對上直入那一段" +
		"之後的狀態。\n\n")
	out.WriteString("⚠ **差異本身不是缺陷**——直入本來就沒有前面的劇情。有意義的是" +
		"**其中哪幾格那一段的程式碼真的會讀**（`會讀且漏掉` 那一欄）：那些才是" +
		"「直入驗過了，但驗的不是主線會遇到的情況」。\n\n")
	out.WriteString("⚠ 「會讀」由**靜態可達指令**的運算元推出來（`code 01h`／`03h` 是記憶體讀，" +
		"`02h` 是立即值）。靜態可達是下界 ⇒ **會讀的格子只會少不會多**，" +
		"所以 `會讀且漏掉` 也是下界。\n\n")
	out.WriteString("⚠ **比較的時間點不是「剛進段的那一瞬間」。** 主線那一份是隊伍" +
		"**第一次站上那一段地城**時存的，那時 initial lifecycle 已經跑完、隊伍也走了幾步。" +
		"所以有些差異只是「主線走得比較前面」，不是交接漏掉的東西——" +
		"`4BF0h`／`4BF1h`（移動前的座標快照）與 `4BF2h`（`LastECL`）都屬於這一類。" +
		"這一欄要當成**上界**看：它是「直入與主線在這一段的狀態差多少」，" +
		"其中一部分才是交接。\n\n")
	out.WriteString("⚠ 直入這一側只有**一名**隊員（和 `cmd/azure-bonds-game -segment` 同一條路），" +
		"主線是六人。隊伍相關的格子必然不同。\n\n")
	out.WriteString("| 段 | 這一段會讀幾格 | 主線走到時有值幾格 | 直入漏掉幾格 | 會讀且漏掉 |\n" +
		"|---|---:|---:|---:|---|\n")
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
		fmt.Fprintf(&out, "| `%s` | %d | %d | %d | %s |\n",
			row.ID, row.ReadCells, row.CampaignCells, row.Dropped, risky)
	}
	fmt.Fprintf(&out, "\n合計 %d 段比得成：直入漏掉 **%d** 格，"+
		"其中**那一段的碼真的會讀**的有 **%d** 格，分佈在 **%d** 段。\n",
		doc.Segments, doc.TotalDropped, doc.TotalRisky, doc.RiskySegments)
	if doc.Failed > 0 {
		fmt.Fprintf(&out, "\n⚠ 另有 **%d** 段**比不成**（見上表的備註）。"+
			"比不成不等於沒有差異——失敗路徑會讓「漏掉 0 格」看起來和「完全對得上」"+
			"一模一樣，所以兩者分開數。\n", doc.Failed)
	}
	return out.String()
}
