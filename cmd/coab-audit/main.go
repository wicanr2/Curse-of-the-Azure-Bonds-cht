package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/sourceaudit"
)

func main() {
	root := flag.String("root", ".", "repository root")
	baselinePath := flag.String("baseline", "docs/audit/go-han-literals-baseline.json", "baseline JSON")
	writeBaseline := flag.Bool("write-baseline", false, "replace baseline with current findings")
	flag.Parse()

	findings, err := sourceaudit.ScanGoHanLiterals(*root)
	if err != nil {
		log.Fatal(err)
	}
	if *writeBaseline {
		data, err := sourceaudit.EncodeBaseline(findings)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, '\n')
		if err := os.WriteFile(*baselinePath, data, 0o644); err != nil {
			log.Fatal(err)
		}
		printSummary(findings)
		return
	}
	baseline, err := sourceaudit.LoadBaseline(*baselinePath)
	if err != nil {
		log.Fatal(err)
	}
	added, removed := sourceaudit.Compare(findings, baseline.Findings)
	if len(added) != 0 || len(removed) != 0 {
		for _, finding := range added {
			fmt.Printf("NEW %s %s %s x%d\n", finding.Path, finding.Function, finding.SHA256, finding.Occurrences)
		}
		for _, finding := range removed {
			fmt.Printf("REMOVED %s %s %s x%d\n", finding.Path, finding.Function, finding.SHA256, finding.Occurrences)
		}
		log.Fatalf("Go Han literal baseline drift: new=%d removed=%d; migrate data or refresh the reduced baseline in the same change", len(added), len(removed))
	}
	printSummary(findings)
}

func printSummary(findings []sourceaudit.Finding) {
	summary := sourceaudit.Summary(findings)
	categories := make([]string, 0, len(summary))
	for category := range summary {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	total := 0
	for _, category := range categories {
		fmt.Printf("%s=%d\n", category, summary[category])
		total += summary[category]
	}
	fmt.Printf("total=%d\n", total)
}
