// Command pc98-s98-audit cross-checks local Hoot S98 traces with MSCDRV.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, tooltext.Text("pc98_s98_audit.usage"))
		os.Exit(2)
	}
	driver, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatal(err)
	}
	var reports []pc98music.S98TrackAudit
	for _, argument := range os.Args[2:] {
		parts := strings.SplitN(argument, ":", 2)
		if len(parts) != 2 {
			fatal(tooltext.Errorf("pc98_s98_audit.trace_argument", argument))
		}
		selector, err := strconv.Atoi(parts[0])
		if err != nil {
			fatal(tooltext.Errorf("pc98_s98_audit.selector_parse", parts[0], err))
		}
		raw, err := os.ReadFile(parts[1])
		if err != nil {
			fatal(err)
		}
		report, err := pc98music.AuditS98Track(driver, raw, selector)
		if err != nil {
			fatal(tooltext.Errorf("pc98_s98_audit.selector_audit", selector, err))
		}
		reports = append(reports, report)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(reports); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
