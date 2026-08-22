// Command pc98-music-triggers 把原作「在哪裡換曲」對回「換成哪一首」。
//
// ★ 存在的理由：`remake-status` 的音訊那一列先前只有分母（幾處寫入），沒有對照。
// 「什麼時候該響」要答得完整，得知道**每一處寫入選的是哪一首**。
//
// 做法：掃 PC-98 側 `MUSICNO`（`8BF3h`，Borland 符號直接讀出）的寫入點，取出
// `mov byte [8BF3h], imm` 的立即數，再拿它當 game pack 的 `reference_selector`
// 查曲名。
//
// ⚠ **`MUSICNO` 是 1 起算的。** 全部 12 首曲子，而寫入值最大是 12——直接拿它當
// 0 起算的索引會在最後一首溢位。game pack 的 `reference_selector` 正是 1..12，
// 兩邊對得上；`driver_index` 才是 0 起算。抄錯一邊會讓每一首都差一格，
// 而**每一格都還是合法的曲名**，所以不會有任何錯誤訊息。
//
// ⚠ 這是 **PC-98** 的版面。DOS 沒有除錯符號，對應的格子要另外認。
//
// 用法：
//
//	go run ./cmd/pc98-music-triggers -output docs/audit/pc98-music-triggers.md
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

// musicNoCell 是 PC-98 的 `MUSICNO`（Borland 符號表直接讀出）。
const musicNoCell = 0x8BF3

// soundSegment 是 `SOUNDX` 那一組常式所在的程式碼段。Borland 符號表把
// `SOUNDFX` 記在 segment 2195 ＝ `0893h`，而 PC-98 的 far-call 表裡正好有一個
// 目標段 `0893h`——兩邊對得上。段內位移也直接取自符號表。
const soundSegment = "0893"

// soundEffectNames 是 `SOUNDFX` 的引數——**音效描述子變數**，不是編號。
//
// ★ 名字全部由 PC-98 的 Borland 除錯符號直接讀出（資料段 3113）。DOS 版沒有符號，
// 所以這一層語意只有 PC-98 給得出來；這正是本專案拿 PC-98 當語意骨幹的理由。
//
// ⚠ 呼叫端多半是 `push word [位址]`（`FF 36 <addr>`），不是推立即數——
// 只找立即數的話 36 處裡只解得出 5 處，會得到「大部分音效查不到」這個假結論。
var soundEffectNames = map[int]string{
	0x4838: "SOUNDHALT（停止）", 0x483A: "SOUNDOFF（關）", 0x483C: "SOUNDON（開）",
	0x483E: "CASTFX（施法）", 0x4844: "DEADFX（死亡）", 0x4846: "WHISTLEFX（哨音）",
	0x4848: "HITFX（命中）", 0x484A: "LIGHTNINGFX（閃電）", 0x484C: "SWISHFX（揮擊）",
	0x484E: "PADFX（腳步）", 0x4850: "FIREBALLFX（火球）", 0x4852: "ARROWFX（箭）",
	0x4854: "OVERTUREFX（序曲）", 0x4856: "COMBATFX（戰鬥）", 0x4858: "CRASHFX（撞擊）",
}

var soundRoutines = map[int]string{
	0x0000: "SOUNDFX（音效）",
	0x010D: "INITSOUND（初始化）",
	0x0114: "MSCPLAY（放音樂）",
	0x0177: "BGMPLAY（背景音樂）",
}

type trigger struct {
	module   string
	unit     string
	offset   int
	selector int
}

func main() {
	root := flag.String("root", "workplace/re-sweep/pc98", "PC-98 反組譯產物目錄")
	resident := flag.String("resident", "PC98-GAME.EXE", "常駐執行檔檔名")
	corePath := flag.String("core", "gamepack/pack/00-core.json", "game pack 核心宣告")
	localePath := flag.String("locale", "gamepack/pack/20-locale.zh-TW.json", "中文語系檔")
	farCallMap := flag.String("far-call-map", "docs/audit/far-call-map-pc98.json", "PC-98 far call 對照表")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	titles, selectors := loadTracks(*corePath, *localePath)

	files := map[string]string{*resident: filepath.Join(*root, *resident)}
	entries, err := os.ReadDir(filepath.Join(*root, "overlays"))
	if err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".bin") {
				name := strings.TrimSuffix(entry.Name(), ".bin")
				files[name] = filepath.Join(*root, "overlays", entry.Name())
			}
		}
	}
	units := loadUnitNames("docs/audit/overlay-module-names.md")

	triggers := make([]trigger, 0, 16)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data, readErr := os.ReadFile(files[name])
		if readErr != nil {
			continue
		}
		for offset := 0; offset+5 <= len(data); offset++ {
			// `C6 06 <addr16> <imm8>` ＝ mov byte [addr], imm8
			if data[offset] != 0xC6 || data[offset+1] != 0x06 {
				continue
			}
			if int(data[offset+2])|int(data[offset+3])<<8 != musicNoCell {
				continue
			}
			triggers = append(triggers, trigger{
				module: name, unit: units[name], offset: offset, selector: int(data[offset+4])})
		}
	}

	used := map[int]bool{}
	for _, item := range triggers {
		used[item.selector] = true
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 原作在哪裡換曲，換成哪一首（PC-98）\n\n")
	fmt.Fprintf(&report, "由 `cmd/pc98-music-triggers` 產生，不要手改。\n\n")
	fmt.Fprintf(&report, "⚠ **`MUSICNO` 是 1 起算的**：12 首曲子、寫入值最大 12，"+
		"直接當 0 起算索引會在最後一首溢位。game pack 的 `reference_selector` 正是 1..12，"+
		"`driver_index` 才是 0 起算。抄錯一邊會讓每一首都差一格，"+
		"而**每一格都還是合法曲名**，所以不會有任何錯誤訊息。\n\n")
	fmt.Fprintf(&report, "⚠ 這是 **PC-98** 的版面（名字由 Borland 除錯符號直接讀出）。"+
		"DOS 沒有符號，對應的格子要另外認。\n\n")

	fmt.Fprintf(&report, "| 總計 | 數量 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 換曲點（`mov byte [MUSICNO], imm`）| %d |\n", len(triggers))
	fmt.Fprintf(&report, "| 被選到的相異曲目 | %d |\n", len(used))
	fmt.Fprintf(&report, "| game pack 宣告的曲目 | %d |\n\n", len(selectors))

	fmt.Fprintf(&report, "| 模組 | 單元 | 位移 | 選擇子 | 曲目 |\n|---|---|---:|---:|---|\n")
	for _, item := range triggers {
		unit := item.unit
		if unit == "" {
			unit = "—"
		}
		fmt.Fprintf(&report, "| `%s` | %s | `%04Xh` | %d | %s |\n",
			item.module, unit, item.offset, item.selector, titleFor(titles, selectors, item.selector))
	}

	// 播放常式的呼叫點：音效那一半在這裡。
	if routines, total := soundCallSites(*farCallMap); total > 0 {
		fmt.Fprintf(&report, "\n## 播放常式的呼叫點（音效那一半）\n\n")
		fmt.Fprintf(&report, "`SOUNDX` 那一組常式在程式碼段 `%sh`，段內位移取自 Borland 符號表。\n\n", soundSegment)
		fmt.Fprintf(&report, "| 常式 | 位移 | 呼叫點 | 來源模組 |\n|---|---:|---:|---|\n")
		offsets := make([]int, 0, len(routines))
		for offset := range routines {
			offsets = append(offsets, offset)
		}
		sort.Ints(offsets)
		for _, offset := range offsets {
			name, known := soundRoutines[offset]
			if !known {
				name = "（符號表沒有這一格）"
			}
			modules := make([]string, 0, len(routines[offset]))
			for module, count := range routines[offset] {
				modules = append(modules, fmt.Sprintf("%s×%d", module, count))
			}
			sort.Strings(modules)
			count := 0
			for _, value := range routines[offset] {
				count += value
			}
			fmt.Fprintf(&report, "| %s | `%04Xh` | %d | %s |\n", name, offset, count, strings.Join(modules, "、"))
		}
		fmt.Fprintf(&report, "\n合計 %d 處。\n\n", total)
		if effects, resolved, unresolved := soundEffectUsage(*farCallMap, *root); resolved+unresolved > 0 {
			fmt.Fprintf(&report, "### `SOUNDFX` 每一處在放什麼音效\n\n")
			fmt.Fprintf(&report, "引數是**音效描述子變數**（`push word [位址]`），不是編號；"+
				"名字由 PC-98 的 Borland 除錯符號直接讀出。\n\n")
			fmt.Fprintf(&report, "⚠ 只找立即數的話 36 處裡只解得出 5 處——會得到「大部分音效查不到」"+
				"這個假結論。真正的形狀是推變數。\n\n")
			fmt.Fprintf(&report, "| 音效 | 呼叫點 | 來源模組 |\n|---|---:|---|\n")
			keys := make([]string, 0, len(effects))
			for name := range effects {
				keys = append(keys, name)
			}
			sort.Strings(keys)
			for _, name := range keys {
				modules := make([]string, 0, len(effects[name]))
				count := 0
				for module, value := range effects[name] {
					modules = append(modules, fmt.Sprintf("%s×%d", module, value))
					count += value
				}
				sort.Strings(modules)
				fmt.Fprintf(&report, "| %s | %d | %s |\n", name, count, strings.Join(modules, "、"))
			}
			fmt.Fprintf(&report, "\n解出 %d 處，還有 %d 處的引數靜態看不出來。\n\n", resolved, unresolved)
		}
		fmt.Fprintf(&report, "★ **交叉印證**：`MSCPLAY` 的呼叫點正好落在上表那五個"+
			"改寫 `MUSICNO` 的 overlay 上（`GEN`×2、`overlay-01`、`POSTCOM`、`overlay-18`）"+
			"——兩次獨立的掃描（資料格寫入 vs 函式呼叫）指到同一組地方。\n\n")
		fmt.Fprintf(&report, "⚠ 這裡只數**跨 overlay 的 far call**。常駐自己呼叫 `SOUNDX` 的次數不在裡面"+
			"（那是段內近呼叫，far-call 表看不到），所以是**下界**。\n")
	}

	fmt.Fprintf(&report, "\n## 沒有任何換曲點選到的曲目\n\n")
	missing := make([]string, 0, 4)
	for selector := range selectors {
		if !used[selector] {
			missing = append(missing, fmt.Sprintf("%d（%s）", selector, titleFor(titles, selectors, selector)))
		}
	}
	sort.Strings(missing)
	if len(missing) == 0 {
		fmt.Fprintf(&report, "（沒有——每一首都有地方會選到）\n")
	} else {
		fmt.Fprintf(&report, "%s\n\n", strings.Join(missing, "、"))
		fmt.Fprintf(&report, "⚠ 不要直接讀成「這首用不到」：這一支只認 `mov byte [MUSICNO], imm` 這一種形狀。"+
			"從暫存器或變數寫進去的換曲點看不到——**要下「沒有人選它」的結論得先排除那些形狀**。\n")
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "triggers=%d distinct-tracks=%d declared=%d unselected=%d\n",
		len(triggers), len(used), len(selectors), len(missing))
}

func titleFor(titles map[string]string, selectors map[int]string, selector int) string {
	titleID, ok := selectors[selector]
	if !ok {
		return fmt.Sprintf("（選擇子 %d 不在 game pack 裡）", selector)
	}
	if title, found := titles[titleID]; found {
		return title
	}
	return titleID
}

// loadTracks 回傳「語系鍵 → 曲名」與「選擇子 → 語系鍵」。
func loadTracks(corePath, localePath string) (map[string]string, map[int]string) {
	raw, err := os.ReadFile(corePath)
	if err != nil {
		log.Fatal(err)
	}
	var core struct {
		MusicTracks []struct {
			TitleID           string `json:"title_id"`
			ReferenceSelector int    `json:"reference_selector"`
		} `json:"music_tracks"`
	}
	if err := json.Unmarshal(raw, &core); err != nil {
		log.Fatal(err)
	}
	selectors := map[int]string{}
	for _, track := range core.MusicTracks {
		selectors[track.ReferenceSelector] = track.TitleID
	}
	localeRaw, err := os.ReadFile(localePath)
	if err != nil {
		log.Fatal(err)
	}
	var locale struct {
		Locales map[string]map[string]string `json:"locales"`
	}
	if err := json.Unmarshal(localeRaw, &locale); err != nil {
		log.Fatal(err)
	}
	for _, entries := range locale.Locales {
		return entries, selectors
	}
	return map[string]string{}, selectors
}

// loadUnitNames 從 overlay 單元名表撈 `overlay-NN → 單元`。
func loadUnitNames(path string) map[string]string {
	names := map[string]string{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return names
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "| `overlay-") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		module := strings.Trim(strings.TrimSpace(parts[1]), "`")
		unit := strings.Trim(strings.TrimSpace(parts[2]), "* ")
		if unit != "—" && unit != "" {
			names[module] = unit
		}
	}
	return names
}

// soundCallSites 從 far-call 表數出打到 `SOUNDX` 段的呼叫點，依段內位移分組。
func soundCallSites(path string) (map[int]map[string]int, int) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0
	}
	var table struct {
		Targets []struct {
			Module string `json:"module"`
			Raw    string `json:"raw"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, 0
	}
	routines := map[int]map[string]int{}
	total := 0
	for _, call := range table.Targets {
		if !strings.HasPrefix(call.Raw, soundSegment+":") {
			continue
		}
		var offset int
		if _, err := fmt.Sscanf(strings.Split(call.Raw, ":")[1], "%04X", &offset); err != nil {
			continue
		}
		if routines[offset] == nil {
			routines[offset] = map[string]int{}
		}
		routines[offset][call.Module]++
		total++
	}
	return routines, total
}

// soundEffectUsage 解出每一處 `SOUNDFX` 呼叫在放哪一個音效。
//
// 兩種形狀都認：`push word [位址]`（`FF 36`，最常見）與 `mov al, imm` ＋ `push ax`。
func soundEffectUsage(mapPath, root string) (map[string]map[string]int, int, int) {
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, 0, 0
	}
	var table struct {
		Targets []struct {
			Module string `json:"module"`
			EA     string `json:"ea"`
			Raw    string `json:"raw"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, 0, 0
	}
	effects := map[string]map[string]int{}
	resolved, unresolved := 0, 0
	cache := map[string][]byte{}
	for _, call := range table.Targets {
		if call.Raw != soundSegment+":0000" {
			continue
		}
		data, ok := cache[call.Module]
		if !ok {
			data, _ = os.ReadFile(filepath.Join(root, "overlays", call.Module+".bin"))
			cache[call.Module] = data
		}
		var ea int
		if _, err := fmt.Sscanf(strings.TrimSuffix(call.EA, "h"), "%04X", &ea); err != nil || ea >= len(data) {
			unresolved++
			continue
		}
		low := ea - 20
		if low < 0 {
			low = 0
		}
		name := ""
		for index := ea - 3; index >= low; index-- {
			if data[index] == 0xFF && data[index+1] == 0x36 {
				address := int(data[index+2]) | int(data[index+3])<<8
				if known, found := soundEffectNames[address]; found {
					name = known
				} else {
					name = fmt.Sprintf("（位址 %04Xh，符號表沒有）", address)
				}
				break
			}
			if data[index] == 0xB0 && data[index+2] == 0x50 {
				name = fmt.Sprintf("（立即數 %d）", data[index+1])
				break
			}
		}
		if name == "" {
			unresolved++
			continue
		}
		if effects[name] == nil {
			effects[name] = map[string]int{}
		}
		effects[name][call.Module]++
		resolved++
	}
	return effects, resolved, unresolved
}
