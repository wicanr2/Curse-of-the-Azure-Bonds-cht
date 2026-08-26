// Command re-gap-audit 對兩平台 overlay 的未定義區段（RE-11）做機械分類。
//
// 分母＝IDA 匯出的 `undefined_ranges`（spec 559 全模組掃描的副產物）。
// 分類只看位元組形狀：碎屑（≤2B 的對齊縫）、同值填充、文字（可列印比率
// 或 Pascal 長度前綴鏈）；其餘標「待人工」並附前 16 bytes——**那一堆才是
// RE-11 還沒回答的部分**，逐段判定屬於後續輪。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

type idaExport struct {
	Overlay struct {
		Module string `json:"module"`
	} `json:"overlay"`
	UndefinedRanges []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"undefined_ranges"`
}

type gap struct {
	Platform string `json:"platform"`
	Module   string `json:"module"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Size     int    `json:"size"`
	Class    string `json:"class"`
	// FillByte 只在同值填充時有意義。
	FillByte byte   `json:"fill_byte,omitempty"`
	Preview  string `json:"preview,omitempty"`
	// LedgerState／LedgerFunction：待人工段完整落在台帳某函式範圍內時，
	// 記下那支函式的解讀狀態與名字。「已解讀」「邊界碎片」代表那幾個
	// bytes 其實已被人讀過（IDA 匯出把函式中段標成 undefined 而已），
	// 不是漏碼——overlay-14 0044h 那 258 bytes 就是這一種。
	LedgerState    string `json:"ledger_state,omitempty"`
	LedgerFunction string `json:"ledger_function,omitempty"`
}

// ledgerFunc 是 coab-function-index.json 的一列（只取需要的欄位）。
type ledgerFunc struct {
	Platform string `json:"platform"`
	Module   string `json:"module"`
	EA       int    `json:"ea"`
	Size     int    `json:"size"`
	Name     string `json:"ida_name"`
	State    string `json:"state"`
}

const (
	classCrumb   = "crumb"
	classFill    = "fill"
	classText    = "text"
	classPending = "pending"
)

func classify(data []byte) (string, byte) {
	if len(data) <= 2 {
		return classCrumb, 0
	}
	same := true
	for _, value := range data {
		if value != data[0] {
			same = false
			break
		}
	}
	if same {
		return classFill, data[0]
	}
	if textCoverage(data)*100 >= len(data)*85 {
		return classText, 0
	}
	// Pascal 長度前綴鏈：len byte ＋ len 個文字位元組，反覆到結尾。
	covered, index := 0, 0
	for index < len(data) {
		length := int(data[index])
		if length == 0 || index+1+length > len(data) {
			break
		}
		if textCoverage(data[index+1:index+1+length]) != length {
			break
		}
		covered += 1 + length
		index += 1 + length
	}
	if covered*100 >= len(data)*80 {
		return classText, 0
	}
	return classPending, 0
}

// textCoverage 回傳 data 裡「像文字」的位元組數：可列印 ASCII 算 1、
// 合法 Shift-JIS 雙位元組（PC-98 版訊息的編碼）整對算 2。
// 這是形狀判定，不是解碼——單獨的高位元組不計入。
func textCoverage(data []byte) int {
	covered, index := 0, 0
	for index < len(data) {
		value := data[index]
		if value >= 0x20 && value < 0x7F {
			covered++
			index++
			continue
		}
		if (value >= 0x81 && value <= 0x9F || value >= 0xE0 && value <= 0xEF) &&
			index+1 < len(data) {
			trail := data[index+1]
			if trail >= 0x40 && trail <= 0xFC && trail != 0x7F {
				covered += 2
				index += 2
				continue
			}
		}
		index++
	}
	return covered
}

func main() {
	sweep := flag.String("sweep", "workplace/re-sweep", tooltext.Text("re_gap_audit.usage_sweep"))
	jsonPath := flag.String("json", "docs/audit/re-gap-audit.json", tooltext.Text("re_gap_audit.usage_json"))
	mdPath := flag.String("md", "docs/audit/re-gap-audit.md", tooltext.Text("re_gap_audit.usage_md"))
	pendingMin := flag.Int("pending-min", 16, tooltext.Text("re_gap_audit.usage_pending_min"))
	ledgerPath := flag.String("ledger", "docs/audit/coab-function-index.json",
		tooltext.Text("re_gap_audit.usage_ledger"))
	flag.Parse()

	ledger := map[string][]ledgerFunc{}
	if raw, err := os.ReadFile(*ledgerPath); err == nil {
		var parsed struct {
			Functions []ledgerFunc `json:"functions"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			log.Fatalf("%s: %v", *ledgerPath, err)
		}
		for _, fn := range parsed.Functions {
			key := fn.Platform + "/" + fn.Module
			ledger[key] = append(ledger[key], fn)
		}
	} else {
		log.Printf("ledger unavailable (%v); pending gaps will not be cross-checked", err)
	}

	gaps := []gap{}
	for _, platform := range []string{"dos", "pc98"} {
		// 常駐檔匯出（START.EXE.json／PC98-GAME.EXE.json）undefined_ranges 皆為 0，
		// 且 bytes 不在 overlays/ 底下，分母只收 overlay。
		paths, err := filepath.Glob(filepath.Join(*sweep, platform, "out", "overlay-*.json"))
		if err != nil || len(paths) == 0 {
			log.Fatalf("no IDA exports under %s/%s/out", *sweep, platform)
		}
		sort.Strings(paths)
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			export := idaExport{}
			if err := json.Unmarshal(raw, &export); err != nil {
				log.Fatalf("%s: %v", path, err)
			}
			module := export.Overlay.Module
			if module == "" {
				module = strings.TrimSuffix(filepath.Base(path), ".json")
			}
			code, err := os.ReadFile(filepath.Join(*sweep, platform, "overlays", module+".bin"))
			if err != nil {
				log.Fatalf("%s code: %v", module, err)
			}
			for _, entry := range export.UndefinedRanges {
				if entry.End <= entry.Start || entry.End > len(code) {
					log.Fatalf("%s/%s range %04X..%04X is outside %d-byte code",
						platform, module, entry.Start, entry.End, len(code))
				}
				data := code[entry.Start:entry.End]
				class, fill := classify(data)
				preview := ""
				if class == classPending {
					limit := len(data)
					if limit > 16 {
						limit = 16
					}
					preview = fmt.Sprintf("% X", data[:limit])
				}
				record := gap{Platform: platform, Module: module,
					Start: entry.Start, End: entry.End, Size: len(data),
					Class: class, FillByte: fill, Preview: preview}
				if class == classPending {
					for _, fn := range ledger[platform+"/"+module] {
						if fn.EA <= entry.Start && entry.End <= fn.EA+fn.Size {
							record.LedgerState = fn.State
							record.LedgerFunction = fn.Name
							break
						}
					}
				}
				gaps = append(gaps, record)
			}
		}
	}

	perPlatform := map[string]*stats{}
	perModule := map[moduleKey]*stats{}
	for _, entry := range gaps {
		for _, bucket := range []*stats{
			ensure(perPlatform, entry.Platform),
			ensureModule(perModule, moduleKey{entry.Platform, entry.Module}),
		} {
			bucket.ranges++
			bucket.bytes += entry.Size
			switch entry.Class {
			case classCrumb:
				bucket.crumb++
			case classFill:
				bucket.fill++
			case classText:
				bucket.text++
			case classPending:
				bucket.pending++
				bucket.pendingBytes += entry.Size
				if entry.LedgerState == "" {
					bucket.pendingOutside++
					bucket.pendingOutsideBytes += entry.Size
				} else {
					bucket.pendingRead++
					bucket.pendingReadBytes += entry.Size
				}
			}
		}
	}

	encoded, err := json.MarshalIndent(map[string]any{
		"schema": "coab-re-gap-audit/1", "gaps": gaps,
	}, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*jsonPath, append(encoded, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}

	var md strings.Builder
	md.WriteString(tooltext.Text("re_gap_audit.md_header"))
	md.WriteString(tooltext.Text("re_gap_audit.md_class_header"))
	for _, platform := range []string{"dos", "pc98"} {
		bucket := perPlatform[platform]
		fmt.Fprintf(&md, "| %s | %d | %d | %d | %d | %d | %d／%d | %d／%d | **%d／%d** |\n",
			platform, bucket.ranges, bucket.bytes, bucket.crumb, bucket.fill,
			bucket.text, bucket.pending, bucket.pendingBytes,
			bucket.pendingRead, bucket.pendingReadBytes,
			bucket.pendingOutside, bucket.pendingOutsideBytes)
	}
	md.WriteString(tooltext.Text("re_gap_audit.md_module_header"))
	moduleKeys := make([]moduleKey, 0, len(perModule))
	for key := range perModule {
		moduleKeys = append(moduleKeys, key)
	}
	sort.Slice(moduleKeys, func(i, j int) bool {
		if moduleKeys[i].platform != moduleKeys[j].platform {
			return moduleKeys[i].platform < moduleKeys[j].platform
		}
		return moduleKeys[i].module < moduleKeys[j].module
	})
	for _, key := range moduleKeys {
		bucket := perModule[key]
		if bucket.pending == 0 {
			continue
		}
		largest := 0
		largestAt := ""
		for _, entry := range gaps {
			if entry.Platform == key.platform && entry.Module == key.module &&
				entry.Class == classPending && entry.Size > largest {
				largest = entry.Size
				largestAt = fmt.Sprintf("`%04X..%04X`", entry.Start, entry.End)
			}
		}
		fmt.Fprintf(&md, "| %s | `%s` | %d | %d | %s（%d B）|\n",
			key.platform, key.module, bucket.pending, bucket.pendingBytes, largestAt, largest)
	}
	fmt.Fprintf(&md, tooltext.Text("re_gap_audit.md_pending_header"), *pendingMin)
	pending := make([]gap, 0)
	for _, entry := range gaps {
		if entry.Class == classPending && entry.Size >= *pendingMin {
			pending = append(pending, entry)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		outsideI, outsideJ := pending[i].LedgerState == "", pending[j].LedgerState == ""
		if outsideI != outsideJ {
			return outsideI
		}
		return pending[i].Size > pending[j].Size
	})
	for _, entry := range pending {
		ledgerCell := tooltext.Text("re_gap_audit.outside_function")
		if entry.LedgerState != "" {
			ledgerCell = fmt.Sprintf("%s（`%s`）", entry.LedgerState, entry.LedgerFunction)
		}
		fmt.Fprintf(&md, "| %s | `%s` | `%04X..%04X` | %d | %s | `%s` |\n",
			entry.Platform, entry.Module, entry.Start, entry.End, entry.Size,
			ledgerCell, entry.Preview)
	}
	if err := os.WriteFile(*mdPath, []byte(md.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	for _, platform := range []string{"dos", "pc98"} {
		bucket := perPlatform[platform]
		fmt.Printf(tooltext.Text("re_gap_audit.summary"), platform,
			bucket.ranges, bucket.bytes, bucket.crumb, bucket.fill,
			bucket.text, bucket.pending, bucket.pendingBytes,
			bucket.pendingOutside, bucket.pendingOutsideBytes)
	}
}

func ensure(m map[string]*stats, key string) *stats {
	if m[key] == nil {
		m[key] = &stats{}
	}
	return m[key]
}

type stats struct {
	ranges, bytes, crumb, fill, text, pending, pendingBytes int
	pendingRead, pendingReadBytes                           int
	pendingOutside, pendingOutsideBytes                     int
}

func ensureModule(m map[moduleKey]*stats, key moduleKey) *stats {
	if m[key] == nil {
		m[key] = &stats{}
	}
	return m[key]
}

type moduleKey struct{ platform, module string }
