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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
	0: tooltext.Text("h.6fe38ed1ee10"), 1: tooltext.Text("h.b549e3f43fcb"), 2: tooltext.Text("h.c560201b331c"), 3: tooltext.Text("h.90e1b1b8a537"), 4: tooltext.Text("h.828d4cec1f3c"),
	5: tooltext.Text("h.625dd417c2c3"), 6: tooltext.Text("h.53ce36b5da9e"), 7: tooltext.Text("h.f7865e52f8b7"), 8: tooltext.Text("h.c014266366bb"),
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
	saveDir := flag.String("saves", "workplace/campaign-frames/saves", tooltext.Text("h.bf6191f67bbd"))
	pngDir := flag.String("png", "workplace/campaign-frames/png", tooltext.Text("h.ad9578556818"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	flag.Parse()

	entries, err := os.ReadDir(*saveDir)
	if err != nil {
		log.Fatal(tooltext.Format("h.32f913ef9506", *saveDir, err))
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
	fmt.Fprint(&report, tooltext.Format("h.d2ff81ecba32"))
	fmt.Fprint(&report, tooltext.Text("h.863aa9d1e1f3")+
		tooltext.Text("h.c56d6264d1fe"))
	fmt.Fprint(&report, tooltext.Text("h.43420567f3c2")+
		tooltext.Text("h.9f824029e160")+
		tooltext.Text("h.40b7577514ac"))
	fmt.Fprint(&report, tooltext.Format("h.13c83a8a875e"))
	fmt.Fprint(&report, tooltext.Format("h.1699a61010f4", len(frames)))
	fmt.Fprint(&report, tooltext.Format("h.f95bb44c1835", rendered))
	fmt.Fprint(&report, tooltext.Format("h.d5a867e544ae", blank))

	fmt.Fprint(&report, tooltext.Text("h.6d8daf9e80f3")+
		tooltext.Text("h.9ed82e0b9822")+
		tooltext.Text("h.a4eaa55ec852"))

	fmt.Fprint(&report, tooltext.Format("h.50a6361592c4"))
	fmt.Fprintf(&report, "|---|---|---|---:|---:|---:|---:|---|\n")
	for _, item := range frames {
		dungeon := "—"
		if item.inDungeon {
			dungeon = tooltext.Text("h.b5141d3d19e9")
		}
		verdict := "✅"
		if !item.rendered {
			verdict = tooltext.Text("h.f785cc62d4e2")
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
	fmt.Fprint(&report, tooltext.Format("h.2221fe2d902e"))
	if len(duplicates) == 0 {
		fmt.Fprint(&report, tooltext.Format("h.be6c71d35815"))
	} else {
		fmt.Fprint(&report, tooltext.Text("h.595c18804d34")+
			tooltext.Text("h.2d1be78f79d9")+
			tooltext.Text("h.1db2a580b6d1")+
			tooltext.Text("h.dfdc72ae9122"))
		fmt.Fprint(&report, tooltext.Format("h.902cf4104bcb"))
		for _, digest := range duplicates {
			list := groups[digest]
			same := true
			for _, item := range list[1:] {
				if item.currentCity != list[0].currentCity {
					same = false
				}
			}
			note := tooltext.Text("h.5a99f7e301be")
			if !same {
				note = tooltext.Text("h.4bdbae6b818d")
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
		item.verdict = tooltext.Text("h.048744588782")
	case rows < 8:
		item.verdict = tooltext.Format("h.78ba8371c616", rows)
	}
}

func packed(source image.Image, x, y int) uint32 {
	r, g, b, a := source.At(x, y).RGBA()
	return uint32(r>>8)<<24 | uint32(g>>8)<<16 | uint32(b>>8)<<8 | uint32(a>>8)
}
