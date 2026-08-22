// Command checkfx-callsites 找出原版每一處 `CHECKFX(時機, 對象)` 的呼叫點。
//
// ★ 存在的理由：`docs/audit/checkfx-timing-table.md` 已經有「每個時機對應哪些
// 效果碼」，但沒有「**誰在什麼時候問這個時機**」。少了呼叫點，就無法回答
// remake 那一側該把哪一次查詢接在哪裡——也無法回答一個更基本的問題：
// **這個時機到底會不會被問到。**
//
// ⚠ 兩種呼叫都要掃，只掃一種會得到假零：
//
//	far call  跨 overlay 的呼叫，來源在 `docs/audit/far-call-map-dos.json`。
//	near call `PUTDAMAGE` 這類**住在 overlay-23 自己裡面**的呼叫端
//	          （DOS `overlay-23:1FDFh`，spec 581）用的是 `E8` 近呼叫，
//	          far-call 表裡一筆都沒有。第一版只掃 far call，於是「時機 06h
//	          沒有呼叫點」——而 spec 581 明明白白寫著它有。
//
// ⚠ 時機是 `mov al, imm` ＋ `push ax` 推進去的，往回找這個樣式。窗口要夠寬：
// 16 bytes 會漏掉引數多的呼叫端（實測 overlay-22 那一處的 `B0 12 50` 在 −22）。
// 找不到樣式的呼叫點會被列成「沒解出來」，不會被當成不存在。
//
// 用法：
//
//	go run ./cmd/checkfx-callsites -output docs/audit/checkfx-callsites.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// checkfxEntry 是 `overlay-23 entry#4` 在自己模組裡的位移（far-call 表的 code_offset）。
const checkfxEntry = 0x03FE

// checkfxStubOffset 是 `overlay-23 entry#4` 的 stub 位移（far-call 表的 raw ＝ `0141:0034`）。
const checkfxStubOffset = 0x0034

type farCall struct {
	Raw        string `json:"raw"`
	Module     string `json:"module"`
	Function   string `json:"function"`
	EA         string `json:"ea"`
	Target     string `json:"target"`
	Entry      int    `json:"entry"`
	CodeOffset string `json:"code_offset"`
}

type site struct {
	where  string
	timing int // −1 ＝ 沒解出來
}

func main() {
	mapPath := flag.String("far-call-map", "docs/audit/far-call-map-dos.json", "far call 對照表")
	overlays := flag.String("overlays", "workplace/re-sweep/dos/overlays", "overlay 二進位所在目錄")
	window := flag.Int("window", 32, "往回找 `mov al,imm`＋`push ax` 的窗口大小")
	resident := flag.String("resident", "workplace/re-sweep/dos/START.EXE", "常駐執行檔（用來排除常駐側呼叫）")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	raw, err := os.ReadFile(*mapPath)
	if err != nil {
		log.Fatal(err)
	}
	var table struct {
		Targets []farCall `json:"targets"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		log.Fatal(err)
	}

	cache := map[string][]byte{}
	load := func(module string) []byte {
		if data, ok := cache[module]; ok {
			return data
		}
		data, readErr := os.ReadFile(filepath.Join(*overlays, module+".bin"))
		if readErr != nil {
			data = nil
		}
		cache[module] = data
		return data
	}

	sites := make([]site, 0, 32)
	// 1. 跨 overlay 的 far call。
	for _, call := range table.Targets {
		if call.Target != "overlay-23" || call.Entry != 4 {
			continue
		}
		data := load(call.Module)
		offset, parseErr := strconv.ParseUint(strings.TrimSuffix(call.EA, "h"), 16, 32)
		if data == nil || parseErr != nil {
			continue
		}
		sites = append(sites, site{
			where:  fmt.Sprintf("`%s`/%s @ `%s`（far）", call.Module, call.Function, call.EA),
			timing: timingBefore(data, int(offset), *window),
		})
	}
	// 2. overlay-23 自己裡面的 near call。
	if data := load("overlay-23"); data != nil {
		for ea := 0; ea+3 <= len(data); ea++ {
			if data[ea] != 0xE8 {
				continue
			}
			rel := int(int16(uint16(data[ea+1]) | uint16(data[ea+2])<<8))
			if (ea+3+rel)&0xFFFF != checkfxEntry {
				continue
			}
			sites = append(sites, site{
				where:  fmt.Sprintf("`overlay-23` @ `%04Xh`（near）", ea),
				timing: timingBefore(data, ea, *window),
			})
		}
	}

	byTiming := map[int][]string{}
	unresolved := make([]string, 0, 4)
	for _, item := range sites {
		if item.timing < 0 {
			unresolved = append(unresolved, item.where)
			continue
		}
		byTiming[item.timing] = append(byTiming[item.timing], item.where)
	}
	timings := make([]int, 0, len(byTiming))
	for timing := range byTiming {
		timings = append(timings, timing)
	}
	sort.Ints(timings)

	var report strings.Builder
	fmt.Fprintf(&report, "# `CHECKFX` 的呼叫點：每個時機是誰在什麼時候問的\n\n")
	fmt.Fprintf(&report, "由 `cmd/checkfx-callsites` 產生，不要手改。理由與兩種呼叫的差別見該檔註解。\n\n")
	fmt.Fprintf(&report, "`CHECKFX` ＝ `overlay-23 entry#4`（模組內位移 `%04Xh`）。"+
		"時機清單見 [`checkfx-timing-table.md`](checkfx-timing-table.md)。\n\n", checkfxEntry)
	fmt.Fprintf(&report, "| 時機 | 呼叫點數 | 在哪 |\n|---|---:|---|\n")
	for _, timing := range timings {
		fmt.Fprintf(&report, "| `%02Xh` | %d | %s |\n",
			timing, len(byTiming[timing]), strings.Join(byTiming[timing], "、"))
	}
	fmt.Fprintf(&report, "\n共 %d 處呼叫點。\n\n", len(sites))
	if len(unresolved) > 0 {
		fmt.Fprintf(&report, "## 沒解出時機的呼叫點\n\n")
		fmt.Fprintf(&report, "時機不是用 `mov al, imm` 推進去的（可能來自變數或暫存器）。"+
			"**這些不代表時機不存在**，只代表這一支靜態掃不出來。\n\n")
		for _, where := range unresolved {
			fmt.Fprintf(&report, "- %s\n", where)
		}
		fmt.Fprintf(&report, "\n")
	}

	// 常駐側：Borland 的 overlay 呼叫一律是 far call 到固定的 stub。
	//
	// ⚠ 這裡只比**位移**不比段——而這對「沒有」這個方向是安全的：位移根本沒出現
	// 過，就沒有任何段值能讓它變成一次 CHECKFX 呼叫。反過來（宣稱「有」）才需要
	// 比段，因為 stub 位移會撞號。
	//
	// ⚠⚠ 而且一定要有**正對照**：常駐側如果根本不用 far call 叫 overlay，那
	// 「找不到 CHECKFX」就什麼都不代表。下面同時數常駐側叫得到幾種 stub 位移。
	residentCalls, residentStubs := -1, -1
	if payload, readErr := os.ReadFile(*resident); readErr == nil {
		stubs := map[int]bool{}
		for _, call := range table.Targets {
			if !strings.HasPrefix(call.Target, "overlay-") {
				continue
			}
			if _, offset, ok := splitRaw(call.Raw); ok {
				stubs[offset] = true
			}
		}
		residentCalls = countFarCallsTo(payload, checkfxStubOffset)
		residentStubs = 0
		for offset := range stubs {
			if countFarCallsTo(payload, offset) > 0 {
				residentStubs++
			}
		}
		fmt.Fprintf(&report, "## 常駐執行檔那一側\n\n")
		fmt.Fprintf(&report, "| 檢查 | 結果 |\n|---|---:|\n")
		fmt.Fprintf(&report, "| 常駐側呼叫 `CHECKFX` 的 stub 位移 `%02Xh` 幾次 | **%d** |\n",
			checkfxStubOffset, residentCalls)
		fmt.Fprintf(&report, "| 正對照：常駐側叫得到幾種 overlay stub 位移 | %d／%d |\n\n",
			residentStubs, len(stubs))
		if residentCalls == 0 && residentStubs > 0 {
			fmt.Fprintf(&report, "正對照成立（常駐側**確實會**用 far call 叫 overlay），"+
				"而 `CHECKFX` 的 stub 位移一次都沒出現 ⇒ **常駐側不呼叫 `CHECKFX`**。\n\n")
		}
	}

	fmt.Fprintf(&report, "## 沒有呼叫點的時機\n\n")
	missing := make([]string, 0, 4)
	for timing := 0; timing <= 0x16; timing++ {
		if len(byTiming[timing]) == 0 {
			missing = append(missing, fmt.Sprintf("`%02Xh`", timing))
		}
	}
	if len(missing) == 0 {
		fmt.Fprintf(&report, "（沒有）\n")
	} else {
		fmt.Fprintf(&report, "%s ——分派表裡有效果碼，但找不到任何呼叫端。\n\n", strings.Join(missing, "、"))
		if len(unresolved) == 0 && residentCalls == 0 && residentStubs > 0 {
			fmt.Fprintf(&report, "三個方向都排除掉了：跨 overlay 的 far call、"+
				"overlay-23 內部的近呼叫、常駐執行檔（正對照成立）。"+
				"而且 %d 處呼叫點的時機**全部**是 `mov al, imm` 推進去的，沒有一處來自變數。\n\n",
				len(sites))
			fmt.Fprintf(&report, "⇒ 可以說：**這兩個時機在 DOS 版是死的**。"+
				"只在它們底下出現的效果碼，原作永遠不會執行——remake 不實作它們**是對的**，"+
				"把它們算成缺口反而會為死碼寫程式。\n")
		} else {
			fmt.Fprintf(&report, "⚠ 還不能讀成「原作不會用到」：仍有沒排除掉的呼叫形狀。\n")
		}
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "sites=%d timings=%d unresolved=%d no-callsite=%d\n",
		len(sites), len(timings), len(unresolved), len(missing))
}

// timingBefore 往回找 `mov al, imm`（`B0 xx`）緊接著 `push ax`（`50`）。
func timingBefore(data []byte, callEA, window int) int {
	low := callEA - window
	if low < 0 {
		low = 0
	}
	for index := callEA - 3; index >= low; index-- {
		if data[index] == 0xB0 && data[index+2] == 0x50 {
			return int(data[index+1])
		}
	}
	return -1
}

// splitRaw 把 far-call 表的 `段:位移` 拆開。
func splitRaw(raw string) (segment, offset int, ok bool) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	segmentValue, err1 := strconv.ParseUint(parts[0], 16, 32)
	offsetValue, err2 := strconv.ParseUint(parts[1], 16, 32)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return int(segmentValue), int(offsetValue), true
}

// countFarCallsTo 數 `9A <位移低> <位移高>` 出現幾次。段值不比——重定位會改寫它，
// 而對「沒有出現」這個方向不比段是安全的。
func countFarCallsTo(payload []byte, offset int) int {
	count := 0
	low, high := byte(offset&0xFF), byte((offset>>8)&0xFF)
	for index := 0; index+3 <= len(payload); index++ {
		if payload[index] == 0x9A && payload[index+1] == low && payload[index+2] == high {
			count++
		}
	}
	return count
}
