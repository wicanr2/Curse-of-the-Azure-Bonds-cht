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
