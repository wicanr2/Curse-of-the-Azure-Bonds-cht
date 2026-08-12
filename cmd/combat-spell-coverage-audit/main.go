package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/spellcoverage"
)

func main() {
	source := flag.String("source", "internal/game/combat_state.go", "runtime source to audit")
	strict := flag.Bool("strict", false, "return failure when any row is incomplete")
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	report, err := spellcoverage.Build(pack, *source)
	if err != nil {
		log.Fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "spell_count=%d covered=%d incomplete=%d\n", report.SpellCount, report.CoveredCount, report.MissingCount)
	if *strict && report.MissingCount != 0 {
		os.Exit(2)
	}
}
