// Command save-field-coverage 量出 DOS 角色記錄裡**每一個位元組**的狀態，
// 並與 `internal/party` 的台帳對帳。
//
// ★ 為什麼用突變量測，而不是掃程式碼。 掃 `data[0x...]` 這種寫法會漏掉算出來的
// 索引（`data[0x12D+class*5]`），得到的「未讀」清單裡會有一堆假缺口——而假缺口
// 看起來就像真缺口。突變法不看程式碼長什麼樣：改一個位元組、重新解析、比對結果，
// **有差別就是有讀**。
//
// ⚠ 這是**下界**：只在特定職業／種族才會被讀的位元組，若基準記錄沒有涵蓋那個
// 條件就會被判成沒讀。所以基準記錄有好幾份（不同職業、不同種族、有無法術），
// 而且「台帳說 decoded 但一份都沒量到」會被當成錯誤——修法是補一份基準記錄，
// 不是把台帳改成 documented。
//
// 用法：
//
//	./tools/go.sh run ./cmd/save-field-coverage -output docs/audit/dos-save-field-coverage.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// mutations 是每個位元組要試的值。原值以外全試，讓「只在某些值下才有差別」的
// 欄位也露出來（例如只判 0／非 0 的旗標）。
var mutations = []byte{0x00, 0x01, 0x02, 0x7F, 0x80, 0xFF}

type fieldRow struct {
	Offset   string `json:"offset"`
	Size     int    `json:"size"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Spec     string `json:"spec,omitempty"`
	Consumed int    `json:"consumed_bytes"`
}

// sidecarReport 是 `.SWG`／`.FX` 兩份 sidecar 的台帳摘要。它們沒有突變探針
// ——兩者的解析器是逐欄直讀，沒有算出來的索引，台帳蓋滿就足以說明覆蓋。
type sidecarReport struct {
	Name     string         `json:"name"`
	Size     int            `json:"size"`
	ByStatus map[string]int `json:"bytes_by_status"`
	Fields   []fieldRow     `json:"fields"`
}

type report struct {
	Schema      string          `json:"schema"`
	RecordSize  int             `json:"record_size"`
	Baselines   int             `json:"baselines"`
	ByStatus    map[string]int  `json:"bytes_by_status"`
	ConsumedAll int             `json:"consumed_bytes"`
	Fields      []fieldRow      `json:"fields"`
	Mismatches  []string        `json:"mismatches"`
	Sidecars    []sidecarReport `json:"sidecars"`
}

func main() {
	output := flag.String("output", "", tooltext.Text("h.fff5cb9e9bc2"))
	outputJSON := flag.String("json", "", tooltext.Text("h.7299c956bbb9"))
	flag.Parse()

	result, err := analyze()
	if err != nil {
		log.Fatal(err)
	}

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
		if err := os.WriteFile(*output, renderMarkdown(result), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprint(os.Stderr, tooltext.Format("h.7a01b550fd79", result.RecordSize, result.ByStatus["decoded"], result.ByStatus["documented"], result.ByStatus["unknown"], result.ConsumedAll, len(result.Mismatches)))
	if len(result.Mismatches) > 0 {
		for _, line := range result.Mismatches {
			fmt.Fprintln(os.Stderr, "  ", line)
		}
		os.Exit(2)
	}
}

// analyze 量測並對帳。抽成函式讓 `main_test.go` 跑同一份邏輯——
// 報告只有人在看才有用，閘要在 `go test ./...` 裡。
func analyze() (report, error) {
	if err := party.ValidateDOSPlayerRecordFields(); err != nil {
		return report{}, err
	}
	baselines := baselineRecords()
	consumed := make([]bool, party.DOSPlayerRecordSize)
	for _, baseline := range baselines {
		base, baseErr := party.ParseOriginalDOSPlayerRecord(baseline.data, "probe")
		if baseErr != nil {
			return report{}, tooltext.Errorf("h.af0c63659919", baseline.name, baseErr)
		}
		for offset := 0; offset < party.DOSPlayerRecordSize; offset++ {
			if consumed[offset] {
				continue
			}
			for _, value := range mutations {
				if baseline.data[offset] == value {
					continue
				}
				mutated := append([]byte(nil), baseline.data...)
				mutated[offset] = value
				got, err := party.ParseOriginalDOSPlayerRecord(mutated, "probe")
				if err != nil || !reflect.DeepEqual(got, base) {
					consumed[offset] = true
					break
				}
			}
		}
	}

	result := report{
		Schema:     "coab-dos-save-field-coverage/1",
		RecordSize: party.DOSPlayerRecordSize,
		Baselines:  len(baselines),
		ByStatus:   map[string]int{},
	}
	for _, field := range party.DOSPlayerRecordFields {
		row := fieldRow{
			Offset: fmt.Sprintf("+%03Xh", field.Offset), Size: field.Size,
			Name: field.Name, Status: string(field.Status), Spec: field.Spec,
		}
		for offset := field.Offset; offset < field.Offset+field.Size; offset++ {
			if consumed[offset] {
				row.Consumed++
				result.ConsumedAll++
			}
		}
		result.ByStatus[string(field.Status)] += field.Size
		// 對帳：量到有讀，台帳卻沒說 decoded ⇒ 台帳低估了，硬錯。
		if row.Consumed > 0 && field.Status != party.DOSFieldDecoded {
			result.Mismatches = append(result.Mismatches, tooltext.Format("h.86de433776c8", field.Offset, field.Name, row.Consumed, field.Status))
		}
		// 台帳說 decoded，卻沒有任何基準記錄量到 ⇒ 補基準記錄，別改台帳。
		if row.Consumed == 0 && field.Status == party.DOSFieldDecoded {
			result.Mismatches = append(result.Mismatches, fmt.Sprintf(
				tooltext.Text("h.37066a8f88d7")+
					tooltext.Text("h.dcb099c34d2b"), field.Offset, field.Name))
		}
		result.Fields = append(result.Fields, row)
	}
	sort.Strings(result.Mismatches)

	for _, sidecar := range []struct {
		name   string
		fields []monster.RecordField
		size   int
	}{
		{tooltext.Text("h.aa095065b9d9"), monster.ItemRecordFields, monster.ItemRecordSize},
		{tooltext.Text("h.9a96e67a7e7b"), monster.AffectRecordFields, monster.AffectRecordSize},
	} {
		if err := monster.ValidateRecordFields(sidecar.name, sidecar.fields, sidecar.size); err != nil {
			return report{}, err
		}
		entry := sidecarReport{Name: sidecar.name, Size: sidecar.size, ByStatus: map[string]int{}}
		for _, field := range sidecar.fields {
			entry.ByStatus[string(field.Status)] += field.Size
			entry.Fields = append(entry.Fields, fieldRow{
				Offset: fmt.Sprintf("+%02Xh", field.Offset), Size: field.Size,
				Name: field.Name, Status: string(field.Status), Spec: field.Spec,
			})
		}
		result.Sidecars = append(result.Sidecars, entry)
	}
	return result, nil
}

type baseline struct {
	name string
	data []byte
}

// baselineRecords 是幾份能解析成功的角色記錄。**不同職業與種族要各一份**：
// 只在某個職業才被讀的位元組，基準沒涵蓋就會被判成沒讀。
func baselineRecords() []baseline {
	make := func(name string, race, classCombo byte, classSlot int) baseline {
		data := makeRecord(name, race, classCombo, classSlot)
		return baseline{name: name, data: data}
	}
	return []baseline{
		make("FIGHTER", 7, 2, 2),
		make("MAGE", 2, 5, 5),
		make("CLERIC", 7, 0, 0),
		make("THIEF", 5, 6, 6),
		make("MULTI", 2, 13, 2), // 戰士／法師：會走第二職業那條路
	}
}

// makeRecord 組一份最小可解析的記錄，其餘位元組故意填成非零的固定樣式——
// 全 0 的背景會讓「把某格改成 0」量不到差別。
func makeRecord(name string, race, classCombo byte, classSlot int) []byte {
	data := make([]byte, party.DOSPlayerRecordSize)
	for index := range data {
		data[index] = 0x33
	}
	data[0] = byte(len(name))
	copy(data[1:], name)
	for index := len(name) + 1; index < 0x10; index++ {
		data[index] = 0
	}
	// 屬性：基準與現值都填合法值。
	for offset := 0x10; offset <= 0x1B; offset++ {
		data[offset] = 12
	}
	data[0x1C], data[0x1D] = 0, 0
	// 記憶法術與已知法術用合法的法術編號（1..100）。
	for offset := 0x01E; offset < 0x072; offset++ {
		data[offset] = 0
	}
	data[0x01E], data[0x01F] = 1, 3
	for offset := 0x079; offset < 0x0DD; offset++ {
		data[offset] = 0
	}
	data[0x079], data[0x07A] = 1, 0
	data[0x074] = race
	data[0x075] = classCombo
	data[0x076], data[0x077] = 20, 0
	data[0x078] = 24
	data[0x109+classSlot] = 4
	data[0x0E6] = 4
	data[0x119] = 0
	data[0x11B] = 0
	data[0x1A4] = 20
	// 貨幣與經驗值留固定樣式即可（0x33 是合法值）。
	return data
}

func renderMarkdown(result report) []byte {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.900e8e3d36e1"))
	out.WriteString(tooltext.Text("h.961a46bf1039"))
	out.WriteString(tooltext.Text("h.5341046d8624") +
		tooltext.Text("h.b4046823dbc7"))
	out.WriteString(tooltext.Text("h.ae311b79f57a") +
		tooltext.Text("h.3a22d28bbf51"))
	out.WriteString(tooltext.Text("h.0805f40e9349") +
		tooltext.Text("h.4dadea83e94b") +
		tooltext.Text("h.d43ef8c03413"))
	out.WriteString(tooltext.Text("h.03dfffb541b0") +
		tooltext.Text("h.d5e41bb036a6"))
	out.WriteString(tooltext.Text("h.663d2c28008a") +
		tooltext.Text("h.d4697a03be54") +
		tooltext.Text("h.c2027ef6bbb6"))

	fmt.Fprint(&out, tooltext.Format("h.5a9885a1c455"))
	for _, status := range []string{"decoded", "documented", "unknown"} {
		count := result.ByStatus[status]
		fmt.Fprintf(&out, "| `%s` | %d | %d%% |\n", status, count, count*100/result.RecordSize)
	}
	fmt.Fprint(&out, tooltext.Format("h.7482263f1fa2", result.RecordSize))
	fmt.Fprint(&out, tooltext.Format("h.10ec9931c77e", result.Baselines, result.ConsumedAll, result.RecordSize))

	out.WriteString(tooltext.Text("h.e0eb54a1c9cc"))
	out.WriteString(tooltext.Text("h.693b6853416c") +
		tooltext.Text("h.0d1a8dd925d8"))
	for _, sidecar := range result.Sidecars {
		fmt.Fprintf(&out, "### %s（%d bytes）\n\n", sidecar.Name, sidecar.Size)
		fmt.Fprintf(&out, "`decoded` %d／`documented` %d／`unknown` %d\n\n",
			sidecar.ByStatus["decoded"], sidecar.ByStatus["documented"], sidecar.ByStatus["unknown"])
		out.WriteString(tooltext.Text("h.08ba5269af27"))
		for _, row := range sidecar.Fields {
			fmt.Fprintf(&out, "| `%s` | %d | %s | `%s` | spec %s |\n",
				row.Offset, row.Size, row.Name, row.Status, row.Spec)
		}
		out.WriteString("\n")
	}

	out.WriteString(tooltext.Text("h.fc0d0475192c"))
	for _, row := range result.Fields {
		spec := row.Spec
		if spec != "" {
			spec = "spec " + spec
		} else {
			spec = "—"
		}
		fmt.Fprintf(&out, "| `%s` | %d | %s | `%s` | %s | %d |\n",
			row.Offset, row.Size, row.Name, row.Status, spec, row.Consumed)
	}
	return []byte(out.String())
}
