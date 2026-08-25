// Command glossary-audit renders the translation glossary report and fails when
// a proper noun has drifted.
//
// 用法：
//
//	go run ./cmd/glossary-audit -output docs/audit/glossary.md -json docs/audit/glossary.json
//	go run ./cmd/glossary-audit -check      # 只回傳結果，不寫檔
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/glossary"
)

func main() {
	root := flag.String("root", ".", "repository root")
	output := flag.String("output", "", "write the Markdown report to this path")
	outputJSON := flag.String("json", "", "write the machine-readable report to this path")
	check := flag.Bool("check", false, "only report; write nothing")
	flag.Parse()

	report, err := glossary.Run(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !*check {
		if *outputJSON != "" {
			encoded, err := json.MarshalIndent(report, "", " ")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		}
		if *output != "" {
			if err := os.WriteFile(*output, renderMarkdown(report), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		}
	}
	for _, issue := range report.Issues {
		fmt.Fprintf(os.Stderr, "%s: %s\n", issue.Code, issue.Detail)
	}
	fmt.Fprintf(os.Stderr, "terms=%d issues=%d\n", len(report.Terms), len(report.Issues))
	if len(report.Issues) > 0 {
		os.Exit(1)
	}
}

func renderMarkdown(report glossary.Report) []byte {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.9137082673f5"))
	out.WriteString(tooltext.Text("h.81ed26c4db2f") +
		"[`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)。\n\n")
	out.WriteString(tooltext.Text("h.a55b2834f9a5"))
	for _, source := range report.Sources {
		fmt.Fprintf(&out, "| `%s` | %d |\n", source.Path, source.Count)
	}

	fmt.Fprint(&out, tooltext.Format("h.fdbec7290a7f"))
	fmt.Fprint(&out, tooltext.Format("h.a2e3db92cbe4", len(report.Terms)))
	fmt.Fprint(&out, tooltext.Format("h.e91815f54eec", len(report.Issues)))

	if len(report.Issues) > 0 {
		out.WriteString(tooltext.Text("h.8c0ccd9edada"))
		for _, issue := range report.Issues {
			fmt.Fprintf(&out, "| `%s` | `%s` | %s |\n", issue.Code, issue.Term, issue.Detail)
		}
	}

	byCategory := map[string][]glossary.Term{}
	for _, term := range report.Terms {
		byCategory[term.Category] = append(byCategory[term.Category], term)
	}
	categories := make([]string, 0, len(byCategory))
	for category := range byCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	out.WriteString(tooltext.Text("h.830b313707a0"))
	for _, category := range categories {
		terms := byCategory[category]
		sort.Slice(terms, func(i, j int) bool { return terms[i].Source < terms[j].Source })
		fmt.Fprint(&out, tooltext.Format("h.3ea6b930ce13", category))
		for _, term := range terms {
			forbidden := "—"
			if len(term.Forbidden) > 0 {
				forbidden = strings.Join(term.Forbidden, "、")
			}
			fmt.Fprintf(&out, "| `%s` | %s | %d | %s |\n",
				term.Source, term.Chinese, term.Uses, forbidden)
		}
		out.WriteString("\n")
	}
	return []byte(strings.TrimRight(out.String(), "\n") + "\n")
}
