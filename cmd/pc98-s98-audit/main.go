// Command pc98-s98-audit cross-checks local Hoot S98 traces with MSCDRV.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "用法：pc98-s98-audit MSCDRV.EXE SELECTOR:TRACE.s98 [...]")
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
			fatal(fmt.Errorf("trace 參數 %q 必須是 SELECTOR:PATH", argument))
		}
		selector, err := strconv.Atoi(parts[0])
		if err != nil {
			fatal(fmt.Errorf("selector %q：%w", parts[0], err))
		}
		raw, err := os.ReadFile(parts[1])
		if err != nil {
			fatal(err)
		}
		report, err := pc98music.AuditS98Track(driver, raw, selector)
		if err != nil {
			fatal(fmt.Errorf("selector %d：%w", selector, err))
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
