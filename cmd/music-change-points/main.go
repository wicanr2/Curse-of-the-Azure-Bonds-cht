// Command music-change-points 回答「原作的換曲點，remake 接上幾個」。
//
// ★ 存在的理由：這個數字以前只活在報表的敘述裡（「13 個裡 10 個接上了」），
// 而接上第 11 個的時候沒有任何東西會更新它。現在分母來自
// `internal/audiomap.ChangePoints`（掃描結果，見
// `docs/audit/pc98-music-triggers.md`），分子現場查 game pack。
//
// ⚠ 「接上」只表示**有落點**——pack 有那首曲子而且有一條指向它的綁定。
// 「在原作會發的那一刻發」是另一件事，要實機比對。
//
// 用法：
//
//	go run ./cmd/music-change-points -output docs/audit/music-change-points.md \
//	    -json docs/audit/music-change-points.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiomap"
)

type summary struct {
	ChangePoints int `json:"change_points"`
	Wired        int `json:"wired"`
	Tracks       int `json:"tracks"`
	// DistinctTracks 是被換曲點選到的相異曲目數。
	DistinctTracks int `json:"distinct_tracks"`
}

func main() {
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	jsonPath := flag.String("json", "", tooltext.Text("h.665ef4d19c47"))
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	trackBySelector := map[int]string{}
	for _, track := range pack.MusicTracks {
		trackBySelector[int(track.ReferenceSelector)] = track.ID
	}
	bindings := make([]audiomap.Binding, 0, len(pack.MusicBindings))
	for _, binding := range pack.MusicBindings {
		bindings = append(bindings, audiomap.Binding{
			Context: binding.Context, TrackID: binding.TrackID})
	}
	results := audiomap.Resolve(trackBySelector, bindings)

	distinct := map[string]bool{}
	for _, result := range results {
		if result.TrackID != "" {
			distinct[result.TrackID] = true
		}
	}
	stats := summary{
		ChangePoints:   len(results),
		Wired:          audiomap.Wired(results),
		Tracks:         len(pack.MusicTracks),
		DistinctTracks: len(distinct),
	}

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.558b94319f3f"))
	fmt.Fprint(&report, tooltext.Text("h.964d43066305")+
		tooltext.Text("h.7e28bef717b3")+
		tooltext.Text("h.ce75b3757941"))
	fmt.Fprint(&report, tooltext.Format("h.e811720ed9a3"))
	fmt.Fprint(&report, tooltext.Format("h.f3d550613e4e", stats.ChangePoints))
	fmt.Fprint(&report, tooltext.Format("h.3473de6fe825", stats.Wired))
	fmt.Fprint(&report, tooltext.Format("h.63b44d9315a0", stats.DistinctTracks))
	fmt.Fprint(&report, tooltext.Format("h.37c6d52044b1", stats.Tracks))
	fmt.Fprint(&report, tooltext.Text("h.aaa3eb5d99fd")+
		tooltext.Text("h.865b7f882270"))
	fmt.Fprint(&report, tooltext.Format("h.7dccd6fc1ae3"))
	fmt.Fprintf(&report, "|---|---|---:|---|---|\n")
	for _, result := range results {
		context := result.Point.Context
		if context == "" {
			context = tooltext.Text("h.7a81abbd938e")
		} else {
			context = "`" + context + "`"
		}
		mark := "—"
		if result.Wired {
			mark = "✅"
		}
		fmt.Fprintf(&report, "| `%s` | %s | %d | %s | %s |\n",
			result.Point.Site, result.Point.Event, result.Point.Selector, context, mark)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if writeErr := os.WriteFile(*output, []byte(text), 0o644); writeErr != nil {
		log.Fatal(writeErr)
	}
	if *jsonPath != "" {
		payload, marshalErr := json.MarshalIndent(stats, "", "  ")
		if marshalErr != nil {
			log.Fatal(marshalErr)
		}
		if writeErr := os.WriteFile(*jsonPath, append(payload, '\n'), 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
	}
	fmt.Fprintf(os.Stderr, "change-points=%d wired=%d distinct-tracks=%d tracks=%d\n",
		stats.ChangePoints, stats.Wired, stats.DistinctTracks, stats.Tracks)
}
