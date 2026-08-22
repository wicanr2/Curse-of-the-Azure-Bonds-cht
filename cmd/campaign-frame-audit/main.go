// Command campaign-frame-audit 檢查「開場到結局」那條路線的每個劇情檢查點，
// 由**真的前端**畫出來的那一張長什麼樣。
//
// ★ 存在的理由：`TestRealNewGameRunsToTheEnding` 驅動的是 `*State`，畫面那一層
// 一次都沒被跑到。`remake-status` 的「開場到結局的同一 session」那一列因此寫著
// 「那是**測試路徑**；真人從開場玩到結局沒有自動化的證明」。
// `tools/campaign-frames.sh` 把每個檢查點交給 `cmd/azure-bonds-game` 畫一張，
// 這一支負責判讀那些 PNG。
//
// ⚠ **不要用「畫面有幾種顏色」判斷有沒有畫出來。** 一張正常的文字畫面就只有
// 三四種顏色，會被判成空白；而一張壞掉的畫面也可能顏色很多。這裡量的是
// **非背景像素怎麼分佈**：總量、佔了幾條掃描列、集中在哪。
//
// ⚠ 另一個看不出來的壞法是**字型沒載入**：畫面其他部分完全正常，只有每一個字
// 變成豆腐框。像素統計分不出豆腐框與真字，所以那一項要靠人看——報表把
// 「有多少列有字」印出來當線索，但不宣稱它能驗字型。
//
// 用法：
//
//	go run ./cmd/campaign-frame-audit -output docs/audit/campaign-frames.md
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// modeNames 是 `game.Mode` 的名字。存檔存的是數字，報表要讓人看得懂。
var modeNames = map[int]string{
	0: "標題", 1: "荒野", 2: "事件", 3: "地圖", 4: "場所",
	5: "戰鬥", 6: "手札", 7: "角色建立", 8: "地城",
}

type frame struct {
	name        string
	mode        int
	inDungeon   bool
	block       int
	currentCity int
	rendered    bool
	digest      string
	pixels      int
	rows        int
	coverage    float64
	verdict     string
}

func main() {
	saveDir := flag.String("saves", "workplace/campaign-frames/saves", "戰役檢查點存檔目錄")
	pngDir := flag.String("png", "workplace/campaign-frames/png", "前端畫出來的 PNG 目錄")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	entries, err := os.ReadDir(*saveDir)
	if err != nil {
		log.Fatalf("讀不到檢查點目錄 %s：%v（先跑 tools/campaign-frames.sh）", *saveDir, err)
	}
	frames := make([]frame, 0, 32)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		item := frame{name: name}
		if payload, readErr := os.ReadFile(filepath.Join(*saveDir, entry.Name())); readErr == nil {
			var save struct {
				Mode int `json:"mode"`
				Area struct {
					InDungeon           bool `json:"InDungeon"`
					Current3DMapBlockID int  `json:"Current3DMapBlockID"`
					CurrentCity         int  `json:"CurrentCity"`
				} `json:"area"`
			}
			if json.Unmarshal(payload, &save) == nil {
				item.mode = save.Mode
				item.inDungeon = save.Area.InDungeon
				item.block = save.Area.Current3DMapBlockID
				item.currentCity = save.Area.CurrentCity
			}
		}
		measure(&item, filepath.Join(*pngDir, name+".png"))
		frames = append(frames, item)
	}
	sort.Slice(frames, func(left, right int) bool { return frames[left].name < frames[right].name })

	rendered, blank := 0, 0
	for _, item := range frames {
		if item.rendered {
			rendered++
		}
		if item.verdict != "" {
			blank++
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 開場到結局：每個檢查點由**真的前端**畫出來長什麼樣\n\n")
	fmt.Fprintf(&report, "由 `cmd/campaign-frame-audit` 產生，不要手改。畫面由 "+
		"`tools/campaign-frames.sh` 用 `cmd/azure-bonds-game -party-load -screenshot` 產生。\n\n")
	fmt.Fprintf(&report, "★ 這一份補的是「**測試路徑不是玩家路徑**」那個缺口："+
		"`TestRealNewGameRunsToTheEnding` 驅動 `*State` 一路打到提朗瑟克斯，"+
		"但畫面那一層一次都沒被跑到。這裡讓玩家真正會執行的那支程式把每個檢查點載入並畫一張。\n\n")
	fmt.Fprintf(&report, "| 指標 | 數字 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 劇情檢查點 | %d |\n", len(frames))
	fmt.Fprintf(&report, "| 前端畫得出來 | %d |\n", rendered)
	fmt.Fprintf(&report, "| 判定為空白／幾乎空白 | %d |\n\n", blank)

	fmt.Fprintf(&report, "⚠ **「畫得出來」不等於「畫對了」。** 像素統計驗得到「不是空白」，"+
		"驗不到字型有沒有載入（少了 `-eten-font` 每個字都是豆腐框，而畫面其他部分完全正常）、"+
		"也驗不到畫的是不是**那一段**該有的畫面。\n\n")

	fmt.Fprintf(&report, "| 檢查點 | 模式 | 地城 | block | 非背景像素 | 佔用掃描列 | 覆蓋率 | 判定 |\n")
	fmt.Fprintf(&report, "|---|---|---|---:|---:|---:|---:|---|\n")
	for _, item := range frames {
		dungeon := "—"
		if item.inDungeon {
			dungeon = "是"
		}
		verdict := "✅"
		if !item.rendered {
			verdict = "**沒有畫面**"
		} else if item.verdict != "" {
			verdict = "**" + item.verdict + "**"
		}
		fmt.Fprintf(&report, "| %s | %s | %s | %d | %d | %d | %.1f%% | %s |\n",
			item.name, modeNames[item.mode], dungeon, item.block,
			item.pixels, item.rows, item.coverage*100, verdict)
	}
	fmt.Fprintf(&report, "\n")

	// 逐張畫面雜湊分組。**畫面一樣不一定是壞掉**——世界地圖那一張只由
	// `Area.CurrentCity` 決定（`drawOverlandMap` 標的是目前城市，不是 `MapX`／
	// `MapY`），所以同城市的兩個檢查點畫出來本來就會一模一樣。
	// 這一節把分組印出來並附上各自的 `CurrentCity`，讓人一眼看出是哪一種。
	groups := map[string][]frame{}
	for _, item := range frames {
		if item.digest == "" {
			continue
		}
		groups[item.digest] = append(groups[item.digest], item)
	}
	duplicates := make([]string, 0, 4)
	for digest, list := range groups {
		if len(list) > 1 {
			duplicates = append(duplicates, digest)
		}
	}
	sort.Strings(duplicates)
	fmt.Fprintf(&report, "## 畫出來一模一樣的檢查點\n\n")
	if len(duplicates) == 0 {
		fmt.Fprintf(&report, "（沒有。）\n")
	} else {
		fmt.Fprintf(&report, "⚠ **一樣不等於壞掉。** 世界地圖那一張只由 `Area.CurrentCity` 決定"+
			"（`drawOverlandMap` 標的是目前城市，不是 `MapX`／`MapY`），"+
			"所以同城市的兩個檢查點畫出來本來就一樣。"+
			"要判斷是不是缺陷，看下表的 `CurrentCity` 是不是也相同。\n\n")
		fmt.Fprintf(&report, "| 雜湊 | 檢查點 | 模式 | `CurrentCity` | 判讀 |\n|---|---|---|---:|---|\n")
		for _, digest := range duplicates {
			list := groups[digest]
			same := true
			for _, item := range list[1:] {
				if item.currentCity != list[0].currentCity {
					same = false
				}
			}
			note := "**同城市 ⇒ 畫面本來就該一樣**"
			if !same {
				note = "**城市不同卻畫出同一張——要查**"
			}
			names := make([]string, 0, len(list))
			modes := make([]string, 0, len(list))
			for _, item := range list {
				names = append(names, item.name)
				modes = append(modes, modeNames[item.mode])
			}
			fmt.Fprintf(&report, "| `%s` | %s | %s | %d | %s |\n",
				digest, strings.Join(names, "<br>"), strings.Join(modes, "／"),
				list[0].currentCity, note)
		}
	}
	fmt.Fprintf(&report, "\n")

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "checkpoints=%d rendered=%d blank=%d\n", len(frames), rendered, blank)
}

// measure 量一張畫面的非背景像素分佈。
//
// 背景色取**出現最多的那個顏色**，而不是寫死深藍：地城與戰鬥畫面的底色不一樣，
// 寫死一個值會把整張地城圖算成「全部都是內容」。
func measure(item *frame, path string) {
	handle, err := os.Open(path)
	if err != nil {
		return
	}
	defer handle.Close()
	decoded, _, err := image.Decode(handle)
	if err != nil {
		return
	}
	item.rendered = true
	if raw, readErr := os.ReadFile(path); readErr == nil {
		item.digest = fmt.Sprintf("%x", sha256.Sum256(raw))[:8]
	}
	bounds := decoded.Bounds()
	counts := map[uint32]int{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			counts[packed(decoded, x, y)]++
		}
	}
	background, best := uint32(0), -1
	for value, count := range counts {
		if count > best {
			background, best = value, count
		}
	}
	rows := 0
	total := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		rowCount := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if packed(decoded, x, y) != background {
				rowCount++
			}
		}
		if rowCount > 0 {
			rows++
		}
		total += rowCount
	}
	item.pixels = total
	item.rows = rows
	area := bounds.Dx() * bounds.Dy()
	if area > 0 {
		item.coverage = float64(total) / float64(area)
	}
	// ⚠ 門檻刻意寬鬆：這一支要抓的是「整張什麼都沒有」，不是「畫得不夠好」。
	// 一張只有幾行字的正常畫面覆蓋率就是個位數百分比。
	switch {
	case total == 0:
		item.verdict = "整張單色"
	case rows < 8:
		item.verdict = fmt.Sprintf("只有 %d 條掃描列有東西", rows)
	}
}

func packed(source image.Image, x, y int) uint32 {
	r, g, b, a := source.At(x, y).RGBA()
	return uint32(r>>8)<<24 | uint32(g>>8)<<16 | uint32(b>>8)<<8 | uint32(a>>8)
}
