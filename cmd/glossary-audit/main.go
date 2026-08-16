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
	out.WriteString("# 譯名一致性稽核\n\n")
	out.WriteString("由 `cmd/glossary-audit` 產生，不要手改。表在 " +
		"[`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md)。\n\n")
	out.WriteString("## 掃描範圍\n\n| 檔案 | 字串數 |\n|---|---:|\n")
	for _, source := range report.Sources {
		fmt.Fprintf(&out, "| `%s` | %d |\n", source.Path, source.Count)
	}

	fmt.Fprintf(&out, "\n## 結果\n\n| 項目 | 數量 |\n|---|---:|\n")
	fmt.Fprintf(&out, "| 詞條 | %d |\n", len(report.Terms))
	fmt.Fprintf(&out, "| **不一致** | **%d** |\n", len(report.Issues))

	if len(report.Issues) > 0 {
		out.WriteString("\n## 不一致\n\n| 代碼 | 詞條 | 說明 |\n|---|---|---|\n")
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
	out.WriteString("\n## 逐詞條\n\n")
	for _, category := range categories {
		terms := byCategory[category]
		sort.Slice(terms, func(i, j int) bool { return terms[i].Source < terms[j].Source })
		fmt.Fprintf(&out, "### %s\n\n| 原文 | 繁中 | 出現次數 | 禁用寫法 |\n|---|---|---:|---|\n", category)
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
