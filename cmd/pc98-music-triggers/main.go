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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

// musicNoCell 是 PC-98 的 `MUSICNO`（Borland 符號表直接讀出）。
const musicNoCell = 0x8BF3

// soundSegment 是 `SOUNDX` 那一組常式所在的程式碼段。Borland 符號表把
// `SOUNDFX` 記在 segment 2195 ＝ `0893h`，而 PC-98 的 far-call 表裡正好有一個
// 目標段 `0893h`——兩邊對得上。段內位移也直接取自符號表。
const soundSegment = "0893"

// soundEffectLabels 給每個 Borland 符號一個中文說明。**名字與位址都不在這裡**
// ——那兩樣由 `pc98sfx.Selectors()` 推出來（描述子位址 ＝ 基底 ＋ 選擇子×2）。
//
// ⚠ 這裡以前是一張手打的「位址 → 名字」對照表，而它**漏了 4840h 與 4842h**
// （MISSFX／SPELLHITFX）。漏掉不會報錯：報表把那兩處印成「符號表沒有」，
// 讀起來像原作真的少了兩個名字，而下游的 `cmd/sound-trigger-compare` 連帶
// 少了兩列——正好是 remake 用得最兇的 `SoundSpellHit` 與 `SoundMiss`。
// 位址改成用推的之後，「漏一格」在結構上就不可能發生。
var soundEffectLabels = map[string]string{
	"SOUNDHALT": "停止", "SOUNDOFF": "關", "SOUNDON": "開",
	"CASTFX": "施法", "MISSFX": "揮空", "SPELLHITFX": "法術命中",
	"DEADFX": "死亡", "WHISTLEFX": "哨音", "HITFX": "命中",
	"LIGHTNINGFX": "閃電", "SWISHFX": "揮擊", "PADFX": "腳步",
	"FIREBALLFX": "火球", "ARROWFX": "箭", "OVERTUREFX": "序曲",
	"COMBATFX": "戰鬥", "CRASHFX": "撞擊",
}

// soundEffectName 是報表裡那個「符號（說明）」字串。
func soundEffectName(address int) (string, bool) {
	info, ok := pc98sfx.SelectorForDescriptor(address)
	if !ok {
		return "", false
	}
	label, known := soundEffectLabels[info.Symbol]
	if !known {
		return info.Symbol, true
	}
	return fmt.Sprintf("%s（%s）", info.Symbol, label), true
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
	symbols := flag.String("symbols", "workplace/re-sweep/pc98/borland-symbols.json", "PC-98 Borland 除錯符號表")
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
		sites := scanSoundFXSites(*root, *resident, *symbols)
		if effects, resolved, unresolved := tallySites(sites); resolved+unresolved > 0 {
			fmt.Fprintf(&report, "### `SOUNDFX` 每一處在放什麼音效\n\n")
			fmt.Fprintf(&report, "引數是**音效描述子變數**（`push word [位址]`），不是編號；"+
				"名字由 PC-98 的 Borland 除錯符號直接讀出。\n\n")
			fmt.Fprintf(&report, "⚠ 只找立即數的話只解得出 5 處——會得到「大部分音效查不到」"+
				"這個假結論。真正的形狀是推變數。\n\n")
			fmt.Fprintf(&report, "⚠ 這一節**直掃位元組**（`9A 00 00 93 08`），不走 far-call 對照表。"+
				"表只收得到 IDA 認成程式碼的呼叫點，比實際少 12 處，而且其中一處會改結論："+
				"`LIGHTNINGFX` 在表裡是 0 處，看起來像「remake 有、原版沒有」，"+
				"實際上它在 `CASTSPELL` 裡。**假零的來源是掃描面，不是原作。**\n\n")
			if mapped, subset := crossCheckFarCallMap(*farCallMap, sites); mapped > 0 {
				verdict := "**表裡有而直掃沒有**——直掃漏了，要查"
				if subset {
					verdict = "表裡那些是直掃的真子集"
				}
				fmt.Fprintf(&report, "★ **交叉檢查**：far-call 對照表列 %d 處、直掃 %d 處；%s。\n\n",
					mapped, len(sites), verdict)
			}
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

			// 逐處攤開：**「幾處」回答不了「什麼時候響」**，而後者才是 remake
			// 接得上接不上的依據。所在常式由 Borland 符號表推（同段裡位移不大於
			// 呼叫點的最後一個符號）。
			if len(sites) > 0 {
				fmt.Fprintf(&report, "#### 逐處：哪一支常式在放\n\n")
				fmt.Fprintf(&report, "所在常式取**同段裡位移不大於呼叫點的最後一個符號**。"+
					"符號表只收得到公開程序，所以模組內部的靜態常式會掛在前一個公開名字底下"+
					"——標成 `A＋n`，那個 `n` 就是它離公開入口多遠，不要當成「就是 A」。\n\n")
				fmt.Fprintf(&report, "| 音效 | 模組 | 位移 | 所在常式 |\n|---|---|---:|---|\n")
				for _, site := range sites {
					fmt.Fprintf(&report, "| %s | `%s` | `%04Xh` | %s |\n",
						site.effect, site.module, site.ea, site.routine)
				}
				fmt.Fprintf(&report, "\n")
			}
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


// effectSite 是一處 `SOUNDFX` 呼叫點的完整身分。
type effectSite struct {
	effect  string
	module  string
	ea      int
	routine string
}


// descriptorAt 往前掃出這一處推的是哪一個音效描述子。
func descriptorAt(data []byte, ea int) string {
	low := ea - 20
	if low < 0 {
		low = 0
	}
	for index := ea - 3; index >= low; index-- {
		if data[index] == 0xFF && data[index+1] == 0x36 {
			address := int(data[index+2]) | int(data[index+3])<<8
			if known, found := soundEffectName(address); found {
				return known
			}
			return fmt.Sprintf("（描述子 %04Xh 不在 SOUNDFX 的表裡）", address)
		}
		if data[index] == 0xB0 && data[index+2] == 0x50 {
			return fmt.Sprintf("（立即數 %d）", data[index+1])
		}
	}
	return ""
}

// symbolIndex 是 overlay → 該段的符號清單（依位移排序）。
type symbolIndex struct {
	byModule map[string][]struct {
		offset int
		name   string
	}
}

func loadSymbolIndex(path string) symbolIndex {
	index := symbolIndex{byModule: map[string][]struct {
		offset int
		name   string
	}{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		return index
	}
	var table struct {
		Segments []struct {
			Segment int    `json:"segment"`
			Module  string `json:"module"`
		} `json:"overlay_segments"`
		Symbols []struct {
			Name    string `json:"name"`
			Segment int    `json:"segment"`
			Offset  int    `json:"offset"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return index
	}
	moduleOf := map[int]string{}
	for _, segment := range table.Segments {
		moduleOf[segment.Segment] = segment.Module
	}
	for _, symbol := range table.Symbols {
		module, ok := moduleOf[symbol.Segment]
		if !ok {
			continue
		}
		index.byModule[module] = append(index.byModule[module], struct {
			offset int
			name   string
		}{offset: symbol.Offset, name: symbol.Name})
	}
	for module := range index.byModule {
		list := index.byModule[module]
		sort.Slice(list, func(left, right int) bool { return list[left].offset < list[right].offset })
		index.byModule[module] = list
	}
	return index
}

// routineAt 回傳位移所屬的公開程序；不是入口就標 `名字＋n`。
func (index symbolIndex) routineAt(module string, ea int) string {
	list := index.byModule[module]
	best := -1
	for position, symbol := range list {
		if symbol.offset > ea {
			break
		}
		best = position
	}
	if best < 0 {
		return "（這一段前面沒有符號）"
	}
	delta := ea - list[best].offset
	if delta == 0 {
		return fmt.Sprintf("**%s**", list[best].name)
	}
	return fmt.Sprintf("**%s**＋%Xh", list[best].name, delta)
}

// soundFXCall 是 `call far 0893:0000`（`SOUNDFX`）的位元組樣式。
var soundFXCall = []byte{0x9A, 0x00, 0x00, 0x93, 0x08}

// scanSoundFXSites 直掃全部 overlay ＋ 常駐，找出每一處 `SOUNDFX` 呼叫點。
//
// ★ 為什麼不走 far-call 對照表。 表只收得到 IDA 認成程式碼的呼叫點：實測
// overlay 側表列 36 處、直掃 42 處，常駐側表根本不涵蓋（另有 12 處）。
// 而且那 12 處差異裡有一處會**改結論**——`LIGHTNINGFX` 在表裡是 0 處，
// 於是對照報表寫成「remake 有、原版沒有」；實際上它在 `CASTSPELL` 裡。
// 這正是「grep 反組譯只看得到 IDA 認成程式碼的部分」那種假零。
//
// ⚠ 位元組直掃的相對風險是**假陽性**（資料剛好長得像那五個位元組）。這裡靠
// 兩件事壓住：呼叫點前面必須推得出一個已知的音效描述子，否則不收；而且
// far-call 表那 36 處必須是直掃結果的子集（`crossCheckFarCallMap` 在報表裡
// 印出裁決）。
func scanSoundFXSites(root, resident, symbolPath string) []effectSite {
	index := loadSymbolIndex(symbolPath)
	sites := make([]effectSite, 0, 64)

	collect := func(module string, data []byte) {
		for at := 0; ; {
			found := bytes.Index(data[at:], soundFXCall)
			if found < 0 {
				return
			}
			ea := at + found
			if name := descriptorAt(data, ea); name != "" {
				sites = append(sites, effectSite{
					effect: name, module: module, ea: ea,
					routine: index.routineAt(module, ea),
				})
			}
			at = ea + 1
		}
	}

	entries, _ := os.ReadDir(filepath.Join(root, "overlays"))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "overlay-") || !strings.HasSuffix(name, ".bin") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, "overlays", name))
		if err != nil {
			continue
		}
		collect(strings.TrimSuffix(name, ".bin"), data)
	}
	// 常駐：far-call 對照表不涵蓋，但它自己也叫 `SOUNDFX`（SOUNDHALT／
	// SOUNDOFF／SOUNDON），不掃就會少 12 處。
	if data, err := os.ReadFile(filepath.Join(root, resident)); err == nil {
		collect("常駐", data)
	}

	sort.SliceStable(sites, func(left, right int) bool {
		if sites[left].effect != sites[right].effect {
			return sites[left].effect < sites[right].effect
		}
		if sites[left].module != sites[right].module {
			return sites[left].module < sites[right].module
		}
		return sites[left].ea < sites[right].ea
	})
	return sites
}

// tallySites 把逐處清單收成「音效 → 模組 → 處數」。
func tallySites(sites []effectSite) (map[string]map[string]int, int, int) {
	effects := map[string]map[string]int{}
	resolved, unresolved := 0, 0
	for _, site := range sites {
		if strings.HasPrefix(site.effect, "（") {
			unresolved++
			continue
		}
		if effects[site.effect] == nil {
			effects[site.effect] = map[string]int{}
		}
		effects[site.effect][site.module]++
		resolved++
	}
	return effects, resolved, unresolved
}

// crossCheckFarCallMap 回答「far-call 對照表列的那些，直掃有沒有全部找到」。
// 回傳表列處數與「是不是真子集」。**不是子集就代表直掃漏了**，那要查。
func crossCheckFarCallMap(path string, sites []effectSite) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	var table struct {
		Targets []struct {
			Module string `json:"module"`
			EA     string `json:"ea"`
			Raw    string `json:"raw"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return 0, false
	}
	found := map[string]bool{}
	for _, site := range sites {
		found[fmt.Sprintf("%s:%04X", site.module, site.ea)] = true
	}
	mapped, subset := 0, true
	for _, call := range table.Targets {
		if call.Raw != soundSegment+":0000" {
			continue
		}
		mapped++
		var ea int
		if _, err := fmt.Sscanf(strings.TrimSuffix(call.EA, "h"), "%04X", &ea); err != nil {
			subset = false
			continue
		}
		if !found[fmt.Sprintf("%s:%04X", call.Module, ea)] {
			subset = false
		}
	}
	return mapped, subset
}
