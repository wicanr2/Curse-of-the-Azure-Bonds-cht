// Package eclcatalog builds a deterministic, evidence-bounded inventory of
// the original CoAB ECL corpus. It catalogs parser/control-flow observations;
// it does not promote static discovery order to original runtime order.
package eclcatalog

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

const FormatVersion = 1

var memberNames = []string{
	"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX",
}

var lifecycleNames = []string{
	"per_turn", "search_location", "pre_camp", "camp_interrupted", "initial",
}

type Catalog struct {
	FormatVersion int      `json:"format_version"`
	Generator     string   `json:"generator"`
	Source        Source   `json:"source"`
	Limitations   []string `json:"limitations"`
	Summary       Summary  `json:"summary"`
	Members       []Member `json:"members"`
}

type Source struct {
	Archive string `json:"archive"`
	SHA256  string `json:"sha256"`
}

type Summary struct {
	MemberCount                     int            `json:"member_count"`
	BlockCount                      int            `json:"block_count"`
	LifecycleEntryCount             int            `json:"lifecycle_entry_count"`
	UniqueReachableInstructionCount int            `json:"unique_reachable_instruction_count"`
	OpcodeCounts                    map[string]int `json:"opcode_counts"`
	OrderedEffectCandidateCount     int            `json:"ordered_effect_candidate_count"`
}

type Member struct {
	Name   string  `json:"name"`
	SHA256 string  `json:"sha256"`
	Blocks []Block `json:"blocks"`
}

type Block struct {
	ID                      string                   `json:"id"`
	DecodedSize             int                      `json:"decoded_size"`
	LifecycleEntries        []LifecycleEntry         `json:"lifecycle_entries"`
	Instructions            []Instruction            `json:"instructions"`
	Edges                   []Edge                   `json:"edges"`
	OrderedEffectCandidates []OrderedEffectCandidate `json:"ordered_effect_candidates,omitempty"`
}

type LifecycleEntry struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	CodeAddress string `json:"code_address"`
	Offset      string `json:"offset"`
}

type Instruction struct {
	Offset        string    `json:"offset"`
	CodeAddress   string    `json:"code_address"`
	Opcode        string    `json:"opcode"`
	Name          string    `json:"name"`
	NextOffset    string    `json:"next_offset"`
	Operands      []Operand `json:"operands,omitempty"`
	EffectKinds   []string  `json:"effect_kinds,omitempty"`
	ReachableFrom []string  `json:"reachable_from"`
}

type Operand struct {
	Code         string `json:"code"`
	Low          string `json:"low"`
	High         string `json:"high,omitempty"`
	Word         string `json:"word,omitempty"`
	PackedLength int    `json:"packed_length,omitempty"`
	PackedSHA256 string `json:"packed_sha256,omitempty"`
}

type Edge struct {
	From          string   `json:"from"`
	To            string   `json:"to"`
	Kind          string   `json:"kind"`
	ReachableFrom []string `json:"reachable_from"`
}

type OrderedEffectCandidate struct {
	StartOffset   string            `json:"start_offset"`
	EndOffset     string            `json:"end_offset"`
	Effects       []CandidateEffect `json:"effects"`
	ReachableFrom []string          `json:"reachable_from"`
	Evidence      string            `json:"evidence"`
}

type CandidateEffect struct {
	Offset string   `json:"offset"`
	Opcode string   `json:"opcode"`
	Name   string   `json:"name"`
	Kinds  []string `json:"kinds"`
}

type aggregateInstruction struct {
	value ecl.Instruction
	from  map[string]bool
}

type aggregateEdge struct {
	value ecl.Edge
	from  map[string]bool
}

type aggregateCandidate struct {
	value OrderedEffectCandidate
	from  map[string]bool
}

func BuildFile(path string) (Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, err
	}
	return Build(filepath.Base(path), data)
}

func Build(archiveName string, archiveData []byte) (Catalog, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return Catalog{}, fmt.Errorf("open archive: %w", err)
	}
	catalog := Catalog{
		FormatVersion: FormatVersion,
		Generator:     "cmd/ecl-event-catalog",
		Source: Source{
			Archive: archiveName,
			SHA256:  digest(archiveData),
		},
		Limitations: []string{
			"TraceGraph follows statically visible GOTO/GOSUB and sequential fallthrough only.",
			"IF conditions, dynamic ON GOTO/ON GOSUB destinations, menus, CALL targets, and runtime state are not executed.",
			"ordered_effect_candidates are straight-line static candidates, not proof of original runtime order or side effects.",
		},
		Summary: Summary{OpcodeCounts: make(map[string]int)},
	}

	files := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		files[file.Name] = file
	}
	for _, name := range memberNames {
		file := files[name]
		if file == nil {
			return Catalog{}, fmt.Errorf("%s is absent from archive", name)
		}
		data, err := readZipFile(file)
		if err != nil {
			return Catalog{}, fmt.Errorf("read %s: %w", name, err)
		}
		member, err := buildMember(name, data, &catalog.Summary)
		if err != nil {
			return Catalog{}, err
		}
		catalog.Members = append(catalog.Members, member)
	}
	catalog.Summary.MemberCount = len(catalog.Members)
	return catalog, nil
}

func Encode(catalog Catalog) ([]byte, error) {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// EncodeMarkdown produces the human-reviewable companion to the complete JSON
// artifact. It intentionally lists candidates rather than claiming runtime
// order or semantic closure.
func EncodeMarkdown(catalog Catalog) []byte {
	var output strings.Builder
	fmt.Fprintln(&output, "# CoAB ECL 全事件靜態清冊摘要")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "> 本檔由 `cmd/ecl-event-catalog` 產生；請勿手動編輯。完整機器資料見")
	fmt.Fprintln(&output, "> [`ecl-event-catalog.json`](ecl-event-catalog.json)。")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## 證據邊界")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "- 原始 archive：`%s`\n", catalog.Source.Archive)
	fmt.Fprintf(&output, "- SHA-256：`%s`\n", catalog.Source.SHA256)
	fmt.Fprintln(&output, "- 位址空間：每個 block 解碼 payload offset；`code_address=0x8000+offset`。")
	fmt.Fprintln(&output, "- 推論等級：靜態 framing／直接 GOTO／GOSUB 可達性為 `exact`；effect kind 與直線序列是 audit 分類候選，不是原版 runtime 語意。")
	for _, limitation := range catalog.Limitations {
		fmt.Fprintf(&output, "- 限制：%s\n", limitation)
	}
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## 摘要")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "| 項目 | 數量 |")
	fmt.Fprintln(&output, "|---|---:|")
	fmt.Fprintf(&output, "| ECL DAX member | %d |\n", catalog.Summary.MemberCount)
	fmt.Fprintf(&output, "| block | %d |\n", catalog.Summary.BlockCount)
	fmt.Fprintf(&output, "| lifecycle entry | %d |\n", catalog.Summary.LifecycleEntryCount)
	fmt.Fprintf(&output, "| 不重複靜態可達 instruction | %d |\n", catalog.Summary.UniqueReachableInstructionCount)
	fmt.Fprintf(&output, "| 跨 effect-kind 直線候選 | %d |\n", catalog.Summary.OrderedEffectCandidateCount)
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## Block 與 lifecycle entry")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "| Member | SHA-256 | Block | decoded bytes | lifecycle offsets | instructions | candidates |")
	fmt.Fprintln(&output, "|---|---|---:|---:|---|---:|---:|")
	for _, member := range catalog.Members {
		for _, block := range member.Blocks {
			entries := make([]string, 0, len(block.LifecycleEntries))
			for _, entry := range block.LifecycleEntries {
				entries = append(entries, fmt.Sprintf("%s=%s", entry.Name, entry.Offset))
			}
			fmt.Fprintf(&output, "| `%s` | `%s` | `%s` | %d | %s | %d | %d |\n",
				member.Name, member.SHA256, block.ID, block.DecodedSize,
				strings.Join(entries, "<br>"), len(block.Instructions), len(block.OrderedEffectCandidates))
		}
	}
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## 跨類型副作用序列候選")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "以下每列只證明同一靜態直線區段可看見不同 effect kind。IF、動態選單、")
	fmt.Fprintln(&output, "CALL consumer、pause commit 與實際分支仍須逐列回到原始 bytes、IDA 與 runtime 閉合。")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "| Member／Block | 範圍 | lifecycle | 靜態 effect sequence | 證據 |")
	fmt.Fprintln(&output, "|---|---|---|---|---|")
	for _, member := range catalog.Members {
		for _, block := range member.Blocks {
			for _, candidate := range block.OrderedEffectCandidates {
				effects := make([]string, 0, len(candidate.Effects))
				for _, effect := range candidate.Effects {
					effects = append(effects, fmt.Sprintf("`%s %s [%s]`", effect.Offset, effect.Name, strings.Join(effect.Kinds, "+")))
				}
				fmt.Fprintf(&output, "| `%s/%s` | `%s..%s` | %s | %s | `%s` |\n",
					member.Name, block.ID, candidate.StartOffset, candidate.EndOffset,
					strings.Join(candidate.ReachableFrom, ", "), strings.Join(effects, " → "), candidate.Evidence)
			}
		}
	}
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "## 使用方式")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "1. 先以 JSON 的 member／block／offset 回讀原始 bytes 與完整 instruction operands。")
	fmt.Fprintln(&output, "2. 對候選建立 `re-closure-record-template.md`，補 writer→projection→consumer。")
	fmt.Fprintln(&output, "3. 以原版 runtime 或等價第一級證據確認實際分支與 commit phase。")
	fmt.Fprintln(&output, "4. 完成 ordered event contract 後，才將候選升格成 READY 規格與 engine transaction。")
	return []byte(output.String())
}

func buildMember(name string, data []byte, summary *Summary) (Member, error) {
	blocks, err := dax.Parse(data)
	if err != nil {
		return Member{}, fmt.Errorf("parse %s: %w", name, err)
	}
	member := Member{Name: name, SHA256: digest(data)}
	for _, rawBlock := range blocks {
		block, err := buildBlock(name, rawBlock, summary)
		if err != nil {
			return Member{}, err
		}
		member.Blocks = append(member.Blocks, block)
	}
	return member, nil
}

func buildBlock(memberName string, rawBlock dax.Block, summary *Summary) (Block, error) {
	points, _, err := ecl.EntryPoints(rawBlock.Data, len(lifecycleNames))
	if err != nil {
		return Block{}, fmt.Errorf("%s block 0x%02X entry points: %w", memberName, rawBlock.Entry.ID, err)
	}
	block := Block{ID: hexByte(rawBlock.Entry.ID), DecodedSize: len(rawBlock.Data)}
	instructions := make(map[int]*aggregateInstruction)
	edges := make(map[string]*aggregateEdge)
	candidates := make(map[string]*aggregateCandidate)
	for index, point := range points {
		entryName := lifecycleNames[index]
		start := int(point) - ecl.CodeAddressBase
		block.LifecycleEntries = append(block.LifecycleEntries, LifecycleEntry{
			Index:       index,
			Name:        entryName,
			CodeAddress: hexWord(point),
			Offset:      hexOffset(start),
		})
		graph, err := ecl.TraceGraph(rawBlock.Data, []int{start}, len(rawBlock.Data)*8)
		if err != nil {
			return Block{}, fmt.Errorf("%s block 0x%02X %s graph: %w", memberName, rawBlock.Entry.ID, entryName, err)
		}
		for _, instruction := range graph.Instructions {
			record := instructions[instruction.Offset]
			if record == nil {
				record = &aggregateInstruction{value: instruction, from: make(map[string]bool)}
				instructions[instruction.Offset] = record
			}
			record.from[entryName] = true
		}
		for _, edge := range graph.Edges {
			key := fmt.Sprintf("%04X:%04X:%s", edge.From, edge.To, edge.Kind)
			record := edges[key]
			if record == nil {
				record = &aggregateEdge{value: edge, from: make(map[string]bool)}
				edges[key] = record
			}
			record.from[entryName] = true
		}
		for _, candidate := range straightLineCandidates(graph.Instructions, entryName) {
			key := candidateKey(candidate)
			record := candidates[key]
			if record == nil {
				record = &aggregateCandidate{value: candidate, from: make(map[string]bool)}
				candidates[key] = record
			}
			record.from[entryName] = true
		}
	}

	offsets := make([]int, 0, len(instructions))
	for offset := range instructions {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)
	for _, offset := range offsets {
		record := instructions[offset]
		instruction := exportInstruction(record.value, sortedKeys(record.from))
		block.Instructions = append(block.Instructions, instruction)
		summary.OpcodeCounts[instruction.Opcode]++
	}

	edgeKeys := sortedAggregateKeys(edges)
	for _, key := range edgeKeys {
		record := edges[key]
		block.Edges = append(block.Edges, Edge{
			From:          hexOffset(record.value.From),
			To:            hexOffset(record.value.To),
			Kind:          record.value.Kind,
			ReachableFrom: sortedKeys(record.from),
		})
	}

	candidateKeys := sortedAggregateKeys(candidates)
	for _, key := range candidateKeys {
		record := candidates[key]
		candidate := record.value
		candidate.ReachableFrom = sortedKeys(record.from)
		block.OrderedEffectCandidates = append(block.OrderedEffectCandidates, candidate)
	}

	summary.BlockCount++
	summary.LifecycleEntryCount += len(block.LifecycleEntries)
	summary.UniqueReachableInstructionCount += len(block.Instructions)
	summary.OrderedEffectCandidateCount += len(block.OrderedEffectCandidates)
	return block, nil
}

func straightLineCandidates(input []ecl.Instruction, entryName string) []OrderedEffectCandidate {
	instructions := append([]ecl.Instruction(nil), input...)
	sort.Slice(instructions, func(i, j int) bool { return instructions[i].Offset < instructions[j].Offset })
	var output []OrderedEffectCandidate
	for start := 0; start < len(instructions); {
		end := start + 1
		for end < len(instructions) && instructions[end-1].Next == instructions[end].Offset &&
			!endsStraightLine(instructions[end-1].Command.Opcode) {
			end++
		}
		var effects []CandidateEffect
		kindSet := make(map[string]bool)
		for _, instruction := range instructions[start:end] {
			kinds := effectKinds(instruction.Command.Opcode)
			if len(kinds) == 0 {
				continue
			}
			for _, kind := range kinds {
				kindSet[kind] = true
			}
			effects = append(effects, CandidateEffect{
				Offset: hexOffset(instruction.Offset), Opcode: hexByte(instruction.Command.Opcode),
				Name: instruction.Command.Name, Kinds: kinds,
			})
		}
		if len(effects) >= 2 && len(kindSet) >= 2 {
			output = append(output, OrderedEffectCandidate{
				StartOffset: hexOffset(instructions[start].Offset),
				EndOffset:   hexOffset(instructions[end-1].Next),
				Effects:     effects, ReachableFrom: []string{entryName},
				Evidence: "static_straight_line_candidate",
			})
		}
		start = end
	}
	return output
}

func exportInstruction(value ecl.Instruction, from []string) Instruction {
	output := Instruction{
		Offset: hexOffset(value.Offset), CodeAddress: hexWord(uint16(ecl.CodeAddressBase + value.Offset)),
		Opcode: hexByte(value.Command.Opcode), Name: value.Command.Name,
		NextOffset: hexOffset(value.Next), EffectKinds: effectKinds(value.Command.Opcode),
		ReachableFrom: from,
	}
	for _, operand := range value.Operands {
		record := Operand{Code: hexByte(operand.Code), Low: hexByte(operand.Low)}
		if operand.WordSet {
			record.High = hexByte(operand.High)
			record.Word = hexWord(operand.Word)
		}
		if len(operand.Packed) != 0 {
			record.PackedLength = len(operand.Packed)
			record.PackedSHA256 = digest(operand.Packed)
		}
		output.Operands = append(output.Operands, record)
	}
	return output
}

func effectKinds(opcode byte) []string {
	switch opcode {
	case 0x0A:
		return []string{"party_load"}
	case 0x0B, 0x0C, 0x1C:
		return []string{"combat_setup"}
	case 0x0D, 0x0E, 0x31, 0x3A, 0x3D:
		return []string{"presentation"}
	case 0x0F, 0x10, 0x15, 0x2B, 0x33, 0x39:
		return []string{"interaction_boundary"}
	case 0x11, 0x12:
		return []string{"text"}
	case 0x20:
		return []string{"ecl_transition"}
	case 0x21, 0x37:
		return []string{"resource_load"}
	case 0x22, 0x23:
		return []string{"combat_setup", "surprise"}
	case 0x24:
		return []string{"combat_boundary"}
	case 0x27:
		return []string{"treasure_boundary"}
	case 0x28, 0x32, 0x3F, 0x40:
		return []string{"inventory"}
	case 0x29, 0x2C:
		return []string{"interaction_boundary", "encounter"}
	case 0x2D:
		return []string{"external_call"}
	case 0x2E:
		return []string{"damage"}
	case 0x34:
		return []string{"clock"}
	case 0x36, 0x3E:
		return []string{"party_mutation"}
	case 0x38:
		return []string{"program_boundary"}
	case 0x3B, 0x3C:
		return []string{"spell_effect"}
	default:
		return nil
	}
}

func endsStraightLine(opcode byte) bool {
	switch opcode {
	case 0x00, 0x01, 0x02, 0x13, 0x25, 0x26:
		return true
	default:
		return false
	}
}

func candidateKey(candidate OrderedEffectCandidate) string {
	data, _ := json.Marshal(struct {
		Start   string            `json:"start"`
		End     string            `json:"end"`
		Effects []CandidateEffect `json:"effects"`
	}{candidate.StartOffset, candidate.EndOffset, candidate.Effects})
	return string(data)
}

func sortedKeys(values map[string]bool) []string {
	output := make([]string, 0, len(values))
	for value := range values {
		output = append(output, value)
	}
	sort.Strings(output)
	return output
}

func sortedAggregateKeys[T any](values map[string]*T) []string {
	output := make([]string, 0, len(values))
	for value := range values {
		output = append(output, value)
	}
	sort.Strings(output)
	return output
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hexByte(value byte) string   { return fmt.Sprintf("0x%02X", value) }
func hexWord(value uint16) string { return fmt.Sprintf("0x%04X", value) }
func hexOffset(value int) string  { return fmt.Sprintf("0x%04X", value) }
