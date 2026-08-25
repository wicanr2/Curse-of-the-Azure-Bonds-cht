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
//
// ⚠ `SOUNDHALT`／`SOUNDOFF`／`SOUNDON`（`4838h`／`483Ah`／`483Ch`）**不是這種格**。
// 它們是 `SOUNDFX` 的**選擇子常數**，和 `CASTFX`…`CRASHFX` 排在同一張表裡，
// 資料段裡就寫死成 255／0／1，全程式一處都沒有寫入。這一支本來把它們當狀態格掃，
// 掃出 0 處寫入，然後照「原作沒寫到就不算待辦」的規則印成 ✅ ——
// **三個假零被當成三項做完的工作**。選擇子的身分歸 `internal/pc98sfx` 管。
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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
	{0x8BF3, "MUSICNO", "select-track", tooltext.Text("h.f7c22d2c0a4a")},
	{0x8BE1, "MUSICNUM", "load-track", tooltext.Text("h.f1c005cd581b")},
	// ⚠ `stop-track` 沒有自己的格：它是 `MUSICNUM := 255` 這個**值**。
	// 只按位址分類會把「停止」算成「載入」，於是原作明明會停音樂，
	// 報表卻說這一類已經接上了。分類要看值。
	{0x8BE1, "MUSICNUM", "stop-track", tooltext.Text("h.f5f3ae981fba")},
	{0x8BE3, "MUSICSW", "music-switch", tooltext.Text("h.d83a548325e9")},
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
	root := flag.String("root", "workplace/re-sweep/pc98", tooltext.Text("h.6f34d54efdfd"))
	resident := flag.String("resident", "PC98-GAME.EXE", tooltext.Text("h.a70cfce00962"))
	symbols := flag.String("symbols", "workplace/re-sweep/pc98/borland-symbols.json", tooltext.Text("h.33ca57ae26bd"))
	remake := flag.String("remake", "internal/game", tooltext.Text("h.82c568be0082"))
	frontend := flag.String("frontend", "cmd/azure-bonds-game", tooltext.Text("h.b2cfe774d76f"))
	output := flag.String("output", "", tooltext.Text("h.fff5cb9e9bc2"))
	outputJSON := flag.String("json", "", tooltext.Text("h.7299c956bbb9"))
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

	report := buildReport(sites, byAction, emitted,
		playerCanStopMusic(*remake, *frontend))
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
	Action  string `json:"action"`
	Symbol  string `json:"symbol"`
	Meaning string `json:"meaning"`
	Sites   int    `json:"sites"`
	Wired   bool   `json:"wired"`
	// Used 是「game-pack 真的有東西會觸發它」。⚠ 和 `Wired` 分開問：
	// 規則層發得出這個動作（能力）不等於現在有哪一段劇情會發（有沒有用到）。
	// 混成一格會讓「寫了但沒人呼叫」看起來像做完了。
	Used     bool     `json:"used"`
	UsedNote string   `json:"used_note,omitempty"`
	RemakeBy string   `json:"remake_by,omitempty"`
	Values   []int    `json:"values,omitempty"`
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
	"select-track": {"play", tooltext.Text("h.9be517378512")},
	"load-track":   {"play", tooltext.Text("h.8aa109986bc4")},
	"stop-track":   {"stop", tooltext.Text("h.b6da1e70452c")},
	"music-switch": {"stop", tooltext.Text("h.48b913622502")},
}

// playerCanStopMusic 問「玩家真的按得到停止嗎」。
//
// ★ 這個問題本來問錯了。原本問的是「game-pack 裡有沒有 `track_id` 是空的
// binding」，那是把「停止」當成**劇情資料**——但原作根本沒有「這一段不放音樂」
// 這種資料：派曲常式（`sub_18AA7`）查不到就 `ret`，音樂繼續放。原作**唯一**會停
// 音樂的地方是玩家把音樂關掉（`MUSICSW`），而那是**按鍵**不是資料。
//
// ⚠ 所以「共用 engine 不收空的 `track_id`」從來就不是卡點：那條路一開始就不該走。
// 這裡改問「規則層有沒有開關，而且前端有沒有把它綁到按鍵上」——兩者缺一都等於
// 玩家按不到（spec 1192）。
func playerCanStopMusic(rulesDir, frontendDir string) bool {
	return declaresMethod(rulesDir, "ToggleMusicSwitch") &&
		callsSelector(frontendDir, "ToggleMusicSwitch")
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
					row.UsedNote = tooltext.Text("h.a9c802f9ce36") +
						tooltext.Text("h.c734a1201fc8")
				}
			default:
				row.Used = row.Wired
			}
		}
		// ⚠ 原作**一處都沒寫**的格子不算待辦：沒有那個動作就沒有要接的東西。
		if row.Sites == 0 {
			row.Wired, row.Used = true, true
			row.RemakeBy = tooltext.Text("h.f8966e9410e2")
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
	out.WriteString(tooltext.Text("h.0bfb3ba682c4"))
	out.WriteString(tooltext.Text("h.d219533ee3a9"))
	out.WriteString(tooltext.Text("h.a536125fe2f4") +
		tooltext.Text("h.5f992d4f0d5c"))
	out.WriteString(tooltext.Text("h.37abc71d59cc") +
		tooltext.Text("h.9e79f707b264"))
	out.WriteString(tooltext.Text("h.805ad7b6ee0e") +
		tooltext.Text("h.f66564471c5c"))
	fmt.Fprint(&out, tooltext.Format("h.ac2fd653402c", strings.Join(doc.Emitted, "`、`")))
	out.WriteString(tooltext.Text("h.e1d8e22f4409") +
		tooltext.Text("h.3b9ba87ff577"))
	out.WriteString(tooltext.Text("h.eff4e3ffb727") +
		tooltext.Text("h.325d2cd3a26e") +
		tooltext.Text("h.726d3de930e6") +
		"（spec 1192）。\n\n")
	fmt.Fprint(&out, tooltext.Text("h.97409a55249f")+
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
	fmt.Fprintf(&out, tooltext.Text("h.9361c4d2c3de")+
		tooltext.Text("h.311e5905eb51"),
		doc.Sites, len(doc.Actions), doc.Wired, doc.Used)

	out.WriteString(tooltext.Text("h.85049007ca2b"))
	out.WriteString("`SOUNDHALT`（`4838h`）、`SOUNDOFF`（`483Ah`）、`SOUNDON`（`483Ch`）" +
		tooltext.Text("h.3e47983cee35") +
		tooltext.Text("h.a403a620b84e") +
		tooltext.Text("h.a30fbecf7454"))
	out.WriteString(tooltext.Text("h.85fb8446978c") +
		tooltext.Text("h.637f9daebe30") +
		tooltext.Text("h.6108a81c26af") +
		tooltext.Text("h.0d13e6c24112") +
		tooltext.Text("h.2b25e85e87d8"))

	out.WriteString(tooltext.Text("h.9ef47d96e206"))
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
