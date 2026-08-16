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
	Schema      string         `json:"schema"`
	RecordSize  int            `json:"record_size"`
	Baselines   int            `json:"baselines"`
	ByStatus    map[string]int `json:"bytes_by_status"`
	ConsumedAll int            `json:"consumed_bytes"`
	Fields      []fieldRow     `json:"fields"`
	Mismatches  []string       `json:"mismatches"`
	Sidecars    []sidecarReport `json:"sidecars"`
}

func main() {
	output := flag.String("output", "", "Markdown 輸出路徑")
	outputJSON := flag.String("json", "", "JSON 輸出路徑")
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
	fmt.Fprintf(os.Stderr,
		"record=%d decoded=%d documented=%d unknown=%d 量到有讀=%d 對帳不符=%d\n",
		result.RecordSize, result.ByStatus["decoded"], result.ByStatus["documented"],
		result.ByStatus["unknown"], result.ConsumedAll, len(result.Mismatches))
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
			return report{}, fmt.Errorf("基準記錄 %q 解析不過：%w", baseline.name, baseErr)
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
			result.Mismatches = append(result.Mismatches, fmt.Sprintf(
				"+%03Xh（%s）量到 %d 個位元組有讀，台帳卻標成 %s",
				field.Offset, field.Name, row.Consumed, field.Status))
		}
		// 台帳說 decoded，卻沒有任何基準記錄量到 ⇒ 補基準記錄，別改台帳。
		if row.Consumed == 0 && field.Status == party.DOSFieldDecoded {
			result.Mismatches = append(result.Mismatches, fmt.Sprintf(
				"+%03Xh（%s）台帳標 decoded，但沒有任何基準記錄量到它——"+
					"補一份能觸發它的基準記錄", field.Offset, field.Name))
		}
		result.Fields = append(result.Fields, row)
	}
	sort.Strings(result.Mismatches)

	for _, sidecar := range []struct {
		name   string
		fields []monster.RecordField
		size   int
	}{
		{".SWG 物品記錄", monster.ItemRecordFields, monster.ItemRecordSize},
		{".FX 效果記錄", monster.AffectRecordFields, monster.AffectRecordSize},
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
	out.WriteString("# DOS 角色記錄的逐位元組覆蓋（`CHARREC`，1A6h bytes）\n\n")
	out.WriteString("由 `cmd/save-field-coverage` 產生，不要手改。\n\n")
	out.WriteString("- **狀態**來自 `internal/party.DOSPlayerRecordFields` 的人工台帳，" +
		"每一段都必須有出處；`unknown` 是「還沒查到」，不是「沒有用」。\n")
	out.WriteString("- **量到有讀**是機器量的：改一個位元組、重新解析、比對結果，" +
		"有差別就是有讀。掃程式碼會漏掉算出來的索引，量測不會。\n")
	out.WriteString("- 量測是**下界**：只在某些職業／種族才被讀的位元組，" +
		"基準記錄沒涵蓋就會被判成沒讀。所以基準有多份，而且「台帳說 decoded 卻沒量到」" +
		"會讓這支回非零離開碼。\n")
	out.WriteString("- ⚠ **匯入不會因為 `unknown` 而遺失資料**：`LoadSAVGAMSlot` 保留整份原始" +
		"記錄，寫回時只動已知位移（spec 185）。這份報告衡量的是**理解程度**，不是保真度。\n")
	out.WriteString("- 對帳的粒度是**每一段**，不是每一個位元組：長度可變的內容" +
		"（名字尾巴、沒用到的法術槽）在單一份基準記錄裡本來就量不到，" +
		"逐位元組要求會產生一堆假警報。\n\n")

	fmt.Fprintf(&out, "## 摘要\n\n| 狀態 | 位元組 | 佔比 |\n|---|---:|---:|\n")
	for _, status := range []string{"decoded", "documented", "unknown"} {
		count := result.ByStatus[status]
		fmt.Fprintf(&out, "| `%s` | %d | %d%% |\n", status, count, count*100/result.RecordSize)
	}
	fmt.Fprintf(&out, "| 合計 | %d | 100%% |\n", result.RecordSize)
	fmt.Fprintf(&out, "\n解析器實際讀到的位元組（%d 份基準記錄的聯集）：**%d／%d**。\n\n",
		result.Baselines, result.ConsumedAll, result.RecordSize)

	out.WriteString("## Sidecar 記錄\n\n")
	out.WriteString("`.SWG`（物品）與 `.FX`（效果）沒有突變探針：兩者的解析器逐欄直讀、" +
		"沒有算出來的索引，台帳蓋滿就足以說明覆蓋。\n\n")
	for _, sidecar := range result.Sidecars {
		fmt.Fprintf(&out, "### %s（%d bytes）\n\n", sidecar.Name, sidecar.Size)
		fmt.Fprintf(&out, "`decoded` %d／`documented` %d／`unknown` %d\n\n",
			sidecar.ByStatus["decoded"], sidecar.ByStatus["documented"], sidecar.ByStatus["unknown"])
		out.WriteString("| 位移 | 長度 | 欄位 | 狀態 | 出處 |\n|---|---:|---|---|---|\n")
		for _, row := range sidecar.Fields {
			fmt.Fprintf(&out, "| `%s` | %d | %s | `%s` | spec %s |\n",
				row.Offset, row.Size, row.Name, row.Status, row.Spec)
		}
		out.WriteString("\n")
	}

	out.WriteString("## 角色記錄逐段\n\n| 位移 | 長度 | 欄位 | 狀態 | 出處 | 量到有讀 |\n|---|---:|---|---|---|---:|\n")
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
