package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcatalog"
)

func main() {
	archive := flag.String("archive", "curseoftheazurebonds.zip", "original CoAB archive")
	reviews := flag.String("reviews", "docs/audit/ecl-ordered-effect-reviews.json", "review overlay for stable candidate IDs; empty disables")
	output := flag.String("output", "", "write deterministic JSON to this path; stdout when empty")
	summaryOutput := flag.String("summary-output", "", "write generated Markdown summary to this path")
	check := flag.String("check", "", "compare deterministic JSON with this committed artifact")
	checkSummary := flag.String("check-summary", "", "compare generated Markdown with this committed artifact")
	phasesOutput := flag.String("phases-output", "", "write the ordered-effect phase ledger JSON to this path")
	phasesSummary := flag.String("phases-summary-output", "", "write the ordered-effect phase ledger Markdown to this path")
	checkPhases := flag.String("check-phases", "", "compare the phase ledger JSON with this committed artifact")
	checkPhasesSummary := flag.String("check-phases-summary", "", "compare the phase ledger Markdown with this committed artifact")
	flag.Parse()

	catalog, err := eclcatalog.BuildFile(*archive)
	if err != nil {
		log.Fatal(err)
	}
	if *reviews != "" {
		reviewData, readErr := os.ReadFile(*reviews)
		if readErr != nil {
			log.Fatal(readErr)
		}
		if err := eclcatalog.ApplyReviewLedger(&catalog, reviewData); err != nil {
			log.Fatal(err)
		}
	}
	data, err := eclcatalog.Encode(catalog)
	if err != nil {
		log.Fatal(err)
	}
	if *check != "" {
		current, err := os.ReadFile(*check)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(current, data) {
			log.Fatalf("ECL event catalog drift: regenerate %s", *check)
		}
	}
	summary := eclcatalog.EncodeMarkdown(catalog)
	if *checkSummary != "" {
		current, err := os.ReadFile(*checkSummary)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(current, summary) {
			log.Fatalf("ECL event catalog summary drift: regenerate %s", *checkSummary)
		}
	}
	if *output != "" {
		if err := os.WriteFile(*output, data, 0o644); err != nil {
			log.Fatal(err)
		}
	} else if *check == "" && *summaryOutput == "" && *checkSummary == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			log.Fatal(err)
		}
	}
	if *summaryOutput != "" {
		if err := os.WriteFile(*summaryOutput, summary, 0o644); err != nil {
			log.Fatal(err)
		}
	}

	ledger := eclcatalog.BuildPhaseLedger()
	if err := eclcatalog.VerifyPhaseCoverage(catalog, ledger); err != nil {
		log.Fatal(err)
	}
	phaseData, err := eclcatalog.EncodePhaseJSON(ledger)
	if err != nil {
		log.Fatal(err)
	}
	phaseMarkdown := eclcatalog.EncodePhaseMarkdown(ledger)
	writeOrCheck(*checkPhases, *phasesOutput, phaseData, "ordered-effect phase ledger")
	writeOrCheck(*checkPhasesSummary, *phasesSummary, phaseMarkdown, "ordered-effect phase summary")

	fmt.Fprintf(os.Stderr,
		"members=%d blocks=%d entries=%d instructions=%d ordered_effect_candidates=%d phase_rows=%d\n",
		catalog.Summary.MemberCount, catalog.Summary.BlockCount,
		catalog.Summary.LifecycleEntryCount, catalog.Summary.UniqueReachableInstructionCount,
		catalog.Summary.OrderedEffectCandidateCount, len(ledger.Rows),
	)
}

// writeOrCheck keeps the generated artifact and the committed copy in one
// direction only: a drift is a failure, never a silent overwrite of the check
// target.
func writeOrCheck(checkPath, outputPath string, data []byte, label string) {
	if checkPath != "" {
		current, err := os.ReadFile(checkPath)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(current, data) {
			log.Fatalf("%s drift: regenerate %s", label, checkPath)
		}
	}
	if outputPath != "" {
		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			log.Fatal(err)
		}
	}
}
