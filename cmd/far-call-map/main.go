// Command far-call-map 把 overlay 裡的每一個 far call 解析成
// `overlay-NN entry#K`。
//
// ★ 為什麼需要這支。 IDA 把 overlay 的 `.bin` 從位址 0 載入，所以
// `call far ptr 014D:00B6` 這種指令會被它「解析」成同一個 `.bin` 裡的
// `loc_1584+2`——那個標籤指到的是**這個 overlay 自己的位元組**，
// 與真正的目標毫無關係。規格裡那一票 `<far 1590h>`、`<far 145Ah+3>`
// 就是這樣來的：名字看起來像位址，其實是工具的誤解。
//
// 真正的目標是常駐段裡的 **TPOV 控制記錄**：
//
//	段 → 控制記錄（每個 overlay 一筆，記錄裡 +20h 起是 entry stub 表）
//	位移 → (位移 − 20h) / 5 ＝ entry 編號（stub 每筆 5 bytes）
//
// 段與檔案位移之間差一個常數（EXE header 的段數）。**這支不寫死那個常數**，
// 而是掃過所有 far call、對每個候選 delta 數「有幾個目標正好落在 stub 邊界」，
// 取命中率最高的那個並把分佈印出來。猜錯的 delta 會讓命中率掉到接近 0，
// 而不是安靜地產生一份看起來合理的對照表。
//
// 用法：
//
//	./tools/go.sh run ./cmd/far-call-map -platform dos -output docs/audit/far-call-map.md
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// stubTableOffset 是控制記錄裡 entry stub 表的起點。
	stubTableOffset = 0x20
	// stubSize 是一筆 stub 的長度：`CD 3F`、handler u16、flags u8。
	stubSize = 5
	// maxDelta 是搜尋段位移常數的上界（段數，不是位元組）。
	maxDelta = 0x400
)

type manifest struct {
	Platform string `json:"platform"`
	Overlays []struct {
		Index      int    `json:"index"`
		Module     string `json:"module"`
		EntryCount int    `json:"entry_count"`
		Entries    []struct {
			Index            int `json:"index"`
			ExecutableOffset int `json:"executable_offset"`
			StubOffset       int `json:"stub_offset"`
			CodeOffset       int `json:"code_offset"`
		} `json:"entries"`
	} `json:"overlays"`
}

type bodyFile struct {
	Functions []struct {
		EA    int    `json:"ea"`
		Name  string `json:"name"`
		Items []struct {
			EA     int    `json:"ea"`
			Bytes  string `json:"bytes"`
			Disasm string `json:"disasm"`
		} `json:"items"`
	} `json:"functions"`
}

// record 是一個 overlay 的控制記錄在**檔案**裡的段位址與 entry 數。
type record struct {
	Module     string
	Index      int
	FilePara   int
	EntryCount int
	CodeOffset map[int]int
}

// callSite 是一個 far call 指令。
type callSite struct {
	Module   string
	Function string
	EA       int
	Segment  int
	Offset   int
}

// resolved 是解出來的目標。
type resolved struct {
	Module     string `json:"module"`
	Function   string `json:"function"`
	EA         string `json:"ea"`
	Target     string `json:"target"`
	Entry      int    `json:"entry"`
	CodeOffset string `json:"code_offset,omitempty"`
	Raw        string `json:"raw"`
}

type report struct {
	Platform     string     `json:"platform"`
	SegmentDelta int        `json:"segment_delta_paragraphs"`
	FarCalls     int        `json:"far_calls"`
	Resolved     int        `json:"resolved"`
	Resident     int        `json:"resident"`
	Unresolved   int        `json:"unresolved"`
	Targets      []resolved `json:"targets"`
	// EntryUsage 是每個 entry 被叫過幾次，用來回答「哪些 entry 沒有呼叫端」。
	EntryUsage map[string]int `json:"entry_usage"`
}

func main() {
	platform := flag.String("platform", "dos", "dos or pc98")
	root := flag.String("root", "workplace/re-sweep", tooltext.Text("h.b0a80c708155"))
	output := flag.String("output", "", tooltext.Text("h.fff5cb9e9bc2"))
	outputJSON := flag.String("json", "", tooltext.Text("h.7299c956bbb9"))
	flag.Parse()

	base := filepath.Join(*root, *platform)
	var loaded manifest
	if err := readJSON(filepath.Join(base, "ovr-manifest.json"), &loaded); err != nil {
		log.Fatal(err)
	}
	records := make([]record, 0, len(loaded.Overlays))
	for _, overlay := range loaded.Overlays {
		entry := record{
			Module: overlay.Module, Index: overlay.Index,
			EntryCount: overlay.EntryCount, CodeOffset: map[int]int{},
		}
		for _, stub := range overlay.Entries {
			if stub.Index == 0 {
				entry.FilePara = (stub.ExecutableOffset - stubTableOffset) >> 4
			}
			entry.CodeOffset[stub.Index] = stub.CodeOffset
		}
		if len(overlay.Entries) == 0 {
			continue
		}
		records = append(records, entry)
	}

	calls, err := collectFarCalls(filepath.Join(base, "overlays", "full"), *platform)
	if err != nil {
		log.Fatal(err)
	}
	if len(calls) == 0 {
		log.Fatal(tooltext.Text("h.f7cbe32a3373"))
	}

	delta, hits := bestDelta(calls, records)
	if hits == 0 {
		log.Fatal(tooltext.Text("h.e0a80a0db55c"))
	}

	result := report{
		Platform: *platform, SegmentDelta: delta,
		FarCalls: len(calls), EntryUsage: map[string]int{},
	}
	byPara := map[int]record{}
	for _, entry := range records {
		byPara[entry.FilePara-delta] = entry
	}
	for _, call := range calls {
		item := resolved{
			Module: call.Module, Function: call.Function,
			EA:  fmt.Sprintf("%04Xh", call.EA),
			Raw: fmt.Sprintf("%04X:%04X", call.Segment, call.Offset),
		}
		owner, ok := byPara[call.Segment]
		if !ok {
			// 段不是任何 overlay 的控制記錄 ⇒ 常駐程式碼裡的 far call。
			item.Target = "resident"
			item.Entry = -1
			result.Resident++
			result.Targets = append(result.Targets, item)
			continue
		}
		index, exact := entryIndex(call.Offset)
		if !exact || index >= owner.EntryCount {
			item.Target = owner.Module
			item.Entry = -1
			result.Unresolved++
			result.Targets = append(result.Targets, item)
			continue
		}
		item.Target = owner.Module
		item.Entry = index
		if code, found := owner.CodeOffset[index]; found {
			item.CodeOffset = fmt.Sprintf("%04Xh", code)
		}
		result.Resolved++
		result.EntryUsage[fmt.Sprintf("%s#%d", owner.Module, index)]++
		result.Targets = append(result.Targets, item)
	}
	sort.Slice(result.Targets, func(i, j int) bool {
		if result.Targets[i].Module != result.Targets[j].Module {
			return result.Targets[i].Module < result.Targets[j].Module
		}
		return result.Targets[i].EA < result.Targets[j].EA
	})

	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	if *output != "" {
		if err := os.WriteFile(*output, renderMarkdown(result, records, delta), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintf(os.Stderr, "platform=%s delta=%#x far_calls=%d resolved=%d resident=%d unresolved=%d\n",
		result.Platform, delta, result.FarCalls, result.Resolved, result.Resident, result.Unresolved)
}

// entryIndex 把 stub 表裡的位移換成 entry 編號。位移沒有落在 stub 邊界上
// 就回 false——**不要四捨五入到最近的 entry**，那會把「假設錯了」變成
// 一份看起來合理的對照表。
func entryIndex(offset int) (int, bool) {
	if offset < stubTableOffset {
		return 0, false
	}
	shifted := offset - stubTableOffset
	if shifted%stubSize != 0 {
		return 0, false
	}
	return shifted / stubSize, true
}

// bestDelta 掃過所有候選段位移，取「落在 stub 邊界且 entry 編號在範圍內」
// 最多的那個。正確的 delta 命中率會遠高於其他值；接近的假值只會命中零星幾筆。
func bestDelta(calls []callSite, records []record) (int, int) {
	paraCount := map[int]int{}
	for _, call := range calls {
		paraCount[call.Segment]++
	}
	bestValue, bestHits := 0, 0
	for delta := 0; delta <= maxDelta; delta++ {
		byPara := map[int]record{}
		for _, entry := range records {
			byPara[entry.FilePara-delta] = entry
		}
		hits := 0
		for _, call := range calls {
			owner, ok := byPara[call.Segment]
			if !ok {
				continue
			}
			index, exact := entryIndex(call.Offset)
			if exact && index < owner.EntryCount {
				hits++
			}
		}
		if hits > bestHits {
			bestValue, bestHits = delta, hits
		}
	}
	return bestValue, bestHits
}

// collectFarCalls 從 full body dump 抓出每一個 `9A` 開頭的 far call。
// 用**位元組**判斷而不是 disasm 字串：IDA 印出來的標籤正是要被取代的東西。
func collectFarCalls(dir, platform string) ([]callSite, error) {
	names, err := filepath.Glob(filepath.Join(dir, platform+"-overlay-*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var calls []callSite
	for _, name := range names {
		var body bodyFile
		if err := readJSON(name, &body); err != nil {
			return nil, err
		}
		module := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(name), platform+"-"), ".json")
		for _, function := range body.Functions {
			for _, item := range function.Items {
				raw, err := hex.DecodeString(item.Bytes)
				if err != nil || len(raw) != 5 || raw[0] != 0x9A {
					continue
				}
				calls = append(calls, callSite{
					Module:   module,
					Function: function.Name,
					EA:       item.EA,
					Offset:   int(binary.LittleEndian.Uint16(raw[1:3])),
					Segment:  int(binary.LittleEndian.Uint16(raw[3:5])),
				})
			}
		}
	}
	return calls, nil
}

func renderMarkdown(result report, records []record, delta int) []byte {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.459f6021ab07"))
	out.WriteString(tooltext.Text("h.875d85c8c66e"))
	fmt.Fprint(&out, tooltext.Text("h.19d06c7650bd")+
		tooltext.Text("h.7ef00e2f3772")+
		tooltext.Text("h.6668473d970d"))
	fmt.Fprintf(&out, tooltext.Text("h.788eb6848b9b")+
		tooltext.Text("h.d281e7784702")+
		tooltext.Text("h.28189a5481ff")+
		tooltext.Text("h.40e2168246c1"), delta)
	fmt.Fprint(&out, tooltext.Format("h.13c83a8a875e"))
	fmt.Fprint(&out, tooltext.Format("h.a30ac1014cea", result.FarCalls))
	fmt.Fprint(&out, tooltext.Format("h.356307303e23", result.Resolved))
	fmt.Fprint(&out, tooltext.Format("h.8144948a5436", result.Resident))
	fmt.Fprint(&out, tooltext.Format("h.2f7c1bfc04a4", result.Unresolved))

	out.WriteString(tooltext.Text("h.a26541153689"))
	out.WriteString(tooltext.Text("h.459d83222c48"))
	out.WriteString(tooltext.Text("h.692e2b57286c"))
	sorted := append([]record(nil), records...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Index < sorted[j].Index })
	for _, entry := range sorted {
		used := 0
		for index := 0; index < entry.EntryCount; index++ {
			if result.EntryUsage[fmt.Sprintf("%s#%d", entry.Module, index)] > 0 {
				used++
			}
		}
		fmt.Fprintf(&out, "| `%s` | %d | %d | %d |\n",
			entry.Module, entry.EntryCount, used, entry.EntryCount-used)
	}

	out.WriteString(tooltext.Text("h.76f00c64beb8"))
	out.WriteString(tooltext.Text("h.ea44e542857c") +
		tooltext.Text("h.c881c758a4ea") +
		tooltext.Text("h.b198af74b85f"))
	callers := map[string][]string{}
	for _, item := range result.Targets {
		if item.Entry < 0 {
			continue
		}
		key := fmt.Sprintf("%s entry#%d", item.Target, item.Entry)
		label := fmt.Sprintf("%s %s", item.Module, item.Function)
		if len(callers[key]) == 0 || callers[key][len(callers[key])-1] != label {
			callers[key] = append(callers[key], label)
		}
	}
	keys := make([]string, 0, len(callers))
	for key := range callers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return entrySortKey(keys[i]) < entrySortKey(keys[j])
	})
	out.WriteString(tooltext.Text("h.aea4f359fd3a"))
	for _, key := range keys {
		unique := uniqueStrings(callers[key])
		code := ""
		for _, item := range result.Targets {
			if item.Entry >= 0 && fmt.Sprintf("%s entry#%d", item.Target, item.Entry) == key {
				code = item.CodeOffset
				break
			}
		}
		fmt.Fprintf(&out, "| `%s` | `%s` | %s |\n", key, code, strings.Join(unique, "、"))
	}
	return []byte(out.String())
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// entrySortKey 讓 `overlay-9 entry#12` 排在 `overlay-10 entry#2` 前面，
// 並且同一個 overlay 內依 entry 編號排——字串排序會把 10 排在 2 前面。
func entrySortKey(key string) string {
	var module, entry int
	if _, err := fmt.Sscanf(key, "overlay-%d entry#%d", &module, &entry); err != nil {
		return key
	}
	return fmt.Sprintf("%03d/%03d", module, entry)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
