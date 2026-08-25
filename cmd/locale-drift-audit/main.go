package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/localeaudit"
)

func main() {
	root := flag.String("root", ".", "CoAB repository root")
	jsonOutput := flag.Bool("json", false, tooltext.Text("h.5f5ebcca8586"))
	flag.Parse()
	report, err := localeaudit.Run(*root)
	if err != nil {
		log.Fatal(err)
	}
	if *jsonOutput {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("assets.locale keys=%d product_refs=%d orphans=%d missing=%d\n", report.AssetsLocale.KeyCount, report.AssetsLocale.ReferencedCount, len(report.AssetsLocale.OrphanKeys), len(report.ProductUsage.MissingAssetsKeys))
		fmt.Printf("gamepack locale keys=%d referenced_message_ids=%d en_zh_equal=%d orphans=%d\n", report.GamepackLocale.KeyCount, report.GamepackLocale.ReferencedCount, len(report.GamepackLocale.EqualEnglish), len(report.GamepackLocale.OrphanKeys))
		fmt.Printf("product files=%d dynamic_text_calls=%d violations=%d\n", report.ProductUsage.ProductFiles, report.ProductUsage.DynamicTextCalls, localeaudit.ViolationCount(report))
		for _, issue := range report.Issues {
			fmt.Printf("%s %s %s: %s\n", issue.Severity, issue.Code, issue.Path, issue.Detail)
		}
	}
	if localeaudit.ViolationCount(report) != 0 {
		os.Exit(1)
	}
}
