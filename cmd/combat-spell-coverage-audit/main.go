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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/spellcoverage"
)

func main() {
	source := flag.String("source", "internal/game", "runtime source file or package directory to audit")
	strict := flag.Bool("strict", false, "return failure when any row is incomplete")
	output := flag.String("output", "", "write the Markdown ledger to this path")
	quiet := flag.Bool("quiet", false, "skip the JSON dump on stdout")
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	report, err := spellcoverage.Build(pack, *source)
	if err != nil {
		log.Fatal(err)
	}
	if !*quiet {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			log.Fatal(err)
		}
	}
	if *output != "" {
		markdown, err := renderMarkdown(report)
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*output, markdown, 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintf(os.Stderr,
		tooltext.Text("h.96329173ba49")+
			tooltext.Text("h.773f451f2f1f"),
		report.SpellCount, report.CoveredCount, report.MissingCount,
		report.Original.Declarable, report.Original.Declared,
		report.Original.CombatCastable, report.Original.UnsupportedClass)
	if *strict && report.MissingCount != 0 {
		os.Exit(2)
	}
}

// renderMarkdown 產生逐支法術的台帳。分母是**原作那 100 筆**，
// 不是 game pack 宣告過的那幾支——後者當分母會得到與玩家無關的比例。
func renderMarkdown(report spellcoverage.Report) ([]byte, error) {
	table, err := gamepack.Spells()
	if err != nil {
		return nil, err
	}
	declared := map[int]spellcoverage.SpellReport{}
	for _, item := range report.Spells {
		declared[int(item.SpellID)] = item
	}

	var out strings.Builder
	out.WriteString(tooltext.Text("h.f6433cd125ea"))
	out.WriteString(tooltext.Text("h.e8b3da3bc176"))
	out.WriteString(tooltext.Text("h.b6fb04f4914a") +
		tooltext.Text("h.d73eba3498cf"))
	out.WriteString(tooltext.Text("h.1beae0403b79") +
		tooltext.Text("h.cc02bb12bdb7"))
	out.WriteString(tooltext.Text("h.e1e119766432") +
		tooltext.Text("h.7aeaa6433af5"))
	out.WriteString(tooltext.Text("h.86e0f965bc77") +
		tooltext.Text("h.10872e04ecbb"))

	original := report.Original
	fmt.Fprint(&out, tooltext.Format("h.082c81548ffe"))
	fmt.Fprint(&out, tooltext.Format("h.7e27e2a01918", original.TableSpells))
	fmt.Fprint(&out, tooltext.Format("h.3f555fcdab09", original.Placeholders))
	fmt.Fprint(&out, tooltext.Format("h.0a4df91b176f", original.Playable))
	fmt.Fprint(&out, tooltext.Format("h.0910816097f1", original.CampOnly))
	fmt.Fprint(&out, tooltext.Format("h.d9007af1f6c0", original.CombatCastable))
	fmt.Fprint(&out, tooltext.Format("h.dd58c83eb108", original.UnsupportedClass))
	fmt.Fprint(&out, tooltext.Format("h.e43f0dbc53ff", original.Declarable))
	fmt.Fprint(&out, tooltext.Format("h.31a132405c46", original.Declared))
	fmt.Fprint(&out, tooltext.Format("h.62a7a40517ec", original.HandlerObserved))
	fmt.Fprint(&out, tooltext.Format("h.e4601f10b654", original.Declarable-original.Declared))

	fmt.Fprint(&out, tooltext.Format("h.61baa020fc07", original.MissingEffectReady))
	fmt.Fprint(&out, tooltext.Format("h.02fb830488b5", original.MissingEffectUnhandled))
	fmt.Fprint(&out, tooltext.Format("h.6a58faaadd4a", original.MissingDamageClass))
	fmt.Fprint(&out, tooltext.Format("h.5449f67adc5d", original.EffectKindsUsed, original.EffectKindsInterpreted))
	out.WriteString(tooltext.Text("h.435d8aed43b4"))
	out.WriteString(tooltext.Text("h.cd8c188c4ccc"))
	if len(original.MissingEffectReadyNames) > 0 {
		fmt.Fprint(&out, tooltext.Format("h.2bdcf08f037e", strings.Join(original.MissingEffectReadyNames, "、")))
	}

	levels := make([]int, 0, len(original.MissingByLevel))
	for level := range original.MissingByLevel {
		levels = append(levels, level)
	}
	sort.Ints(levels)
	if len(levels) == 0 {
		out.WriteString(tooltext.Text("h.b844f287874c"))
	} else {
		out.WriteString(tooltext.Text("h.461d077e3e11"))
		for index, level := range levels {
			if index > 0 {
				out.WriteString("、")
			}
			fmt.Fprint(&out, tooltext.Format("h.c4ebc3a62035", level, original.MissingByLevel[level]))
		}
		out.WriteString("。\n")
	}
	out.WriteString(tooltext.Text("h.acf1c15ce9c1"))
	out.WriteString(tooltext.Text("h.351c02f9019f"))
	out.WriteString("|---:|---|---|---:|---|---|---|---|\n")
	for _, spell := range table.Spells {
		name := spell.Name
		class := spell.CasterClass
		if spell.Placeholder {
			name = tooltext.Text("h.72ca4342dab7")
			class = "—"
		}
		data := describeData(spell)
		handler, visual, sound := "—", "—", "—"
		if item, ok := declared[spell.SpellID]; ok {
			handler = item.Handler.Status
			visual = item.Visual.Status
			sound = item.Sound.Status
		} else if spell.CampOnly {
			handler = tooltext.Text("h.7521f10ebf82")
		} else if spell.Placeholder {
			handler = tooltext.Text("h.335fa0b90d1a")
		} else {
			handler = tooltext.Text("h.11b8e2909fa5")
		}
		fmt.Fprintf(&out, "| %d | %s | %s | %d | %s | %s | %s | %s |\n",
			spell.SpellID, name, class, spell.Level, data, handler, visual, sound)
	}
	return []byte(out.String()), nil
}

// describeData 把資料側已知的東西壓成一格，讓「缺的是效果不是資料」看得出來。
func describeData(spell gamepack.SpellEntry) string {
	parts := []string{tooltext.Format("h.4ce12bfd0cc0", spell.CastingTimeSegments)}
	switch spell.TargetModeKind {
	case "self":
		parts = append(parts, tooltext.Text("h.fb43354d2aae"))
	case "fixed":
		parts = append(parts, tooltext.Format("h.3e18e8bcd8d3", spell.TargetCount))
	case "area":
		parts = append(parts, tooltext.Format("h.0c242793a0b6", spell.AreaRadius))
	case "weighted-picks":
		parts = append(parts, tooltext.Text("h.25a1d4282753"))
	case "locked-or-area":
		parts = append(parts, tooltext.Text("h.833edd624755"))
	}
	if spell.RequiresSave() {
		parts = append(parts, tooltext.Format("h.d94dfd9df3af", spell.SaveCategoryName))
	}
	if spell.EffectID != 0 {
		parts = append(parts, tooltext.Format("h.c7df37db55e7", spell.EffectID))
	}
	return strings.Join(parts, "／")
}
