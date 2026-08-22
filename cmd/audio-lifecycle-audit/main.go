// Command audio-lifecycle-audit 回答「原機音訊的**播放生命週期**，remake 對上幾種」。
//
// ★ 存在的理由：`cmd/dseg-writers` 已經盤出「誰決定什麼時候該響」（六格音訊狀態
// 的 17 處寫入），但那是**分母不是覆蓋率**——它說原作在幾個地方改音訊狀態，
// 沒說 remake 有沒有對應的動作。這一支把寫入點按**生命週期動作**分類，再拿去對
// remake 真的會發出來的動作，差額就是待辦。
//
// 生命週期動作按寫哪一格分（PC-98 版面，名字取自 Borland 除錯符號）：
//
//	MUSICNO   `8BF3h`  選曲：換成哪一首
//	MUSICNUM  `8BE1h`  曲目編號：驅動程式要載哪一份資料
//	MUSICSW   `8BE3h`  音樂開關：整個音樂要不要響
//	SOUNDHALT `4838h`  音效停止
//	SOUNDOFF  `483Ah`  音效關
//	SOUNDON   `483Ch`  音效開
//
// ⚠ **這是 PC-98 的版面**。DOS 版沒有除錯符號，位址不能直接套（spec 1187）。
//
// ⚠ 位元組直掃，不走 far-call 對照表：表比實際少，而**下界看起來和全集一樣合理**
// （spec 1186 已經被這個坑咬過一次）。代價是偽陽性——剛好長得像 opcode 的立即數
// 會被算進來，所以每一處都印出所屬常式讓人看得出合不合理。
//
// 用法：
//
//	go run ./cmd/audio-lifecycle-audit -output docs/audit/audio-lifecycle.md \
//	    -json docs/audit/audio-lifecycle.json
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
)

// cell 是一格音訊狀態，以及它代表的生命週期動作。
type cell struct {
	Address uint16
	Symbol  string
	Action  string
	Meaning string
}

var cells = []cell{
	{0x8BF3, "MUSICNO", "select-track", "選曲：換成哪一首"},
	{0x8BE1, "MUSICNUM", "load-track", "曲目編號：驅動程式要載哪一份資料"},
	// ⚠ `stop-track` 沒有自己的格：它是 `MUSICNUM := 255` 這個**值**。
	// 只按位址分類會把「停止」算成「載入」，於是原作明明會停音樂，
	// 報表卻說這一類已經接上了。分類要看值。
	{0x8BE1, "MUSICNUM", "stop-track", "停止：曲目編號寫 `255`（沒有曲子）"},
	{0x8BE3, "MUSICSW", "music-switch", "音樂開關：整個音樂要不要響"},
	{0x4838, "SOUNDHALT", "sfx-halt", "音效停止"},
	{0x483A, "SOUNDOFF", "sfx-off", "音效關"},
	{0x483C, "SOUNDON", "sfx-on", "音效開"},
}

// site 是一處寫入。
type site struct {
	File    string `json:"file"`
	Module  string `json:"module"`
	Routine string `json:"routine"`
	Offset  int    `json:"offset"`
	Symbol  string `json:"symbol"`
	Action  string `json:"action"`
	Form    string `json:"form"`
	// Value 是 `mov [addr],imm` 的立即數；其他形式是 -1。
	// ★ 這一欄才看得出**語意**：`MUSICSW := 0` 是關音樂，`:= 1` 是開。
	Value int `json:"value"`
}

func main() {
	root := flag.String("root", "workplace/re-sweep/pc98", "PC-98 掃描產物目錄")
	resident := flag.String("resident", "PC98-GAME.EXE", "常駐執行檔（相對 root）")
	symbols := flag.String("symbols", "workplace/re-sweep/pc98/borland-symbols.json", "Borland 符號表")
	remake := flag.String("remake", "internal/game", "remake 規則層目錄")
	packCore := flag.String("pack", "gamepack/pack/00-core.json", "game-pack 的 core JSON")
	output := flag.String("output", "", "Markdown 輸出路徑")
	outputJSON := flag.String("json", "", "JSON 輸出路徑")
	flag.Parse()

	index := loadSymbolIndex(*symbols)
	// ⚠ 兩筆 `MUSICNUM` 共用同一個位址，`wanted` 只留「按位址」的那一筆；
	// `stop-track` 由值決定，在 `scanFile` 裡改寫。
	wanted := map[uint16]cell{}
	for _, item := range cells {
		if _, taken := wanted[item.Address]; taken {
			continue
		}
		wanted[item.Address] = item
	}

	sites := make([]site, 0, 32)
	sites = append(sites, scanFile(filepath.Join(*root, *resident), "PC98-GAME.EXE", "", wanted, index)...)
	overlays, err := filepath.Glob(filepath.Join(*root, "overlays", "overlay-*.bin"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(overlays)
	for _, path := range overlays {
		name := strings.TrimSuffix(filepath.Base(path), ".bin")
		sites = append(sites, scanFile(path, name, name, wanted, index)...)
	}

	byAction := map[string][]site{}
	for _, item := range sites {
		byAction[item.Action] = append(byAction[item.Action], item)
	}
	emitted := scanRemakeActions(*remake)

	report := buildReport(sites, byAction, emitted, packUsesStop(*packCore))
	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(report, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	text := renderMarkdown(report, byAction)
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "sites=%d actions=%d wired=%d used=%d\n",
		len(sites), len(report.Actions), report.Wired, report.Used)
}

// scanFile 掃一個檔案裡對這幾格的寫入。編碼與 `cmd/dseg-writers` 同一套。
func scanFile(path, file, module string, wanted map[uint16]cell, index symbolIndex) []site {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	out := make([]site, 0, 8)
	word := func(i int) uint16 { return uint16(buf[i]) | uint16(buf[i+1])<<8 }
	isDisp16 := func(modrm byte) bool { return modrm&0xC7 == 0x06 }
	for i := 0; i+4 < len(buf); i++ {
		op, next := buf[i], buf[i+1]
		var address uint16
		form, value := "", -1
		switch {
		case op == 0xA2 || op == 0xA3:
			address, form = word(i+1), "mov [addr],acc"
		case (op == 0xC6 || op == 0xC7) && next == 0x06:
			address, form = word(i+2), "mov [addr],imm"
			// ⚠ `C6` 是 byte、`C7` 是 word：立即數寬度不同，讀錯會把下一個
			// 位元組算進值裡，而算出來的數字每一個都還是合法曲號 ⇒ 不會報錯。
			if op == 0xC6 {
				value = int(buf[i+4])
			} else if i+5 < len(buf) {
				value = int(word(i + 4))
			}
		case (op == 0x88 || op == 0x89) && isDisp16(next):
			address, form = word(i+2), "mov [addr],reg"
		default:
			continue
		}
		item, ok := wanted[address]
		if !ok {
			continue
		}
		// `MUSICNUM := 255` 是「沒有曲子」＝停止，不是載入。
		if item.Symbol == "MUSICNUM" && value == 0xFF {
			item.Action = "stop-track"
		}
		routine := ""
		if module != "" {
			routine = index.routineAt(module, i)
		}
		out = append(out, site{
			File: file, Module: module, Routine: routine, Offset: i,
			Symbol: item.Symbol, Action: item.Action, Form: form, Value: value,
		})
	}
	return out
}

// actionRow 是一種生命週期動作的處置。
type actionRow struct {
	Action  string   `json:"action"`
	Symbol  string   `json:"symbol"`
	Meaning string   `json:"meaning"`
	Sites   int      `json:"sites"`
	Wired   bool     `json:"wired"`
	// Used 是「game-pack 真的有東西會觸發它」。⚠ 和 `Wired` 分開問：
	// 規則層發得出這個動作（能力）不等於現在有哪一段劇情會發（有沒有用到）。
	// 混成一格會讓「寫了但沒人呼叫」看起來像做完了。
	Used     bool    `json:"used"`
	UsedNote string  `json:"used_note,omitempty"`
	RemakeBy string  `json:"remake_by,omitempty"`
	Values  []int    `json:"values,omitempty"`
	Routines []string `json:"routines,omitempty"`
}

type reportDoc struct {
	Schema  string      `json:"schema"`
	Sites   int         `json:"sites"`
	Actions []actionRow `json:"actions"`
	Wired   int         `json:"wired"`
	Used    int         `json:"used"`
	// Emitted 是 remake 規則層真的會發出來的動作字串。
	Emitted []string `json:"remake_actions"`
}

// remakeCounterpart 說每一種原作動作在 remake 由誰負責。
//
// ⚠ 這張表是**手寫的對照**，不是掃出來的——但它只用來說「誰負責」，
// **有沒有接上是掃 remake 的動作字串決定的**，不是這張表說了算。
var remakeCounterpart = map[string]struct {
	action string
	by     string
}{
	"select-track": {"play", "`State.requestMusicForCurrentBlock`（game-pack 的 music binding）"},
	"load-track":   {"play", "同上：remake 的 `TrackID` 就是曲目，載入由 adapter 負責"},
	"stop-track":   {"stop", "`State.requestMusicForCurrentBlock`：binding 的 `TrackID` 是空的就發 `stop`"},
	"music-switch": {"stop", "同上；開關的「開」由一般的選曲表達"},
	"sfx-halt":     {"sound-halt", "尚未建模：`SoundEvent` 只有「放這個音效」，沒有停止"},
	"sfx-off":      {"sound-off", "尚未建模"},
	"sfx-on":       {"sound-on", "尚未建模"},
}

// packUsesStop 問「game-pack 裡有沒有東西會讓音樂停下來」——也就是有沒有
// `track_id` 是空的 music binding。
func packUsesStop(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var pack struct {
		Bindings []struct {
			TrackID string `json:"track_id"`
		} `json:"music_bindings"`
	}
	if json.Unmarshal(raw, &pack) != nil {
		return false
	}
	for _, binding := range pack.Bindings {
		if strings.TrimSpace(binding.TrackID) == "" {
			return true
		}
	}
	return false
}

func buildReport(sites []site, byAction map[string][]site, emitted map[string]bool, stopUsed bool) reportDoc {
	doc := reportDoc{Schema: "coab-audio-lifecycle/1", Sites: len(sites)}
	for _, item := range cells {
		list := byAction[item.Action]
		row := actionRow{
			Action: item.Action, Symbol: item.Symbol, Meaning: item.Meaning, Sites: len(list),
		}
		values := map[int]bool{}
		routines := map[string]bool{}
		for _, one := range list {
			if one.Value >= 0 {
				values[one.Value] = true
			}
			if one.Routine != "" {
				routines[one.Routine] = true
			}
		}
		for value := range values {
			row.Values = append(row.Values, value)
		}
		sort.Ints(row.Values)
		for routine := range routines {
			row.Routines = append(row.Routines, routine)
		}
		sort.Strings(row.Routines)
		if counterpart, ok := remakeCounterpart[item.Action]; ok {
			row.RemakeBy = counterpart.by
			row.Wired = emitted[counterpart.action]
			switch counterpart.action {
			case "stop":
				row.Used = stopUsed
				if !stopUsed {
					// ⚠ 卡點不是「還沒有人寫那條 binding」，是**寫不出來**：
					// 共用 engine 的 pack 驗證會把 `track_id` 空的 binding 判成
					// `references unknown track ""`。這兩種卡點的處置完全不同，
					// 寫成前者會讓人以為補一行 JSON 就好。
					row.UsedNote = "engine 的 pack 驗證不收 `track_id` 空的 binding，" +
						"「這裡不放音樂」目前**表達不出來**（`TestEnginePackCannotExpressStopYet`）"
				}
			default:
				row.Used = row.Wired
			}
		}
		// ⚠ 原作**一處都沒寫**的格子不算待辦：沒有那個動作就沒有要接的東西。
		if row.Sites == 0 {
			row.Wired, row.Used = true, true
			row.RemakeBy = "原作一處都沒寫到這一格"
		}
		if row.Wired {
			doc.Wired++
		}
		if row.Used {
			doc.Used++
		}
		doc.Actions = append(doc.Actions, row)
	}
	for action := range emitted {
		doc.Emitted = append(doc.Emitted, action)
	}
	sort.Strings(doc.Emitted)
	return doc
}

func renderMarkdown(doc reportDoc, byAction map[string][]site) string {
	var out strings.Builder
	out.WriteString("# 原機音訊的播放生命週期：原作有幾種動作，remake 對上幾種\n\n")
	out.WriteString("由 `cmd/audio-lifecycle-audit` 產生，不要手改。\n\n")
	out.WriteString("`cmd/dseg-writers` 盤的是**分母**（誰決定什麼時候該響）；這一份把那些寫入點按" +
		"**生命週期動作**分類，再拿去對 remake 真的會發出來的動作。差額就是待辦。\n\n")
	out.WriteString("⚠ **這是 PC-98 的版面**，名字取自 Borland 除錯符號。DOS 版沒有符號，" +
		"位址不能直接套（spec 1187）。\n\n")
	out.WriteString("⚠ 位元組直掃，不走 far-call 對照表——表比實際少，而**下界看起來和全集一樣合理**。" +
		"代價是偽陽性，所以每一處都印出所屬常式讓人看得出合不合理。\n\n")
	fmt.Fprintf(&out, "remake 規則層目前發得出來的動作：`%s`\n\n",
		strings.Join(doc.Emitted, "`、`"))
	out.WriteString("⚠ **「發得出來」與「有東西會發」是兩件事**，分兩欄問：規則層有這個動作是**能力**，" +
		"現在有哪一段劇情會發是**有沒有用到**。混成一格會讓「寫了但沒人呼叫」看起來像做完了。\n\n")
	fmt.Fprintf(&out, "| 動作 | 格 | 意思 | 原作寫入處 | 規則層發得出來 | pack 有用到 | 由誰負責 |\n"+
		"|---|---|---|---:|---|---|---|\n")
	for _, row := range doc.Actions {
		mark, used := "—", "—"
		if row.Wired {
			mark = "✅"
		}
		if row.Used {
			used = "✅"
		} else if row.UsedNote != "" {
			used = "— " + row.UsedNote
		}
		fmt.Fprintf(&out, "| `%s` | `%s` | %s | %d | %s | %s | %s |\n",
			row.Action, row.Symbol, row.Meaning, row.Sites, mark, used, row.RemakeBy)
	}
	fmt.Fprintf(&out, "\n合計 %d 處寫入、%d 種動作：規則層發得出 **%d** 種，"+
		"game-pack 真的會發的有 **%d** 種。\n\n",
		doc.Sites, len(doc.Actions), doc.Wired, doc.Used)

	out.WriteString("## 逐處\n\n| 檔案 | 常式 | 位移 | 格 | 形式 | 值 |\n|---|---|---|---|---|---|\n")
	for _, item := range cells {
		for _, one := range byAction[item.Action] {
			value := "—"
			if one.Value >= 0 {
				value = fmt.Sprintf("`%d`", one.Value)
			}
			routine := one.Routine
			if routine == "" {
				routine = "—"
			}
			fmt.Fprintf(&out, "| `%s` | %s | `%04Xh` | `%s` | `%s` | %s |\n",
				one.File, routine, one.Offset, one.Symbol, one.Form, value)
		}
	}
	return out.String()
}
