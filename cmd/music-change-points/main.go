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
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	jsonPath := flag.String("json", "", "JSON 輸出路徑（給 cmd/remake-status 用）")
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
	fmt.Fprintf(&report, "# 原作的換曲點，remake 接上幾個\n\n")
	fmt.Fprintf(&report, "由 `cmd/music-change-points` 產生，不要手改。"+
		"換曲點的來源是 `docs/audit/pc98-music-triggers.md`（PC-98 的 Borland 符號表），"+
		"對照表在 `internal/audiomap`。\n\n")
	fmt.Fprintf(&report, "| 項目 | 數 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 原作換曲點 | %d |\n", stats.ChangePoints)
	fmt.Fprintf(&report, "| remake 有落點 | %d |\n", stats.Wired)
	fmt.Fprintf(&report, "| 被選到的相異曲目 | %d |\n", stats.DistinctTracks)
	fmt.Fprintf(&report, "| game pack 宣告的曲目 | %d |\n\n", stats.Tracks)
	fmt.Fprintf(&report, "⚠ 「有落點」＝ pack 有那首曲子而且有一條指向它的綁定。"+
		"**不表示**在原作會發的那一刻發——那要實機比對。\n\n")
	fmt.Fprintf(&report, "| 位置 | 事件 | 曲目 | remake 的 context | 有落點 |\n")
	fmt.Fprintf(&report, "|---|---|---:|---|---|\n")
	for _, result := range results {
		context := result.Point.Context
		if context == "" {
			context = "（照 ECL 段綁）"
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
