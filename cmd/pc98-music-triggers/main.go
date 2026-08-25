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
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

// mscPlayOffset 是 `MSCPLAY` 在 `SOUNDX` 段內的位移（Borland 符號表）。
const mscPlayOffset = 0x0114

// musicNoCell 是 PC-98 的 `MUSICNO`（Borland 符號表直接讀出）。
const musicNoCell = 0x8BF3

// soundSegment 是 `SOUNDX` 那一組常式所在的程式碼段。Borland 符號表把
// `SOUNDFX` 記在 segment 2195 ＝ `0893h`，而 PC-98 的 far-call 表裡正好有一個
// 目標段 `0893h`——兩邊對得上。段內位移也直接取自符號表。
const soundSegment = "0893"

// soundSegmentNumber 是同一個段在 Borland 符號表裡的十進位編號。
const soundSegmentNumber = 0x0893

// soundEffectLabels 給每個 Borland 符號一個中文說明。**名字與位址都不在這裡**
// ——那兩樣由 `pc98sfx.Selectors()` 推出來（描述子位址 ＝ 基底 ＋ 選擇子×2）。
//
// ⚠ 這裡以前是一張手打的「位址 → 名字」對照表，而它**漏了 4840h 與 4842h**
// （MISSFX／SPELLHITFX）。漏掉不會報錯：報表把那兩處印成「符號表沒有」，
// 讀起來像原作真的少了兩個名字，而下游的 `cmd/sound-trigger-compare` 連帶
// 少了兩列——正好是 remake 用得最兇的 `SoundSpellHit` 與 `SoundMiss`。
// 位址改成用推的之後，「漏一格」在結構上就不可能發生。
var soundEffectLabels = map[string]string{
	"SOUNDHALT": tooltext.Text("h.ca4d973c0b00"), "SOUNDOFF": tooltext.Text("h.94ffcbc4a678"), "SOUNDON": tooltext.Text("h.320d1ab4f80f"),
	"CASTFX": tooltext.Text("h.e377c6152522"), "MISSFX": tooltext.Text("h.d7aa5b20173e"), "SPELLHITFX": tooltext.Text("h.0118c5026c62"),
	"DEADFX": tooltext.Text("h.82d3130fa582"), "WHISTLEFX": tooltext.Text("h.d0928210330c"), "HITFX": tooltext.Text("h.393df9bb13ea"),
	"LIGHTNINGFX": tooltext.Text("h.48693c8fd90b"), "SWISHFX": tooltext.Text("h.692fbefa6024"), "PADFX": tooltext.Text("h.9f18c28a4e63"),
	"FIREBALLFX": tooltext.Text("h.010568a02690"), "ARROWFX": tooltext.Text("h.674650b36c70"), "OVERTUREFX": tooltext.Text("h.3df5903dbdcd"),
	"COMBATFX": tooltext.Text("h.625dd417c2c3"), "CRASHFX": tooltext.Text("h.b15e09c78fd2"),
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

// soundRoutineLabels 只給中文說明。**名字與位移一律從 Borland 符號表讀**
// （`loadSoundRoutines`），不在這裡抄第二份。
//
// ⚠ 這裡以前是一張手打的「位移 → 名字」表，只有四格，於是 `015Eh` 被印成
// 「符號表沒有這一格」——而它就在符號表裡，叫 `MSCSTOP`。這是同一個 session
// 裡**第三次**踩到同一種錯（音效描述子表漏兩格、對照表漏兩列、這裡漏一格）：
// 手打的對照表漏掉時不會報錯，只會印出一句看起來很有資訊量的「查無」。
var soundRoutineLabels = map[string]string{
	"SOUNDFX": tooltext.Text("h.5e9d80d0e3c3"), "INITSOUND": tooltext.Text("h.65622e8ee9fb"), "MSCPLAY": tooltext.Text("h.a9ddbaf5be7f"),
	"MSCSTOP": tooltext.Text("h.90a41433b9d4"), "BGMPLAY": tooltext.Text("h.fb7df1d56430"),
}

// loadSoundRoutines 從 Borland 符號表讀出 `SOUNDX` 那一段的全部公開程序。
func loadSoundRoutines(path string, segment int) map[int]string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var table struct {
		Symbols []struct {
			Name    string `json:"name"`
			Segment int    `json:"segment"`
			Offset  int    `json:"offset"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil
	}
	routines := map[int]string{}
	for _, symbol := range table.Symbols {
		if symbol.Segment != segment {
			continue
		}
		if label, ok := soundRoutineLabels[symbol.Name]; ok {
			routines[symbol.Offset] = fmt.Sprintf("%s（%s）", symbol.Name, label)
			continue
		}
		routines[symbol.Offset] = symbol.Name
	}
	return routines
}

type trigger struct {
	module   string
	unit     string
	form     string
	offset   int
	selector int
}

func main() {
	root := flag.String("root", "workplace/re-sweep/pc98", tooltext.Text("h.a51d2488a6c1"))
	resident := flag.String("resident", "PC98-GAME.EXE", tooltext.Text("h.c33c386f5bc7"))
	corePath := flag.String("core", "gamepack/pack/00-core.json", tooltext.Text("h.e93e76cafe80"))
	localePath := flag.String("locale", "gamepack/pack/20-locale.zh-TW.json", tooltext.Text("h.18a8248f0745"))
	farCallMap := flag.String("far-call-map", "docs/audit/far-call-map-pc98.json", tooltext.Text("h.441d3f0b84bd"))
	symbols := flag.String("symbols", "workplace/re-sweep/pc98/borland-symbols.json", tooltext.Text("h.d3ff39c17343"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
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
				module: name, unit: units[name], offset: offset,
				selector: int(data[offset+4]), form: "`mov byte [MUSICNO], imm`"})
		}
		// ★ 第二種形狀：**直接把曲號推給 `MSCPLAY`**，完全不碰 `MUSICNO`。
		//
		// ⚠ 只認第一種的話會漏掉戰鬥音樂：`INITCOMBAT`（COMPREP）用的是
		// `mov al, imm / push ax / call MSCPLAY`，於是「戰鬥」與「地城二」兩首
		// 看起來像**沒有任何地方選到**——而報表的警告雖然寫著「不能讀成用不到」，
		// 也只能停在那裡。掃描面窄一格，兩首曲子就整個消失。
		for offset := 0; offset+8 <= len(data); offset++ {
			if data[offset] != 0xB0 || data[offset+2] != 0x50 {
				continue
			}
			if data[offset+3] != 0x9A {
				continue
			}
			if int(binary.LittleEndian.Uint16(data[offset+4:])) != mscPlayOffset {
				continue
			}
			if int(binary.LittleEndian.Uint16(data[offset+6:])) != soundSegmentNumber {
				continue
			}
			triggers = append(triggers, trigger{
				module: name, unit: units[name], offset: offset + 3,
				selector: int(data[offset+1]), form: "`push imm` → `MSCPLAY`"})
		}
	}
	sort.SliceStable(triggers, func(left, right int) bool {
		if triggers[left].module != triggers[right].module {
			return triggers[left].module < triggers[right].module
		}
		return triggers[left].offset < triggers[right].offset
	})

	used := map[int]bool{}
	for _, item := range triggers {
		used[item.selector] = true
	}

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.2fb54fab6005"))
	fmt.Fprint(&report, tooltext.Format("h.c5036ad81cf9"))
	fmt.Fprint(&report, tooltext.Text("h.d42b996ab403")+
		tooltext.Text("h.c695d120f5c2")+
		tooltext.Text("h.a52b410b22b7")+
		tooltext.Text("h.73b69f9c4fac"))
	fmt.Fprint(&report, tooltext.Text("h.587276721795")+
		tooltext.Text("h.f92eb9453f3b"))

	fmt.Fprint(&report, tooltext.Format("h.91130546f2ec"))
	byForm := map[string]int{}
	for _, item := range triggers {
		byForm[item.form]++
	}
	fmt.Fprint(&report, tooltext.Format("h.7b799cde9e3a", byForm["`mov byte [MUSICNO], imm`"]))
	fmt.Fprint(&report, tooltext.Format("h.ce1a2c9705ed", byForm["`push imm` → `MSCPLAY`"]))
	fmt.Fprint(&report, tooltext.Format("h.9f9c0365b7bf", len(triggers)))
	fmt.Fprint(&report, tooltext.Format("h.63b44d9315a0", len(used)))
	fmt.Fprint(&report, tooltext.Format("h.37c6d52044b1", len(selectors)))

	fmt.Fprint(&report, tooltext.Format("h.83e64ff0b57a"))
	for _, item := range triggers {
		unit := item.unit
		if unit == "" {
			unit = "—"
		}
		fmt.Fprintf(&report, "| `%s` | %s | `%04Xh` | %s | %d | %s |\n",
			item.module, unit, item.offset, item.form, item.selector,
			titleFor(titles, selectors, item.selector))
	}

	// 播放常式的呼叫點：音效那一半在這裡。
	soundRoutines := loadSoundRoutines(*symbols, soundSegmentNumber)
	if routines, total := soundCallSites(*root, *resident); total > 0 {
		fmt.Fprint(&report, tooltext.Format("h.1a332a3387a0"))
		fmt.Fprint(&report, tooltext.Format("h.df84acefda75", soundSegment))
		fmt.Fprint(&report, tooltext.Format("h.9b07ea04d8a4"))
		offsets := make([]int, 0, len(routines))
		for offset := range routines {
			offsets = append(offsets, offset)
		}
		sort.Ints(offsets)
		for _, offset := range offsets {
			name, known := soundRoutines[offset]
			if !known {
				name = tooltext.Text("h.788f39d8378b")
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
		fmt.Fprint(&report, tooltext.Format("h.d1c34d08a667", total))
		sites := scanSoundFXSites(*root, *resident, *symbols)
		if effects, resolved, unresolved := tallySites(sites); resolved+unresolved > 0 {
			fmt.Fprint(&report, tooltext.Format("h.a2687669cf08"))
			fmt.Fprint(&report, tooltext.Text("h.47aecdbf90b9")+
				tooltext.Text("h.81c9c51d2862"))
			fmt.Fprint(&report, tooltext.Text("h.3064ea53271c")+
				tooltext.Text("h.0981e4937a0a"))
			fmt.Fprint(&report, tooltext.Text("h.9ee0225ba472")+
				tooltext.Text("h.7ccda174345f")+
				tooltext.Text("h.2d17f50b39e3")+
				tooltext.Text("h.d860808f8768"))
			if mapped, subset := crossCheckFarCallMap(*farCallMap, sites); mapped > 0 {
				verdict := tooltext.Text("h.53529448ec78")
				if subset {
					verdict = tooltext.Text("h.10fc258e58a2")
				}
				fmt.Fprint(&report, tooltext.Format("h.401dbe4b35a1", mapped, len(sites), verdict))
			}
			fmt.Fprint(&report, tooltext.Format("h.5198002eaba7"))
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
			fmt.Fprint(&report, tooltext.Format("h.837f8054fad9", resolved, unresolved))

			// 逐處攤開：**「幾處」回答不了「什麼時候響」**，而後者才是 remake
			// 接得上接不上的依據。所在常式由 Borland 符號表推（同段裡位移不大於
			// 呼叫點的最後一個符號）。
			if len(sites) > 0 {
				fmt.Fprint(&report, tooltext.Format("h.42b1afe221f7"))
				fmt.Fprint(&report, tooltext.Text("h.8a764bd0b916")+
					tooltext.Text("h.7715c0ba97b5")+
					tooltext.Text("h.eae01ecdd928"))
				fmt.Fprint(&report, tooltext.Format("h.b27432477059"))
				for _, site := range sites {
					fmt.Fprintf(&report, "| %s | `%s` | `%04Xh` | %s |\n",
						site.effect, site.module, site.ea, site.routine)
				}
				fmt.Fprintf(&report, "\n")
			}
		}
		// ★ **「戰鬥進行中會不會換曲」是這張表答得出來的問題**，而且答案是不會。
		// 一定要配正對照才算數：同一次掃描在戰鬥模組裡找得到 `SOUNDFX`，
		// 所以掃描面涵蓋得到那幾個 overlay——0 是原作的 0，不是掃描的假零。
		{
			const mscplay = 0x0114
			combat := map[string]string{
				"overlay-08": "COMBAT", "overlay-13": "COMSTUFF", "overlay-32": "TACMAP",
			}
			music, effects := 0, 0
			for module, count := range routines[mscplay] {
				if _, ok := combat[module]; ok {
					music += count
				}
			}
			for _, site := range sites {
				if _, ok := combat[site.module]; ok {
					effects++
				}
			}
			names := make([]string, 0, len(combat))
			for module, unit := range combat {
				names = append(names, fmt.Sprintf("`%s` %s", module, unit))
			}
			sort.Strings(names)
			fmt.Fprint(&report, tooltext.Format("h.acaff94b123b"))
			fmt.Fprint(&report, tooltext.Format("h.3f1673ee4f58", strings.Join(names, "、")))
			fmt.Fprint(&report, tooltext.Format("h.39c330dd9df7"))
			fmt.Fprint(&report, tooltext.Format("h.7cd06265c23b", music))
			fmt.Fprint(&report, tooltext.Format("h.574f2a56da72", effects))
			switch {
			case music == 0 && effects > 0:
				fmt.Fprint(&report, tooltext.Text("h.c076a2b3ac95")+
					tooltext.Text("h.f8f56cd24f2e")+
					tooltext.Text("h.cc4c458d86e1"))
				fmt.Fprintf(&report, tooltext.Text("h.2be47f1e1b66")+
					tooltext.Text("h.02ec25083533"), effects)
			case music > 0:
				fmt.Fprint(&report, tooltext.Format("h.192b449ff84a", music))
			default:
				fmt.Fprint(&report, tooltext.Text("h.398631ba0315")+
					tooltext.Text("h.49885adebe8f"))
			}
			fmt.Fprint(&report, tooltext.Text("h.b0fae69041c5")+
				tooltext.Text("h.9caf0e04a334"))
		}
		// ★ 交叉印證要**算出來**，不能寫死。 兩次獨立的掃描（換曲點 vs
		// `MSCPLAY` 呼叫點）指到哪些模組是資料，寫死的話資料變了句子不會變——
		// 這一句原本寫著「正好落在那五個改寫 `MUSICNO` 的 overlay 上」，
		// 而改用位元組直掃之後 `MSCPLAY` 多出 `overlay-10`（COMPREP），
		// 那句話就成了假的。
		if callers := routines[mscPlayOffset]; len(callers) > 0 {
			triggerModules := map[string]bool{}
			for _, item := range triggers {
				triggerModules[item.module] = true
			}
			both, onlyPlay := []string{}, []string{}
			for module := range callers {
				if triggerModules[module] {
					both = append(both, module)
				} else {
					onlyPlay = append(onlyPlay, module)
				}
			}
			sort.Strings(both)
			sort.Strings(onlyPlay)
			fmt.Fprintf(&report, tooltext.Text("h.91faae770999")+
				tooltext.Text("h.d7fb57b06437"),
				len(both), strings.Join(both, "、"))
			if len(onlyPlay) > 0 {
				fmt.Fprintf(&report, tooltext.Text("h.32012bf3fdf3")+
					tooltext.Text("h.c8d9d8bf2e80")+
					tooltext.Text("h.fe8b176bd295"),
					len(onlyPlay), strings.Join(onlyPlay, "、"))
			}
		}
		fmt.Fprint(&report, tooltext.Text("h.8d1ee822633e")+
			tooltext.Text("h.ad19403087ca")+
			tooltext.Text("h.07c6e0c180e1"))
	}

	fmt.Fprint(&report, tooltext.Format("h.22ea335e446f"))
	missing := make([]string, 0, 4)
	for selector := range selectors {
		if !used[selector] {
			missing = append(missing, fmt.Sprintf("%d（%s）", selector, titleFor(titles, selectors, selector)))
		}
	}
	sort.Strings(missing)
	if len(missing) == 0 {
		fmt.Fprint(&report, tooltext.Format("h.54f42cb556f8", len(selectors)))
		fmt.Fprint(&report, tooltext.Text("h.ce900d389385")+
			tooltext.Text("h.16516383f7cb")+
			tooltext.Text("h.b8c8ccdb862e")+
			tooltext.Text("h.6527445c688f"))
		fmt.Fprint(&report, "```\n"+
			"cmp byte [LOADMONNUM], 47h\n"+
			tooltext.Text("h.2772b583a21f")+
			tooltext.Text("h.31f4a8dc7609")+
			"push ax / call MSCPLAY\n```\n\n")
		fmt.Fprint(&report, tooltext.Text("h.d9445a32c88e")+
			tooltext.Text("h.5f4074f7ccd5"))
	} else {
		fmt.Fprintf(&report, "%s\n\n", strings.Join(missing, "、"))
		fmt.Fprint(&report, tooltext.Text("h.b44aedfd37b9")+
			tooltext.Text("h.9740ebf5d834"))
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
		return tooltext.Format("h.74e4d1121349", selector)
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

// soundCallSites 統計 `SOUNDX` 那一段每個入口各被叫幾次。
//
// ⚠ **位元組直掃，不走 far-call 對照表。** 表只收得到 IDA 認成程式碼的呼叫點，
// 也不涵蓋常駐：實測 `MSCPLAY` 表列 5 處、直掃 **7** 處（多的兩處在 overlay-10），
// `SOUNDFX` 表列 36 處、直掃 54 處。**下界看起來和全集一樣合理**，所以這裡不用它。
func soundCallSites(root, resident string) (map[int]map[string]int, int) {
	var segment int
	if _, err := fmt.Sscanf(soundSegment, "%04X", &segment); err != nil {
		return nil, 0
	}
	scan := func(name string, data []byte, routines map[int]map[string]int) int {
		count := 0
		for ea := 0; ea+4 < len(data); ea++ {
			if data[ea] != 0x9A {
				continue
			}
			if int(binary.LittleEndian.Uint16(data[ea+3:])) != segment {
				continue
			}
			offset := int(binary.LittleEndian.Uint16(data[ea+1:]))
			if routines[offset] == nil {
				routines[offset] = map[string]int{}
			}
			routines[offset][name]++
			count++
		}
		return count
	}
	routines := map[int]map[string]int{}
	total := 0
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
		total += scan(strings.TrimSuffix(name, ".bin"), data, routines)
	}
	if data, err := os.ReadFile(filepath.Join(root, resident)); err == nil {
		total += scan(tooltext.Text("h.4671035ec213"), data, routines)
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
			return tooltext.Format("h.a0c98a430541", address)
		}
		if data[index] == 0xB0 && data[index+2] == 0x50 {
			return tooltext.Format("h.db85e3591778", data[index+1])
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
		return tooltext.Text("h.50083effc596")
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
		collect(tooltext.Text("h.4671035ec213"), data)
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
