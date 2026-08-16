package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/spellcoverage"
)

func main() {
	source := flag.String("source", "internal/game/combat_state.go", "runtime source to audit")
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
		"declared=%d covered=%d incomplete=%d｜原作可戰鬥施放 %d 支中已宣告 %d 支\n",
		report.SpellCount, report.CoveredCount, report.MissingCount,
		report.Original.CombatCastable, report.Original.Declared)
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
	out.WriteString("# 玩家法術覆蓋（原作 100 筆 → remake）\n\n")
	out.WriteString("由 `cmd/combat-spell-coverage-audit -output` 產生，不要手改。\n\n")
	out.WriteString("- 分母是原作法術主表的 **100 筆**（`gamepack/rules/spell-table.json`，" +
		"由 `cmd/spell-table-export` 從常駐資料段量出來，spec 1111）。\n")
	out.WriteString("- **占位**：沒有名字的 13 筆，玩家取不到。" +
		"**紮營**：`+0Bh = 0` 的 8 支，不會出現在戰鬥選單（spec 827）。\n")
	out.WriteString("- `handler`／`visual`／`sound` 三欄是 game pack 宣告與 runtime callsite 的機器觀察，" +
		"**不代表原版規則、時序或素材已還原**。\n")
	out.WriteString("- 「資料」欄一律有值：施法時間、目標形狀、豁免、持續時間係數整張表都在，" +
		"沒有 handler 的法術缺的是**效果**，不是資料。\n\n")

	original := report.Original
	fmt.Fprintf(&out, "## 摘要\n\n| 分類 | 支數 |\n|---|---:|\n")
	fmt.Fprintf(&out, "| 表的筆數 | %d |\n", original.TableSpells)
	fmt.Fprintf(&out, "| ├ 占位（玩家取不到）| %d |\n", original.Placeholders)
	fmt.Fprintf(&out, "| └ 玩家取得到 | %d |\n", original.Playable)
	fmt.Fprintf(&out, "| 　├ 只能紮營施放 | %d |\n", original.CampOnly)
	fmt.Fprintf(&out, "| 　└ **戰鬥可施放（真正的分母）** | **%d** |\n", original.CombatCastable)
	fmt.Fprintf(&out, "| 　　├ game pack 已宣告 | %d |\n", original.Declared)
	fmt.Fprintf(&out, "| 　　├ 其中 runtime handler 已觀察到 | %d |\n", original.HandlerObserved)
	fmt.Fprintf(&out, "| 　　└ **尚未宣告** | **%d** |\n", original.CombatCastable-original.Declared)

	levels := make([]int, 0, len(original.MissingByLevel))
	for level := range original.MissingByLevel {
		levels = append(levels, level)
	}
	sort.Ints(levels)
	out.WriteString("\n未宣告的戰鬥法術，依環數：")
	for index, level := range levels {
		if index > 0 {
			out.WriteString("、")
		}
		fmt.Fprintf(&out, "%d 環 %d 支", level, original.MissingByLevel[level])
	}
	out.WriteString("。\n\n## 逐支\n\n")
	out.WriteString("| 編號 | 名稱 | 職業 | 環 | 資料 | handler | visual | sound |\n")
	out.WriteString("|---:|---|---|---:|---|---|---|---|\n")
	for _, spell := range table.Spells {
		name := spell.Name
		class := spell.CasterClass
		if spell.Placeholder {
			name = "（占位）"
			class = "—"
		}
		data := describeData(spell)
		handler, visual, sound := "—", "—", "—"
		if item, ok := declared[spell.SpellID]; ok {
			handler = item.Handler.Status
			visual = item.Visual.Status
			sound = item.Sound.Status
		} else if spell.CampOnly {
			handler = "紮營法術"
		} else if spell.Placeholder {
			handler = "取不到"
		} else {
			handler = "未宣告"
		}
		fmt.Fprintf(&out, "| %d | %s | %s | %d | %s | %s | %s | %s |\n",
			spell.SpellID, name, class, spell.Level, data, handler, visual, sound)
	}
	return []byte(out.String()), nil
}

// describeData 把資料側已知的東西壓成一格，讓「缺的是效果不是資料」看得出來。
func describeData(spell gamepack.SpellEntry) string {
	parts := []string{fmt.Sprintf("%d 節", spell.CastingTimeSegments)}
	switch spell.TargetModeKind {
	case "self":
		parts = append(parts, "自己")
	case "fixed":
		parts = append(parts, fmt.Sprintf("%d 個目標", spell.TargetCount))
	case "area":
		parts = append(parts, fmt.Sprintf("半徑 %d", spell.AreaRadius))
	case "weighted-picks":
		parts = append(parts, "逐一點選")
	case "locked-or-area":
		parts = append(parts, "鎖定目標")
	}
	if spell.RequiresSave() {
		parts = append(parts, fmt.Sprintf("豁免 %s", spell.SaveCategoryName))
	}
	if spell.EffectID != 0 {
		parts = append(parts, fmt.Sprintf("效果 %d", spell.EffectID))
	}
	return strings.Join(parts, "／")
}
