// Command segment-seams 回答「分段驗收的**接縫**對得上嗎」。
//
// ★ 存在的理由：使用者 2026-08-16 決定分階段驗收算數——每一段用 debug 進入點
// 直入、各自對 reference 驗證無誤即算該段完成，不必跑一次連續全程。但同一條
// 決定還寫著：
//
//	⚠ 段與段之間的狀態交接本身就是一段。debug 旗標注入的是合成起始狀態，
//	  未必等於上一段真的跑出來的結束狀態；兩端都綠不等於接縫通過。
//
// 這一支把那句話變成數字。`internal/segment` 的 `EnterFrom` 是**宣告**——直接
// 進入時寫進 `4BF2h`（LastECL）的值；主線跑一次錄下來的段落轉移是**事實**。
// 兩者不一致的段，直入時跑的 initial lifecycle 就和主線實際到達時不是同一條路。
//
// ⚠ 這**不是**「接縫通過了」的證明，是接縫的**分母**：它只比 LastECL 這一個
// 欄位。隊伍、旗標、攜帶物、ECL 記憶體的交接還沒有對照——那些要另外做。
// 但 LastECL 是**唯一一個直入時被明文合成出來的**，所以它先。
//
// 用法：
//
//	COAB_BLOCK_EDGES=/src/workplace/campaign-frames/block-edges.json \
//	    tools/go.sh test ./internal/game/ -run TestRealNewGameRunsToTheEnding -count=1
//	go run ./cmd/segment-seams -edges workplace/campaign-frames/block-edges.json \
//	    -output docs/audit/segment-seams.md -json docs/audit/segment-seams.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

type edge struct {
	From  uint8 `json:"from"`
	To    uint8 `json:"to"`
	Count int   `json:"count"`
}

// seamRow 是一段的接縫狀態。
type seamRow struct {
	ID string `json:"id"`
	// Declared 是註冊表宣告的前一段（直入時寫進 LastECL 的值）。
	Declared uint8 `json:"declared_enter_from"`
	// Observed 是主線實際從哪些段走進來的。
	Observed []uint8 `json:"observed_from,omitempty"`
	// Verdict：`match`／`mismatch`／`fresh-start`／`not-visited`。
	Verdict string `json:"verdict"`
}

type report struct {
	Schema string `json:"schema"`
	// Segments 是註冊表裡的段數（分母）。
	Segments int `json:"segments"`
	// Visited 是主線真的到過的段數。
	Visited int `json:"visited"`
	// Matched 是「宣告的前一段就是主線實際的前一段之一」。
	Matched int `json:"matched"`
	// Mismatched 是主線到過、但實際的前一段和宣告的不同。
	Mismatched int `json:"mismatched"`
	// FreshStart 是宣告 `0x00`（全新開局）的段，沒有前一段可比。
	FreshStart int `json:"fresh_start"`
	// NotVisited 是主線一次都沒走到的段。
	NotVisited int `json:"not_visited"`
	// Undeclared 是主線走出來、但沒有任何段宣告的轉移數。
	Undeclared int       `json:"undeclared_edges"`
	Rows       []seamRow `json:"rows"`
}

func main() {
	edgesPath := flag.String("edges", "workplace/campaign-frames/block-edges.json",
		"主線錄下來的段落轉移")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	outputJSON := flag.String("json", "", "JSON 輸出路徑，給 `cmd/remake-status` 取用")
	flag.Parse()

	raw, err := os.ReadFile(*edgesPath)
	if err != nil {
		log.Fatalf("讀不到段落轉移 %s：%v\n"+
			"先跑：COAB_BLOCK_EDGES=/src/%s tools/go.sh test ./internal/game/ "+
			"-run TestRealNewGameRunsToTheEnding -count=1", *edgesPath, err, *edgesPath)
	}
	var edges []edge
	if err := json.Unmarshal(raw, &edges); err != nil {
		log.Fatal(err)
	}
	if len(edges) == 0 {
		log.Fatal("段落轉移是空的：錄製沒有開起來")
	}

	// 主線實際走進每一段的來源。
	arrivals := map[uint8]map[uint8]bool{}
	for _, one := range edges {
		if arrivals[one.To] == nil {
			arrivals[one.To] = map[uint8]bool{}
		}
		arrivals[one.To][one.From] = true
	}

	doc := report{Schema: "coab-segment-seams/1"}
	declared := map[[2]uint8]bool{}
	for _, item := range segment.All() {
		doc.Segments++
		row := seamRow{ID: item.ID, Declared: item.EnterFrom}
		declared[[2]uint8{item.EnterFrom, item.Block}] = true
		from, visited := arrivals[item.Block]
		if visited {
			doc.Visited++
			for source := range from {
				row.Observed = append(row.Observed, source)
			}
			sort.Slice(row.Observed, func(l, r int) bool { return row.Observed[l] < row.Observed[r] })
		}
		row.Verdict = classify(item.EnterFrom, visited, from)
		switch row.Verdict {
		case "fresh-start":
			doc.FreshStart++
		case "not-visited":
			doc.NotVisited++
		case "match":
			doc.Matched++
		default:
			doc.Mismatched++
		}
		doc.Rows = append(doc.Rows, row)
	}
	// 主線走出來、卻沒有任何段宣告的轉移。
	undeclared := make([]edge, 0, len(edges))
	for _, one := range edges {
		if !declared[[2]uint8{one.From, one.To}] {
			undeclared = append(undeclared, one)
		}
	}
	doc.Undeclared = len(undeclared)

	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(doc, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	text := render(doc, undeclared, len(edges))
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "segments=%d visited=%d match=%d mismatch=%d fresh=%d unvisited=%d undeclared=%d\n",
		doc.Segments, doc.Visited, doc.Matched, doc.Mismatched,
		doc.FreshStart, doc.NotVisited, doc.Undeclared)
}

// classify 判一段的接縫狀態。
//
// ⚠ 順序有意義：`fresh-start` 要排在 `not-visited` 前面。宣告 `00h` 的段是
// **全新開局**，它本來就不會有「前一段」——把它算成「主線沒走到」會讓開場那兩段
// 看起來像沒驗到，而實際上它們是這條路的起點。
func classify(declared uint8, visited bool, from map[uint8]bool) string {
	switch {
	case declared == 0x00:
		return "fresh-start"
	case !visited:
		return "not-visited"
	case from[declared]:
		return "match"
	default:
		return "mismatch"
	}
}

func render(doc report, undeclared []edge, total int) string {
	var out strings.Builder
	out.WriteString("# 分段驗收的接縫：直入宣告的前一段，和主線實際的前一段對得上嗎\n\n")
	out.WriteString("由 `cmd/segment-seams` 產生，不要手改。\n\n")
	out.WriteString("★ 使用者 2026-08-16 決定分階段驗收算數：每一段用 debug 進入點直入、" +
		"各自對 reference 驗證無誤即算該段完成，不必跑一次連續全程。同一條決定也寫著" +
		"**段與段之間的狀態交接本身就是一段**——直入注入的是合成起始狀態，" +
		"未必等於上一段真的跑出來的結束狀態。這一份就是那句話的分母。\n\n")
	out.WriteString("⚠ **只比 `LastECL` 這一個欄位**（`4BF2h`，直入時寫進去的 `EnterFrom`）。" +
		"隊伍、旗標、攜帶物、ECL 記憶體的交接還沒有對照。選它先是因為它是" +
		"**唯一一個直入時被明文合成出來的**，其餘欄位直入是沿用當下的狀態。\n\n")
	out.WriteString("⚠ 主線的轉移是**下界**：只錄得到經過換段派曲那個瓶頸的轉移。" +
		"漏掉的轉移音樂也會跟著錯，所以這兩件事共用同一個訊號、不是各自獨立的假設。\n\n")
	fmt.Fprintf(&out, "| 段 | 宣告的前一段 | 主線實際的前一段 | 結果 |\n|---|---|---|---|\n")
	for _, row := range doc.Rows {
		observed := "—"
		if len(row.Observed) > 0 {
			parts := make([]string, 0, len(row.Observed))
			for _, one := range row.Observed {
				parts = append(parts, fmt.Sprintf("`%02Xh`", one))
			}
			observed = strings.Join(parts, "、")
		}
		fmt.Fprintf(&out, "| `%s` | `%02Xh` | %s | %s |\n",
			row.ID, row.Declared, observed, verdictText(row.Verdict))
	}
	fmt.Fprintf(&out, "\n合計 %d 段：對得上 **%d**、對不上 **%d**、"+
		"全新開局 %d（沒有前一段可比）、主線沒走到 %d。\n\n",
		doc.Segments, doc.Matched, doc.Mismatched, doc.FreshStart, doc.NotVisited)

	fmt.Fprintf(&out, "## 主線走出來、卻沒有任何段宣告的轉移（%d／%d）\n\n",
		len(undeclared), total)
	out.WriteString("⚠ 這些**不是錯誤**：一段可以有好幾個入口，而 `EnterFrom` 只挑" +
		"「最順著劇情的那一個」。列出來是因為**直入永遠走不到這些入口**——" +
		"如果某一段的行為會因為從哪裡進來而不同，那個差異在分段驗收裡看不見。\n\n")
	if len(undeclared) == 0 {
		out.WriteString("（沒有。）\n")
		return out.String()
	}
	out.WriteString("| 從 | 到 | 主線走過幾次 |\n|---|---|---:|\n")
	for _, one := range undeclared {
		fmt.Fprintf(&out, "| `%02Xh` | `%02Xh` | %d |\n", one.From, one.To, one.Count)
	}
	return out.String()
}

func verdictText(verdict string) string {
	switch verdict {
	case "match":
		return "✅ 對得上"
	case "mismatch":
		return "⚠ **對不上**"
	case "fresh-start":
		return "— 全新開局（`00h`）"
	default:
		return "— 主線沒走到"
	}
}
