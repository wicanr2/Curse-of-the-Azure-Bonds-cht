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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
		tooltext.Text("h.88cb80b6da27"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	outputJSON := flag.String("json", "", tooltext.Text("h.fa63314394ca"))
	flag.Parse()

	raw, err := os.ReadFile(*edgesPath)
	if err != nil {
		log.Fatalf(tooltext.Text("h.4d7cc47cc5c0")+
			tooltext.Text("h.1ce8361a318f")+
			"-run TestRealNewGameRunsToTheEnding -count=1", *edgesPath, err, *edgesPath)
	}
	var edges []edge
	if err := json.Unmarshal(raw, &edges); err != nil {
		log.Fatal(err)
	}
	if len(edges) == 0 {
		log.Fatal(tooltext.Text("h.6221d232aba8"))
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
	out.WriteString(tooltext.Text("h.1947760486e0"))
	out.WriteString(tooltext.Text("h.47fdb6333d38"))
	out.WriteString(tooltext.Text("h.aa25e8479159") +
		tooltext.Text("h.247561b638a4") +
		tooltext.Text("h.2f9a5f8efb91") +
		tooltext.Text("h.00f327af72c4"))
	out.WriteString(tooltext.Text("h.2171d2787a21") +
		tooltext.Text("h.f8c90a60b905") +
		tooltext.Text("h.760a6fe2755a"))
	out.WriteString(tooltext.Text("h.39a7fecfa95b") +
		tooltext.Text("h.0aaf246a7f77"))
	fmt.Fprint(&out, tooltext.Format("h.3ad9cef66da9"))
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
	fmt.Fprintf(&out, tooltext.Text("h.03dc4b7c50e8")+
		tooltext.Text("h.1d5fdf261c05"),
		doc.Segments, doc.Matched, doc.Mismatched, doc.FreshStart, doc.NotVisited)

	fmt.Fprint(&out, tooltext.Format("h.d417fa3f1b15", len(undeclared), total))
	out.WriteString(tooltext.Text("h.27d4b137e4a7") +
		tooltext.Text("h.cb654b7d4d2d") +
		tooltext.Text("h.3513aaa9f055"))
	if len(undeclared) == 0 {
		out.WriteString(tooltext.Text("h.be6c71d35815"))
		return out.String()
	}
	out.WriteString(tooltext.Text("h.10c0470d9e50"))
	for _, one := range undeclared {
		fmt.Fprintf(&out, "| `%02Xh` | `%02Xh` | %d |\n", one.From, one.To, one.Count)
	}
	return out.String()
}

func verdictText(verdict string) string {
	switch verdict {
	case "match":
		return tooltext.Text("h.22ef1b3f681c")
	case "mismatch":
		return tooltext.Text("h.47a668b4902b")
	case "fresh-start":
		return tooltext.Text("h.15fbcc75ee85")
	default:
		return tooltext.Text("h.b235c9ed6810")
	}
}
